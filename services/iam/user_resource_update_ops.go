package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/celerity-provider-aws/services/iam/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

type userUpdateBasic struct {
	userName string
}

func (u *userUpdateBasic) Name() string {
	return "update user basic properties"
}

func (u *userUpdateBasic) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are any basic property changes that require an update
	// For users, the main updateable property is the path, but it requires recreation
	// Most user properties are either computed or require recreation
	return false, saveOpCtx, nil
}

func (u *userUpdateBasic) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Currently no basic properties that can be updated without recreation
	return saveOpCtx, nil
}

type userLoginProfileUpdate struct {
	userName     string
	loginProfile *core.MappingNode
	operation    string // "create", "update", "delete"
}

func (u *userLoginProfileUpdate) Name() string {
	return "update login profile"
}

func (u *userLoginProfileUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if login profile needs to be created, updated, or deleted
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentProfile, _ := pluginutils.GetValueByPath("$.loginProfile", currentStateSpecData)
	newProfile := specData.Fields["loginProfile"]

	// Determine operation based on previous and current state
	if currentProfile == nil && newProfile != nil {
		u.operation = "create"
		u.loginProfile = newProfile
		return true, saveOpCtx, nil
	} else if currentProfile != nil && newProfile == nil {
		u.operation = "delete"
		return true, saveOpCtx, nil
	} else if currentProfile != nil && newProfile != nil {
		// Check if password or passwordResetRequired changed
		u.operation = "update"
		u.loginProfile = newProfile
		return true, saveOpCtx, nil
	}

	return false, saveOpCtx, nil
}

func (u *userLoginProfileUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	switch u.operation {
	case "create":
		password := core.StringValue(u.loginProfile.Fields["password"])
		var passwordResetRequired *bool
		if resetRequiredNode, ok := u.loginProfile.Fields["passwordResetRequired"]; ok && resetRequiredNode != nil {
			resetRequired := core.BoolValue(resetRequiredNode)
			passwordResetRequired = &resetRequired
		}

		createInput := &iam.CreateLoginProfileInput{
			UserName: aws.String(u.userName),
			Password: aws.String(password),
		}
		if passwordResetRequired != nil {
			createInput.PasswordResetRequired = *passwordResetRequired
		}

		_, err := iamService.CreateLoginProfile(ctx, createInput)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to create login profile: %w", err)
		}

	case "update":
		password := core.StringValue(u.loginProfile.Fields["password"])
		var passwordResetRequired *bool
		if resetRequiredNode, ok := u.loginProfile.Fields["passwordResetRequired"]; ok && resetRequiredNode != nil {
			resetRequired := core.BoolValue(resetRequiredNode)
			passwordResetRequired = &resetRequired
		}

		updateInput := &iam.UpdateLoginProfileInput{
			UserName: aws.String(u.userName),
			Password: aws.String(password),
		}
		if passwordResetRequired != nil {
			updateInput.PasswordResetRequired = passwordResetRequired
		}

		_, err := iamService.UpdateLoginProfile(ctx, updateInput)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to update login profile: %w", err)
		}

	case "delete":
		_, err := iamService.DeleteLoginProfile(ctx, &iam.DeleteLoginProfileInput{
			UserName: aws.String(u.userName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to delete login profile: %w", err)
		}
	}

	return saveOpCtx, nil
}

type userInlinePoliciesUpdate struct {
	userName string
	toAdd    []*core.MappingNode
	toUpdate []*core.MappingNode
	toRemove []string
}

func (u *userInlinePoliciesUpdate) Name() string {
	return "update inline policies"
}

func (u *userInlinePoliciesUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Compare current and desired inline policies
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentPolicies, _ := pluginutils.GetValueByPath("$.policies", currentStateSpecData)
	newPolicies := specData.Fields["policies"]

	// Create maps for easier comparison
	currentMap := make(map[string]*core.MappingNode)
	if currentPolicies != nil {
		for _, policy := range currentPolicies.Items {
			policyName := core.StringValue(policy.Fields["policyName"])
			currentMap[policyName] = policy
		}
	}

	newMap := make(map[string]*core.MappingNode)
	if newPolicies != nil {
		for _, policy := range newPolicies.Items {
			policyName := core.StringValue(policy.Fields["policyName"])
			newMap[policyName] = policy
		}
	}

	// Determine what needs to be added, updated, or removed
	for name, policy := range newMap {
		if currentPolicy, exists := currentMap[name]; !exists {
			u.toAdd = append(u.toAdd, policy)
		} else if !policiesEqual(currentPolicy, policy) {
			u.toUpdate = append(u.toUpdate, policy)
		}
	}

	for name := range currentMap {
		if _, exists := newMap[name]; !exists {
			u.toRemove = append(u.toRemove, name)
		}
	}

	return len(u.toAdd) > 0 || len(u.toUpdate) > 0 || len(u.toRemove) > 0, saveOpCtx, nil
}

