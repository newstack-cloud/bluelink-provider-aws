package iam

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (i *iamInstanceProfileResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// IAM instance profiles are created synchronously and are immediately available
	// so will always be stable after a successful create or update operation.
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
