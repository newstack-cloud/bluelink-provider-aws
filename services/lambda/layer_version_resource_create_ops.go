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

type layerVersionCreate struct {
	input *lambda.PublishLayerVersionInput
}

func (u *layerVersionCreate) Name() string {
	return "create layer version"
}

func (u *layerVersionCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToPublishLayerVersionInput(
		specData,
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *layerVersionCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	publishLayerVersionOutput, err := lambdaService.PublishLayerVersion(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	newSaveOpCtx.ProviderUpstreamID = aws.ToString(publishLayerVersionOutput.LayerVersionArn)
	newSaveOpCtx.Data["publishLayerVersionOutput"] = publishLayerVersionOutput

	return newSaveOpCtx, err
}

func changesToPublishLayerVersionInput(
	specData *core.MappingNode,
) (*lambda.PublishLayerVersionInput, bool, error) {
	input := &lambda.PublishLayerVersionInput{}

	valueSetters := []*pluginutils.ValueSetter[*lambda.PublishLayerVersionInput]{
		pluginutils.NewValueSetter(
			"$.layerName",
			func(value *core.MappingNode, input *lambda.PublishLayerVersionInput) {
				input.LayerName = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.description",
			func(value *core.MappingNode, input *lambda.PublishLayerVersionInput) {
				input.Description = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.licenseInfo",
			func(value *core.MappingNode, input *lambda.PublishLayerVersionInput) {
				input.LicenseInfo = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.compatibleRuntimes",
			func(value *core.MappingNode, input *lambda.PublishLayerVersionInput) {
				runtimes := core.StringSliceValue(value)
				if len(runtimes) > 0 {
					input.CompatibleRuntimes = make([]types.Runtime, len(runtimes))
					for i, runtime := range runtimes {
						input.CompatibleRuntimes[i] = types.Runtime(runtime)
					}
				}
			},
		),
		pluginutils.NewValueSetter(
			"$.compatibleArchitectures",
			func(value *core.MappingNode, input *lambda.PublishLayerVersionInput) {
				architectures := core.StringSliceValue(value)
				if len(architectures) > 0 {
					input.CompatibleArchitectures = make([]types.Architecture, len(architectures))
					for i, arch := range architectures {
						input.CompatibleArchitectures[i] = types.Architecture(arch)
					}
				}
			},
		),
		pluginutils.NewValueSetter(
			"$.content",
			func(value *core.MappingNode, input *lambda.PublishLayerVersionInput) {
				content := &types.LayerVersionContentInput{}
				contentSet := false

				if s3Bucket, ok := pluginutils.GetValueByPath("$.s3Bucket", value); ok {
					content.S3Bucket = aws.String(core.StringValue(s3Bucket))
					contentSet = true
				}
				if s3Key, ok := pluginutils.GetValueByPath("$.s3Key", value); ok {
					content.S3Key = aws.String(core.StringValue(s3Key))
					contentSet = true
				}
				if s3ObjectVersion, ok := pluginutils.GetValueByPath("$.s3ObjectVersion", value); ok {
					content.S3ObjectVersion = aws.String(core.StringValue(s3ObjectVersion))
					contentSet = true
				}

				if contentSet {
					input.Content = content
				}
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
