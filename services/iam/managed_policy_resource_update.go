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

func (i *iamManagedPolicyResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Get the policy ARN from the computed ARN field in current state
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

	updateOperations := []pluginutils.SaveOperation[iamservice.Service]{
		&managedPolicyVersionUpdate{},
		&managedPolicyTagsUpdate{},
	}

	hasUpdates, _, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			ProviderUpstreamID: arn,
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
		// Get the updated policy to return computed fields
		getPolicyOutput, err := iamService.GetPolicy(ctx, &iam.GetPolicyInput{
			PolicyArn: aws.String(arn),
		})
		if err != nil {
			return nil, err
		}

		computedFields := i.extractComputedFieldsFromPolicy(getPolicyOutput.Policy)
		return &provider.ResourceDeployOutput{
			ComputedFieldValues: computedFields,
		}, nil
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: i.extractComputedFieldsFromCurrentState(currentStateSpecData),
	}, nil
}

func (i *iamManagedPolicyResourceActions) extractComputedFieldsFromPolicy(
	policy *types.Policy,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if policy != nil {
		if policy.Arn != nil {
			fields["spec.arn"] = core.MappingNodeFromString(*policy.Arn)
		}
		if policy.PolicyId != nil {
			fields["spec.id"] = core.MappingNodeFromString(*policy.PolicyId)
		}
		if policy.AttachmentCount != nil {
			fields["spec.attachmentCount"] = core.MappingNodeFromInt(int(*policy.AttachmentCount))
		}
		if policy.CreateDate != nil {
			fields["spec.createDate"] = core.MappingNodeFromString(policy.CreateDate.Format("2006-01-02T15:04:05Z"))
		}
		if policy.DefaultVersionId != nil {
			fields["spec.defaultVersionId"] = core.MappingNodeFromString(*policy.DefaultVersionId)
		}
		if policy.IsAttachable {
			fields["spec.isAttachable"] = core.MappingNodeFromBool(policy.IsAttachable)
		}
		if policy.PermissionsBoundaryUsageCount != nil {
			fields["spec.permissionsBoundaryUsageCount"] = core.MappingNodeFromInt(int(*policy.PermissionsBoundaryUsageCount))
		}
		if policy.UpdateDate != nil {
			fields["spec.updateDate"] = core.MappingNodeFromString(policy.UpdateDate.Format("2006-01-02T15:04:05Z"))
		}
	}
	return fields
}

func (i *iamManagedPolicyResourceActions) extractComputedFieldsFromCurrentState(
	currentStateSpecData *core.MappingNode,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	fieldsToExtract := map[string]string{
		"$.arn":                           "spec.arn",
		"$.id":                            "spec.id",
		"$.attachmentCount":               "spec.attachmentCount",
		"$.createDate":                    "spec.createDate",
		"$.defaultVersionId":              "spec.defaultVersionId",
		"$.isAttachable":                  "spec.isAttachable",
		"$.permissionsBoundaryUsageCount": "spec.permissionsBoundaryUsageCount",
	}

	for path, field := range fieldsToExtract {
		if v, ok := pluginutils.GetValueByPath(path, currentStateSpecData); ok {
			fields[field] = v
		}
	}

	return fields
}
