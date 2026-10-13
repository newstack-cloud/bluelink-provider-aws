//go:build unit

package lambdasqs

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionQueueLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionQueueLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		lqContributeEnvVarTestCase(loader),
		lqContributeNothingWhenDisabledTestCase(loader),
		lqContributeErrorWhenQueueURLMissingTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, functionQueueLinkFactory(iammock.CreateIamServiceMock()), &s.Suite,
	)
}

// The variable the function needs to address the queue. Stated rather than written: the
// framework applies what every link contributing to the function needs it to say as one
// update of it.
func lqContributeEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "contributes the queue URL env var to the function",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  lqFunctionInfo(nil),
			ResourceBInfo:  lqQueueInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName: "submitOrderFunction",
					FieldPath: fmt.Sprintf(
						"spec.environment.variables[\"SQS_QUEUE_%s\"]",
						"ordersQueue",
					),
					Value:  core.MappingNodeFromString(lqQueueURL),
					Action: provider.ContributionActionSet,
				},
			},
		},
		// The value comes from what the link is handed, so the function is not read and,
		// more importantly, not written.
		ActionsNotCalled: []string{"GetFunction", "UpdateFunctionConfiguration"},
	}
}

// Contributions are desired state, so a link that no longer wants the variable withdraws
// it by contributing nothing. The framework takes it off the function.
func lqContributeNothingWhenDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "contributes nothing when env vars are disabled",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeUpdate,
			ResourceAInfo: lqFunctionInfo(map[string]*core.MappingNode{
				"aws.lambda.sqs.populateEnvVars": core.MappingNodeFromBool(false),
			}),
			ResourceBInfo: lqQueueInfo(),
			LinkContext:   testLinkContext(),
		},
		ExpectedOutput:   &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func lqContributeErrorWhenQueueURLMissingTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:            "returns an error when the queue URL cannot be resolved",
		ServiceFactoryA: func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:    testConfigStore(loader),
		ServiceFactoryB: noopCloudControlServiceFactory,
		ConfigStoreB:    testConfigStore(loader),
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  lqFunctionInfo(nil),
			ResourceBInfo:  &provider.ResourceInfo{ResourceName: "ordersQueue"},
			LinkContext:    testLinkContext(),
		},
		ExpectError:          true,
		ExpectedErrorMessage: "queue URL could not be retrieved",
	}
}

func TestFunctionQueueLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionQueueLinkContributionsSuite))
}
