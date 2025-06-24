package lambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func lambdaAliasResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaAliasDefinition",
		Description: "The definition of an AWS Lambda function alias.",
		Required:    []string{"functionName", "name", "functionVersion"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"functionName": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The name or ARN of the Lambda function.",
				Pattern:      "(arn:(aws[a-zA-Z-]*)?:lambda:)?([a-z]{2}(-gov)?-[a-z]+-\\d{1}:)?(\\d{12}:)?(function:)?([a-zA-Z0-9-_]+)(:(\\$LATEST|[a-zA-Z0-9-_]+))?",
				MinLength:    1,
				MaxLength:    140,
				MustRecreate: true,
			},
			"name": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The name of the alias.",
				Pattern:      "(?!^[0-9]+$)([a-zA-Z0-9-_]+)",
				MinLength:    1,
				MaxLength:    128,
				MustRecreate: true,
			},
			"functionVersion": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The function version that the alias invokes.",
				Pattern:     "(\\$LATEST|[0-9]+)",
				MinLength:   1,
				MaxLength:   1024,
			},
			"description": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "A description of the alias.",
				MinLength:   0,
				MaxLength:   256,
			},
			"routingConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "AliasRoutingConfiguration",
				Description: "The routing configuration of the alias.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"additionalVersionWeights": {
						Type: provider.ResourceDefinitionsSchemaTypeMap,
						MapValues: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeFloat,
						},
						Description: "The second version, and the percentage of traffic that's routed to it. " +
							"The key is the version number and the value is the percentage of traffic.",
					},
				},
			},
			"provisionedConcurrencyConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "ProvisionedConcurrencyConfiguration",
				Description: "Specifies a provisioned concurrency configuration for a function's alias.",
				Required:    []string{"provisionedConcurrentExecutions"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"provisionedConcurrentExecutions": {
						Type:        provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "The amount of provisioned concurrency to allocate for the alias.",
						Minimum:     core.ScalarFromInt(1),
					},
				},
			},
			// Computed fields returned by AWS
			"aliasArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the alias.",
				Computed:    true,
			},
		},
	}
}
