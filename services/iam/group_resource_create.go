package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamGroupResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[iamservice.Service]{
		newGroupCreate(i.uniqueNameGenerator),
		&groupInlinePoliciesCreate{},
		&groupManagedPoliciesCreate{},
	}

	saveOpCtx := pluginutils.SaveOperationContext{
		Data: map[string]any{
			"ResourceDeployInput": input,
		},
	}

	hasUpdates, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		saveOpCtx,
		createOperations,
		input,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	if !hasUpdates {
		return nil, fmt.Errorf("no updates were made during group creation")
	}

	createGroupOutput, ok := saveOpCtx.Data["createGroupOutput"].(*iam.CreateGroupOutput)
	if !ok {
		return nil, fmt.Errorf("createGroupOutput not found in save operation context")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn":     core.MappingNodeFromString(aws.ToString(createGroupOutput.Group.Arn)),
		"spec.groupId": core.MappingNodeFromString(aws.ToString(createGroupOutput.Group.GroupId)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
