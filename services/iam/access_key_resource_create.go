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

func (i *iamAccessKeyResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&accessKeyCreate{},
		&accessKeyStatusUpdate{},
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
		return nil, fmt.Errorf("no updates were made during access key creation")
	}

	createAccessKeyOutput, ok := saveOpCtx.Data["createAccessKeyOutput"].(*iam.CreateAccessKeyOutput)
	if !ok {
		return nil, fmt.Errorf("createAccessKeyOutput not found in save operation context")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.id":              core.MappingNodeFromString(aws.ToString(createAccessKeyOutput.AccessKey.AccessKeyId)),
		"spec.secretAccessKey": core.MappingNodeFromString(aws.ToString(createAccessKeyOutput.AccessKey.SecretAccessKey)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
