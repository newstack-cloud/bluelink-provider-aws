package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *lambdaCodeSigningConfigResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return fmt.Errorf("failed to get Lambda service: %w", err)
	}

	// Extract code signing config ARN from the resource state
	codeSigningConfigArn := core.StringValue(input.ResourceState.SpecData.Fields["codeSigningConfigArn"])

	deleteCodeSigningConfigInput := &lambda.DeleteCodeSigningConfigInput{
		CodeSigningConfigArn: aws.String(codeSigningConfigArn),
	}

	_, err = lambdaService.DeleteCodeSigningConfig(ctx, deleteCodeSigningConfigInput)
	if err != nil {
		return fmt.Errorf("failed to delete Lambda code signing config: %w", err)
	}

	return nil
}
