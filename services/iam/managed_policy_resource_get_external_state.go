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

func (i *iamManagedPolicyResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Safely get the policy ARN from the resource spec
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.CurrentResourceSpec)
	if !hasArn {
		return nil, fmt.Errorf("ARN is required for get external state operation")
	}

	arnStr := core.StringValue(arn)
	if arnStr == "" {
		return nil, fmt.Errorf("ARN is required for get external state operation")
	}

	// Get the managed policy
	getPolicyOutput, err := iamService.GetPolicy(ctx, &iam.GetPolicyInput{
		PolicyArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get IAM managed policy %s: %w", arnStr, err)
	}

	// Get the policy tags
	listPolicyTagsOutput, err := iamService.ListPolicyTags(ctx, &iam.ListPolicyTagsInput{
		PolicyArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list tags for IAM managed policy %s: %w", arnStr, err)
	}

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"policyName": core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.PolicyName)),
		"path":       core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.Path)),
		"arn":        core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.Arn)),
		"id":         core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.PolicyId)),
	}

	// Add optional fields if they exist
	if getPolicyOutput.Policy.Description != nil {
		externalState["description"] = core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.Description))
	}

	// Add computed fields
	if getPolicyOutput.Policy.AttachmentCount != nil {
		externalState["attachmentCount"] = core.MappingNodeFromInt(int(*getPolicyOutput.Policy.AttachmentCount))
	}
	if getPolicyOutput.Policy.CreateDate != nil {
		externalState["createDate"] = core.MappingNodeFromString(getPolicyOutput.Policy.CreateDate.Format("2006-01-02T15:04:05Z"))
	}
	if getPolicyOutput.Policy.DefaultVersionId != nil {
		externalState["defaultVersionId"] = core.MappingNodeFromString(aws.ToString(getPolicyOutput.Policy.DefaultVersionId))
	}
	if getPolicyOutput.Policy.IsAttachable {
		externalState["isAttachable"] = core.MappingNodeFromBool(getPolicyOutput.Policy.IsAttachable)
	}
	if getPolicyOutput.Policy.PermissionsBoundaryUsageCount != nil {
		externalState["permissionsBoundaryUsageCount"] = core.MappingNodeFromInt(int(*getPolicyOutput.Policy.PermissionsBoundaryUsageCount))
	}
	if getPolicyOutput.Policy.UpdateDate != nil {
		externalState["updateDate"] = core.MappingNodeFromString(getPolicyOutput.Policy.UpdateDate.Format("2006-01-02T15:04:05Z"))
	}

	// Add tags if they exist
	if len(listPolicyTagsOutput.Tags) > 0 {
		tags := make([]*core.MappingNode, 0, len(listPolicyTagsOutput.Tags))
		for _, tag := range listPolicyTagsOutput.Tags {
			tags = append(tags, &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"key":   core.MappingNodeFromString(aws.ToString(tag.Key)),
					"value": core.MappingNodeFromString(aws.ToString(tag.Value)),
				},
			})
		}
		externalState["tags"] = &core.MappingNode{
			Items: tags,
		}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}
