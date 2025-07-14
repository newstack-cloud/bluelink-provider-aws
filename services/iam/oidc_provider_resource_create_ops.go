package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type oidcProviderCreate struct {
	url                        string
	clientIdList               []string
	thumbprintList             []string
	uniqueOidcProviderUrlGenerator utils.UniqueNameGenerator
}

func (o *oidcProviderCreate) Name() string {
	return "create OIDC provider"
}

func (o *oidcProviderCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Extract URL from spec data
	url, hasUrl := pluginutils.GetValueByPath("$.url", specData)
	if !hasUrl || core.StringValue(url) == "" {
		// If no URL is provided, generate a unique one
		generator := o.uniqueOidcProviderUrlGenerator
		if generator == nil {
					// Use the default OIDC provider generator
		generator = utils.IAMOidcProviderUrlGenerator
		}

		// Retrieve inputData from SaveOperationContext if available
		inputData, ok := saveOpCtx.Data["ResourceDeployInput"].(*provider.ResourceDeployInput)
		if !ok || inputData == nil {
			return false, saveOpCtx, fmt.Errorf("ResourceDeployInput not found in SaveOperationContext.Data")
		}

		uniqueId, err := generator(inputData)
		if err != nil {
			return false, saveOpCtx, err
		}

		o.url = fmt.Sprintf("https://%s.example.com", uniqueId)
	} else {
		o.url = core.StringValue(url)
	}

	// Extract client IDs
	clientIdList, hasClientIds := pluginutils.GetValueByPath("$.clientIdList", specData)
	if hasClientIds && clientIdList != nil && len(clientIdList.Items) > 0 {
		clientIdItems := clientIdList.Items
		o.clientIdList = make([]string, len(clientIdItems))
		for i, item := range clientIdItems {
			o.clientIdList[i] = core.StringValue(item)
		}
	}

	// Extract thumbprints
	thumbprintList, hasThumbprints := pluginutils.GetValueByPath("$.thumbprintList", specData)
	if hasThumbprints && thumbprintList != nil && len(thumbprintList.Items) > 0 {
		thumbprintItems := thumbprintList.Items
		o.thumbprintList = make([]string, len(thumbprintItems))
		for i, item := range thumbprintItems {
			o.thumbprintList[i] = core.StringValue(item)
		}
	}

	return true, saveOpCtx, nil
}

func (o *oidcProviderCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	input := &iam.CreateOpenIDConnectProviderInput{
		Url:            aws.String(o.url),
		ClientIDList:   o.clientIdList,
		ThumbprintList: o.thumbprintList,
	}

	output, err := iamService.CreateOpenIDConnectProvider(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	newSaveOpCtx.Data["createOidcProviderOutput"] = output
	newSaveOpCtx.Data["oidcProviderArn"] = output.OpenIDConnectProviderArn
	return newSaveOpCtx, nil
}

func newOidcProviderCreate(generator utils.UniqueNameGenerator) *oidcProviderCreate {
	return &oidcProviderCreate{
		uniqueOidcProviderUrlGenerator: generator,
	}
}

type oidcProviderTagsSave struct {
	arn  string
	tags []types.Tag
}

func (o *oidcProviderTagsSave) Name() string {
	return "save OIDC provider tags"
}

func (o *oidcProviderTagsSave) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	// Get the ARN from the create operation output
	arn, ok := saveOpCtx.Data["oidcProviderArn"].(*string)
	if !ok {
		return false, saveOpCtx, fmt.Errorf("OIDC provider ARN not found")
	}
	o.arn = aws.ToString(arn)

	// Extract tags from spec data
	tags, hasTags := pluginutils.GetValueByPath("$.tags", specData)
	if !hasTags || tags == nil || len(tags.Items) == 0 {
		return false, saveOpCtx, nil // No tags to save
	}

	tagItems := tags.Items
	o.tags = make([]types.Tag, len(tagItems))
	for i, item := range tagItems {
		keyNode, hasKey := pluginutils.GetValueByPath("$.key", item)
		valueNode, hasValue := pluginutils.GetValueByPath("$.value", item)
		if !hasKey || !hasValue {
			return false, saveOpCtx, fmt.Errorf("invalid tag format at index %d", i)
		}
		o.tags[i] = types.Tag{
			Key:   aws.String(core.StringValue(keyNode)),
			Value: aws.String(core.StringValue(valueNode)),
		}
	}

	return len(o.tags) > 0, saveOpCtx, nil
}

func (o *oidcProviderTagsSave) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	iamService iamservice.Service,
) (pluginutils.SaveOperationContext, error) {
	input := &iam.TagOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(o.arn),
		Tags:                     o.tags,
	}

	_, err := iamService.TagOpenIDConnectProvider(ctx, input)
	if err != nil {
		return saveOpCtx, fmt.Errorf("failed to tag OIDC provider: %w", err)
	}

	return saveOpCtx, nil
}