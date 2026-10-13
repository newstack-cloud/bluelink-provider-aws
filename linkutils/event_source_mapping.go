package linkutils

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
)

// EventSourceMappingRef identifies a mapping that already exists in AWS.
type EventSourceMappingRef struct {
	UUID string
	ARN  string
	// Adopted reports whether the mapping was already there rather than created
	// by this call.
	Adopted bool
}

// CreateOrAdoptEventSourceMapping creates an event source mapping, and takes over the
// existing one when AWS says there already is one.
//
// Creating is not idempotent, and a link cannot tell from its own state whether it has
// already created a mapping, its state is persisted when it finishes, so an attempt that
// created the mapping and then failed leaves no record. The engine's retry then creates a
// second one and AWS answers with ResourceConflictException, so a conflict means a mapping
// for this source and function already exists and is adopted rather than failed. It is
// found by listing rather than by reading the UUID out of the error text, which is not part
// of the API contract.
func CreateOrAdoptEventSourceMapping(
	ctx context.Context,
	lambdaService lambdaservice.Service,
	createInput *lambda.CreateEventSourceMappingInput,
) (*EventSourceMappingRef, error) {
	output, err := lambdaService.CreateEventSourceMapping(ctx, createInput)
	if err == nil {
		return &EventSourceMappingRef{
			UUID: aws.ToString(output.UUID),
			ARN:  aws.ToString(output.EventSourceMappingArn),
		}, nil
	}

	var conflict *lambdatypes.ResourceConflictException
	if !errors.As(err, &conflict) {
		return nil, err
	}

	existing, lookupErr := findEventSourceMapping(
		ctx,
		lambdaService,
		aws.ToString(createInput.EventSourceArn),
		aws.ToString(createInput.FunctionName),
	)
	if lookupErr != nil {
		return nil, err
	}

	if existing == nil {
		// A conflict with nothing to adopt is not something to paper over, the
		// mapping may belong to a different source or function than the one asked
		// for, and creating or updating blindly could point a live trigger somewhere
		// unintended.
		return nil, err
	}

	if reconcileErr := reconcileAdoptedMapping(
		ctx, lambdaService, existing.UUID, createInput,
	); reconcileErr != nil {
		return nil, reconcileErr
	}

	return existing, nil
}

// Applies the requested settings to a mapping that was adopted rather than created, so an
// adoption cannot quietly leave a trigger configured as an earlier deployment left it.
//
// The source is not something that can be updated, and the mapping was found by
// that ARN in the first place.
func reconcileAdoptedMapping(
	ctx context.Context,
	lambdaService lambdaservice.Service,
	uuid string,
	createInput *lambda.CreateEventSourceMappingInput,
) error {
	updateInput := &lambda.UpdateEventSourceMappingInput{
		UUID:                           aws.String(uuid),
		FunctionName:                   createInput.FunctionName,
		Enabled:                        createInput.Enabled,
		BatchSize:                      createInput.BatchSize,
		MaximumBatchingWindowInSeconds: createInput.MaximumBatchingWindowInSeconds,
		FunctionResponseTypes:          createInput.FunctionResponseTypes,
		FilterCriteria:                 createInput.FilterCriteria,
		ScalingConfig:                  createInput.ScalingConfig,
	}

	_, err := lambdaService.UpdateEventSourceMapping(ctx, updateInput)
	return err
}

func findEventSourceMapping(
	ctx context.Context,
	lambdaService lambdaservice.Service,
	eventSourceARN string,
	functionName string,
) (*EventSourceMappingRef, error) {
	output, err := lambdaService.ListEventSourceMappings(
		ctx,
		&lambda.ListEventSourceMappingsInput{
			EventSourceArn: aws.String(eventSourceARN),
			FunctionName:   aws.String(functionName),
		},
	)
	if err != nil {
		return nil, err
	}

	for _, mapping := range output.EventSourceMappings {
		if aws.ToString(mapping.UUID) == "" {
			continue
		}
		return &EventSourceMappingRef{
			UUID:    aws.ToString(mapping.UUID),
			ARN:     aws.ToString(mapping.EventSourceMappingArn),
			Adopted: true,
		}, nil
	}

	return nil, nil
}
