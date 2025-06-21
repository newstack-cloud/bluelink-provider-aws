package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/celerity-provider-aws/services/iam/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamRoleResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Get the role name from the computed ARN field in current state
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

	// Extract role name from ARN
	roleName, err := extractRoleNameFromARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to extract role name from ARN %s: %w", arn, err)
	}

	updateOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&roleUpdate{},
		&roleInlinePoliciesUpdate{},
		&roleManagedPoliciesUpdate{},
	}

	hasUpdates, _, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			ProviderUpstreamID: roleName,
			Data:               make(map[string]any),
		},
		updateOperations,
		input,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	if hasUpdates {
		// Get the updated role to return computed fields
		getRoleOutput, err := iamService.GetRole(ctx, &iam.GetRoleInput{
			RoleName: aws.String(roleName),
		})
		if err != nil {
			return nil, err
		}

		computedFields := i.extractComputedFieldsFromRole(getRoleOutput.Role)
		return &provider.ResourceDeployOutput{
			ComputedFieldValues: computedFields,
		}, nil
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: i.extractComputedFieldsFromCurrentState(currentStateSpecData),
	}, nil
}

func (i *iamRoleResourceActions) extractComputedFieldsFromRole(
	role *types.Role,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if role != nil {
		if role.Arn != nil {
			fields["spec.arn"] = core.MappingNodeFromString(*role.Arn)
		}
		if role.RoleId != nil {
			fields["spec.roleId"] = core.MappingNodeFromString(*role.RoleId)
		}
	}
	return fields
}

func (i *iamRoleResourceActions) extractComputedFieldsFromCurrentState(
	currentStateSpecData *core.MappingNode,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if v, ok := pluginutils.GetValueByPath("$.arn", currentStateSpecData); ok {
		fields["spec.arn"] = v
	}
	if v, ok := pluginutils.GetValueByPath("$.roleId", currentStateSpecData); ok {
		fields["spec.roleId"] = v
	}
	return fields
}
