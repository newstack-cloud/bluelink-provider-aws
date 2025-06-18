package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaEventSourceMappingResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[Service]{
		&eventSourceMappingUpdate{},
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
		return nil, fmt.Errorf("no values were saved during event source mapping update")
	}

	id, ok := saveOpCtx.Data["id"]
	if !ok {
		return nil, fmt.Errorf("id not found in save operation context")
	}

	idString, ok := id.(string)
	if !ok {
		return nil, fmt.Errorf("id is not a string")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.id": core.MappingNodeFromString(idString),
	}

	if updateOutput, ok := saveOpCtx.Data["updateEventSourceMappingOutput"]; ok {
		if updateOutputTyped, ok := updateOutput.(*lambda.UpdateEventSourceMappingOutput); ok {
			computedFields["spec.eventSourceMappingArn"] = core.MappingNodeFromString(aws.ToString(updateOutputTyped.EventSourceMappingArn))
			computedFields["spec.functionArn"] = core.MappingNodeFromString(aws.ToString(updateOutputTyped.FunctionArn))
			computedFields["spec.state"] = core.MappingNodeFromString(aws.ToString(updateOutputTyped.State))
		}
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
