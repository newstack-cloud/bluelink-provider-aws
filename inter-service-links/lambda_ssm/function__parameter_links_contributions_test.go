//go:build unit

package lambdassm

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionParameterLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionParameterLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		fpContributeEnvVarTestCase(loader),
		fpContributeNothingWhenDisabledTestCase(loader),
		fpContributeErrorWhenMissingTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, functionParameterLinkFactory(iammock.CreateIamServiceMock()), &s.Suite,
	)
}

// The variable the function needs to address the parameter name. Stated rather than written: the
// framework applies what every link contributing to the function needs it to say as one
// update of it.
func fpContributeEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "contributes the parameter name env var to the function",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  fpFunctionInfo(),
			ResourceBInfo:  fpParameterInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName: "apiFunction",
					FieldPath:    "spec.environment.variables[\"" + fpEnvVarName + "\"]",
					Value:        core.MappingNodeFromString(testParameterName),
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
func fpContributeNothingWhenDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	functionInfo := fpFunctionInfo()
	functionInfo.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.lambda.ssm.populateEnvVars",
				core.MappingNodeFromBool(false),
			),
		},
	}

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "contributes nothing when env vars are disabled",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeUpdate,
			ResourceAInfo:  functionInfo,
			ResourceBInfo:  fpParameterInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput:   &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func fpContributeErrorWhenMissingTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:            "returns an error when the parameter name cannot be resolved",
		ServiceFactoryA: func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:    testConfigStore(loader),
		ConfigStoreB:    testConfigStore(loader),
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  fpFunctionInfo(),
			ResourceBInfo:  &provider.ResourceInfo{ResourceName: "missing"},
			LinkContext:    testLinkContext(),
		},
		ExpectError:          true,
		ExpectedErrorMessage: "parameter name could not be retrieved",
	}
}

func TestFunctionParameterLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionParameterLinkContributionsSuite))
}

type FunctionParameterPathLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionParameterPathLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		fppContributeEnvVarTestCase(loader),
		fppContributeNothingWhenDisabledTestCase(loader),
		fppContributeErrorWhenMissingTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, functionParameterPathLinkFactory(iammock.CreateIamServiceMock()), &s.Suite,
	)
}

// The variable the function needs to address the path prefix. Stated rather than written: the
// framework applies what every link contributing to the function needs it to say as one
// update of it.
func fppContributeEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "contributes the path prefix env var to the function",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  fpFunctionInfo(),
			ResourceBInfo:  fppParameterPathInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName: "apiFunction",
					FieldPath:    "spec.environment.variables[\"" + fppDefaultEnvVarName + "\"]",
					Value:        core.MappingNodeFromString(fppPath),
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
func fppContributeNothingWhenDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	functionInfo := fpFunctionInfo()
	functionInfo.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.lambda.ssm.populateEnvVars",
				core.MappingNodeFromBool(false),
			),
		},
	}

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "contributes nothing when env vars are disabled",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeUpdate,
			ResourceAInfo:  functionInfo,
			ResourceBInfo:  fppParameterPathInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput:   &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func fppContributeErrorWhenMissingTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:            "returns an error when the path prefix cannot be resolved",
		ServiceFactoryA: func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:    testConfigStore(loader),
		ConfigStoreB:    testConfigStore(loader),
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  fpFunctionInfo(),
			ResourceBInfo:  &provider.ResourceInfo{ResourceName: "missing"},
			LinkContext:    testLinkContext(),
		},
		ExpectError:          true,
		ExpectedErrorMessage: "parameter path could not be retrieved",
	}
}

func TestFunctionParameterPathLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionParameterPathLinkContributionsSuite))
}

type FunctionParameterTreeLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionParameterTreeLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		fptContributeEnvVarTestCase(loader),
		fptContributeNothingWhenDisabledTestCase(loader),
		fptContributeErrorWhenMissingTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, functionParameterTreeLinkFactory(iammock.CreateIamServiceMock()), &s.Suite,
	)
}

// The variable the function needs to address the tree path prefix. Stated rather than written: the
// framework applies what every link contributing to the function needs it to say as one
// update of it.
func fptContributeEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "contributes the tree path prefix env var to the function",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  fpFunctionInfo(),
			ResourceBInfo:  fptParameterTreeInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName: "apiFunction",
					FieldPath:    "spec.environment.variables[\"" + fptDefaultEnvVarName + "\"]",
					Value:        core.MappingNodeFromString(fptPath),
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
func fptContributeNothingWhenDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	functionInfo := fpFunctionInfo()
	functionInfo.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.lambda.ssm.populateEnvVars",
				core.MappingNodeFromBool(false),
			),
		},
	}

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:                    "contributes nothing when env vars are disabled",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeUpdate,
			ResourceAInfo:  functionInfo,
			ResourceBInfo:  fptParameterTreeInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput:   &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{"UpdateFunctionConfiguration"},
	}
}

func fptContributeErrorWhenMissingTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, ssmservice.Service,
	]{
		Name:            "returns an error when the tree path prefix cannot be resolved",
		ServiceFactoryA: func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:    testConfigStore(loader),
		ConfigStoreB:    testConfigStore(loader),
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  fpFunctionInfo(),
			ResourceBInfo:  &provider.ResourceInfo{ResourceName: "missing"},
			LinkContext:    testLinkContext(),
		},
		ExpectError:          true,
		ExpectedErrorMessage: "parameter path could not be retrieved",
	}
}

func TestFunctionParameterTreeLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionParameterTreeLinkContributionsSuite))
}
