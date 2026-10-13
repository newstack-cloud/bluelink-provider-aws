package lambdalinks

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// ProduceResourceContributions states the environment variable the calling function needs
// in order to invoke the function it is linked to.
//
// Contributed rather than written: the framework applies every contributing link's
// variables to the function as one update, so writing this one here would leave it out of
// the spec the framework deploys and the next deployment of the function would remove it.
func (l *lambdaFunctionFunctionLinkActions) ProduceResourceContributions(
	ctx context.Context,
	input *provider.LinkProduceResourceContributionsInput,
) (*provider.LinkProduceResourceContributionsOutput, error) {
	annotations := getLambdaFunctionLinkAnnotations(
		input.ResourceAInfo,
		input.ResourceBInfo,
	)
	if !annotations.populateEnvVars {
		// Nothing contributed retracts the variable this link contributed before: a
		// contribution is desired state, so withdrawing it is how the variable is removed.
		return &provider.LinkProduceResourceContributionsOutput{}, nil
	}

	otherFunctionARN, hasOtherFunctionARN := utils.ExtractARNFromResourceInfo(
		input.ResourceBInfo,
	)
	if !hasOtherFunctionARN {
		return nil, fmt.Errorf(
			"function ARN could not be retrieved from the linked to %q function resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	return &provider.LinkProduceResourceContributionsOutput{
		Contributions: []*provider.ResourceContribution{
			{
				ResourceName: pluginutils.GetResourceName(input.ResourceAInfo),
				FieldPath: fmt.Sprintf(
					"spec.environment.variables[\"%s\"]",
					invokeLambdaFunctionEnvVarName(
						annotations.envVarName,
						input.ResourceBInfo,
					),
				),
				Value: core.MappingNodeFromString(otherFunctionARN),
				// One variable holds one function ARN. Two links naming the same variable
				// is a disagreement the framework reports rather than a list to add to.
				Action: provider.ContributionActionSet,
			},
		},
	}, nil
}
