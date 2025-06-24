package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaAliasResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&aliasCreate{},
		&aliasPutProvisionedConcurrencyConfig{},
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
		return nil, fmt.Errorf("no values were saved during alias creation")
	}

	createAliasOutput, ok := saveOpCtx.Data["createAliasOutput"]
	if !ok {
		return nil, fmt.Errorf("createAliasOutput not found in save operation context")
	}

	createAliasOutputTyped, ok := createAliasOutput.(*lambda.CreateAliasOutput)
	if !ok {
		return nil, fmt.Errorf("createAliasOutput is not of type *lambda.CreateAliasOutput")
	}

	aliasArn := aws.ToString(createAliasOutputTyped.AliasArn)

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.aliasArn": core.MappingNodeFromString(aliasArn),
		},
	}, nil
}
