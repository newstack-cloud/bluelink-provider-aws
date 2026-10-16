//go:build unit

package utils

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AWSConfigTestSuite struct {
	suite.Suite
}

func (s *AWSConfigTestSuite) TestAWSConfigFromProviderContext() {
	tests := []struct {
		name        string
		providerCtx provider.Context
		meta        map[string]*core.MappingNode
		env         map[string]string
		mockLoader  *testutils.MockAWSConfigLoader
		expectError bool
	}{
		{
			name: "successfully loads config with region",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"region": core.ScalarFromString("us-west-2"),
				},
				nil,
			),
			meta: map[string]*core.MappingNode{},
			env:  map[string]string{},
			mockLoader: &testutils.MockAWSConfigLoader{
				LoadDefaultConfigFunc: func(
					ctx context.Context,
					optFns ...func(*config.LoadOptions) error,
				) (aws.Config, error) {
					cfg := aws.Config{}
					for _, opt := range optFns {
						err := opt(&config.LoadOptions{})
						if err != nil {
							return aws.Config{}, err
						}
					}
					return cfg, nil
				},
			},
			expectError: false,
		},
		{
			name: "handles loader error",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env: map[string]string{},
			mockLoader: &testutils.MockAWSConfigLoader{
				LoadDefaultConfigFunc: func(
					ctx context.Context,
					optFns ...func(*config.LoadOptions) error,
				) (aws.Config, error) {
					return aws.Config{}, assert.AnError
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			cfg, err := AWSConfigFromProviderContext(
				context.Background(),
				tt.providerCtx,
				tt.meta,
				tt.env,
				tt.mockLoader,
			)

			if tt.expectError {
				s.Error(err)
				s.Nil(cfg)
			} else {
				s.NoError(err)
				s.NotNil(cfg)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestRegionOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		expectedRegion string
		meta           map[string]*core.MappingNode
	}{
		{
			name: "region from provider config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"region": core.ScalarFromString("us-east-1"),
				},
				nil,
			),
			expectedRegion: "us-east-1",
			meta:           map[string]*core.MappingNode{},
		},
		{
			name: "no region in provider config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedRegion: "",
			meta:           map[string]*core.MappingNode{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := RegionOptions(tt.providerCtx, tt.meta)

			if tt.expectedRegion == "" {
				s.Empty(opts)
			} else {
				s.Len(opts, 1)

				// Create a LoadOptions to test the option function
				loadOpts := &config.LoadOptions{}
				err := opts[0](loadOpts)
				s.NoError(err)
				s.Equal(tt.expectedRegion, loadOpts.Region)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestRetryConfigOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		env            map[string]string
		expectedRetry  aws.RetryMode
		expectedMaxRet int
	}{
		{
			name: "retry mode and max retries from provider config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"retryMode":  core.ScalarFromString("standard"),
					"maxRetries": core.ScalarFromInt(3),
				},
				nil,
			),
			env:            map[string]string{},
			expectedRetry:  aws.RetryModeStandard,
			expectedMaxRet: 3,
		},
		{
			name: "retry mode from env var",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env: map[string]string{
				"AWS_RETRY_MODE": "adaptive",
			},
			expectedRetry:  aws.RetryModeAdaptive,
			expectedMaxRet: 0,
		},
		{
			name: "no retry config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env:            map[string]string{},
			expectedRetry:  "",
			expectedMaxRet: 0,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := RetryConfigOptions(tt.providerCtx, tt.env)

			if tt.expectedRetry == "" && tt.expectedMaxRet == 0 {
				s.Empty(opts)
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			if tt.expectedRetry != "" {
				s.Equal(tt.expectedRetry, loadOpts.RetryMode)
			}
			if tt.expectedMaxRet > 0 {
				s.Equal(tt.expectedMaxRet, loadOpts.RetryMaxAttempts)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestCredentialOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		expectedConfig *config.LoadOptions
	}{
		{
			name: "static credentials",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"accessKeyId":     core.ScalarFromString("test-access-key"),
					"secretAccessKey": core.ScalarFromString("test-secret-key"),
					"sessionToken":    core.ScalarFromString("test-session-token"),
				},
				nil,
			),
			expectedConfig: &config.LoadOptions{
				Credentials: credentials.NewStaticCredentialsProvider(
					"test-access-key",
					"test-secret-key",
					"test-session-token",
				),
			},
		},
		{
			name: "shared credentials files",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"sharedCredentialsFiles": core.ScalarFromString("/path/to/creds1,/path/to/creds2"),
					"sharedConfigFiles":      core.ScalarFromString("/path/to/config1,/path/to/config2"),
					"profile":                core.ScalarFromString("test-profile"),
				},
				nil,
			),
			expectedConfig: &config.LoadOptions{
				SharedConfigFiles:      []string{"/path/to/config1", "/path/to/config2"},
				SharedCredentialsFiles: []string{"/path/to/creds1", "/path/to/creds2"},
				SharedConfigProfile:    "test-profile",
			},
		},
		{
			name: "no credentials",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := CredentialOptions(tt.providerCtx)

			if tt.expectedConfig.Credentials == nil && len(tt.expectedConfig.SharedConfigFiles) == 0 {
				s.Empty(opts)
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			if tt.expectedConfig.Credentials != nil {
				s.NotNil(loadOpts.Credentials)
			}
			if len(tt.expectedConfig.SharedConfigFiles) > 0 {
				s.Equal(tt.expectedConfig.SharedConfigFiles, loadOpts.SharedConfigFiles)
			}
			if tt.expectedConfig.SharedConfigProfile != "" {
				s.Equal(tt.expectedConfig.SharedConfigProfile, loadOpts.SharedConfigProfile)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestSharedEndpointOptions() {
	tests := []struct {
		name         string
		providerCtx  provider.Context
		expectedFIPS aws.FIPSEndpointState
		expectedDual aws.DualStackEndpointState
	}{
		{
			name: "FIPS and dual stack enabled",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"useFIPSEndpoint":      core.ScalarFromBool(true),
					"useDualStackEndpoint": core.ScalarFromBool(true),
				},
				nil,
			),
			expectedFIPS: aws.FIPSEndpointStateEnabled,
			expectedDual: aws.DualStackEndpointStateEnabled,
		},
		{
			name: "FIPS and dual stack disabled",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"useFIPSEndpoint":      core.ScalarFromBool(false),
					"useDualStackEndpoint": core.ScalarFromBool(false),
				},
				nil,
			),
			expectedFIPS: aws.FIPSEndpointStateDisabled,
			expectedDual: aws.DualStackEndpointStateDisabled,
		},
		{
			name: "no endpoint config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedFIPS: aws.FIPSEndpointStateDisabled,
			expectedDual: aws.DualStackEndpointStateDisabled,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := SharedEndpointOptions(tt.providerCtx)

			if tt.expectedFIPS == aws.FIPSEndpointStateUnset &&
				tt.expectedDual == aws.DualStackEndpointStateUnset {
				s.Empty(opts)
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			if tt.expectedFIPS != aws.FIPSEndpointStateDisabled {
				s.Equal(tt.expectedFIPS, loadOpts.UseFIPSEndpoint)
			}
			if tt.expectedDual != aws.DualStackEndpointStateDisabled {
				s.Equal(tt.expectedDual, loadOpts.UseDualStackEndpoint)
			}
		})
	}
}

func (s *AWSConfigTestSuite) TestCertOptions() {
	tests := []struct {
		name        string
		providerCtx provider.Context
		env         map[string]string
		expectError bool
	}{
		{
			name: "custom CA bundle from file",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"customCaBundle": core.ScalarFromString("__testdata/ca-bundle.pem"),
				},
				nil,
			),
			env:         map[string]string{},
			expectError: false,
		},
		{
			name: "custom CA bundle from env",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env: map[string]string{
				"AWS_CA_BUNDLE": "__testdata/ca-bundle.pem",
			},
			expectError: false,
		},
		{
			name: "no CA bundle",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env:         map[string]string{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts, err := CertOptions(tt.providerCtx, tt.env)

			if tt.expectError {
				s.Error(err)
				s.Nil(opts)
				return
			}

			s.NoError(err)
			if len(opts) == 0 {
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			// Note: We can't easily test the actual CA bundle content
			// as it's loaded from a file. The test just verifies that
			// the option functions don't error.
		})
	}
}

