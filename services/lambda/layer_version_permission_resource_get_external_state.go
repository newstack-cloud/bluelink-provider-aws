package lambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *lambdaLayerVersionPermissionResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	layerVersionArn := core.StringValue(input.CurrentResourceSpec.Fields["layerVersionArn"])
	statementId := core.StringValue(input.CurrentResourceSpec.Fields["statementId"])
	if layerVersionArn == "" || statementId == "" {
		return nil, fmt.Errorf("layerVersionArn and statementId are required")
	}

	layerName, versionNumber, err := parseLayerVersionPermissionArn(layerVersionArn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse layer version ARN: %w", err)
	}

	getLayerVersionPolicyInput := &lambda.GetLayerVersionPolicyInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	}

	_, err = lambdaService.GetLayerVersionPolicy(ctx, getLayerVersionPolicyInput)
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			if apiError.ErrorCode() == "ResourceNotFoundException" {
				return &provider.ResourceGetExternalStateOutput{
					ResourceSpecState: &core.MappingNode{Fields: make(map[string]*core.MappingNode)},
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to get layer version policy: %w", err)
	}

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"layerVersionArn": core.MappingNodeFromString(layerVersionArn),
			"statementId":     core.MappingNodeFromString(statementId),
			"id":              core.MappingNodeFromString(fmt.Sprintf("%s#%s", layerVersionArn, statementId)),
		},
	}

	for key, value := range input.CurrentResourceSpec.Fields {
		if key != "id" {
			resourceSpecState.Fields[key] = value
		}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}
