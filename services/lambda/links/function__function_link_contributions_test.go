//go:build unit

package lambdalinks

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionFunctionLinkContributionsSuite struct {
	suite.Suite
}

func (s *FunctionFunctionLinkContributionsSuite) Test_produce_resource_contributions() {
	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, lambdaservice.Service,
	]{
		{
			// The variable the caller needs to invoke the callee. Stated rather than
			// written: the framework applies what every link contributing to the caller
			// needs it to say as one update of it.
			Name: "contributes the invoke env var to the caller function",
			Input: &provider.LinkProduceResourceContributionsInput{
				LinkUpdateType: provider.LinkUpdateTypeCreate,
				ResourceAInfo:  callerFunctionResourceInfo(),
				ResourceBInfo:  calleeFunctionResourceInfo(),
			},
			ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
				Contributions: []*provider.ResourceContribution{
					{
						ResourceName: "callerFunction",
						FieldPath: fmt.Sprintf(
							"spec.environment.variables[%q]", ffLinkEnvVarName,
						),
						Value:  core.MappingNodeFromString(ffLinkCalleeARN),
						Action: provider.ContributionActionSet,
					},
				},
			},
		},
		{
			// Contributions are desired state, so a link that no longer wants the
			// variable withdraws it by contributing nothing.
			Name: "contributes nothing when env var population is disabled",
			Input: &provider.LinkProduceResourceContributionsInput{
				LinkUpdateType: provider.LinkUpdateTypeUpdate,
				ResourceAInfo:  callerFunctionEnvVarsDisabled(),
				ResourceBInfo:  calleeFunctionResourceInfo(),
			},
			ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{},
		},
		{
			Name: "returns an error when the callee function ARN is missing",
			Input: &provider.LinkProduceResourceContributionsInput{
				LinkUpdateType: provider.LinkUpdateTypeCreate,
				ResourceAInfo:  callerFunctionResourceInfo(),
				ResourceBInfo:  calleeFunctionNoARNResourceInfo(),
			},
			ExpectError:          true,
			ExpectedErrorMessage: "function ARN could not be retrieved",
		},
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, ffLinkFactory(iammock.CreateIamServiceMock()), &s.Suite,
	)
}

// A caller function whose annotations turn the link's environment variable off.
func callerFunctionEnvVarsDisabled() *provider.ResourceInfo {
	info := callerFunctionResourceInfo()
	info.ResourceWithResolvedSubs = &provider.ResolvedResource{
		Metadata: &provider.ResolvedResourceMetadata{
			Annotations: core.MappingNodeFields(
				"aws.lambda.invoke.populateEnvVars",
				core.MappingNodeFromBool(false),
			),
		},
	}

	return info
}

func TestFunctionFunctionLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(FunctionFunctionLinkContributionsSuite))
}
