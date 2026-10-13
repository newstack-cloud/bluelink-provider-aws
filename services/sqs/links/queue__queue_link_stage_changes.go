package sqslinks

import (
	"context"
	"fmt"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *sqsQueueQueueLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	sourceQueueResourceName := linkhelpers.GetResourceNameFromChanges(input.ResourceAChanges)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	// Collect deadLetterTargetArn changes from the dead-letter queue's ARN
	deadLetterTargetArnWriteToPath := fmt.Sprintf(
		"$[%q].redrivePolicy.deadLetterTargetArn",
		sourceQueueResourceName,
	)
	err := linkhelpers.CollectChanges(
		"$.spec.arn",
		deadLetterTargetArnWriteToPath,
		currentLinkData,
		input.ResourceBChanges,
		changes,
	)
	if err != nil {
		return nil, err
	}

	// Collect maxReceiveCount changes from annotations
	maxReceiveCountWriteToPath := fmt.Sprintf(
		"$[%q].redrivePolicy.maxReceiveCount",
		sourceQueueResourceName,
	)

	// Get the annotation value (default to 10 if not present)
	maxReceiveCountAnnotation := linkhelpers.GetAnnotation(
		input.ResourceAChanges,
		"aws.sqs.redrive.maxReceiveCount",
		core.MappingNodeFromInt(10),
	)

	err = linkhelpers.CollectLinkDataChanges(
		maxReceiveCountWriteToPath,
		currentLinkData,
		changes,
		maxReceiveCountAnnotation,
	)
	if err != nil {
		return nil, err
	}

	// What this link will contribute, declared before it runs so the framework knows the
	// resource is waiting on it. The value names the contribution rather than a path built
	// from the resource's name: a name with a hyphen in it is not a valid path segment, and
	// this is the form the framework records once the link has produced it.
	changes.ResourceDataMappings = map[string]string{
		fmt.Sprintf(
			"%s::spec.redrivePolicy",
			pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo),
		): "contributions[0]",
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
