package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/celerity-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
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
				cors := &types.Cors{}
				if allowCredentials, exists := pluginutils.GetValueByPath("$.allowCredentials", value); exists {
					cors.AllowCredentials = aws.Bool(core.BoolValue(allowCredentials))
				}
				if allowHeaders, exists := pluginutils.GetValueByPath("$.allowHeaders", value); exists {
					headers := make([]string, len(allowHeaders.Items))
					for i, header := range allowHeaders.Items {
						headers[i] = core.StringValue(header)
					}
					cors.AllowHeaders = headers
				}
				if allowMethods, exists := pluginutils.GetValueByPath("$.allowMethods", value); exists {
					methods := make([]string, len(allowMethods.Items))
					for i, method := range allowMethods.Items {
						methods[i] = core.StringValue(method)
					}
					cors.AllowMethods = methods
				}
				if allowOrigins, exists := pluginutils.GetValueByPath("$.allowOrigins", value); exists {
					origins := make([]string, len(allowOrigins.Items))
					for i, origin := range allowOrigins.Items {
						origins[i] = core.StringValue(origin)
					}
					cors.AllowOrigins = origins
				}
				if exposeHeaders, exists := pluginutils.GetValueByPath("$.exposeHeaders", value); exists {
					headers := make([]string, len(exposeHeaders.Items))
					for i, header := range exposeHeaders.Items {
						headers[i] = core.StringValue(header)
					}
					cors.ExposeHeaders = headers
				}
				if maxAge, exists := pluginutils.GetValueByPath("$.maxAge", value); exists {
					cors.MaxAge = aws.Int32(int32(core.IntValue(maxAge)))
				}
				input.Cors = cors
			},
		),
	}

	hasValuesToSave := true // functionName is always required
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
	}

	return input, hasValuesToSave, nil
}
