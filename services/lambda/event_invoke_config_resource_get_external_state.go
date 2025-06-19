package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaEventInvokeConfigResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	functionName, hasFunctionName := pluginutils.GetValueByPath(
		"$.functionName",
		input.CurrentResourceSpec,
	)
	if !hasFunctionName {
		return nil, fmt.Errorf("functionName must be defined in the resource spec")
	}

	qualifier, hasQualifier := pluginutils.GetValueByPath(
		"$.qualifier",
		input.CurrentResourceSpec,
	)
	if !hasQualifier {
		return nil, fmt.Errorf("qualifier must be defined in the resource spec")
	}

	getEventInvokeConfigInput := &lambda.GetFunctionEventInvokeConfigInput{
		FunctionName: aws.String(core.StringValue(functionName)),
		Qualifier:    aws.String(core.StringValue(qualifier)),
	}

	result, err := lambdaService.GetFunctionEventInvokeConfig(ctx, getEventInvokeConfigInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get event invoke config: %w", err)
	}

	// Build resource spec state from AWS response
	resourceSpecState := l.buildBaseResourceSpecState(result)

	// Add optional fields if they exist
	err = l.addOptionalConfigurationsToSpec(result, resourceSpecState.Fields)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaEventInvokeConfigResourceActions) buildBaseResourceSpecState(
	output *lambda.GetFunctionEventInvokeConfigOutput,
) *core.MappingNode {
	fields := map[string]*core.MappingNode{
		"functionArn": core.MappingNodeFromString(aws.ToString(output.FunctionArn)),
	}

	if output.LastModified != nil {
		fields["lastModified"] = core.MappingNodeFromString(output.LastModified.String())
	}

	return &core.MappingNode{
		Fields: fields,
	}
}

func (l *lambdaEventInvokeConfigResourceActions) addOptionalConfigurationsToSpec(
	output *lambda.GetFunctionEventInvokeConfigOutput,
	specFields map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetFunctionEventInvokeConfigOutput]{
		{
			Name: "maximumEventAgeInSeconds",
			Condition: func(output *lambda.GetFunctionEventInvokeConfigOutput) bool {
				return output.MaximumEventAgeInSeconds != nil
			},
			Fields: []string{"maximumEventAgeInSeconds"},
			Values: func(output *lambda.GetFunctionEventInvokeConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.MaximumEventAgeInSeconds))),
				}, nil
			},
		},
		{
			Name: "maximumRetryAttempts",
			Condition: func(output *lambda.GetFunctionEventInvokeConfigOutput) bool {
				return output.MaximumRetryAttempts != nil
			},
			Fields: []string{"maximumRetryAttempts"},
			Values: func(output *lambda.GetFunctionEventInvokeConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromInt(int(aws.ToInt32(output.MaximumRetryAttempts))),
				}, nil
			},
		},
		{
			Name: "destinationConfig",
			Condition: func(output *lambda.GetFunctionEventInvokeConfigOutput) bool {
				return output.DestinationConfig != nil &&
					(output.DestinationConfig.OnFailure != nil || output.DestinationConfig.OnSuccess != nil)
			},
			Fields: []string{"destinationConfig"},
			Values: func(output *lambda.GetFunctionEventInvokeConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					destinationConfigToMappingNode(output.DestinationConfig),
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(
		output,
		specFields,
		extractors,
	)
}

func destinationConfigToMappingNode(
	destinationConfig *types.DestinationConfig,
) *core.MappingNode {
	if destinationConfig == nil {
		return &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}

	fields := map[string]*core.MappingNode{}

	if destinationConfig.OnFailure != nil && destinationConfig.OnFailure.Destination != nil {
		fields["onFailure"] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"destination": core.MappingNodeFromString(aws.ToString(destinationConfig.OnFailure.Destination)),
			},
		}
	}

	if destinationConfig.OnSuccess != nil && destinationConfig.OnSuccess.Destination != nil {
		fields["onSuccess"] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"destination": core.MappingNodeFromString(aws.ToString(destinationConfig.OnSuccess.Destination)),
			},
		}
	}

	return &core.MappingNode{Fields: fields}
}
