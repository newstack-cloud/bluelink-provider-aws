package lambda

import (
	"context"

	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaAliasResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	// Lambda aliases are generally stable immediately after creation/update
	// since they're just pointers to function versions
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
