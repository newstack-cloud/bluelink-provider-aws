package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type groupUpdate struct {
	groupName string
	path      string
}

func (g *groupUpdate) Name() string {
	return "update group"
}

func (g *groupUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Extract group name from ARN in the current state
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	if currentStateSpecData == nil {
		return false, saveOpCtx, fmt.Errorf("current state spec data is required for group update")
	}

	arn, hasArn := pluginutils.GetValueByPath("$.arn", currentStateSpecData)
	if !hasArn {
		return false, saveOpCtx, fmt.Errorf("ARN is required for group update")
	}

	arnStr := core.StringValue(arn)
	if arnStr == "" {
		return false, saveOpCtx, fmt.Errorf("ARN is required for group update")
	}

	groupName, err := extractGroupNameFromARN(arnStr)
	if err != nil {
		return false, saveOpCtx, fmt.Errorf("failed to extract group name from ARN: %w", err)
	}

	g.groupName = groupName

	// Check if path needs to be updated
	if path, hasPath := pluginutils.GetValueByPath("$.path", specData); hasPath {
		g.path = core.StringValue(path)
		return true, saveOpCtx, nil
	}

	return false, saveOpCtx, nil
}

func (g *groupUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := iamService.UpdateGroup(ctx, &iam.UpdateGroupInput{
		GroupName: aws.String(g.groupName),
		NewPath:   aws.String(g.path),
	})
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to update group: %w", err)
	}

	return saveOpCtx, nil
}

type groupInlinePoliciesUpdate struct {
	groupName        string
	policiesToAdd    []*core.MappingNode
	policiesToRemove []string
}

func (g *groupInlinePoliciesUpdate) Name() string {
	return "update inline policies"
}

func (g *groupInlinePoliciesUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Compare current and desired inline policies
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentPolicies, _ := pluginutils.GetValueByPath("$.policies", currentStateSpecData)
	newPolicies, _ := pluginutils.GetValueByPath("$.policies", specData)

	// Create maps for easier comparison
	currentMap := make(map[string]*core.MappingNode)
	if currentPolicies != nil {
		for _, policy := range currentPolicies.Items {
			if policyName, hasPolicyName := pluginutils.GetValueByPath("$.policyName", policy); hasPolicyName {
				policyNameStr := core.StringValue(policyName)
				if policyNameStr != "" {
					currentMap[policyNameStr] = policy
				}
			}
		}
	}

	newMap := make(map[string]*core.MappingNode)
	if newPolicies != nil {
		for _, policy := range newPolicies.Items {
			if policyName, hasPolicyName := pluginutils.GetValueByPath("$.policyName", policy); hasPolicyName {
				policyNameStr := core.StringValue(policyName)
				if policyNameStr != "" {
					newMap[policyNameStr] = policy
				}
			}
		}
	}

	// Determine what needs to be added, updated, or removed
	var policiesToAdd []*core.MappingNode
	var policiesToRemove []string

	for policyName, policy := range newMap {
		if _, exists := currentMap[policyName]; !exists {
			policiesToAdd = append(policiesToAdd, policy)
		}
	}

	for policyName := range currentMap {
		if _, exists := newMap[policyName]; !exists {
			policiesToRemove = append(policiesToRemove, policyName)
		}
	}

	g.policiesToAdd = policiesToAdd
	g.policiesToRemove = policiesToRemove

	return len(policiesToAdd) > 0 || len(policiesToRemove) > 0, saveOpCtx, nil
}

func (g *groupInlinePoliciesUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove policies
	for _, policyName := range g.policiesToRemove {
		_, err := iamService.DeleteGroupPolicy(ctx, &iam.DeleteGroupPolicyInput{
			GroupName:  aws.String(g.groupName),
			PolicyName: aws.String(policyName),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to delete group policy %s: %w", policyName, err)
		}
	}

	// Add policies
	for _, policy := range g.policiesToAdd {
		policyName, hasPolicyName := pluginutils.GetValueByPath("$.policyName", policy)
		if !hasPolicyName {
			continue
		}
		policyNameStr := core.StringValue(policyName)
		if policyNameStr == "" {
			continue
		}

		policyDocNode, hasPolicyDoc := pluginutils.GetValueByPath("$.policyDocument", policy)
		if !hasPolicyDoc {
			continue
		}

		policyJSON, err := json.Marshal(policyDocNode)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to marshal inline policy document: %w", err)
		}

		_, err = iamService.PutGroupPolicy(ctx, &iam.PutGroupPolicyInput{
			GroupName:      aws.String(g.groupName),
			PolicyName:     aws.String(policyNameStr),
			PolicyDocument: aws.String(string(policyJSON)),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to put inline policy %s: %w", policyNameStr, err)
		}
	}

	return saveOpCtx, nil
}

type groupManagedPoliciesUpdate struct {
	groupName        string
	policiesToAdd    []string
	policiesToRemove []string
}

func (g *groupManagedPoliciesUpdate) Name() string {
	return "update managed policies"
}

func (g *groupManagedPoliciesUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Compare current and desired managed policies
	currentStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	currentPolicies, _ := pluginutils.GetValueByPath("$.managedPolicyArns", currentStateSpecData)
	newPolicies, _ := pluginutils.GetValueByPath("$.managedPolicyArns", specData)

	// Create maps for easier comparison
	currentMap := make(map[string]bool)
	if currentPolicies != nil {
		for _, policyArn := range currentPolicies.Items {
			policyArnStr := core.StringValue(policyArn)
			if policyArnStr != "" {
				currentMap[policyArnStr] = true
			}
		}
	}

	newMap := make(map[string]bool)
	if newPolicies != nil {
		for _, policyArn := range newPolicies.Items {
			policyArnStr := core.StringValue(policyArn)
			if policyArnStr != "" {
				newMap[policyArnStr] = true
			}
		}
	}

	// Determine what needs to be added or removed
	var policiesToAdd []string
	var policiesToRemove []string

	for policyArn := range newMap {
		if !currentMap[policyArn] {
			policiesToAdd = append(policiesToAdd, policyArn)
		}
	}

	for policyArn := range currentMap {
		if !newMap[policyArn] {
			policiesToRemove = append(policiesToRemove, policyArn)
		}
	}

	g.policiesToAdd = policiesToAdd
	g.policiesToRemove = policiesToRemove

	return len(policiesToAdd) > 0 || len(policiesToRemove) > 0, saveOpCtx, nil
}

func (g *groupManagedPoliciesUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Remove policies
	for _, policyArn := range g.policiesToRemove {
		_, err := iamService.DetachGroupPolicy(ctx, &iam.DetachGroupPolicyInput{
			GroupName: aws.String(g.groupName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to detach group policy %s: %w", policyArn, err)
		}
	}

	// Add policies
	for _, policyArn := range g.policiesToAdd {
		_, err := iamService.AttachGroupPolicy(ctx, &iam.AttachGroupPolicyInput{
			GroupName: aws.String(g.groupName),
			PolicyArn: aws.String(policyArn),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to attach group policy %s: %w", policyArn, err)
		}
	}

	return saveOpCtx, nil
}
