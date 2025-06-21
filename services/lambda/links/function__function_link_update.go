package lambdalinks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/celerity-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/blueprint/state"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionFunctionLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	lambdaService, err := l.getLambdaService(
		ctx,
		provider.NewProviderContextFromLinkContext(
			input.LinkContext,
			"aws",
		),
	)
	if err != nil {
		return nil, err
	}

	functionSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(
		input.ResourceInfo,
	)
	functionARN, hasFunctionARN := pluginutils.GetValueByPath(
		"$.arn",
		functionSpec,
	)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked from %q function resource",
			getResourceNameFromResourceInfo(input.ResourceInfo),
		)
	}

	otherFunctionSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(
		input.OtherResourceInfo,
	)
	otherFunctionARN, hasOtherFunctionARN := pluginutils.GetValueByPath(
		"$.arn",
		otherFunctionSpec,
	)
	if !hasOtherFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked to %q function resource",
			getResourceNameFromResourceInfo(input.OtherResourceInfo),
		)
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return l.removeFunctionPermissionsAndEnvVars(
			ctx,
			core.StringValue(functionARN),
			core.StringValue(otherFunctionARN),
			lambdaService,
		)
	}

	return l.addFunctionPermissionsAndEnvVars(
		ctx,
		input,
		core.StringValue(functionARN),
		core.StringValue(otherFunctionARN),
		lambdaService,
	)
}

func (l *lambdaFunctionFunctionLinkActions) addFunctionPermissionsAndEnvVars(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
	functionARN string,
	otherFunctionARN string,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	codeSigningConfigSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(
		input.OtherResourceInfo,
	)
	codeSigningConfigARN, hasCodeSigningConfigARN := pluginutils.GetValueByPath(
		"$.codeSigningConfigArn",
		codeSigningConfigSpec,
	)
	if !hasCodeSigningConfigARN {
		return nil, fmt.Errorf(
			"code signing config ARN could not be retrieved from code signing config",
		)
	}

	_, err := lambdaService.PutFunctionCodeSigningConfig(
		ctx,
		&lambda.PutFunctionCodeSigningConfigInput{
			FunctionName:         aws.String(functionARN),
			CodeSigningConfigArn: aws.String(core.StringValue(codeSigningConfigARN)),
		},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				input.ResourceInfo.ResourceName: {
					Fields: map[string]*core.MappingNode{
						"codeSigningConfigArn": core.MappingNodeFromString(
							core.StringValue(codeSigningConfigARN),
						),
					},
				},
			},
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) removeFunctionPermissionsAndEnvVars(
	ctx context.Context,
	functionARN string,
	linkedToFunctionARN string,
	lambdaService lambdaservice.Service,
) (*provider.LinkUpdateResourceOutput, error) {
	_, err := lambdaService.DeleteFunctionCodeSigningConfig(
		ctx,
		&lambda.DeleteFunctionCodeSigningConfigInput{
			FunctionName: aws.String(functionARN),
		},
	)
	if err != nil {
		return nil, err
	}

	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The target function is not updated as a part of the link update,
	// the link only requires updates to the linked from function
	// that will need to be able to invoke the target function.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *lambdaFunctionFunctionLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	// There are no intermediary resources to update
	// for the lambda function to code signing config link.
	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func getResourceNameFromResourceInfo(resourceInfo *provider.ResourceInfo) string {
	if resourceInfo == nil {
		return "unknown"
	}

	return resourceInfo.ResourceName
}
