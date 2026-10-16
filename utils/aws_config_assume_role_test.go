//go:build unit

package utils

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

// These drive the resolved credentials to the point of issuing the STS call and
// assert on the request that goes out. Asserting only the credential provider's
// type stays true when every option the config carries is dropped.
type AssumeRoleCredentialsTestSuite struct {
	suite.Suite
}

const stsAssumeRoleResponse = `<AssumeRoleResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <AssumeRoleResult>
    <Credentials>
      <AccessKeyId>ASIAASSUMED</AccessKeyId>
      <SecretAccessKey>assumed-secret</SecretAccessKey>
      <SessionToken>assumed-token</SessionToken>
      <Expiration>2035-01-01T00:00:00Z</Expiration>
    </Credentials>
  </AssumeRoleResult>
</AssumeRoleResponse>`

const stsWebIdentityResponse = `<AssumeRoleWithWebIdentityResponse xmlns="https://sts.amazonaws.com/doc/2011-06-15/">
  <AssumeRoleWithWebIdentityResult>
    <Credentials>
      <AccessKeyId>ASIAWEBID</AccessKeyId>
      <SecretAccessKey>webid-secret</SecretAccessKey>
      <SessionToken>webid-token</SessionToken>
      <Expiration>2035-01-01T00:00:00Z</Expiration>
    </Credentials>
  </AssumeRoleWithWebIdentityResult>
</AssumeRoleWithWebIdentityResponse>`

type recordingHTTPClient struct {
	body     url.Values
	response string
}

func (c *recordingHTTPClient) Do(req *http.Request) (*http.Response, error) {
	if req.Body != nil {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		parsed, err := url.ParseQuery(string(raw))
		if err != nil {
			return nil, err
		}
		c.body = parsed
	}

	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/xml"}},
		Body:       io.NopCloser(strings.NewReader(c.response)),
	}, nil
}

type authorizationCapturingClient struct {
	inner         *recordingHTTPClient
	authorization string
}

func (c *authorizationCapturingClient) Do(req *http.Request) (*http.Response, error) {
	c.authorization = req.Header.Get("Authorization")
	return c.inner.Do(req)
}

// Stands in for LoadDefaultConfig, resolving the ambient credentials a role
// assumption is layered on top of and routing the STS call to the recorder.
func loaderRecording(recorder *recordingHTTPClient) *testutils.MockAWSConfigLoader {
	return &testutils.MockAWSConfigLoader{
		LoadDefaultConfigFunc: func(
			ctx context.Context,
			optFns ...func(*config.LoadOptions) error,
		) (aws.Config, error) {
			loadOpts := &config.LoadOptions{}
			for _, opt := range optFns {
				if err := opt(loadOpts); err != nil {
					return aws.Config{}, err
				}
			}

			return aws.Config{
				Region:     "eu-west-2",
				HTTPClient: recorder,
				Credentials: credentials.NewStaticCredentialsProvider(
					"ambient-key",
					"ambient-secret",
					"",
				),
			}, nil
		},
	}
}

func providerCtxWith(configVars map[string]*core.ScalarValue) provider.Context {
	return plugintestutils.NewTestProviderContext("aws", configVars, nil)
}

// Resolves the credentials, which is what issues the STS call, and returns the
// recorded request parameters.
func (s *AssumeRoleCredentialsTestSuite) resolve(
	configVars map[string]*core.ScalarValue,
	response string,
) url.Values {
	recorder := &recordingHTTPClient{response: response}

	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(configVars),
		map[string]*core.MappingNode{},
		map[string]string{},
		loaderRecording(recorder),
	)
	s.Require().NoError(err)
	s.Require().NotNil(cfg)

	creds, err := cfg.Credentials.Retrieve(context.Background())
	s.Require().NoError(err)
	s.Equal("ASIAASSUMED", creds.AccessKeyID, "the assumed credentials must be the ones served")

	return recorder.body
}

