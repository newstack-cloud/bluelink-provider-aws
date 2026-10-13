package lambdards

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// The cluster endpoint is read from state when the function's environment is built,
// so neither resource is written by this link.
func (l *functionClusterLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *functionClusterLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, region, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getClusterLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)

	// The execution role is only needed for the IAM auth grant; skip the role lookup for
	// password auth (which grants no IAM permission).
	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceAInfo,
			LoadRoleInfo:           annotations.authMode == "iam",
		},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	output := &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}

	// IAM rds-db:connect grant (only for token-based auth).
	if annotations.authMode == "iam" {
		output, err = l.reconcileConnectGrant(ctx, providerCtx, input, setupCtx, annotations, region)
		if err != nil {
			return nil, err
		}
	}

	// Open the security-group path from the function to the cluster on the database port.
	targetSecurityGroupIDs, hasTargetSG := extractClusterSecurityGroupIDs(input.ResourceBInfo)
	if !hasTargetSG {
		return nil, fmt.Errorf(
			"security group could not be retrieved from the linked to %q Aurora cluster resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	return linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:                 linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
			Region:                 region,
			TargetSecurityGroupIDs: targetSecurityGroupIDs,
			TargetPort:             annotations.port,
		},
		output,
	)
}

func (l *functionClusterLinkActions) reconcileConnectGrant(
	ctx context.Context,
	providerCtx provider.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	annotations *clusterLinkAnnotations,
	region string,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// The role is locked around each grant below rather than for this whole
	// phase, which goes on to do slow work the role has no part in.
	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	sid := createConnectSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if _, err := linkutils.GrantRoleAccessUnderLock(
			ctx,
			linkutils.RoleAccessLockFor(
				input.ResourceService,
				input.LinkID,
				input.ResourceAInfo,
				setupCtx,
				providerCtx,
			),
			iamService,
			linkutils.RoleAccessGrant{
				RoleName: setupCtx.RoleName,
				SID:      sid,
			}); err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}, nil
	}

	connectResource, err := rdsClusterConnectResourceARN(input.ResourceBInfo, region, annotations.dbUser)
	if err != nil {
		return nil, err
	}

	result, err := linkutils.GrantRoleAccessUnderLock(
		ctx,
		linkutils.RoleAccessLockFor(
			input.ResourceService,
			input.LinkID,
			input.ResourceAInfo,
			setupCtx,
			providerCtx,
		),
		iamService,
		linkutils.RoleAccessGrant{
			RoleName:  setupCtx.RoleName,
			SID:       sid,
			Statement: connectStatement(sid, connectResource),
		})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q rds-db:connect on Aurora cluster %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	return connectLinkOutput(input, setupCtx.RoleResourceName, sid, connectResource, result), nil
}

// Builds the rds-db:connect resource ARN for the cluster:
// arn:aws:rds-db:<region>:<account>:dbuser:<cluster-resource-id>/<dbUser>.
// The cluster resource id (cluster-...) is read directly from state; the account is extracted
// from the cluster ARN.
func rdsClusterConnectResourceARN(
	clusterInfo *provider.ResourceInfo,
	region, dbUser string,
) (string, error) {
	clusterARN, hasClusterARN := extractClusterARN(clusterInfo)
	if !hasClusterARN {
		return "", fmt.Errorf(
			"ARN could not be retrieved from the linked to %q Aurora cluster resource",
			pluginutils.GetResourceName(clusterInfo),
		)
	}
	resourceID, hasResourceID := extractClusterResourceID(clusterInfo)
	if !hasResourceID {
		return "", fmt.Errorf(
			"resource id could not be retrieved from the linked to %q Aurora cluster resource",
			pluginutils.GetResourceName(clusterInfo),
		)
	}
	account, err := parseClusterARN(clusterARN)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("arn:aws:rds-db:%s:%s:dbuser:%s/%s", region, account, resourceID, dbUser), nil
}

