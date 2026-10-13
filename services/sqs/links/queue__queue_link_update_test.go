//go:build unit

package sqslinks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	sqsmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/sqs_mock"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type QueueQueueLinkUpdateSuite struct {
	suite.Suite
}

func (s *QueueQueueLinkUpdateSuite) Test_link_update_intermediary_resources() {
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
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	]{
		s.createUpdateLinkIntermediaryResourcesTestCase(linkCtx, loader),
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		testCases,
		QueueQueueLink,
		&s.Suite,
	)
}

func TestQueueQueueLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(QueueQueueLinkUpdateSuite))
}

func (s *QueueQueueLinkUpdateSuite) createUpdateLinkIntermediaryResourcesTestCase(
	linkCtx provider.LinkContext,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
	*aws.Config,
	sqsservice.Service,
	*aws.Config,
	sqsservice.Service,
] {
	service := sqsmock.CreateSQSServiceMock()
	configStore := utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		loader,
		utils.AWSConfigCacheKey,
	)

	serviceFactory := func(awsConfig *aws.Config, providerContext provider.Context) sqsservice.Service {
		return service
	}

	return plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
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
