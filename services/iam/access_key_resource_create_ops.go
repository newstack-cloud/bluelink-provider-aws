package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type accessKeyCreate struct {
	userName string
}

func (a *accessKeyCreate) Name() string {
	return "create access key"
}

func (a *accessKeyCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Extract userName from spec data
	userName, hasUserName := pluginutils.GetValueByPath("$.userName", specData)
	if !hasUserName {
		return false, saveOpCtx, fmt.Errorf("userName is required")
	}

	a.userName = core.StringValue(userName)

	return true, saveOpCtx, nil
}

func (a *accessKeyCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	input := &iam.CreateAccessKeyInput{
		UserName: aws.String(a.userName),
	}

	output, err := iamService.CreateAccessKey(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create access key: %w", err)
	}

	newSaveOpCtx.Data["createAccessKeyOutput"] = output
	return newSaveOpCtx, nil
}

type accessKeyStatusUpdate struct {
	accessKeyID string
	status      string
}

func (a *accessKeyStatusUpdate) Name() string {
	return "update access key status"
}

func (a *accessKeyStatusUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the access key ID from the create operation output
	createAccessKeyOutput, ok := saveOpCtx.Data["createAccessKeyOutput"].(*iam.CreateAccessKeyOutput)
	if !ok {
		return false, saveOpCtx, fmt.Errorf("createAccessKeyOutput not found")
	}

	a.accessKeyID = aws.ToString(createAccessKeyOutput.AccessKey.AccessKeyId)

	// Extract status from spec data, default to "Active"
	status, hasStatus := pluginutils.GetValueByPath("$.status", specData)
	if hasStatus {
		a.status = core.StringValue(status)
	} else {
		a.status = "Active"
	}

	// Only update if status is not "Active" (default)
	if a.status == "Active" {
		return false, saveOpCtx, nil
	}

	return true, saveOpCtx, nil
}

func (a *accessKeyStatusUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	status := types.StatusTypeActive
	if a.status == "Inactive" {
		status = types.StatusTypeInactive
	}

	input := &iam.UpdateAccessKeyInput{
		AccessKeyId: aws.String(a.accessKeyID),
		Status:      status,
	}

	_, err := iamService.UpdateAccessKey(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to update access key status: %w", err)
	}

	return newSaveOpCtx, nil
}
