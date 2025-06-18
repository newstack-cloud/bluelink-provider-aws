package lambda

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/celerity-provider-aws/internal/testutils"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type LambdaAliasResourceStabilisedSuite struct {
	suite.Suite
}

func (s *LambdaAliasResourceStabilisedSuite) Test_stabilised_lambda_alias() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, Service]{
		stabilisedBasicAliasTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		AliasResource,
		&s.Suite,
	)
}

func stabilisedBasicAliasTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock()

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString("test-function"),
			"name":            core.MappingNodeFromString("PROD"),
			"functionVersion": core.MappingNodeFromString("1"),
			"aliasArn":        core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:function:test-function:PROD"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, Service]{
		Name: "basic alias is stabilised",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) Service {
			return service
		},
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceHasStabilisedInput{
			ProviderContext: providerCtx,
			ResourceSpec:    resourceSpecState,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func TestLambdaAliasResourceStabilised(t *testing.T) {
	suite.Run(t, new(LambdaAliasResourceStabilisedSuite))
}
