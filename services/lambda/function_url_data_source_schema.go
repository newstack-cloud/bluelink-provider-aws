package lambda

import "github.com/newstack-cloud/bluelink/libs/blueprint/provider"

func lambdaFunctionUrlDataSourceSchema() map[string]*provider.DataSourceSpecSchema {
	return map[string]*provider.DataSourceSpecSchema{
		"functionUrl": {
			Label:       "Function URL",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The HTTP URL endpoint for the Lambda function.",
			Nullable:    false,
		},
		"functionArn": {
			Label:       "Function ARN",
			Type:        provider.DataSourceSpecTypeString,
			Description: "The Amazon Resource Name (ARN) of the Lambda function.",
			Nullable:    false,
		},
		"authType": {
			Label:                "Authentication Type",
			Type:                 provider.DataSourceSpecTypeString,
			Description:          "The type of authentication that the function URL uses (AWS_IAM or NONE).",
			FormattedDescription: "The type of authentication that the function URL uses (`AWS_IAM` or `NONE`).",
			Nullable:             false,
		},
		"creationTime": {
			Label:       "Creation Time",
			Type:        provider.DataSourceSpecTypeString,
			Description: "When the function URL was created, in ISO-8601 format.",
			Nullable:    false,
		},
		"lastModifiedTime": {
			Label:       "Last Modified Time",
			Type:        provider.DataSourceSpecTypeString,
			Description: "When the function URL configuration was last updated, in ISO-8601 format.",
			Nullable:    false,
		},
		"invokeMode": {
			Label: "Invoke Mode",
			Type:  provider.DataSourceSpecTypeString,
			Description: "The invocation mode for the function URL. " +
				"BUFFERED (default) for synchronous invocation, RESPONSE_STREAM for response streaming.",
			FormattedDescription: "The invocation mode for the function URL. " +
				"`BUFFERED` (default) for synchronous invocation, `RESPONSE_STREAM` for response streaming.",
			Nullable: true,
		},
		"cors.allowCredentials": {
			Label:       "CORS Allow Credentials",
			Type:        provider.DataSourceSpecTypeBoolean,
			Description: "Whether to allow cookies or other credentials in requests to the function URL.",
			Nullable:    true,
		},
		"cors.allowHeaders": {
			Label:       "CORS Allow Headers",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The HTTP headers that origins can include in requests to the function URL.",
			Nullable:    true,
		},
		"cors.allowMethods": {
			Label:       "CORS Allow Methods",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The HTTP methods that are allowed when calling the function URL.",
			Nullable:    true,
		},
		"cors.allowOrigins": {
			Label:       "CORS Allow Origins",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The origins that can access the function URL.",
			Nullable:    true,
		},
		"cors.exposeHeaders": {
			Label:       "CORS Expose Headers",
			Type:        provider.DataSourceSpecTypeArray,
			Description: "The HTTP headers in the function response that you want to expose to origins that call the function URL.",
			Nullable:    true,
		},
		"cors.maxAge": {
			Label:       "CORS Max Age",
			Type:        provider.DataSourceSpecTypeInteger,
			Description: "The maximum amount of time, in seconds, that web browsers can cache results of a preflight request.",
			Nullable:    true,
		},
	}
}
