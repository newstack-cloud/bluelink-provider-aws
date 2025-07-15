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

func (i *iamSAMLProviderResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[iamservice.Service]{
		newSAMLProviderCreate(i.uniqueNameGenerator),
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
		return nil, fmt.Errorf("no updates were made during SAML provider creation")
	}

	createSAMLProviderOutput, hasCreateSAMLProviderOutput := saveOpCtx.Data["createSAMLProviderOutput"].(*iam.CreateSAMLProviderOutput)
	if !hasCreateSAMLProviderOutput {
		return nil, fmt.Errorf("createSAMLProviderOutput not found in save operation context")
	}

	getSAMLProviderOutput, err := iamService.GetSAMLProvider(ctx, &iam.GetSAMLProviderInput{
		SAMLProviderArn: createSAMLProviderOutput.SAMLProviderArn,
	})
	if err != nil {
		return nil, err
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn":              core.MappingNodeFromString(aws.ToString(createSAMLProviderOutput.SAMLProviderArn)),
		"spec.samlProviderUUID": core.MappingNodeFromString(aws.ToString(getSAMLProviderOutput.SAMLProviderUUID)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
