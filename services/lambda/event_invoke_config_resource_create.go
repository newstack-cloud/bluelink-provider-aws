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

func (l *lambdaEventInvokeConfigResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[Service]{
		&eventInvokeConfigCreate{},
	}

	hasSavedValues, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			Data: map[string]any{},
		},
		createOperations,
		input,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	if !hasSavedValues {
		return nil, fmt.Errorf("no values were saved during event invoke config creation")
	}

	putEventInvokeConfigOutput, ok := saveOpCtx.Data["putEventInvokeConfigOutput"]
	if !ok {
		return nil, fmt.Errorf("putEventInvokeConfigOutput not found in save operation context")
	}

	putEventInvokeConfigOutputTyped, ok := putEventInvokeConfigOutput.(*lambda.PutFunctionEventInvokeConfigOutput)
	if !ok {
		return nil, fmt.Errorf("putEventInvokeConfigOutput is not of type *lambda.PutFunctionEventInvokeConfigOutput")
	}

	functionArn := aws.ToString(putEventInvokeConfigOutputTyped.FunctionArn)
	computedValues := map[string]*core.MappingNode{
		"spec.functionArn": core.MappingNodeFromString(functionArn),
	}

	if putEventInvokeConfigOutputTyped.LastModified != nil {
		computedValues["spec.lastModified"] = core.MappingNodeFromString(putEventInvokeConfigOutputTyped.LastModified.String())
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedValues,
	}, nil
}
