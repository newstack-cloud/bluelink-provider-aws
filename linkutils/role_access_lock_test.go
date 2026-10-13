//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

// The lock exists to keep two links out of the same role's policies at once, so it has to
// cover the grant. It does not have to cover anything else, and when it did the deployment
// paid for it: links go on to create event source mappings and attach functions to a VPC,
// and with one execution role shared by dozens of functions, every link wanting the role
// queued behind work it had no part in until it timed out.
type RoleAccessLockSuite struct {
	suite.Suite
}

const sharedRoleResourceName = "sharedExecutionRole"

func sharedRoleLock(resourceService provider.ResourceService) RoleAccessLock {
	return RoleAccessLock{
		ResourceService:  resourceService,
		InstanceID:       "instance-1",
		LinkID:           "link-a",
		RoleResourceName: sharedRoleResourceName,
	}
}

func (s *RoleAccessLockSuite) Test_releases_the_role_once_the_grant_is_applied() {
	rs := resourceservicemock.Create()

	_, err := GrantRoleAccessUnderLock(
		context.Background(),
		sharedRoleLock(rs),
		iammock.CreateIamServiceMock(
			iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
			iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
			iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
		),
		RoleAccessGrant{RoleName: "shared-role", SID: "sid-1"},
	)
	s.Require().NoError(err)

	s.Assert().Equal(
		[]string{sharedRoleResourceName},
		rs.ReleasedLocks,
		"the role must be given up as soon as the grant is applied",
	)
}

// A grant that failed has finished with the role just as surely as one that succeeded.
// Holding the lock until the phase ends would keep it for the rest of a deployment that is
// already losing.
func (s *RoleAccessLockSuite) Test_releases_the_role_when_the_grant_fails() {
	rs := resourceservicemock.Create()
	grantErr := errors.New("access denied")

	_, err := GrantRoleAccessUnderLock(
		context.Background(),
		sharedRoleLock(rs),
		iammock.CreateIamServiceMock(
			iammock.WithListRolePoliciesError(grantErr),
		),
		RoleAccessGrant{RoleName: "shared-role", SID: "sid-1"},
	)

	s.Require().Error(err, "the grant's error is what the caller has to act on")
	s.Assert().Equal([]string{sharedRoleResourceName}, rs.ReleasedLocks)
}

func TestRoleAccessLockSuite(t *testing.T) {
	suite.Run(t, new(RoleAccessLockSuite))
}
