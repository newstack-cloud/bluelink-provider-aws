package lambda

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

func lambdaCodeSigningConfigDataSourceSchema() map[string]*provider.DataSourceSpecSchema {
	return map[string]*provider.DataSourceSpecSchema{
		"arn": {
			Label:       "Code Signing Config ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the Lambda code signing configuration.",
			Nullable:    false,
		},
		"codeSigningConfigId": {
			Label:       "Code Signing Config ID",
			Type:        provider.DataSourceSpecTypeString,
			Description: "Unique identifier for the code signing configuration.",
			Nullable:    false,
		},
		"description": {
			Label:       "Description",
			Type:        provider.DataSourceSpecTypeString,
			Description: "Descriptive name for this code signing configuration.",
			Nullable:    true,
		},
		"allowedPublishers.signingProfileVersionArns": {
			Label:       "Signing Profile Version ARNs",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The Amazon Resource Name (ARN) for each of the signing profiles. A signing profile defines a trusted user who can sign a code package.",
			Nullable:    true,
		},
		"codeSigningPolicies.untrustedArtifactOnDeployment": {
			Label:       "Untrusted Artifact On Deployment",
			Type:        provider.DataSourceSpecTypeString,
			Description: "Code signing configuration policy for deployment validation failure. If you set the policy to Enforce, Lambda blocks the deployment request if code-signing validation checks fail. If you set the policy to Warn, Lambda allows the deployment and creates a CloudWatch log.",
			Nullable:    true,
		},
		"tags": {
			Label:       "Tags",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "A map of tags assigned to the code signing configuration.",
			Nullable:    true,
		},
	}
}
