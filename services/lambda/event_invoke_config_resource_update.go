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

func (l *lambdaEventInvokeConfigResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[Service]{
		&eventInvokeConfigUpdate{},
	}

	hasSavedValues, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			Data: map[string]any{},
		},
		updateOperations,
		input,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	if !hasSavedValues {
		return nil, fmt.Errorf("no values were updated during event invoke config update")
	}

	updateEventInvokeConfigOutput, ok := saveOpCtx.Data["updateEventInvokeConfigOutput"]
	if !ok {
		return nil, fmt.Errorf("updateEventInvokeConfigOutput not found in save operation context")
	}

	updateEventInvokeConfigOutputTyped, ok := updateEventInvokeConfigOutput.(*lambda.UpdateFunctionEventInvokeConfigOutput)
	if !ok {
		return nil, fmt.Errorf("updateEventInvokeConfigOutput is not of type *lambda.UpdateFunctionEventInvokeConfigOutput")
	}

	functionArn := aws.ToString(updateEventInvokeConfigOutputTyped.FunctionArn)
	computedValues := map[string]*core.MappingNode{
		"spec.functionArn": core.MappingNodeFromString(functionArn),
	}

	if updateEventInvokeConfigOutputTyped.LastModified != nil {
		computedValues["spec.lastModified"] = core.MappingNodeFromString(updateEventInvokeConfigOutputTyped.LastModified.String())
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedValues,
	}, nil
}
