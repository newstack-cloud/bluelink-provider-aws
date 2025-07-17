package iam

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamRoleResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[iamservice.Service]{
		newRoleCreate(i.uniqueNameGenerator),
		&roleInlinePoliciesCreate{},
		&roleManagedPoliciesCreate{},
		&rolePermissionsBoundaryCreate{},
	}

	saveOpCtx := pluginutils.SaveOperationContext{
		Data: map[string]any{
			"ResourceDeployInput": input,
		},
	}

	hasUpdates, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		saveOpCtx,
		createOperations,
		input,
		iamService,
	)
	if err != nil {
		return nil, err
	}

	if !hasUpdates {
		return nil, fmt.Errorf("no updates were made during role creation")
	}

	createRoleOutput, ok := saveOpCtx.Data["createRoleOutput"].(*iam.CreateRoleOutput)
	if !ok {
		return nil, fmt.Errorf("createRoleOutput not found in save operation context")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn":    core.MappingNodeFromString(aws.ToString(createRoleOutput.Role.Arn)),
		"spec.roleId": core.MappingNodeFromString(aws.ToString(createRoleOutput.Role.RoleId)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}

func changesToCreateRoleInput(
	specData *core.MappingNode,
) (*iam.CreateRoleInput, bool, error) {
	input := &iam.CreateRoleInput{}

	valueSetters := []*pluginutils.ValueSetter[*iam.CreateRoleInput]{
		pluginutils.NewValueSetter(
			"$.assumeRolePolicyDocument",
			setCreateRoleAssumeRolePolicyDocument,
		),
		pluginutils.NewValueSetter(
			"$.description",
			setCreateRoleDescription,
		),
		pluginutils.NewValueSetter(
			"$.maxSessionDuration",
			setCreateRoleMaxSessionDuration,
		),
		pluginutils.NewValueSetter(
			"$.path",
			setCreateRolePath,
		),
		pluginutils.NewValueSetter(
			"$.roleName",
			setCreateRoleName,
		),
		pluginutils.NewValueSetter(
			"$.tags",
			setCreateRoleTags,
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates, nil
}

func setCreateRoleAssumeRolePolicyDocument(
	value *core.MappingNode,
	input *iam.CreateRoleInput,
) {
	// Convert the structured policy document to JSON string
	policyJSON, err := json.Marshal(value)
	if err != nil {
		// Fallback to string value if JSON marshaling fails
		input.AssumeRolePolicyDocument = aws.String(core.StringValue(value))
		return
	}
	input.AssumeRolePolicyDocument = aws.String(string(policyJSON))
}

func setCreateRoleDescription(
	value *core.MappingNode,
	input *iam.CreateRoleInput,
) {
	input.Description = aws.String(core.StringValue(value))
}

func setCreateRoleMaxSessionDuration(
	value *core.MappingNode,
	input *iam.CreateRoleInput,
) {
	input.MaxSessionDuration = aws.Int32(int32(core.IntValue(value)))
}

func setCreateRolePath(
	value *core.MappingNode,
	input *iam.CreateRoleInput,
) {
	input.Path = aws.String(core.StringValue(value))
}

func setCreateRoleName(
	value *core.MappingNode,
	input *iam.CreateRoleInput,
) {
	input.RoleName = aws.String(core.StringValue(value))
}

func setCreateRoleTags(
	value *core.MappingNode,
	input *iam.CreateRoleInput,
) {
	tags := make([]types.Tag, 0, len(value.Items))
	for _, item := range value.Items {
		key := core.StringValue(item.Fields["key"])
		tagValue := core.StringValue(item.Fields["value"])
		tags = append(tags, types.Tag{
			Key:   aws.String(key),
			Value: aws.String(tagValue),
		})
	}
	// Sort tags by key before setting them
	input.Tags = sortTagsByKey(tags)
}
