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

func (i *iamInstanceProfileResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Get the instance profile name from the resource spec
	instanceProfileName, hasInstanceProfileName := pluginutils.GetValueByPath("$.instanceProfileName", input.CurrentResourceSpec)
	if !hasInstanceProfileName {
		return nil, fmt.Errorf("instance profile name is required for get external state")
	}

	instanceProfileNameStr := core.StringValue(instanceProfileName)

	if instanceProfileNameStr == "" {
		return nil, fmt.Errorf("instance profile name is required for get external state")
	}

	// Get the instance profile
	getInstanceProfileInput := &iam.GetInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileNameStr),
	}

	getInstanceProfileOutput, err := iamService.GetInstanceProfile(ctx, getInstanceProfileInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get instance profile: %w", err)
	}

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"instanceProfileName": core.MappingNodeFromString(aws.ToString(getInstanceProfileOutput.InstanceProfile.InstanceProfileName)),
		"path":                core.MappingNodeFromString(aws.ToString(getInstanceProfileOutput.InstanceProfile.Path)),
		"arn":                 core.MappingNodeFromString(aws.ToString(getInstanceProfileOutput.InstanceProfile.Arn)),
	}

	// Add role information if present
	if len(getInstanceProfileOutput.InstanceProfile.Roles) > 0 {
		role := getInstanceProfileOutput.InstanceProfile.Roles[0]
		externalState["role"] = core.MappingNodeFromString(aws.ToString(role.RoleName))
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}
