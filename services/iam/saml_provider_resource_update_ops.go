package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type samlProviderMetadataUpdate struct {
	arn                  string
	samlMetadataDocument string
}

func (s *samlProviderMetadataUpdate) Name() string {
	return "update SAML provider metadata"
}

func (s *samlProviderMetadataUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the SAML provider ARN from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	if currentStateSpecData == nil {
		return false, saveOpCtx, fmt.Errorf("current state spec data is required for SAML provider update")
	}
	arn, hasArn := pluginutils.GetValueByPath("$.arn", currentStateSpecData)
	if !hasArn {
		return false, saveOpCtx, fmt.Errorf("SAML provider ARN is required for update")
	}
	s.arn = core.StringValue(arn)

	// Check if samlMetadataDocument was modified
	metadataModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.samlMetadataDocument" {
			metadataModified = true
			break
		}
	}

	if !metadataModified {
		return false, saveOpCtx, nil
	}

	// Get new metadata document
	samlMetadataDocument, hasMetadata := pluginutils.GetValueByPath("$.samlMetadataDocument", specData)
	if !hasMetadata || core.StringValue(samlMetadataDocument) == "" {
		return false, saveOpCtx, fmt.Errorf("SAML metadata document must be provided and non-empty")
	}
	s.samlMetadataDocument = core.StringValue(samlMetadataDocument)

	return true, saveOpCtx, nil
}

func (s *samlProviderMetadataUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := iamService.UpdateSAMLProvider(ctx, &iam.UpdateSAMLProviderInput{
		SAMLMetadataDocument: aws.String(s.samlMetadataDocument),
		SAMLProviderArn:      aws.String(s.arn),
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to update SAML provider metadata: %w", err)
	}

	return saveOpCtx, nil
}

type samlProviderTagsUpdate struct {
	arn          string
	tagsToAdd    []types.Tag
	tagsToRemove []string
}

func (s *samlProviderTagsUpdate) Name() string {
	return "update SAML provider tags"
}

func (s *samlProviderTagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the SAML provider ARN from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	if currentStateSpecData == nil {
		return false, saveOpCtx, fmt.Errorf("current state spec data is required for SAML provider update")
	}
	arn, hasArn := pluginutils.GetValueByPath("$.arn", currentStateSpecData)
	if !hasArn {
		return false, saveOpCtx, fmt.Errorf("SAML provider ARN is required for update")
	}
	s.arn = core.StringValue(arn)

	// Check if tags were modified
	tagsModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.tags" {
			tagsModified = true
			break
		}
	}

	if !tagsModified {
		return false, saveOpCtx, nil
	}

	// Get current tags
	currentTags := []types.Tag{}
	if currentTagsList, exists := pluginutils.GetValueByPath("$.tags", currentStateSpecData); exists && currentTagsList != nil {
		for _, item := range currentTagsList.Items {
			keyNode, hasKey := pluginutils.GetValueByPath("$.key", item)
			valueNode, hasValue := pluginutils.GetValueByPath("$.value", item)
			if hasKey && hasValue {
				currentTags = append(currentTags, types.Tag{
					Key:   aws.String(core.StringValue(keyNode)),
					Value: aws.String(core.StringValue(valueNode)),
				})
			}
		}
	}

	// Get desired tags
	desiredTags := []types.Tag{}
	if desiredTagsList, exists := pluginutils.GetValueByPath("$.tags", specData); exists && desiredTagsList != nil {
		for _, item := range desiredTagsList.Items {
			keyNode, hasKey := pluginutils.GetValueByPath("$.key", item)
			valueNode, hasValue := pluginutils.GetValueByPath("$.value", item)
			if hasKey && hasValue {
				desiredTags = append(desiredTags, types.Tag{
					Key:   aws.String(core.StringValue(keyNode)),
					Value: aws.String(core.StringValue(valueNode)),
				})
			}
		}
	}

	// Calculate tags to add and remove
	s.tagsToAdd, s.tagsToRemove = diffTags(desiredTags, currentTags)

	return len(s.tagsToAdd) > 0 || len(s.tagsToRemove) > 0, saveOpCtx, nil
}

func (s *samlProviderTagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove tags
	if len(s.tagsToRemove) > 0 {
		_, err := iamService.UntagSAMLProvider(ctx, &iam.UntagSAMLProviderInput{
			SAMLProviderArn: aws.String(s.arn),
			TagKeys:         s.tagsToRemove,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove tags: %w", err)
		}
	}

	// Add tags
	if len(s.tagsToAdd) > 0 {
		_, err := iamService.TagSAMLProvider(ctx, &iam.TagSAMLProviderInput{
			SAMLProviderArn: aws.String(s.arn),
			Tags:            sortTagsByKeyForSAML(s.tagsToAdd),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add tags: %w", err)
		}
	}

	return saveOpCtx, nil
}

// diffTags computes the tags to add and the tag keys to remove
func diffTags(desired, current []types.Tag) (toAdd []types.Tag, toRemove []string) {
	currentMap := make(map[string]string)
	for _, tag := range current {
		currentMap[aws.ToString(tag.Key)] = aws.ToString(tag.Value)
	}

	desiredMap := make(map[string]string)
	for _, tag := range desired {
		key := aws.ToString(tag.Key)
		value := aws.ToString(tag.Value)
		desiredMap[key] = value

		// If the tag doesn't exist in current or has a different value, add it
		if currentValue, exists := currentMap[key]; !exists || currentValue != value {
			toAdd = append(toAdd, tag)
		}
	}

	// Find tags to remove (exist in current but not in desired)
	for key := range currentMap {
		if _, exists := desiredMap[key]; !exists {
			toRemove = append(toRemove, key)
		}
	}

	return toAdd, toRemove
}
