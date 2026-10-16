package utils

import (
	"bytes"
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/credentials/stscreds"
	"github.com/aws/aws-sdk-go-v2/feature/ec2/imds"
	"github.com/aws/aws-sdk-go-v2/service/sts"
	"github.com/aws/aws-sdk-go-v2/service/sts/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// AWSConfigLoader defines the interface for loading AWS configurations.
type AWSConfigLoader interface {
	LoadDefaultConfig(
		ctx context.Context,
		optFns ...func(*config.LoadOptions) error,
	) (aws.Config, error)
}

// DefaultAWSConfigLoader implements AWSConfigLoader using the AWS SDK.
type DefaultAWSConfigLoader struct{}

func (l *DefaultAWSConfigLoader) LoadDefaultConfig(
	ctx context.Context,
	optFns ...func(*config.LoadOptions) error,
) (aws.Config, error) {
	return config.LoadDefaultConfig(ctx, optFns...)
}

// AWSConfigFromProviderContext creates an AWS config from the given
// provider context and environment variables.
func AWSConfigFromProviderContext(
	ctx context.Context,
	providerContext provider.Context,
	meta map[string]*core.MappingNode,
	env map[string]string,
	loader AWSConfigLoader,
) (*aws.Config, error) {
	if loader == nil {
		loader = &DefaultAWSConfigLoader{}
	}

	opts := []func(*config.LoadOptions) error{}
	opts = append(opts, RegionOptions(providerContext, meta)...)
	opts = append(opts, RetryConfigOptions(providerContext, env)...)
	opts = append(opts, CredentialOptions(providerContext)...)
	opts = append(opts, SharedEndpointOptions(providerContext)...)
	opts = append(opts, EC2MetadataServiceOptions(providerContext, env)...)

	certOpts, err := CertOptions(providerContext, env)
	if err != nil {
		return nil, err
	}
	opts = append(opts, certOpts...)

	opts = append(opts, HTTPClientOptions(providerContext)...)

	cfg, err := loader.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, err
	}

	applyAssumeRoleCredentials(&cfg, providerContext)

	return &cfg, nil
}

// Replaces the loaded config's credentials with a role assumed through STS when
// provider config names one, and leaves them untouched when it does not.
//
// LoadDefaultConfig cannot do this as it assumes a role only when a shared config
// profile declares one, so a role named in provider config has to be applied
// here.
func applyAssumeRoleCredentials(
	cfg *aws.Config,
	providerContext provider.Context,
) {
	stsClient := newSTSClient(*cfg, providerContext)

	if webIdentity := readWebIdentitySettings(providerContext); webIdentity.isUsable() {
		cfg.Credentials = aws.NewCredentialsCache(
			stscreds.NewWebIdentityRoleProvider(
				stsClient,
				webIdentity.roleARN,
				webIdentity.tokenRetriever,
				webIdentity.apply,
			),
		)
		return
	}

	if assumeRole := readAssumeRoleSettings(providerContext); assumeRole.isUsable() {
		cfg.Credentials = aws.NewCredentialsCache(
			stscreds.NewAssumeRoleProvider(
				stsClient,
				assumeRole.roleARN,
				assumeRole.apply,
			),
		)
	}
}

// Honours endpoint.sts the way every other service client honours its own
// override, so role assumption can be pointed at a local STS.
func newSTSClient(cfg aws.Config, providerContext provider.Context) *sts.Client {
	endpoint, hasEndpoint := GetEndpointFromProviderConfig(
		providerContext,
		"sts",
		Services["sts"],
	)
	if !hasEndpoint || core.IsScalarNil(endpoint) {
		return sts.NewFromConfig(cfg)
	}

	return sts.NewFromConfig(cfg, func(o *sts.Options) {
		o.BaseEndpoint = aws.String(core.StringValueFromScalar(endpoint))
	})
}

