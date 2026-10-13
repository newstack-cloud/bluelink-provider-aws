//go:build unit

package elasticachesecretsmanager

import (
	"context"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	elasticachemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/elasticache_mock"
	secretsmanagermock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/secretsmanager_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	elasticacheservice "github.com/newstack-cloud/bluelink-provider-aws/services/elasticache/service"
	secretsmanagerservice "github.com/newstack-cloud/bluelink-provider-aws/services/secretsmanager/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type ReplicationGroupSecretLinkUpdateSuite struct {
	suite.Suite
}

func matchModifyReplicationGroup(strategy elasticachetypes.AuthTokenUpdateStrategyType) func(arg any) bool {
	return func(arg any) bool {
		input, ok := arg.(*elasticache.ModifyReplicationGroupInput)
		if !ok {
			return false
		}
		return aws.ToString(input.ReplicationGroupId) == rgReplicationGroupID &&
			aws.ToString(input.AuthToken) == rgAuthTokenValue &&
			input.AuthTokenUpdateStrategy == strategy &&
			aws.ToBool(input.ApplyImmediately)
	}
}

func (s *ReplicationGroupSecretLinkUpdateSuite) harnessCase(
	name string,
	input *provider.LinkUpdateLinkedResourcesInput,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.LinkUpdateResourceTestCase[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	cloudcontrolservice.Service,
] {
	return plugintestutils.LinkUpdateResourceTestCase[
		*aws.Config,
		cloudcontrolservice.Service,
		*aws.Config,
		cloudcontrolservice.Service,
	]{
		Name:            name,
		ServiceFactoryA: noopCloudControlServiceFactory,
		ConfigStoreA:    testConfigStore(loader),
		ServiceFactoryB: noopCloudControlServiceFactory,
		ConfigStoreB:    testConfigStore(loader),
		Input:           input,
		ExpectedOutputMatcher: func(
			actual *provider.LinkUpdateLinkedResourcesOutput,
		) (plugintestutils.EqualityCheckValues, error) {
			return plugintestutils.EqualityCheckValues{Expected: true, Actual: actual != nil}, nil
		},
	}
}

// Runs the link's update phase directly, which the harness above cannot do for the paths
// that fail before any AWS call.
func (s *ReplicationGroupSecretLinkUpdateSuite) runUpdate(
	input *provider.LinkUpdateLinkedResourcesInput,
	elastiCacheSvc elasticacheservice.Service,
	secretsManagerSvc secretsmanagerservice.Service,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	loader := &testutils.MockAWSConfigLoader{}
	link := replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc)(
		pluginutils.LinkServiceDeps[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{
			ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
				ServiceFactory: noopCloudControlServiceFactory,
				ConfigStore:    testConfigStore(loader),
			},
			ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
				ServiceFactory: noopCloudControlServiceFactory,
				ConfigStore:    testConfigStore(loader),
			},
		},
	)

	return link.UpdateLinkedResources(context.Background(), input)
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_applies_auth_token_with_rotate_strategy_on_create() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	input := &provider.LinkUpdateLinkedResourcesInput{
		LinkUpdateType: provider.LinkUpdateTypeCreate,
		ResourceAInfo:  replicationGroupInfo(nil),
		ResourceBInfo:  secretInfo(),
		LinkContext:    testLinkContext(),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		[]plugintestutils.LinkUpdateResourceTestCase[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{s.harnessCase("applies auth token on create", input, loader)},
		replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc),
		&s.Suite,
	)

	secretsManagerSvc.AssertCalledWith(
		&s.Suite, "GetSecretValue", 0, plugintestutils.Any,
		func(arg any) bool {
			in, ok := arg.(*secretsmanager.GetSecretValueInput)
			return ok && aws.ToString(in.SecretId) == rgSecretARN
		},
	)
	// ROTATE is the only strategy ElastiCache accepts for a server without an existing AUTH
	// token, so it is the default on first configuration.
	elastiCacheSvc.AssertCalledWith(
		&s.Suite, "ModifyReplicationGroup", 0, plugintestutils.Any,
		matchModifyReplicationGroup(elasticachetypes.AuthTokenUpdateStrategyTypeRotate),
	)
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_rotates_auth_token_on_update() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	input := &provider.LinkUpdateLinkedResourcesInput{
		LinkUpdateType: provider.LinkUpdateTypeUpdate,
		ResourceAInfo:  replicationGroupInfo(nil),
		ResourceBInfo:  secretInfo(),
		LinkContext:    testLinkContext(),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		[]plugintestutils.LinkUpdateResourceTestCase[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{s.harnessCase("rotates auth token on update", input, loader)},
		replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc),
		&s.Suite,
	)

	elastiCacheSvc.AssertCalledWith(
		&s.Suite, "ModifyReplicationGroup", 0, plugintestutils.Any,
		matchModifyReplicationGroup(elasticachetypes.AuthTokenUpdateStrategyTypeRotate),
	)
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_annotation_overrides_update_strategy() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	annotations := map[string]*core.MappingNode{
		"aws.elasticache.secretsmanager.authTokenUpdateStrategy": core.MappingNodeFromString("SET"),
	}
	input := &provider.LinkUpdateLinkedResourcesInput{
		LinkUpdateType: provider.LinkUpdateTypeUpdate,
		ResourceAInfo:  replicationGroupInfo(annotations),
		ResourceBInfo:  secretInfo(),
		LinkContext:    testLinkContext(),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		[]plugintestutils.LinkUpdateResourceTestCase[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{s.harnessCase("annotation overrides strategy", input, loader)},
		replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc),
		&s.Suite,
	)

	// The annotation forces SET on an update (which would otherwise default to ROTATE),
	// retiring the previous token after a rotation window.
	elastiCacheSvc.AssertCalledWith(
		&s.Suite, "ModifyReplicationGroup", 0, plugintestutils.Any,
		matchModifyReplicationGroup(elasticachetypes.AuthTokenUpdateStrategyTypeSet),
	)
}

