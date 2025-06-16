package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaFunctionVersionResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	functionARN := core.StringValue(
		input.ResourceSpec.Fields["functionArn"],
	)
	version := core.StringValue(
		input.ResourceSpec.Fields["version"],
	)
	functionOutput, err := lambdaService.GetFunction(
		ctx,
		&lambda.GetFunctionInput{
			FunctionName: &functionARN,
			Qualifier:    aws.String(version),
		},
	)
	if err != nil {
		return nil, err
	}

	functionState := functionOutput.Configuration.State
	hasStabilised := functionState == types.StateActive
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: hasStabilised,
	}, nil
}
