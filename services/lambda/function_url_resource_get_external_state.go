package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionUrlResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, fmt.Errorf("failed to get Lambda service: %w", err)
	}

	functionARN, hasFunctionARN := pluginutils.GetValueByPath(
		"$.functionArn",
		input.CurrentResourceSpec,
	)
	if !hasFunctionARN {
		return nil, fmt.Errorf("functionArn must be defined in the resource spec")
	}

	qualifier, _ := pluginutils.GetValueByPath(
		"$.qualifier",
		input.CurrentResourceSpec,
	)

	getFunctionUrlInput := &lambda.GetFunctionUrlConfigInput{
		FunctionName: aws.String(core.StringValue(functionARN)),
	}
	if qualifier != nil {
		getFunctionUrlInput.Qualifier = aws.String(core.StringValue(qualifier))
	}

	result, err := lambdaService.GetFunctionUrlConfig(ctx, getFunctionUrlInput)
	if err != nil {
		return nil, fmt.Errorf("failed to get function URL config: %w", err)
	}

	// Build resource spec state from AWS response
	resourceSpecState := l.buildBaseResourceSpecState(result)

	// Add optional fields if they exist
	err = l.addOptionalConfigurationsToSpec(result, resourceSpecState.Fields)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaFunctionUrlResourceActions) buildBaseResourceSpecState(
	output *lambda.GetFunctionUrlConfigOutput,
) *core.MappingNode {
	return &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"functionUrl": core.MappingNodeFromString(aws.ToString(output.FunctionUrl)),
			"functionArn": core.MappingNodeFromString(aws.ToString(output.FunctionArn)),
			"authType":    core.MappingNodeFromString(string(output.AuthType)),
		},
	}
}

func (l *lambdaFunctionUrlResourceActions) addOptionalConfigurationsToSpec(
	output *lambda.GetFunctionUrlConfigOutput,
	specFields map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetFunctionUrlConfigOutput]{
		{
			Name: "invokeMode",
			Condition: func(output *lambda.GetFunctionUrlConfigOutput) bool {
				return output.InvokeMode != ""
			},
			Fields: []string{"invokeMode"},
			Values: func(output *lambda.GetFunctionUrlConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					core.MappingNodeFromString(string(output.InvokeMode)),
				}, nil
			},
		},
		{
			Name: "cors",
			Condition: func(output *lambda.GetFunctionUrlConfigOutput) bool {
				return output.Cors != nil
			},
			Fields: []string{"cors"},
			Values: func(output *lambda.GetFunctionUrlConfigOutput) ([]*core.MappingNode, error) {
				return []*core.MappingNode{
					functionUrlCorsToMappingNode(output.Cors),
				}, nil
			},
		},
	}

	return pluginutils.RunOptionalValueExtractors(
		output,
		specFields,
		extractors,
	)
}

func functionUrlCorsToMappingNode(
	cors *types.Cors,
) *core.MappingNode {
	if cors == nil {
		return &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}

	fields := map[string]*core.MappingNode{}

	if cors.AllowCredentials != nil {
		fields["allowCredentials"] = core.MappingNodeFromBool(aws.ToBool(cors.AllowCredentials))
	}

	if len(cors.AllowHeaders) > 0 {
		headers := make([]*core.MappingNode, len(cors.AllowHeaders))
		for i, header := range cors.AllowHeaders {
			headers[i] = core.MappingNodeFromString(header)
		}
		fields["allowHeaders"] = &core.MappingNode{Items: headers}
	}

	if len(cors.AllowMethods) > 0 {
		methods := make([]*core.MappingNode, len(cors.AllowMethods))
		for i, method := range cors.AllowMethods {
			methods[i] = core.MappingNodeFromString(method)
		}
		fields["allowMethods"] = &core.MappingNode{Items: methods}
	}

	if len(cors.AllowOrigins) > 0 {
		origins := make([]*core.MappingNode, len(cors.AllowOrigins))
		for i, origin := range cors.AllowOrigins {
			origins[i] = core.MappingNodeFromString(origin)
		}
		fields["allowOrigins"] = &core.MappingNode{Items: origins}
	}

	if len(cors.ExposeHeaders) > 0 {
		headers := make([]*core.MappingNode, len(cors.ExposeHeaders))
		for i, header := range cors.ExposeHeaders {
			headers[i] = core.MappingNodeFromString(header)
		}
		fields["exposeHeaders"] = &core.MappingNode{Items: headers}
	}

	if cors.MaxAge != nil {
		fields["maxAge"] = core.MappingNodeFromInt(int(aws.ToInt32(cors.MaxAge)))
	}

	return &core.MappingNode{Fields: fields}
}
