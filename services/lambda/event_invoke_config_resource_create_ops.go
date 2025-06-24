package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type eventInvokeConfigCreate struct {
	input *lambda.PutFunctionEventInvokeConfigInput
}

func (e *eventInvokeConfigCreate) Name() string {
	return "create event invoke config"
}

func (e *eventInvokeConfigCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToPutEventInvokeConfigInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	e.input = input
	return hasValues, saveOpCtx, nil
}

func (e *eventInvokeConfigCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	putEventInvokeConfigOutput, err := lambdaService.PutFunctionEventInvokeConfig(ctx, e.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(putEventInvokeConfigOutput.FunctionArn)
	newSaveOpCtx.Data["putEventInvokeConfigOutput"] = putEventInvokeConfigOutput
	newSaveOpCtx.Data["functionArn"] = aws.ToString(putEventInvokeConfigOutput.FunctionArn)
	if putEventInvokeConfigOutput.LastModified != nil {
		newSaveOpCtx.Data["lastModified"] = putEventInvokeConfigOutput.LastModified.String()
	}

	return newSaveOpCtx, nil
}

func changesToPutEventInvokeConfigInput(
	specData *core.MappingNode,
) (*lambda.PutFunctionEventInvokeConfigInput, bool, error) {
	input := &lambda.PutFunctionEventInvokeConfigInput{}

	functionName, hasFunctionName := pluginutils.GetValueByPath(
		"$.functionName",
		specData,
	)
	if !hasFunctionName {
		return nil, false, fmt.Errorf("functionName must be defined in the resource spec")
	}

	input.FunctionName = aws.String(core.StringValue(functionName))

	qualifier, hasQualifier := pluginutils.GetValueByPath(
		"$.qualifier",
		specData,
	)
	if !hasQualifier {
		return nil, false, fmt.Errorf("qualifier must be defined in the resource spec")
	}

	input.Qualifier = aws.String(core.StringValue(qualifier))

	valueSetters := []*pluginutils.ValueSetter[*lambda.PutFunctionEventInvokeConfigInput]{
		pluginutils.NewValueSetter(
			"$.maximumEventAgeInSeconds",
			func(value *core.MappingNode, input *lambda.PutFunctionEventInvokeConfigInput) {
				input.MaximumEventAgeInSeconds = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumRetryAttempts",
			func(value *core.MappingNode, input *lambda.PutFunctionEventInvokeConfigInput) {
				input.MaximumRetryAttempts = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.destinationConfig",
			func(value *core.MappingNode, input *lambda.PutFunctionEventInvokeConfigInput) {
				destinationConfig := &types.DestinationConfig{}

				if onFailure, exists := pluginutils.GetValueByPath("$.onFailure", value); exists {
					if destination, destExists := pluginutils.GetValueByPath("$.destination", onFailure); destExists {
						destinationConfig.OnFailure = &types.OnFailure{
							Destination: aws.String(core.StringValue(destination)),
						}
					}
				}

				if onSuccess, exists := pluginutils.GetValueByPath("$.onSuccess", value); exists {
					if destination, destExists := pluginutils.GetValueByPath("$.destination", onSuccess); destExists {
						destinationConfig.OnSuccess = &types.OnSuccess{
							Destination: aws.String(core.StringValue(destination)),
						}
					}
				}

				input.DestinationConfig = destinationConfig
			},
		),
	}

	hasValuesToSave := true // functionName and qualifier are always required
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
	}

	return input, hasValuesToSave, nil
}
