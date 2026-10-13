//go:build unit

package lambdards

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

type FunctionProxyLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionProxyLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		fpContributeEnvVarsTestCase(loader),
		fpContributeNothingWhenDisabledTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, functionProxyLinkFactory(iammock.CreateIamServiceMock(), noopEC2ServiceFactory()), &s.Suite,
	)
}

// One contribution per variable, ordered by name so that what a link produces does not
// depend on map iteration. Stated rather than written: the framework applies what every
// link contributing to the function needs it to say as one update of it.
func fpContributeEnvVarsTestCase(
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
			ResourceAInfo:  fpFunctionInfo(fpEnvVarAnnotations()),
			ResourceBInfo:  fpProxyInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName: "apiFunction",
					FieldPath: fmt.Sprintf(
						"spec.environment.variables[%q]", fpPrefix+"_DATABASE",
					),
					Value:  core.MappingNodeFromString("orders"),
					Action: provider.ContributionActionSet,
				},
				{
					ResourceName: "apiFunction",
					FieldPath: fmt.Sprintf(
						"spec.environment.variables[%q]", fpPrefix+"_HOST",
					),
					Value:  core.MappingNodeFromString(testProxyEndpoint),
					Action: provider.ContributionActionSet,
				},
				{
					ResourceName: "apiFunction",
					FieldPath: fmt.Sprintf(
						"spec.environment.variables[%q]", fpPrefix+"_PORT",
					),
					Value:  core.MappingNodeFromString("5432"),
					Action: provider.ContributionActionSet,
				},
			},
		},
		// The values come from what the link is handed, so the function is not read and,
		// more importantly, not written.
		ActionsNotCalled: []string{"GetFunction", "UpdateFunctionConfiguration"},
	}
}

// Contributions are desired state, so a link that no longer wants the variables withdraws
// them by contributing nothing. The framework takes them off the function.
func fpContributeNothingWhenDisabledTestCase(
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
			LinkUpdateType: provider.LinkUpdateTypeUpdate,
			ResourceAInfo: fpFunctionInfo(map[string]*core.MappingNode{
				"aws.lambda.rds.ordersProxy.populateEnvVars": core.MappingNodeFromBool(false),
			}),
			ResourceBInfo: fpProxyInfo(),
			LinkContext:   testLinkContext(),
		},
		ExpectedOutput:   &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func TestFunctionProxyLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionProxyLinkContributionsSuite))
}
