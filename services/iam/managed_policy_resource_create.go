package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamManagedPolicyResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[iamservice.Service]{
		newManagedPolicyCreate(i.uniqueNameGenerator),
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
		return nil, fmt.Errorf("no updates were made during managed policy creation")
	}

	createPolicyOutput, ok := saveOpCtx.Data["createPolicyOutput"].(*iam.CreatePolicyOutput)
	if !ok {
		return nil, fmt.Errorf("createPolicyOutput not found in save operation context")
	}

	computedFields := map[string]*core.MappingNode{
		"spec.arn":                           core.MappingNodeFromString(aws.ToString(createPolicyOutput.Policy.Arn)),
		"spec.id":                            core.MappingNodeFromString(aws.ToString(createPolicyOutput.Policy.PolicyId)),
		"spec.attachmentCount":               core.MappingNodeFromInt(int(aws.ToInt32(createPolicyOutput.Policy.AttachmentCount))),
		"spec.createDate":                    core.MappingNodeFromString(createPolicyOutput.Policy.CreateDate.Format("2006-01-02T15:04:05Z")),
		"spec.defaultVersionId":              core.MappingNodeFromString(aws.ToString(createPolicyOutput.Policy.DefaultVersionId)),
		"spec.isAttachable":                  core.MappingNodeFromBool(createPolicyOutput.Policy.IsAttachable),
		"spec.permissionsBoundaryUsageCount": core.MappingNodeFromInt(int(aws.ToInt32(createPolicyOutput.Policy.PermissionsBoundaryUsageCount))),
		"spec.updateDate":                    core.MappingNodeFromString(createPolicyOutput.Policy.UpdateDate.Format("2006-01-02T15:04:05Z")),
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: computedFields,
	}, nil
}

func changesToCreatePolicyInput(
	specData *core.MappingNode,
) (*iam.CreatePolicyInput, bool, error) {
	input := &iam.CreatePolicyInput{}

	valueSetters := []*pluginutils.ValueSetter[*iam.CreatePolicyInput]{
		pluginutils.NewValueSetter(
			"$.policyDocument",
			setCreatePolicyDocument,
		),
		pluginutils.NewValueSetter(
			"$.description",
			setCreatePolicyDescription,
		),
		pluginutils.NewValueSetter(
			"$.path",
			setCreatePolicyPath,
		),
		pluginutils.NewValueSetter(
			"$.policyName",
			setCreatePolicyName,
		),
		pluginutils.NewValueSetter(
			"$.tags",
			setCreatePolicyTags,
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates, nil
}

func setCreatePolicyDocument(
	value *core.MappingNode,
	input *iam.CreatePolicyInput,
) {
	// Convert the structured policy document to JSON string
	policyJSON, err := json.Marshal(value)
	if err != nil {
		// Fallback to string value if JSON marshaling fails
		input.PolicyDocument = aws.String(core.StringValue(value))
		return
	}
	input.PolicyDocument = aws.String(string(policyJSON))
}

func setCreatePolicyDescription(
	value *core.MappingNode,
	input *iam.CreatePolicyInput,
) {
	input.Description = aws.String(core.StringValue(value))
}

func setCreatePolicyPath(
	value *core.MappingNode,
	input *iam.CreatePolicyInput,
) {
	input.Path = aws.String(core.StringValue(value))
}

func setCreatePolicyName(
	value *core.MappingNode,
	input *iam.CreatePolicyInput,
) {
	input.PolicyName = aws.String(core.StringValue(value))
}

func setCreatePolicyTags(
	value *core.MappingNode,
	input *iam.CreatePolicyInput,
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
	sort.Slice(tags, func(i, j int) bool {
		return aws.ToString(tags[i].Key) < aws.ToString(tags[j].Key)
	})
	input.Tags = tags
}
