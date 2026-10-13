package lambdadynamodb

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// ProduceResourceContributions states that the table needs a stream, which is what this
// link exists to consume.
//
// Contributed rather than written. Enabling the stream through the DynamoDB API left the
// table's spec saying nothing about it, so the next deployment of the table for any other
// reason would state a spec with no stream specification and turn the stream off. Stating
// it as a contribution puts it in the spec the framework deploys.
func (l *dynamoDBTableLambdaFunctionLinkActions) ProduceResourceContributions(
	ctx context.Context,
	input *provider.LinkProduceResourceContributionsInput,
) (*provider.LinkProduceResourceContributionsOutput, error) {
	tableName := pluginutils.GetResourceName(input.ResourceAInfo)
	if tableName == "" {
		return nil, fmt.Errorf(
			"the DynamoDB table in the link relationship could not be identified",
		)
	}

	// A stream that is already running keeps the view type it was created with. Stating
	// the annotation's view type instead would change a live stream because a link was
	// deployed, which is not what adding a consumer asks for.
	viewType := getStreamViewTypeAnnotation(input.ResourceAInfo)
	if isStreamEnabled(input.ResourceAInfo) {
		if current, has := currentStreamViewType(input.ResourceAInfo); has {
			viewType = current
		}
	}

	return &provider.LinkProduceResourceContributionsOutput{
		Contributions: []*provider.ResourceContribution{
			// Only the view type. The table schema has no streamEnabled field -- the
			// presence of streamSpecification is what turns the stream on -- and sending
			// one makes Cloud Control reject the update with
			// "#/StreamSpecification: extraneous key [StreamEnabled] is not permitted".
			{
				ResourceName: tableName,
				FieldPath:    "spec.streamSpecification.streamViewType",
				Value:        core.MappingNodeFromString(viewType),
				Action:       provider.ContributionActionSet,
				// Removing the link does not turn the stream off. Retraction is the
				// default because contributions are desired state, but a stream can have
				// consumers this link knows nothing about, and disabling it would break
				// them to tidy up after one consumer that left.
				RetainOnRemoval: true,
			},
		},
	}, nil
}

// The view type of a stream that is already running, which is recorded on the table once
// the stream exists.
func currentStreamViewType(resourceInfo *provider.ResourceInfo) (string, bool) {
	if resourceInfo == nil || resourceInfo.CurrentResourceState == nil {
		return "", false
	}

	viewType, has := pluginutils.GetValueByPath(
		"$.streamSpecification.streamViewType",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if !has {
		return "", false
	}

	value := core.StringValue(viewType)

	return value, value != ""
}