// RegionOptions returns the region options derived from the given provider context.
func RegionOptions(
	providerContext provider.Context,
	meta map[string]*core.MappingNode,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	region, hasRegion := providerContext.ProviderConfigVariable("region")
	if hasRegion && !core.IsScalarNil(region) {
		opts = append(opts, config.WithRegion(core.StringValueFromScalar(region)))
	}

	// A nil or empty region in the metadata (e.g. a data source fetch without a region
	// filter) must not override the provider-configured region above.
	if len(meta) > 0 {
		if region, hasRegion := meta["region"]; hasRegion &&
			core.StringValue(region) != "" {
			opts = append(opts, config.WithRegion(core.StringValue(region)))
		}
	}

	return opts
}

// RetryConfigOptions returns the retry config options derived from the
// given provider context and environment variables.
func RetryConfigOptions(
	providerContext provider.Context,
	env map[string]string,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	retryMode, hasRetryMode := getProviderConfigValueFallbackToEnv(
		providerContext,
		env,
		"retryMode",
		"AWS_RETRY_MODE",
	)
	if hasRetryMode && !core.IsScalarNil(retryMode) {
		retryModeValue := core.StringValueFromScalar(retryMode)
		opts = append(opts, config.WithRetryMode(aws.RetryMode(retryModeValue)))
	}

	maxRetries, hasMaxRetries := providerContext.ProviderConfigVariable("maxRetries")
	if hasMaxRetries && !core.IsScalarNil(maxRetries) {
		maxRetriesValue := core.IntValueFromScalar(maxRetries)
		opts = append(opts, config.WithRetryMaxAttempts(maxRetriesValue))
	}

	return opts
}

// CredentialOptions returns the credential options derived from the given provider context.
func CredentialOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	accessKeyID, hasAccessKeyID := providerContext.ProviderConfigVariable(
		"accessKeyId",
	)
	secretAccessKey, hasSecretAccessKey := providerContext.ProviderConfigVariable(
		"secretAccessKey",
	)
	sessionToken, _ := providerContext.ProviderConfigVariable(
		"sessionToken",
	)

	if hasAccessKeyID && !core.IsScalarNil(accessKeyID) &&
		hasSecretAccessKey && !core.IsScalarNil(secretAccessKey) {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				core.StringValueFromScalar(accessKeyID),
				core.StringValueFromScalar(secretAccessKey),
				core.StringValueFromScalar(sessionToken),
			),
		))
	}

	sharedCredentialsFiles, hasSharedCredentialsFiles := providerContext.ProviderConfigVariable(
		"sharedCredentialsFiles",
	)
	if hasSharedCredentialsFiles && !core.IsScalarNil(sharedCredentialsFiles) {
		credentialsFilesStr := core.StringValueFromScalar(sharedCredentialsFiles)
		opts = append(opts, config.WithSharedCredentialsFiles(
			sharedFilePaths(credentialsFilesStr),
		))
	}

	sharedConfigFiles, hasSharedConfigFiles := providerContext.ProviderConfigVariable(
		"sharedConfigFiles",
	)
	if hasSharedConfigFiles && !core.IsScalarNil(sharedConfigFiles) {
		configFilesStr := core.StringValueFromScalar(sharedConfigFiles)
		opts = append(opts, config.WithSharedConfigFiles(
			sharedFilePaths(configFilesStr),
		))
	}

	profile, hasProfile := providerContext.ProviderConfigVariable(
		"profile",
	)
	if hasProfile && !core.IsScalarNil(profile) {
		opts = append(opts, config.WithSharedConfigProfile(
			core.StringValueFromScalar(profile),
		))
	}

	return opts
}

// Splits a comma-separated list of shared config or credentials file paths and
// resolves each one.
func sharedFilePaths(value string) []string {
	parts := strings.Split(value, ",")
	paths := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		paths = append(paths, resolveSharedFilePath(trimmed))
	}

	return paths
}

// Resolves the references in a shared config or credentials file path.
func resolveSharedFilePath(path string) string {
	return expandHomePath(expandEnvRefs(path))
}

