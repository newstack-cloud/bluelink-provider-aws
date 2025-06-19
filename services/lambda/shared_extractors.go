package lambda

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
)

func functionHandlerValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetFunctionOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetFunctionOutput]{
		Name: "handler",
		Condition: func(output *lambda.GetFunctionOutput) bool {
			return output.Configuration.Handler != nil
		},
		Fields: []string{"handler"},
		Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
			return []*core.MappingNode{
				core.MappingNodeFromString(aws.ToString(output.Configuration.Handler)),
			}, nil
		},
	}
}

func functionKMSKeyArnValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetFunctionOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetFunctionOutput]{
		Name: "kmsKeyArn",
		Condition: func(output *lambda.GetFunctionOutput) bool {
			return output.Configuration.KMSKeyArn != nil
		},
		Fields: []string{"kmsKeyArn"},
		Values: func(output *lambda.GetFunctionOutput) ([]*core.MappingNode, error) {
			return []*core.MappingNode{
				core.MappingNodeFromString(aws.ToString(output.Configuration.KMSKeyArn)),
			}, nil
		},
	}
}

func aliasDescriptionValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetAliasOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetAliasOutput]{
		Name: "description",
		Condition: func(output *lambda.GetAliasOutput) bool {
			return output.Description != nil
		},
		Fields: []string{"description"},
		Values: func(output *lambda.GetAliasOutput) ([]*core.MappingNode, error) {
			return []*core.MappingNode{
				core.MappingNodeFromString(aws.ToString(output.Description)),
			}, nil
		},
	}
}

// Layer version extractors

func layerVersionDescriptionValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput]{
		Name: "description",
		Condition: func(output *lambda.GetLayerVersionOutput) bool {
			return output.Description != nil
		},
		Fields: []string{"description"},
		Values: func(output *lambda.GetLayerVersionOutput) ([]*core.MappingNode, error) {
			return []*core.MappingNode{
				core.MappingNodeFromString(*output.Description),
			}, nil
		},
	}
}

func layerVersionLicenseInfoValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput]{
		Name: "licenseInfo",
		Condition: func(output *lambda.GetLayerVersionOutput) bool {
			return output.LicenseInfo != nil
		},
		Fields: []string{"licenseInfo"},
		Values: func(output *lambda.GetLayerVersionOutput) ([]*core.MappingNode, error) {
			return []*core.MappingNode{
				core.MappingNodeFromString(*output.LicenseInfo),
			}, nil
		},
	}
}

func layerVersionCreatedDateValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput]{
		Name: "createdDate",
		Condition: func(output *lambda.GetLayerVersionOutput) bool {
			return output.CreatedDate != nil
		},
		Fields: []string{"createdDate"},
		Values: func(output *lambda.GetLayerVersionOutput) ([]*core.MappingNode, error) {
			return []*core.MappingNode{
				core.MappingNodeFromString(*output.CreatedDate),
			}, nil
		},
	}
}

func layerVersionCompatibleRuntimesValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput]{
		Name: "compatibleRuntimes",
		Condition: func(output *lambda.GetLayerVersionOutput) bool {
			return len(output.CompatibleRuntimes) > 0
		},
		Fields: []string{"compatibleRuntimes"},
		Values: func(output *lambda.GetLayerVersionOutput) ([]*core.MappingNode, error) {
			runtimeItems := make([]*core.MappingNode, len(output.CompatibleRuntimes))
			for i, runtime := range output.CompatibleRuntimes {
				runtimeItems[i] = core.MappingNodeFromString(string(runtime))
			}
			return []*core.MappingNode{
				{Items: runtimeItems},
			}, nil
		},
	}
}

func layerVersionCompatibleArchitecturesValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput]{
		Name: "compatibleArchitectures",
		Condition: func(output *lambda.GetLayerVersionOutput) bool {
			return len(output.CompatibleArchitectures) > 0
		},
		Fields: []string{"compatibleArchitectures"},
		Values: func(output *lambda.GetLayerVersionOutput) ([]*core.MappingNode, error) {
			archItems := make([]*core.MappingNode, len(output.CompatibleArchitectures))
			for i, arch := range output.CompatibleArchitectures {
				archItems[i] = core.MappingNodeFromString(string(arch))
			}
			return []*core.MappingNode{
				{Items: archItems},
			}, nil
		},
	}
}

func layerVersionContentValueExtractor() pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput] {
	return pluginutils.OptionalValueExtractor[*lambda.GetLayerVersionOutput]{
		Name: "content",
		Condition: func(output *lambda.GetLayerVersionOutput) bool {
			return output.Content != nil
		},
		Fields: []string{"content"},
		Values: func(output *lambda.GetLayerVersionOutput) ([]*core.MappingNode, error) {
			contentFields := make(map[string]*core.MappingNode)

			if output.Content.CodeSha256 != nil {
				contentFields["codeSha256"] = core.MappingNodeFromString(*output.Content.CodeSha256)
			}
			contentFields["codeSize"] = core.MappingNodeFromInt(int(output.Content.CodeSize))
			if output.Content.Location != nil {
				contentFields["location"] = core.MappingNodeFromString(*output.Content.Location)
			}
			if output.Content.SigningJobArn != nil {
				contentFields["signingJobArn"] = core.MappingNodeFromString(*output.Content.SigningJobArn)
			}
			if output.Content.SigningProfileVersionArn != nil {
				contentFields["signingProfileVersionArn"] = core.MappingNodeFromString(*output.Content.SigningProfileVersionArn)
			}

			return []*core.MappingNode{
				{Fields: contentFields},
			}, nil
		},
	}
}
