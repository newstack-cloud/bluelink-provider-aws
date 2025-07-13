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

type accessKeyUpdate struct {
	accessKeyID string
	status      string
}

func (a *accessKeyUpdate) Name() string {
	return "update access key status"
}

func (a *accessKeyUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the access key ID from the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	if currentStateSpecData == nil {
		return false, saveOpCtx, fmt.Errorf("current state spec data is required for access key update")
	}
	accessKeyID, hasAccessKeyID := pluginutils.GetValueByPath("$.id", currentStateSpecData)
	if !hasAccessKeyID {
		return false, saveOpCtx, fmt.Errorf("access key ID is required for update")
	}

	a.accessKeyID = core.StringValue(accessKeyID)

	// Extract status from spec data, default to "Active"
	status, hasStatus := pluginutils.GetValueByPath("$.status", specData)
	if hasStatus {
		a.status = core.StringValue(status)
	} else {
		a.status = "Active"
	}

	// Get current status from changes
	currentStatus, hasCurrentStatus := pluginutils.GetValueByPath("$.status", currentStateSpecData)
	if hasCurrentStatus {
		currentStatusStr := core.StringValue(currentStatus)
		// Only update if status has changed
		if currentStatusStr == a.status {
			return false, saveOpCtx, nil
		}
	}

	return true, saveOpCtx, nil
}

func (a *accessKeyUpdate) Execute(
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