func (s *AWSConfigTestSuite) TestHTTPClientOptions() {
	tests := []struct {
		name           string
		providerCtx    provider.Context
		expectedConfig *config.LoadOptions
	}{
		{
			name: "custom HTTP client config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"httpProxy": core.ScalarFromString("http://proxy.example.com:8080"),
				},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
		{
			name: "no HTTP client config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			expectedConfig: &config.LoadOptions{},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := HTTPClientOptions(tt.providerCtx)

			if len(opts) == 0 {
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			// Note: We can't easily test the actual HTTP client configuration
			// as it involves network settings. The test just verifies that
			// the option functions don't error.
		})
	}
}

func (s *AWSConfigTestSuite) TestEC2MetadataServiceOptions() {
	tests := []struct {
		name             string
		providerCtx      provider.Context
		env              map[string]string
		expectedMode     imds.EndpointModeState
		expectedEndpoint string
	}{
		{
			name: "EC2 metadata service config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"ec2MetadataServiceEndpoint":     core.ScalarFromString("http://169.254.169.254"),
					"ec2MetadataServiceEndpointMode": core.ScalarFromString("IPv4"),
				},
				nil,
			),
			env:              map[string]string{},
			expectedMode:     imds.EndpointModeStateIPv4,
			expectedEndpoint: "http://169.254.169.254",
		},
		{
			name: "EC2 metadata service from env",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env: map[string]string{
				"AWS_EC2_METADATA_SERVICE_ENDPOINT":      "http://169.254.169.254",
				"AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE": "IPv4",
			},
			expectedMode:     imds.EndpointModeStateIPv4,
			expectedEndpoint: "http://169.254.169.254",
		},
		{
			name: "no EC2 metadata service config",
			providerCtx: plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{},
				nil,
			),
			env:              map[string]string{},
			expectedMode:     imds.EndpointModeStateIPv4,
			expectedEndpoint: "",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := EC2MetadataServiceOptions(tt.providerCtx, tt.env)

			if tt.expectedMode == imds.EndpointModeStateIPv4 && tt.expectedEndpoint == "" {
				s.Empty(opts)
				return
			}

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				err := opt(loadOpts)
				s.NoError(err)
			}

			// Note: We can't easily test the actual EC2 metadata service configuration
			// as it involves network settings. The test just verifies that
			// the option functions don't error.
		})
	}
}

