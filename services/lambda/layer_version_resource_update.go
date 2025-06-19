package lambda

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaLayerVersionResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	// Layer versions are immutable - they cannot be updated.
	// Any changes to a layer version require creating a new version.
	// This should not be called due to MustRecreate: true on all fields in the schema.
	return nil, fmt.Errorf("layer versions are immutable and cannot be updated. " +
		"Any changes require creating a new layer version")
}
