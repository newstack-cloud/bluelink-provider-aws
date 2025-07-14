package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type managedPolicyVersionUpdate struct {
	policyDocument *core.MappingNode
}

func (m *managedPolicyVersionUpdate) Name() string {
	return "update policy version"
}

func (m *managedPolicyVersionUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if the policy document has changed
	if policyDocNode, ok := specData.Fields["policyDocument"]; ok && policyDocNode != nil {
		m.policyDocument = policyDocNode
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (m *managedPolicyVersionUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	policyArn := saveOpCtx.ProviderUpstreamID

	// Convert the structured policy document to JSON string
	policyJSON, err := json.Marshal(m.policyDocument)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to marshal policy document: %w", err)
	}

	// Create a new policy version
	_, err = iamService.CreatePolicyVersion(ctx, &iam.CreatePolicyVersionInput{
		PolicyArn:      aws.String(policyArn),
		PolicyDocument: aws.String(string(policyJSON)),
		SetAsDefault:   true,
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create new policy version for %s: %w", policyArn, err)
	}

	return saveOpCtx, nil
}

type managedPolicyTagsUpdate struct {
	tags []*core.MappingNode
}

func (m *managedPolicyTagsUpdate) Name() string {
	return "update tags"
}

func (m *managedPolicyTagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are tags to update
	if tagsNode, ok := specData.Fields["tags"]; ok && tagsNode != nil && len(tagsNode.Items) > 0 {
		m.tags = tagsNode.Items
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (m *managedPolicyTagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	policyArn := saveOpCtx.ProviderUpstreamID

	// Get current tags to determine what needs to be removed
	currentTagsOutput, err := iamService.ListPolicyTags(ctx, &iam.ListPolicyTagsInput{
		PolicyArn: aws.String(policyArn),
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to list current tags for policy %s: %w", policyArn, err)
	}

	// Build a map of current tag keys
	currentTagKeys := make(map[string]bool)
	for _, tag := range currentTagsOutput.Tags {
		currentTagKeys[aws.ToString(tag.Key)] = true
	}

	// Build a map of new tag keys
	newTagKeys := make(map[string]bool)
	for _, tagNode := range m.tags {
		key := core.StringValue(tagNode.Fields["key"])
		newTagKeys[key] = true
	}

	// Remove tags that are no longer present
	tagsToRemove := make([]string, 0)
	for key := range currentTagKeys {
		if !newTagKeys[key] {
			tagsToRemove = append(tagsToRemove, key)
		}
	}

	if len(tagsToRemove) > 0 {
		_, err = iamService.UntagPolicy(ctx, &iam.UntagPolicyInput{
			PolicyArn: aws.String(policyArn),
			TagKeys:   tagsToRemove,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove tags from policy %s: %w", policyArn, err)
		}
	}

	// Add new tags
	if len(m.tags) > 0 {
		// Convert tags to the format expected by AWS
		tags := make([]types.Tag, 0, len(m.tags))
		for _, tagNode := range m.tags {
			key := core.StringValue(tagNode.Fields["key"])
			tagValue := core.StringValue(tagNode.Fields["value"])
			tags = append(tags, types.Tag{
				Key:   aws.String(key),
				Value: aws.String(tagValue),
			})
		}

		_, err = iamService.TagPolicy(ctx, &iam.TagPolicyInput{
			PolicyArn: aws.String(policyArn),
			Tags:      tags,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add tags to policy %s: %w", policyArn, err)
		}
	}

	return saveOpCtx, nil
}
