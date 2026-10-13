//go:build unit

package eventslambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	eventsservice "github.com/newstack-cloud/bluelink-provider-aws/services/events/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionEventBusLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionEventBusLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
	]{
		febContributeEnvVarTestCase(loader),
		febContributeNothingWhenDisabledTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, functionEventBusLinkFactory(), &s.Suite,
	)
}

// The variable the function needs to address the event bus. Stated rather than written:
// the framework applies what every link contributing to the function needs it to say as
// one update of it.
func febContributeEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
	]{
		Name:                    "contributes the event bus name env var to the function",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  functionResourceInfoA(),
			ResourceBInfo:  eventBusResourceInfoB(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName: "publishOrderFunction",
					FieldPath:    "spec.environment.variables[\"EVENT_BUS_orderEventBus\"]",
					Value:        core.MappingNodeFromString("order-events"),
					Action:       provider.ContributionActionSet,
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
func febContributeNothingWhenDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	functionInfo := functionResourceInfoA()
	functionInfo.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.lambda.events.populateEnvVars",
				core.MappingNodeFromBool(false),
			),
		},
	}

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, eventsservice.Service,
	]{
		Name:                    "contributes nothing when env vars are disabled",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeUpdate,
			ResourceAInfo:  functionInfo,
			ResourceBInfo:  eventBusResourceInfoB(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput:   &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func TestFunctionEventBusLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionEventBusLinkContributionsSuite))
}
