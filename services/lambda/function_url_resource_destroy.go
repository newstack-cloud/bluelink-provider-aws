package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionUrlResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return fmt.Errorf("failed to get Lambda service: %w", err)
	}

	functionARNValue, hasFunctionARN := pluginutils.GetValueByPath(
		"$.functionArn",
		input.ResourceState.SpecData,
	)
	if !hasFunctionARN {
		return fmt.Errorf("failed to get functionArn from spec: %w", err)
	}

	qualifier, hasQualifier := pluginutils.GetValueByPath(
		"$.qualifier",
		input.ResourceState.SpecData,
	)

	deleteFunctionUrlInput := &lambda.DeleteFunctionUrlConfigInput{
		FunctionName: aws.String(core.StringValue(functionARNValue)),
	}
	if hasQualifier {
		deleteFunctionUrlInput.Qualifier = aws.String(core.StringValue(qualifier))
	}

	_, err = lambdaService.DeleteFunctionUrlConfig(ctx, deleteFunctionUrlInput)
	if err != nil {
		return fmt.Errorf("failed to delete function URL: %w", err)
	}

	return nil
}
