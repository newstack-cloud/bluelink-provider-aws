package lambda

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaLayerVersionPermissionResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	// Layer version permissions are immutable - they cannot be updated.
	// Any changes to a layer version permission require recreating it.
	// This should not be called due to MustRecreate: true on all fields in the schema.
	return nil, fmt.Errorf("layer version permissions are immutable and cannot be updated. " +
		"Any changes require recreating the permission")
}
