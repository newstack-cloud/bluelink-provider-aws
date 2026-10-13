//go:build unit

package linkutils

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	"github.com/stretchr/testify/suite"
)

type EventSourceMappingSuite struct {
	suite.Suite
}

func esmCreateInput() *lambda.CreateEventSourceMappingInput {
	return &lambda.CreateEventSourceMappingInput{
		EventSourceArn: aws.String("arn:aws:sqs:eu-west-2:123456789012:orders"),
		FunctionName:   aws.String("OrdersConsumer-handle"),
	}
}

func (s *EventSourceMappingSuite) Test_returns_the_created_mapping() {
	svc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingOutput(&lambda.CreateEventSourceMappingOutput{
			UUID:                  aws.String("uuid-new"),
			EventSourceMappingArn: aws.String("arn:esm:new"),
		}),
	)

	ref, err := CreateOrAdoptEventSourceMapping(context.Background(), svc, esmCreateInput())

	s.Require().NoError(err)
	s.Assert().Equal("uuid-new", ref.UUID)
	s.Assert().False(ref.Adopted)
}

// Creating is not idempotent and a link cannot tell from its own state whether it
// already created a mapping, state is persisted when the link finishes, so an
// attempt that created the mapping and then failed leaves no record. The engine's
// retry then hits ResourceConflictException, which failed the whole deployment.
func (s *EventSourceMappingSuite) Test_adopts_the_existing_mapping_on_conflict() {
	svc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingError(&lambdatypes.ResourceConflictException{
			Message: aws.String("An event source mapping ... already exists"),
		}),
		lambdamock.WithListEventSourceMappingsOutput(&lambda.ListEventSourceMappingsOutput{
			EventSourceMappings: []lambdatypes.EventSourceMappingConfiguration{
				{
					UUID:                  aws.String("uuid-existing"),
					EventSourceMappingArn: aws.String("arn:esm:existing"),
				},
			},
		}),
	)

	ref, err := CreateOrAdoptEventSourceMapping(context.Background(), svc, esmCreateInput())

	s.Require().NoError(err)
	s.Assert().Equal("uuid-existing", ref.UUID)
	s.Assert().True(ref.Adopted)
}

// A conflict with nothing to adopt is left as a failure. The mapping AWS is
// objecting to may not be the one asked for, and adopting blind could point a
// live trigger at something unintended.
func (s *EventSourceMappingSuite) Test_keeps_the_conflict_when_nothing_matches() {
	svc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingError(&lambdatypes.ResourceConflictException{
			Message: aws.String("already exists"),
		}),
		lambdamock.WithListEventSourceMappingsOutput(&lambda.ListEventSourceMappingsOutput{}),
	)

	_, err := CreateOrAdoptEventSourceMapping(context.Background(), svc, esmCreateInput())

	var conflict *lambdatypes.ResourceConflictException
	s.Assert().ErrorAs(err, &conflict, "the original conflict is what the caller needs to see")
}

// Any other failure is passed through untouched, so a genuine problem is not
// turned into a lookup.
func (s *EventSourceMappingSuite) Test_passes_other_errors_through() {
	throttleErr := errors.New("throttled")
	svc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingError(throttleErr),
	)

	_, err := CreateOrAdoptEventSourceMapping(context.Background(), svc, esmCreateInput())

	s.Assert().ErrorIs(err, throttleErr)
}

// An adopted mapping is brought into line with the requested settings. It cannot
// be assumed to match, a source with a stable name collides with mappings left by
// earlier deployments too, and one of those left in Disabled state would
// otherwise be adopted as a trigger that never fires.
func (s *EventSourceMappingSuite) Test_reconciles_the_adopted_mapping() {
	var updated *lambda.UpdateEventSourceMappingInput
	svc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingError(&lambdatypes.ResourceConflictException{
			Message: aws.String("already exists"),
		}),
		lambdamock.WithListEventSourceMappingsOutput(&lambda.ListEventSourceMappingsOutput{
			EventSourceMappings: []lambdatypes.EventSourceMappingConfiguration{
				{
					UUID:                  aws.String("uuid-stale"),
					EventSourceMappingArn: aws.String("arn:esm:stale"),
					State:                 aws.String("Disabled"),
				},
			},
		}),
	)
	svc.MockUpdateEventSourceMapping = func(
		ctx context.Context,
		in *lambda.UpdateEventSourceMappingInput,
	) (*lambda.UpdateEventSourceMappingOutput, error) {
		updated = in
		return &lambda.UpdateEventSourceMappingOutput{}, nil
	}

	createInput := esmCreateInput()
	createInput.Enabled = aws.Bool(true)
	createInput.BatchSize = aws.Int32(10)

	ref, err := CreateOrAdoptEventSourceMapping(context.Background(), svc, createInput)

	s.Require().NoError(err)
	s.Assert().Equal("uuid-stale", ref.UUID)
	s.Require().NotNil(updated, "an adopted mapping must be reconciled, not trusted")
	s.Assert().Equal("uuid-stale", aws.ToString(updated.UUID))
	s.Assert().True(aws.ToBool(updated.Enabled), "a disabled leftover must be enabled again")
	s.Assert().Equal(int32(10), aws.ToInt32(updated.BatchSize))
}

// A reconcile that fails is a failure, the mapping exists but is not configured
// as asked, and reporting success would leave a trigger in an unknown state.
func (s *EventSourceMappingSuite) Test_fails_when_the_adopted_mapping_cannot_be_reconciled() {
	rejected := errors.New("update rejected")
	svc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithCreateEventSourceMappingError(&lambdatypes.ResourceConflictException{
			Message: aws.String("already exists"),
		}),
		lambdamock.WithListEventSourceMappingsOutput(&lambda.ListEventSourceMappingsOutput{
			EventSourceMappings: []lambdatypes.EventSourceMappingConfiguration{
				{UUID: aws.String("uuid-stale")},
			},
		}),
	)
	svc.MockUpdateEventSourceMapping = func(
		ctx context.Context,
		in *lambda.UpdateEventSourceMappingInput,
	) (*lambda.UpdateEventSourceMappingOutput, error) {
		return nil, rejected
	}

	_, err := CreateOrAdoptEventSourceMapping(context.Background(), svc, esmCreateInput())

	s.Assert().ErrorIs(err, rejected)
}

func TestEventSourceMappingSuite(t *testing.T) {
	suite.Run(t, new(EventSourceMappingSuite))
}
