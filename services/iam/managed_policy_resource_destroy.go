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

func (i *iamManagedPolicyResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	// Get the policy ARN from the resource state
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.ResourceState.SpecData)
	if !hasArn {
		return fmt.Errorf("ARN is required for destroy operation")
	}

	arnStr := core.StringValue(arn)
	if arnStr == "" {
		return fmt.Errorf("ARN cannot be empty for destroy operation")
	}

	// Delete the managed policy
	_, err = iamService.DeletePolicy(ctx, &iam.DeletePolicyInput{
		PolicyArn: aws.String(arnStr),
	})
	if err != nil {
		return fmt.Errorf("failed to delete IAM managed policy %s: %w", arnStr, err)
	}

	return nil
}
