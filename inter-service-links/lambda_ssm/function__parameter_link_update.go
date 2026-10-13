package lambdassm

import (
	"context"
	"errors"
	"fmt"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// The parameter's name is read from state when the function's access is granted, so
// neither resource is written by this link.
func (l *functionParameterLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *functionParameterLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, region, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceAInfo,
			LoadRoleInfo:           true,
		},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	// The role is locked around each grant below rather than for this whole
	// phase, which goes on to do slow work the role has no part in.
	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	sid := createParameterAccessSID(input.ResourceBInfo)

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
		// The endpoint this link provisioned is removed here; returning early would
		// leave it, and its security group, behind.
		ec2Service, err := l.getEC2Service(ctx, providerCtx)
		if err != nil {
			return nil, err
		}
		return linkutils.ReconcileLinkNetworking(
			ctx,
			ec2Service,
			input,
			ssmParameterNetworkingActivation(setupCtx, region),
			&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
		)
	}

	// The IAM policy grants access to the parameter by ARN.
	parameterARN, hasParameterARN := extractParameterARN(input.ResourceBInfo)
	if !hasParameterARN {
		return nil, fmt.Errorf(
			"parameter ARN could not be retrieved from the linked to %q SSM parameter resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	annotations := getParameterLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
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
			Statement: parameterAccessStatement(sid, parameterARN, annotations.accessLevel),
		})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q access to SSM parameter %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	statementNode := specAccessStatementNode(
		sid,
		parameterActionsForAccessLevel(annotations.accessLevel),
		[]string{parameterARN},
	)
	output := accessLinkOutput(input, setupCtx.RoleResourceName, sid, statementNode, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-isolated caller reaches SSM through an interface VPC endpoint; this is a no-op
	// for non-VPC functions.
	return linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		ssmParameterNetworkingActivation(setupCtx, region),
		output,
	)
}

func parameterAccessStatement(sid, parameterARN, accessLevel string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   parameterActionsForAccessLevel(accessLevel),
		"Resource": []string{parameterARN},
	}
}

func accessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid string,
	statementNode *core.MappingNode,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		statementNode,
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it: inline
	// placements map the statement by Sid; managed (overflow) placements map the attached
	// managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(
		mappings,
		roleLinkData,
		roleResourceName,
		linkDataKey,
		sid,
		result,
	)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

func specAccessStatementNode(sid string, actions, resources []string) *core.MappingNode {
	actionItems := make([]*core.MappingNode, len(actions))
	for i, action := range actions {
		actionItems[i] = core.MappingNodeFromString(action)
	}
	resourceItems := make([]*core.MappingNode, len(resources))
	for i, resource := range resources {
		resourceItems[i] = core.MappingNodeFromString(resource)
	}
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: actionItems},
		"resource", &core.MappingNode{Items: resourceItems},
	)
}

func extractParameterName(parameterInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(parameterInfo)
	nameNode, has := pluginutils.GetValueByPath("$.name", spec)
	if !has {
		return "", false
	}
	return core.StringValue(nameNode), true
}

func extractParameterARN(parameterInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(parameterInfo)
	arnNode, has := pluginutils.GetValueByPath("$.arn", spec)
	if !has {
		return "", false
	}
	return core.StringValue(arnNode), true
}

func parameterEnvVarName(userDefinedEnvVarName string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("SSM_PARAMETER_%s", resourceInfo.ResourceName)
}

func createParameterAccessSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("SSMAccess%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

func parameterActionsForAccessLevel(accessLevel string) []string {
	switch accessLevel {
	case "write":
		return []string{"ssm:PutParameter"}
	case "readwrite":
		return []string{
			"ssm:GetParameter",
			"ssm:GetParameters",
			"ssm:GetParametersByPath",
			"ssm:PutParameter",
		}
	case "read":
		fallthrough
	default:
		return []string{"ssm:GetParameter", "ssm:GetParameters", "ssm:GetParametersByPath"}
	}
}

type parameterLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
	accessLevel     string
}

func getParameterLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *parameterLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.ssm.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.ssm.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.ssm.%s.envVarName", otherResourceInfo.ResourceName),
		},
	)

	accessLevel, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.ssm.%s.accessLevel", otherResourceInfo.ResourceName),
			Default: "read",
		},
	)

	return &parameterLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
		accessLevel:     accessLevel,
	}
}

// Shared by the create and destroy paths so a teardown removes exactly what the create
// path provisioned.
//
// Destroy used to return before reaching the activation, so the VPC endpoint and its
// security group were left behind on every teardown. That group's ingress rule
// references the caller's group, which then blocks the caller's group, and with it the
// whole VPC, from being deleted.
func ssmParameterNetworkingActivation(
	setupCtx *linkutils.LambdaLinkSetupContext,
	region string,
) linkutils.NetworkingActivation {
	return linkutils.NetworkingActivation{
		Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
		Region:       region,
		AWSService:   "ssm",
		EndpointType: ec2types.VpcEndpointTypeInterface,
	}
}
