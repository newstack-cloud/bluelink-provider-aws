package lambdas3

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// ProduceResourceContributions states the environment variable the function needs in order
// to address the bucket it is linked to.
//
// Contributed rather than written: the framework applies every contributing link's
// variables to the function as one update, so writing this one here would leave it out of
// the spec the framework deploys and the next deployment of the function would remove it.
func (l *functionBucketLinkActions) ProduceResourceContributions(
	ctx context.Context,
	input *provider.LinkProduceResourceContributionsInput,
) (*provider.LinkProduceResourceContributionsOutput, error) {
	annotations := getBucketLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	if !annotations.populateEnvVars {
		// Nothing contributed retracts the variable this link contributed before: a
		// contribution is desired state, so withdrawing it is how the variable is removed.
		return &provider.LinkProduceResourceContributionsOutput{}, nil
	}

	bucketName, hasBucketName := extractBucketName(input.ResourceBInfo)
	if !hasBucketName {
		return nil, fmt.Errorf(
			"bucket name could not be retrieved from the linked to %q S3 bucket resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	return &provider.LinkProduceResourceContributionsOutput{
		Contributions: []*provider.ResourceContribution{
			{
				ResourceName: pluginutils.GetResourceName(input.ResourceAInfo),
				FieldPath: fmt.Sprintf(
					"spec.environment.variables[\"%s\"]",
					bucketEnvVarName(annotations.envVarName, input.ResourceBInfo),
				),
				Value: core.MappingNodeFromString(bucketName),
				// One variable holds one bucket name. Two links naming the same variable
				// is a disagreement the framework reports rather than a list to add to.
				Action: provider.ContributionActionSet,
			},
		},
	}, nil
}
