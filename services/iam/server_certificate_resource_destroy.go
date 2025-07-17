package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *iamServerCertificateResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	iamService, err := a.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	serverCertName, hasServerCertName := pluginutils.GetValueByPath(
		"$.serverCertificateName",
		input.ResourceState.SpecData,
	)
	if !hasServerCertName {
		return fmt.Errorf("server certificate name is required for destroy")
	}

	serverCertNameStr := core.StringValue(serverCertName)
	if serverCertNameStr == "" {
		return fmt.Errorf("server certificate name is required for destroy")
	}

	_, err = iamService.DeleteServerCertificate(ctx, &iam.DeleteServerCertificateInput{
		ServerCertificateName: aws.String(serverCertNameStr),
	})
	if err != nil {
		return fmt.Errorf("failed to delete server certificate: %w", err)
	}

	return nil
}
