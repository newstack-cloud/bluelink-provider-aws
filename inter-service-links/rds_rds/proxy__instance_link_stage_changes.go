package rdsrds

import (
	"context"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// StageChanges reports what this link will do at deploy time.
//
// Neither resource's spec is modified, so there are no field changes to stage. The
// link's whole effect is a pair of security group rules, which are reconciled against
// live networking at deploy and so can only be surfaced as known-on-deploy.
func (l *proxyInstanceLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	linkutils.StageNetworkAccessKnownOnDeploy(input.ResourceAChanges, changes)

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
