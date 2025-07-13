package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamGroupResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Get the group ARN from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(input.Changes)
	arnValue, err := core.GetPathValue(
		"$.arn",
		currentStateSpecData,
		core.MappingNodeMaxTraverseDepth,
	)
	if err != nil {
		return nil, err
	}

	arn := core.StringValue(arnValue)
	if arn == "" {
		return nil, fmt.Errorf("ARN is required for update operation")
	}

	// Extract group name from ARN
	groupName, err := extractGroupNameFromARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to extract group name from ARN: %w", err)
	}

	updateOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&groupUpdate{groupName: groupName},
		&groupInlinePoliciesUpdate{groupName: groupName},
		&groupManagedPoliciesUpdate{groupName: groupName},
	}

	saveOpCtx := pluginutils.SaveOperationContext{
		Data: map[string]any{
			"ResourceDeployInput": input,
		},
	}

	_, _, err = pluginutils.RunSaveOperations(
		ctx,
		saveOpCtx,
		updateOperations,
		input,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	// Get current group details for computed fields
	getGroupOutput, err := iamService.GetGroup(ctx, &iam.GetGroupInput{
		GroupName: aws.String(groupName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get group details after update: %w", err)
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn":       core.MappingNodeFromString(aws.ToString(getGroupOutput.Group.Arn)),
		"spec.groupId":   core.MappingNodeFromString(aws.ToString(getGroupOutput.Group.GroupId)),
		"spec.groupName": core.MappingNodeFromString(aws.ToString(getGroupOutput.Group.GroupName)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
