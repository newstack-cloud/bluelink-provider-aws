package lambda

import (
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/blueprint/schema"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func extractRegionFromFilters(filters *provider.ResolvedDataSourceFilters) *core.MappingNode {
	for _, filter := range filters.Filters {
		if core.StringValueFromScalar(filter.Field) == "region" &&
			pluginutils.GetDataSourceFilterOperator(
				filter,
			) == schema.DataSourceFilterOperatorEquals {
			return pluginutils.GetDataSourceFilterSearchValue(filter, 0)
		}
	}

	return nil
}

func extractFunctionNameOrARNFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	return pluginutils.ExtractFirstMatchFromFilters(
		filters,
		[]string{"arn", "name"},
	)
}

func extractQualifierFromFilters(
	filters *provider.ResolvedDataSourceFilters,
) *core.MappingNode {
	return pluginutils.ExtractMatchFromFilters(
		filters,
		"qualifier",
	)
}
