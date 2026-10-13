package sqslinks

import (
	"embed"

	"github.com/aws/aws-sdk-go-v2/aws"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

//go:embed descriptions
var descriptions embed.FS

// QueueQueueLink returns a link implementation for
// a link from an SQS queue to another SQS queue (dead-letter queue).
func QueueQueueLink(
	linkServiceDeps pluginutils.LinkServiceDeps[
		*aws.Config,
		sqsservice.Service,
		*aws.Config,
		sqsservice.Service,
	],
) provider.Link {
	description, _ := descriptions.ReadFile("descriptions/queue__queue.md")

	actions := &sqsQueueQueueLinkActions{
		sqsServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
		awsConfigStore:    linkServiceDeps.ResourceAService.ConfigStore,
	}

	return &providerv1.LinkDefinition{
		ResourceTypeA: "aws/sqs/queue",
		ResourceTypeB: "aws/sqs/queue",
		// Neither resource is written by this link. The redrive policy is contributed
		// rather than written, and the framework applies what every link
		// contributing to the resource needs it to say as one update, so no
		// exclusive lock is taken here.
		Modifies: provider.LinkModifiesNeither,
		CardinalityA: provider.LinkCardinality{
			// A source queue can only have one dead-letter queue.
			Max: 1,
		},
		// It doesn't matter which queue is created first,
		// the source queue will be configured with a redrive policy
		// to send failed messages to the dead-letter queue once both have been created.
		Kind:                             provider.LinkKindSoft,
		PriorityResource:                 provider.LinkPriorityResourceNone,
		PlainTextSummary:                 "A link that configures the linked to queue as the dead-letter queue for the linked from queue.",
		FormattedDescription:             string(description),
		AnnotationDefinitions:            sqsQueueQueueLinkAnnotations(),
		StageChangesFunc:                 actions.StageChanges,
		ProduceResourceContributionsFunc: actions.ProduceResourceContributions,
		UpdateLinkedResourcesFunc:        actions.UpdateLinkedResources,
		UpdateIntermediaryResourcesFunc:  actions.UpdateIntermediaryResources,
	}
}

type sqsQueueQueueLinkActions struct {
	sqsServiceFactory pluginutils.ServiceFactory[*aws.Config, sqsservice.Service]
	awsConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
}
