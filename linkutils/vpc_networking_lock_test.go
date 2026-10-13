//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

type NetworkingLockSuite struct {
	suite.Suite
}

func TestNetworkingLockSuite(t *testing.T) {
	suite.Run(t, new(NetworkingLockSuite))
}

// Narrowing the lock key to the endpoint was not enough on its own. Every handler
// reaching the same service still queues on that one key, and a holder that kept the
// lock for the rest of its phase made the queue as long as the phase rather than as
// long as the endpoint work. Callers waiting on a popular endpoint then timed out on a
// holder that had finished with it long before.
func (s *NetworkingLockSuite) Test_endpoint_lock_is_released_once_the_endpoint_is_reconciled() {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
			RouteTables: []ec2types.RouteTable{
				{
					RouteTableId: aws.String("rtb-1"),
					Associations: []ec2types.RouteTableAssociation{
						{SubnetId: aws.String("subnet-1")},
					},
				},
			},
		}),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		ec2mock.WithCreateVpcEndpointOutput(&ec2.CreateVpcEndpointOutput{
			VpcEndpoint: &ec2types.VpcEndpoint{VpcEndpointId: aws.String("vpce-s3")},
		}),
		ec2mock.WithDescribeManagedPrefixListsOutput(gatewayPrefixListOutput()),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		gatewayActivation(),
		&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
	)

	s.Require().NoError(err)
	s.Equal(
		[]string{"appVpc::endpoint::Gateway::s3"},
		rs.ReleasedLocks,
		"the endpoint must be given up as soon as it is reconciled",
	)
}

// A reconciliation that failed has finished with the endpoint just as one
// that succeeded, and holding the lock until the phase ends would keep it for the rest
// of a deployment that is already failing.
func (s *NetworkingLockSuite) Test_endpoint_lock_is_released_when_reconciliation_fails() {
	reconcileErr := errors.New("rate exceeded")
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
			RouteTables: []ec2types.RouteTable{
				{
					RouteTableId: aws.String("rtb-1"),
					Associations: []ec2types.RouteTableAssociation{
						{SubnetId: aws.String("subnet-1")},
					},
				},
			},
		}),
		ec2mock.WithDescribeVpcEndpointsError(reconcileErr),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		gatewayActivation(),
		&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
	)

	s.Require().Error(err)
	s.Equal(
		[]string{"appVpc::endpoint::Gateway::s3"},
		rs.ReleasedLocks,
		"a failed reconciliation must not keep the endpoint locked",
	)
}

// The caller has to act on what went wrong with the endpoint, not on a release that
// failed; an unreleased lock is cleared with the phase.
func (s *NetworkingLockSuite) Test_reconciliation_error_wins_over_a_release_error() {
	reconcileErr := errors.New("rate exceeded")
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
			RouteTables: []ec2types.RouteTable{
				{
					RouteTableId: aws.String("rtb-1"),
					Associations: []ec2types.RouteTableAssociation{
						{SubnetId: aws.String("subnet-1")},
					},
				},
			},
		}),
		ec2mock.WithDescribeVpcEndpointsError(reconcileErr),
	)
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
		resourceservicemock.WithReleaseResourceLockError(errors.New("lock not held")),
	)

	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		gatewayActivation(),
		&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
	)

	s.Require().Error(err)
	s.Contains(err.Error(), reconcileErr.Error())
}

// The lock a reconciliation actually takes, read back from the release it performs.
func (s *NetworkingLockSuite) releasedLockFor(
	activation NetworkingActivation,
	ec2Svc *ec2mock.Ec2ServiceMock,
) []string {
	rs := resourceservicemock.Create(
		resourceservicemock.WithLookupResourceInState(flexVPCStateForNetworking()),
	)

	_, err := ReconcileLinkNetworking(
		context.Background(),
		ec2Svc,
		networkingInput(provider.LinkUpdateTypeCreate, rs),
		activation,
		&provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
	)
	s.Require().NoError(err)

	return rs.ReleasedLocks
}

