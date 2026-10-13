package lambdalinks

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// ProduceResourceContributions states the code signing configuration the function is
// linked to.
//
// Contributed rather than written. Putting the ARN on the function through the Lambda API
// left the function's spec saying nothing about it, so the next deployment of the function
// for any other reason would state a spec without it and take the code signing
// configuration off. Stating it as a contribution puts it in the spec the framework
// deploys.
func (l *lambdaFunctionCodeSigningConfigLinkActions) ProduceResourceContributions(
	ctx context.Context,
	input *provider.LinkProduceResourceContributionsInput,
) (*provider.LinkProduceResourceContributionsOutput, error) {
	codeSigningConfigARN, hasARN := pluginutils.GetValueByPath(
		"$.codeSigningConfigArn",
		pluginutils.GetCurrentStateSpecDataFromResourceInfo(input.ResourceBInfo),
	)
	if !hasARN {
		return nil, fmt.Errorf(
			"code signing config ARN could not be retrieved from code signing config",
		)
	}

	return &provider.LinkProduceResourceContributionsOutput{
		Contributions: []*provider.ResourceContribution{
			{
				ResourceName: pluginutils.GetResourceName(input.ResourceAInfo),
				FieldPath:    "spec.codeSigningConfigArn",
				Value:        core.MappingNodeFromString(core.StringValue(codeSigningConfigARN)),
				// The function has one code signing configuration. Two links naming it is
				// a disagreement the framework reports rather than a list to add to.
				Action: provider.ContributionActionSet,
			},
		},
	}, nil
}
