//go:build unit

package rdsrds

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/rds_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type ProxyInstanceLinkTestSuite struct {
	suite.Suite
}

func TestProxyInstanceLinkTestSuite(t *testing.T) {
	suite.Run(t, new(ProxyInstanceLinkTestSuite))
}

func registeringRDSMock() *rds.RegisterDBProxyTargetsOutput {
	return &rds.RegisterDBProxyTargetsOutput{}
}

// The rule pairs the proxy's own group with the one group on the instance that the flex
// VPC provisioned; the author's own group is left alone, since pairing against it could open
// a path to everything else sharing it.
func (s *ProxyInstanceLinkTestSuite) Test_opens_the_path_from_the_proxy_to_the_instance() {
	ec2Svc := pairingEC2Mock()
	rdsSvc := rdsmock.CreateRDSServiceMock(
		rdsmock.WithDescribeDBProxyTargetsOutput(&rds.DescribeDBProxyTargetsOutput{}),
		rdsmock.WithRegisterDBProxyTargetsOutput(registeringRDSMock()),
	)

	_, err := runUpdate(
		provider.LinkUpdateTypeCreate,
		proxyInfo(proxySpec(), nil),
		instanceInfo(instanceSpec(core.MappingNodeFields("port", core.MappingNodeFromInt(5433)))),
		ec2Svc,
		rdsSvc,
	)

	s.Require().NoError(err)
	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupIngress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupIngressInput)
		return ok && aws.ToString(in.GroupId) == testInstanceSGID &&
			len(in.IpPermissions) == 1 &&
			aws.ToInt32(in.IpPermissions[0].FromPort) == 5433 &&
			aws.ToString(in.IpPermissions[0].UserIdGroupPairs[0].GroupId) == testProxySGID
	})
	ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupEgress", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
		return ok && aws.ToString(in.GroupId) == testProxySGID
	})
}

// The port the rule opens. An annotation if the author set one, then the instance's own
// port from state (top-level or on its endpoint), and otherwise a default. The rule has
// to open something even before the instance reports a port, or the first deploy opens
// nothing and the target waits on a health it never reaches.
func (s *ProxyInstanceLinkTestSuite) Test_port_the_rule_opens() {
	cases := []struct {
		name         string
		instance     *core.MappingNode
		annotations  map[string]*core.MappingNode
		expectedPort int32
	}{
		{
			name:         "the instance's own port",
			instance:     instanceSpec(core.MappingNodeFields("port", core.MappingNodeFromInt(5433))),
			expectedPort: 5433,
		},
		{
			name: "the port on the instance's endpoint",
			instance: instanceSpec(core.MappingNodeFields(
				"endpoint", core.MappingNodeFields("port", core.MappingNodeFromInt(3306)),
			)),
			expectedPort: 3306,
		},
		{
			name:     "an annotation, which wins over the instance state",
			instance: instanceSpec(core.MappingNodeFields("port", core.MappingNodeFromInt(5433))),
			annotations: map[string]*core.MappingNode{
				"aws.rds.dbProxy." + testInstanceName + ".port": core.MappingNodeFromInt(6000),
			},
			expectedPort: 6000,
		},
		{
			name:         "the default when the instance has no port yet",
			instance:     instanceSpec(),
			expectedPort: defaultDatabasePort,
		},
	}

	for _, testCase := range cases {
		s.Run(testCase.name, func() {
			ec2Svc := pairingEC2Mock()
			rdsSvc := rdsmock.CreateRDSServiceMock(
				rdsmock.WithDescribeDBProxyTargetsOutput(&rds.DescribeDBProxyTargetsOutput{}),
				rdsmock.WithRegisterDBProxyTargetsOutput(registeringRDSMock()),
			)

			_, err := runUpdate(
				provider.LinkUpdateTypeCreate,
				proxyInfo(proxySpec(), testCase.annotations),
				instanceInfo(testCase.instance),
				ec2Svc,
				rdsSvc,
			)

			s.Require().NoError(err)
			ec2Svc.AssertCalledWith(&s.Suite, "AuthorizeSecurityGroupIngress", 0, plugintestutils.Any, func(arg any) bool {
				in, ok := arg.(*ec2.AuthorizeSecurityGroupIngressInput)
				return ok && len(in.IpPermissions) == 1 &&
					aws.ToInt32(in.IpPermissions[0].FromPort) == testCase.expectedPort
			})
		})
	}
}

// A proxy always sits in a VPC, so incomplete networking means its state has not caught
// up rather than that it has none. Reporting it beats opening nothing and leaving the
// target to time out with no explanation.
func (s *ProxyInstanceLinkTestSuite) Test_incomplete_proxy_networking_is_reported() {
	cases := map[string]*core.MappingNode{
		"no vpc id": core.MappingNodeFields(
			"dbProxyName", core.MappingNodeFromString(testProxyName),
			"vpcSubnetIds", core.MappingNodeItems(core.MappingNodeFromString(testSubnetID)),
			"vpcSecurityGroupIds", core.MappingNodeItems(core.MappingNodeFromString(testProxySGID)),
		),
		"no subnets": core.MappingNodeFields(
			"dbProxyName", core.MappingNodeFromString(testProxyName),
			"vpcId", core.MappingNodeFromString(testVPCID),
			"vpcSecurityGroupIds", core.MappingNodeItems(core.MappingNodeFromString(testProxySGID)),
		),
		"no security groups": core.MappingNodeFields(
			"dbProxyName", core.MappingNodeFromString(testProxyName),
			"vpcId", core.MappingNodeFromString(testVPCID),
			"vpcSubnetIds", core.MappingNodeItems(core.MappingNodeFromString(testSubnetID)),
		),
		"empty state": core.MappingNodeFields(),
	}

	for name, spec := range cases {
		s.Run(name, func() {
			_, err := runUpdate(
				provider.LinkUpdateTypeCreate,
				proxyInfo(spec, nil),
				instanceInfo(instanceSpec()),
				pairingEC2Mock(),
				rdsmock.CreateRDSServiceMock(),
			)

			s.Require().Error(err)
			s.Contains(err.Error(), "VPC networking could not be retrieved")
		})
	}
}

// The two resource types name the field differently: AWS::RDS::DBInstance uses
// VPCSecurityGroups and AWS::RDS::DBProxy uses vpcSecurityGroupIds. Reading the proxy's
// name off the instance found nothing and reported an instance that has a security group
// as having none, which failed the link on a real deployment.
func (s *ProxyInstanceLinkTestSuite) Test_instance_without_security_groups_is_reported() {
	cases := map[string]*core.MappingNode{
		"no groups at all": core.MappingNodeFields(
			"dbInstanceIdentifier", core.MappingNodeFromString(testInstanceID),
		),
		"groups under the proxy's field name": core.MappingNodeFields(
			"dbInstanceIdentifier", core.MappingNodeFromString(testInstanceID),
			"vpcSecurityGroupIds", core.MappingNodeItems(
				core.MappingNodeFromString(testInstanceSGID),
			),
		),
	}

	for name, spec := range cases {
		s.Run(name, func() {
			_, err := runUpdate(
				provider.LinkUpdateTypeCreate,
				proxyInfo(proxySpec(), nil),
				instanceInfo(spec),
				pairingEC2Mock(),
				rdsmock.CreateRDSServiceMock(),
			)

			s.Require().Error(err)
			s.Contains(err.Error(), "security group could not be retrieved")
		})
	}
}
