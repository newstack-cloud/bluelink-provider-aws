package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaCodeSigningConfigResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	codeSigningConfigArn := core.StringValue(
		input.CurrentResourceSpec.Fields["codeSigningConfigArn"],
	)

	getCodeSigningConfigInput := &lambda.GetCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(codeSigningConfigArn),
	}

	result, err := lambdaService.GetCodeSigningConfig(ctx, getCodeSigningConfigInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda code signing config: %w", err)
	}

	// Build resource spec state from AWS response
	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"codeSigningConfigArn": core.MappingNodeFromString(
				aws.ToString(result.CodeSigningConfig.CodeSigningConfigArn),
			),
			"codeSigningConfigId": core.MappingNodeFromString(
				aws.ToString(result.CodeSigningConfig.CodeSigningConfigId),
			),
		},
	}

	// Build allowed publishers
	if result.CodeSigningConfig.AllowedPublishers != nil {
		allowedPublishers := &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"signingProfileVersionArns": buildSigningProfileVersionArnsNode(
					result.CodeSigningConfig.AllowedPublishers.SigningProfileVersionArns,
				),
			},
		}
		resourceSpecState.Fields["allowedPublishers"] = allowedPublishers
	}

	// Build code signing policies if present
	if result.CodeSigningConfig.CodeSigningPolicies != nil {
		codeSigningPolicies := &core.MappingNode{
			Fields: map[string]*core.MappingNode{},
		}
		if result.CodeSigningConfig.CodeSigningPolicies.UntrustedArtifactOnDeployment != "" {
			codeSigningPolicies.Fields["untrustedArtifactOnDeployment"] = core.MappingNodeFromString(
				string(result.CodeSigningConfig.CodeSigningPolicies.UntrustedArtifactOnDeployment),
			)
		}
		resourceSpecState.Fields["codeSigningPolicies"] = codeSigningPolicies
	}

	// Add description if present
	if result.CodeSigningConfig.Description != nil {
		resourceSpecState.Fields["description"] = core.MappingNodeFromString(aws.ToString(result.CodeSigningConfig.Description))
	}

	// Get tags
	listTagsInput := &lambda.ListTagsInput{
		Resource: aws.String(codeSigningConfigArn),
	}
	tagsOutput, err := lambdaService.ListTags(ctx, listTagsInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda code signing config tags: %w", err)
	}

	// Add tags if present
	if tagsOutput != nil && len(tagsOutput.Tags) > 0 {
		tagNodes := make([]*core.MappingNode, 0, len(tagsOutput.Tags))
		for key, value := range tagsOutput.Tags {
			tagNodes = append(tagNodes, &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"key":   core.MappingNodeFromString(key),
					"value": core.MappingNodeFromString(value),
				},
			})
		}
		resourceSpecState.Fields["tags"] = &core.MappingNode{
			Items: tagNodes,
		}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func buildSigningProfileVersionArnsNode(arns []string) *core.MappingNode {
	arnNodes := make([]*core.MappingNode, len(arns))
	for i, arn := range arns {
		arnNodes[i] = core.MappingNodeFromString(arn)
	}
	return &core.MappingNode{
		Items: arnNodes,
	}
}
