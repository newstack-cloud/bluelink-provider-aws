package overlays

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
)

const lambdaFunctionType = "aws/lambda/function"

func init() {
	RegisterBehaviour(lambdaFunctionType, &Behaviour{
		SettledAfterLinkUpdate: lambdaFunctionSettledAfterLinkUpdate,
	})
	Register(lambdaFunctionType, lambdaFunctionOverlay)
}

// Declares the VPC attachment lists as unordered.
//
// Lambda treats a function's subnets and security groups as sets and returns them in an
// order of its own choosing, so comparing them positionally reports drift on a function
// nobody has touched. The claim is about the service's semantics so it belongs here rather
// than in the comparison, which cannot know which arrays are sets.
func lambdaFunctionOverlay(
	schema *provider.ResourceDefinitionsSchema,
) *provider.ResourceDefinitionsSchema {
	vpcConfig, ok := schema.Attributes["vpcConfig"]
	if !ok || vpcConfig.Attributes == nil {
		return schema
	}

	for _, field := range []string{"subnetIds", "securityGroupIds"} {
		if attr, has := vpcConfig.Attributes[field]; has {
			attr.IgnoreItemOrder = true
		}
	}

	return schema
}

// Reports whether a Lambda function has finished applying a configuration change a link
// made.
//
// UpdateFunctionConfiguration returns once the change is accepted, while the function
// stays in LastUpdateStatus InProgress and rejects any further configuration update with
// ResourceConflictException. Without this, the next link to write the function reaches it
// at exactly the moment the API will refuse.
func lambdaFunctionSettledAfterLinkUpdate(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
	awsConfig *aws.Config,
) (bool, error) {
	functionID := lambdaFunctionIdentifier(input)
	if functionID == "" {
		// Nothing to look up, so nothing to wait for. Reporting settled leaves the
		// resource locks to keep links apart, which is where this started.
		return true, nil
	}

	service := lambdaservice.NewService(awsConfig, input.ProviderContext)
	output, err := service.GetFunction(ctx, &lambda.GetFunctionInput{
		FunctionName: aws.String(functionID),
	})
	if err != nil {
		return false, err
	}

	var config *lambdatypes.FunctionConfiguration
	if output != nil {
		config = output.Configuration
	}

	// A function that has never been updated reports no status at all, which is settled
	// rather than unknown.
	if config == nil || config.LastUpdateStatus == "" ||
		config.LastUpdateStatus == lambdatypes.LastUpdateStatusSuccessful {
		return true, nil
	}

	if config.LastUpdateStatus == lambdatypes.LastUpdateStatusFailed {
		reason := aws.ToString(config.LastUpdateStatusReason)
		err := fmt.Errorf(
			"update to lambda function %q failed after it was accepted: %s",
			functionID,
			reason,
		)
		if isTransientUpdateFailureReason(reason) {
			// Surfaced as retryable so the engine re-runs the idempotent link update
			// rather than failing the deployment over a permission that does exist.
			return false, &provider.RetryableError{ChildError: err}
		}
		return false, err
	}

	return false, nil
}

func lambdaFunctionIdentifier(input *provider.ResourceHasStabilisedInput) string {
	if input.ResourceSpec == nil {
		return ""
	}

	for _, field := range []string{"arn", "functionName"} {
		if value, ok := input.ResourceSpec.Fields[field]; ok {
			if identifier := core.StringValue(value); identifier != "" {
				return identifier
			}
		}
	}

	return ""
}

// Reports whether an update failed for a reason that clears on its own.
//
// Attaching a function to a VPC makes Lambda check, at that moment, that the execution role
// can manage network interfaces. The link grants that permission immediately beforehand and
// IAM is eventually consistent, so the check can fail for a permission that does exist and
// will succeed moments later.
//
// Matched on the status reason rather than an API error because the failure happens after
// UpdateFunctionConfiguration has already returned successfully: there is no error to
// inspect, only a status the function reports later.
func isTransientUpdateFailureReason(reason string) bool {
	lowered := strings.ToLower(reason)
	for _, fragment := range []string{
		"execution role doesn't have permission",
		"execution role does not have permission",
		"does not have permissions to call",
		"cannot access",
		"ensure the role can perform",
	} {
		if strings.Contains(lowered, fragment) {
			return true
		}
	}
	return false
}
