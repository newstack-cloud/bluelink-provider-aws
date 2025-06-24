package lambda

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *lambdaLayerVersionResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// Layer versions are immutable once created, so they're immediately stable
	// after successful creation. No polling or waiting is required.
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
