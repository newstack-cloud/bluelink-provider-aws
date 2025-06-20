package lambda

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/celerity-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaAliasResourceActions) Update(
	ctx context.Context,
	input *provider.ResourceDeployInput,
) (*provider.ResourceDeployOutput, error) {
	lambdaService, err := l.getLambdaService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	updateOperations := []pluginutils.SaveOperation[lambdaservice.Service]{
		&aliasUpdate{},
		&aliasPutProvisionedConcurrencyConfig{},
	}

	hasSavedValues, saveOpCtx, err := pluginutils.RunSaveOperations(
		ctx,
		pluginutils.SaveOperationContext{
			Data: map[string]any{},
		},
		updateOperations,
		input,
		lambdaService,
	)
	if err != nil {
		return nil, err
	}

	if !hasSavedValues {
		return nil, fmt.Errorf("no values were saved during alias update")
	}

	aliasArn, ok := saveOpCtx.Data["aliasArn"]
	if !ok {
		return nil, fmt.Errorf("aliasArn not found in save operation context")
	}

	aliasArnString, ok := aliasArn.(string)
	if !ok {
		return nil, fmt.Errorf("aliasArn is not a string")
	}

	return &provider.ResourceDeployOutput{
		ComputedFieldValues: map[string]*core.MappingNode{
			"spec.aliasArn": core.MappingNodeFromString(aliasArnString),
		},
	}, nil
}

type aliasUpdate struct {
	input *lambda.UpdateAliasInput
}

func (u *aliasUpdate) Name() string {
	return "update alias"
}

func (u *aliasUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToUpdateAliasInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *aliasUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	updateAliasOutput, err := lambdaService.UpdateAlias(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.Data["updateAliasOutput"] = updateAliasOutput
	newSaveOpCtx.Data["aliasArn"] = aws.ToString(updateAliasOutput.AliasArn)

	return newSaveOpCtx, nil
}

func changesToUpdateAliasInput(
	specData *core.MappingNode,
) (*lambda.UpdateAliasInput, bool, error) {
	input := &lambda.UpdateAliasInput{}

	valueSetters := []*pluginutils.ValueSetter[*lambda.UpdateAliasInput]{
		pluginutils.NewValueSetter(
			"$.functionName",
			func(value *core.MappingNode, input *lambda.UpdateAliasInput) {
				input.FunctionName = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.name",
			func(value *core.MappingNode, input *lambda.UpdateAliasInput) {
				input.Name = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.functionVersion",
			func(value *core.MappingNode, input *lambda.UpdateAliasInput) {
				input.FunctionVersion = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.description",
			func(value *core.MappingNode, input *lambda.UpdateAliasInput) {
				input.Description = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.routingConfig",
			func(value *core.MappingNode, input *lambda.UpdateAliasInput) {
				routingConfig := &types.AliasRoutingConfiguration{}
				if weights, exists := pluginutils.GetValueByPath("$.additionalVersionWeights", value); exists {
					additionalVersionWeights := make(map[string]float64)
					for key, valueNode := range weights.Fields {
						additionalVersionWeights[key] = core.FloatValue(valueNode)
					}
					if len(additionalVersionWeights) > 0 {
						routingConfig.AdditionalVersionWeights = additionalVersionWeights
					}
				}
				input.RoutingConfig = routingConfig
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