func (u *userInlinePoliciesUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove policies
	for _, policyName := range u.toRemove {
		_, err := iamService.DeleteUserPolicy(ctx, &iam.DeleteUserPolicyInput{
			UserName:   aws.String(u.userName),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to delete inline policy %s: %w", policyName, err)
		}
	}

	// Add and update policies (both use PutUserPolicy)
	allPolicies := append(u.toAdd, u.toUpdate...)
	for _, policyNode := range allPolicies {
		policyName := core.StringValue(policyNode.Fields["policyName"])
		policyDocNode := policyNode.Fields["policyDocument"]

		policyJSON, err := json.Marshal(policyDocNode)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to marshal policy document: %w", err)
		}

		_, err = iamService.PutUserPolicy(ctx, &iam.PutUserPolicyInput{
			UserName:       aws.String(u.userName),
			PolicyName:     aws.String(policyName),
			PolicyDocument: aws.String(string(policyJSON)),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to put inline policy %s: %w", policyName, err)
		}
	}

	return saveOpCtx, nil
}

type userManagedPoliciesUpdate struct {
	userName string
	toAttach []string
	toDetach []string
}

func (u *userManagedPoliciesUpdate) Name() string {
	return "update managed policies"
}

func (u *userManagedPoliciesUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Compare current and desired managed policies
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentPolicies, _ := pluginutils.GetValueByPath("$.managedPolicyArns", currentStateSpecData)
	newPolicies := specData.Fields["managedPolicyArns"]

	currentSet := make(map[string]bool)
	if currentPolicies != nil {
		for _, policyArn := range currentPolicies.Items {
			currentSet[core.StringValue(policyArn)] = true
		}
	}

	newSet := make(map[string]bool)
	if newPolicies != nil {
		for _, policyArn := range newPolicies.Items {
			newSet[core.StringValue(policyArn)] = true
		}
	}

	// Determine policies to attach and detach
	for arn := range newSet {
		if !currentSet[arn] {
			u.toAttach = append(u.toAttach, arn)
		}
	}

	for arn := range currentSet {
		if !newSet[arn] {
			u.toDetach = append(u.toDetach, arn)
		}
	}

	return len(u.toAttach) > 0 || len(u.toDetach) > 0, saveOpCtx, nil
}

func (u *userManagedPoliciesUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Detach policies
	for _, policyArn := range u.toDetach {
		_, err := iamService.DetachUserPolicy(ctx, &iam.DetachUserPolicyInput{
			UserName:  aws.String(u.userName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to detach managed policy %s: %w", policyArn, err)
		}
	}

	// Attach policies
	for _, policyArn := range u.toAttach {
		_, err := iamService.AttachUserPolicy(ctx, &iam.AttachUserPolicyInput{
			UserName:  aws.String(u.userName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to attach managed policy %s: %w", policyArn, err)
		}
	}

	return saveOpCtx, nil
}

type userPermissionsBoundaryUpdate struct {
	userName  string
	operation string // "set", "delete"
	boundary  string
}

func (u *userPermissionsBoundaryUpdate) Name() string {
	return "update permissions boundary"
}

func (u *userPermissionsBoundaryUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentBoundary, _ := pluginutils.GetValueByPath("$.permissionsBoundary", currentStateSpecData)
	newBoundary := specData.Fields["permissionsBoundary"]

	currentValue := ""
	if currentBoundary != nil {
		currentValue = core.StringValue(currentBoundary)
	}

	newValue := ""
	if newBoundary != nil {
		newValue = core.StringValue(newBoundary)
	}

	if currentValue != newValue {
		if newValue != "" {
			u.operation = "set"
			u.boundary = newValue
		} else {
			u.operation = "delete"
		}
		return true, saveOpCtx, nil
	}

	return false, saveOpCtx, nil
}

func (u *userPermissionsBoundaryUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	switch u.operation {
	case "set":
		_, err := iamService.PutUserPermissionsBoundary(ctx, &iam.PutUserPermissionsBoundaryInput{
			UserName:            aws.String(u.userName),
			PermissionsBoundary: aws.String(u.boundary),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to set permissions boundary: %w", err)
		}

	case "delete":
		_, err := iamService.DeleteUserPermissionsBoundary(ctx, &iam.DeleteUserPermissionsBoundaryInput{
			UserName: aws.String(u.userName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to delete permissions boundary: %w", err)
		}
	}

	return saveOpCtx, nil
}

type userTagsUpdate struct {
	userName string
	toAdd    []types.Tag
	toRemove []string
}

func (u *userTagsUpdate) Name() string {
	return "update tags"
}

func (u *userTagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Compare current and desired tags
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentTags, _ := pluginutils.GetValueByPath("$.tags", currentStateSpecData)
	newTags := specData.Fields["tags"]

	currentMap := make(map[string]string)
	if currentTags != nil {
		for _, tag := range currentTags.Items {
			key := core.StringValue(tag.Fields["key"])
			value := core.StringValue(tag.Fields["value"])
			currentMap[key] = value
		}
	}

	newMap := make(map[string]string)
	if newTags != nil {
		for _, tag := range newTags.Items {
			key := core.StringValue(tag.Fields["key"])
			value := core.StringValue(tag.Fields["value"])
			newMap[key] = value
		}
	}

	// Determine tags to add/update and remove
	for key, value := range newMap {
		if currentValue, exists := currentMap[key]; !exists || currentValue != value {
			u.toAdd = append(u.toAdd, types.Tag{
				Key:   aws.String(key),
				Value: aws.String(value),
			})
		}
	}

	for key := range currentMap {
		if _, exists := newMap[key]; !exists {
			u.toRemove = append(u.toRemove, key)
		}
	}

	return len(u.toAdd) > 0 || len(u.toRemove) > 0, saveOpCtx, nil
}

func (u *userTagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove tags
	if len(u.toRemove) > 0 {
		_, err := iamService.UntagUser(ctx, &iam.UntagUserInput{
			UserName: aws.String(u.userName),
			TagKeys:  u.toRemove,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove tags: %w", err)
		}
	}

	// Add/update tags
	if len(u.toAdd) > 0 {
		_, err := iamService.TagUser(ctx, &iam.TagUserInput{
			UserName: aws.String(u.userName),
			Tags:     u.toAdd,
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add tags: %w", err)
		}
	}

	return saveOpCtx, nil
}

type userGroupMembershipUpdate struct {
	userName string
	toAdd    []string
	toRemove []string
}

func (u *userGroupMembershipUpdate) Name() string {
	return "update group membership"
}

func (u *userGroupMembershipUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Compare current and desired groups
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentGroups, _ := pluginutils.GetValueByPath("$.groups", currentStateSpecData)
	newGroups := specData.Fields["groups"]

	currentSet := make(map[string]bool)
	if currentGroups != nil {
		for _, group := range currentGroups.Items {
			currentSet[core.StringValue(group)] = true
		}
	}

	newSet := make(map[string]bool)
	if newGroups != nil {
		for _, group := range newGroups.Items {
			newSet[core.StringValue(group)] = true
		}
	}

	// Determine groups to add and remove
	for groupName := range newSet {
		if !currentSet[groupName] {
			u.toAdd = append(u.toAdd, groupName)
		}
	}

	for groupName := range currentSet {
		if !newSet[groupName] {
			u.toRemove = append(u.toRemove, groupName)
		}
	}

	return len(u.toAdd) > 0 || len(u.toRemove) > 0, saveOpCtx, nil
}

func (u *userGroupMembershipUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove from groups
	for _, groupName := range u.toRemove {
		_, err := iamService.RemoveUserFromGroup(ctx, &iam.RemoveUserFromGroupInput{
			UserName:  aws.String(u.userName),
			GroupName: aws.String(groupName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to remove user from group %s: %w", groupName, err)
		}
	}

	// Add to groups
	for _, groupName := range u.toAdd {
		_, err := iamService.AddUserToGroup(ctx, &iam.AddUserToGroupInput{
			UserName:  aws.String(u.userName),
			GroupName: aws.String(groupName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to add user to group %s: %w", groupName, err)
		}
	}

	return saveOpCtx, nil
}

// Helper function to compare policy documents.
func policiesEqual(policy1, policy2 *core.MappingNode) bool {
	// Simple comparison - in a real implementation, you might want to do a deep comparison
	// of the policy documents
	doc1JSON, err1 := json.Marshal(policy1.Fields["policyDocument"])
	doc2JSON, err2 := json.Marshal(policy2.Fields["policyDocument"])

	if err1 != nil || err2 != nil {
		return false
	}

	return string(doc1JSON) == string(doc2JSON)
}
