package iam

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamRoleResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	// Get the role name from the computed ARN field
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.ResourceState.SpecData)
	if !hasArn {
		return fmt.Errorf("ARN is required for destroy operation")
	}

	arnStr := core.StringValue(arn)
	if arnStr == "" {
		return fmt.Errorf("ARN cannot be empty for destroy operation")
	}

	// Extract role name from ARN
	roleName, err := extractRoleNameFromARN(arnStr)
	if err != nil {
		return fmt.Errorf("failed to extract role name from ARN %s: %w", arnStr, err)
	}

	// Remove inline policies that are managed by this blueprint
	if policiesNode, exists := pluginutils.GetValueByPath("$.policies", input.ResourceState.SpecData); exists && policiesNode != nil && len(policiesNode.Items) > 0 {
		for _, policyNode := range policiesNode.Items {
			policyName, hasPolicyName := pluginutils.GetValueByPath("$.policyName", policyNode)
			if !hasPolicyName {
				continue // Skip policies without names
			}
			_, err := iamService.DeleteRolePolicy(ctx, &iam.DeleteRolePolicyInput{
				RoleName:   aws.String(roleName),
				PolicyName: aws.String(core.StringValue(policyName)),
			})
			if err != nil {
				return fmt.Errorf("failed to delete inline policy %s: %w", core.StringValue(policyName), err)
			}
		}
	}

	// Detach managed policies that are managed by this blueprint
	if managedPolicyArnsNode, exists := pluginutils.GetValueByPath("$.managedPolicyArns", input.ResourceState.SpecData); exists && managedPolicyArnsNode != nil && len(managedPolicyArnsNode.Items) > 0 {
		for _, policyArnNode := range managedPolicyArnsNode.Items {
			policyArn := core.StringValue(policyArnNode)
			_, err := iamService.DetachRolePolicy(ctx, &iam.DetachRolePolicyInput{
				RoleName:  aws.String(roleName),
				PolicyArn: aws.String(policyArn),
			})
			if err != nil {
				return fmt.Errorf("failed to detach managed policy %s: %w", policyArn, err)
			}
		}
	}

	// Remove permissions boundary if it exists
	if permissionsBoundaryNode, exists := pluginutils.GetValueByPath("$.permissionsBoundary", input.ResourceState.SpecData); exists && permissionsBoundaryNode != nil {
		_, err := iamService.DeleteRolePermissionsBoundary(ctx, &iam.DeleteRolePermissionsBoundaryInput{
			RoleName: aws.String(roleName),
		})
		if err != nil {
			return fmt.Errorf("failed to delete permissions boundary for role %s: %w", roleName, err)
		}
	}

	// Now attempt to delete the role
	_, err = iamService.DeleteRole(ctx, &iam.DeleteRoleInput{
		RoleName: aws.String(roleName),
	})

	// Handle DeleteConflict errors with helpful error messages
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) && apiError.ErrorCode() == "DeleteConflict" {
			return fmt.Errorf("failed to delete role %s: %s. This may be due to policies, instance profiles, or other resources attached outside of this blueprint. Please remove them manually and try again", roleName, apiError.ErrorMessage())
		}
		return fmt.Errorf("failed to delete role %s: %w", roleName, err)
	}

	return nil
}
