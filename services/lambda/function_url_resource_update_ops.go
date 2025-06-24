package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type functionUrlUpdate struct {
	input *lambda.UpdateFunctionUrlConfigInput
}

func (u *functionUrlUpdate) Name() string {
	return "update function URL"
}

func (u *functionUrlUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToUpdateFunctionUrlInput(
		saveOpCtx.ProviderUpstreamID,
		specData,
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *functionUrlUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	updateFunctionUrlOutput, err := lambdaService.UpdateFunctionUrlConfig(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(updateFunctionUrlOutput.FunctionUrl)
	newSaveOpCtx.Data["updateFunctionUrlOutput"] = updateFunctionUrlOutput
	newSaveOpCtx.Data["functionUrl"] = aws.ToString(updateFunctionUrlOutput.FunctionUrl)
	newSaveOpCtx.Data["functionArn"] = aws.ToString(updateFunctionUrlOutput.FunctionArn)

	return newSaveOpCtx, nil
}

func changesToUpdateFunctionUrlInput(
	functionARN string,
	specData *core.MappingNode,
) (*lambda.UpdateFunctionUrlConfigInput, bool, error) {
	input := &lambda.UpdateFunctionUrlConfigInput{}

	input.FunctionName = aws.String(functionARN)

	valueSetters := []*pluginutils.ValueSetter[*lambda.UpdateFunctionUrlConfigInput]{
		pluginutils.NewValueSetter(
			"$.authType",
			func(value *core.MappingNode, input *lambda.UpdateFunctionUrlConfigInput) {
				input.AuthType = types.FunctionUrlAuthType(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.qualifier",
			func(value *core.MappingNode, input *lambda.UpdateFunctionUrlConfigInput) {
				input.Qualifier = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.invokeMode",
			func(value *core.MappingNode, input *lambda.UpdateFunctionUrlConfigInput) {
				input.InvokeMode = types.InvokeMode(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.cors",
			func(value *core.MappingNode, input *lambda.UpdateFunctionUrlConfigInput) {
				setCorsValue(value, func(cors *types.Cors) {
					input.Cors = cors
				})
			},
		),
	}

	hasValuesToSave := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasValuesToSave = hasValuesToSave || valueSetter.DidSet()
	}

	return input, hasValuesToSave, nil
}
