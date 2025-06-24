package lambda

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

func lambdaAliasDataSourceSchema() map[string]*provider.DataSourceSpecSchema {
	return map[string]*provider.DataSourceSpecSchema{
		"arn": {
			Label:       "Alias ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the Lambda alias.",
			Nullable:    false,
		},
		"description": {
			Label:       "Description",
			Type:        provider.DataSourceSpecTypeString,
			Description: "Description of the alias.",
			Nullable:    true,
		},
		"functionName": {
			Label:       "Function Name",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The name of the Lambda function.",
			Nullable:    false,
		},
		"functionVersion": {
			Label:       "Function Version",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The function version that the alias invokes.",
			Nullable:    false,
		},
		"invokeArn": {
			Label:       "Invoke ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The invocation ARN of the alias to be used when invoking from API Gateway.",
			Nullable:    false,
		},
		"name": {
			Label:       "Alias Name",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The name of the alias.",
			Nullable:    false,
		},
		"routingConfig.additionalVersionWeights": {
			Label:       "Additional Version Weights",
			Type:        provider.DataSourceSpecTypeString,
			Description: "A JSON-encoded string of the additional version weights for the alias.",
			Nullable:    true,
		},
	}
}
