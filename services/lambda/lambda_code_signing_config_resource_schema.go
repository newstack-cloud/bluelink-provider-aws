package lambda

import (
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func lambdaCodeSigningConfigResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaCodeSigningConfigDefinition",
		Description: "The definition of an AWS Lambda code signing configuration.",
		Required:    []string{"allowedPublishers"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"allowedPublishers": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "AllowedPublishers",
				Description: "Signing profiles for this code signing configuration.",
				Required:    []string{"signingProfileVersionArns"},
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"signingProfileVersionArns": {
						Type:        provider.ResourceDefinitionsSchemaTypeArray,
						Description: "The Amazon Resource Name (ARN) for each of the signing profiles. A signing profile defines a trusted user who can sign a code package.",
						Items: &provider.ResourceDefinitionsSchema{
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The ARN of a signing profile version.",
							Pattern:     "arn:(aws[a-zA-Z-]*)?:signer:[a-z]{2}((-gov)|(-iso(b?)))?-[a-z]+-\\d{1}:\\d{12}:/signing-profiles/[a-zA-Z0-9_]{2,64}/[a-zA-Z0-9]{10}",
						},
						MinLength: 1,
						MaxLength: 20,
					},
				},
			},
			"codeSigningPolicies": {
				Type:        provider.ResourceDefinitionsSchemaTypeObject,
				Label:       "CodeSigningPolicies",
				Description: "The code signing policies define the actions to take if the validation checks fail.",
				Attributes: map[string]*provider.ResourceDefinitionsSchema{
					"untrustedArtifactOnDeployment": {
						Type:        provider.ResourceDefinitionsSchemaTypeString,
						Description: "Code signing configuration policy for deployment validation failure. If you set the policy to Enforce, Lambda blocks the deployment request if code-signing validation checks fail. If you set the policy to Warn, Lambda allows the deployment and creates a CloudWatch log.",
						Default:     core.MappingNodeFromString("Warn"),
						AllowedValues: []*core.MappingNode{
							core.MappingNodeFromString("Warn"),
							core.MappingNodeFromString("Enforce"),
						},
					},
				},
			},
			"description": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "Descriptive name for this code signing configuration.",
				MinLength:   0,
				MaxLength:   256,
			},
			"tags": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of tags to apply to the code signing configuration.",
				FormattedDescription: "A list of [tags](https://docs.aws.amazon.com/lambda/latest/dg/tagging.html) " +
					"to apply to the code signing configuration.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:                 provider.ResourceDefinitionsSchemaTypeObject,
					Label:                "Tag",
					Description:          "A tag to apply to the code signing configuration.",
					FormattedDescription: "A [tag](https://docs.aws.amazon.com/lambda/latest/dg/tagging.html) to apply to the code signing configuration.",
					Required:             []string{"key", "value"},
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"key": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The key of the tag.",
							MinLength:   1,
							MaxLength:   128,
						},
						"value": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The value of the tag.",
							MinLength:   0,
							MaxLength:   256,
						},
					},
				},
			},
			// Computed fields returned by AWS
			"codeSigningConfigArn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the code signing configuration.",
				Computed:    true,
			},
			"codeSigningConfigId": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "Unique identifier for the code signing configuration.",
				Computed:    true,
			},
		},
	}
}
