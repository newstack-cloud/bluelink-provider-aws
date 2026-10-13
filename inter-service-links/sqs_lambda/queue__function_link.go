package sqslambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// QueueToFunctionLinkDeps is a type alias for the core dependencies for a link from
// an SQS queue to a Lambda function (trigger). The queue is a Cloud Control–backed
// resource, so resource A's service is the Cloud Control service; the link only reads
// the queue ARN from state and does not call it.
type QueueToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	lambdaservice.Service,
]

// SQSQueueLambdaFunctionLink returns a link implementation for a link from an SQS
// queue to a Lambda function. This creates an Event Source Mapping that triggers the
// Lambda function when messages are sent to the queue. It also adds IAM permissions
// to the Lambda function's execution role to consume from the queue.
func SQSQueueLambdaFunctionLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
) func(QueueToFunctionLinkDeps) provider.Link {
	return func(
		linkServiceDeps QueueToFunctionLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/queue__function.md")

		actions := &sqsQueueLambdaFunctionLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceBService.ServiceFactory,
			iamServiceFactory:    iamServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceBService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/sqs/queue",
			ResourceTypeB: "aws/lambda/function",
			// Neither resource in the relationship is written; the link does its work
			// through intermediary resources.
			Modifies: provider.LinkModifiesNeither,
			// The queue's ARN is read when the Event Source Mapping is created, which is
			// this link's own work once both resources have been deployed, so neither has
			// to exist before the other.
			Kind:                            provider.LinkKindSoft,
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that configures an SQS queue to trigger a Lambda function via an event source mapping.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           sqsQueueLambdaFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateLinkedResourcesFunc:       actions.UpdateLinkedResources,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type sqsQueueLambdaFunctionLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	iamServiceFactory    pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *sqsQueueLambdaFunctionLinkActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}

func (l *sqsQueueLambdaFunctionLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.iamServiceFactory(awsConfig, providerContext), nil
}