func (s *AWSConfigTestSuite) TestAWSConfigCacheKey() {
	tests := []struct {
		name      string
		sessionID string
		meta      map[string]*core.MappingNode
		expected  string
	}{
		{
			name:      "no meta",
			sessionID: "test-session-id",
			meta:      map[string]*core.MappingNode{},
			expected:  "test-session-id",
		},
		{
			name:      "with meta",
			sessionID: "test-session-id",
			meta: map[string]*core.MappingNode{
				"region": core.MappingNodeFromString("us-east-1"),
			},
			expected: "test-session-id--region-us-east-1",
		},
		{
			name:      "nil meta",
			sessionID: "test-session-id",
			meta:      nil,
			expected:  "test-session-id",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			actual := AWSConfigCacheKey(tt.sessionID, tt.meta)
			s.Equal(tt.expected, actual)
		})
	}
}

// The provider's own defaults for these fields are written with a "~", and the AWS SDK
// takes the paths literally, so an unexpanded one resolves against the working directory.
// The file is then missing and the SDK reports the named profile as non-existent, which
// points at the profile rather than the path.
func (s *AWSConfigTestSuite) Test_credential_options_resolve_shared_file_paths() {
	home, err := os.UserHomeDir()
	s.Require().NoError(err)

	tests := []struct {
		name                     string
		sharedCredentialsFiles   string
		sharedConfigFiles        string
		expectedCredentialsFiles []string
		expectedConfigFiles      []string
	}{
		{
			name:                     "a leading home marker is resolved",
			sharedCredentialsFiles:   "~/.aws/credentials",
			sharedConfigFiles:        "~",
			expectedCredentialsFiles: []string{filepath.Join(home, ".aws/credentials")},
			expectedConfigFiles:      []string{home},
		},
		{
			// Only a leading marker is a home reference.
			name:                     "other paths are left alone",
			sharedCredentialsFiles:   "relative/credentials",
			sharedConfigFiles:        "/etc/~/config",
			expectedCredentialsFiles: []string{"relative/credentials"},
			expectedConfigFiles:      []string{"/etc/~/config"},
		},
		{
			name:                     "the separator after the marker follows the platform",
			sharedCredentialsFiles:   `~\.aws\credentials`,
			expectedCredentialsFiles: []string{backslashHomePath(home)},
		},
		{
			// The value is a comma-separated list, which authors write with spaces.
			name:                     "a list is split and trimmed",
			sharedCredentialsFiles:   "/etc/aws/credentials, ~/.aws/credentials",
			sharedConfigFiles:        "/etc/aws/config, ~/.aws/config",
			expectedCredentialsFiles: []string{"/etc/aws/credentials", filepath.Join(home, ".aws/credentials")},
			expectedConfigFiles:      []string{"/etc/aws/config", filepath.Join(home, ".aws/config")},
		},
		{
			// An empty entry would reach the SDK as a path to a file that cannot exist,
			// replacing the search list with nothing.
			name:                     "empty entries are dropped",
			sharedCredentialsFiles:   "/etc/aws/credentials,",
			sharedConfigFiles:        "",
			expectedCredentialsFiles: []string{"/etc/aws/credentials"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := CredentialOptions(plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"sharedCredentialsFiles": core.ScalarFromString(tt.sharedCredentialsFiles),
					"sharedConfigFiles":      core.ScalarFromString(tt.sharedConfigFiles),
				},
				nil,
			))

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				s.Require().NoError(opt(loadOpts))
			}

			s.Equal(tt.expectedCredentialsFiles, nilWhenEmpty(loadOpts.SharedCredentialsFiles))
			s.Equal(tt.expectedConfigFiles, nilWhenEmpty(loadOpts.SharedConfigFiles))
		})
	}
}

