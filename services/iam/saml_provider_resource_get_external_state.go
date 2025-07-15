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

func (i *iamSAMLProviderResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Safely get the SAML provider ARN from the resource spec
	arn, hasArn := pluginutils.GetValueByPath("$.arn", input.CurrentResourceSpec)
	if !hasArn {
		return nil, fmt.Errorf("SAML provider ARN is required for get external state")
	}

	arnStr := core.StringValue(arn)
	if arnStr == "" {
		return nil, fmt.Errorf("SAML provider ARN is required for get external state")
	}

	// Get the SAML provider details
	result, err := iamService.GetSAMLProvider(ctx, &iam.GetSAMLProviderInput{
		SAMLProviderArn: aws.String(arnStr),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get SAML provider: %w", err)
	}

	// Extract the name from the ARN
	name, err := extractNameFromArn(arnStr)
	if err != nil {
		return nil, fmt.Errorf("failed to extract name from ARN: %w", err)
	}

	// Build the external state
	externalState := map[string]*core.MappingNode{
		"arn":  core.MappingNodeFromString(arnStr),
		"name": core.MappingNodeFromString(name),
	}

	// Add SAML metadata document if present
	if result.SAMLMetadataDocument != nil {
		externalState["samlMetadataDocument"] = core.MappingNodeFromString(aws.ToString(result.SAMLMetadataDocument))
	}

	// Get tags
	tagsResult, err := iamService.ListSAMLProviderTags(ctx, &iam.ListSAMLProviderTagsInput{
		SAMLProviderArn: aws.String(arnStr),
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
