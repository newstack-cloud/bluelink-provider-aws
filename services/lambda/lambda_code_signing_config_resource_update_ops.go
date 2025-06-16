package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

type codeSigningConfigUpdate struct {
	input *lambda.UpdateCodeSigningConfigInput
}

func (u *codeSigningConfigUpdate) Name() string {
	return "update code signing config"
}

func (u *codeSigningConfigUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToUpdateCodeSigningConfigInput(specData)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input

	// Store the ARN for other operations
	if input.CodeSigningConfigArn != nil {
		saveOpCtx.Data["codeSigningConfigArn"] = aws.ToString(input.CodeSigningConfigArn)
	}

	return hasValues, saveOpCtx, nil
}

func (u *codeSigningConfigUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	updateCodeSigningConfigOutput, err := lambdaService.UpdateCodeSigningConfig(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.Data["updateCodeSigningConfigOutput"] = updateCodeSigningConfigOutput
	newSaveOpCtx.Data["codeSigningConfigArn"] = aws.ToString(updateCodeSigningConfigOutput.CodeSigningConfig.CodeSigningConfigArn)

	return newSaveOpCtx, nil
}

func changesToUpdateCodeSigningConfigInput(
	specData *core.MappingNode,
) (*lambda.UpdateCodeSigningConfigInput, bool, error) {
	input := &lambda.UpdateCodeSigningConfigInput{}

	valueSetters := []*pluginutils.ValueSetter[*lambda.UpdateCodeSigningConfigInput]{
		pluginutils.NewValueSetter(
			"$.codeSigningConfigArn",
			func(value *core.MappingNode, input *lambda.UpdateCodeSigningConfigInput) {
				input.CodeSigningConfigArn = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.description",
			func(value *core.MappingNode, input *lambda.UpdateCodeSigningConfigInput) {
				input.Description = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.allowedPublishers",
			func(value *core.MappingNode, input *lambda.UpdateCodeSigningConfigInput) {
				allowedPublishers := &types.AllowedPublishers{}
				if arns, exists := pluginutils.GetValueByPath("$.signingProfileVersionArns", value); exists {
					var arnList []string
					for _, arnNode := range arns.Items {
						arnList = append(arnList, core.StringValue(arnNode))
					}
					allowedPublishers.SigningProfileVersionArns = arnList
				}
				input.AllowedPublishers = allowedPublishers
			},
		),
		pluginutils.NewValueSetter(
			"$.codeSigningPolicies",
			func(value *core.MappingNode, input *lambda.UpdateCodeSigningConfigInput) {
				codeSigningPolicies := &types.CodeSigningPolicies{}
				if policy, exists := pluginutils.GetValueByPath("$.untrustedArtifactOnDeployment", value); exists {
					policyStr := core.StringValue(policy)
					codeSigningPolicies.UntrustedArtifactOnDeployment = types.CodeSigningPolicy(policyStr)
				}
				input.CodeSigningPolicies = codeSigningPolicies
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
