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

func (i *iamAccessKeyResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	// Safely get the access key ID from the resource state
	accessKeyID, hasAccessKeyID := pluginutils.GetValueByPath("$.id", input.ResourceState.SpecData)
	if !hasAccessKeyID {
		return fmt.Errorf("access key ID is required for destroy")
	}

	// Get the user name from the resource state
	userName, hasUserName := pluginutils.GetValueByPath("$.userName", input.ResourceState.SpecData)
	if !hasUserName {
		return fmt.Errorf("user name is required for destroy")
	}

	accessKeyIDStr := core.StringValue(accessKeyID)
	userNameStr := core.StringValue(userName)

	if accessKeyIDStr == "" {
		return fmt.Errorf("access key ID is required for destroy")
	}

	if userNameStr == "" {
		return fmt.Errorf("user name is required for destroy")
	}

	// Delete the access key
	_, err = iamService.DeleteAccessKey(ctx, &iam.DeleteAccessKeyInput{
		AccessKeyId: aws.String(accessKeyIDStr),
		UserName:    aws.String(userNameStr),
	})
	if err != nil {
		return fmt.Errorf("failed to delete access key: %w", err)
	}

	return nil
}
