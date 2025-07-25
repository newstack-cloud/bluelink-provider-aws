package iam

import (
	"context"

	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamOIDCProviderResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&oidcProviderClientIdsUpdate{},
		&oidcProviderThumbprintsUpdate{},
		&oidcProviderTagsUpdate{},
	}

	saveOpCtx := pluginutils.SaveOperationContext{
		Data: map[string]any{
			"ResourceDeployInput": input,
		},
	}

	_, _, err = pluginutils.RunSaveOperations(
		ctx,
		saveOpCtx,
		updateOperations,
		input,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: i.extractComputedFieldsFromCurrentState(currentStateSpecData),
	}, nil
}

func (i *iamOIDCProviderResourceActions) extractComputedFieldsFromCurrentState(
	currentStateSpecData *core.MappingNode,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if v, ok := pluginutils.GetValueByPath("$.arn", currentStateSpecData); ok {
		fields["spec.arn"] = v
	}
	return fields
}