// Extracts the account id from a cluster ARN of the form
// arn:aws:rds:<region>:<account>:cluster:<cluster-id>.
func parseClusterARN(clusterARN string) (account string, err error) {
	parts := strings.Split(clusterARN, ":")
	if len(parts) < 7 || parts[5] != "cluster" {
		return "", fmt.Errorf("unexpected Aurora cluster ARN format: %q", clusterARN)
	}
	return parts[4], nil
}

func extractClusterEndpoint(clusterInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(clusterInfo)
	node, has := pluginutils.GetValueByPath("$.endpoint.address", spec)
	if !has {
		return "", false
	}
	return core.StringValue(node), true
}

func extractClusterReaderEndpoint(clusterInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(clusterInfo)
	node, has := pluginutils.GetValueByPath("$.readEndpoint.address", spec)
	if !has {
		return "", false
	}
	return core.StringValue(node), true
}

func extractClusterARN(clusterInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(clusterInfo)
	node, has := pluginutils.GetValueByPath("$.dbClusterArn", spec)
	if !has {
		return "", false
	}
	return core.StringValue(node), true
}

func extractClusterResourceID(clusterInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(clusterInfo)
	node, has := pluginutils.GetValueByPath("$.dbClusterResourceId", spec)
	if !has {
		return "", false
	}
	return core.StringValue(node), true
}

func extractClusterSecurityGroupIDs(clusterInfo *provider.ResourceInfo) ([]string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(clusterInfo)
	node, has := pluginutils.GetValueByPath("$.vpcSecurityGroupIds", spec)
	if !has || node == nil || len(node.Items) == 0 {
		return nil, false
	}

	// Every group is returned rather than the first. Which one the link should pair
	// against is decided by matching them to the groups the flex VPC created, not by
	// their order here.
	groupIDs := make([]string, 0, len(node.Items))
	for _, item := range node.Items {
		if groupID := core.StringValue(item); groupID != "" {
			groupIDs = append(groupIDs, groupID)
		}
	}

	return groupIDs, len(groupIDs) > 0
}

type clusterEnvVarNameSet struct {
	host       string
	port       string
	database   string
	readerHost string
}

func (s clusterEnvVarNameSet) all() []string {
	return []string{s.host, s.port, s.database, s.readerHost}
}

func clusterEnvVarNames(prefix string) clusterEnvVarNameSet {
	return clusterEnvVarNameSet{
		host:       fmt.Sprintf("%s_HOST", prefix),
		port:       fmt.Sprintf("%s_PORT", prefix),
		database:   fmt.Sprintf("%s_DATABASE", prefix),
		readerHost: fmt.Sprintf("%s_READER_HOST", prefix),
	}
}

func clusterEnvVarPrefix(userDefinedPrefix string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedPrefix != "" {
		return userDefinedPrefix
	}
	return "RDS_" + strings.ToUpper(pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

type clusterLinkAnnotations struct {
	populateEnvVars bool
	envVarPrefix    string
	authMode        string
	dbUser          string
	port            int32
	databaseName    string
	readerEndpoint  bool
}

func getClusterLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *clusterLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.rds.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.rds.populateEnvVars",
			Default:     true,
		},
	)
	envVarPrefix, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.rds.%s.envVarPrefix", otherResourceInfo.ResourceName),
		},
	)
	authMode, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.rds.%s.authMode", otherResourceInfo.ResourceName),
			Default: "password",
		},
	)
	dbUser, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.rds.%s.dbUser", otherResourceInfo.ResourceName),
			Default: "*",
		},
	)
	port, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     fmt.Sprintf("aws.lambda.rds.%s.port", otherResourceInfo.ResourceName),
			Default: 5432,
		},
	)
	databaseName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.rds.%s.databaseName", otherResourceInfo.ResourceName),
		},
	)
	readerEndpoint, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:     fmt.Sprintf("aws.lambda.rds.%s.readerEndpoint", otherResourceInfo.ResourceName),
			Default: false,
		},
	)

	return &clusterLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarPrefix:    envVarPrefix,
		authMode:        authMode,
		dbUser:          dbUser,
		port:            int32(port),
		databaseName:    databaseName,
		readerEndpoint:  readerEndpoint,
	}
}
