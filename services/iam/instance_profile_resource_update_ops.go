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

type instanceProfileRoleUpdate struct {
	instanceProfileName string
	oldRoleName         string
	newRoleName         string
}

func (i *instanceProfileRoleUpdate) Name() string {
	return "update instance profile role"
}

func (i *instanceProfileRoleUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Extract role from spec data
	role, hasRole := pluginutils.GetValueByPath("$.role", specData)
	if !hasRole {
		return false, saveOpCtx, fmt.Errorf("role is required")
	}

	newRoleSpec := core.StringValue(role)
	newRoleName, err := extractRoleNameFromRoleSpec(newRoleSpec)
	if err != nil {
		return false, saveOpCtx, fmt.Errorf("failed to extract new role name: %w", err)
	}

	i.newRoleName = newRoleName

	// Get the current state spec data
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	if currentStateSpecData == nil {
		return false, saveOpCtx, fmt.Errorf("current state spec data is required for instance profile update")
	}

	// Get the instance profile name from the current state
	arn, hasArn := pluginutils.GetValueByPath("$.arn", currentStateSpecData)
	if !hasArn {
		return false, saveOpCtx, fmt.Errorf("ARN is required for instance profile update")
	}

	instanceProfileName, err := extractInstanceProfileNameFromARN(core.StringValue(arn))
	if err != nil {
		return false, saveOpCtx, fmt.Errorf("failed to extract instance profile name: %w", err)
	}

	i.instanceProfileName = instanceProfileName

	// Get the old role name from the current state
	oldRole, hasOldRole := pluginutils.GetValueByPath("$.role", currentStateSpecData)
	if hasOldRole {
		oldRoleSpec := core.StringValue(oldRole)
		oldRoleName, err := extractRoleNameFromRoleSpec(oldRoleSpec)
		if err != nil {
			return false, saveOpCtx, fmt.Errorf("failed to extract old role name: %w", err)
		}
		i.oldRoleName = oldRoleName
	}

	// Only update if the role has changed
	if i.oldRoleName == i.newRoleName {
		return false, saveOpCtx, nil
	}

	return true, saveOpCtx, nil
}

func (i *instanceProfileRoleUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	// Remove the old role if it exists
	if i.oldRoleName != "" {
		removeRoleInput := &iam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: aws.String(i.instanceProfileName),
			RoleName:            aws.String(i.oldRoleName),
		}

		_, err := iamService.RemoveRoleFromInstanceProfile(ctx, removeRoleInput)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove old role from instance profile: %w", err)
		}
	}

	// Add the new role
	addRoleInput := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(i.instanceProfileName),
		RoleName:            aws.String(i.newRoleName),
	}

	_, err := iamService.AddRoleToInstanceProfile(ctx, addRoleInput)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to add new role to instance profile: %w", err)
	}

	return newSaveOpCtx, nil
}
