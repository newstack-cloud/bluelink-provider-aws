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