// envRefPattern matches an environment variable reference in any of the forms a path is
// written with: Windows "%NAME%", and "${NAME}" or "$NAME" elsewhere. The braced form is
// listed before the bare one so it wins where both could match.
//
// A name may not hold a path separator, so a stray "%" or "$" elsewhere in a path cannot
// start a reference; the Windows form also allows parentheses, for the likes of
// %ProgramFiles(x86)%.
var envRefPattern = regexp.MustCompile(
	`%[A-Za-z_][A-Za-z0-9_()]*%|\$\{[A-Za-z_][A-Za-z0-9_]*\}|\$[A-Za-z_][A-Za-z0-9_]*`,
)

// Expands environment variable references in a path.
//
// A value from provider configuration never passes through a shell, and neither the AWS
// SDK nor the file APIs of any platform expand environment references themselves, so
// "%USERPROFILE%\.aws\credentials" or "$HOME/.aws/credentials" would otherwise reach the
// filesystem as-is and be reported as a missing profile rather than as a path that was
// never resolved.
//
// Every form is expanded on every platform, so one configuration value resolves the same
// way wherever the plugin runs. A reference to a variable that is not set is left exactly
// as written, otherwise, collapsing it to nothing would turn a typo into a different path
// that looks plausible.
func expandEnvRefs(path string) string {
	return envRefPattern.ReplaceAllStringFunc(path, func(reference string) string {
		if value, found := os.LookupEnv(envRefName(reference)); found {
			return value
		}
		return reference
	})
}

// The variable name inside a reference, in whichever form it was written.
func envRefName(reference string) string {
	switch {
	case strings.HasPrefix(reference, "%"):
		return reference[1 : len(reference)-1]
	case strings.HasPrefix(reference, "${"):
		return reference[2 : len(reference)-1]
	default:
		return reference[1:]
	}
}

// Resolves a leading "~" to the current user's home directory.
//
// The AWS SDK takes these paths literally, so a "~" reaches the filesystem
// unexpanded and resolves against the working directory instead of the home
// directory. The file is then missing and the named profile cannot be found,
// which surfaces as the profile not existing rather than as a bad path. The
// provider's own defaults are written with a "~", so this affects every
// deployment that relies on them.
func expandHomePath(path string) string {
	rest, isHomeRelative := homeRelativePath(path)
	if !isHomeRelative {
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		// Nothing better to fall back on; the SDK will report the missing file.
		return path
	}

	if rest == "" {
		return home
	}

	return filepath.Join(home, rest)
}

// Splits a leading "~" off a path, reporting whether there was one.
//
// The separator is whatever the platform accepts: os.IsPathSeparator takes a
// backslash on Windows, so "~\.aws\config" is treated as a home reference there,
// matching what os.path.expanduser accepts and so what the AWS CLI accepts. On
// every other platform a backslash is an ordinary character in a file name and is
// left alone.
func homeRelativePath(path string) (string, bool) {
	if path == "~" {
		return "", true
	}

	if len(path) > 1 && path[0] == '~' && os.IsPathSeparator(path[1]) {
		return path[2:], true
	}

	return "", false
}

// SharedEndpointOptions returns the shared endpoint options derived from the given provider context.
func SharedEndpointOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	useFIPSEndpoint, hasUseFIPSEndpoint := providerContext.ProviderConfigVariable(
		"useFIPSEndpoint",
	)
	if hasUseFIPSEndpoint && !core.IsScalarNil(useFIPSEndpoint) {
		useFIPSEndpointValue := core.BoolValueFromScalar(useFIPSEndpoint)
		fipsEndpointState := aws.FIPSEndpointStateDisabled
		if useFIPSEndpointValue {
			fipsEndpointState = aws.FIPSEndpointStateEnabled
		}
		opts = append(opts, config.WithUseFIPSEndpoint(fipsEndpointState))
	}

	useDualStackEndpoint, hasUseDualStackEndpoint := providerContext.ProviderConfigVariable(
		"useDualStackEndpoint",
	)
	if hasUseDualStackEndpoint && !core.IsScalarNil(useDualStackEndpoint) {
		useDualStackEndpointValue := core.BoolValueFromScalar(useDualStackEndpoint)
		dualStackEndpointState := aws.DualStackEndpointStateDisabled
		if useDualStackEndpointValue {
			dualStackEndpointState = aws.DualStackEndpointStateEnabled
		}
		opts = append(opts, config.WithUseDualStackEndpoint(dualStackEndpointState))
	}

	return opts
}

