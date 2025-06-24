package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// GetExternalState retrieves the current state of the Lambda function version from AWS.
func (l *lambdaFunctionVersionResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	functionARN := core.StringValue(
		input.CurrentResourceSpec.Fields["arn"],
	)

	functionOutput, err := lambdaService.GetFunction(
		ctx,
		&lambda.GetFunctionInput{
			FunctionName: &functionARN,
		},
	)
	if err != nil {
		return nil, err
	}

	// Build resource spec state from AWS response
	resourceSpecState := l.buildBaseResourceSpecState(functionOutput)

	// Add optional fields if they exist
	err = l.addOptionalConfigurationsToSpec(functionOutput, input, resourceSpecState.Fields)
	if err != nil {
		return nil, err
	}

	// Get provisioned concurrency config
	provisionedConcurrencyOutput, err := lambdaService.GetProvisionedConcurrencyConfig(
		ctx,
		&lambda.GetProvisionedConcurrencyConfigInput{
			FunctionName: aws.String(aws.ToString(functionOutput.Configuration.FunctionName)),
			Qualifier:    aws.String(aws.ToString(functionOutput.Configuration.Version)),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get provisioned concurrency config: %w", err)
	}

	// Add provisioned concurrency config if present
	if provisionedConcurrencyOutput.RequestedProvisionedConcurrentExecutions != nil {
		resourceSpecState.Fields["provisionedConcurrencyConfig"] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"provisionedConcurrentExecutions": core.MappingNodeFromInt(
					int(
						aws.ToInt32(
							provisionedConcurrencyOutput.RequestedProvisionedConcurrentExecutions,
						),
					),
				),
			},
		}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaFunctionVersionResourceActions) buildBaseResourceSpecState(
	output *lambda.GetFunctionOutput,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(
				aws.ToString(output.Configuration.FunctionArn),
			),
			"functionArnWithVersion": core.MappingNodeFromString(
				aws.ToString(output.Configuration.FunctionArn) + ":" +
					aws.ToString(output.Configuration.Version),
			),
			"version": core.MappingNodeFromString(
				aws.ToString(output.Configuration.Version),
			),
			"functionName": core.MappingNodeFromString(
				aws.ToString(output.Configuration.FunctionName),
			),
		},
	}
}

func (l *lambdaFunctionVersionResourceActions) addOptionalConfigurationsToSpec(
	output *lambda.GetFunctionOutput,
	input *provider.ResourceGetExternalStateInput,
	specFields map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetFunctionOutput]{
		{
			Name: "description",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.Description != nil
			},
			Fields: []string{"description"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.Configuration.Description)),
				}, nil
			},
		},
		{
			Name: "runtimePolicy",
			Condition: func(output *lambda.GetFunctionOutput) bool {
				return output.Configuration.RuntimeVersionConfig != nil
			},
			Fields: []string{"runtimePolicy"},
			Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
				runtimePolicy := &core.MappingNode{
					Fields: map[string]*core.MappingNode{},
				}

				if output.Configuration.RuntimeVersionConfig.RuntimeVersionArn != nil {
					runtimePolicy.Fields["runtimeVersionArn"] = core.MappingNodeFromString(
						aws.ToString(output.Configuration.RuntimeVersionConfig.RuntimeVersionArn),
					)
				}

				// The updateRuntimeOn field is not persisted in AWS, so we get it from the input spec
				inputRuntimePolicy, hasRuntimePolicy := input.CurrentResourceSpec.Fields["runtimePolicy"]
				if hasRuntimePolicy {
					if updateRuntimeOn, ok := inputRuntimePolicy.Fields["updateRuntimeOn"]; ok {
						runtimePolicy.Fields["updateRuntimeOn"] = updateRuntimeOn
					}
				}

				return []*core.MappingNode{runtimePolicy}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(output, specFields, extractors)
}
