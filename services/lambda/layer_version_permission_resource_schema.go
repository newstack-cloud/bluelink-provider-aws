package lambda

import (
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func lambdaLayerVersionPermissionResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "LambdaLayerVersionPermissionDefinition",
		Description: "The definition of an AWS Lambda layer version permission.",
		Required:    []string{"layerVersionArn", "statementId", "action", "principal"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"layerVersionArn": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The name or Amazon Resource Name (ARN) of the layer.",
				Pattern:      "^(arn:[a-zA-Z0-9-]+:lambda:[a-zA-Z0-9-]+:\\d{12}:layer:[a-zA-Z0-9-_]+)|[a-zA-Z0-9-_]+$",
				MinLength:    1,
				MaxLength:    140,
				MustRecreate: true,
			},
			"statementId": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "An identifier that distinguishes the policy from others on the same layer version.",
				Pattern:      "^[a-zA-Z0-9-_]+$",
				MinLength:    1,
				MaxLength:    100,
				MustRecreate: true,
			},
			"action": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "The API action that grants access to the layer. For example, `lambda:GetLayerVersion`.",
				Pattern:      "^lambda:GetLayerVersion$",
				MaxLength:    22,
				MustRecreate: true,
			},
			"principal": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "An account ID, or `*` to grant layer usage permission to all accounts in an organization, or all AWS accounts (if `organizationId` is not specified). For the last case, make sure that you really do want all AWS accounts to have usage permission to this layer.",
				Pattern:      "^(\\d{12}|\\*|arn:(aws[a-zA-Z-]*):iam::\\d{12}:root)$",
				MustRecreate: true,
			},
			"organizationId": {
				Type:         provider.ResourceDefinitionsSchemaTypeString,
				Description:  "With the principal set to `*`, grant permission to all accounts in the specified organization.",
				Pattern:      "^o-[a-z0-9]{10,32}$",
				MaxLength:    34,
				MustRecreate: true,
			},
			"id": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The computed unique identifier combining layer version and statement ID.",
				FormattedDescription: "The computed unique identifier combining layer version and statement ID." +
					"An example of this would be `arn:aws:lambda:us-west-2:123456789012:layer:my-layer:1#engineering-org`",
				Computed: true,
			},
		},
	}
}
