package rdsrds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	rdsservice "github.com/newstack-cloud/bluelink-provider-aws/services/rds/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

// ProxyToInstanceLinkDeps is a type alias for the core dependencies of a link from an RDS
// proxy to the DB instance it fronts. Both are Cloud Control–backed; the link only reads
// their networking from state and calls EC2 to open the path between them.
type ProxyToInstanceLinkDeps pluginutils.LinkServiceDeps[
	*aws.Config,
	cloudcontrolservice.Service,
	*aws.Config,
	cloudcontrolservice.Service,
]

// ProxyInstanceLink returns a link implementation for a link from an RDS proxy to the DB
// instance it fronts. It opens the security-group path between them on the database port
// and then registers the instance as the proxy's target, in that order.
//
// Both halves live here because the order between them matters and nothing else can
// enforce it. A target registered before the path is open is reported as UNAVAILABLE, and anything
// waiting on its health waits for a rule that has not been written.
//
// This is why there is no AWS::RDS::DBProxyTargetGroup resource. That type's create
// blocks until the registered target reports healthy, so as a resource it ran in an
// earlier phase than the link writing the rule it depended on, and spent its whole
// timeout waiting for something that had not happened yet. RDS creates the default
// target group with the proxy regardless, so nothing needs to create it.
//
// The path is not implied by the two sharing a security group. Security groups do not
// permit traffic within a group by default, so a proxy and instance in the same group
// still need a rule naming that group on both sides; ReconcileLinkNetworking writes
// exactly that when the caller and target resolve to the same group.
func ProxyInstanceLink(
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
	rdsServiceFactory pluginutils.ServiceFactory[*aws.Config, rdsservice.Service],
) func(ProxyToInstanceLinkDeps) provider.Link {
	return func(
		linkServiceDeps ProxyToInstanceLinkDeps,
	) provider.Link {
		description, _ := descriptions.ReadFile("descriptions/proxy__instance.md")

		actions := &proxyInstanceLinkActions{
			ec2ServiceFactory: ec2ServiceFactory,
			rdsServiceFactory: rdsServiceFactory,
			awsConfigStore:    linkServiceDeps.ResourceAService.ConfigStore,
		}

		return &providerv1.LinkDefinition{
			ResourceTypeA: "aws/rds/dbProxy",
			ResourceTypeB: "aws/rds/dbInstance",
			// Neither resource in the relationship is written; the link does its work
			// through intermediary resources.
			Modifies: provider.LinkModifiesNeither,
			// A proxy carries its own VPC attachment in its spec rather than being placed
			// by a separate link, so there is no placement to wait on.
			Kind: provider.LinkKindSoft,
			// The rule that names the group the instance carries is this link's own work,
			// and reads the instance's port when both resources have been deployed, so
			// neither has to exist before the other.
			PriorityResource:                provider.LinkPriorityResourceNone,
			PlainTextSummary:                "A link that opens the network path from an RDS proxy to the DB instance it fronts and registers it as the proxy's target.",
			FormattedDescription:            string(description),
			AnnotationDefinitions:           proxyInstanceLinkAnnotations(),
			StageChangesFunc:                actions.StageChanges,
			UpdateLinkedResourcesFunc:       actions.UpdateLinkedResources,
			UpdateIntermediaryResourcesFunc: actions.UpdateIntermediaryResources,
		}
	}
}

type proxyInstanceLinkActions struct {
	awsConfigStore    pluginutils.ServiceConfigStore[*aws.Config]
	ec2ServiceFactory pluginutils.ServiceFactory[*aws.Config, ec2service.Service]
	rdsServiceFactory pluginutils.ServiceFactory[*aws.Config, rdsservice.Service]
}

func (l *proxyInstanceLinkActions) getRDSService(
	ctx context.Context,
	providerContext provider.Context,
) (rdsservice.Service, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, err
	}
	return l.rdsServiceFactory(awsConfig, providerContext), nil
}

func (l *proxyInstanceLinkActions) getEC2ServiceWithRegion(
	ctx context.Context,
	providerContext provider.Context,
) (ec2service.Service, string, error) {
	awsConfig, err := l.awsConfigStore.FromProviderContext(ctx, providerContext, nil)
	if err != nil {
		return nil, "", err
	}
	return l.ec2ServiceFactory(awsConfig, providerContext), awsConfig.Region, nil
}