func sgPairMock() *ec2mock.Ec2ServiceMock {
	return ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
}

func gatewayEndpointMock(endpointID string) *ec2mock.Ec2ServiceMock {
	return ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeRouteTablesOutput(&ec2.DescribeRouteTablesOutput{
			RouteTables: []ec2types.RouteTable{
				{
					RouteTableId: aws.String("rtb-1"),
					Associations: []ec2types.RouteTableAssociation{
						{SubnetId: aws.String("subnet-1")},
					},
				},
			},
		}),
		ec2mock.WithDescribeVpcEndpointsOutput(&ec2.DescribeVpcEndpointsOutput{}),
		ec2mock.WithCreateVpcEndpointOutput(&ec2.CreateVpcEndpointOutput{
			VpcEndpoint: &ec2types.VpcEndpoint{VpcEndpointId: aws.String(endpointID)},
		}),
		ec2mock.WithDescribeManagedPrefixListsOutput(gatewayPrefixListOutput()),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
}

// Every VPC-attached link reconciles its networking here, and locking the flex VPC put
// all of them in one queue behind endpoint creation that runs for minutes. Because the
// engine holds a link's locks until its stage ends, a link queued here also held its
// execution role, and links wanting that role timed out waiting on a holder that was
// itself waiting.
//
// So the property that matters is not that a lock is taken, but that two links with
// nothing in common take different ones.
func (s *NetworkingLockSuite) Test_endpoints_for_different_services_take_different_locks() {
	s.Equal(
		[]string{"appVpc::endpoint::Gateway::s3"},
		s.releasedLockFor(gatewayActivation(), gatewayEndpointMock("vpce-s3")),
	)

	dynamodb := gatewayActivation()
	dynamodb.AWSService = "dynamodb"
	s.Equal(
		[]string{"appVpc::endpoint::Gateway::dynamodb"},
		s.releasedLockFor(dynamodb, gatewayEndpointMock("vpce-ddb")),
		"endpoints for different services do not conflict and must not serialise",
	)
}

// Security group pairing writes to the groups it names, so links pairing different groups
// are independent, and the order the target's ids arrive in is not significant: two links
// naming the same groups have to agree on the key or both would write the same rules.
func (s *NetworkingLockSuite) Test_security_group_pairing_is_keyed_on_the_groups_it_writes() {
	pair := sgPairActivation()
	pair.TargetSecurityGroupIDs = []string{"sg-author-owned", "sg-db"}

	reversed := sgPairActivation()
	reversed.TargetSecurityGroupIDs = []string{"sg-db", "sg-author-owned"}

	s.Equal(
		[]string{"appVpc::sgpair::sg-author-owned,sg-db"},
		s.releasedLockFor(pair, sgPairMock()),
	)
	s.Equal(
		s.releasedLockFor(pair, sgPairMock()),
		s.releasedLockFor(reversed, sgPairMock()),
		"the key must not depend on the order the group ids are given in",
	)

	// A different pair is different work, and pairing never shares a key with endpoint
	// work.
	s.Equal(
		[]string{"appVpc::sgpair::sg-db"},
		s.releasedLockFor(sgPairActivation(), sgPairMock()),
	)
}

// With nothing identifiable to narrow to, the VPC is still the safe key, an unrecognised
// activation that quietly took no lock would let concurrent callers race.
func (s *NetworkingLockSuite) Test_an_activation_with_nothing_to_narrow_to_locks_the_vpc() {
	activation := vpcAttachedActivation()
	activation.AWSService = ""
	activation.EndpointType = ""
	activation.TargetSecurityGroupIDs = nil

	s.Equal(
		[]string{"appVpc"},
		s.releasedLockFor(activation, ec2mock.CreateEc2ServiceMock(
			ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		)),
	)
}
