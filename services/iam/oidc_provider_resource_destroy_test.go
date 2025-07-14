package iam

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMOidcProviderResourceDestroySuite struct {
	suite.Suite
}

func (s *IAMOidcProviderResourceDestroySuite) Test_destroy_iam_oidc_provider() {
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		destroyOidcProviderTestCase(providerCtx, loader),
		destroyOidcProviderMissingArnTestCase(providerCtx, loader),
		destroyOidcProviderServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		OidcProviderResource,
		&s.Suite,
	)
}

func destroyOidcProviderTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteOpenIDConnectProviderOutput(&iam.DeleteOpenIDConnectProviderOutput{}),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM OIDC provider",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn": core.MappingNodeFromString(oidcProviderArn),
						"url": core.MappingNodeFromString("https://example.com"),
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
				},
			},
		},
		ExpectError: false,
	}
}

func destroyOidcProviderMissingArnTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock()

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM OIDC provider missing ARN",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"url": core.MappingNodeFromString("https://example.com"),
					},
				},
			},
		},
		ExpectedError: "OIDC provider ARN is required for destroy",
	}
}

func destroyOidcProviderServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteOpenIDConnectProviderError(fmt.Errorf("failed to delete OIDC provider")),
	)

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, iamservice.Service]{
		Name: "Destroy IAM OIDC provider service error",
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
		Input: &provider.ResourceDestroyInput{
			ProviderContext: providerCtx,
			ResourceState: &state.ResourceState{
				SpecData: &core.MappingNode{
					Fields: map[string]*core.MappingNode{
						"arn": core.MappingNodeFromString(oidcProviderArn),
						"url": core.MappingNodeFromString("https://example.com"),
					},
				},
			},
		},
		ExpectedError: "failed to delete OIDC provider",
	}
}

func TestIAMOidcProviderResourceDestroy(t *testing.T) {
	suite.Run(t, new(IAMOidcProviderResourceDestroySuite))
}