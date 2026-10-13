//go:build unit

package lambdalinks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionCodeSigningConfigLinkUpdateSuite struct {
	suite.Suite
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) Test_link_update_intermediary_resources() {
	loader := &testutils.MockAWSConfigLoader{}
	linkCtx := plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {
				"region": core.ScalarFromString("us-west-2"),
			},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		s.createUpdateLinkIntermediaryResourcesTestCase(linkCtx, loader),
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		testCases,
		FunctionCodeSigningConfigLink,
		&s.Suite,
	)
}

func TestFunctionCodeSigningConfigLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionCodeSigningConfigLinkUpdateSuite))
}

func (s *FunctionCodeSigningConfigLinkUpdateSuite) createUpdateLinkIntermediaryResourcesTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	lambdaservice.Service,
] {
	service := lambdamock.CreateLambdaServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) lambdaservice.Service {
		return service
	}

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	]{
		Name:                           "Returns empty link data as there are no intermediary resources to create or update",
		ServiceFactoryA:                serviceFactory,
		ConfigStoreA:                   configStore,
		ServiceFactoryB:                serviceFactory,
		ConfigStoreB:                   configStore,
		IntermediariesServiceMockCalls: &service.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			LinkContext:    linkCtx,
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{
			IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
			LinkData: &core.MappingNode{
				Fields: map[string]*core.MappingNode{},
			},
		},
	}
}
