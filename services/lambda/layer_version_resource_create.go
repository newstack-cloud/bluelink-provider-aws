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

func (l *lambdaLayerVersionResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&layerVersionCreate{},
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
		return nil, fmt.Errorf("no values were saved during layer version creation")
	}

	publishLayerVersionOutput, ok := saveOpCtx.Data["publishLayerVersionOutput"].(*lambda.PublishLayerVersionOutput)
	if !ok {
		return nil, fmt.Errorf("publishLayerVersionOutput not found in save operation context")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.layerArn":        core.MappingNodeFromString(aws.ToString(publishLayerVersionOutput.LayerArn)),
		"spec.layerVersionArn": core.MappingNodeFromString(aws.ToString(publishLayerVersionOutput.LayerVersionArn)),
		"spec.version":         core.MappingNodeFromInt(int(publishLayerVersionOutput.Version)),
		"spec.createdDate":     core.MappingNodeFromString(aws.ToString(publishLayerVersionOutput.CreatedDate)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
