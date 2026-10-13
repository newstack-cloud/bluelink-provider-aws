//go:build unit

package lambdadynamodb

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	dynamodbservice "github.com/newstack-cloud/bluelink-provider-aws/services/dynamodb/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type TableFunctionLinkContributionsSuite struct {
	suite.Suite
}

func (s *TableFunctionLinkContributionsSuite) Test_produce_resource_contributions() {
	loader := &testutils.MockAWSConfigLoader{}

	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service,
	]{
		tfContributeStreamTestCase(loader),
		tfKeepsRunningStreamViewTypeTestCase(loader),
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, tableFunctionLinkFactory(iammock.CreateIamServiceMock()), &s.Suite,
	)
}

// The stream is what this link consumes, so it states that the table has one. Stated
// rather than enabled through the DynamoDB API: a stream enabled through the API leaves
// the table's spec saying nothing about it, and the next deployment of the table for any
// other reason would turn it off.
func tfContributeStreamTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service,
] {
	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service,
	]{
		Name:         "contributes the stream specification to the table",
		ConfigStoreA: testConfigStore(loader),
		ConfigStoreB: testConfigStore(loader),
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  tableResourceInfoStreamsDisabled(),
			ResourceBInfo:  streamFunctionResourceInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName:    "ordersTable",
					FieldPath:       "spec.streamSpecification.streamViewType",
					Value:           core.MappingNodeFromString("NEW_AND_OLD_IMAGES"),
					Action:          provider.ContributionActionSet,
					RetainOnRemoval: true,
				},
			},
		},
	}
}

// A stream that is already running keeps the view type it was created with. Stating the
// annotation's view type instead would change a live stream because a link was deployed,
// which is not what adding a consumer asks for.
func tfKeepsRunningStreamViewTypeTestCase(
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkProduceContributionsTestCase[
	*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service,
] {
	tableInfo := tableResourceInfoStreamsEnabled()
	tableInfo.CurrentResourceState.SpecData.Fields["streamSpecification"] = core.MappingNodeFields(
		"streamViewType", core.MappingNodeFromString("KEYS_ONLY"),
	)

	return plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, dynamodbservice.Service, *aws.Config, lambdaservice.Service,
	]{
		Name:         "keeps the view type of a stream that is already running",
		ConfigStoreA: testConfigStore(loader),
		ConfigStoreB: testConfigStore(loader),
		Input: &provider.LinkProduceResourceContributionsInput{
			LinkUpdateType: provider.LinkUpdateTypeUpdate,
			ResourceAInfo:  tableInfo,
			ResourceBInfo:  streamFunctionResourceInfo(),
			LinkContext:    testLinkContext(),
		},
		ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
			Contributions: []*provider.ResourceContribution{
				{
					ResourceName:    "ordersTable",
					FieldPath:       "spec.streamSpecification.streamViewType",
					Value:           core.MappingNodeFromString("KEYS_ONLY"),
					Action:          provider.ContributionActionSet,
					RetainOnRemoval: true,
				},
			},
		},
	}
}

func TestTableFunctionLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(TableFunctionLinkContributionsSuite))
}
