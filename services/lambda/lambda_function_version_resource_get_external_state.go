package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
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

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionArn": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionArn),
			),
			"functionArnWithVersion": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionArn) + ":" +
					aws.ToString(functionOutput.Configuration.Version),
			),
			"version": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.Version),
			),
			"functionName": core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.FunctionName),
			),
		},
	}

	if functionOutput.Configuration.Description != nil {
		resourceSpecState.Fields["description"] = core.MappingNodeFromString(
			aws.ToString(functionOutput.Configuration.Description),
		)
	}

	// Provisioned concurrency configuration is not a part of the core function config,
	// so we need to request it separately.
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

	if functionOutput.Configuration.RuntimeVersionConfig != nil {
		runtimePolicy := &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}

		if functionOutput.Configuration.RuntimeVersionConfig.RuntimeVersionArn != nil {
			runtimePolicy.Fields["runtimeVersionArn"] = core.MappingNodeFromString(
				aws.ToString(functionOutput.Configuration.RuntimeVersionConfig.RuntimeVersionArn),
			)
		}

		// The updateRuntimeOn field is not persisted in AWS, so we get it from the input spec
		inputRuntimePolicy, hasRuntimePolicy := input.CurrentResourceSpec.Fields["runtimePolicy"]
		if hasRuntimePolicy {
			if updateRuntimeOn, ok := inputRuntimePolicy.Fields["updateRuntimeOn"]; ok {
				runtimePolicy.Fields["updateRuntimeOn"] = updateRuntimeOn
			}
		}

		resourceSpecState.Fields["runtimePolicy"] = runtimePolicy
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}
