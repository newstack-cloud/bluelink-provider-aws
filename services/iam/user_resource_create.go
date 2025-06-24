package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
)

func (i *iamUserResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[iamservice.Service]{
		newUserCreate(i.uniqueNameGenerator),
		&userLoginProfileCreate{},
		&userInlinePoliciesCreate{},
		&userManagedPoliciesCreate{},
		&userPermissionsBoundaryCreate{},
		&userGroupMembershipCreate{},
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
		return nil, fmt.Errorf("no updates were made during user creation")
	}

	createUserOutput, ok := saveOpCtx.Data["createUserOutput"].(*iam.CreateUserOutput)
	if !ok {
		return nil, fmt.Errorf("createUserOutput not found in save operation context")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn":    core.MappingNodeFromString(aws.ToString(createUserOutput.User.Arn)),
		"spec.userId": core.MappingNodeFromString(aws.ToString(createUserOutput.User.UserId)),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}

func changesToCreateUserInput(
	specData *core.MappingNode,
) (*iam.CreateUserInput, bool, error) {
	input := &iam.CreateUserInput{}

	valueSetters := []*pluginutils.ValueSetter[*iam.CreateUserInput]{
		pluginutils.NewValueSetter(
			"$.path",
			setCreateUserPath,
		),
		pluginutils.NewValueSetter(
			"$.userName",
			setCreateUserName,
		),
		pluginutils.NewValueSetter(
			"$.tags",
			setCreateUserTags,
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates, nil
}

func setCreateUserPath(
	value *core.MappingNode,
	input *iam.CreateUserInput,
) {
	input.Path = aws.String(core.StringValue(value))
}

func setCreateUserName(
	value *core.MappingNode,
	input *iam.CreateUserInput,
) {
	input.UserName = aws.String(core.StringValue(value))
}

func setCreateUserTags(
	value *core.MappingNode,
	input *iam.CreateUserInput,
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
	input.Tags = tags
}
