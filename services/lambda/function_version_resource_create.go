package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionVersionResourceActions) Create(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	createOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&functionVersionCreate{},
		&functionRuntimeManagementConfigUpdate{
			path:                 "$.runtimePolicy",
			fieldChangesPathRoot: "spec.runtimePolicy",
		},
		&functionVersionPutProvisionedConcurrencyConfig{},
	}

	hasSavedValues, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			Data: map[string]any{},
		},
		createOperations,
		input,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	if !hasSavedValues {
		return nil, fmt.Errorf("no values were saved during function version creation")
	}

	publishVersionOutput, ok := saveOpCtx.Data["publishVersionOutput"].(*lambda.PublishVersionOutput)
	if !ok {
		return nil, fmt.Errorf("publishVersionOutput not found in save operation context")
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.functionArn": core.MappingNodeFromString(aws.ToString(publishVersionOutput.FunctionArn)),
			"spec.version":     core.MappingNodeFromString(aws.ToString(publishVersionOutput.Version)),
			"spec.functionArnWithVersion": core.MappingNodeFromString(
				aws.ToString(publishVersionOutput.FunctionArn) +
					":" + aws.ToString(publishVersionOutput.Version),
			),
		},
	}, nil
}

func changesToPublishVersionInput(
	specData *core.MappingNode,
) (*lambda.PublishVersionInput, bool, error) {
	input := &lambda.PublishVersionInput{}

	valueSetters := []*pluginutils.ValueSetter[*lambda.PublishVersionInput]{
		pluginutils.NewValueSetter(
			"$.functionName",
			func(value *core.MappingNode, input *lambda.PublishVersionInput) {
				input.FunctionName = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.codeSha256",
			func(value *core.MappingNode, input *lambda.PublishVersionInput) {
				input.CodeSha256 = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.description",
			func(value *core.MappingNode, input *lambda.PublishVersionInput) {
				input.Description = aws.String(core.StringValue(value))
			},
		),
	}

	hasValuesToSave := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasValuesToSave = hasValuesToSave || valueSetter.DidSet()
	}

	return input, hasValuesToSave, nil
}

func changesToPutProvisionedConcurrencyConfigInput(
	functionARN string,
	version string,
	specData *core.MappingNode,
) (*lambda.PutProvisionedConcurrencyConfigInput, bool) {
	input := &lambda.PutProvisionedConcurrencyConfigInput{
		FunctionName: &functionARN,
		Qualifier:    &version,
	}

	provisionedConcurrentExecutions, ok := pluginutils.GetValueByPath(
		"$.provisionedConcurrentExecutions",
		specData,
	)
	if !ok {
		// When provisioned concurrent executions is not set,
		// there is no need to update provisioned concurrency config.
		return nil, false
	}

	input.ProvisionedConcurrentExecutions = aws.Int32(
		int32(core.IntValue(provisionedConcurrentExecutions)),
	)

	return input, true
}
