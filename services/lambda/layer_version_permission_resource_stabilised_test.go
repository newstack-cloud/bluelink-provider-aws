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

type LambdaLayerVersionPermissionsResourceStabilisedSuite struct {
	suite.Suite
}

func (s *LambdaLayerVersionPermissionsResourceStabilisedSuite) Test_stabilised_lambda_layer_version_permissions() {
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
		stabilisedBasicLayerVersionPermissionTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceHasStabilisedTestCases(
		testCases,
		LayerVersionPermissionResource,
		&s.Suite,
	)
}

func stabilisedBasicLayerVersionPermissionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, Service] {
	service := createLambdaServiceMock()

	// Create test data for layer version permission stabilisation check
	specData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString("arn:aws:lambda:us-west-2:123456789012:layer:test-layer:1"),
			"statementId":     core.MappingNodeFromString("test-statement"),
			"action":          core.MappingNodeFromString("lambda:GetLayerVersion"),
			"principal":       core.MappingNodeFromString("123456789012"),
		},
	}

	return plugintestutils.ResourceHasStabilisedTestCase[*aws.Config, Service]{
		Name: "layer version permission is always stabilised",
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
			InstanceID:      "test-instance-id",
			ResourceID:      "test-layer-version-permissions-id",
			ResourceSpec:    specData,
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceHasStabilisedOutput{
			Stabilised: true,
		},
	}
}

func TestLambdaLayerVersionPermissionsResourceStabilised(t *testing.T) {
	suite.Run(t, new(LambdaLayerVersionPermissionsResourceStabilisedSuite))
}
