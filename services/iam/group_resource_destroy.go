package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (i *iamGroupResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	// Get the group ARN from the resource state
	arn := core.StringValue(input.ResourceState.SpecData.Fields["arn"])
	if arn == "" {
		return fmt.Errorf("ARN is required for destroy operation")
	}

	// Extract group name from ARN
	groupName, err := extractGroupNameFromARN(arn)
	if err != nil {
		return fmt.Errorf("failed to extract group name from ARN: %w", err)
	}

	// Before deleting the group, we need to clean up all attached resources
	if err := i.cleanupGroupResources(ctx, iamService, groupName); err != nil {
		return fmt.Errorf("failed to cleanup group resources: %w", err)
	}

	// Finally, delete the group
	_, err = iamService.DeleteGroup(ctx, &iam.DeleteGroupInput{
		GroupName: aws.String(groupName),
	})
	if err != nil {
		return fmt.Errorf("failed to delete IAM group %s: %w", groupName, err)
	}

	return nil
}

// cleanupGroupResources removes all resources attached to the group before deletion.
func (i *iamGroupResourceActions) cleanupGroupResources(
	ctx context.Context,
	iamService iamservice.Service,
	groupName string,
) error {
	// 1. Detach all managed policies
	if err := i.detachAllManagedPolicies(ctx, iamService, groupName); err != nil {
		return fmt.Errorf("failed to detach managed policies: %w", err)
	}

	// 2. Delete all inline policies
	if err := i.deleteAllInlinePolicies(ctx, iamService, groupName); err != nil {
		return fmt.Errorf("failed to delete inline policies: %w", err)
	}

	return nil
}

func (i *iamGroupResourceActions) detachAllManagedPolicies(
	ctx context.Context,
	iamService iamservice.Service,
	groupName string,
) error {
	// List all attached managed policies
	result, err := iamService.ListAttachedGroupPolicies(ctx, &iam.ListAttachedGroupPoliciesInput{
		GroupName: aws.String(groupName),
	})
	if err != nil {
		return fmt.Errorf("failed to list attached policies for group %s: %w", groupName, err)
	}

	// Detach each managed policy
	for _, policy := range result.AttachedPolicies {
		_, err := iamService.DetachGroupPolicy(ctx, &iam.DetachGroupPolicyInput{
			GroupName: aws.String(groupName),
			PolicyArn: policy.PolicyArn,
		})
		if err != nil {
			return fmt.Errorf("failed to detach policy %s from group: %w", aws.ToString(policy.PolicyArn), err)
		}
	}

	return nil
}

func (i *iamGroupResourceActions) deleteAllInlinePolicies(
	ctx context.Context,
	iamService iamservice.Service,
	groupName string,
) error {
	// List all inline policies
	result, err := iamService.ListGroupPolicies(ctx, &iam.ListGroupPoliciesInput{
		GroupName: aws.String(groupName),
	})
	if err != nil {
		return fmt.Errorf("failed to list inline policies for group %s: %w", groupName, err)
	}

	// Delete each inline policy
	for _, policyName := range result.PolicyNames {
		_, err := iamService.DeleteGroupPolicy(ctx, &iam.DeleteGroupPolicyInput{
			GroupName:  aws.String(groupName),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return fmt.Errorf("failed to delete inline policy %s from group: %w", policyName, err)
		}
	}

	return nil
}
