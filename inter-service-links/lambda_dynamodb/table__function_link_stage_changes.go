package lambdadynamodb

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *dynamoDBTableLambdaFunctionLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	// This link creates a Lambda event source mapping as a link-owned intermediary
	// resource (the DynamoDB table's stream cannot be modelled as a spec field of either
	// resource). It is projected into the link's linkData under the "intermediaries" map so
	// its create/update is surfaced as link changes; the stream/function ARNs are derived
	// from the linked resources and are known on deploy when either resource is new.
	changes := &provider.LinkChanges{}
	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	identity := tableFunctionESMIdentity(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	err := linkutils.CollectIntermediaryChanges(currentLinkData, changes, linkutils.StageIntermediary{
		Identity: identity,
		DerivedLeaves: []linkutils.DerivedLeaf{
			{Leaf: "eventSourceArn", ResourceChanges: input.ResourceAChanges, ResourceSpecPath: "$.spec.streamArn"},
			{Leaf: "functionArn", ResourceChanges: input.ResourceBChanges, ResourceSpecPath: "$.spec.arn"},
		},
	})
	if err != nil {
		return nil, err
	}

	// The link grants the function's execution role read access to the table's stream as
	// part of UpdateIntermediaryResources. The grant is scoped to the stream ARN, so it
	// only changes when either resource is (re)created; surface that as a known-on-deploy
	// change to the role's link-data field, mirroring the access link.
	if pluginutils.IsResourceNew(input.ResourceAChanges) ||
		pluginutils.IsResourceNew(input.ResourceBChanges) {
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createStreamLinkDataExecutionRoleName(&input.ResourceBChanges.AppliedResourceInfo),
		)
	}

	// What this link will contribute to the table, declared before it runs so the
	// framework knows the table is waiting on it. The stream is what this link consumes,
	// and stating it in the table's spec is what stops a later deployment of the table
	// turning it off.
	tableName := pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo)
	changes.ResourceDataMappings = map[string]string{
		fmt.Sprintf("%s::spec.streamSpecification.streamViewType", tableName): "contributions[0]",
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
