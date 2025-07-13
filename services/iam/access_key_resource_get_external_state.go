package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamAccessKeyResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Safely get the access key ID from the resource spec
	accessKeyID, hasAccessKeyID := pluginutils.GetValueByPath("$.id", input.CurrentResourceSpec)
	if !hasAccessKeyID {
		return nil, fmt.Errorf("access key ID is required for get external state")
	}

	// Get the user name from the resource spec
	userName, hasUserName := pluginutils.GetValueByPath("$.userName", input.CurrentResourceSpec)
	if !hasUserName {
		return nil, fmt.Errorf("user name is required for get external state")
	}

	accessKeyIDStr := core.StringValue(accessKeyID)
	userNameStr := core.StringValue(userName)

	if accessKeyIDStr == "" {
		return nil, fmt.Errorf("access key ID is required for get external state")
	}

	if userNameStr == "" {
		return nil, fmt.Errorf("user name is required for get external state")
	}

	// List access keys for the user to find the specific one
	result, err := iamService.ListAccessKeys(ctx, &iam.ListAccessKeysInput{
		UserName: aws.String(userNameStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list access keys: %w", err)
	}

	// Find the specific access key
	var foundAccessKey types.AccessKeyMetadata
	for _, accessKey := range result.AccessKeyMetadata {
		if aws.ToString(accessKey.AccessKeyId) == accessKeyIDStr {
			foundAccessKey = accessKey
			break
		}
	}

	if aws.ToString(foundAccessKey.AccessKeyId) == "" {
		return nil, fmt.Errorf("access key %s not found for user %s", accessKeyIDStr, userNameStr)
	}

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"id":       core.MappingNodeFromString(aws.ToString(foundAccessKey.AccessKeyId)),
		"userName": core.MappingNodeFromString(userNameStr),
		"status":   core.MappingNodeFromString(string(foundAccessKey.Status)),
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}