func (s *AssumeRoleCredentialsTestSuite) Test_assumes_the_configured_role_with_every_option() {
	body := s.resolve(map[string]*core.ScalarValue{
		"region":                       core.ScalarFromString("eu-west-2"),
		"assumeRole.roleArn":           core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
		"assumeRole.sessionName":       core.ScalarFromString("celerity-deploy"),
		"assumeRole.externalId":        core.ScalarFromString("external-id"),
		"assumeRole.duration":          core.ScalarFromString("45m"),
		"assumeRole.sourceIdentity":    core.ScalarFromString("deploy-principal"),
		"assumeRole.transitiveTagKeys": core.ScalarFromString("team,env"),
		"assumeRole.tags.team":         core.ScalarFromString("platform"),
		"assumeRole.policyArns.0":      core.ScalarFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
	}, stsAssumeRoleResponse)

	s.Equal("AssumeRole", body.Get("Action"))
	s.Equal("arn:aws:iam::123456789012:role/deploy", body.Get("RoleArn"))
	s.Equal("celerity-deploy", body.Get("RoleSessionName"))
	s.Equal("external-id", body.Get("ExternalId"))
	s.Equal("2700", body.Get("DurationSeconds"))
	s.Equal("deploy-principal", body.Get("SourceIdentity"))
	s.Equal("team", body.Get("TransitiveTagKeys.member.1"))
	s.Equal("env", body.Get("TransitiveTagKeys.member.2"))
	s.Equal("team", body.Get("Tags.member.1.Key"))
	s.Equal("platform", body.Get("Tags.member.1.Value"))
	s.Equal(
		"arn:aws:iam::aws:policy/ReadOnlyAccess",
		body.Get("PolicyArns.member.1.arn"),
	)
}

func (s *AssumeRoleCredentialsTestSuite) Test_assumes_a_web_identity_role_with_its_token() {
	recorder := &recordingHTTPClient{response: stsWebIdentityResponse}

	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"assumeRoleWithWebIdentity.roleArn": core.ScalarFromString(
				"arn:aws:iam::123456789012:role/ci",
			),
			"assumeRoleWithWebIdentity.webIdentityToken": core.ScalarFromString("oidc-token"),
			"assumeRoleWithWebIdentity.sessionName":      core.ScalarFromString("ci-run"),
			"assumeRoleWithWebIdentity.duration":         core.ScalarFromString("30m"),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		loaderRecording(recorder),
	)
	s.Require().NoError(err)

	creds, err := cfg.Credentials.Retrieve(context.Background())
	s.Require().NoError(err)
	s.Equal("ASIAWEBID", creds.AccessKeyID)

	s.Equal("AssumeRoleWithWebIdentity", recorder.body.Get("Action"))
	s.Equal("arn:aws:iam::123456789012:role/ci", recorder.body.Get("RoleArn"))
	s.Equal("oidc-token", recorder.body.Get("WebIdentityToken"))
	s.Equal("ci-run", recorder.body.Get("RoleSessionName"))
	s.Equal("1800", recorder.body.Get("DurationSeconds"))
}

// A role ARN with no token is not usable: installing a provider for it would
// panic inside stscreds on first retrieval. Validation reports it, and the
// runtime falls through to whatever else is configured.
func (s *AssumeRoleCredentialsTestSuite) Test_web_identity_without_a_token_falls_through() {
	body := s.resolve(map[string]*core.ScalarValue{
		"assumeRole.roleArn": core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
		"assumeRoleWithWebIdentity.roleArn": core.ScalarFromString(
			"arn:aws:iam::123456789012:role/ci",
		),
	}, stsAssumeRoleResponse)

	s.Equal(
		"AssumeRole",
		body.Get("Action"),
		"a token-less web identity role must not displace a usable assumeRole config",
	)
	s.Equal("arn:aws:iam::123456789012:role/deploy", body.Get("RoleArn"))
}

func (s *AssumeRoleCredentialsTestSuite) Test_web_identity_without_a_token_is_not_assumed_at_all() {
	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"assumeRoleWithWebIdentity.roleArn": core.ScalarFromString(
				"arn:aws:iam::123456789012:role/ci",
			),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		loaderRecording(&recordingHTTPClient{response: stsWebIdentityResponse}),
	)
	s.Require().NoError(err)

	s.IsType(
		credentials.StaticCredentialsProvider{},
		cfg.Credentials,
		"no usable role means the ambient credentials are left in place",
	)
}

