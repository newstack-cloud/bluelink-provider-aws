package lambdaelasticache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// ProduceResourceContributions states the environment variables the function needs in
// order to reach the cache it is linked to.
//
// Contributed rather than written: the framework applies every contributing link's
// variables to the function as one update, so writing these here would leave them out of
// the spec the framework deploys and the next deployment of the function would remove them.
func (l *functionCacheLinkActions) ProduceResourceContributions(
	ctx context.Context,
	input *provider.LinkProduceResourceContributionsInput,
) (*provider.LinkProduceResourceContributionsOutput, error) {
	annotations := getCacheLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)
	if !annotations.populateEnvVars {
		// Nothing contributed retracts both variables this link contributed before: a
		// contribution is desired state, so withdrawing them is how they are removed.
		return &provider.LinkProduceResourceContributionsOutput{}, nil
	}

	endpoint, hasEndpoint := extractCacheEndpoint(input.ResourceBInfo)
	if !hasEndpoint {
		return nil, fmt.Errorf(
			"endpoint could not be retrieved from the linked to %q ElastiCache replication group resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	names := cacheEnvVarNames(
		cacheEnvVarPrefix(annotations.envVarPrefix, input.ResourceBInfo),
	)
	functionName := pluginutils.GetResourceName(input.ResourceAInfo)

	// Ordered by variable name so that the contributions a link produces do not depend on
	// map iteration, which is what keeps a redeploy from looking like a change.
	values := map[string]string{
		names.host: endpoint,
		names.port: strconv.Itoa(int(annotations.port)),
	}

	contributions := make([]*provider.ResourceContribution, 0, len(values))
	for _, name := range sortedKeys(values) {
		contributions = append(contributions, &provider.ResourceContribution{
			ResourceName: functionName,
			FieldPath: fmt.Sprintf(
				"spec.environment.variables[%q]",
				name,
			),
			Value: core.MappingNodeFromString(values[name]),
			// One variable holds one value. Two links naming the same variable is a
			// disagreement the framework reports rather than a list to add to.
			Action: provider.ContributionActionSet,
		})
	}

	return &provider.LinkProduceResourceContributionsOutput{
		Contributions: contributions,
	}, nil
}
