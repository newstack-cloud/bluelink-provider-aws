package lambda

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaLayerVersionResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	// Get the layer name and version from the current resource spec
	layerName := core.StringValue(input.CurrentResourceSpec.Fields["layerName"])
	versionValue := input.CurrentResourceSpec.Fields["version"]
	var versionNumber int64
	if versionValue != nil {
		if versionValue.Scalar != nil && versionValue.Scalar.IntValue != nil {
			versionNumber = int64(*versionValue.Scalar.IntValue)
		}
	}

	getLayerVersionInput := &lambda.GetLayerVersionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	}

	getLayerVersionOutput, err := lambdaService.GetLayerVersion(ctx, getLayerVersionInput)
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			// If the layer version doesn't exist, return empty state (this might indicate external deletion)
			if apiError.ErrorCode() == "ResourceNotFoundException" {
				return &provider.ResourceGetExternalStateOutput{
					ResourceSpecState: &core.MappingNode{Fields: make(map[string]*core.MappingNode)},
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to get layer version: %w", err)
	}

	// Build the resource spec state from the API response
	resourceSpecState := l.buildBaseResourceSpecState(getLayerVersionOutput)

	// Add optional fields using pluginutils extractors
	err = l.addOptionalConfigurationsToSpec(getLayerVersionOutput, resourceSpecState.Fields)
	if err != nil {
		return nil, err
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

func (l *lambdaLayerVersionResourceActions) buildBaseResourceSpecState(
	output *lambda.GetLayerVersionOutput,
) *core.MappingNode {
	fields := make(map[string]*core.MappingNode)

	// Add required fields
	if output.LayerArn != nil {
		fields["layerArn"] = core.MappingNodeFromString(*output.LayerArn)
	}
	if output.LayerVersionArn != nil {
		fields["layerVersionArn"] = core.MappingNodeFromString(*output.LayerVersionArn)
	}
	fields["version"] = core.MappingNodeFromInt(int(output.Version))

	return &core.MappingNode{Fields: fields}
}

func (l *lambdaLayerVersionResourceActions) addOptionalConfigurationsToSpec(
	output *lambda.GetLayerVersionOutput,
	specFields map[string]*core.MappingNode,
) error {
	extractors := []pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput]{
		layerVersionDescriptionValueExtractor(),
		layerVersionLicenseInfoValueExtractor(),
		layerVersionCreatedDateValueExtractor(),
		layerVersionCompatibleRuntimesValueExtractor(),
		layerVersionCompatibleArchitecturesValueExtractor(),
		layerVersionContentValueExtractor(),
	}

	return pluginutils.RunOptionalValueExtractors(output, specFields, extractors)
}

// parseLayerVersionArn extracts the layer name and version number from a layer version ARN
// Format: arn:aws:lambda:region:account-id:layer:layer-name:version.
func parseLayerVersionArn(arn string) (layerName string, versionNumber int64, err error) {
	parts := strings.Split(arn, ":")
	if len(parts) != 8 || parts[0] != "arn" || parts[1] != "aws" || parts[2] != "lambda" || parts[5] != "layer" {
		return "", 0, fmt.Errorf("invalid layer version ARN format: %s", arn)
	}

	layerName = parts[6]
	versionNumber, err = strconv.ParseInt(parts[7], 10, 64)
	if err != nil {
		return "", 0, fmt.Errorf("invalid version number in ARN %s: %w", arn, err)
	}

	return layerName, versionNumber, nil
}
