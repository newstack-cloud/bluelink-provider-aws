package iam

import (
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func toIAMTag(tag *utils.Tag) types.Tag {
	return types.Tag{
		Key:   aws.String(tag.Key),
		Value: aws.String(tag.Value),
	}
}

func extractIAMTags(tags []types.Tag) *core.MappingNode {
	tagItems := make([]*core.MappingNode, len(tags))
	for i, tag := range tags {
		tagItems[i] = &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString(aws.ToString(tag.Key)),
				"value": core.MappingNodeFromString(aws.ToString(tag.Value)),
			},
		}
	}
	return &core.MappingNode{
		Items: tagItems,
	}
}

func iamTagsFromSpecData(specData *core.MappingNode) ([]types.Tag, error) {
	var iamTags []types.Tag
	tags, hasTags := pluginutils.GetValueByPath("$.tags", specData)
	if hasTags {
		parsedTags, err := iamTagsFromValue(tags)
		if err != nil {
			return nil, err
		}
		iamTags = parsedTags
	}
	return iamTags, nil
}

func iamTagsFromValue(value *core.MappingNode) ([]types.Tag, error) {
	iamTags := []types.Tag{}
	if core.IsArrayMappingNode(value) {
		for i, item := range value.Items {
			keyNode, hasKey := pluginutils.GetValueByPath("$.key", item)
			valueNode, hasValue := pluginutils.GetValueByPath("$.value", item)
			if !hasKey || !hasValue {
				return iamTags, fmt.Errorf("invalid tag format at index %d", i)
			}
			iamTags = append(iamTags, types.Tag{
				Key:   aws.String(core.StringValue(keyNode)),
				Value: aws.String(core.StringValue(valueNode)),
			})
		}
	}
	return iamTags, nil
}

func sortTagsByKey(tags []types.Tag) []types.Tag {
	sort.Slice(tags, func(i, j int) bool {
		return aws.ToString(tags[i].Key) < aws.ToString(tags[j].Key)
	})
	return tags
}
