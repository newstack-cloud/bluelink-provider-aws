package lambdalinks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/newstack-cloud/celerity-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/providerv1"
)

// FunctionFunctionLink returns a link implementation for
// a link from a lambda function to another lambda function.
// The first lambda function will be configured with permissions
// and environment variables to be able to invoke the second lambda function.
func FunctionFunctionLink(
	linkServiceDeps pluginutils.LinkServiceDeps[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	],
) provider.Link {
	description, _ := descriptions.ReadFile("descriptions/function__function.md")

	actions := &lambdaFunctionFunctionLinkActions{
		lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
		awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
	}

	return &providerv1.LinkDefinition{
		ResourceTypeA: "aws/lambda/function",
		ResourceTypeB: "aws/lambda/function",
		// It doesn't matter which lambda function is created first,
		// the caller function will be configured to be able to invoke
		// the callee function once both have been created.
		Kind:                            provider.LinkKindSoft,
		PriorityResource:                provider.LinkPriorityResourceNone,
		PlainTextSummary:                "A link that configures a lambda function to be able to invoke another lambda function.",
		FormattedDescription:            string(description),
		AnnotationDefinitions:           lambdaFunctionFunctionLinkAnnotations(),
		StageChangesFunc:                actions.StageChanges,
		UpdateResourceAFunc:             actions.UpdateResourceA,
		UpdateResourceBFunc:             actions.UpdateResourceB,
		UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
	}
}

type lambdaFunctionFunctionLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaFunctionFunctionLinkActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}
