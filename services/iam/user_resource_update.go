package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/celerity-provider-aws/services/iam/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamUserResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Get the user ARN from the current state
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

	// Extract user name from ARN
	userName, err := extractUserNameFromARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to extract user name from ARN: %w", err)
	}

	updateOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&userUpdateBasic{userName: userName},
		&userLoginProfileUpdate{userName: userName},
		&userInlinePoliciesUpdate{userName: userName},
		&userManagedPoliciesUpdate{userName: userName},
		&userPermissionsBoundaryUpdate{userName: userName},
		&userTagsUpdate{userName: userName},
		&userGroupMembershipUpdate{userName: userName},
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

	// Get current user details for computed fields
	getUserOutput, err := iamService.GetUser(ctx, &iam.GetUserInput{
		UserName: aws.String(userName),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user details after update: %w", err)
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn":    core.MappingNodeFromString(aws.ToString(getUserOutput.User.Arn)),
		"spec.userId": core.MappingNodeFromString(aws.ToString(getUserOutput.User.UserId)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}
