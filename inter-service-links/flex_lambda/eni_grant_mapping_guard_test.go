//go:build unit

package flexlambda

import (
	"fmt"
	"testing"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/suite"
)

// Guards the promise the ENI grant makes about its own contribution: the mapping path it
// records selects the statement by @.sid, so the statement it puts in the link data has to
// carry a sid.
//
// The grant recorded it with IAM's JSON casing, "Sid", which was inert until composition
// began reading the mappings and then failed staging for every instance that had placed a
// function in a VPC. The existing test could not catch it: it built its expected value by
// calling specENIStatementNode, the function under test, so it passed under either casing.
type ENIGrantMappingGuardSuite struct {
	suite.Suite
}

const (
	guardRoleResourceName = "ordersExecutionRole"
	guardFunctionName     = "listOrdersFunction"
	guardSID              = "VPCNetworkInterfaceslistOrdersFunction"
)

// The link data and mappings the grant records for an inline placement, built the
// way the implementation builds them.
func (s *ENIGrantMappingGuardSuite) grantOutput() (map[string]string, *core.MappingNode) {
	linkDataKey := fmt.Sprintf("%sExecutionRole", guardFunctionName)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specENIStatementNode(guardSID),
	)

	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(
		mappings,
		roleLinkData,
		guardRoleResourceName,
		linkDataKey,
		guardSID,
		linkutils.RoleAccessResult{PlacedSlotInline: true},
	)

	return mappings, core.MappingNodeFields(linkDataKey, roleLinkData)
}

func (s *ENIGrantMappingGuardSuite) Test_every_mapping_selects_the_data_it_points_at() {
	mappings, linkData := s.grantOutput()

	s.Require().NotEmpty(mappings, "the inline placement must record a mapping")
	s.Assert().Empty(linkutils.ValidateResourceDataMappings(mappings, linkData))
}

func (s *ENIGrantMappingGuardSuite) Test_the_statement_uses_the_resource_spec_casing() {
	// Field names spelled out literally rather than taken from the builder, so putting the
	// capital back fails here. The IAM role spec schema declares these lowercase; the
	// capitalised form belongs only in the policy JSON sent to AWS.
	statement := specENIStatementNode(guardSID)

	s.Require().NotNil(statement.Fields)
	for _, field := range []string{"sid", "effect", "action", "resource"} {
		s.Assert().Contains(statement.Fields, field)
	}
	for _, field := range []string{"Sid", "Effect", "Action", "Resource"} {
		s.Assert().NotContains(statement.Fields, field)
	}
	s.Assert().Equal(guardSID, core.StringValue(statement.Fields["sid"]))
}

func TestENIGrantMappingGuardSuite(t *testing.T) {
	suite.Run(t, new(ENIGrantMappingGuardSuite))
}
