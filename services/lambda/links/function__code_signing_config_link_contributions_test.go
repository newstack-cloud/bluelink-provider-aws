//go:build unit

package lambdalinks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionCodeSigningConfigLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionCodeSigningConfigLinkContributionsSuite) Test_produce_resource_contributions() {
	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, lambdaservice.Service,
	]{
		{
			// Stated rather than put on the function through the Lambda API: a
			// configuration applied through the API is absent from the function's spec,
			// and the next deployment of the function would take it off.
			Name: "contributes the code signing configuration to the function",
			Input: &provider.LinkProduceResourceContributionsInput{
				LinkUpdateType: provider.LinkUpdateTypeCreate,
				ResourceAInfo: &provider.ResourceInfo{
					ResourceName: "test-function",
					CurrentResourceState: &state.ResourceState{
						SpecData: core.MappingNodeFields(
							"arn", core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test"),
						),
					},
				},
				ResourceBInfo: &provider.ResourceInfo{
					ResourceName: "test-code-signing-config",
					CurrentResourceState: &state.ResourceState{
						SpecData: core.MappingNodeFields(
							"codeSigningConfigArn",
							core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1"),
						),
					},
				},
			},
			ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
				Contributions: []*provider.ResourceContribution{
					{
						ResourceName: "test-function",
						FieldPath:    "spec.codeSigningConfigArn",
						Value: core.MappingNodeFromString(
							"arn:aws:lambda:us-west-2:123456789012:code-signing-config:csc-1",
						),
						Action: provider.ContributionActionSet,
					},
				},
			},
		},
		{
			Name: "returns an error when the code signing config ARN is missing",
			Input: &provider.LinkProduceResourceContributionsInput{
				LinkUpdateType: provider.LinkUpdateTypeCreate,
				ResourceAInfo: &provider.ResourceInfo{
					ResourceName:         "test-function",
					CurrentResourceState: &state.ResourceState{SpecData: core.MappingNodeFields()},
				},
				ResourceBInfo: &provider.ResourceInfo{
					ResourceName:         "test-code-signing-config",
					CurrentResourceState: &state.ResourceState{SpecData: core.MappingNodeFields()},
				},
			},
			ExpectError:          true,
			ExpectedErrorMessage: "code signing config ARN could not be retrieved",
		},
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, FunctionCodeSigningConfigLink, &s.Suite,
	)
}

func TestFunctionCodeSigningConfigLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionCodeSigningConfigLinkContributionsSuite))
}
