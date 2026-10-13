//go:build unit

package lambdasqs

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type FunctionQueueLinkUpdateSuite struct {
	suite.Suite
}

func lqGetFunctionOutput(vars map[string]string) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(lqFunctionARN),
			Role:        aws.String(lqRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
		},
	}
}

func (s *FunctionQueueLinkUpdateSuite) Test_update_intermediary_resources_grants_send() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(lqGetFunctionOutput(map[string]string{})),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(lqRoleState()))

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "grants sqs:SendMessage scoped to the queue",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    lqFunctionInfo(nil),
			ResourceBInfo:    lqQueueInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  rs,
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: lqMatchAccessOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return lqMatchAccessPolicy(arg, []string{"sqs:SendMessage"}) },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionQueueLinkFactory(iamSvc),
		&s.Suite,
	)
}

// Checks the queue access statement contains the wanted actions (a subset)
// scoped to the queue ARN.
func lqMatchAccessPolicy(arg any, wantActions []string) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok || aws.ToString(input.RoleName) != lqRoleName {
		return false
	}
	var doc struct {
		Statement []struct {
			Sid      string
			Action   []string
			Resource string
		}
	}
	if err := json.Unmarshal([]byte(aws.ToString(input.PolicyDocument)), &doc); err != nil {
		return false
	}
	for _, statement := range doc.Statement {
		if statement.Sid != "SQSAccessordersQueue" {
			continue
		}
		if statement.Resource != lqQueueARN {
			return false
		}
		actions := map[string]bool{}
		for _, a := range statement.Action {
			actions[a] = true
		}
		for _, want := range wantActions {
			if !actions[want] {
				return false
			}
		}
		return true
	}
	return false
}

func lqMatchAccessOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	summary := map[string]any{}
	if actual != nil {
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields["submitOrderFunctionExecutionRole"] != nil &&
			actual.LinkData.Fields["submitOrderFunctionExecutionRole"].Fields[linkutils.PermissionFieldName] != nil
	}
	return plugintestutils.EqualityCheckValues{
		Expected: map[string]any{"hasStatement": true},
		Actual:   summary,
	}, nil
}

func TestFunctionQueueLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionQueueLinkUpdateSuite))
}
