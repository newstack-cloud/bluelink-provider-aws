package lambdasns

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

// The topic's ARN is read from state when the publish grant is reconciled, so neither
// resource is written by this link.
func (l *functionTopicLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *functionTopicLinkActions) UpdateIntermediaryResources(
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
		&linkutils.LambdaLinkSetupData{LambdaFuncResourceInfo: input.ResourceAInfo, LoadRoleInfo: true},
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

	sid := createPublishSID(input.ResourceBInfo)

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
			snsNetworkingActivation(setupCtx, region),
			&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
		)
	}

	topicARN, hasTopicARN := extractTopicARN(input.ResourceBInfo)
	if !hasTopicARN {
		return nil, fmt.Errorf(
			"topic ARN could not be retrieved from the linked to %q topic resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
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
			Statement: snsPublishStatement(sid, topicARN),
		})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q permission to publish to SNS topic %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	output := publishAccessLinkOutput(input, setupCtx.RoleResourceName, sid, topicARN, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-isolated caller needs an interface VPC endpoint for SNS; this is a no-op for
	// non-VPC functions.
	return linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		snsNetworkingActivation(setupCtx, region),
		output,
	)
}

func snsPublishStatement(sid, topicARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   "sns:Publish",
		"Resource": topicARN,
	}
}

func publishAccessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, topicARN string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specPublishStatementNode(sid, topicARN),
	)

	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

func specPublishStatementNode(sid, topicARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", core.MappingNodeFromString("sns:Publish"),
		"resource", core.MappingNodeFromString(topicARN),
	)
}

func extractTopicARN(topicInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(topicInfo)
	topicARNNode, has := pluginutils.GetValueByPath("$.topicArn", spec)
	if !has {
		return "", false
	}
	return core.StringValue(topicARNNode), true
}

func topicEnvVarName(userDefinedEnvVarName string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("SNS_TOPIC_%s", resourceInfo.ResourceName)
}

func createPublishSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("SNSPublish%s", pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

type topicLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
}

func getTopicLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *topicLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.sns.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.sns.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.sns.%s.envVarName", otherResourceInfo.ResourceName),
		},
	)

	return &topicLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
	}
}

// Shared by the create and destroy paths so a teardown removes exactly what the create
// path provisioned.
//
// Destroy used to return before reaching the activation, so the VPC endpoint and its
// security group were left behind on every teardown. That group's ingress rule
// references the caller's group, which then blocks the caller's group, and with it the
// whole VPC, from being deleted.
func snsNetworkingActivation(
	setupCtx *linkutils.LambdaLinkSetupContext,
	region string,
) linkutils.NetworkingActivation {
	return linkutils.NetworkingActivation{
		Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
		Region:       region,
		AWSService:   "sns",
		EndpointType: ec2types.VpcEndpointTypeInterface,
	}
}
