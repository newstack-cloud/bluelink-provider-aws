//go:build unit

package ssm

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ssmmock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ssm_mock"
	ssmservice "github.com/newstack-cloud/bluelink-provider-aws/services/ssm/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type SSMParameterTreeThrottleSuite struct {
	suite.Suite
}

func (s *SSMParameterTreeThrottleSuite) treeResource(
	service ssmservice.Service,
) provider.Resource {
	loader := &testutils.MockAWSConfigLoader{}
	return ParameterTreeResource(
		func(*aws.Config, provider.Context) ssmservice.Service { return service },
		utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
	)
}

func (s *SSMParameterTreeThrottleSuite) providerCtx() provider.Context {
	return plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func (s *SSMParameterTreeThrottleSuite) Test_create_surfaces_throttling_as_retryable() {
	// The codes Parameter Store reports when a burst of writes exceeds the account's
	// throughput, which is what a tree of any size produces.
	for _, code := range []string{
		"ThrottlingException",
		"TooManyUpdates",
		"RequestLimitExceeded",
	} {
		service := ssmmock.CreateSSMServiceMock(
			ssmmock.WithPutParameterError(&smithy.GenericAPIError{
				Code:    code,
				Message: "Rate exceeded",
			}),
		)

		_, err := s.treeResource(service).Deploy(
			context.Background(),
			deployInput(s.providerCtx(), treeSpecData()),
		)

		s.Require().Error(err)
		retryable := &provider.RetryableError{}
		s.Assert().True(
			errors.As(err, &retryable),
			"a write throttled with %s must reach the deployment as retryable, got %T: %v",
			code, err, err,
		)
	}
}

func (s *SSMParameterTreeThrottleSuite) Test_create_still_fails_on_a_rejected_request() {
	// A rejected request does not become valid by being sent again, so retrying one only
	// delays the failure the operator needs to see. The plain error stands in for a
	// transport or SDK failure, which carries no API code to classify on.
	for _, failure := range []error{
		&smithy.GenericAPIError{Code: "InvalidKeyId", Message: "the KMS key is not usable"},
		&smithy.GenericAPIError{Code: "ParameterLimitExceeded", Message: "too many parameters"},
		&smithy.GenericAPIError{Code: "AccessDeniedException", Message: "not authorised"},
		errors.New("a plain error"),
	} {
		service := ssmmock.CreateSSMServiceMock(
			ssmmock.WithPutParameterError(failure),
		)

		_, err := s.treeResource(service).Deploy(
			context.Background(),
			deployInput(s.providerCtx(), treeSpecData()),
		)

		s.Require().Error(err)
		retryable := &provider.RetryableError{}
		s.Assert().False(
			errors.As(err, &retryable),
			"%v must not be retried", failure,
		)
	}
}

func TestSSMParameterTreeThrottleSuite(t *testing.T) {
	suite.Run(t, new(SSMParameterTreeThrottleSuite))
}
