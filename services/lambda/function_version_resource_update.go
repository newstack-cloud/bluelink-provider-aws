package lambda

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
)

func (l *lambdaFunctionVersionResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// function ARN and version allow us to update configuration associated
	// with a specific function version.
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)
	functionARNWithVersionValue, err := core.GetPathValue(
		"$.functionArnWithVersion",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, err
	}
	functionARNWithVersion := core.StringValue(functionARNWithVersionValue)

	functionARN, err := core.GetPathValue(
		"$.functionArn",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, err
	}

	version, err := core.GetPathValue(
		"$.version",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&functionVersionPutProvisionedConcurrencyConfig{},
		&functionRuntimeManagementConfigUpdate{
			path:                 "$.runtimePolicy",
			fieldChangesPathRoot: "spec.runtimePolicy",
		},
	}

	hasUpdates, _, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			ProviderUpstreamID: functionARNWithVersion,
			Data: map[string]any{
				"functionARN": core.StringValue(functionARN),
				"version":     core.StringValue(version),
			},
		},
		updateOperations,
		input,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	if hasUpdates {
		arnWithVersionParts := strings.Split(functionARNWithVersion, ":")
		getFunctionOutput, err := lambdaService.GetFunction(ctx, &lambda.GetFunctionInput{
			FunctionName: &arnWithVersionParts[0],
			Qualifier:    &arnWithVersionParts[1],
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get updated function configuration: %w", err)
		}

		computedFields := l.extractComputedFieldsFromFunctionConfig(
			getFunctionOutput.Configuration,
		)
		return &provider.ResourceDeployOutput{
			ComputedFieldValues: computedFields,
		}, nil
	}

	// If no updates were made, return the current computed fields from the current state.
	currentStateComputedFields := l.extractComputedFieldsFromCurrentState(
		currentStateSpecData,
	)
	return &provider.ResourceDeployOutput{
		ComputedFieldValues: currentStateComputedFields,
	}, nil
}

func (l *lambdaFunctionVersionResourceActions) extractComputedFieldsFromCurrentState(
	currentStateSpecData *core.MappingNode,
) map[string]*core.MappingNode {
	computedFields := map[string]*core.MappingNode{}

	functionArn, _ := pluginutils.GetValueByPath(
		"$.functionArn",
		currentStateSpecData,
	)
	computedFields["spec.functionArn"] = functionArn

	version, _ := pluginutils.GetValueByPath(
		"$.version",
		currentStateSpecData,
	)
	computedFields["spec.version"] = version

	functionArnWithVersion, _ := pluginutils.GetValueByPath(
		"$.functionArnWithVersion",
		currentStateSpecData,
	)
	computedFields["spec.functionArnWithVersion"] = functionArnWithVersion

	return computedFields
}

func (l *lambdaFunctionVersionResourceActions) extractComputedFieldsFromFunctionConfig(
	functionConfig *types.FunctionConfiguration,
) map[string]*core.MappingNode {
	return map[string]*core.MappingNode{
		"spec.functionArn": core.MappingNodeFromString(
			aws.ToString(functionConfig.FunctionArn),
		),
		"spec.version": core.MappingNodeFromString(
			aws.ToString(functionConfig.Version),
		),
		"spec.functionArnWithVersion": core.MappingNodeFromString(
			aws.ToString(functionConfig.FunctionArn) +
				":" + aws.ToString(functionConfig.Version),
		),
	}
}
