package lambdassm

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// ProduceResourceContributions states the environment variable the function needs in order
// to address the SSM parameter tree it is linked to.
//
// Declared here rather than inherited from the parameter link this one embeds. The
// embedded link contributes a single parameter's name, and a tree link contributing that
// would put the wrong value on the function.
func (l *functionParameterTreeLinkActions) ProduceResourceContributions(
	ctx context.Context,
	input *provider.LinkProduceResourceContributionsInput,
) (*provider.LinkProduceResourceContributionsOutput, error) {
	annotations := getParameterLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	if !annotations.populateEnvVars {
		return &provider.LinkProduceResourceContributionsOutput{}, nil
	}

	// The environment variable holds the path prefix, since the runtime enumerates the
	// tree via ssm:GetParametersByPath(path).
	path, hasPath := extractParameterPath(input.ResourceBInfo)
	if !hasPath {
		return nil, fmt.Errorf(
			"parameter path could not be retrieved from the linked to %q SSM parameter tree resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	return &provider.LinkProduceResourceContributionsOutput{
		Contributions: []*provider.ResourceContribution{
			{
				ResourceName: pluginutils.GetResourceName(input.ResourceAInfo),
				FieldPath: fmt.Sprintf(
					"spec.environment.variables[\"%s\"]",
					parameterPathEnvVarName(annotations.envVarName, input.ResourceBInfo),
				),
				Value:  core.MappingNodeFromString(path),
				Action: provider.ContributionActionSet,
			},
		},
	}, nil
}
