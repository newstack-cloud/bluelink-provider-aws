package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type instanceProfileCreate struct {
	instanceProfileName string
	path                string
	uniqueNameGenerator utils.UniqueNameGenerator
}

func (i *instanceProfileCreate) Name() string {
	return "create instance profile"
}

func (i *instanceProfileCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Extract instanceProfileName from spec data
	instanceProfileName, hasInstanceProfileName := pluginutils.GetValueByPath("$.instanceProfileName", specData)
	if hasInstanceProfileName {
		i.instanceProfileName = core.StringValue(instanceProfileName)
	}

	// Extract path from spec data, default to "/"
	path, hasPath := pluginutils.GetValueByPath("$.path", specData)
	if hasPath {
		i.path = core.StringValue(path)
	} else {
		i.path = "/"
	}

	// Generate instance profile name if not provided
	if i.instanceProfileName == "" {
		resourceDeployInput, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
		if !ok {
			return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in save operation context")
		}

		generatedName, err := i.uniqueNameGenerator(resourceDeployInput)
		if err != nil {
			return false, saveOpCtx, fmt.Errorf("failed to generate instance profile name: %w", err)
		}
		i.instanceProfileName = generatedName
	}

	return true, saveOpCtx, nil
}

func (i *instanceProfileCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	input := &iam.CreateInstanceProfileInput{
		InstanceProfileName: aws.String(i.instanceProfileName),
		Path:                aws.String(i.path),
	}

	output, err := iamService.CreateInstanceProfile(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create instance profile: %w", err)
	}

	newSaveOpCtx.Data["createInstanceProfileOutput"] = output
	return newSaveOpCtx, nil
}

type instanceProfileRoleAdd struct {
	instanceProfileName string
	roleName            string
}

func (i *instanceProfileRoleAdd) Name() string {
	return "add role to instance profile"
}

func (i *instanceProfileRoleAdd) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Extract role from spec data
	role, hasRole := pluginutils.GetValueByPath("$.role", specData)
	if !hasRole {
		return false, saveOpCtx, fmt.Errorf("role is required")
	}

	roleSpec := core.StringValue(role)
	roleName, err := extractRoleNameFromRoleSpec(roleSpec)
	if err != nil {
		return false, saveOpCtx, fmt.Errorf("failed to extract role name: %w", err)
	}

	i.roleName = roleName

	// Get the instance profile name from the create operation output
	createInstanceProfileOutput, ok := saveOpCtx.Data["createInstanceProfileOutput"].(*iam.CreateInstanceProfileOutput)
	if !ok {
		return false, saveOpCtx, fmt.Errorf("createInstanceProfileOutput not found")
	}

	i.instanceProfileName = aws.ToString(createInstanceProfileOutput.InstanceProfile.InstanceProfileName)

	return true, saveOpCtx, nil
}

func (i *instanceProfileRoleAdd) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	input := &iam.AddRoleToInstanceProfileInput{
		InstanceProfileName: aws.String(i.instanceProfileName),
		RoleName:            aws.String(i.roleName),
	}

	_, err := iamService.AddRoleToInstanceProfile(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to add role to instance profile: %w", err)
	}

	return newSaveOpCtx, nil
}
