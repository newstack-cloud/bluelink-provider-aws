package flexlambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// VPCToFunctionLinkDeps is a type alias for the core dependencies for a link from a
// flex VPC to a Lambda function. Resource A is the flex VPC (backed by the EC2 service)
// and resource B is the Lambda function.
type VPCToFunctionLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	ec2service.Service,
	*aws.Config,
	lambdaservice.Service,
]

// VPCFunctionLink returns a link implementation that places a Lambda function within a
// flex VPC. The function's VpcConfig is set from the VPC's computed subnets (selected by
// the aws.flexvpc.lambda.subnetType annotation) and Bluelink-managed security group.
//
// The function's execution role is granted the network interface permissions Lambda
// requires of a VPC-attached function, so the IAM service is needed alongside the two
// resource services.
func VPCFunctionLink(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
) func(VPCToFunctionLinkDeps) provider.Link {
	return func(
		linkServiceDeps VPCToFunctionLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/vpc__function.md")

		actions := &vpcFunctionLinkActions{
			lambdaServiceFactory: linkServiceDeps.ResourceBService.ServiceFactory,
			awsConfigStore:       linkServiceDeps.ResourceBService.ConfigStore,
			ec2ServiceFactory:    linkServiceDeps.ResourceAService.ServiceFactory,
			ec2ConfigStore:       linkServiceDeps.ResourceAService.ConfigStore,
			iamServiceFactory:    iamServiceFactory,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/flex/vpc",
			ResourceTypeB: "aws/lambda/function",
			// Only one side of the relationship is written, so the deployer takes no
			// exclusive lock on the other and links that merely read it are not
			// serialised against one another.
			//
			// This link writes the function rather than contributing to it, which is
			// deliberate. Removal has to detach the function before revoking the
			// execution role's network interface permissions, for the reason given on
			// UpdateLinkedResources, and a contribution cannot be ordered that way: a link
			// being removed is destroyed before any resource is deployed, so its teardown
			// would revoke the permissions while the function was still attached. Placing
			// the function through a contribution needs a way to order a link's teardown
			// after the update that retracts what it contributed.
			Modifies: provider.LinkModifiesResourceB,
			// The function is placed into the VPC by this link once both have been
			// deployed, and detached by it before either is destroyed, so neither has to
			// exist before the other.
			Kind:             provider.LinkKindSoft,
			PriorityResource: provider.LinkPriorityResourceNone,
			// Every link that reads the function's live VPC attachment waits for this
			// one, and is torn down before it.
			Provides:                        linkutils.NetworkAttachedProvided(provider.LinkPriorityResourceB),
			PlainTextSummary:                "A link that places a Lambda function within a flex VPC's subnets.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           vpcFunctionLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateLinkedResourcesFunc:       actions.UpdateLinkedResources,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type vpcFunctionLinkActions struct {
	lambdaServiceFactory pluginutils.ServiceFactory[*aws.Config, lambdaservice.Service]
	awsConfigStore       pluginutils.ServiceConfigStore[*aws.Config]
	// The link mints a security group for each function it places, so it needs the
	// EC2 service that backs resource A as well as the Lambda service for resource B.
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
	ec2ConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
	// Attaching a function to a VPC requires its execution role to be able to manage
	// network interfaces, which this link grants before it sets the attachment.
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
}

func (l *vpcFunctionLinkActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.iamServiceFactory(awsConfig, providerContext), nil
}

func (l *vpcFunctionLinkActions) getEC2Service(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, error) {
	awsConfig, err := l.ec2ConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.ec2ServiceFactory(awsConfig, providerContext), nil
}

func (l *vpcFunctionLinkActions) getLambdaService(
	ctx context.Context,
	providerContext provider.Context,
) (lambdaservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}

	return l.lambdaServiceFactory(awsConfig, providerContext), nil
}
