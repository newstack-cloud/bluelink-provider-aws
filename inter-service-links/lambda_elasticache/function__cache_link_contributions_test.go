//go:build unit

package lambdaelasticache

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

type FunctionCacheLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionCacheLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		lcContributeEnvVarsTestCase(loader),
		lcContributeNothingWhenDisabledTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, functionCacheLinkFactory(iammock.CreateIamServiceMock(), noopEC2ServiceFactory()), &s.Suite,
	)
}

// One contribution per variable, ordered by name so that what a link produces does not
// depend on map iteration.
func lcContributeEnvVarsTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                    "contributes the connection env vars to the function",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  lcFunctionInfo(lcEnvVarAnnotations()),
			ResourceBInfo:  lcCacheInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName: "apiFunction",
					FieldPath: fmt.Sprintf(
						"spec.environment.variables[%q]", lcPrefix+"_HOST",
					),
					Value:  core.MappingNodeFromString(testCacheEndpoint),
					Action: provider.ContributionActionSet,
				},
				{
					ResourceName: "apiFunction",
					FieldPath: fmt.Sprintf(
						"spec.environment.variables[%q]", lcPrefix+"_PORT",
					),
					Value:  core.MappingNodeFromString("6379"),
					Action: provider.ContributionActionSet,
				},
			},
		},
		ActionsNotCalled: []string{"GetFunction", "UpdateFunctionConfiguration"},
	}
}

// Contributions are desired state, so a link that no longer wants the variables withdraws
// them by contributing nothing. The framework takes them off the function.
func lcContributeNothingWhenDisabledTestCase(
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
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo: lcFunctionInfo(map[string]*core.MappingNode{
				"aws.lambda.elasticache.sessionCache.populateEnvVars": core.MappingNodeFromBool(false),
			}),
			ResourceBInfo: lcCacheInfo(),
			LinkContext:   testLinkContext(),
		},
		ExpectedOutput:   &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{"GetFunction", "UpdateFunctionConfiguration"},
	}
}

func TestFunctionCacheLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionCacheLinkContributionsSuite))
}
