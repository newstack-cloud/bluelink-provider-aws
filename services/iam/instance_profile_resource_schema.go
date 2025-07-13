package iam

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func iamInstanceProfileResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "IAMInstanceProfileDefinition",
		Description: "The definition of an AWS IAM instance profile.",
		Required:    []string{"role"},
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"instanceProfileName": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the instance profile to create.",
				FormattedDescription: "The name of the instance profile to create. " +
					"If you do not specify a name, a unique name will be generated automatically.",
				Pattern:      `^[\w+=,.@-]+$`,
				MinLength:    1,
				MaxLength:    128,
				MustRecreate: true,
				Nullable:     true,
			},
			"path": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The path to the instance profile. For more information about paths, see IAM Identifiers in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				FormattedDescription: "The path to the instance profile. For more information about paths, see " +
					"[IAM Identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html) in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				Pattern:      `(\u002F)|(\u002F[\u0021-\u007E]+\u002F)`,
				MinLength:    1,
				MaxLength:    512,
				Default:      core.MappingNodeFromString("/"),
				MustRecreate: true,
				Nullable:     true,
			},
			"role": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name or ARN of the role to associate with the instance profile.",
				FormattedDescription: "The name or ARN of the role to associate with the instance profile. " +
					"This is the role that will be assumed by EC2 instances when they launch.",
				Pattern: `^(arn:aws:iam::\d{12}:role/[\w+=,.@-]+|[\w+=,.@-]+)$`,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("MyRole"),
					core.MappingNodeFromString("arn:aws:iam::123456789012:role/MyRole"),
				},
			},

			// Computed fields
			"arn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the IAM instance profile.",
				FormattedDescription: "The Amazon Resource Name (ARN) of the IAM instance profile. " +
					"This is a computed field that is automatically set after the instance profile is created.",
				Computed: true,
			},
		},
		Examples: []*core.MappingNode{
			{
				Fields: map[string]*core.MappingNode{
					"instanceProfileName": core.MappingNodeFromString("MyInstanceProfile"),
					"path":                core.MappingNodeFromString("/"),
					"role":                core.MappingNodeFromString("MyRole"),
				},
			},
		},
	}
}
