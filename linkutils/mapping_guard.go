package linkutils

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
)

// maxMappingTraverseDepth bounds path resolution when checking a mapping. Link
// data is shallow (a per-resource object holding a statement or an ARN), so a
// generous bound still costs nothing and stops a malformed path looping.
const maxMappingTraverseDepth = 20

// fieldSelectorPattern matches an array item selector that filters on a field,
// as in [@.sid="Something"]. The captured group is the field name the selector
// requires the item to carry.
var fieldSelectorPattern = regexp.MustCompile(`\[@\.([A-Za-z0-9_]+)\s*=`)

// ValidateResourceDataMappings reports the mappings whose path selects on a field the link
// data it points at does not have.
//
// A ResourceDataMappings entry is a promise with two halves: the key says where a
// contribution belongs on the target resource, and the value says where the contribution's
// data lives in the link's own data. When the key selects an array item by field, as inline
// policy statements do with [@.sid="..."], the data at the value path has to carry that
// field - composition locates the item by those conditions, so a contribution that does not
// satisfy its own selector can never be applied or found again. This is the same check
// composition performs at deploy time, done where it is cheap (a flex VPC link recorded its
// statement with IAM's "Sid" against a path selecting on @.sid, which was inert until
// composition began reading the mappings and then failed staging).
//
// Errors are returned sorted so a failing test names the same mapping every run.
func ValidateResourceDataMappings(
	mappings map[string]string,
	linkData *core.MappingNode,
) []error {
	keys := make([]string, 0, len(mappings))
	for key := range mappings {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	errs := []error{}
	for _, mappingKey := range keys {
		if err := validateMapping(mappingKey, mappings[mappingKey], linkData); err != nil {
			errs = append(errs, err)
		}
	}
	return errs
}

func validateMapping(mappingKey, dataPath string, linkData *core.MappingNode) error {
	required, hasSelector := finalSelectorFields(mappingKey)
	if !hasSelector {
		// Nothing selects on a field, so there is no promise to break. A plain path
		// that is simply absent is a different concern, reported by composition as
		// an unresolved projection rather than as a broken mapping.
		return nil
	}

	value, err := core.GetPathValue(rootedPath(dataPath), linkData, maxMappingTraverseDepth)
	if err != nil {
		return fmt.Errorf(
			"mapping %q points at link data path %q, which could not be read: %w",
			mappingKey, dataPath, err,
		)
	}
	if value == nil {
		return fmt.Errorf(
			"mapping %q points at link data path %q, which holds nothing",
			mappingKey, dataPath,
		)
	}

	for _, field := range required {
		if value.Fields == nil {
			return fmt.Errorf(
				"mapping %q selects on %q but the link data at %q is not an object",
				mappingKey, field, dataPath,
			)
		}
		if _, ok := value.Fields[field]; !ok {
			return fmt.Errorf(
				"mapping %q selects on %q but the link data at %q has no %q field (it has: %s). "+
					"A contribution has to carry the fields its own path selects on, or it can never "+
					"be located again",
				mappingKey, field, dataPath, field, strings.Join(fieldNames(value), ", "),
			)
		}
	}
	return nil
}

// rootedPath turns a link-data field path into one GetPathValue accepts.
//
// The path helpers produce forms like ["roleName"].permission, which is rooted at
// the link data object but written without one. A bracket accessor follows the
// root directly, while a named field needs the separating dot.
func rootedPath(dataPath string) string {
	trimmed := strings.TrimPrefix(dataPath, ".")
	if strings.HasPrefix(trimmed, "[") {
		return "$" + trimmed
	}
	return "$." + trimmed
}

// finalSelectorFields returns the fields the LAST array selector in a mapping path filters
// on, which are the only ones the contribution itself has to carry.
//
// A path can select more than once: the inline statement path picks the policy by name and
// then the statement by sid. Every selector but the last locates a container on the TARGET
// resource, which composition creates as it walks; only the final selector identifies the
// contributed item.
func finalSelectorFields(mappingKey string) ([]string, bool) {
	matches := fieldSelectorPattern.FindAllStringSubmatch(mappingKey, -1)
	if len(matches) == 0 {
		return nil, false
	}
	last := matches[len(matches)-1]
	return []string{last[1]}, true
}

func fieldNames(node *core.MappingNode) []string {
	if node.Fields == nil {
		return nil
	}
	names := make([]string, 0, len(node.Fields))
	for name := range node.Fields {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
