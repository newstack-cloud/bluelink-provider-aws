package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
)

type functionUrlCreate struct {
	input *lambda.CreateFunctionUrlConfigInput
}

func (u *functionUrlCreate) Name() string {
	return "create function URL"
}

func (u *functionUrlCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToCreateFunctionUrlInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *functionUrlCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createFunctionUrlOutput, err := lambdaService.CreateFunctionUrlConfig(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(createFunctionUrlOutput.FunctionUrl)
	newSaveOpCtx.Data["createFunctionUrlOutput"] = createFunctionUrlOutput
	newSaveOpCtx.Data["functionUrl"] = aws.ToString(createFunctionUrlOutput.FunctionUrl)
	newSaveOpCtx.Data["functionArn"] = aws.ToString(createFunctionUrlOutput.FunctionArn)

	return newSaveOpCtx, nil
}

func changesToCreateFunctionUrlInput(
	specData *core.MappingNode,
) (*lambda.CreateFunctionUrlConfigInput, bool, error) {
	input := &lambda.CreateFunctionUrlConfigInput{}

	functionARN, hasFunctionARN := pluginutils.GetValueByPath(
		"$.functionArn",
		specData,
	)
	if !hasFunctionARN {
		return nil, false, fmt.Errorf("functionArn must be defined in the resource spec")
	}

	input.FunctionName = aws.String(core.StringValue(functionARN))

	valueSetters := []*pluginutils.ValueSetter[*lambda.CreateFunctionUrlConfigInput]{
		pluginutils.NewValueSetter(
			"$.authType",
			func(value *core.MappingNode, input *lambda.CreateFunctionUrlConfigInput) {
				input.AuthType = types.FunctionUrlAuthType(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.qualifier",
			func(value *core.MappingNode, input *lambda.CreateFunctionUrlConfigInput) {
				input.Qualifier = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.invokeMode",
			func(value *core.MappingNode, input *lambda.CreateFunctionUrlConfigInput) {
				input.InvokeMode = types.InvokeMode(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.cors",
			func(value *core.MappingNode, input *lambda.CreateFunctionUrlConfigInput) {
				setCorsValue(value, func(cors *types.Cors) {
					input.Cors = cors
				})
			},
		),
	}

	hasValuesToSave := true // functionName is always required
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
	}

	return input, hasValuesToSave, nil
}
