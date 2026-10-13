package lambdards

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// The proxy endpoint is read from state when the function's environment is built and
// the connect grant is reconciled, so neither resource is written by this link.
func (l *functionProxyLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *functionProxyLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, region, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getProxyLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)

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

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}

	// IAM rds-db:connect grant (only for token-based auth).
	if annotations.authMode == "iam" {
		output, err = l.reconcileConnectGrant(ctx, providerCtx, input, setupCtx, annotations, region)
		if err != nil {
			return nil, err
		}
	}

	// Open the security-group path from the function to the proxy on the database port.
	targetSecurityGroupIDs, hasTargetSG := extractProxySecurityGroupIDs(input.ResourceBInfo)
	if !hasTargetSG {
		return nil, fmt.Errorf(
			"security group could not be retrieved from the linked to %q RDS proxy resource",
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

func (l *functionProxyLinkActions) reconcileConnectGrant(
	ctx context.Context,
	providerCtx provider.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
	setupCtx *linkutils.LambdaLinkSetupContext,
	annotations *proxyLinkAnnotations,
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

	connectResource, err := rdsConnectResourceARN(input.ResourceBInfo, region, annotations.dbUser)
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
				"cannot grant Lambda %q rds-db:connect on RDS proxy %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	return connectLinkOutput(input, setupCtx.RoleResourceName, sid, connectResource, result), nil
}

func connectStatement(sid, connectResource string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   []string{"rds-db:connect"},
		"Resource": []string{connectResource},
	}
}

func connectLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, connectResource string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		core.MappingNodeFields(
			"sid", core.MappingNodeFromString(sid),
			"effect", core.MappingNodeFromString("Allow"),
			"action", &core.MappingNode{Items: []*core.MappingNode{
				core.MappingNodeFromString("rds-db:connect"),
			}},
			"resource", &core.MappingNode{Items: []*core.MappingNode{
				core.MappingNodeFromString(connectResource),
			}},
		),
	)

	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

// Builds the rds-db:connect resource ARN for the proxy:
// arn:aws:rds-db:<region>:<account>:dbuser:<proxy-resource-id>/<dbUser>.
// The proxy resource id (prx-...) and account are extracted from the proxy ARN.
func rdsConnectResourceARN(
	proxyInfo *provider.ResourceInfo,
	region, dbUser string,
) (string, error) {
	proxyARN, hasProxyARN := extractProxyARN(proxyInfo)
	if !hasProxyARN {
		return "", fmt.Errorf(
			"ARN could not be retrieved from the linked to %q RDS proxy resource",
			pluginutils.GetResourceName(proxyInfo),
		)
	}
	account, proxyResourceID, err := parseProxyARN(proxyARN)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("arn:aws:rds-db:%s:%s:dbuser:%s/%s", region, account, proxyResourceID, dbUser), nil
}

// Extracts the account id and proxy resource id (prx-...) from a proxy ARN of the
// form arn:aws:rds:<region>:<account>:db-proxy:<proxy-resource-id>.
func parseProxyARN(proxyARN string) (account, proxyResourceID string, err error) {
	parts := strings.Split(proxyARN, ":")
	if len(parts) < 7 || parts[5] != "db-proxy" {
		return "", "", fmt.Errorf("unexpected RDS proxy ARN format: %q", proxyARN)
	}
	return parts[4], parts[6], nil
}

func extractProxyEndpoint(proxyInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(proxyInfo)
	node, has := pluginutils.GetValueByPath("$.endpoint", spec)
	if !has {
		return "", false
	}
	return core.StringValue(node), true
}

func extractProxyARN(proxyInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(proxyInfo)
	node, has := pluginutils.GetValueByPath("$.dbProxyArn", spec)
	if !has {
		return "", false
	}
	return core.StringValue(node), true
}

func extractProxySecurityGroupIDs(proxyInfo *provider.ResourceInfo) ([]string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(proxyInfo)
	node, has := pluginutils.GetValueByPath("$.vpcSecurityGroupIds", spec)
	if !has || node == nil || len(node.Items) == 0 {
		return nil, false
	}

	// Every group is returned rather than the first. Which one the link should pair
	// against is decided by matching them to the groups the flex VPC minted, not by
	// their order here.
	groupIDs := make([]string, 0, len(node.Items))
	for _, item := range node.Items {
		if groupID := core.StringValue(item); groupID != "" {
			groupIDs = append(groupIDs, groupID)
		}
	}

	return groupIDs, len(groupIDs) > 0
}

type proxyEnvVarNameSet struct {
	host     string
	port     string
	database string
}

func (s proxyEnvVarNameSet) all() []string {
	return []string{s.host, s.port, s.database}
}

func proxyEnvVarNames(prefix string) proxyEnvVarNameSet {
	return proxyEnvVarNameSet{
		host:     prefix + "_HOST",
		port:     prefix + "_PORT",
		database: prefix + "_DATABASE",
	}
}

func proxyEnvVarPrefix(userDefinedPrefix string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedPrefix != "" {
		return userDefinedPrefix
	}
	return "RDS_" + strings.ToUpper(pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createConnectSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("RDSConnect%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type proxyLinkAnnotations struct {
	populateEnvVars bool
	envVarPrefix    string
	authMode        string
	dbUser          string
	port            int32
	databaseName    string
}

func getProxyLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *proxyLinkAnnotations {
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

	return &proxyLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarPrefix:    envVarPrefix,
		authMode:        authMode,
		dbUser:          dbUser,
		port:            int32(port),
		databaseName:    databaseName,
	}
}
