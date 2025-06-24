package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
)

func (l *lambdaCodeSigningConfigResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&codeSigningConfigCreate{},
		&tagsUpdate{pathRoot: "$.tags"},
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
		return nil, fmt.Errorf("no values were saved during code signing config creation")
	}

	createCodeSigningConfigOutput, ok := saveOpCtx.Data["createCodeSigningConfigOutput"]
	if !ok {
		return nil, fmt.Errorf("createCodeSigningConfigOutput not found in save operation context")
	}

	createCodeSigningConfigOutputTyped, ok := createCodeSigningConfigOutput.(*lambda.CreateCodeSigningConfigOutput)
	if !ok {
		return nil, fmt.Errorf("createCodeSigningConfigOutput is not of type *lambda.CreateCodeSigningConfigOutput")
	}

	codeSigningConfigArn := aws.ToString(createCodeSigningConfigOutputTyped.CodeSigningConfig.CodeSigningConfigArn)
	codeSigningConfigId := aws.ToString(createCodeSigningConfigOutputTyped.CodeSigningConfig.CodeSigningConfigId)

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.codeSigningConfigArn": core.MappingNodeFromString(codeSigningConfigArn),
			"spec.codeSigningConfigId":  core.MappingNodeFromString(codeSigningConfigId),
		},
	}, nil
}
