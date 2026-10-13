//go:build unit

package sqslinks

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	sqsservice "github.com/newstack-cloud/bluelink-provider-aws/services/sqs/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type QueueQueueLinkContributionsSuite struct {
	suite.Suite
}

func sourceQueueInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "source-queue",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"queueUrl",
				core.MappingNodeFromString("https://sqs.us-west-2.amazonaws.com/123456789012/source"),
			),
		},
	}
}

func deadLetterQueueInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "dead-letter-queue",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"arn", core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:dlq"),
			),
		},
	}
}

func (s *QueueQueueLinkContributionsSuite) Test_produce_resource_contributions() {
	testCases := []plugintestutils.LinkProduceContributionsTestCase[
		*aws.Config, sqsservice.Service, *aws.Config, sqsservice.Service,
	]{
		{
			// Stated rather than set as a queue attribute through the SQS API: a policy
			// set through the API is absent from the queue's spec, so the next deployment
			// of the queue would drop it and undeliverable messages would stop reaching
			// the dead-letter queue, silently.
			Name: "contributes the redrive policy to the source queue",
			Input: &provider.LinkProduceResourceContributionsInput{
				LinkUpdateType: provider.LinkUpdateTypeCreate,
				ResourceAInfo:  sourceQueueInfo(),
				ResourceBInfo:  deadLetterQueueInfo(),
			},
			ExpectedOutput: &provider.LinkProduceResourceContributionsOutput{
				Contributions: []*provider.ResourceContribution{
					{
						ResourceName: "source-queue",
						FieldPath:    "spec.redrivePolicy",
						Value: core.MappingNodeFields(
							"deadLetterTargetArn",
							core.MappingNodeFromString("arn:aws:sqs:us-west-2:123456789012:dlq"),
							"maxReceiveCount",
							core.MappingNodeFromInt(10),
						),
						Action: provider.ContributionActionSet,
					},
				},
			},
		},
		{
			Name: "returns an error when the dead-letter queue ARN is missing",
			Input: &provider.LinkProduceResourceContributionsInput{
				LinkUpdateType: provider.LinkUpdateTypeCreate,
				ResourceAInfo:  sourceQueueInfo(),
				ResourceBInfo: &provider.ResourceInfo{
					ResourceName:         "dead-letter-queue",
					CurrentResourceState: &state.ResourceState{SpecData: core.MappingNodeFields()},
				},
			},
			ExpectError:          true,
			ExpectedErrorMessage: "queue ARN could not be retrieved",
		},
	}

	plugintestutils.RunLinkProduceContributionsTestCases(
		testCases, QueueQueueLink, &s.Suite,
	)
}

func TestQueueQueueLinkContributionsSuite(t *testing.T) {
	suite.Run(t, new(QueueQueueLinkContributionsSuite))
}