// A SET annotation is safe as a static value: on first configuration there is no previous
// token to retire, so the link falls back to ROTATE instead of failing.
func (s *ReplicationGroupSecretLinkUpdateSuite) Test_set_strategy_falls_back_to_rotate_on_first_configuration() {
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	annotations := map[string]*core.MappingNode{
		"aws.elasticache.secretsmanager.authTokenUpdateStrategy": core.MappingNodeFromString("SET"),
	}

	_, err := s.runUpdate(
		&provider.LinkUpdateLinkedResourcesInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			ResourceAInfo:  replicationGroupInfo(annotations),
			ResourceBInfo:  secretInfo(),
			LinkContext:    testLinkContext(),
		},
		elastiCacheSvc,
		secretsManagerSvc,
	)

	s.Require().NoError(err)
	elastiCacheSvc.AssertCalledWith(
		&s.Suite, "ModifyReplicationGroup", 0, plugintestutils.Any,
		matchModifyReplicationGroup(elasticachetypes.AuthTokenUpdateStrategyTypeRotate),
	)
}

func (s *ReplicationGroupSecretLinkUpdateSuite) Test_destroy_is_a_no_op() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock()
	secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock()

	input := &provider.LinkUpdateLinkedResourcesInput{
		LinkUpdateType: provider.LinkUpdateTypeDestroy,
		ResourceAInfo:  replicationGroupInfo(nil),
		ResourceBInfo:  secretInfo(),
		LinkContext:    testLinkContext(),
	}

	plugintestutils.RunLinkUpdateResourceTestCases(
		[]plugintestutils.LinkUpdateResourceTestCase[
			*aws.Config, cloudcontrolservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{s.harnessCase("destroy is a no-op", input, loader)},
		replicationGroupSecretLinkFactory(elastiCacheSvc, secretsManagerSvc),
		&s.Suite,
	)

	elastiCacheSvc.AssertNotCalled(&s.Suite, "ModifyReplicationGroup")
	secretsManagerSvc.AssertNotCalled(&s.Suite, "GetSecretValue")
}

// Test_auth_token_value_never_appears_in_output_state asserts that the link data returned from a
// successful create records only the secret ARN and the strategy, never the token value.
func (s *ReplicationGroupSecretLinkUpdateSuite) Test_auth_token_value_never_appears_in_output_state() {
	loader := &testutils.MockAWSConfigLoader{}
	elastiCacheFactory := elasticachemock.CreateElastiCacheServiceMockFactory(
		elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
	)
	secretsManagerFactory := secretsmanagermock.CreateSecretsManagerServiceMockFactory(
		secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
			SecretString: aws.String(rgAuthTokenValue),
		}),
	)

	actions := &replicationGroupSecretLinkActions{
		elastiCacheServiceFactory:    elastiCacheFactory,
		secretsManagerServiceFactory: secretsManagerFactory,
		awsConfigStore:               testConfigStore(loader),
	}

	output, err := actions.UpdateLinkedResources(context.Background(), &provider.LinkUpdateLinkedResourcesInput{
		LinkUpdateType: provider.LinkUpdateTypeCreate,
		ResourceAInfo:  replicationGroupInfo(nil),
		ResourceBInfo:  secretInfo(),
		LinkContext:    testLinkContext(),
	})
	s.Require().NoError(err)
	s.Require().NotNil(output)

	// The secret ARN and strategy marker are recorded, but never the token value.
	authTokenData := output.LinkData.Fields["authToken"]
	s.Require().NotNil(authTokenData)
	s.Equal(rgSecretARN, core.StringValue(authTokenData.Fields["secretArn"]))
	s.True(core.BoolValue(authTokenData.Fields["configured"]))
	assertNoTokenValue(&s.Suite, output.LinkData)
}

