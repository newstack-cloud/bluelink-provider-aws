package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (i *iamOIDCProviderResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Safely get the OIDC provider ARN from the resource spec
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.CurrentResourceSpec)
	if !hasArn {
		return nil, fmt.Errorf("OIDC provider ARN is required for get external state")
	}

	arnStr := core.StringValue(arn)
	if arnStr == "" {
		return nil, fmt.Errorf("OIDC provider ARN is required for get external state")
	}

	// Get the OIDC provider details
	result, err := iamService.GetOpenIDConnectProvider(ctx, &iam.GetOpenIDConnectProviderInput{
		OpenIDConnectProviderArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get OIDC provider: %w", err)
	}

	// Extract the URL from the ARN
	url, err := extractUrlFromArn(arnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract URL from ARN: %w", err)
	}

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"arn": core.MappingNodeFromString(arnStr),
		"url": core.MappingNodeFromString(url),
	}

	// Add client IDs if present
	if len(result.ClientIDList) > 0 {
		clientIdItems := make([]*core.MappingNode, len(result.ClientIDList))
		for i, clientId := range result.ClientIDList {
			clientIdItems[i] = core.MappingNodeFromString(clientId)
		}
		externalState["clientIdList"] = &core.MappingNode{
			Items: clientIdItems,
		}
	}

	// Add thumbprints if present
	if len(result.ThumbprintList) > 0 {
		thumbprintItems := make([]*core.MappingNode, len(result.ThumbprintList))
		for i, thumbprint := range result.ThumbprintList {
			thumbprintItems[i] = core.MappingNodeFromString(thumbprint)
		}
		externalState["thumbprintList"] = &core.MappingNode{
			Items: thumbprintItems,
		}
	}

	// Get tags
	tagsResult, err := iamService.ListOpenIDConnectProviderTags(ctx, &iam.ListOpenIDConnectProviderTagsInput{
		OpenIDConnectProviderArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	if len(tagsResult.Tags) > 0 {
		externalState["tags"] = extractIAMTags(tagsResult.Tags)
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: externalState,
		},
	}, nil
}
