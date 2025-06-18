package lambda

import (
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/blueprint/state"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaEventSourceMappingResourceDestroySuite struct {
	suite.Suite
}

func (s *LambdaEventSourceMappingResourceDestroySuite) Test_destroy() {
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

	testCases := []plugintestutils.ResourceDestroyTestCase[*aws.Config, Service]{
		createSuccessfulEventSourceMappingDestroyTestCase(providerCtx, loader),
		createFailingEventSourceMappingDestroyTestCase(providerCtx, loader),
		createEventSourceMappingDestroyWithNoIDTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDestroyTestCases(
		testCases,
		EventSourceMappingResource,
		&s.Suite,
	)
}

func createSuccessfulEventSourceMappingDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock(
		WithDeleteEventSourceMappingOutput(&lambda.DeleteEventSourceMappingOutput{}),
	)

	expectedUUID := "12345678-1234-1234-1234-123456789012"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, Service]{
		Name: "successfully deletes event source mapping",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
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
						"id": core.MappingNodeFromString(expectedUUID),
					},
				},
			},
		},
		ExpectError: false,
		DestroyActionsCalled: map[string]any{
			"DeleteEventSourceMapping": &lambda.DeleteEventSourceMappingInput{
				UUID: aws.String(expectedUUID),
			},
		},
	}
}

func createFailingEventSourceMappingDestroyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock(
		WithDeleteEventSourceMappingError(errors.New("failed to delete event source mapping")),
	)

	expectedUUID := "12345678-1234-1234-1234-123456789012"

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, Service]{
		Name: "fails to delete event source mapping",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
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
						"id": core.MappingNodeFromString(expectedUUID),
					},
				},
			},
		},
		ExpectError: true,
		DestroyActionsCalled: map[string]any{
			"DeleteEventSourceMapping": &lambda.DeleteEventSourceMappingInput{
				UUID: aws.String(expectedUUID),
			},
		},
	}
}

func createEventSourceMappingDestroyWithNoIDTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDestroyTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock()

	return plugintestutils.ResourceDestroyTestCase[*aws.Config, Service]{
		Name: "handles destroy with no ID gracefully",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
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
						"id": core.MappingNodeFromString(""),
					},
				},
			},
		},
		ExpectError:          false,
		DestroyActionsCalled: map[string]any{},
	}
}

func TestLambdaEventSourceMappingResourceDestroySuite(t *testing.T) {
	suite.Run(t, new(LambdaEventSourceMappingResourceDestroySuite))
}
