package lambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func lambdaFunctionVersionResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaFunctionVersionDefinition",
		Description: "The definition of an AWS Lambda function version.",
		Required:    []string{"functionName"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"functionName": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name or ARN of the Lambda function. The length constraint applies only to the full ARN. " +
					"If you specify only the function name, it is limited to 64 characters in length.",
				Pattern:      "^(arn:(aws[a-zA-Z-]*)?:lambda:)?([a-z]{2}((-gov)|(-iso([a-z]?)))?-[a-z]+-\\d{1}:)?(\\d{12}:)?(function:)?([a-zA-Z0-9-_]+)(:(\\$LATEST|[a-zA-Z0-9-_]+))?$",
				MinLength:    1,
				MaxLength:    140,
				MustRecreate: true,
			},
			"description": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "A description for the version to override the description in the function configuration.",
				MinLength:    0,
				MaxLength:    256,
				MustRecreate: true,
			},
			"codeSha256": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "Only publish a version if the hash value matches the value that's specified. " +
					"Use this option to avoid publishing a version if the function code has changed since you last updated it.",
				MustRecreate: true,
			},
			"provisionedConcurrencyConfig": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "ProvisionedConcurrencyConfiguration",
				Description: "Specifies a provisioned concurrency configuration for a function's version.",
				Required:    []string{"provisionedConcurrentExecutions"},
				// Provisioned concurrency can be updated for a function version without
				// having to publish new versions.
				MustRecreate: false,
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"provisionedConcurrentExecutions": {
						Type:         provider.ResourceDefinitionsSchemaTypeInteger,
						Description:  "The amount of provisioned concurrency to allocate for the version.",
						MustRecreate: false,
					},
				},
			},
			"runtimePolicy": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "RuntimePolicy",
				Description: "The runtime management configuration for the version.",
				Required:    []string{"updateRuntimeOn"},
				// Runtime policy can be updated for a function version without
				// having to publish new versions.
				MustRecreate: false,
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"updateRuntimeOn": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "The runtime update mode to use.",
						FormattedDescription: "The runtime update mode to use.\n\n" +
							"- **Auto (default)** - Automatically update to the most recent and secure runtime " +
							"version using a [Two-phase runtime version rollout](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-update.html#runtime-management-two-phase)." +
							" This is the best choice for most cases as it ensures you will always benefit from runtime updates.\n" +
							"- **FunctionUpdate** - Lambda updates the runtime of your function to the most recent and secure runtime version " +
							"when you update your function. This approach synchronizes runtime updates with function deployments, " +
							"giving you control over when runtime updates are applied and allowing you to detect and mitigate rare runtime update incompatibilities early. " +
							"When using this setting, you need to regularly update your functions to keep their runtime up-to-date.\n" +
							"- **Manual** - You specify a runtime version in your function configuration. " +
							"The function will use this runtime version indefinitely. In the rare case where a runtime version is incompatible with an existing function, " +
							"this allows you to roll back your function to an earlier runtime version. " +
							"For more information, see [Roll back a runtime version](https://docs.aws.amazon.com/lambda/latest/dg/runtimes-update.html#runtime-management-rollback).",
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("Auto"),
							core.MappingNodeFromString("FunctionUpdate"),
							core.MappingNodeFromString("Manual"),
						},
						MustRecreate: false,
					},
					"runtimeVersionArn": {
						Type:         provider.ResourceDefinitionsSchemaTypeString,
						Description:  "The ARN of the runtime version you want the function to use.",
						Pattern:      "^arn:(aws[a-zA-Z-]*):lambda:[a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\\d{1}::runtime:.+$",
						MinLength:    26,
						MaxLength:    2048,
						MustRecreate: false,
					},
				},
			},

			// Computed fields
			"functionArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the Lambda function.",
				Computed:    true,
			},
			"version": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The version number of the Lambda function.",
				Computed:    true,
			},
			// Due to the blueprint framework's API for defining an external ID field
			// for a resource, we need to create a composite field that combines the function ARN
			// and the version number.
			"functionArnWithVersion": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) for the lambda function including the version number.",
				Computed:    true,
			},
		},
	}
}
