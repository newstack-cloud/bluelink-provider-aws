package lambda

import (
	"context"

	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *lambdaEventInvokeConfigResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// Event Invoke Config resources are immediately stable after creation/update
	// as they don't have any asynchronous configuration that needs to settle.
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
