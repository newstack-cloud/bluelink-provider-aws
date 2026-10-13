package rdsrds

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// Neither resource is modified by this link. The proxy already names the instance as its
// target through the target group, and the instance knows nothing of the proxy; all this
// link contributes is the network path between them, which lives on their security
// groups.
func (l *proxyInstanceLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData:             core.MappingNodeFields(),
		ResourceDataMappings: map[string]string{},
	}, nil
}

// UpdateIntermediaryResources opens the security-group path from the proxy to the
// instance on the database port, and revokes it on destroy.
func (l *proxyInstanceLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	output := &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}

	caller, hasCaller := callerNetworkingFromProxy(input.ResourceAInfo)
	if !hasCaller {
		// A proxy always sits in a VPC, so this means its networking has not reached
		// state yet rather than that it has none. Reporting it is better than opening
		// nothing and leaving the target group to time out with no explanation.
		return nil, fmt.Errorf(
			"VPC networking could not be retrieved from the linked from %q RDS proxy resource",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}

	targetSecurityGroupIDs, hasTargetSG := instanceSecurityGroupIDs(input.ResourceBInfo)
	if !hasTargetSG {
		return nil, fmt.Errorf(
			"security group could not be retrieved from the linked to %q RDS instance resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	ec2Service, region, err := l.getEC2ServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	reconciled, err := linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:                 caller,
			Region:                 region,
			TargetSecurityGroupIDs: targetSecurityGroupIDs,
			TargetPort:             databasePort(input.ResourceAInfo, input.ResourceBInfo),
		},
		output,
	)
	if err != nil {
		return nil, err
	}

	// Registration follows the rule, in that order and in this one link: a target
	// registered before the path is open sits UNAVAILABLE, and anything waiting on its
	// health waits for the rule that has not been written yet.
	proxyName, instanceIdentifier, err := proxyAndInstanceNames(
		input.ResourceAInfo,
		input.ResourceBInfo,
	)
	if err != nil {
		return nil, err
	}

	rdsService, err := l.getRDSService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		if err := deregisterTarget(ctx, rdsService, proxyName, instanceIdentifier); err != nil {
			return nil, err
		}
		return reconciled, nil
	}

	if err := registerTarget(ctx, rdsService, proxyName, instanceIdentifier); err != nil {
		return nil, err
	}

	return reconciled, nil
}

// proxyAndInstanceNames reads the physical names registration needs from each side's
// state.
func proxyAndInstanceNames(
	proxyInfo *provider.ResourceInfo,
	instanceInfo *provider.ResourceInfo,
) (string, string, error) {
	proxySpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(proxyInfo)
	proxyName, hasProxyName := pluginutils.GetValueByPath("$.dbProxyName", proxySpec)
	if !hasProxyName || core.StringValue(proxyName) == "" {
		return "", "", fmt.Errorf(
			"the name of the linked from %q RDS proxy resource could not be retrieved",
			pluginutils.GetResourceName(proxyInfo),
		)
	}

	instanceSpec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(instanceInfo)
	instanceID, hasInstanceID := pluginutils.GetValueByPath(
		"$.dbInstanceIdentifier",
		instanceSpec,
	)
	if !hasInstanceID || core.StringValue(instanceID) == "" {
		return "", "", fmt.Errorf(
			"the identifier of the linked to %q RDS instance resource could not be retrieved",
			pluginutils.GetResourceName(instanceInfo),
		)
	}

	return core.StringValue(proxyName), core.StringValue(instanceID), nil
}

// Adapts an RDS proxy's own VPC placement into the
// platform-agnostic CallerNetworking that ReconcileLinkNetworking consumes. Unlike a
// Lambda function, a proxy is placed by its own spec rather than by a placement link,
// so its networking is read straight from its state.
func callerNetworkingFromProxy(proxyInfo *provider.ResourceInfo) (linkutils.CallerNetworking, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(proxyInfo)

	vpcID, hasVPC := pluginutils.GetValueByPath("$.vpcId", spec)
	if !hasVPC || core.StringValue(vpcID) == "" {
		return linkutils.CallerNetworking{}, false
	}

	subnetIDs := stringListAtPath(spec, "$.vpcSubnetIds")
	securityGroupIDs := stringListAtPath(spec, "$.vpcSecurityGroupIds")
	if len(subnetIDs) == 0 || len(securityGroupIDs) == 0 {
		return linkutils.CallerNetworking{}, false
	}

	return linkutils.CallerNetworking{
		VPCID:            core.StringValue(vpcID),
		SubnetIDs:        subnetIDs,
		SecurityGroupIDs: securityGroupIDs,
	}, true
}

func instanceSecurityGroupIDs(instanceInfo *provider.ResourceInfo) ([]string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(instanceInfo)

	// vpcSecurityGroups, not vpcSecurityGroupIds: AWS::RDS::DBInstance names the field
	// VPCSecurityGroups while AWS::RDS::DBProxy names its own vpcSecurityGroupIds.
	// Reading the proxy's name off the instance finds nothing and reports the instance
	// as having no security group at all.
	//
	// Every group the instance carries is passed, not one chosen here. Which group the
	// rule should name is decided by matching against the groups the flex VPC provisioned,
	// so that an author's own group is never paired against.
	groupIDs := stringListAtPath(spec, "$.vpcSecurityGroups")
	return groupIDs, len(groupIDs) > 0
}

func stringListAtPath(spec *core.MappingNode, path string) []string {
	node, has := pluginutils.GetValueByPath(path, spec)
	if !has || node == nil || len(node.Items) == 0 {
		return nil
	}

	values := make([]string, 0, len(node.Items))
	for _, item := range node.Items {
		if value := core.StringValue(item); value != "" {
			values = append(values, value)
		}
	}
	return values
}