func (s *AssumeRoleCredentialsTestSuite) Test_web_identity_takes_precedence_when_both_are_usable() {
	recorder := &recordingHTTPClient{response: stsWebIdentityResponse}

	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"assumeRole.roleArn": core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
			"assumeRoleWithWebIdentity.roleArn": core.ScalarFromString(
				"arn:aws:iam::123456789012:role/ci",
			),
			"assumeRoleWithWebIdentity.webIdentityToken": core.ScalarFromString("oidc-token"),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		loaderRecording(recorder),
	)
	s.Require().NoError(err)

	_, err = cfg.Credentials.Retrieve(context.Background())
	s.Require().NoError(err)
	s.Equal("AssumeRoleWithWebIdentity", recorder.body.Get("Action"))
	s.Equal("arn:aws:iam::123456789012:role/ci", recorder.body.Get("RoleArn"))
}

// The STS call that assumes the role has to be made as the ambient principal,
// not as the role being assumed.
func (s *AssumeRoleCredentialsTestSuite) Test_the_sts_call_is_signed_with_the_ambient_credentials() {
	recorder := &recordingHTTPClient{response: stsAssumeRoleResponse}
	captured := &authorizationCapturingClient{inner: recorder}

	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"assumeRole.roleArn": core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		&testutils.MockAWSConfigLoader{
			LoadDefaultConfigFunc: func(
				ctx context.Context,
				optFns ...func(*config.LoadOptions) error,
			) (aws.Config, error) {
				return aws.Config{
					Region:     "eu-west-2",
					HTTPClient: captured,
					Credentials: credentials.NewStaticCredentialsProvider(
						"ambient-key",
						"ambient-secret",
						"",
					),
				}, nil
			},
		},
	)
	s.Require().NoError(err)

	_, err = cfg.Credentials.Retrieve(context.Background())
	s.Require().NoError(err)
	s.Contains(
		captured.authorization,
		"ambient-key",
		"AssumeRole must be signed as the ambient principal",
	)
}

func (s *AssumeRoleCredentialsTestSuite) Test_leaves_ambient_credentials_when_no_role_configured() {
	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"region": core.ScalarFromString("eu-west-2"),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		loaderRecording(&recordingHTTPClient{response: stsAssumeRoleResponse}),
	)
	s.Require().NoError(err)
	s.IsType(credentials.StaticCredentialsProvider{}, cfg.Credentials)
}

// Resolved credentials are cached, so a second consumer does not pay another
// AssumeRole call.
func (s *AssumeRoleCredentialsTestSuite) Test_assumed_credentials_are_cached() {
	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"assumeRole.roleArn": core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		loaderRecording(&recordingHTTPClient{response: stsAssumeRoleResponse}),
	)
	s.Require().NoError(err)

	cache, isCache := cfg.Credentials.(*aws.CredentialsCache)
	s.Require().True(isCache, "got %T", cfg.Credentials)
	s.True(cache.IsCredentialsProvider(&stscreds.AssumeRoleProvider{}))
}

func (s *AssumeRoleCredentialsTestSuite) Test_assume_role_sends_an_inline_session_policy() {
	body := s.resolve(map[string]*core.ScalarValue{
		"assumeRole.roleArn": core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
		"assumeRole.policy":  core.ScalarFromString(`{"Version":"2012-10-17"}`),
	}, stsAssumeRoleResponse)

	s.Equal(`{"Version":"2012-10-17"}`, body.Get("Policy"))
}

func (s *AssumeRoleCredentialsTestSuite) Test_assume_role_sends_every_tag_and_policy_arn() {
	body := s.resolve(map[string]*core.ScalarValue{
		"assumeRole.roleArn":      core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
		"assumeRole.tags.team":    core.ScalarFromString("platform"),
		"assumeRole.tags.env":     core.ScalarFromString("prod"),
		"assumeRole.policyArns.0": core.ScalarFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
		"assumeRole.policyArns.1": core.ScalarFromString("arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"),
	}, stsAssumeRoleResponse)

	// Tag order follows map iteration, so assert on the set rather than the index.
	s.ElementsMatch(
		[]string{"team", "env"},
		[]string{body.Get("Tags.member.1.Key"), body.Get("Tags.member.2.Key")},
	)
	s.ElementsMatch(
		[]string{"platform", "prod"},
		[]string{body.Get("Tags.member.1.Value"), body.Get("Tags.member.2.Value")},
	)
	s.Equal(
		"arn:aws:iam::aws:policy/ReadOnlyAccess",
		body.Get("PolicyArns.member.1.arn"),
	)
	s.Equal(
		"arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess",
		body.Get("PolicyArns.member.2.arn"),
	)
}

