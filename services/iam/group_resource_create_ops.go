package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type groupCreate struct {
	input                    *iam.CreateGroupInput
	uniqueGroupNameGenerator utils.UniqueNameGenerator
}

func (g *groupCreate) Name() string {
	return "create group"
}

func (g *groupCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	createInput, hasValues, err := changesToCreateGroupInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}

	// Generate unique group name if not provided
	if createInput.GroupName == nil || aws.ToString(createInput.GroupName) == "" {
		generator := g.uniqueGroupNameGenerator
		if generator == nil {
			generator = utils.IAMGroupNameGenerator
		}
		inputData, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
		if !ok || inputData == nil {
			return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
		}
		uniqueGroupName, err := generator(inputData)
		if err != nil {
			return false, saveOpCtx, err
		}
		createInput.GroupName = aws.String(uniqueGroupName)
		hasValues = true
	}

	g.input = createInput
	return hasValues, saveOpCtx, nil
}

func (g *groupCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createGroupOutput, err := iamService.CreateGroup(ctx, g.input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create IAM group: %w", err)
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(createGroupOutput.Group.Arn)
	newSaveOpCtx.Data["createGroupOutput"] = createGroupOutput
	newSaveOpCtx.Data["groupArn"] = aws.ToString(createGroupOutput.Group.Arn)

	return newSaveOpCtx, nil
}

func changesToCreateGroupInput(
	specData *core.MappingNode,
) (*iam.CreateGroupInput, bool, error) {
	input := &iam.CreateGroupInput{}

	valueSetters := []*pluginutils.ValueSetter[*iam.CreateGroupInput]{
		pluginutils.NewValueSetter(
			"$.path",
			setCreateGroupPath,
		),
		pluginutils.NewValueSetter(
			"$.groupName",
			setCreateGroupName,
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates, nil
}

func setCreateGroupPath(
	value *core.MappingNode,
	input *iam.CreateGroupInput,
) {
	input.Path = aws.String(core.StringValue(value))
}

func setCreateGroupName(
	value *core.MappingNode,
	input *iam.CreateGroupInput,
) {
	input.GroupName = aws.String(core.StringValue(value))
}

func newGroupCreate(generator utils.UniqueNameGenerator) *groupCreate {
	return &groupCreate{
		uniqueGroupNameGenerator: generator,
	}
}

type groupInlinePoliciesCreate struct {
	policies []*core.MappingNode
}

func (g *groupInlinePoliciesCreate) Name() string {
	return "create inline policies"
}

func (g *groupInlinePoliciesCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are inline policies to create
	if policies, hasPolicies := pluginutils.GetValueByPath("$.policies", specData); hasPolicies && policies != nil && len(policies.Items) > 0 {
		g.policies = policies.Items
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (g *groupInlinePoliciesCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the group name from the created group
	createGroupOutput, ok := saveOpCtx.Data["createGroupOutput"].(*iam.CreateGroupOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createGroupOutput not found in save operation context")
	}

	groupName := aws.ToString(createGroupOutput.Group.GroupName)

	// Create each inline policy
	for _, policyNode := range g.policies {
		policyName, hasPolicyName := pluginutils.GetValueByPath("$.policyName", policyNode)
		if !hasPolicyName {
			continue
		}
		policyNameStr := core.StringValue(policyName)
		if policyNameStr == "" {
			continue
		}

		policyDocNode, hasPolicyDoc := pluginutils.GetValueByPath("$.policyDocument", policyNode)
		if !hasPolicyDoc {
			continue
		}

		policyJSON, err := json.Marshal(policyDocNode)
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to marshal inline policy document: %w", err)
		}

		_, err = iamService.PutGroupPolicy(ctx, &iam.PutGroupPolicyInput{
			GroupName:      aws.String(groupName),
			PolicyName:     aws.String(policyNameStr),
			PolicyDocument: aws.String(string(policyJSON)),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to put inline policy %s: %w", policyNameStr, err)
		}
	}

	return saveOpCtx, nil
}

type groupManagedPoliciesCreate struct {
	managedPolicyArns []*core.MappingNode
}

func (g *groupManagedPoliciesCreate) Name() string {
	return "create managed policies"
}

func (g *groupManagedPoliciesCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Check if there are managed policies to attach
	if managedPolicyArns, hasManagedPolicyArns := pluginutils.GetValueByPath("$.managedPolicyArns", specData); hasManagedPolicyArns && managedPolicyArns != nil && len(managedPolicyArns.Items) > 0 {
		g.managedPolicyArns = managedPolicyArns.Items
		return true, saveOpCtx, nil
	}
	return false, saveOpCtx, nil
}

func (g *groupManagedPoliciesCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	// Get the group name from the created group
	createGroupOutput, ok := saveOpCtx.Data["createGroupOutput"].(*iam.CreateGroupOutput)
	if !ok {
		return saveOpCtx, fmt.Errorf("createGroupOutput not found in save operation context")
	}

	groupName := aws.ToString(createGroupOutput.Group.GroupName)

	// Attach each managed policy
	for _, policyArnNode := range g.managedPolicyArns {
		policyArnStr := core.StringValue(policyArnNode)
		if policyArnStr == "" {
			continue
		}

		_, err := iamService.AttachGroupPolicy(ctx, &iam.AttachGroupPolicyInput{
			GroupName: aws.String(groupName),
			PolicyArn: aws.String(policyArnStr),
		})
		if err != nil {
			return saveOpCtx, fmt.Errorf("failed to attach managed policy %s: %w", policyArnStr, err)
		}
	}

	return saveOpCtx, nil
}
