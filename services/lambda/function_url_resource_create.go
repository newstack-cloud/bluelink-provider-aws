package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/celerity-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionUrlResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&functionUrlCreate{},
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
		return nil, fmt.Errorf("no values were saved during function URL creation")
	}

	createFunctionUrlOutput, ok := saveOpCtx.Data["createFunctionUrlOutput"]
	if !ok {
		return nil, fmt.Errorf("createFunctionUrlOutput not found in save operation context")
	}

	createFunctionUrlOutputTyped, ok := createFunctionUrlOutput.(*lambda.CreateFunctionUrlConfigOutput)
	if !ok {
		return nil, fmt.Errorf("createFunctionUrlOutput is not of type *lambda.CreateFunctionUrlConfigOutput")
	}

	functionUrl := aws.ToString(createFunctionUrlOutputTyped.FunctionUrl)
	functionArn := aws.ToString(createFunctionUrlOutputTyped.FunctionArn)

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.functionUrl": core.MappingNodeFromString(functionUrl),
			"spec.functionArn": core.MappingNodeFromString(functionArn),
		},
	}, nil
}
