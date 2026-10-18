//go:build unit

package flex

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/stretchr/testify/suite"
)

type VPCExternalModeTestSuite struct {
	suite.Suite
}

func TestVPCExternalModeTestSuite(t *testing.T) {
	suite.Run(t, new(VPCExternalModeTestSuite))
}

func vpcSpec(fields ...any) *schema.Resource {
	return &schema.Resource{Spec: core.MappingNodeFields(fields...)}
}

// Every input that provisions something is invalid wherever the VPC already
// exists.
func (s *VPCExternalModeTestSuite) Test_provisioning_inputs_are_rejected_in_both_existing_vpc_modes() {
	vpcSchema := vpcResourceSchema()

	for _, field := range []string{
		"preset", "cidrBlock", "enableDNSSupport", "enableDNSHostnames", "region", "tags",
	} {
		attribute, declared := vpcSchema.Attributes[field]
		s.Require().True(declared, "expected a %q input", field)
		s.Require().NotNil(attribute.ValidateFunc, "%q must reject provisioning modes", field)

		for _, mode := range []string{ModeReference, ModeExternal} {
			diagnostics := attribute.ValidateFunc(
				field,
				core.MappingNodeFromString("anything"),
				vpcSpec("mode", core.MappingNodeFromString(mode)),
			)
			s.NotEmpty(diagnostics, "%q must be rejected when mode is %q", field, mode)
		}

		s.Empty(
			attribute.ValidateFunc(
				field,
				core.MappingNodeFromString("anything"),
				vpcSpec("mode", core.MappingNodeFromString(ModeCreate)),
			),
			"%q is how a created VPC is configured, so create mode must accept it", field,
		)
	}
}

// The declaration is only meaningful for a VPC Bluelink does not own.
func (s *VPCExternalModeTestSuite) Test_the_external_declaration_is_rejected_in_the_owned_modes() {
	attribute := vpcResourceSchema().Attributes["externalNetwork"]
	s.Require().NotNil(attribute, "expected an externalNetwork input")
	s.Require().NotNil(attribute.ValidateFunc)

	for _, mode := range []string{ModeCreate, ModeReference} {
		s.NotEmpty(
			attribute.ValidateFunc(
				"externalNetwork",
				core.MappingNodeFields("vpcId", core.MappingNodeFromString("vpc-0abc")),
				vpcSpec("mode", core.MappingNodeFromString(mode)),
			),
			"externalNetwork must be rejected when mode is %q", mode,
		)
	}

	s.Empty(attribute.ValidateFunc(
		"externalNetwork",
		core.MappingNodeFields("vpcId", core.MappingNodeFromString("vpc-0abc")),
		vpcSpec("mode", core.MappingNodeFromString(ModeExternal)),
	))
}

// An external VPC with no declaration has no identity, so it cannot be resolved
// at all. Reported against mode, because a ValidateFunc on an absent field never
// runs.
func (s *VPCExternalModeTestSuite) Test_external_mode_without_a_declaration_is_reported() {
	attribute := vpcResourceSchema().Attributes["mode"]
	s.Require().NotNil(attribute.ValidateFunc)

	diagnostics := attribute.ValidateFunc(
		"mode",
		core.MappingNodeFromString(ModeExternal),
		vpcSpec("mode", core.MappingNodeFromString(ModeExternal)),
	)
	s.Require().Len(diagnostics, 1)
	s.Equal(core.DiagnosticLevelError, diagnostics[0].Level)
	s.Contains(diagnostics[0].Message, "externalNetwork")

	s.Empty(
		attribute.ValidateFunc(
			"mode",
			core.MappingNodeFromString(ModeExternal),
			vpcSpec(
				"mode", core.MappingNodeFromString(ModeExternal),
				"externalNetwork", core.MappingNodeFields(
					"vpcId", core.MappingNodeFromString("vpc-0abc"),
				),
			),
		),
		"a declared external network satisfies the requirement",
	)

	for _, mode := range []string{ModeCreate, ModeReference} {
		s.Empty(
			attribute.ValidateFunc(
				"mode",
				core.MappingNodeFromString(mode),
				vpcSpec("mode", core.MappingNodeFromString(mode)),
			),
			"%q needs no external declaration", mode,
		)
	}
}

// Placement reads subnets by tier, so the declaration is keyed by tier. Egress
// is per tier because the commonest external shape is public subnets routed to
// an internet gateway beside private subnets with no NAT.
func (s *VPCExternalModeTestSuite) Test_the_declaration_is_keyed_by_tier_with_per_tier_egress() {
	external := vpcResourceSchema().Attributes["externalNetwork"]
	s.Require().NotNil(external.Attributes["vpcId"])

	subnets := external.Attributes["subnets"]
	s.Require().NotNil(subnets, "expected a subnets declaration")

	for _, tier := range []string{SubnetTierPrivate, SubnetTierPublic} {
		declared := subnets.Attributes[tier]
		s.Require().NotNil(declared, "expected a %q tier", tier)
		s.Require().NotNil(declared.Attributes["ids"])
		s.Equal(provider.ResourceDefinitionsSchemaTypeArray, declared.Attributes["ids"].Type)

		egress := declared.Attributes["egress"]
		s.Require().NotNil(egress, "expected %q to declare its egress", tier)

		allowed := []string{}
		for _, value := range egress.AllowedValues {
			allowed = append(allowed, core.StringValue(value))
		}
		s.ElementsMatch([]string{EgressInternet, EgressNone}, allowed,
			"the egress vocabulary is what the published spec already documents")
	}
}

// Changing which VPC an app is placed in is not an in-place update.
func (s *VPCExternalModeTestSuite) Test_the_external_declaration_forces_recreation() {
	s.True(vpcResourceSchema().Attributes["externalNetwork"].MustRecreate)
	s.True(vpcResourceSchema().Attributes["mode"].MustRecreate)
}
