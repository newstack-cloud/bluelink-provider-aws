package lambdassm

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// FunctionParameterPathLink returns a link implementation for a link from a Lambda function
// to an SSM parameter path (a synthetic resource representing a parameter hierarchy). The
// Lambda function is granted access to every parameter beneath the path prefix with a single
// statement and, optionally, an environment variable holding the prefix. When the function is
// VPC-isolated, an SSM interface VPC endpoint is activated on the linked flex VPC.
//
// Unlike the per-parameter aws/lambda/function::aws/ssm/parameter link, one link covers an
// entire configuration namespace: the grant and environment variable live and die with the
// namespace, not with any individual parameter beneath it.
func FunctionParameterPathLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(FunctionToParameterLinkDeps) provider.Link {
	return func(
		linkServiceDeps FunctionToParameterLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/function__parameter_path.md")

		actions := &functionParameterPathLinkActions{
			functionParameterLinkActions{
				lambdaServiceFactory: linkServiceDeps.ResourceAService.ServiceFactory,
				iamServiceFactory:    iamServiceFactory,
				ec2ServiceFactory:    ec2ServiceFactory,
				awsConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
			},
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/lambda/function",
			ResourceTypeB: "aws/ssm/parameterPath",
			// Neither resource is written by this link. The function's environment
			// variable is contributed rather than written, and the framework applies
			// what every link contributing to the function needs it to say as one
			// update, so no exclusive lock on the function is taken here.
			Modifies: provider.LinkModifiesNeither,
			// ReconcileLinkNetworking reads the function's live VPC attachment, so the
			// placement link that establishes it has to run first.
			Requires: linkutils.NetworkAttachedRequired(provider.LinkPriorityResourceA),
			// It doesn't matter which resource is created first; the lambda function will
			// be configured to access the parameter path once both have been created.
			Kind:             provider.LinkKindSoft,
			PriorityResource: provider.LinkPriorityResourceNone,
			PlainTextSummary: "A link that grants a Lambda function access to all SSM parameters " +
				"beneath a path prefix.",
			FormattedDescription:             string(description),
			AnnotationDefinitions:            functionParameterPathLinkAnnotations(),
			StageChangesFunc:                 actions.StageChanges,
			ProduceResourceContributionsFunc: actions.ProduceResourceContributions,
			UpdateLinkedResourcesFunc:        actions.UpdateLinkedResources,
			UpdateIntermediaryResourcesFunc:  actions.UpdateIntermediaryResources,
		}
	}
}

// The per-parameter link's actions are embedded for the shared service getters; every
// link hook is explicitly wired to a path-scoped override in the definition above.
type functionParameterPathLinkActions struct {
	functionParameterLinkActions
}
