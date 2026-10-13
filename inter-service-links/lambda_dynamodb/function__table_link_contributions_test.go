//go:build unit

package lambdadynamodb

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	dynamodbmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/dynamodb_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionTableLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionTableLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		contributeTableNameEnvVarTestCase(loader),
		contributeTableNameFromARNTestCase(loader),
		contributeNothingWhenEnvVarsDisabledTestCase(loader),
		contributeErrorWhenTableNameMissingTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases,
		functionTableLinkFactory(iammock.CreateIamServiceMock()),
		&s.Suite,
	)
}

// A function whose annotations turn the link's environment variable off.
func functionWithEnvVarsDisabled() *provider.ResourceInfo {
	info := functionResourceInfoA()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.lambda.dynamodb.ordersTable.populateEnvVars",
				core.MappingNodeFromBool(false),
			),
		},
	}

	return info
}

func envVarContribution(value string) *provider.ResourceContribution {
	return &provider.ResourceContribution{
		ResourceName: "processOrdersFunction",
		FieldPath: fmt.Sprintf(
			"spec.environment.variables[\"%s\"]",
			fdtEnvVarName,
		),
		Value: core.MappingNodeFromString(value),
		// One variable holds one table name, so the contribution states it rather
		// than adding to what is there.
		Action: provider.ContributionActionSet,
	}
}

// The variable the function needs to address the table. Stated rather than written: the
// framework applies what every link contributing to the function needs it to say as one
// update of it.
func contributeTableNameEnvVarTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name:                    "contributes the table name env var to the function",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  functionResourceInfoA(),
			ResourceBInfo:  tableResourceInfoB(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				envVarContribution(fdtTableName),
			},
		},
		// The value is worked out from what the link is given, so the function is not
		// read and, more importantly, not written.
		ActionsNotCalled: []string{
			"GetFunction",
			"UpdateFunctionConfiguration",
		},
	}
}

// A name-less (auto-named) table has no tableName in state, so the link derives the
// physical name from the table ARN instead.
func contributeTableNameFromARNTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	namelessTableInfo := &provider.ResourceInfo{
		ResourceName: "ordersTable",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString(fdtTableARN),
			),
		},
	}

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name:            "derives the table name from the ARN for a name-less table",
		ServiceFactoryA: func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:    testConfigStore(loader),
		ServiceFactoryB: dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:    testConfigStore(loader),
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  functionResourceInfoA(),
			ResourceBInfo:  namelessTableInfo,
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				envVarContribution("orders"),
			},
		},
	}
}

// Contributions are desired state, so a link that no longer wants the variable withdraws
// it by contributing nothing. The framework takes it off the function; there is no call
// here to remove it.
func contributeNothingWhenEnvVarsDisabledTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name:                    "contributes nothing when env vars are disabled",
		ServiceFactoryA:         func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:            testConfigStore(loader),
		ServiceFactoryB:         dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:            testConfigStore(loader),
		CurrentServiceMockCalls: &lambdaSvc.MockCalls,
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeUpdate,
			ResourceAInfo:  functionWithEnvVarsDisabled(),
			ResourceBInfo:  tableResourceInfoB(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{},
		ActionsNotCalled: []string{
			"UpdateFunctionConfiguration",
		},
	}
}

func contributeErrorWhenTableNameMissingTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config,
	lambdaservice.Service,
	*aws.Config,
	dynamodbservice.Service,
] {
	lambdaSvc := lambdamock.CreateLambdaServiceMock()

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		dynamodbservice.Service,
	]{
		Name:            "returns an error when the table name cannot be resolved",
		ServiceFactoryA: func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:    testConfigStore(loader),
		ServiceFactoryB: dynamodbmock.CreateDynamoDBServiceMockFactory(),
		ConfigStoreB:    testConfigStore(loader),
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  functionResourceInfoA(),
			ResourceBInfo: &provider.ResourceInfo{
				ResourceName: "ordersTable",
			},
			LinkContext: testLinkContext(),
		},
		ExpectError:          true,
		ExpectedErrorMessage: "table name could not be retrieved",
	}
}

func TestFunctionTableLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionTableLinkContributionsSuite))
}
