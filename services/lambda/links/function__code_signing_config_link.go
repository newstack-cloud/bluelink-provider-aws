package lambdalinks

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
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
	description, _ := descriptions.ReadFile("descriptions/function__code_signing_config.md")

	actions := &lambdaFunctionCodeSigningConfigLinkActions{
		lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
		awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
	}

	return &providerv1.LinkDefinition{
		ResourceTypeA: "aws/lambda/function",
		ResourceTypeB: "aws/lambda/codeSigningConfig",
		// Neither resource is written by this link. The code signing configuration is contributed
		// rather than written, and the framework applies what every link
		// contributing to the resource needs it to say as one update, so no
		// exclusive lock is taken here.
		Modifies: provider.LinkModifiesNeither,
		CardinalityA: provider.LinkCardinality{
			// A lambda function can only have one signing config.
			Max: 1,
		},
		Kind:                             provider.LinkKindHard,
		PriorityResource:                 provider.LinkPriorityResourceB,
		PlainTextSummary:                 "A link from a lambda function to a code signing config.",
		FormattedDescription:             string(description),
		AnnotationDefinitions:            lambdaFunctionCodeSigningConfigLinkAnnotations(),
		StageChangesFunc:                 actions.StageChanges,
		ProduceResourceContributionsFunc: actions.ProduceResourceContributions,
		UpdateLinkedResourcesFunc:        actions.UpdateLinkedResources,
		UpdateIntermediaryResourcesFunc:  actions.UpdateIntermediaryResources,
	}
}

type lambdaFunctionCodeSigningConfigLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
}
