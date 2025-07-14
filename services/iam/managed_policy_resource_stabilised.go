package iam

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (i *iamManagedPolicyResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// IAM managed policies are typically available immediately after creation
	// Unlike Lambda functions which have states, IAM policies are stable once they exist.
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
