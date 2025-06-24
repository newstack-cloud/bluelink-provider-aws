package lambda

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

func lambdaFunctionDataSourceSchema() map[string]*provider.DataSourceSpecSchema {
	return map[string]*provider.DataSourceSpecSchema{
		"architecture": {
			Label:       "Architecture",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The architecture of the Lambda function to retrieve.",
			Nullable:    false,
		},
		"arn": {
			Label: "Function ARN",
			Type:  provider.DataSourceSpecTypeString,
			FormattedDescription: "The ARN of the Lambda function to retrieve without a qualifier (without `:QUALIFIER` or `:VERSION` suffix)." +
				" See `qualifiedArn` for the ARN with a qualifier.",
			Nullable: false,
		},
		"codeSHA256": {
			Label:       "Code SHA256",
			Type:        provider.DataSourceSpecTypeString,
			Description: "A base64-encoded representation of the SHA-256 sum of the lambda zip archive.",
			Nullable:    true,
		},
		"codeSigningConfigArn": {
			Label:       "Code Signing Config ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the code signing configuration associated with the function.",
			Nullable:    true,
		},
		"deadLetterConfig.targetArn": {
			Label:       "Dead Letter Config",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the SQS queue or SNS topic to send dead letter events to.",
			Nullable:    true,
		},
		"environment.variables": {
			Label:       "Environment Variables",
			Type:        provider.DataSourceSpecTypeString,
			Description: "A json-encoded string of the environment variables of the Lambda function.",
			Nullable:    true,
		},
		"ephemeralStorage.size": {
			Label:       "Ephemeral Storage Size",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The size of the ephemeral storage used for the Lambda function.",
			Nullable:    true,
		},
		"fileSystemConfig.arn": {
			Label:       "FileSystem Config ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the EFS file system mounted for the function.",
			Nullable:    true,
		},
		"fileSystemConfig.localMountPath": {
			Label:       "Local Mount Path",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The path to the mount point for the file system.",
			Nullable:    true,
		},
		"handler": {
			Label:       "Handler",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The handler (code entry point) of the Lambda function.",
			Nullable:    true,
		},
		"imageUri": {
			Label:       "Image URI",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The URI of the container image used for the Lambda function.",
			Nullable:    true,
		},
		"kmsKeyArn": {
			Label: "KMS Key ARN",
			Type:  provider.DataSourceSpecTypeString,
			Description: "The ARN of the KMS key used to encrypt the" +
				" function's environment variables and SnapStart snapshots.",
			Nullable: true,
		},
		"layers": {
			Label: "Layers",
			Type:  provider.DataSourceSpecTypeArray,
			Description: "A list of layer ARNs " +
				"that are included in the function's execution environment.",
			Nullable: true,
		},
		"loggingConfig.applicationLogLevel": {
			Label:       "Application Log Level",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The log level for the application logs.",
			Nullable:    true,
		},
		"loggingConfig.logFormat": {
			Label:       "Log Format",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The format of the log output.",
			Nullable:    true,
		},
		"loggingConfig.logGroup": {
			Label:       "Log Group",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The name of the CloudWatch Logs log group for the function.",
			Nullable:    true,
		},
		"loggingConfig.systemLogLevel": {
			Label:       "System Log Level",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The log level for the system logs.",
			Nullable:    true,
		},
		"memorySize": {
			Label:       "Memory Size",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The amount of memory in MB available to the function at runtime.",
			Nullable:    true,
		},
		"name": {
			Label:       "Name",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The name of the Lambda function.",
			Nullable:    true,
		},
		"qualifiedArn": {
			Label:                "Qualified ARN",
			Type:                 provider.DataSourceSpecTypeString,
			FormattedDescription: "The ARN of the Lambda function with a qualifier (`:QUALIFIER` or `:VERSION` suffix).",
			Nullable:             true,
		},
		"reservedConcurrentExecutions": {
			Label:       "Reserved Concurrent Executions",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The number of concurrent executions reserved for the function.",
			Nullable:    true,
		},
		"role": {
			Label:       "Role",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the IAM role associated with the function.",
			Nullable:    true,
		},
		"runtime": {
			Label:       "Runtime",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The runtime environment for the Lambda function.",
			Nullable:    true,
		},
		"signingJobArn": {
			Label:       "Signing Job ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The ARN of the signing job for the Lambda function.",
			Nullable:    true,
		},
		"sourceCodeSize": {
			Label:       "Source Code Size",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The size of the source code of the Lambda function.",
			Nullable:    true,
		},
		"timeout": {
			Label:       "Timeout",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The timeout for the Lambda function.",
			Nullable:    true,
		},
		"tracingConfig.mode": {
			Label:       "Tracing Config Mode",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The mode of tracing for the Lambda function.",
			Nullable:    true,
		},
		"version": {
			Label: "Version",
			Type:  provider.DataSourceSpecTypeString,
			FormattedDescription: "The version of the Lambda function. " +
				"If `qualifier` is not provided in the filter fields, the most recent published version will be used. " +
				"If there isn't a published version, `$LATEST` will be used for this field.",
			Nullable: true,
		},
		"vpcConfig.ipv6AllowedForDualStack": {
			Label:       "IPv6 Allowed for Dual Stack",
			Type:        provider.DataSourceSpecTypeBoolean,
			Description: "Whether outbound IPv6 traffic is allowed for dual-stack subnets.",
			Nullable:    true,
		},
		"vpcConfig.securityGroupIds": {
			Label:       "Security Group IDs",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The IDs of the security groups for the Lambda function.",
			Nullable:    true,
		},
		"vpcConfig.subnetIds": {
			Label:       "Subnet IDs",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The IDs of the subnets for the Lambda function.",
			Nullable:    true,
		},
	}
}
