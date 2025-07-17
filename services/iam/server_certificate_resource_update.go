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

func (a *iamServerCertificateResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := a.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&serverCertificateUpdate{},
		&serverCertificateTagsUpdate{},
	}

	seedSaveOpCtx := pluginutils.SaveOperationContext{
		Data: map[string]any{
			"ResourceDeployInput": input,
		},
	}

	hasUpdates, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		seedSaveOpCtx,
		updateOperations,
		input,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	if hasUpdates {
		finalServerCertificateName, ok := saveOpCtx.Data["finalServerCertificateName"].(string)
		if !ok {
			return nil, fmt.Errorf("finalServerCertificateName is expected to be present in the save operation context")
		}

		// Re-fetch the server certificate to get the updated ARN
		// as when the server certificate name or path is updated, the ARN changes.
		getServerCertificateOutput, err := iamService.GetServerCertificate(ctx, &iam.GetServerCertificateInput{
			ServerCertificateName: aws.String(finalServerCertificateName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to get server certificate: %w", err)
		}

		return &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(
					aws.ToString(getServerCertificateOutput.ServerCertificate.ServerCertificateMetadata.Arn),
				),
			},
		}, nil
	}

	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)
	arn, hasARN := pluginutils.GetValueByPath(
		"$.arn",
		currentStateSpecData,
	)
	if !hasARN {
		return nil, fmt.Errorf("ARN is expected to be present in the current state for update operation")
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.arn": arn,
		},
	}, nil
}
