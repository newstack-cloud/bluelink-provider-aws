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

func (i *iamInstanceProfileResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	// Get the instance profile name from the resource state
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.ResourceState.SpecData)
	if !hasArn {
		return fmt.Errorf("ARN is required for instance profile destruction")
	}

	instanceProfileName, err := extractInstanceProfileNameFromARN(core.StringValue(arn))
	if err != nil {
		return fmt.Errorf("failed to extract instance profile name: %w", err)
	}

	// Get the role name from the resource state
	role, hasRole := pluginutils.GetValueByPath("$.role", input.ResourceState.SpecData)
	if hasRole {
		roleSpec := core.StringValue(role)
		roleName, err := extractRoleNameFromRoleSpec(roleSpec)
		if err != nil {
			return fmt.Errorf("failed to extract role name: %w", err)
		}

		// Remove role from instance profile
		removeRoleInput := &iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: aws.String(instanceProfileName),
			RoleName:            aws.String(roleName),
		}

		_, err = iamService.RemoveRoleFromInstanceProfile(ctx, removeRoleInput)
		if err != nil {
			return fmt.Errorf("failed to remove role from instance profile: %w", err)
		}
	}

	// Delete the instance profile
	deleteInstanceProfileInput := &iam.DeleteInstanceProfileInput{
		InstanceProfileName: aws.String(instanceProfileName),
	}

	_, err = iamService.DeleteInstanceProfile(ctx, deleteInstanceProfileInput)
	if err != nil {
		return fmt.Errorf("failed to delete instance profile: %w", err)
	}

	return nil
}
