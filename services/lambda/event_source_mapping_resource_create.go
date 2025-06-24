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

func (l *lambdaEventSourceMappingResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&eventSourceMappingCreate{},
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
		return nil, fmt.Errorf("no values were saved during event source mapping creation")
	}

	createEventSourceMappingOutput, ok := saveOpCtx.Data["createEventSourceMappingOutput"]
	if !ok {
		return nil, fmt.Errorf("createEventSourceMappingOutput not found in save operation context")
	}

	createEventSourceMappingOutputTyped, ok := createEventSourceMappingOutput.(*lambda.CreateEventSourceMappingOutput)
	if !ok {
		return nil, fmt.Errorf("createEventSourceMappingOutput is not of type *lambda.CreateEventSourceMappingOutput")
	}

	uuid := aws.ToString(createEventSourceMappingOutputTyped.UUID)
	eventSourceMappingArn := aws.ToString(createEventSourceMappingOutputTyped.EventSourceMappingArn)
	functionArn := aws.ToString(createEventSourceMappingOutputTyped.FunctionArn)
	state := aws.ToString(createEventSourceMappingOutputTyped.State)

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.id":                    core.MappingNodeFromString(uuid),
			"spec.eventSourceMappingArn": core.MappingNodeFromString(eventSourceMappingArn),
			"spec.functionArn":           core.MappingNodeFromString(functionArn),
			"spec.state":                 core.MappingNodeFromString(state),
		},
	}, nil
}