// Nothing else expands these: the value never passes through a shell, and neither the AWS
// SDK nor any platform's file APIs resolve an environment reference, so an unexpanded
// "%USERPROFILE%" or "$HOME" reaches the filesystem verbatim and is reported as a missing
// profile.
func (s *AWSConfigTestSuite) Test_credential_options_expand_environment_references() {
	s.T().Setenv("USERPROFILE", `C:\Users\orders`)
	s.T().Setenv("ProgramFiles(x86)", `C:\Program Files (x86)`)
	s.T().Setenv("AWS_HOME", "/home/orders")

	tests := []struct {
		name                     string
		sharedCredentialsFiles   string
		expectedCredentialsFiles []string
	}{
		{
			name:                     "a reference is replaced by its value",
			sharedCredentialsFiles:   `%USERPROFILE%\.aws\credentials`,
			expectedCredentialsFiles: []string{`C:\Users\orders\.aws\credentials`},
		},
		{
			// %ProgramFiles(x86)% is a real variable, so the name has to allow parentheses.
			name:                     "a name with parentheses in it",
			sharedCredentialsFiles:   `%ProgramFiles(x86)%\aws\credentials`,
			expectedCredentialsFiles: []string{`C:\Program Files (x86)\aws\credentials`},
		},
		{
			// Collapsing it to nothing would turn a typo into a different path that looks
			// plausible, and the error would name that path rather than the reference.
			name:                     "a reference to a variable that is not set is left as written",
			sharedCredentialsFiles:   `%NOT_SET_ANYWHERE%\.aws\credentials`,
			expectedCredentialsFiles: []string{`%NOT_SET_ANYWHERE%\.aws\credentials`},
		},
		{
			// A path separator cannot appear in a name, so a stray marker stays put.
			name:                     "a stray percent sign does not start a reference",
			sharedCredentialsFiles:   "/mnt/50%/aws/credentials",
			expectedCredentialsFiles: []string{"/mnt/50%/aws/credentials"},
		},
		{
			name:                     "a bare POSIX reference",
			sharedCredentialsFiles:   "$AWS_HOME/.aws/credentials",
			expectedCredentialsFiles: []string{"/home/orders/.aws/credentials"},
		},
		{
			// The braced form has to win over the bare one, which would otherwise read the
			// name as "AWS_HOME" and leave the closing brace behind.
			name:                     "a braced POSIX reference",
			sharedCredentialsFiles:   "${AWS_HOME}/.aws/credentials",
			expectedCredentialsFiles: []string{"/home/orders/.aws/credentials"},
		},
		{
			name:                     "an unset POSIX reference is left as written",
			sharedCredentialsFiles:   "$NOT_SET_ANYWHERE/.aws/credentials",
			expectedCredentialsFiles: []string{"$NOT_SET_ANYWHERE/.aws/credentials"},
		},
		{
			// A dollar is an ordinary character in a POSIX file name.
			name:                     "a stray dollar sign does not start a reference",
			sharedCredentialsFiles:   "/mnt/cost$/aws/credentials",
			expectedCredentialsFiles: []string{"/mnt/cost$/aws/credentials"},
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			opts := CredentialOptions(plugintestutils.NewTestProviderContext(
				"aws",
				map[string]*core.ScalarValue{
					"sharedCredentialsFiles": core.ScalarFromString(tt.sharedCredentialsFiles),
				},
				nil,
			))

			loadOpts := &config.LoadOptions{}
			for _, opt := range opts {
				s.Require().NoError(opt(loadOpts))
			}

			s.Equal(tt.expectedCredentialsFiles, nilWhenEmpty(loadOpts.SharedCredentialsFiles))
		})
	}
}

// What the platform makes of a backslash after the marker: a path separator on Windows,
// where "~\.aws\credentials" is a home reference, and an ordinary character in a file
// name everywhere else.
func backslashHomePath(home string) string {
	if runtime.GOOS == "windows" {
		return filepath.Join(home, ".aws", "credentials")
	}
	return `~\.aws\credentials`
}

// The SDK's load options hold an empty slice where no path survived, which the test
// tables express as no expectation at all.
func nilWhenEmpty(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}
	return paths
}

func TestAWSConfigSuite(t *testing.T) {
	suite.Run(t, new(AWSConfigTestSuite))
}
