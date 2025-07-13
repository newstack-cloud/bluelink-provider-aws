package iam

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func iamGroupResourceSchema() *provider.ResourceDefinitionsSchema {
	return &provider.ResourceDefinitionsSchema{
		Type:        provider.ResourceDefinitionsSchemaTypeObject,
		Label:       "IAMGroupDefinition",
		Description: "The definition of an AWS IAM group.",
		Attributes: map[string]*provider.ResourceDefinitionsSchema{
			"managedPolicyArns": {
				Type:        provider.ResourceDefinitionsSchemaTypeArray,
				Description: "A list of Amazon Resource Names (ARNs) of the IAM managed policies that you want to attach to the group.",
				FormattedDescription: "A list of Amazon Resource Names (ARNs) of the IAM managed policies that you want to attach to the group. " +
					"For more information about ARNs, see [Amazon Resource Names (ARNs) and AWS Service Namespaces](https://docs.aws.amazon.com/general/latest/gr/aws-arns-and-namespaces.html) " +
					"in the AWS General Reference.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:        provider.ResourceDefinitionsSchemaTypeString,
					Description: "The ARN of an IAM managed policy.",
					Pattern:     `^arn:(aws[a-zA-Z-]*)?:iam::(aws|\d{12}):policy/[\w+=,.@-]+$`,
					Examples: []*core.MappingNode{
						core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
						core.MappingNodeFromString("arn:aws:iam::aws:policy/PowerUserAccess"),
						core.MappingNodeFromString("arn:aws:iam::123456789012:policy/MyCustomPolicy"),
					},
				},
				Nullable: true,
			},
			"path": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The path for the group name. For more information about paths, see IAM identifiers in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				FormattedDescription: "The path for the group name. For more information about paths, see " +
					"[IAM identifiers](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_identifiers.html) in the IAM User Guide. " +
					"This parameter is optional. If it is not included, it defaults to a slash (/).",
				Pattern:      `(\u002F)|(\u002F[\u0021-\u007E]+\u002F)`,
				MinLength:    1,
				MaxLength:    512,
				Default:      core.MappingNodeFromString("/"),
				MustRecreate: true,
				Nullable:     true,
			},
			"policies": {
				Type: provider.ResourceDefinitionsSchemaTypeArray,
				Description: "Adds or updates an inline policy document that is embedded in the specified IAM group. " +
					"When you embed an inline policy in a group, the inline policy is used as part of the group's access (permissions) policy.",
				FormattedDescription: "Adds or updates an inline policy document that is embedded in the specified IAM group. " +
					"When you embed an inline policy in a group, the inline policy is used as part of the group's access (permissions) policy. " +
					"For more information about policies, see [Managed Policies and Inline Policies](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_managed-vs-inline.html) " +
					"in the IAM User Guide.",
				Items: &provider.ResourceDefinitionsSchema{
					Type:  provider.ResourceDefinitionsSchemaTypeObject,
					Label: "Policy",
					Attributes: map[string]*provider.ResourceDefinitionsSchema{
						"policyName": {
							Type:        provider.ResourceDefinitionsSchemaTypeString,
							Description: "The name of the policy document.",
							Pattern:     `[\w+=,.@-]+`,
							MinLength:   1,
							MaxLength:   128,
						},
						"policyDocument": {
							Type:        provider.ResourceDefinitionsSchemaTypeObject,
							Description: "The policy document.",
							Label:       "PolicyDocument",
							Required:    []string{"Version", "Statement"},
							Attributes: map[string]*provider.ResourceDefinitionsSchema{
								"Version": {
									Type:        provider.ResourceDefinitionsSchemaTypeString,
									Description: "The policy language version. The only valid value is '2012-10-17'.",
									Pattern:     `^2012-10-17$`,
									Examples: []*core.MappingNode{
										core.MappingNodeFromString("2012-10-17"),
									},
								},
								"Statement": {
									Type:        provider.ResourceDefinitionsSchemaTypeArray,
									Description: "An array of individual statements in the policy document.",
									Items: &provider.ResourceDefinitionsSchema{
										Type:     provider.ResourceDefinitionsSchemaTypeObject,
										Label:    "Statement",
										Required: []string{"Effect", "Action"},
										Attributes: map[string]*provider.ResourceDefinitionsSchema{
											"Effect": {
												Type:        provider.ResourceDefinitionsSchemaTypeString,
												Description: "The effect of the statement. Valid values are 'Allow' and 'Deny'.",
												Pattern:     `^(Allow|Deny)$`,
												Examples: []*core.MappingNode{
													core.MappingNodeFromString("Allow"),
													core.MappingNodeFromString("Deny"),
												},
											},
											"Action": {
												Type:        provider.ResourceDefinitionsSchemaTypeArray,
												Description: "The action or actions that will be allowed or denied.",
												Items: &provider.ResourceDefinitionsSchema{
													Type: provider.ResourceDefinitionsSchemaTypeString,
													Examples: []*core.MappingNode{
														core.MappingNodeFromString("s3:GetObject"),
														core.MappingNodeFromString("ec2:DescribeInstances"),
														core.MappingNodeFromString("iam:ListUsers"),
													},
												},
											},
											"Resource": {
												Type:        provider.ResourceDefinitionsSchemaTypeArray,
												Description: "The resource or resources that the statement covers.",
												Items: &provider.ResourceDefinitionsSchema{
													Type: provider.ResourceDefinitionsSchemaTypeString,
													Examples: []*core.MappingNode{
														core.MappingNodeFromString("*"),
														core.MappingNodeFromString("arn:aws:s3:::my-bucket/*"),
														core.MappingNodeFromString("arn:aws:ec2:us-west-2:123456789012:instance/*"),
													},
												},
												Nullable: true,
											},
											"Condition": {
												Type:        provider.ResourceDefinitionsSchemaTypeObject,
												Description: "The conditions under which the statement is in effect.",
												Nullable:    true,
											},
										},
									},
								},
							},
						},
					},
				},
				Required: []string{"policyName", "policyDocument"},
				Nullable: true,
			},
			"groupName": {
				Type: provider.ResourceDefinitionsSchemaTypeString,
				Description: "The name of the group to create. Do not include the path in this value. " +
					"The group name must be unique within the account. Group names are not distinguished by case. " +
					"If not specified, a unique name will be generated.",
				FormattedDescription: "The name of the group to create. Do not include the path in this value. " +
					"The group name must be unique within the account. Group names are not distinguished by case. " +
					"If not specified, a unique name will be generated.",
				Pattern:      `[\w+=,.@-]+`,
				MinLength:    1,
				MaxLength:    128,
				MustRecreate: true,
				Nullable:     true,
				Examples: []*core.MappingNode{
					core.MappingNodeFromString("MyGroup"),
					core.MappingNodeFromString("developers-group"),
					core.MappingNodeFromString("service-account-group"),
				},
			},

			// Computed fields
			"arn": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The Amazon Resource Name (ARN) of the IAM group.",
				FormattedDescription: "The Amazon Resource Name (ARN) of the IAM group. " +
					"This is a computed field that is automatically set after the group is created.",
				Computed: true,
			},
			"groupId": {
				Type:        provider.ResourceDefinitionsSchemaTypeString,
				Description: "The stable and unique string identifying the group.",
				FormattedDescription: "The stable and unique string identifying the group. " +
					"This is a computed field that is automatically set after the group is created.",
				Computed: true,
			},
		},
	}
}
