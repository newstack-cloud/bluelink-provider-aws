package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaAliasResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	// Extract function name and alias name from the current resource spec
	functionName := core.StringValue(input.CurrentResourceSpec.Fields["functionName"])
	aliasName := core.StringValue(input.CurrentResourceSpec.Fields["name"])

	getAliasInput := &lambda.GetAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(aliasName),
	}

	result, err := lambdaService.GetAlias(ctx, getAliasInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda alias: %w", err)
	}

	// Build resource spec state from AWS response
	resourceSpecState := l.buildBaseResourceSpecState(result, functionName, aliasName)

	// Add optional fields if they exist
	err = l.addOptionalConfigurationsToSpec(result, resourceSpecState.Fields)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaAliasResourceActions) buildBaseResourceSpecState(
	output *lambda.GetAliasOutput,
	functionName string,
	aliasName string,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString(functionName),
			"name":            core.MappingNodeFromString(aliasName),
			"functionVersion": core.MappingNodeFromString(aws.ToString(output.FunctionVersion)),
			"aliasArn":        core.MappingNodeFromString(aws.ToString(output.AliasArn)),
		},
	}
}

func (l *lambdaAliasResourceActions) addOptionalConfigurationsToSpec(
	output *lambda.GetAliasOutput,
	specFields map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetAliasOutput]{
		{
			Name: "description",
			Condition: func(output *lambda.GetAliasOutput) bool {
				return output.Description != nil
			},
			Fields: []string{"description"},
			Values: func(output *lambda.GetAliasOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(aws.ToString(output.Description)),
				}, nil
			},
		},
		{
			Name: "routingConfig",
			Condition: func(output *lambda.GetAliasOutput) bool {
				return output.RoutingConfig != nil && output.RoutingConfig.AdditionalVersionWeights != nil && len(output.RoutingConfig.AdditionalVersionWeights) > 0
			},
			Fields: []string{"routingConfig"},
			Values: func(output *lambda.GetAliasOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"additionalVersionWeights": buildAdditionalVersionWeightsNode(output.RoutingConfig.AdditionalVersionWeights),
						},
					},
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(output, specFields, extractors)
}

func buildAdditionalVersionWeightsNode(weights map[string]float64) *core.MappingNode {
	weightsMap := &core.MappingNode{
		Fields: make(map[string]*core.MappingNode),
	}
	for version, weight := range weights {
		weightsMap.Fields[version] = core.MappingNodeFromFloat(weight)
	}
	return weightsMap
}
