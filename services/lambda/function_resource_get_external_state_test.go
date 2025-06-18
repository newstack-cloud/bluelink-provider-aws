package lambda

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type LambdaFunctionResourceGetExternalStateSuite struct {
	suite.Suite
}

func (s *LambdaFunctionResourceGetExternalStateSuite) Test_get_external_state() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			pluginutils.SessionIDKey: core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service]{
		createBasicFunctionStateTestCase(providerCtx, loader),
		createAllOptionalConfigsTestCase(providerCtx, loader),
		createGetFunctionErrorTestCase(providerCtx, loader),
		createGetFunctionCodeSigningErrorTestCase(providerCtx, loader),
		createEphemeralStorageTestCase(providerCtx, loader),
		createImageConfigTestCase(providerCtx, loader),
		createTracingAndRuntimeVersionTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceGetExternalStateTestCases(
		testCases,
		FunctionResource,
		&s.Suite,
	)
}

func TestLambdaFunctionResourceGetExternalStateSuite(t *testing.T) {
	suite.Run(t, new(LambdaFunctionResourceGetExternalStateSuite))
}

// Test case generator functions below.

func createBasicFunctionStateTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service]{
		Name: "successfully gets basic function state",
		ServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"architecture": core.MappingNodeFromString("x86_64"),
					"functionName": core.MappingNodeFromString("test-function"),
					"runtime":      core.MappingNodeFromString("nodejs18.x"),
					"handler":      core.MappingNodeFromString("index.handler"),
					"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				},
			},
		},
		ExpectError: false,
	}
}

func createAllOptionalConfigsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service] {
	tags := map[string]string{
		"Environment": "test",
		"Project":     "celerity",
		"Service":     "lambda",
	}

	// Create sorted tag items for expected output
	tagItems := []*core.MappingNode{
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Environment"),
				"value": core.MappingNodeFromString("test"),
			},
		},
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Project"),
				"value": core.MappingNodeFromString("celerity"),
			},
		},
		{
			Fields: map[string]*core.MappingNode{
				"key":   core.MappingNodeFromString("Service"),
				"value": core.MappingNodeFromString("lambda"),
			},
		},
	}

	expectedOutput := &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"arn":          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				"architecture": core.MappingNodeFromString("x86_64"),
				"functionName": core.MappingNodeFromString("test-function"),
				"runtime":      core.MappingNodeFromString("nodejs18.x"),
				"handler":      core.MappingNodeFromString("index.handler"),
				"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"description":  core.MappingNodeFromString("Test function"),
				"memorySize":   core.MappingNodeFromInt(256),
				"timeout":      core.MappingNodeFromInt(30),
				"environment": {
					Fields: map[string]*core.MappingNode{
						"TEST_VAR": core.MappingNodeFromString("test-value"),
					},
				},
				"deadLetterConfig": {
					Fields: map[string]*core.MappingNode{
						"targetArn": core.MappingNodeFromString("arn:aws:sqs:us-east-1:123456789012:test-queue"),
					},
				},
				"vpcConfig": {
					Fields: map[string]*core.MappingNode{
						"securityGroupIds": {
							Items: []*core.MappingNode{
								core.MappingNodeFromString("sg-12345678"),
							},
						},
						"subnetIds": {
							Items: []*core.MappingNode{
								core.MappingNodeFromString("subnet-12345678"),
							},
						},
					},
				},
				"fileSystemConfig": {
					Fields: map[string]*core.MappingNode{
						"arn":            core.MappingNodeFromString("arn:aws:elasticfilesystem:us-east-1:123456789012:access-point/fsap-1234567890abcdef0"),
						"localMountPath": core.MappingNodeFromString("/mnt/efs"),
					},
				},
				"loggingConfig": {
					Fields: map[string]*core.MappingNode{
						"applicationLogLevel": core.MappingNodeFromString("DEBUG"),
						"logFormat":           core.MappingNodeFromString("JSON"),
						"logGroup":            core.MappingNodeFromString("/aws/lambda/test-function"),
						"systemLogLevel":      core.MappingNodeFromString("INFO"),
					},
				},
				"snapStart": {
					Fields: map[string]*core.MappingNode{
						"applyOn": core.MappingNodeFromString("PublishedVersions"),
					},
				},
				"layers": {
					Items: []*core.MappingNode{
						core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:layer:test-layer-1:1"),
						core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:layer:test-layer-2:2"),
					},
				},
				"tags": {
					Items: tagItems,
				},
				"code": {
					Fields: map[string]*core.MappingNode{
						"s3Bucket": core.MappingNodeFromString("test-bucket"),
						"s3Key":    core.MappingNodeFromString("test-key"),
					},
				},
				"snapStartResponseApplyOn":            core.MappingNodeFromString("PublishedVersions"),
				"snapStartResponseOptimizationStatus": core.MappingNodeFromString("On"),
			},
		},
	}

	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service]{
		Name: "successfully gets function state with all optional configurations",
		ServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
					Architectures: []types.Architecture{
						types.ArchitectureX8664,
					},
					Description: aws.String("Test function"),
					MemorySize:  aws.Int32(256),
					Timeout:     aws.Int32(30),
					Environment: &types.EnvironmentResponse{
						Variables: map[string]string{
							"TEST_VAR": "test-value",
						},
					},
					DeadLetterConfig: &types.DeadLetterConfig{
						TargetArn: aws.String("arn:aws:sqs:us-east-1:123456789012:test-queue"),
					},
					VpcConfig: &types.VpcConfigResponse{
						SecurityGroupIds: []string{"sg-12345678"},
						SubnetIds:        []string{"subnet-12345678"},
					},
					FileSystemConfigs: []types.FileSystemConfig{
						{
							Arn:            aws.String("arn:aws:elasticfilesystem:us-east-1:123456789012:access-point/fsap-1234567890abcdef0"),
							LocalMountPath: aws.String("/mnt/efs"),
						},
					},
					LoggingConfig: &types.LoggingConfig{
						ApplicationLogLevel: types.ApplicationLogLevelDebug,
						LogFormat:           types.LogFormatJson,
						LogGroup:            aws.String("/aws/lambda/test-function"),
						SystemLogLevel:      types.SystemLogLevelInfo,
					},
					SnapStart: &types.SnapStartResponse{
						ApplyOn:            types.SnapStartApplyOnPublishedVersions,
						OptimizationStatus: types.SnapStartOptimizationStatusOn,
					},
					Layers: []types.Layer{
						{
							Arn:      aws.String("arn:aws:lambda:us-east-1:123456789012:layer:test-layer-1:1"),
							CodeSize: 1024,
						},
						{
							Arn:      aws.String("arn:aws:lambda:us-east-1:123456789012:layer:test-layer-2:2"),
							CodeSize: 2048,
						},
					},
				},
				Code: &types.FunctionCodeLocation{
					Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
				},
				Tags: tags,
			}),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		CheckTags:      true,
		ExpectedOutput: expectedOutput,
		ExpectError:    false,
	}
}

func createGetFunctionErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service]{
		Name: "handles get function error",
		ServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionError(errors.New("failed to get function")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				},
			},
		},
		ExpectError: true,
	}
}

func createGetFunctionCodeSigningErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service]{
		Name: "handles get function code signing config error",
		ServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(createBaseTestFunctionConfig(
				"test-function",
				types.RuntimeNodejs18x,
				"index.handler",
				"arn:aws:iam::123456789012:role/test-role",
			)),
			WithGetFunctionCodeSigningError(errors.New("failed to get code signing config")),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
				},
			},
		},
		ExpectError: true,
	}
}

func createEphemeralStorageTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service]{
		Name: "successfully gets function state with ephemeral storage",
		ServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
					Architectures: []types.Architecture{
						types.ArchitectureX8664,
					},
					EphemeralStorage: &types.EphemeralStorage{
						Size: aws.Int32(1024), // 1024 MB = 1 GB
					},
				},
				Code: &types.FunctionCodeLocation{
					Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
				},
			}),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"architecture": core.MappingNodeFromString("x86_64"),
					"functionName": core.MappingNodeFromString("test-function"),
					"runtime":      core.MappingNodeFromString("nodejs18.x"),
					"handler":      core.MappingNodeFromString("index.handler"),
					"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
					"ephemeralStorage": {
						Fields: map[string]*core.MappingNode{
							"size": core.MappingNodeFromInt(1024),
						},
					},
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createImageConfigTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service]{
		Name: "successfully gets function state with image configuration",
		ServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
					Architectures: []types.Architecture{
						types.ArchitectureX8664,
					},
					ImageConfigResponse: &types.ImageConfigResponse{
						ImageConfig: &types.ImageConfig{
							Command: []string{
								"app.lambda_handler",
								"--config",
								"config.json",
							},
							EntryPoint: []string{
								"/var/runtime/bootstrap",
							},
							WorkingDirectory: aws.String("/var/task"),
						},
					},
				},
				Code: &types.FunctionCodeLocation{
					Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
				},
			}),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"architecture": core.MappingNodeFromString("x86_64"),
					"functionName": core.MappingNodeFromString("test-function"),
					"runtime":      core.MappingNodeFromString("nodejs18.x"),
					"handler":      core.MappingNodeFromString("index.handler"),
					"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
					"imageConfig": {
						Fields: map[string]*core.MappingNode{
							"command": {
								Items: []*core.MappingNode{
									core.MappingNodeFromString("app.lambda_handler"),
									core.MappingNodeFromString("--config"),
									core.MappingNodeFromString("config.json"),
								},
							},
							"entryPoint": {
								Items: []*core.MappingNode{
									core.MappingNodeFromString("/var/runtime/bootstrap"),
								},
							},
							"workingDirectory": core.MappingNodeFromString("/var/task"),
						},
					},
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}

func createTracingAndRuntimeVersionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service] {
	return plugintestutils.ResourceGetExternalStateTestCase[*aws.Config, Service]{
		Name: "successfully gets function state with tracing and runtime version config",
		ServiceFactory: createLambdaServiceMockFactory(
			WithGetFunctionOutput(&lambda.GetFunctionOutput{
				Configuration: &types.FunctionConfiguration{
					FunctionName: aws.String("test-function"),
					FunctionArn:  aws.String("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					Runtime:      types.RuntimeNodejs18x,
					Handler:      aws.String("index.handler"),
					Role:         aws.String("arn:aws:iam::123456789012:role/test-role"),
					Architectures: []types.Architecture{
						types.ArchitectureX8664,
					},
					TracingConfig: &types.TracingConfigResponse{
						Mode: types.TracingModeActive,
					},
					RuntimeVersionConfig: &types.RuntimeVersionConfig{
						RuntimeVersionArn: aws.String("arn:aws:lambda:us-east-1::runtime-version/test"),
					},
				},
				Code: &types.FunctionCodeLocation{
					Location: aws.String("https://test-bucket.s3.amazonaws.com/test-key"),
				},
			}),
			WithGetFunctionCodeSigningOutput(&lambda.GetFunctionCodeSigningConfigOutput{}),
			WithGetFunctionRecursionOutput(&lambda.GetFunctionRecursionConfigOutput{}),
			WithGetFunctionConcurrencyOutput(&lambda.GetFunctionConcurrencyOutput{}),
		),
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceGetExternalStateInput{
			ProviderContext: providerCtx,
			CurrentResourceSpec: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn": core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		ExpectedOutput: &provider.ResourceGetExternalStateOutput{
			ResourceSpecState: &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"arn":          core.MappingNodeFromString("arn:aws:lambda:us-east-1:123456789012:function:test-function"),
					"architecture": core.MappingNodeFromString("x86_64"),
					"functionName": core.MappingNodeFromString("test-function"),
					"runtime":      core.MappingNodeFromString("nodejs18.x"),
					"handler":      core.MappingNodeFromString("index.handler"),
					"role":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
					"tracingConfig": {
						Fields: map[string]*core.MappingNode{
							"mode": core.MappingNodeFromString("Active"),
						},
					},
					"runtimeManagementConfig": {
						Fields: map[string]*core.MappingNode{
							"runtimeVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-east-1::runtime-version/test"),
						},
					},
					"code": {
						Fields: map[string]*core.MappingNode{
							"s3Bucket": core.MappingNodeFromString("test-bucket"),
							"s3Key":    core.MappingNodeFromString("test-key"),
						},
					},
				},
			},
		},
		ExpectError: false,
	}
}