func assertNoTokenValue(s *suite.Suite, node *core.MappingNode) {
	if node == nil {
		return
	}
	s.NotEqual(rgAuthTokenValue, core.StringValue(node))
	for _, field := range node.Fields {
		assertNoTokenValue(s, field)
	}
	for _, item := range node.Items {
		assertNoTokenValue(s, item)
	}
}

// The token reaches the link as the secret's value, so a token ElastiCache would reject is
// caught before ModifyReplicationGroup is called: the API's own rejection names neither the
// secret the value came from nor what is wrong with it.
func (s *ReplicationGroupSecretLinkUpdateSuite) Test_rejects_an_auth_token_elasticache_would_not_accept() {
	testCases := []struct {
		name          string
		token         string
		errorContains string
	}{
		{
			name:  "accepts a valid token",
			token: rgAuthTokenValue,
		},
		{
			name:          "rejects a token shorter than 16 characters",
			token:         "too-short",
			errorContains: "between 16 and 128 characters",
		},
		{
			name:          "rejects a token longer than 128 characters",
			token:         strings.Repeat("a", 129),
			errorContains: "between 16 and 128 characters",
		},
		{
			name:          "rejects non-printable-ASCII characters",
			token:         "valid-length-token-with-é-in-it",
			errorContains: "printable ASCII",
		},
		{
			name:          "rejects forbidden special characters",
			token:         "valid-length-token-with-@-in-it",
			errorContains: `may contain only letters, digits and the characters`,
		},
		{
			// ElastiCache allows only eight printable special characters, so punctuation that is
			// neither a delimiter nor obviously unsafe is still rejected by the API.
			// Accepting it here just moves the failure to ModifyReplicationGroup.
			name:          "rejects punctuation outside the allowed special characters",
			token:         "valid-length-token-with-(-in-it",
			errorContains: `may contain only letters, digits and the characters`,
		},
		{
			name:  "accepts the allowed special characters",
			token: "valid-length-token!&#$^<>-ok",
		},
	}

	for _, testCase := range testCases {
		s.Run(testCase.name, func() {
			elastiCacheSvc := elasticachemock.CreateElastiCacheServiceMock(
				elasticachemock.WithModifyReplicationGroupOutput(&elasticache.ModifyReplicationGroupOutput{}),
			)
			secretsManagerSvc := secretsmanagermock.CreateSecretsManagerServiceMock(
				secretsmanagermock.WithGetSecretValueOutput(&secretsmanager.GetSecretValueOutput{
					SecretString: aws.String(testCase.token),
				}),
			)

			_, err := s.runUpdate(
				&provider.LinkUpdateLinkedResourcesInput{
					LinkUpdateType: provider.LinkUpdateTypeCreate,
					ResourceAInfo:  replicationGroupInfo(nil),
					ResourceBInfo:  secretInfo(),
					LinkContext:    testLinkContext(),
				},
				elastiCacheSvc,
				secretsManagerSvc,
			)

			if testCase.errorContains == "" {
				s.Require().NoError(err)
				elastiCacheSvc.AssertCalled(&s.Suite, "ModifyReplicationGroup")
				return
			}

			s.Require().Error(err)
			s.Contains(err.Error(), testCase.errorContains)
			s.Contains(err.Error(), rgSecretARN)
			// The token is a secret: it must not reach a log or a deployment report.
			s.NotContains(err.Error(), testCase.token)
			elastiCacheSvc.AssertNotCalled(&s.Suite, "ModifyReplicationGroup")
		})
	}
}

func TestReplicationGroupSecretLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(ReplicationGroupSecretLinkUpdateSuite))
}
