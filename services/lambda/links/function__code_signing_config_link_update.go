package lambdalinks

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/blueprint/state"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionCodeSigningConfigLinkActions) UpdateResourceA(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	codeSigningConfigSpec := pluginutils.GetSpecDataFromResourceInfo(
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

	functionSpec := pluginutils.GetSpecDataFromResourceInfo(
		input.ResourceInfo,
	)
	functionARN, hasFunctionARN := pluginutils.GetValueByPath(
		"$.arn",
		functionSpec,
	)
	if !hasFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from function",
		)
	}

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

	_, err = lambdaService.PutFunctionCodeSigningConfig(
		ctx,
		&lambda.PutFunctionCodeSigningConfigInput{
			FunctionName:         aws.String(core.StringValue(functionARN)),
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

func (l *lambdaFunctionCodeSigningConfigLinkActions) UpdateResourceB(
	ctx context.Context,
	input *provider.LinkUpdateResourceInput,
) (*provider.LinkUpdateResourceOutput, error) {
	// The code signing config is not updated as a part of the link update.
	return &provider.LinkUpdateResourceOutput{
		LinkData: &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		},
	}, nil
}

func (l *lambdaFunctionCodeSigningConfigLinkActions) UpdateIntermediaryResources(
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
