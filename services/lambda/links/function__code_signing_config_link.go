package lambdalinks

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/newstack-cloud/celerity-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/providerv1"
)

// FunctionCodeSigningConfigLink returns a link implementation for
// a link from a lambda function to a code signing config.
func FunctionCodeSigningConfigLink(
	linkServiceDeps pluginutils.LinkServiceDeps[
		*aws.Config,
		lambdaservice.Service,
		*aws.Config,
		lambdaservice.Service,
	],
) provider.Link {
	actions := &lambdaFunctionCodeSigningConfigLinkActions{
		lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
		awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
	}
	return &providerv1.LinkDefinition{
		ResourceTypeA:                   "aws/lambda/function",
		ResourceTypeB:                   "aws/lambda/codeSigningConfig",
		Kind:                            provider.LinkKindHard,
		PriorityResource:                provider.LinkPriorityResourceB,
		PlainTextSummary:                "A link from a lambda function to a code signing config.",
		FormattedDescription:            "The link type used to link a lambda function to a code signing config.",
		AnnotationDefinitions:           lambdaFunctionCodeSigningConfigLinkAnnotations(),
		StageChangesFunc:                actions.StageChanges,
		UpdateResourceAFunc:             actions.UpdateResourceA,
		UpdateResourceBFunc:             actions.UpdateResourceB,
		UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
	}
}

type lambdaFunctionCodeSigningConfigLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}

func (l *lambdaFunctionCodeSigningConfigLinkActions) getLambdaService(
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
