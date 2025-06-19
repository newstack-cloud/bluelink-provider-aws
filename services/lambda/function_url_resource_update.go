package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionUrlResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)
	functionARNValue, err := core.GetPathValue(
		"$.functionArn",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[Service]{
		&functionUrlUpdate{},
	}

	hasSavedValues, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			ProviderUpstreamID: core.StringValue(functionARNValue),
			Data: map[string]any{
				"functionArn": core.StringValue(functionARNValue),
			},
		},
		updateOperations,
		input,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	if !hasSavedValues {
		functionURLValue, _ := pluginutils.GetValueByPath(
			"$.functionUrl",
			currentStateSpecData,
		)
		return &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.functionUrl": functionURLValue,
				"spec.functionArn": functionARNValue,
			},
		}, nil
	}

	updateFunctionUrlOutput, ok := saveOpCtx.Data["updateFunctionUrlOutput"]
	if !ok {
		return nil, fmt.Errorf("updateFunctionUrlOutput not found in save operation context")
	}

	updateFunctionUrlOutputTyped, ok := updateFunctionUrlOutput.(*lambda.UpdateFunctionUrlConfigOutput)
	if !ok {
		return nil, fmt.Errorf("updateFunctionUrlOutput is not of type *lambda.UpdateFunctionUrlConfigOutput")
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.functionUrl": core.MappingNodeFromString(aws.ToString(updateFunctionUrlOutputTyped.FunctionUrl)),
			"spec.functionArn": core.MappingNodeFromString(aws.ToString(updateFunctionUrlOutputTyped.FunctionArn)),
		},
	}, nil
}
