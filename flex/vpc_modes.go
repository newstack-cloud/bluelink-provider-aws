package flex

import (
	"fmt"
	"slices"
	"strings"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
)

// The modes a flex VPC can be declared in.
//
// Create and Reference both describe a VPC Bluelink owns.
// Reference resolves one another blueprint created,
// by the bluelink:flex-vpc:name tag, so its subnets,
// tiers and groups are all discoverable from tags Bluelink itself wrote.
//
// External describes a VPC Bluelink did not create and does not own. None of
// those tags exist on it, so its topology is declared rather than discovered,
// and nothing about it is provisioned.
const (
	ModeCreate    = "create"
	ModeReference = "reference"
	ModeExternal  = "external"
)

// Modes in which the VPC already exists and is not provisioned here.
var existingVPCModes = []string{ModeReference, ModeExternal}

func conflictsWithModeIn(
	modes ...string,
) func(string, *core.MappingNode, *schema.Resource) []*core.Diagnostic {
	return func(
		fieldName string,
		value *core.MappingNode,
		resource *schema.Resource,
	) []*core.Diagnostic {
		modeValue, _ := core.GetPathValue(
			"$.mode",
			resource.Spec,
			core.MappingNodeMaxTraverseDepth,
		)
		mode := core.StringValue(modeValue)
		if !slices.Contains(modes, mode) {
			return nil
		}

		return []*core.Diagnostic{{
			Level: core.DiagnosticLevelError,
			Message: fmt.Sprintf(
				"%q cannot be set when \"mode\" has a value of %q, as the VPC already "+
					"exists and is not provisioned by this resource.",
				fieldName, mode,
			),
		}}
	}
}

// Reports a field that is only meaningful in one mode and has to
// be set there. The schema's own Required list cannot express it, since the
// field is invalid in the other modes.
func requiredInMode(
	mode string,
	fieldPath string,
	fieldName string,
) func(string, *core.MappingNode, *schema.Resource) []*core.Diagnostic {
	return func(
		_ string,
		_ *core.MappingNode,
		resource *schema.Resource,
	) []*core.Diagnostic {
		modeValue, _ := core.GetPathValue(
			"$.mode", resource.Spec, core.MappingNodeMaxTraverseDepth,
		)
		if core.StringValue(modeValue) != mode {
			return nil
		}

		fieldValue, _ := core.GetPathValue(
			fieldPath, resource.Spec, core.MappingNodeMaxTraverseDepth,
		)
		if !core.IsNilMappingNode(fieldValue) {
			return nil
		}

		return []*core.Diagnostic{{
			Level: core.DiagnosticLevelError,
			Message: fmt.Sprintf(
				"%q must be set when \"mode\" is %q, as it is what identifies the VPC.",
				fieldName, mode,
			),
		}}
	}
}

// Subnet tiers an external VPC declares. These are the only tiers placement
// consumes, which is why the declaration is keyed by tier rather than by subnet
// name as the discovered form is.
const (
	SubnetTierPrivate = "private"
	SubnetTierPublic  = "public"
)

// What outbound reach a tier's topology provides. Declared because it is not
// derivable with any confidence from outside.
const (
	EgressInternet = "internet"
	EgressNone     = "none"
)

var egressValues = []string{EgressInternet, EgressNone}

func joinQuoted(values []string) string {
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, fmt.Sprintf("%q", value))
	}

	return strings.Join(quoted, ", ")
}

func externalSubnetTierSchema(tier string) *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:  provider.ResourceDefinitionsSchemaTypeObject,
		Label: "ExternalSubnetTier",
		Description: fmt.Sprintf(
			"The %s subnets of the external VPC and the outbound reach they provide.", tier,
		),
		Nullable: true,
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"ids": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: fmt.Sprintf("The IDs of the %s subnets.", tier),
				Items: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
				},
			},
			"egress": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: fmt.Sprintf(
					"What outbound reach this tier's topology provides, one of %s. Declared "+
						"because it is not derivable from outside, for example, a NAT gateway in a route "+
						"table is necessary and not sufficient, and a centralised egress VPC "+
						"reached over a transit gateway attachment presents as a plain route "+
						"to a transit gateway.",
					joinQuoted(egressValues),
				),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString(EgressInternet),
					core.MappingNodeFromString(EgressNone),
				},
			},
		},
	}
}
