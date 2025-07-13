package iam

import (
	"context"
	"fmt"

	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamInstanceProfileResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&instanceProfileRoleUpdate{},
	}

	saveOpCtx := pluginutils.SaveOperationContext{
		Data: map[string]any{
			"ResourceDeployInput": input,
		},
	}

	hasUpdates, _, err := pluginutils.RunSaveOperations(
		ctx,
		saveOpCtx,
		updateOperations,
		input,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	if !hasUpdates {
		return nil, fmt.Errorf("no updates were made during instance profile update")
	}

	return &provider.ResourceDeployOutput{}, nil
}
