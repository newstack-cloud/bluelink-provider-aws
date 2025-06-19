package lambda

import (
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func lambdaLayerVersionResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaLayerVersionDefinition",
		Description: "The definition of an AWS Lambda layer version.",
		Required:    []string{"layerName", "content"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"layerName": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name or Amazon Resource Name (ARN) of the layer. " +
					"The length constraint applies only to the full ARN. If you specify only the layer name, " +
					"it is limited to 64 characters in length.",
				Pattern:      "^(arn:(aws[a-zA-Z-]*)?:lambda:)?([a-z]{2}((-gov)|(-iso([a-z]?)))?-[a-z]+-\\d{1}:)?(\\d{12}:)?(layer:)?([a-zA-Z0-9-_]+)$",
				MinLength:    1,
				MaxLength:    140,
				MustRecreate: true,
			},
			"content": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "LayerVersionContentInput",
				Description: "The function layer archive.",
				Required:    []string{},
				// Layer content cannot be updated - must recreate
				MustRecreate: true,
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"s3Bucket": {
						Type:         provider.ResourceDefinitionsSchemaTypeString,
						Description:  "An Amazon S3 bucket in the same AWS Region as your function. The bucket can be in a different AWS account.",
						MinLength:    3,
						MaxLength:    63,
						MustRecreate: true,
					},
					"s3Key": {
						Type:         provider.ResourceDefinitionsSchemaTypeString,
						Description:  "The Amazon S3 key of the deployment package.",
						MinLength:    1,
						MaxLength:    1024,
						MustRecreate: true,
					},
					"s3ObjectVersion": {
						Type:         provider.ResourceDefinitionsSchemaTypeString,
						Description:  "For versioned objects, the version of the deployment package object to use.",
						MinLength:    1,
						MaxLength:    1024,
						MustRecreate: true,
					},
				},
			},
			"compatibleArchitectures": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of compatible instruction set architectures.",
				Items: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
				},
			},
			"compatibleRuntimes": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of compatible function runtimes. Used for filtering with ListLayers and ListLayerVersions.",
				Items: &provider.ResourceDefinitionsSchema{
					Type: provider.ResourceDefinitionsSchemaTypeString,
				},
			},
			"description": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The description of the version.",
				MinLength:    0,
				MaxLength:    256,
				MustRecreate: true,
			},
			"licenseInfo": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The layer's software license. It can be any of the following:\n\n" +
					"- An [SPDX license identifier](https://spdx.org/licenses/). For example, `MIT`.\n" +
					"- The URL of a license hosted on the internet. For example, `https://opensource.org/licenses/MIT`.\n" +
					"- The full text of the license.",
				MaxLength:    512,
				MustRecreate: true,
			},

			// Computed fields
			"layerArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ARN of the layer.",
				Computed:    true,
			},
			"layerVersionArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The ARN of the layer version.",
				Computed:    true,
			},
			"version": {
				Type:        provider.ResourceDefinitionsSchemaTypeInteger,
				Description: "The version number.",
				Computed:    true,
			},
			"createdDate": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The date that the layer version was created, in ISO-8601 format (YYYY-MM-DDThh:mm:ss.sTZD).",
				Computed:    true,
			},
		},
	}
}
