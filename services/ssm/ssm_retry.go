package ssm

import (
	"errors"

	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// The SSM API error codes that represent transient throttling rather than a rejected
// request.
//
// Parameter Store meters PutParameter against an account-wide throughput limit that is low
// by default, and a resource writing a tree of parameters spends that budget in one burst.
// The SDK's own retries are bounded in attempts and finish within seconds, which is shorter
// than the window a saturated limit stays saturated for.
var retryableSSMErrorCodes = map[string]bool{
	"ThrottlingException":  true,
	"TooManyUpdates":       true,
	"RequestLimitExceeded": true,
}

// Marks a throttled SSM operation as retryable, leaving every other error
// to fail the deployment as before.
//
// A plain error for a throttle abandons the whole resource for a condition that clears on
// its own; wrapping it as a provider.RetryableError hands it to the provider's retry
// policy, which backs off over a far longer window than the SDK does.
//
// Safe for the writes in a parameter tree or path as they are individually idempotent and the
// create path already tolerates finding a parameter it wrote on an earlier attempt, so a
// retried operation converges rather than conflicting.
func retryableSSMError(err error) error {
	if err == nil {
		return nil
	}

	if isSSMErrorRetryable(err) {
		return &provider.RetryableError{ChildError: err}
	}

	return err
}

// Classifies an SSM API error as retryable, matching on the smithy API error code
// so it is decoupled from concrete generated type names.
func isSSMErrorRetryable(err error) bool {
	if err == nil {
		return false
	}

	if apiErr, ok := errors.AsType[smithy.APIError](err); ok {
		return retryableSSMErrorCodes[apiErr.ErrorCode()]
	}

	return false
}
