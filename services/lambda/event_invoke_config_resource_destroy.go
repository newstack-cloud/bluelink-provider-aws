package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaEventInvokeConfigResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return fmt.Errorf("failed to get Lambda service: %w", err)
	}

	functionName, hasFunctionName := pluginutils.GetValueByPath(
		"$.functionName",
		input.ResourceState.SpecData,
	)
	if !hasFunctionName {
		return fmt.Errorf("failed to get functionName from spec")
	}

	qualifier, hasQualifier := pluginutils.GetValueByPath(
		"$.qualifier",
		input.ResourceState.SpecData,
	)
	if !hasQualifier {
		return fmt.Errorf("failed to get qualifier from spec")
	}

	deleteEventInvokeConfigInput := &lambda.DeleteFunctionEventInvokeConfigInput{
		FunctionName: aws.String(core.StringValue(functionName)),
		Qualifier:    aws.String(core.StringValue(qualifier)),
	}

	_, err = lambdaService.DeleteFunctionEventInvokeConfig(ctx, deleteEventInvokeConfigInput)
	if err != nil {
		return fmt.Errorf("failed to delete event invoke config: %w", err)
	}

	return nil
}