// CertOptions returns the cert options derived from the given provider context and environment variables.
func CertOptions(
	providerContext provider.Context,
	env map[string]string,
) ([]func(*config.LoadOptions) error, error) {
	opts := []func(*config.LoadOptions) error{}

	customCABundle, hasCustomCABundle := getProviderConfigValueFallbackToEnv(
		providerContext,
		env,
		"customCABundle",
		"AWS_CA_BUNDLE",
	)
	if hasCustomCABundle && !core.IsScalarNil(customCABundle) {
		path := core.StringValueFromScalar(customCABundle)
		bundleData, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		opts = append(opts, config.WithCustomCABundle(
			bytes.NewReader(bundleData),
		))
	}

	return opts, nil
}

// HTTPClientOptions returns the http client options derived from the given provider context.
func HTTPClientOptions(
	providerContext provider.Context,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	insecureScalarValue, hasInsecure := providerContext.ProviderConfigVariable(
		"insecure",
	)
	insecure := core.BoolValueFromScalar(insecureScalarValue)

	httpProxyScalarValue, hasHTTPProxy := providerContext.ProviderConfigVariable(
		"httpProxy",
	)
	httpProxy := core.StringValueFromScalar(httpProxyScalarValue)

	httpsProxyScalarValue, hasHTTPSProxy := providerContext.ProviderConfigVariable(
		"httpsProxy",
	)
	httpsProxy := core.StringValueFromScalar(httpsProxyScalarValue)

	// The AWS SDK will automatically pick up the HTTP_PROXY and HTTPS_PROXY
	// environment variables, we only need to configure http proxies if defined
	// in the provider config.
	if (hasInsecure && insecure) ||
		(hasHTTPProxy && !core.IsScalarNil(httpProxyScalarValue)) ||
		(hasHTTPSProxy && !core.IsScalarNil(httpsProxyScalarValue)) {
		customClient := awshttp.NewBuildableClient().WithTransportOptions(
			func(t *http.Transport) {
				if insecure {
					t.TLSClientConfig = &tls.Config{
						InsecureSkipVerify: true,
					}
				}

				finalProxyURL := strings.TrimSpace(httpProxy)
				if finalProxyURL == "" {
					finalProxyURL = strings.TrimSpace(httpsProxy)
				}
				if finalProxyURL != "" {
					parsedProxyURL, err := url.Parse(finalProxyURL)
					if err != nil {
						log.Fatalf("Failed to parse proxy URL: %v", err)
					}

					t.Proxy = http.ProxyURL(parsedProxyURL)
				}
			},
		)
		opts = append(opts, config.WithHTTPClient(customClient))
	}

	return opts
}

// EC2MetadataServiceOptions returns the ec2 metadata service options
// derived from the given provider context and environment variables.
func EC2MetadataServiceOptions(
	providerContext provider.Context,
	env map[string]string,
) []func(*config.LoadOptions) error {
	opts := []func(*config.LoadOptions) error{}

	ec2MetadataServiceEndpoint, hasEC2MetadataServiceEndpoint := getProviderConfigValueFallbackToEnv(
		providerContext,
		env,
		"ec2MetadataServiceEndpoint",
		"AWS_EC2_METADATA_SERVICE_ENDPOINT",
	)
	ec2MetadataServiceEndpointMode, hasEC2MetadataServiceEndpointMode := getProviderConfigValueFallbackToEnv(
		providerContext,
		env,
		"ec2MetadataServiceEndpointMode",
		"AWS_EC2_METADATA_SERVICE_ENDPOINT_MODE",
	)

	if hasEC2MetadataServiceEndpoint && !core.IsScalarNil(ec2MetadataServiceEndpoint) {
		opts = append(
			opts,
			config.WithEC2IMDSEndpoint(
				core.StringValueFromScalar(ec2MetadataServiceEndpoint),
			),
		)
	}

	if hasEC2MetadataServiceEndpointMode && !core.IsScalarNil(ec2MetadataServiceEndpointMode) {
		opts = append(
			opts,
			config.WithEC2IMDSEndpointMode(
				imds.EndpointModeState(
					imdsEndpointModeStateFromString(
						core.StringValueFromScalar(ec2MetadataServiceEndpointMode),
					),
				),
			),
		)
	}

	return opts
}

