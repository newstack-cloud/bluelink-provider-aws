package utils

import (
	"slices"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Tag is an intermediary representation of a tag that is used across
// AWS services (The SDK provides a different type for each service).
// This is used to provide a consistent interface for the tag changes
// that should be converted to the upstream service's tag type.
type Tag struct {
	Key   string
	Value string
}

// TagsDiffResult is the result of the DiffTags function,
// it contains the tags that should be set and the tag keys that should be removed.
type TagsDiffResult[UpstreamTag any] struct {
	ToSet    []UpstreamTag
	ToRemove []string
}

// DiffTags provides a general purpose utility to derive the difference
// between two sets of tags stored in a resource spec for the purpose of making
// calls to the upstream service to apply tag changes.
// There is a limitation in the `Changes` data that the plugin receives
// in that when the tags are stored in a list, it does not provide sufficient information
// on the key of the tags to be removed as it just reports an updated or removed index
// in the list.
// For this reason, resource implementations need to use the actual current and upcoming
// resource spec data to derive the tag changes so that the correct tags are removed, added
// and replaced.
//
// tagsRootPath is the path to the tags field in the resource spec,
// the expected format is to use "$" to represent the root of the spec (e.g. "$.tags").
func DiffTags[UpstreamTag any](
	changes *provider.Changes,
	tagsRootPath string,
	transformTag func(tag *Tag) UpstreamTag,
) *TagsDiffResult[UpstreamTag] {
	result := &TagsDiffResult[UpstreamTag]{
		ToSet:    []UpstreamTag{},
		ToRemove: []string{},
	}

	currentSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentSpecTags, _ := pluginutils.GetValueByPath(tagsRootPath, currentSpecData)
	newSpecData := pluginutils.GetResolvedResourceSpecData(changes)
	newSpecTags, _ := pluginutils.GetValueByPath(tagsRootPath, newSpecData)

	currentTags := toTagsMap(currentSpecTags)
	desiredTags := toTagsMap(newSpecTags)

	// Calculate tags to add/update
	toSetIntermediary := []*Tag{}
	for key, value := range desiredTags {
		toSetIntermediary = append(toSetIntermediary, &Tag{
			Key:   key,
			Value: value,
		})
	}

	// Calculate tags to remove
	for key := range currentTags {
		if _, exists := desiredTags[key]; !exists {
			result.ToRemove = append(result.ToRemove, key)
		}
	}

	// Sort the tags to set and remove by key, this is mostly helpful
	// for deterministic comparison of output.
	slices.SortFunc(toSetIntermediary, func(i, j *Tag) int {
		return strings.Compare(i.Key, j.Key)
	})
	slices.Sort(result.ToRemove)

	for _, tag := range toSetIntermediary {
		result.ToSet = append(result.ToSet, transformTag(tag))
	}

	return result
}

func toTagsMap(specTags *core.MappingNode) map[string]string {
	tagMap := make(map[string]string)
	if core.IsArrayMappingNode(specTags) {
		for _, item := range specTags.Items {
			key, _ := pluginutils.GetValueByPath("$.key", item)
			value, _ := pluginutils.GetValueByPath("$.value", item)
			tagMap[core.StringValue(key)] = core.StringValue(value)
		}
	} else if core.IsObjectMappingNode(specTags) {
		for key, value := range specTags.Fields {
			tagMap[key] = core.StringValue(value)
		}
	}
	return tagMap
}
