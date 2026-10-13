//go:build unit

package rdsrds

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	rdsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/rds_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type ProxyInstanceTargetTestSuite struct {
	suite.Suite
}

func TestProxyInstanceTargetTestSuite(t *testing.T) {
	suite.Run(t, new(ProxyInstanceTargetTestSuite))
}

func (s *ProxyInstanceTargetTestSuite) Test_registers_a_target_that_is_not_yet_present() {
	rdsSvc := rdsmock.CreateRDSServiceMock(
		rdsmock.WithDescribeDBProxyTargetsOutput(&rds.DescribeDBProxyTargetsOutput{}),
		rdsmock.WithRegisterDBProxyTargetsOutput(&rds.RegisterDBProxyTargetsOutput{}),
	)

	_, err := runUpdate(
		provider.LinkUpdateTypeCreate,
		proxyInfo(proxySpec(), nil),
		instanceInfo(instanceSpec()),
		pairingEC2Mock(),
		rdsSvc,
	)

	s.Require().NoError(err)
	rdsSvc.AssertCalledWith(&s.Suite, "RegisterDBProxyTargets", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*rds.RegisterDBProxyTargetsInput)
		return ok && aws.ToString(in.DBProxyName) == testProxyName &&
			len(in.DBInstanceIdentifiers) == 1 &&
			in.DBInstanceIdentifiers[0] == testInstanceID
	})
}

// A re-deploy, or a retry of this link, finds the target it registered last time and
// must not fail on its own work.
func (s *ProxyInstanceTargetTestSuite) Test_registration_is_skipped_when_already_registered() {
	rdsSvc := rdsmock.CreateRDSServiceMock(
		rdsmock.WithDescribeDBProxyTargetsOutput(&rds.DescribeDBProxyTargetsOutput{
			Targets: []rdstypes.DBProxyTarget{
				{RdsResourceId: aws.String(testOtherInstanceID)},
				{RdsResourceId: aws.String(testInstanceID)},
			},
		}),
	)

	_, err := runUpdate(
		provider.LinkUpdateTypeCreate,
		proxyInfo(proxySpec(), nil),
		instanceInfo(instanceSpec()),
		pairingEC2Mock(),
		rdsSvc,
	)

	s.Require().NoError(err)
	rdsSvc.AssertNotCalled(&s.Suite, "RegisterDBProxyTargets")
}

// Two links racing to register the same target both succeed as far as the caller is
// concerned; the loser simply finds the work already done.
func (s *ProxyInstanceTargetTestSuite) Test_already_registered_fault_is_not_an_error() {
	rdsSvc := rdsmock.CreateRDSServiceMock(
		rdsmock.WithDescribeDBProxyTargetsOutput(&rds.DescribeDBProxyTargetsOutput{}),
		rdsmock.WithRegisterDBProxyTargetsError(&rdstypes.DBProxyTargetAlreadyRegisteredFault{}),
	)

	_, err := runUpdate(
		provider.LinkUpdateTypeCreate,
		proxyInfo(proxySpec(), nil),
		instanceInfo(instanceSpec()),
		pairingEC2Mock(),
		rdsSvc,
	)

	s.Require().NoError(err)
}

// A proxy with no target group yet has nothing registered. That is the state this link
// exists to change, not a failure to report.
func (s *ProxyInstanceTargetTestSuite) Test_missing_target_group_is_treated_as_unregistered_instead_of_reporting_error() {
	rdsSvc := rdsmock.CreateRDSServiceMock(
		rdsmock.WithDescribeDBProxyTargetsError(&rdstypes.DBProxyTargetGroupNotFoundFault{}),
		rdsmock.WithRegisterDBProxyTargetsOutput(&rds.RegisterDBProxyTargetsOutput{}),
	)

	_, err := runUpdate(
		provider.LinkUpdateTypeCreate,
		proxyInfo(proxySpec(), nil),
		instanceInfo(instanceSpec()),
		pairingEC2Mock(),
		rdsSvc,
	)

	s.Require().NoError(err)
	rdsSvc.AssertCalled(&s.Suite, "RegisterDBProxyTargets")
}

func (s *ProxyInstanceTargetTestSuite) Test_registration_failure_is_reported() {
	rdsSvc := rdsmock.CreateRDSServiceMock(
		rdsmock.WithDescribeDBProxyTargetsOutput(&rds.DescribeDBProxyTargetsOutput{}),
		rdsmock.WithRegisterDBProxyTargetsError(errors.New("insufficient permissions")),
	)

	_, err := runUpdate(
		provider.LinkUpdateTypeCreate,
		proxyInfo(proxySpec(), nil),
		instanceInfo(instanceSpec()),
		pairingEC2Mock(),
		rdsSvc,
	)

	s.Require().Error(err)
	s.Contains(err.Error(), testProxyName)
	s.Contains(err.Error(), "insufficient permissions")
}

func (s *ProxyInstanceTargetTestSuite) Test_deregisters_on_destroy() {
	rdsSvc := rdsmock.CreateRDSServiceMock(
		rdsmock.WithDeregisterDBProxyTargetsOutput(&rds.DeregisterDBProxyTargetsOutput{}),
	)

	_, err := runUpdate(
		provider.LinkUpdateTypeDestroy,
		proxyInfo(proxySpec(), nil),
		instanceInfo(instanceSpec()),
		revokingEC2Mock(),
		rdsSvc,
	)

	s.Require().NoError(err)
	rdsSvc.AssertCalledWith(&s.Suite, "DeregisterDBProxyTargets", 0, plugintestutils.Any, func(arg any) bool {
		in, ok := arg.(*rds.DeregisterDBProxyTargetsInput)
		return ok && aws.ToString(in.DBProxyName) == testProxyName &&
			len(in.DBInstanceIdentifiers) == 1 &&
			in.DBInstanceIdentifiers[0] == testInstanceID
	})
	rdsSvc.AssertNotCalled(&s.Suite, "RegisterDBProxyTargets")
}

// When a whole blueprint comes down the instance or proxy is commonly destroyed before
// the link, so a target that is already gone is the ordinary case rather than a fault.
func (s *ProxyInstanceTargetTestSuite) Test_deregistering_something_already_gone_is_not_an_error() {
	for name, mockErr := range map[string]error{
		"proxy gone":        &rdstypes.DBProxyNotFoundFault{},
		"target group gone": &rdstypes.DBProxyTargetGroupNotFoundFault{},
		"target gone":       &rdstypes.DBProxyTargetNotFoundFault{},
	} {
		s.Run(name, func() {
			_, err := runUpdate(
				provider.LinkUpdateTypeDestroy,
				proxyInfo(proxySpec(), nil),
				instanceInfo(instanceSpec()),
				revokingEC2Mock(),
				rdsmock.CreateRDSServiceMock(
					rdsmock.WithDeregisterDBProxyTargetsError(mockErr),
				),
			)

			s.NoError(err)
		})
	}
}
