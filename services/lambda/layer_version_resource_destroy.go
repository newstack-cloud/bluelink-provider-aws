package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *lambdaLayerVersionResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return err
	}

	// Parse the layer version ARN from the resource state to extract layer name and version number
	layerVersionArn := core.StringValue(
		input.ResourceState.SpecData.Fields["layerVersionArn"],
	)

	layerName, versionNumber, err := parseLayerVersionArn(layerVersionArn)
	if err != nil {
		return fmt.Errorf("failed to parse layer version ARN: %w", err)
	}

	deleteLayerVersionInput := &lambda.DeleteLayerVersionInput{
		LayerName:     aws.String(layerName),
		VersionNumber: aws.Int64(versionNumber),
	}

	_, err = lambdaService.DeleteLayerVersion(ctx, deleteLayerVersionInput)
	return err
}