func imdsEndpointModeStateFromString(s string) imds.EndpointModeState {
	switch strings.ToLower(s) {
	case "ipv4":
		return imds.EndpointModeStateIPv4
	case "ipv6":
		return imds.EndpointModeStateIPv6
	default:
		return imds.EndpointModeStateUnset
	}
}

// Reads the assumeRole.* provider config fields. A role ARN of "" means no role
// is configured.
func readAssumeRoleSettings(providerContext provider.Context) assumeRoleSettings {
	// Wrapped in plugin config for the prefix helpers that emulate arrays and maps.
	pluginConfig := core.PluginConfig(providerContext.ProviderConfigVariables())

	return assumeRoleSettings{
		roleARN:           readConfigString(providerContext, "assumeRole.roleArn"),
		externalID:        readConfigString(providerContext, "assumeRole.externalId"),
		policy:            readConfigString(providerContext, "assumeRole.policy"),
		sessionName:       readConfigString(providerContext, "assumeRole.sessionName"),
		sourceIdentity:    readConfigString(providerContext, "assumeRole.sourceIdentity"),
		duration:          readConfigDuration(providerContext, "assumeRole.duration"),
		policyARNs:        pluginConfig.SliceFromPrefix("assumeRole.policyArns"),
		tags:              pluginConfig.MapFromPrefix("assumeRole.tags"),
		transitiveTagKeys: readConfigString(providerContext, "assumeRole.transitiveTagKeys"),
	}
}

type assumeRoleSettings struct {
	roleARN           string
	externalID        string
	policy            string
	sessionName       string
	sourceIdentity    string
	duration          time.Duration
	policyARNs        []*core.ScalarValue
	tags              map[string]*core.ScalarValue
	transitiveTagKeys string
}

func (s assumeRoleSettings) isUsable() bool {
	return s.roleARN != ""
}

func (s assumeRoleSettings) apply(o *stscreds.AssumeRoleOptions) {
	o.RoleARN = s.roleARN

	if s.externalID != "" {
		o.ExternalID = aws.String(s.externalID)
	}
	if s.duration > 0 {
		o.Duration = s.duration
	}
	if s.policy != "" {
		o.Policy = aws.String(s.policy)
	}
	if len(s.policyARNs) > 0 {
		o.PolicyARNs = toSTSPolicyARNs(s.policyARNs)
	}
	if s.sessionName != "" {
		o.RoleSessionName = s.sessionName
	}
	if s.sourceIdentity != "" {
		o.SourceIdentity = aws.String(s.sourceIdentity)
	}
	if len(s.tags) > 0 {
		o.Tags = toSTSTags(s.tags)
	}
	if s.transitiveTagKeys != "" {
		o.TransitiveTagKeys = strings.Split(s.transitiveTagKeys, ",")
	}
}

// Reads the assumeRoleWithWebIdentity.* provider config fields. The token
// retriever is held here rather than set through the option function because
// NewWebIdentityRoleProvider takes it positionally, and because whether one could
// be built is what decides if web identity is usable at all.
func readWebIdentitySettings(providerContext provider.Context) webIdentitySettings {
	pluginConfig := core.PluginConfig(providerContext.ProviderConfigVariables())

	return webIdentitySettings{
		roleARN:        readConfigString(providerContext, "assumeRoleWithWebIdentity.roleArn"),
		policy:         readConfigString(providerContext, "assumeRoleWithWebIdentity.policy"),
		sessionName:    readConfigString(providerContext, "assumeRoleWithWebIdentity.sessionName"),
		duration:       readConfigDuration(providerContext, "assumeRoleWithWebIdentity.duration"),
		policyARNs:     pluginConfig.SliceFromPrefix("assumeRoleWithWebIdentity.policyArns"),
		tokenRetriever: webIdentityTokenRetriever(providerContext),
	}
}

