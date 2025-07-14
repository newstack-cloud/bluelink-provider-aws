package iam

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (i *iamOidcProviderResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// IAM OIDC providers are created synchronously and are immediately available
	// so will always be stable after a successful create or update operation.
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}