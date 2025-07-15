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

func (i *iamSAMLProviderResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&samlProviderMetadataUpdate{},
		&samlProviderTagsUpdate{},
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

	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)

	if hasUpdates {
		arnValue, err := core.GetPathValue(
			"$.arn",
			currentStateSpecData,
			core.MappingNodeMaxTraverseDepth,
		)
		if err != nil {
			return nil, err
		}

		arn := core.StringValue(arnValue)
		if arn == "" {
			return nil, fmt.Errorf("ARN is required for update operation")
		}

		getSAMLProviderOutput, err := iamService.GetSAMLProvider(ctx, &iam.GetSAMLProviderInput{
			SAMLProviderArn: aws.String(arn),
		})
		if err != nil {
			return nil, err
		}

		return &provider.ResourceDeployOutput{
			ComputedFieldValues: i.extractComputedFieldsFromSAMLProvider(arn, getSAMLProviderOutput),
		}, nil
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: i.extractComputedFieldsFromCurrentState(currentStateSpecData),
	}, nil
}

func (i *iamSAMLProviderResourceActions) extractComputedFieldsFromSAMLProvider(
	arn string,
	getSAMLProviderOutput *iam.GetSAMLProviderOutput,
) map[string]*core.MappingNode {
	return map[string]*core.MappingNode{
		"spec.arn":              core.MappingNodeFromString(arn),
		"spec.samlProviderUUID": core.MappingNodeFromString(aws.ToString(getSAMLProviderOutput.SAMLProviderUUID)),
	}
}

func (i *iamSAMLProviderResourceActions) extractComputedFieldsFromCurrentState(
	currentStateSpecData *core.MappingNode,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if v, ok := pluginutils.GetValueByPath("$.arn", currentStateSpecData); ok {
		fields["spec.arn"] = v
	}
	if v, ok := pluginutils.GetValueByPath("$.samlProviderUUID", currentStateSpecData); ok {
		fields["spec.samlProviderUUID"] = v
	}
	return fields
}
