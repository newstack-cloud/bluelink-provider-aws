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

func (i *iamSAMLProviderResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	// Safely get the SAML provider ARN from the resource state
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.ResourceState.SpecData)
	if !hasArn {
		return fmt.Errorf("SAML provider ARN is required for destroy")
	}

	arnStr := core.StringValue(arn)
	if arnStr == "" {
		return fmt.Errorf("SAML provider ARN is required for destroy")
	}

	// Delete the SAML provider
	_, err = iamService.DeleteSAMLProvider(ctx, &iam.DeleteSAMLProviderInput{
		SAMLProviderArn: aws.String(arnStr),
	})
	if err != nil {
		return fmt.Errorf("failed to delete SAML provider: %w", err)
	}

	return nil
}
