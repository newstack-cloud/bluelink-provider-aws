package lambdalinks

import (
	"context"
	"errors"
	"fmt"

	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// The target function's ARN is contributed to the invoking function's environment rather
// than written here, so neither resource is written by this link.
func (l *lambdaFunctionFunctionLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(
		input.LinkContext,
		"aws",
	)
	lambdaService, region, err := l.getLambdaServiceWithRegion(
		ctx,
		providerCtx,
	)
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

	// The execution role is shared by every link that grants it access, so lock it
	// for the read-modify-write of its policy set.
	err = input.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      pluginutils.GetInstanceID(input.ResourceAInfo),
			ResourceName:    setupCtx.RoleResourceName,
			ProviderContext: providerCtx,
			AcquiredBy:      input.LinkID,
		},
	)
	if err != nil {
		return nil, err
	}

	iamService, err := l.getIamService(
		ctx,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	sid := createLambdaFunctionInvokeSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		_, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
			RoleName: setupCtx.RoleName,
			SID:      sid,
		})
		if err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	otherFunctionARN, hasOtherFunctionARN := utils.ExtractARNFromResourceInfo(
		input.ResourceBInfo,
	)
	if !hasOtherFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked to %q function resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	result, err := linkutils.ReconcileRoleAccessPolicy(ctx, iamService, linkutils.RoleAccessGrant{
		RoleName:  setupCtx.RoleName,
		SID:       sid,
		Statement: lambdaInvokeAccessStatement(sid, otherFunctionARN),
	})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q permission to invoke Lambda %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	output := invokeAccessLinkOutput(input, setupCtx.RoleResourceName, sid, otherFunctionARN, result)

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	// A VPC-attached caller reaching another Lambda needs an interface VPC endpoint for
	// the Lambda service; ReconcileLinkNetworking is a no-op for non-VPC functions.
	return linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:       linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
			Region:       region,
			AWSService:   "lambda",
			EndpointType: ec2types.VpcEndpointTypeInterface,
		},
		output,
	)
}

// lambdaInvokeAccessStatement builds the IAM policy statement (canonical
// PascalCase, as the IAM API expects) granting permission to invoke the target
// function.
func lambdaInvokeAccessStatement(sid, otherFunctionARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   "lambda:InvokeFunction",
		"Resource": otherFunctionARN,
	}
}

// invokeAccessLinkOutput records the granted statement in link data and, for
// inline placements, maps it onto the role's spec so the framework attributes the
// statement to this link and does not treat it as drift / strip it on redeploy.
func invokeAccessLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, otherFunctionARN string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specInvokeStatementNode(sid, otherFunctionARN),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it:
	// inline placements map the statement by Sid; managed (overflow) placements map
	// the attached managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

// specInvokeStatementNode builds the statement in the camelCase spec form the
// role's external state uses (after Cloud Control name translation), so the drift
// comparison against link data matches.
func specInvokeStatementNode(sid, otherFunctionARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", core.MappingNodeFromString("lambda:InvokeFunction"),
		"resource", core.MappingNodeFromString(otherFunctionARN),
	)
}

func invokeLambdaFunctionEnvVarName(
	userDefinedEnvVarName string,
	resourceInfo *provider.ResourceInfo,
) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}

	return fmt.Sprintf(
		"INVOKE_LAMBDA_FUNCTION_%s",
		resourceInfo.ResourceName,
	)
}

func createLambdaFunctionInvokeSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"InvokeLambdaFunction%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

type lambdaFunctionLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
}

func getLambdaFunctionLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *lambdaFunctionLinkAnnotations {
	// Intra-service link annotations use aws.lambda.invoke.* pattern
	// where "invoke" is the feature being configured
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key: fmt.Sprintf(
				"aws.lambda.invoke.%s.populateEnvVars",
				otherResourceInfo.ResourceName,
			),
			FallbackKey: "aws.lambda.invoke.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf(
				"aws.lambda.invoke.%s.envVarName",
				otherResourceInfo.ResourceName,
			),
		},
	)

	return &lambdaFunctionLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
	}
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"%sExecutionRole",
		resourceInfo.ResourceName,
	)
}
