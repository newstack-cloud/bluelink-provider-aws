package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type roleUpdate struct {
	updateRoleInput             *iam.UpdateRoleInput
	updateAssumeRolePolicyInput *iam.UpdateAssumeRolePolicyInput
}

func (u *roleUpdate) Name() string {
	return "update IAM role"
}

func (u *roleUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	roleName := saveOpCtx.ProviderUpstreamID

	hasChanges := false

	// Check for basic role updates (description, maxSessionDuration, path)
	updateRoleInput, hasRoleUpdate, err := changesToUpdateRoleInput(roleName, specData, changes)
	if err != nil {
		return false, saveOpCtx, err
	}
	if hasRoleUpdate {
		u.updateRoleInput = updateRoleInput
		hasChanges = true
	}

	// Check for assume role policy updates
	updateAssumeRolePolicyInput, hasAssumeRolePolicyUpdate, err := changesToUpdateAssumeRolePolicyInput(roleName, specData, changes)
	if err != nil {
		return false, saveOpCtx, err
	}
	if hasAssumeRolePolicyUpdate {
		u.updateAssumeRolePolicyInput = updateAssumeRolePolicyInput
		hasChanges = true
	}

	return hasChanges, saveOpCtx, nil
}

func (u *roleUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data:               saveOpCtx.Data,
		ProviderUpstreamID: saveOpCtx.ProviderUpstreamID,
	}

	// Execute basic role updates
	if u.updateRoleInput != nil {
		updateRoleOutput, err := iamService.UpdateRole(ctx, u.updateRoleInput)
		if err != nil {
			return saveOpCtx, err
		}
		newSaveOpCtx.Data["updateRoleOutput"] = updateRoleOutput
	}

	// Execute assume role policy updates
	if u.updateAssumeRolePolicyInput != nil {
		updateAssumeRolePolicyOutput, err := iamService.UpdateAssumeRolePolicy(ctx, u.updateAssumeRolePolicyInput)
		if err != nil {
			return saveOpCtx, err
		}
		newSaveOpCtx.Data["updateAssumeRolePolicyOutput"] = updateAssumeRolePolicyOutput
	}

	return newSaveOpCtx, nil
}

func changesToUpdateRoleInput(
	roleName string,
	specData *core.MappingNode,
	changes *provider.Changes,
) (*iam.UpdateRoleInput, bool, error) {
	input := &iam.UpdateRoleInput{
		RoleName: aws.String(roleName),
	}

	hasValuesToSave := false

	// Check if description was modified
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.description" {
			if description, exists := pluginutils.GetValueByPath("$.description", specData); exists {
				input.Description = aws.String(core.StringValue(description))
				hasValuesToSave = true
			}
		}
		if fieldChange.FieldPath == "spec.maxSessionDuration" {
			if maxSessionDuration, exists := pluginutils.GetValueByPath("$.maxSessionDuration", specData); exists {
				input.MaxSessionDuration = aws.Int32(int32(core.IntValue(maxSessionDuration)))
				hasValuesToSave = true
			}
		}
	}

	return input, hasValuesToSave, nil
}

func changesToUpdateAssumeRolePolicyInput(
	roleName string,
	specData *core.MappingNode,
	changes *provider.Changes,
) (*iam.UpdateAssumeRolePolicyInput, bool, error) {
	// Check if assumeRolePolicyDocument was modified
	assumeRolePolicyModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.assumeRolePolicyDocument" {
			assumeRolePolicyModified = true
			break
		}
	}

	if !assumeRolePolicyModified {
		return nil, false, nil
	}

	assumeRolePolicyDocument, exists := pluginutils.GetValueByPath("$.assumeRolePolicyDocument", specData)
	if !exists {
		return nil, false, nil
	}

	// Convert the structured policy document to JSON string
	policyJSON, err := json.Marshal(assumeRolePolicyDocument)
	if err != nil {
		return nil, false, err
	}

	// URL encode the policy document
	policyDocument := string(policyJSON)
	encodedPolicyDocument := url.QueryEscape(policyDocument)

	input := &iam.UpdateAssumeRolePolicyInput{
		RoleName:       aws.String(roleName),
		PolicyDocument: aws.String(encodedPolicyDocument),
	}

	return input, true, nil
}

type roleInlinePoliciesUpdate struct {
	policies []*core.MappingNode
}

func (r *roleInlinePoliciesUpdate) Name() string {
	return "update inline policies"
}

func (r *roleInlinePoliciesUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if policies were modified
	policiesModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.policies" {
			policiesModified = true
			break
		}
	}

	if !policiesModified {
		return false, saveOpCtx, nil
	}

	// Check if there are inline policies to update
	if policiesNode, ok := specData.Fields["policies"]; ok && policiesNode != nil && len(policiesNode.Items) > 0 {
		r.policies = policiesNode.Items
		return true, saveOpCtx, nil
	}

	return false, saveOpCtx, nil
}

func (r *roleInlinePoliciesUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	roleName := saveOpCtx.ProviderUpstreamID

	// Update each inline policy
	for _, policyNode := range r.policies {
		policyName := core.StringValue(policyNode.Fields["policyName"])
		policyDocNode := policyNode.Fields["policyDocument"]

		policyJSON, err := json.Marshal(policyDocNode)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to marshal inline policy document: %w", err)
		}

		_, err = iamService.PutRolePolicy(ctx, &iam.PutRolePolicyInput{
			RoleName:       aws.String(roleName),
			PolicyName:     aws.String(policyName),
			PolicyDocument: aws.String(string(policyJSON)),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to update inline policy %s: %w", policyName, err)
		}
	}

	return saveOpCtx, nil
}

type roleManagedPoliciesUpdate struct {
	managedPolicyArns []string
}

func (r *roleManagedPoliciesUpdate) Name() string {
	return "update managed policies"
}

func (r *roleManagedPoliciesUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if managedPolicyArns was modified
	managedPolicyArnsModified := false
	for _, fieldChange := range changes.ModifiedFields {
		if fieldChange.FieldPath == "spec.managedPolicyArns" {
			managedPolicyArnsModified = true
			break
		}
	}

	if !managedPolicyArnsModified {
		return false, saveOpCtx, nil
	}

	// Check if there are managed policy ARNs to update
	if managedPolicyArnsNode, ok := specData.Fields["managedPolicyArns"]; ok && managedPolicyArnsNode != nil && len(managedPolicyArnsNode.Items) > 0 {
		r.managedPolicyArns = make([]string, len(managedPolicyArnsNode.Items))
		for i, arnNode := range managedPolicyArnsNode.Items {
			r.managedPolicyArns[i] = core.StringValue(arnNode)
		}
		return true, saveOpCtx, nil
	}

	return false, saveOpCtx, nil
}

func (r *roleManagedPoliciesUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	roleName := saveOpCtx.ProviderUpstreamID

	// Attach each managed policy
	for _, policyArn := range r.managedPolicyArns {
		_, err := iamService.AttachRolePolicy(ctx, &iam.AttachRolePolicyInput{
			RoleName:  aws.String(roleName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to attach managed policy %s: %w", policyArn, err)
		}
	}

	return saveOpCtx, nil
}
