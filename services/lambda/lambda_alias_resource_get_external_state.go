package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
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
	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionName":    core.MappingNodeFromString(functionName),
			"name":            core.MappingNodeFromString(aliasName),
			"functionVersion": core.MappingNodeFromString(aws.ToString(result.FunctionVersion)),
			"aliasArn":        core.MappingNodeFromString(aws.ToString(result.AliasArn)),
		},
	}

	if result.Description != nil {
		resourceSpecState.Fields["description"] = core.MappingNodeFromString(aws.ToString(result.Description))
	}

	if result.RoutingConfig != nil && result.RoutingConfig.AdditionalVersionWeights != nil && len(result.RoutingConfig.AdditionalVersionWeights) > 0 {
		routingConfig := &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"additionalVersionWeights": buildAdditionalVersionWeightsNode(result.RoutingConfig.AdditionalVersionWeights),
			},
		}
		resourceSpecState.Fields["routingConfig"] = routingConfig
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
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
