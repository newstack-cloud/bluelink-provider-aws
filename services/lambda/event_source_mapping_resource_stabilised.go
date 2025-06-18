package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaEventSourceMappingResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	uuid := core.StringValue(
		input.ResourceSpec.Fields["id"],
	)
	getEventSourceMappingOutput, err := lambdaService.GetEventSourceMapping(
		ctx,
		&lambda.GetEventSourceMappingInput{
			UUID: &uuid,
		},
	)
	if err != nil {
		return nil, err
	}

	state := aws.ToString(getEventSourceMappingOutput.State)
	return &provider.ResourceHasStabilisedOutput{
		// When an event source mapping has finished being created or updated,
		// it will be in either the "Enabled" or "Disabled" state.
		Stabilised: state == "Enabled" || state == "Disabled",
	}, nil
}
