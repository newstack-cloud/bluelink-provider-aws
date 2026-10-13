//go:build unit

package lambdasns

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionTopicLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionTopicLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		ltContributeEnvVarTestCase(loader),
		ltContributeNothingWhenDisabledTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases,
		functionTopicLinkFactory(),
		&s.Suite,
	)
}

// The variable the function needs to address the topic ARN. Stated rather than written: the
// framework applies what every link contributing to the function needs it to say as one
// update of it.
func ltContributeEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "contributes the topic ARN env var to the function",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         noopCloudControlServiceFactory,
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  ltFunctionInfo(),
			ResourceBInfo:  ltTopicInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName: "publishOrderFunction",
					FieldPath:    "spec.environment.variables[\"SNS_TOPIC_ordersTopic\"]",
					Value:        core.MappingNodeFromString(ltTopicARN),
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
func ltContributeNothingWhenDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	functionInfo := ltFunctionInfo()
	functionInfo.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.lambda.sns.populateEnvVars",
				core.MappingNodeFromBool(false),
			),
		},
	}

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
			ResourceAInfo:  functionInfo,
			ResourceBInfo:  ltTopicInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput:   &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func TestFunctionTopicLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionTopicLinkContributionsSuite))
}
