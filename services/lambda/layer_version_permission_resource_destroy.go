package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (l *lambdaLayerVersionPermissionResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	layerVersionArn := core.StringValue(
		input.ResourceState.SpecData.Fields["layerVersionArn"],
	)
	statementId := core.StringValue(
		input.ResourceState.SpecData.Fields["statementId"],
	)

	if layerVersionArn == "" || statementId == "" {
		return fmt.Errorf("layerVersionArn and statementId are required for destruction")
	}

	layerName, versionNumber, err := parseLayerVersionPermissionArn(layerVersionArn)
	if err != nil {
		return fmt.Errorf("failed to parse layer version ARN: %w", err)
	}

	removeInput := &lambda.RemoveLayerVersionPermissionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
		StatementId:   aws.String(statementId),
	}

	_, err = lambdaService.RemoveLayerVersionPermission(ctx, removeInput)
	return err
}
