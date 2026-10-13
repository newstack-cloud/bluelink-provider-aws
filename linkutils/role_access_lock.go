package linkutils

import (
	"context"

	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// RoleAccessLock identifies the execution role lock a grant is made under.
type RoleAccessLock struct {
	ResourceService provider.ResourceService
	// InstanceID is the blueprint instance the role and the lock belong to.
	InstanceID string
	// LinkID identifies the caller, so the release matches the acquisition.
	LinkID string
	// RoleResourceName is the role's name in the blueprint, which is what the
	// lock is keyed on.
	RoleResourceName string
	ProviderContext  provider.Context
}

// RoleAccessLockFor builds the lock descriptor from the pieces a link has to hand, rather
// than from a link input, so it is usable from every phase (whose input types differ).
//
// resourceInfo only supplies the blueprint instance, so either side of the link serves;
// callers pass whichever they already hold.
func RoleAccessLockFor(
	resourceService provider.ResourceService,
	linkID string,
	resourceInfo *provider.ResourceInfo,
	setupCtx *LambdaLinkSetupContext,
	providerCtx provider.Context,
) RoleAccessLock {
	return RoleAccessLock{
		ResourceService:  resourceService,
		InstanceID:       pluginutils.GetInstanceID(resourceInfo),
		LinkID:           linkID,
		RoleResourceName: setupCtx.RoleResourceName,
		ProviderContext:  providerCtx,
	}
}

// GrantRoleAccessUnderLock applies a role access grant while holding the role's lock, and
// gives the lock up as soon as the grant is applied.
//
// An execution role is shared as several links grant access to the same one, and the lock is
// what keeps their read-modify-write of its policies apart.
// This avoids holding a lock on an entire role that would serialise every role update,
// slowing down a deployment significantly.
//
// The lock is released whether or not the grant succeeded.
func GrantRoleAccessUnderLock(
	ctx context.Context,
	lock RoleAccessLock,
	iamService iamservice.Service,
	grant RoleAccessGrant,
) (RoleAccessResult, error) {
	err := lock.ResourceService.AcquireResourceLock(
		ctx,
		&provider.AcquireResourceLockInput{
			InstanceID:      lock.InstanceID,
			ResourceName:    lock.RoleResourceName,
			ProviderContext: lock.ProviderContext,
			AcquiredBy:      lock.LinkID,
		},
	)
	if err != nil {
		return RoleAccessResult{}, err
	}

	result, grantErr := ReconcileRoleAccessPolicy(ctx, iamService, grant)

	releaseErr := lock.ResourceService.ReleaseResourceLock(
		ctx,
		&provider.ReleaseResourceLockInput{
			InstanceID:      lock.InstanceID,
			ResourceName:    lock.RoleResourceName,
			ProviderContext: lock.ProviderContext,
			AcquiredBy:      lock.LinkID,
		},
	)

	// The grant's own error wins as a release that failed leaves the lock to be
	// cleared with the phase.
	if grantErr != nil {
		return result, grantErr
	}

	return result, releaseErr
}
