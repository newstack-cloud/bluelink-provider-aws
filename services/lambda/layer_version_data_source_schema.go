package lambda

import "github.com/newstack-cloud/celerity/libs/blueprint/provider"

func lambdaLayerVersionDataSourceSchema() map[string]*provider.DataSourceSpecSchema {
	return map[string]*provider.DataSourceSpecSchema{
		"arn": {
			Label:       "Layer ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the layer.",
			Nullable:    false,
		},
		"compatibleArchitectures": {
			Label:       "Compatible Architectures",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "A list of compatible instruction set architectures.",
			Nullable:    true,
		},
		"compatibleRuntimes": {
			Label:       "Compatible Runtimes",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The layer's compatible runtimes.",
			Nullable:    true,
		},
		"content.codeSha256": {
			Label:       "Code SHA256",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The SHA-256 hash of the layer archive.",
			Nullable:    true,
		},
		"content.codeSize": {
			Label:       "Code Size",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The size of the layer archive in bytes.",
			Nullable:    true,
		},
		"content.location": {
			Label:       "Download Location",
			Type:        provider.DataSourceSpecTypeString,
			Description: "A link to the layer archive in Amazon S3 that is valid for 10 minutes.",
			Nullable:    true,
		},
		"content.signingJobArn": {
			Label:       "Signing Job ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The Amazon Resource Name (ARN) of a signing job.",
			Nullable:    true,
		},
		"content.signingProfileVersionArn": {
			Label:       "Signing Profile Version ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The Amazon Resource Name (ARN) for a signing profile version.",
			Nullable:    true,
		},
		"createdDate": {
			Label:       "Created Date",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The date that the layer version was created, in ISO-8601 format.",
			Nullable:    true,
		},
		"description": {
			Label:       "Description",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The description of the version.",
			Nullable:    true,
		},
		"layerVersionArn": {
			Label:       "Layer Version ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the layer version.",
			Nullable:    true,
		},
		"licenseInfo": {
			Label:       "License Info",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The layer's software license.",
			Nullable:    true,
		},
		"version": {
			Label:       "Version",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The version number.",
			Nullable:    false,
		},
	}
}
