package provider

import (
	"fmt"
	"regexp"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/validation"
)

func validateAssumeRoleDuration(
	key string,
	value *core.ScalarValue,
	pluginConfig core.PluginConfig,
) []*core.Diagnostic {
	stringVal := core.StringValueFromScalar(value)
	duration, err := time.ParseDuration(stringVal)
	if err != nil {
		return []*core.Diagnostic{
			{
				Level: core.DiagnosticLevelError,
				Message: fmt.Sprintf(
					"Invalid duration %q for field %q: %s",
					stringVal, key, err.Error(),
				),
			},
		}
	}

	if duration.Minutes() < 15 || duration.Hours() > 12 {
		return []*core.Diagnostic{
			{
				Level: core.DiagnosticLevelError,
				Message: fmt.Sprintf(
					"Duration %q for field %q must be between 15 minutes and 12 hours",
					stringVal, key,
				),
			},
		}
	}

	return []*core.Diagnostic{}
}

var partitionRegexp = regexp.MustCompile(`^aws(-[a-z]+)*$`)
var regionRegexp = regexp.MustCompile(`^[a-z]{2}(-[a-z]+)+-\d{1,2}$`)
var accountIDRegexp = regexp.MustCompile(
	`^(aws|aws-managed|third-party|aws-marketplace|\d{12}|cw.{10})$`,
)

func validateARN(
	key string,
	value *core.ScalarValue,
	pluginConfig core.PluginConfig,
) []*core.Diagnostic {
	diagnostics := []*core.Diagnostic{}

	stringVal := core.StringValueFromScalar(value)
	parsedARN, err := arn.Parse(stringVal)
	if err != nil {
		return []*core.Diagnostic{
			{
				Level: core.DiagnosticLevelError,
				Message: fmt.Sprintf(
					"Invalid ARN %q for field %q: %s",
					stringVal, key, err.Error(),
				),
			},
		}
	}

	if parsedARN.Partition == "" {
		diagnostics = append(diagnostics, &core.Diagnostic{
			Level: core.DiagnosticLevelError,
			Message: fmt.Sprintf(
				"ARN %q for field %q is missing a partition value",
				stringVal,
				key,
			),
		})
	} else if !partitionRegexp.MatchString(parsedARN.Partition) {
		diagnostics = append(diagnostics, &core.Diagnostic{
			Level: core.DiagnosticLevelError,
			Message: fmt.Sprintf(
				"ARN %q for field %q has an invalid partition value: %s",
				stringVal, key, parsedARN.Partition,
			),
		})
	}

	if parsedARN.Region != "" && !regionRegexp.MatchString(parsedARN.Region) {
		diagnostics = append(diagnostics, &core.Diagnostic{
			Level: core.DiagnosticLevelError,
			Message: fmt.Sprintf(
				"ARN %q for field %q has an invalid region value: %s",
				stringVal, key, parsedARN.Region,
			),
		})
	}

	if parsedARN.AccountID != "" &&
		!accountIDRegexp.MatchString(parsedARN.AccountID) {
		diagnostics = append(diagnostics, &core.Diagnostic{
			Level: core.DiagnosticLevelError,
			Message: fmt.Sprintf(
				"ARN %q for field %q has an invalid account ID: %s",
				stringVal, key, parsedARN.AccountID,
			),
		})
	}

	if parsedARN.Resource == "" {
		diagnostics = append(diagnostics, &core.Diagnostic{
			Level: core.DiagnosticLevelError,
			Message: fmt.Sprintf(
				"ARN %q for field %q is missing a resource value",
				stringVal, key,
			),
		})
	}

	return diagnostics
}

var validateAssumeRoleSessionName = validation.WrapForPluginConfig(
	validation.AllOf(
		validation.StringLengthRange(2, 64),
		validation.StringMatchesPattern(
			regexp.MustCompile(`[\w+=,.@\-]*`),
		),
	),
)

var validateAssumeRoleSourceIdentity = validation.WrapForPluginConfig(
	validation.AllOf(
		validation.StringLengthRange(2, 64),
		validation.StringMatchesPattern(
			regexp.MustCompile(`[\w+=,.@\-]*`),
		),
	),
)

const (
	webIdentityTokenKey     = "assumeRoleWithWebIdentity.webIdentityToken"
	webIdentityTokenFileKey = "assumeRoleWithWebIdentity.webIdentityTokenFile"
)

// A web identity role cannot be assumed without a token, and the token fields are
// independently optional so nothing else pairs them with the role ARN.
//
// This is a warning rather than an error, because the token may come from the
// environment. AWS_WEB_IDENTITY_TOKEN_FILE is how IRSA and the EKS pod identity
// webhook supply it, and a ValidateFunc cannot see the environment. The runtime
// defers to whatever the SDK resolved in that case, so an error here would reject
// a configuration that works.
func validateWebIdentityRoleARN(
	key string,
	value *core.ScalarValue,
	pluginConfig core.PluginConfig,
) []*core.Diagnostic {
	diagnostics := validateARN(key, value, pluginConfig)

	if hasWebIdentityToken(pluginConfig) {
		return diagnostics
	}

	return append(diagnostics, &core.Diagnostic{
		Level: core.DiagnosticLevelWarning,
		Message: fmt.Sprintf(
			"%q is set but neither %q nor %q is. The role is only assumed if a "+
				"token is supplied, either in configuration or through the "+
				"AWS_WEB_IDENTITY_TOKEN_FILE environment variable.",
			key,
			webIdentityTokenKey,
			webIdentityTokenFileKey,
		),
	})
}

func hasWebIdentityToken(pluginConfig core.PluginConfig) bool {
	for _, key := range []string{webIdentityTokenKey, webIdentityTokenFileKey} {
		if hasNonEmptyString(pluginConfig, key) {
			return true
		}
	}

	return false
}

func hasNonEmptyString(pluginConfig core.PluginConfig, key string) bool {
	value, ok := pluginConfig.Get(key)
	if !ok || core.IsScalarNil(value) {
		return false
	}

	return core.StringValueFromScalar(value) != ""
}
