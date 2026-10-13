package sqslinks

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// ProduceResourceContributions states the redrive policy that sends the source queue's
// undeliverable messages to the dead-letter queue it is linked to.
//
// Contributed rather than written. Setting the queue attribute through the SQS API left
// the queue's spec saying nothing about the policy, so the next deployment of the queue for
// any other reason would state a spec without it and messages would stop being sent to the
// dead-letter queue, silently.
func (l *sqsQueueQueueLinkActions) ProduceResourceContributions(
	ctx context.Context,
	input *provider.LinkProduceResourceContributionsInput,
) (*provider.LinkProduceResourceContributionsOutput, error) {
	deadLetterQueueARN, hasARN := pluginutils.GetValueByPath(
		"$.arn",
		pluginutils.GetCurrentStateSpecDataFromResourceInfo(input.ResourceBInfo),
	)
	if !hasARN {
		return nil, fmt.Errorf(
			"queue ARN could not be retrieved from dead-letter queue %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	maxReceiveCount, _ := pluginutils.GetIntAnnotation(
		input.ResourceAInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     "aws.sqs.redrive.maxReceiveCount",
			Default: 10,
		},
	)

	return &provider.LinkProduceResourceContributionsOutput{
		Contributions: []*provider.ResourceContribution{
			{
				ResourceName: pluginutils.GetResourceName(input.ResourceAInfo),
				FieldPath:    "spec.redrivePolicy",
				Value: core.MappingNodeFields(
					"deadLetterTargetArn",
					core.MappingNodeFromString(core.StringValue(deadLetterQueueARN)),
					"maxReceiveCount",
					core.MappingNodeFromInt(maxReceiveCount),
				),
				// A queue has one redrive policy. Two links naming it is a disagreement
				// the framework reports rather than a list to add to.
				Action: provider.ContributionActionSet,
			},
		},
	}, nil
}