func (s *AssumeRoleCredentialsTestSuite) Test_web_identity_reads_the_token_from_a_file() {
	tokenFile := filepath.Join(s.T().TempDir(), "token.jwt")
	s.Require().NoError(os.WriteFile(tokenFile, []byte("file-token"), 0600))

	recorder := &recordingHTTPClient{response: stsWebIdentityResponse}
	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"assumeRoleWithWebIdentity.roleArn": core.ScalarFromString(
				"arn:aws:iam::123456789012:role/ci",
			),
			"assumeRoleWithWebIdentity.webIdentityTokenFile": core.ScalarFromString(tokenFile),
			"assumeRoleWithWebIdentity.policy": core.ScalarFromString(
				`{"Version":"2012-10-17"}`,
			),
			"assumeRoleWithWebIdentity.policyArns.0": core.ScalarFromString(
				"arn:aws:iam::aws:policy/ReadOnlyAccess",
			),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		loaderRecording(recorder),
	)
	s.Require().NoError(err)

	_, err = cfg.Credentials.Retrieve(context.Background())
	s.Require().NoError(err)
	s.Equal("file-token", recorder.body.Get("WebIdentityToken"))
	s.Equal(`{"Version":"2012-10-17"}`, recorder.body.Get("Policy"))
	s.Equal(
		"arn:aws:iam::aws:policy/ReadOnlyAccess",
		recorder.body.Get("PolicyArns.member.1.arn"),
	)
}

// An inline token is the more specific of the two, so it wins. Pinned because the
// previous inert implementation resolved the other way round.
func (s *AssumeRoleCredentialsTestSuite) Test_inline_token_takes_precedence_over_a_token_file() {
	tokenFile := filepath.Join(s.T().TempDir(), "token.jwt")
	s.Require().NoError(os.WriteFile(tokenFile, []byte("file-token"), 0600))

	recorder := &recordingHTTPClient{response: stsWebIdentityResponse}
	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"assumeRoleWithWebIdentity.roleArn": core.ScalarFromString(
				"arn:aws:iam::123456789012:role/ci",
			),
			"assumeRoleWithWebIdentity.webIdentityToken":     core.ScalarFromString("inline-token"),
			"assumeRoleWithWebIdentity.webIdentityTokenFile": core.ScalarFromString(tokenFile),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		loaderRecording(recorder),
	)
	s.Require().NoError(err)

	_, err = cfg.Credentials.Retrieve(context.Background())
	s.Require().NoError(err)
	s.Equal("inline-token", recorder.body.Get("WebIdentityToken"))
}

// Role assumption has to honour endpoint.sts like every other service client, or
// it cannot be exercised against a local STS.
func (s *AssumeRoleCredentialsTestSuite) Test_sts_endpoint_override_is_honoured() {
	recorder := &urlCapturingClient{
		inner: &recordingHTTPClient{response: stsAssumeRoleResponse},
	}
	cfg, err := AWSConfigFromProviderContext(
		context.Background(),
		providerCtxWith(map[string]*core.ScalarValue{
			"assumeRole.roleArn": core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
			"endpoint.sts":       core.ScalarFromString("http://localhost:4566"),
		}),
		map[string]*core.MappingNode{},
		map[string]string{},
		&testutils.MockAWSConfigLoader{
			LoadDefaultConfigFunc: func(
				ctx context.Context,
				optFns ...func(*config.LoadOptions) error,
			) (aws.Config, error) {
				return aws.Config{
					Region:     "eu-west-2",
					HTTPClient: recorder,
					Credentials: credentials.NewStaticCredentialsProvider(
						"ambient-key",
						"ambient-secret",
						"",
					),
				}, nil
			},
		},
	)
	s.Require().NoError(err)

	_, err = cfg.Credentials.Retrieve(context.Background())
	s.Require().NoError(err)
	s.Equal("http://localhost:4566", recorder.url)
}

type urlCapturingClient struct {
	inner *recordingHTTPClient
	url   string
}

func (c *urlCapturingClient) Do(req *http.Request) (*http.Response, error) {
	c.url = req.URL.Scheme + "://" + req.URL.Host
	return c.inner.Do(req)
}

func TestAssumeRoleCredentialsTestSuite(t *testing.T) {
	suite.Run(t, new(AssumeRoleCredentialsTestSuite))
}
