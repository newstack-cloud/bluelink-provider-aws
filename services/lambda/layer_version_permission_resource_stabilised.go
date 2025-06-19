package lambda

import (
	"context"

	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaLayerVersionPermissionResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// Layer version permissions are immediately stabilized once created
	// There's no async operation that needs to be waited for
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
