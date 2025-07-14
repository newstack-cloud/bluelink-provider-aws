package iam

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMOidcProviderResourceCreateSuite struct {
	suite.Suite
}

func (s *IAMOidcProviderResourceCreateSuite) Test_create_iam_oidc_provider() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		createBasicOidcProviderTestCase(providerCtx, loader),
		createOidcProviderWithTagsTestCase(providerCtx, loader),
		createOidcProviderWithGeneratedUrlTestCase(providerCtx, loader),
		createOidcProviderServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		OidcProviderResource,
		&s.Suite,
	)
}

func createBasicOidcProviderTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/token.actions.githubusercontent.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateOpenIDConnectProviderOutput(&iam.CreateOpenIDConnectProviderOutput{
			OpenIDConnectProviderArn: aws.String(oidcProviderArn),
		}),
	)

	// Create test data for OIDC provider creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://token.actions.githubusercontent.com"),
			"clientIdList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("sts.amazonaws.com"),
				},
			},
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
					core.MappingNodeFromString("9e99a48a9960b14926bb7f3b02e22da2b0ab7280"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create basic IAM OIDC provider",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.url",
					},
					{
						FieldPath: "spec.clientIdList",
					},
					{
						FieldPath: "spec.thumbprintList",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(oidcProviderArn),
			},
		},
		SaveActionsCalled: []plugintestutils.SaveActionCalled{
			{
				SaveAction: "create OIDC provider",
			},
		},
	}
}

func createOidcProviderWithTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/accounts.google.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateOpenIDConnectProviderOutput(&iam.CreateOpenIDConnectProviderOutput{
			OpenIDConnectProviderArn: aws.String(oidcProviderArn),
		}),
		iammock.WithTagOpenIDConnectProviderOutput(&iam.TagOpenIDConnectProviderOutput{}),
	)

	// Create test data for OIDC provider creation with tags
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://accounts.google.com"),
			"clientIdList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("123456789012-abcdef.apps.googleusercontent.com"),
				},
			},
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
			"tags": &core.MappingNode{
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Production"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Service"),
							"value": core.MappingNodeFromString("Authentication"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM OIDC provider with tags",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.url",
					},
					{
						FieldPath: "spec.clientIdList",
					},
					{
						FieldPath: "spec.thumbprintList",
					},
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(oidcProviderArn),
			},
		},
		SaveActionsCalled: []plugintestutils.SaveActionCalled{
			{
				SaveAction: "create OIDC provider",
			},
			{
				SaveAction: "save OIDC provider tags",
			},
		},
	}
}

func createOidcProviderWithGeneratedUrlTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	// Since the URL will be generated, we'll use a regex or just check that it's created
	generatedArn := "arn:aws:iam::123456789012:oidc-provider/oidc-provider-test-generated.example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithCreateOpenIDConnectProviderOutput(&iam.CreateOpenIDConnectProviderOutput{
			OpenIDConnectProviderArn: aws.String(generatedArn),
		}),
	)

	// Create test data without URL - it should be generated
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"clientIdList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("my-app-id"),
				},
			},
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM OIDC provider with generated URL",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.clientIdList",
					},
					{
						FieldPath: "spec.thumbprintList",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn": core.MappingNodeFromString(generatedArn),
			},
		},
		SaveActionsCalled: []plugintestutils.SaveActionCalled{
			{
				SaveAction: "create OIDC provider",
			},
		},
	}
}

func createOidcProviderServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithCreateOpenIDConnectProviderError(fmt.Errorf("failed to create OIDC provider")),
	)

	// Create test data for OIDC provider creation
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Create IAM OIDC provider service error",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: specData,
					},
				},
				NewFields: []provider.FieldChange{
					{
						FieldPath: "spec.url",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedError: "failed to create OIDC provider",
	}
}

func TestIAMOidcProviderResourceCreate(t *testing.T) {
	suite.Run(t, new(IAMOidcProviderResourceCreateSuite))
}