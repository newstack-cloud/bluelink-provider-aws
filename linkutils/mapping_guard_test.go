//go:build unit

package linkutils

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/suite"
)

type MappingGuardSuite struct {
	suite.Suite
}

const roleName = "celerityLambdaExec_fbde8c62"
const testSID = "VPCNetworkInterfacesstatsControllerscheduledAggregatelambdafunc"

// statementNode builds a permission statement with the given field casing, so a
// test can reproduce both the correct form and the one that shipped.
func statementNode(sidKey string) *core.MappingNode {
	return core.MappingNodeFields(
		sidKey, core.MappingNodeFromString(testSID),
		"effect", core.MappingNodeFromString("Allow"),
		"resource", core.MappingNodeFromString("*"),
	)
}

func linkDataFor(statement *core.MappingNode) *core.MappingNode {
	return core.MappingNodeFields(
		"statsController_scheduledAggregate_lambda_funcExecutionRole",
		core.MappingNodeFields(PermissionFieldName, statement),
	)
}

func (s *MappingGuardSuite) mappings() map[string]string {
	return map[string]string{
		InlineAccessStatementPath(roleName, testSID): PermissionFieldPath(
			"statsController_scheduledAggregate_lambda_funcExecutionRole",
		),
	}
}

func (s *MappingGuardSuite) Test_accepts_a_statement_carrying_the_field_it_selects_on() {
	errs := ValidateResourceDataMappings(s.mappings(), linkDataFor(statementNode("sid")))
	s.Assert().Empty(errs)
}

func (s *MappingGuardSuite) Test_rejects_the_casing_that_shipped() {
	// The real defect: the flex VPC link recorded "Sid" (IAM's JSON casing) while
	// its own mapping path selects on @.sid. Inert until composition read the
	// mappings, then a hard staging failure on every instance with a VPC-attached
	// function.
	errs := ValidateResourceDataMappings(s.mappings(), linkDataFor(statementNode("Sid")))

	s.Require().Len(errs, 1)
	s.Assert().Contains(errs[0].Error(), `selects on "sid"`)
	s.Assert().Contains(errs[0].Error(), `has no "sid" field`)
	// Naming what it does have is what turns "the structure does not match" into a
	// one-look diagnosis.
	s.Assert().Contains(errs[0].Error(), "Sid")
}

func (s *MappingGuardSuite) Test_reports_link_data_missing_at_the_mapped_path() {
	errs := ValidateResourceDataMappings(s.mappings(), core.MappingNodeFields())

	s.Require().Len(errs, 1)
	s.Assert().Contains(errs[0].Error(), "holds nothing")
}

func (s *MappingGuardSuite) Test_ignores_a_mapping_that_selects_on_no_field() {
	// A plain path makes no promise about the shape of what it points at, so there
	// is nothing here to break. Composition reports an absent one separately.
	mappings := map[string]string{
		"someFunc_lambda_func::spec.vpcConfig.subnetIds": `["someRole"].subnetIds`,
	}

	s.Assert().Empty(ValidateResourceDataMappings(mappings, core.MappingNodeFields()))
}

func (s *MappingGuardSuite) Test_checks_every_selector_in_a_path() {
	// The inline statement path selects twice: the policy by name and the
	// statement by sid. Both have to be satisfiable.
	mappings := map[string]string{
		InlineAccessStatementPath(roleName, testSID): PermissionFieldPath("role"),
	}
	// policyName is selected on the target resource, not the contribution, so only
	// sid is required of the link data.
	linkData := core.MappingNodeFields(
		"role", core.MappingNodeFields(PermissionFieldName, statementNode("sid")),
	)

	s.Assert().Empty(ValidateResourceDataMappings(mappings, linkData))
}

func (s *MappingGuardSuite) Test_reports_every_broken_mapping_not_just_the_first() {
	// Aborting on the first would have hidden that 37 links shared the fault.
	mappings := map[string]string{
		InlineAccessStatementPath("roleA", "sidA"): PermissionFieldPath("a"),
		InlineAccessStatementPath("roleB", "sidB"): PermissionFieldPath("b"),
	}
	linkData := core.MappingNodeFields(
		"a", core.MappingNodeFields(PermissionFieldName, statementNode("Sid")),
		"b", core.MappingNodeFields(PermissionFieldName, statementNode("Sid")),
	)

	s.Assert().Len(ValidateResourceDataMappings(mappings, linkData), 2)
}

func (s *MappingGuardSuite) Test_accepts_a_managed_policy_arn_mapping() {
	// The overflow placement maps an ARN by value rather than by field, so it has
	// no field selector and nothing to satisfy.
	arn := "arn:aws:iam::123456789012:policy/celerity-link-access-1"
	mappings := map[string]string{
		ManagedAccessArnPath(roleName, arn): ManagedPolicyArnFieldPath("role"),
	}
	linkData := core.MappingNodeFields(
		"role", core.MappingNodeFields(
			ManagedPolicyArnFieldName, core.MappingNodeFromString(arn),
		),
	)

	s.Assert().Empty(ValidateResourceDataMappings(mappings, linkData))
}

func (s *MappingGuardSuite) Test_handles_no_mappings() {
	s.Assert().Empty(ValidateResourceDataMappings(map[string]string{}, core.MappingNodeFields()))
}

func TestMappingGuardSuite(t *testing.T) {
	suite.Run(t, new(MappingGuardSuite))
}
