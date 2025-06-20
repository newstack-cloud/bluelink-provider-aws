package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/celerity-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

type aliasCreate struct {
	input *lambda.CreateAliasInput
}

func (u *aliasCreate) Name() string {
	return "create alias"
}

func (u *aliasCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToCreateAliasInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *aliasCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	createAliasOutput, err := lambdaService.CreateAlias(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(createAliasOutput.AliasArn)
	newSaveOpCtx.Data["createAliasOutput"] = createAliasOutput
	newSaveOpCtx.Data["aliasArn"] = aws.ToString(createAliasOutput.AliasArn)

	return newSaveOpCtx, nil
}

type aliasPutProvisionedConcurrencyConfig struct {
	input *lambda.PutProvisionedConcurrencyConfigInput
}

func (u *aliasPutProvisionedConcurrencyConfig) Name() string {
	return "put provisioned concurrency config for alias"
}

func (u *aliasPutProvisionedConcurrencyConfig) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	provisionedConcurrencyConfigData, exists := pluginutils.GetValueByPath(
		"$.provisionedConcurrencyConfig",
		specData,
	)
	if !exists {
		return false, saveOpCtx, nil
	}

	aliasArn := extractAliasArn(saveOpCtx)
	input, hasUpdates := changesToAliasProvisionedConcurrencyConfigInput(
		aliasArn,
		provisionedConcurrencyConfigData,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func extractAliasArn(saveOpCtx pluginutils.SaveOperationContext) string {
	if arn, ok := saveOpCtx.Data["aliasArn"]; ok && arn != nil {
		if arnStr, ok := arn.(string); ok {
			return arnStr
		}
	}
	return ""
}

func (u *aliasPutProvisionedConcurrencyConfig) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.PutProvisionedConcurrencyConfig(ctx, u.input)
	return saveOpCtx, err
}

func changesToCreateAliasInput(
	specData *core.MappingNode,
) (*lambda.CreateAliasInput, bool, error) {
	input := &lambda.CreateAliasInput{}

	valueSetters := []*pluginutils.ValueSetter[*lambda.CreateAliasInput]{
		pluginutils.NewValueSetter(
			"$.functionName",
			func(value *core.MappingNode, input *lambda.CreateAliasInput) {
				input.FunctionName = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.name",
			func(value *core.MappingNode, input *lambda.CreateAliasInput) {
				input.Name = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.functionVersion",
			func(value *core.MappingNode, input *lambda.CreateAliasInput) {
				input.FunctionVersion = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.description",
			func(value *core.MappingNode, input *lambda.CreateAliasInput) {
				input.Description = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.routingConfig",
			func(value *core.MappingNode, input *lambda.CreateAliasInput) {
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

func changesToAliasProvisionedConcurrencyConfigInput(
	aliasArn string,
	specData *core.MappingNode,
) (*lambda.PutProvisionedConcurrencyConfigInput, bool) {
	input := &lambda.PutProvisionedConcurrencyConfigInput{
		FunctionName: aws.String(aliasArn),
	}

	provisionedConcurrentExecutions, ok := pluginutils.GetValueByPath(
		"$.provisionedConcurrentExecutions",
		specData,
	)
	if !ok {
		return nil, false
	}

	input.ProvisionedConcurrentExecutions = aws.Int32(
		int32(core.IntValue(provisionedConcurrentExecutions)),
	)

	return input, true
}
