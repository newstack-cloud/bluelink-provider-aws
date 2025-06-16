package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

type functionVersionCreate struct {
	input *lambda.PublishVersionInput
}

func (u *functionVersionCreate) Name() string {
	return "create function version"
}

func (u *functionVersionCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToPublishVersionInput(
		specData,
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *functionVersionCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	publishVersionOutput, err := lambdaService.PublishVersion(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(
		publishVersionOutput.FunctionArn,
	) + ":" + aws.ToString(publishVersionOutput.Version)
	newSaveOpCtx.Data["publishVersionOutput"] = publishVersionOutput
	newSaveOpCtx.Data["functionARN"] = aws.ToString(publishVersionOutput.FunctionArn)
	newSaveOpCtx.Data["version"] = aws.ToString(publishVersionOutput.Version)

	return newSaveOpCtx, err
}

type functionVersionPutProvisionedConcurrencyConfig struct {
	input *lambda.PutProvisionedConcurrencyConfigInput
}

func (u *functionVersionPutProvisionedConcurrencyConfig) Name() string {
	return "put provisioned concurrency config"
}

func (u *functionVersionPutProvisionedConcurrencyConfig) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	provisionedConcurrencyConfigData, _ := pluginutils.GetValueByPath(
		"$.provisionedConcurrencyConfig",
		specData,
	)
	functionARN, version := extractFunctionARNAndVersion(saveOpCtx)
	input, hasUpdates := changesToPutProvisionedConcurrencyConfigInput(
		functionARN,
		version,
		provisionedConcurrencyConfigData,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func extractFunctionARNAndVersion(saveOpCtx pluginutils.SaveOperationContext) (string, string) {
	var functionARN, version string
	if arn, ok := saveOpCtx.Data["functionARN"]; ok && arn != nil {
		if arn, ok := arn.(string); ok {
			functionARN = arn
		}
	}
	if ver, ok := saveOpCtx.Data["version"]; ok && ver != nil {
		if versionStr, ok := ver.(string); ok {
			version = versionStr
		}
	}
	return functionARN, version
}

func (u *functionVersionPutProvisionedConcurrencyConfig) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.PutProvisionedConcurrencyConfig(ctx, u.input)
	return saveOpCtx, err
}
