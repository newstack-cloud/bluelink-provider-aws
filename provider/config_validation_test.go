//go:build unit

package provider

import (
	"testing"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/stretchr/testify/suite"
)

type ConfigValidationTestSuite struct {
	suite.Suite
}

const testWebIdentityRoleARN = "arn:aws:iam::123456789012:role/ci"

// Through the real validator against the real definition, so the ValidateFunc
// being wired onto the field is part of what is asserted. Reverting that wiring
// is otherwise invisible.
func (s *ConfigValidationTestSuite) validate(
	config map[string]*core.ScalarValue,
) []*core.Diagnostic {
	diagnostics, err := core.ValidateConfigDefinition(
		"aws",
		"provider",
		config,
		providerConfigDefinition(),
	)
	s.Require().NoError(err)

	return diagnostics
}

func (s *ConfigValidationTestSuite) levels(
	config map[string]*core.ScalarValue,
) (errors, warnings []string) {
	for _, diagnostic := range s.validate(config) {
		switch diagnostic.Level {
		case core.DiagnosticLevelError:
			errors = append(errors, diagnostic.Message)
		case core.DiagnosticLevelWarning:
			warnings = append(warnings, diagnostic.Message)
		}
	}

	return errors, warnings
}

// The token can come from AWS_WEB_IDENTITY_TOKEN_FILE, which a ValidateFunc
// cannot see, so a role without a configured token must not be fatal. This is the
// shape IRSA and the EKS pod identity webhook produce.
func (s *ConfigValidationTestSuite) Test_web_identity_role_without_a_token_warns() {
	errors, warnings := s.levels(map[string]*core.ScalarValue{
		"assumeRoleWithWebIdentity.roleArn": core.ScalarFromString(testWebIdentityRoleARN),
	})

	s.Empty(errors, "an environment-supplied token is a working configuration")
	s.Require().Len(warnings, 1)
	s.Contains(warnings[0], "AWS_WEB_IDENTITY_TOKEN_FILE")
}

func (s *ConfigValidationTestSuite) Test_web_identity_role_with_a_token_is_clean() {
	for _, tokenKey := range []string{
		"assumeRoleWithWebIdentity.webIdentityToken",
		"assumeRoleWithWebIdentity.webIdentityTokenFile",
	} {
		s.Run(tokenKey, func() {
			errors, warnings := s.levels(map[string]*core.ScalarValue{
				"assumeRoleWithWebIdentity.roleArn": core.ScalarFromString(
					testWebIdentityRoleARN,
				),
				tokenKey: core.ScalarFromString("a-token"),
			})

			s.Empty(errors)
			s.Empty(warnings)
		})
	}
}

// An empty token is not a token, so it must not satisfy the pairing.
func (s *ConfigValidationTestSuite) Test_empty_token_does_not_satisfy_the_pairing() {
	_, warnings := s.levels(map[string]*core.ScalarValue{
		"assumeRoleWithWebIdentity.roleArn":          core.ScalarFromString(testWebIdentityRoleARN),
		"assumeRoleWithWebIdentity.webIdentityToken": core.ScalarFromString(""),
	})

	s.Len(warnings, 1, "an empty token should still warn")
}

// The ARN validation the field carried before the cross-field check was added has
// to survive it.
func (s *ConfigValidationTestSuite) Test_web_identity_role_arn_is_still_validated() {
	errors, _ := s.levels(map[string]*core.ScalarValue{
		"assumeRoleWithWebIdentity.roleArn":          core.ScalarFromString("not-an-arn"),
		"assumeRoleWithWebIdentity.webIdentityToken": core.ScalarFromString("a-token"),
	})

	s.Require().Len(errors, 1)
	s.Contains(errors[0], "Invalid ARN")
}

func (s *ConfigValidationTestSuite) Test_plain_assume_role_is_unaffected_by_the_pairing() {
	errors, warnings := s.levels(map[string]*core.ScalarValue{
		"assumeRole.roleArn": core.ScalarFromString("arn:aws:iam::123456789012:role/deploy"),
	})

	s.Empty(errors)
	s.Empty(warnings, "the pairing applies to web identity only")
}

func (s *ConfigValidationTestSuite) Test_duration_is_validated_on_both_assume_role_paths() {
	for _, key := range []string{
		"assumeRole.duration",
		"assumeRoleWithWebIdentity.duration",
	} {
		s.Run(key, func() {
			errors, _ := s.levels(map[string]*core.ScalarValue{
				key: core.ScalarFromString("not-a-duration"),
			})
			s.NotEmpty(errors, "%q must reject an unparseable duration", key)
		})
	}
}

func TestConfigValidationTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigValidationTestSuite))
}
