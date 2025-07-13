package iam

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func iamAccessKeyResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "IAMAccessKeyDefinition",
		Description: "The definition of an AWS IAM access key.",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"userName": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the IAM user that the new access key will belong to.",
				FormattedDescription: "The name of the IAM user that the new access key will belong to. " +
					"This field is required and must reference an existing IAM user.",
				Pattern:   `[\w+=,.@-]+`,
				MinLength: 1,
				MaxLength: 128,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("john.doe"),
					core.MappingNodeFromString("service-account"),
					core.MappingNodeFromString("developer-user"),
				},
			},
			"status": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The status of the access key. Valid values are 'Active' and 'Inactive'.",
				FormattedDescription: "The status of the access key. Valid values are 'Active' and 'Inactive'. " +
					"Access keys are created in 'Active' status by default.",
				Default: core.MappingNodeFromString("Active"),
				AllowedValues: []*core.MappingNode{
					core.MappingNodeFromString("Active"),
					core.MappingNodeFromString("Inactive"),
				},
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("Active"),
					core.MappingNodeFromString("Inactive"),
				},
				Nullable: true,
			},

			// Computed fields
			"id": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The access key ID.",
				FormattedDescription: "The access key ID. " +
					"This is a computed field that is automatically set after the access key is created.",
				Computed: true,
			},
			"secretAccessKey": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The secret access key.",
				FormattedDescription: "The secret access key. " +
					"This is a computed field that is only available during initial creation.",
				Computed: true,
			},
		},
	}
}