type webIdentitySettings struct {
	roleARN        string
	policy         string
	sessionName    string
	duration       time.Duration
	policyARNs     []*core.ScalarValue
	tokenRetriever stscreds.IdentityTokenRetriever
}

// stscreds calls GetIdentityToken without a nil check, so a role with no
// retriever panics inside the SDK rather than erroring.
func (s webIdentitySettings) isUsable() bool {
	return s.roleARN != "" && s.tokenRetriever != nil
}

func (s webIdentitySettings) apply(o *stscreds.WebIdentityRoleOptions) {
	o.RoleARN = s.roleARN

	if s.duration > 0 {
		o.Duration = s.duration
	}

	if s.policy != "" {
		o.Policy = aws.String(s.policy)
	}

	if len(s.policyARNs) > 0 {
		o.PolicyARNs = toSTSPolicyARNs(s.policyARNs)
	}

	if s.sessionName != "" {
		o.RoleSessionName = s.sessionName
	}
}

// An inline token takes precedence over a token file, matching the order the
// fields are documented in.
func webIdentityTokenRetriever(
	providerContext provider.Context,
) stscreds.IdentityTokenRetriever {
	if token := readConfigString(providerContext, "assumeRoleWithWebIdentity.webIdentityToken"); token != "" {
		return staticTokenRetriever(token)
	}

	if file := readConfigString(providerContext, "assumeRoleWithWebIdentity.webIdentityTokenFile"); file != "" {
		return stscreds.IdentityTokenFile(file)
	}

	return nil
}

func readConfigString(providerContext provider.Context, key string) string {
	value, has := providerContext.ProviderConfigVariable(key)
	if !has || core.IsScalarNil(value) {
		return ""
	}
	return core.StringValueFromScalar(value)
}

// A parse failure resolves to no duration, leaving the SDK default. Validation
// rejects an unparseable duration on the engine's path, but this is also reached
// by harnesses that construct a config directly and never validate.
func readConfigDuration(providerContext provider.Context, key string) time.Duration {
	raw := readConfigString(providerContext, key)
	if raw == "" {
		return 0
	}

	duration, err := time.ParseDuration(raw)
	if err != nil {
		return 0
	}
	return duration
}

func getProviderConfigValueFallbackToEnv(
	providerContext provider.Context,
	env map[string]string,
	key string,
	envKey string,
) (*core.ScalarValue, bool) {
	providerConfigValue, hasProviderConfigValue := providerContext.ProviderConfigVariable(key)
	if hasProviderConfigValue && !core.IsScalarNil(providerConfigValue) {
		return providerConfigValue, true
	}

	envValue, hasEnvValue := env[envKey]
	if hasEnvValue {
		return core.ScalarFromString(envValue), true
	}

	return nil, false
}

func toSTSTags(tagValues map[string]*core.ScalarValue) []types.Tag {
	tags := make([]types.Tag, 0, len(tagValues))

	for key := range tagValues {
		tags = append(tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(core.StringValueFromScalar(tagValues[key])),
		})
	}

	return tags
}

func toSTSPolicyARNs(policyARNConfigValues []*core.ScalarValue) []types.PolicyDescriptorType {
	policyARNs := make([]types.PolicyDescriptorType, 0, len(policyARNConfigValues))

	for _, policyARN := range policyARNConfigValues {
		policyARNs = append(policyARNs, types.PolicyDescriptorType{
			Arn: aws.String(core.StringValueFromScalar(policyARN)),
		})
	}

	return policyARNs
}

type staticTokenRetriever string

func (s staticTokenRetriever) GetIdentityToken() ([]byte, error) {
	return []byte(s), nil
}

// AWSConfigCacheKey creates a cache key for the given session ID and metadata
// specific to a specific request for an action on a resource, data source, or link.
func AWSConfigCacheKey(
	sessionID string,
	meta map[string]*core.MappingNode,
) string {
	if len(meta) == 0 {
		return sessionID
	}

	if region, hasRegion := meta["region"]; hasRegion {
		return fmt.Sprintf("%s--region-%s", sessionID, core.StringValue(region))
	}

	return sessionID
}
