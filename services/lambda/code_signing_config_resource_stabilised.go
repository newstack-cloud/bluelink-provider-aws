package lambda

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *lambdaCodeSigningConfigResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// Lambda code signing configurations are generally stable immediately after creation/update
	// since they're configuration resources that don't require provisioning
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
