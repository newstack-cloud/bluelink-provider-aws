package iam

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type samlProviderCreate struct {
	name                            string
	samlMetadataDocument            string
	tags                            []types.Tag
	uniqueSAMLProviderNameGenerator utils.UniqueNameGenerator
}

func (s *samlProviderCreate) Name() string {
	return "create SAML provider"
}

func (s *samlProviderCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Extract name from spec data or generate one
	name, hasName := pluginutils.GetValueByPath("$.name", specData)
	if hasName && core.StringValue(name) != "" {
		s.name = core.StringValue(name)
	} else {
		// Generate a unique name if not provided
		generatedName, err := s.uniqueSAMLProviderNameGenerator(&provider.ResourceDeployInput{
			Changes: changes,
		})
		if err != nil {
			return false, saveOpCtx, fmt.Errorf("failed to generate unique name: %w", err)
		}
		s.name = generatedName
	}

	// Extract SAML metadata document
	samlMetadataDocument, hasMetadata := pluginutils.GetValueByPath("$.samlMetadataDocument", specData)
	if !hasMetadata || core.StringValue(samlMetadataDocument) == "" {
		return false, saveOpCtx, fmt.Errorf("SAML metadata document must be provided and non-empty")
	}
	s.samlMetadataDocument = core.StringValue(samlMetadataDocument)

	// Extract tags
	tags, hasTags := pluginutils.GetValueByPath("$.tags", specData)
	if hasTags && tags != nil && len(tags.Items) > 0 {
		tagItems := tags.Items
		s.tags = make([]types.Tag, len(tagItems))
		for i, item := range tagItems {
			keyNode, hasKey := pluginutils.GetValueByPath("$.key", item)
			valueNode, hasValue := pluginutils.GetValueByPath("$.value", item)
			if !hasKey || !hasValue {
				return false, saveOpCtx, fmt.Errorf("invalid tag format at index %d", i)
			}
			s.tags[i] = types.Tag{
				Key:   aws.String(core.StringValue(keyNode)),
				Value: aws.String(core.StringValue(valueNode)),
			}
		}
	} else {
		s.tags = nil
	}

	return true, saveOpCtx, nil
}

func (s *samlProviderCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	input := &iam.CreateSAMLProviderInput{
		Name:                 aws.String(s.name),
		SAMLMetadataDocument: aws.String(s.samlMetadataDocument),
		Tags:                 sortTagsByKeyForSAML(s.tags),
	}

	output, err := iamService.CreateSAMLProvider(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create SAML provider: %w", err)
	}

	newSaveOpCtx.Data["createSAMLProviderOutput"] = output
	return newSaveOpCtx, nil
}

func newSAMLProviderCreate(generator utils.UniqueNameGenerator) *samlProviderCreate {
	return &samlProviderCreate{
		uniqueSAMLProviderNameGenerator: generator,
	}
}

// sortTagsByKeyForSAML sorts a slice of types.Tag by their Key field.
func sortTagsByKeyForSAML(tags []types.Tag) []types.Tag {
	sorted := make([]types.Tag, len(tags))
	copy(sorted, tags)
	sort.Slice(sorted, func(i, j int) bool {
		return aws.ToString(sorted[i].Key) < aws.ToString(sorted[j].Key)
	})
	return sorted
}
