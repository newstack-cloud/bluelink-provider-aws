//go:build unit

package overlays

import (
	"context"
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionOverlaySuite struct {
	suite.Suite
}

// The settle predicate is only reached if the overlay is registered under exactly the
// blueprint type the generated resource declares. A mismatch is silent: BehaviourFor
// returns nil, settledAfterLinkUpdate reports the resource settled, and links go back to
// colliding with no error anywhere to say why.
func (s *LambdaFunctionOverlaySuite) Test_settle_predicate_is_registered_for_the_generated_blueprint_type() {
	behaviour := BehaviourFor("aws/lambda/function")

	s.Require().NotNil(behaviour, "no behaviour overlay registered for aws/lambda/function")
	s.Require().NotNil(
		behaviour.SettledAfterLinkUpdate,
		"the lambda function overlay has no SettledAfterLinkUpdate predicate",
	)
}

// Exercised directly because the predicate builds its own Lambda client from the
// aws.Config it is handed, leaving no seam to inject a mock service: everything past the
// identifier needs a real API call. The paths that stop before that call are covered
// through the registered behaviour below.
func (s *LambdaFunctionOverlaySuite) Test_identifier_is_read_from_the_fields_the_spec_actually_carries() {
	cases := []struct {
		name     string
		spec     *core.MappingNode
		expected string
	}{
		{
			name: "arn, which the generated spec computes on deploy",
			spec: core.MappingNodeFields(
				"arn",
				core.MappingNodeFromString(
					"arn:aws:lambda:eu-west-2:123456789012:function:OrdersController-create",
				),
			),
			expected: "arn:aws:lambda:eu-west-2:123456789012:function:OrdersController-create",
		},
		{
			name: "functionName, the primary identifier field",
			spec: core.MappingNodeFields(
				"functionName",
				core.MappingNodeFromString("OrdersController-create"),
			),
			expected: "OrdersController-create",
		},
		{
			name: "arn wins when both are present",
			spec: core.MappingNodeFields(
				"functionName",
				core.MappingNodeFromString("OrdersController-create"),
				"arn",
				core.MappingNodeFromString(
					"arn:aws:lambda:eu-west-2:123456789012:function:OrdersController-create",
				),
			),
			expected: "arn:aws:lambda:eu-west-2:123456789012:function:OrdersController-create",
		},
	}

	for _, testCase := range cases {
		s.Run(testCase.name, func() {
			identifier := lambdaFunctionIdentifier(&provider.ResourceHasStabilisedInput{
				ResourceSpec: testCase.spec,
			})

			s.Equal(testCase.expected, identifier)
		})
	}
}

// A spec with no identifier in it is the silent path: with nothing to look up there is
// nothing to wait for, so the predicate reports settled and leaves the resource locks to
// keep links apart. It must reach that answer rather than call AWS with an empty name,
// which is why this goes through the registered behaviour with no service available.
func (s *LambdaFunctionOverlaySuite) Test_settled_when_the_spec_carries_no_identifier() {
	cases := map[string]*core.MappingNode{
		"neither field": core.MappingNodeFields(
			"runtime",
			core.MappingNodeFromString("nodejs24.x"),
		),
		"nil spec": nil,
	}

	for name, spec := range cases {
		s.Run(name, func() {
			settled, err := BehaviourFor("aws/lambda/function").SettledAfterLinkUpdate(
				context.Background(),
				&provider.ResourceHasStabilisedInput{ResourceSpec: spec},
				nil,
			)

			s.Require().NoError(err)
			s.True(settled)
		})
	}
}

// A VPC attach makes Lambda check the execution role's ENI permissions at that moment,
// and the link granted them immediately beforehand. IAM is eventually consistent, so the
// check can fail for a permission that does exist, and reporting that as terminal fails a
// deployment that would have succeeded on the next attempt.
//
// Classification is exercised directly for the same reason as the identifier above: it is
// only reachable past a GetFunction call the predicate makes with a client it builds
// itself.
func (s *LambdaFunctionOverlaySuite) Test_a_transient_update_failure_is_reported_as_retryable() {
	reasons := []string{
		"The function's execution role doesn't have permission to perform this operation.",
		"The provided execution role does not have permissions to call CreateNetworkInterface on EC2",
		"Cannot access stream arn:aws:kinesis:eu-west-2:123456789012:stream/orders",
	}

	for _, reason := range reasons {
		s.Run(reason, func() {
			s.True(isTransientUpdateFailureReason(reason))
		})
	}
}

func (s *LambdaFunctionOverlaySuite) Test_a_genuine_update_failure_is_not_reported_as_retryable() {
	reasons := []string{
		"The provided subnets do not have an available IP address",
		"InvalidParameterValueException: Uploaded file must be a non-empty zip",
	}

	for _, reason := range reasons {
		s.Run(reason, func() {
			s.False(isTransientUpdateFailureReason(reason))
		})
	}
}

// The declaration exists so drift and change staging stop reporting a reordered subnet
// list as a change; without it every VPC-attached function drifts permanently.
func (s *LambdaFunctionOverlaySuite) Test_marks_vpc_lists_unordered() {
	schema := &provider.ResourceDefinitionsSchema{
		Type: provider.ResourceDefinitionsSchemaTypeObject,
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"vpcConfig": {
				Type: provider.ResourceDefinitionsSchemaTypeObject,
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"subnetIds":               {Type: provider.ResourceDefinitionsSchemaTypeArray},
					"securityGroupIds":        {Type: provider.ResourceDefinitionsSchemaTypeArray},
					"ipv6AllowedForDualStack": {Type: provider.ResourceDefinitionsSchemaTypeBoolean},
				},
			},
		},
	}

	out := Apply(lambdaFunctionType, schema)
	vpc := out.Attributes["vpcConfig"].Attributes

	s.Assert().True(vpc["subnetIds"].IgnoreItemOrder, "subnetIds should be declared unordered")
	s.Assert().True(
		vpc["securityGroupIds"].IgnoreItemOrder,
		"securityGroupIds should be declared unordered",
	)
	s.Assert().False(
		vpc["ipv6AllowedForDualStack"].IgnoreItemOrder,
		"a scalar field should not be marked unordered",
	)
}

// A schema without vpcConfig must pass through rather than panic.
func (s *LambdaFunctionOverlaySuite) Test_tolerates_a_missing_vpc_config() {
	schema := &provider.ResourceDefinitionsSchema{
		Type:       provider.ResourceDefinitionsSchemaTypeObject,
		Attributes: map[string]*provider.ResourceDefinitionsSchema{},
	}

	s.Assert().Same(
		schema,
		Apply(lambdaFunctionType, schema),
		"the schema should pass through unchanged",
	)
}

func TestLambdaFunctionOverlaySuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionOverlaySuite))
}
