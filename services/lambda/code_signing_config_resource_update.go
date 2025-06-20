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

func (l *lambdaCodeSigningConfigResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&codeSigningConfigUpdate{},
		&tagsUpdate{pathRoot: "$.tags"},
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
		return nil, fmt.Errorf("no values were saved during code signing config update")
	}

	codeSigningConfigArn, ok := saveOpCtx.Data["codeSigningConfigArn"]
	if !ok {
		return nil, fmt.Errorf("codeSigningConfigArn not found in save operation context")
	}

	codeSigningConfigArnString, ok := codeSigningConfigArn.(string)
	if !ok {
		return nil, fmt.Errorf("codeSigningConfigArn is not a string")
	}

	// If we have an update output, extract computed fields
	computedFields := map[string]*core.MappingNode{
		"spec.codeSigningConfigArn": core.MappingNodeFromString(codeSigningConfigArnString),
	}

	if updateOutput, ok := saveOpCtx.Data["updateCodeSigningConfigOutput"]; ok {
		if updateOutputTyped, ok := updateOutput.(*lambda.UpdateCodeSigningConfigOutput); ok {
			computedFields["spec.codeSigningConfigId"] = core.MappingNodeFromString(aws.ToString(updateOutputTyped.CodeSigningConfig.CodeSigningConfigId))
		}
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
