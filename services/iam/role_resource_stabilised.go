package iam

import (
	"context"

	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (i *iamRoleResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// IAM roles are typically available immediately after creation
	// Unlike Lambda functions which have states, IAM roles are stable once they exist.
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
