package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *lambdaAliasResourceActions) Destroy(
	ctx context.Context,
	input *provider.ResourceDestroyInput,
) error {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return fmt.Errorf("failed to get Lambda service: %w", err)
	}

	// Extract function name and alias name from the resource state
	functionName := core.StringValue(input.ResourceState.SpecData.Fields["functionName"])
	aliasName := core.StringValue(input.ResourceState.SpecData.Fields["name"])

	deleteAliasInput := &lambda.DeleteAliasInput{
		FunctionName: aws.String(functionName),
		Name:         aws.String(aliasName),
	}

	_, err = lambdaService.DeleteAlias(ctx, deleteAliasInput)
	if err != nil {
		return fmt.Errorf("failed to delete Lambda alias: %w", err)
	}

	return nil
}
