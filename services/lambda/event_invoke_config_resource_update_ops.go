package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

type eventInvokeConfigUpdate struct {
	input *lambda.UpdateFunctionEventInvokeConfigInput
}

func (e *eventInvokeConfigUpdate) Name() string {
	return "update event invoke config"
}

func (e *eventInvokeConfigUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToUpdateEventInvokeConfigInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	e.input = input
	return hasValues, saveOpCtx, nil
}

func (e *eventInvokeConfigUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	updateEventInvokeConfigOutput, err := lambdaService.UpdateFunctionEventInvokeConfig(ctx, e.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(updateEventInvokeConfigOutput.FunctionArn)
	newSaveOpCtx.Data["updateEventInvokeConfigOutput"] = updateEventInvokeConfigOutput
	newSaveOpCtx.Data["functionArn"] = aws.ToString(updateEventInvokeConfigOutput.FunctionArn)
	if updateEventInvokeConfigOutput.LastModified != nil {
		newSaveOpCtx.Data["lastModified"] = updateEventInvokeConfigOutput.LastModified.String()
	}

	return newSaveOpCtx, nil
}

func changesToUpdateEventInvokeConfigInput(
	specData *core.MappingNode,
) (*lambda.UpdateFunctionEventInvokeConfigInput, bool, error) {
	input := &lambda.UpdateFunctionEventInvokeConfigInput{}

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

	valueSetters := []*pluginutils.ValueSetter[*lambda.UpdateFunctionEventInvokeConfigInput]{
		pluginutils.NewValueSetter(
			"$.maximumEventAgeInSeconds",
			func(value *core.MappingNode, input *lambda.UpdateFunctionEventInvokeConfigInput) {
				input.MaximumEventAgeInSeconds = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.maximumRetryAttempts",
			func(value *core.MappingNode, input *lambda.UpdateFunctionEventInvokeConfigInput) {
				input.MaximumRetryAttempts = aws.Int32(int32(core.IntValue(value)))
			},
		),
		pluginutils.NewValueSetter(
			"$.destinationConfig",
			func(value *core.MappingNode, input *lambda.UpdateFunctionEventInvokeConfigInput) {
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

	hasValuesToSave := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasValuesToSave = hasValuesToSave || valueSetter.DidSet()
	}

	return input, hasValuesToSave, nil
}
