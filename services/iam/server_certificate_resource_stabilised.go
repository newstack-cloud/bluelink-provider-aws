package iam

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (a *iamServerCertificateResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// Server certificates are always considered stable straight after creation.
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
