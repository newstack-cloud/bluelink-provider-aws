package lambda

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaLayerVersionPermissionResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[Service]{
		&layerVersionPermissionCreate{},
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
		return nil, fmt.Errorf("no values were saved during layer version permission creation")
	}

	computedFields := map[string]*core.MappingNode{}
	if saveOpCtx.ProviderUpstreamID != "" {
		computedFields["id"] = core.MappingNodeFromString(saveOpCtx.ProviderUpstreamID)
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
