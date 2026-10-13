package snssqs

import (
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// SubscriptionQueueLink returns a link implementation for an SNS subscription (sqs
// protocol) delivering to an SQS queue. The subscription wiring and all of its
// configuration are modelled on the aws/sns/subscription resource; this link only grants
// the queue's resource-based policy allowing SNS to send messages to it, scoped to the
// subscription's topic. It is reference-implied, it activates when the subscription's
// endpoint references the queue.
func SubscriptionQueueLink() provider.Link {
	description, _ := descriptions.ReadFile("descriptions/subscription__queue.md")

	actions := &subscriptionQueueLinkActions{}

	return &providerv1.LinkDefinition{
		ResourceTypeA: "aws/sns/subscription",
		ResourceTypeB: "aws/sqs/queue",
		// Neither resource in the relationship is written; the link does its work
		// through intermediary resources.
		Modifies:                        provider.LinkModifiesNeither,
		Kind:                            provider.LinkKindSoft,
		PriorityResource:                provider.LinkPriorityResourceNone,
		PlainTextSummary:                "A link that grants SNS permission to deliver a subscription's messages to an SQS queue.",
		FormattedDescription:            string(description),
		StageChangesFunc:                actions.StageChanges,
		UpdateLinkedResourcesFunc:       actions.UpdateLinkedResources,
		UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
	}
}

type subscriptionQueueLinkActions struct{}
