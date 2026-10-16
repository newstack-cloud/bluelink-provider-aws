//go:build unit

package utils

import (
	"context"
	"errors"
	"math/rand"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type AWSConfigStoreTestSuite struct {
	suite.Suite
	mockConfigCreator AWSConfigCreator
}

func (s *AWSConfigStoreTestSuite) SetupTest() {
	// Create a mock config creator that returns a simple config
	s.mockConfigCreator = func(
		ctx context.Context,
		providerContext provider.Context,
		meta map[string]*core.MappingNode,
		env map[string]string,
		loader AWSConfigLoader,
	) (*aws.Config, error) {
		var sessionID string
		if v := ctx.Value(pluginutils.ContextSessionIDKey); v != nil {
			sessionID, _ = v.(string)
		} else {
			sessionID = "no-session"
		}
		appID := "test-config-" + sessionID
		if sessionID == "no-session" {
			// Return a unique config each time
			appID += "-" + strconv.Itoa(rand.Int())
		}
		return &aws.Config{
			Region: "us-west-2",
			AppID:  appID,
		}, nil
	}
}

func (s *AWSConfigStoreTestSuite) Test_from_provider_context() {
	env := []string{
		"AWS_EC2_METADATA_SERVICE_ENDPOINT=http://169.254.169.254",
	}
	providerConfig := map[string]*core.ScalarValue{
		"region":               core.ScalarFromString("us-west-2"),
		"retryMode":            core.ScalarFromString("standard"),
		"maxRetries":           core.ScalarFromInt(3),
		"accessKeyId":          core.ScalarFromString("test-access-key"),
		"secretAccessKey":      core.ScalarFromString("test-secret-key"),
		"useFIPSEndpoint":      core.ScalarFromBool(true),
		"useDualStackEndpoint": core.ScalarFromBool(true),
	}
	sessionID := "test-session-1"

	store := NewAWSConfigStore(
		env,
		s.mockConfigCreator,
		&testutils.MockAWSConfigLoader{},
		AWSConfigCacheKey,
	)
	providerContext := plugintestutils.NewTestProviderContext("aws", providerConfig, nil)

	// Create a context with the session ID
	ctx := context.WithValue(context.Background(), pluginutils.ContextSessionIDKey, sessionID)

	// First call should not be cached
	cfg, err := store.FromProviderContext(ctx, providerContext, nil)
	s.NoError(err)
	s.NotNil(cfg)

	// Second call should be cached
	cachedCfg, err := store.FromProviderContext(ctx, providerContext, nil)
	s.NoError(err)
	s.NotNil(cachedCfg)
	s.Equal(cfg, cachedCfg, "Cached config should be identical to original config")

	// Verify cache behavior with different session IDs
	otherSessionCtx := context.WithValue(
		context.Background(),
		pluginutils.ContextSessionIDKey,
		sessionID+"-other",
	)
	otherCfg, err := store.FromProviderContext(otherSessionCtx, providerContext, nil)
	s.NoError(err)
	s.NotNil(otherCfg)
	s.NotEqual(cfg, otherCfg, "Configs for different sessions should be different")
}

func (s *AWSConfigStoreTestSuite) Test_from_provider_context_no_session_id() {
	store := NewAWSConfigStore([]string{}, s.mockConfigCreator, &testutils.MockAWSConfigLoader{}, nil)
	providerContext := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		nil,
	)

	// Call without session ID in context
	cfg, err := store.FromProviderContext(context.Background(), providerContext, nil)
	s.NoError(err)
	s.NotNil(cfg)

	// Second call should still work but won't use cache
	cfg2, err := store.FromProviderContext(context.Background(), providerContext, nil)
	s.NoError(err)
	s.NotNil(cfg2)
	s.NotEqual(cfg, cfg2, "Configs without session ID should not be cached")
}

// A failed construction used to be cached and then served to the next caller as a
// nil config with a nil error, which reads as success and panics at the first use of
// the returned pointer. Harmless while the only per-request variation was a region;
// a correctness bug once the cache key means more.
func (s *AWSConfigStoreTestSuite) Test_does_not_cache_a_failed_config() {
	attempts := 0
	failFirstCreator := func(
		ctx context.Context,
		providerContext provider.Context,
		meta map[string]*core.MappingNode,
		env map[string]string,
		loader AWSConfigLoader,
	) (*aws.Config, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("sts unreachable")
		}
		return &aws.Config{Region: "us-west-2", AppID: "recovered"}, nil
	}

	store := NewAWSConfigStore(
		[]string{},
		failFirstCreator,
		&testutils.MockAWSConfigLoader{},
		AWSConfigCacheKey,
	)
	providerContext := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{"region": core.ScalarFromString("us-west-2")},
		nil,
	)
	ctx := context.WithValue(
		context.Background(),
		pluginutils.ContextSessionIDKey,
		"test-session-failure",
	)

	cfg, err := store.FromProviderContext(ctx, providerContext, nil)
	s.Error(err)
	s.Nil(cfg)

	// The retry must rebuild rather than be served the failure.
	retried, err := store.FromProviderContext(ctx, providerContext, nil)
	s.Require().NoError(err)
	s.Require().NotNil(retried)
	s.Equal("recovered", retried.AppID)
	s.Equal(2, attempts, "a failed config must not satisfy the next request from cache")
}

func TestAWSConfigStoreSuite(t *testing.T) {
	suite.Run(t, new(AWSConfigStoreTestSuite))
}
