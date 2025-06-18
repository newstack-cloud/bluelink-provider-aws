package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaEventSourceMappingResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	// Extract the ID from the resource state
	id := core.StringValue(input.ResourceState.SpecData.Fields["id"])
	if id == "" {
		// If no ID is present, there's nothing to delete
		return nil
	}

	deleteEventSourceMappingInput := &lambda.DeleteEventSourceMappingInput{
		UUID: aws.String(id),
	}

	_, err = lambdaService.DeleteEventSourceMapping(ctx, deleteEventSourceMappingInput)
	if err != nil {
		return err
	}

	return nil
}
