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

func (l *lambdaFunctionUrlResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	functionARN, hasFunctionARN := pluginutils.GetValueByPath(
		"$.functionArn",
		input.ResourceSpec,
	)
	if !hasFunctionARN {
		return nil, fmt.Errorf("functionArn must be defined in the resource spec")
	}
	qualifier, _ := pluginutils.GetValueByPath(
		"$.qualifier",
		input.ResourceSpec,
	)

	getFunctionUrlInput := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(core.StringValue(functionARN)),
	}
	if qualifier != nil {
		getFunctionUrlInput.Qualifier = aws.String(core.StringValue(qualifier))
	}

	_, err = lambdaService.GetFunctionUrlConfig(ctx, getFunctionUrlInput)
	if err != nil {
		return &provider.ResourceHasStabilisedOutput{
			Stabilised: false,
		}, nil
	}

	// If we can successfully get the function URL config, it's stabilised
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: true,
	}, nil
}
