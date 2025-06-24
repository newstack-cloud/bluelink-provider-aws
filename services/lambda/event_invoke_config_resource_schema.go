package lambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func lambdaEventInvokeConfigResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaEventInvokeConfigDefinition",
		Description: "The definition of an AWS Lambda Event Invoke Config.",
		Required:    []string{"functionName", "qualifier"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"functionName": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The name or ARN of the Lambda function, version, or alias.",
				Pattern:      "^(arn:(aws[a-zA-Z-]*)?:lambda:)?([a-z]{2}((-gov)|(-iso([a-z])?)?)?-[a-z]+-\\d{1}:)?(\\d{12}:)?(function:)?([a-zA-Z0-9-_]+)(:([a-zA-Z0-9$_-]+))?$",
				MinLength:    1,
				MaxLength:    140,
				MustRecreate: true,
			},
			"qualifier": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The identifier of a version or alias. Version numbers, $LATEST, or alias name.",
				Pattern:      "^(|[a-zA-Z0-9$_-]{1,129})$",
				MinLength:    1,
				MaxLength:    128,
				MustRecreate: true,
			},
			"maximumEventAgeInSeconds": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The maximum age of a request that Lambda sends to a function for processing.",
				Minimum:     core.ScalarFromInt(60),
				Maximum:     core.ScalarFromInt(21600),
			},
			"maximumRetryAttempts": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The maximum number of times to retry when the function returns an error.",
				Minimum:     core.ScalarFromInt(0),
				Maximum:     core.ScalarFromInt(2),
			},
			"destinationConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "DestinationConfig",
				Description: "A destination for events after they have been sent to a function for processing.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"onFailure": {
						Type:        provider.ResourceDefinitionsSchemaTypeObject,
						Label:       "OnFailure",
						Description: "The destination configuration for failed invocations.",
						Attributes: map[string]*provider.ResourceDefinitionsSchema{
							"destination": {
								Type:        provider.ResourceDefinitionsSchemaTypeString,
								Description: "The Amazon Resource Name (ARN) of the destination resource.",
								MinLength:   1,
								MaxLength:   350,
							},
						},
					},
					"onSuccess": {
						Type:        provider.ResourceDefinitionsSchemaTypeObject,
						Label:       "OnSuccess",
						Description: "The destination configuration for successful invocations.",
						Attributes: map[string]*provider.ResourceDefinitionsSchema{
							"destination": {
								Type:        provider.ResourceDefinitionsSchemaTypeString,
								Description: "The Amazon Resource Name (ARN) of the destination resource.",
								MinLength:   1,
								MaxLength:   350,
							},
						},
					},
				},
			},

			// Computed fields returned by AWS
			"functionArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the function.",
				Computed:    true,
			},
			"lastModified": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The date and time that the configuration was last updated.",
				Computed:    true,
			},
		},
	}
}
