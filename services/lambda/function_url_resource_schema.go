package lambda

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func lambdaFunctionUrlResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaFunctionUrlDefinition",
		Description: "The definition of an AWS Lambda function URL.",
		Required:    []string{"targetFunctionArn", "authType"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"targetFunctionArn": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The name or ARN of the Lambda function.",
				Pattern:      "^(arn:(aws[a-zA-Z-]*)?:lambda:)?([a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\\d{1}:)?(\\d{12}:)?(function:)?([a-zA-Z0-9-_]+)(:((?!\\d+)[0-9a-zA-Z-_]+))?$",
				MinLength:    1,
				MaxLength:    140,
				MustRecreate: true,
			},
			"authType": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The type of authentication that your function URL uses.",
				Pattern:      "^(AWS_IAM|NONE)$",
				MinLength:    1,
				MaxLength:    10,
				MustRecreate: true,
			},
			"qualifier": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The alias name.",
				Pattern:      "((?!^[0-9]+$)([a-zA-Z0-9-_]+))",
				MinLength:    1,
				MaxLength:    128,
				MustRecreate: true,
			},
			"invokeMode": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The invocation mode for the function URL.",
				Pattern:     "^(BUFFERED|RESPONSE_STREAM)$",
				MinLength:   1,
				MaxLength:   20,
			},
			"cors": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "FunctionUrlCors",
				Description: "The Cross-Origin Resource Sharing (CORS) settings for your function URL.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"allowCredentials": {
						Type:        provider.ResourceDefinitionsSchemaTypeBoolean,
						Description: "Whether to allow cookies to be included in CORS requests.",
					},
					"allowHeaders": {
						Type: provider.ResourceDefinitionsSchemaTypeArray,
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						Description: "The allowed headers for CORS requests.",
					},
					"allowMethods": {
						Type: provider.ResourceDefinitionsSchemaTypeArray,
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						Description: "The allowed HTTP methods for CORS requests.",
					},
					"allowOrigins": {
						Type: provider.ResourceDefinitionsSchemaTypeArray,
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						Description: "The allowed origins for CORS requests.",
					},
					"exposeHeaders": {
						Type: provider.ResourceDefinitionsSchemaTypeArray,
						Items: &provider.ResourceDefinitionsSchema{
							Type: provider.ResourceDefinitionsSchemaTypeString,
						},
						Description: "The headers to expose in CORS responses.",
					},
					"maxAge": {
						Type:        provider.ResourceDefinitionsSchemaTypeInteger,
						Description: "The maximum age (in seconds) for CORS preflight requests.",
						Minimum:     core.ScalarFromInt(0),
					},
				},
			},

			// Computed fields returned by AWS
			"functionArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the function.",
				Computed:    true,
			},
			"functionUrl": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The HTTP URL endpoint for your function.",
				Computed:    true,
			},
		},
	}
}
