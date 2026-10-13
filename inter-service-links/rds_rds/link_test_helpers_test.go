//go:build unit

package rdsrds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	rdsservice "github.com/newstack-cloud/bluelink-provider-aws/services/rds/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

const (
	testProxyResource   = "ordersProxy"
	testProxyName       = "orders-proxy"
	testInstanceName    = "ordersInstance"
	testInstanceID      = "orders-instance"
	testOtherInstanceID = "audit-instance"

	testVPCID       = "vpc-1"
	testFlexVPCName = "orders-vpc"
	// The flex VPC resource's name in the blueprint, which is what its lock is keyed on.
	testFlexVPCResource = "appVpc"
	testSubnetID        = "subnet-1"
	testProxySGID       = "sg-proxy"
	// The group the flex VPC minted for the database tier, which is the one the rule
	// pairs against.
	testInstanceSGID = "sg-db"
	// A group the author put on the instance themselves, which must be left alone.
	testAuthorSGID = "sg-author-owned"
)

func testLinkContext() provider.LinkContext {
	return plugintestutils.NewTestLinkContext(
		map[string]map[string]*core.ScalarValue{
			"aws": {"region": core.ScalarFromString("us-west-2")},
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)
}

func testConfigStore() pluginutils.ServiceConfigStore[*aws.Config] {
	return utils.NewAWSConfigStore(
		[]string{},
		utils.AWSConfigFromProviderContext,
		&testutils.MockAWSConfigLoader{},
		utils.AWSConfigCacheKey,
	)
}

func proxySpec() *core.MappingNode {
	return core.MappingNodeFields(
		"dbProxyName", core.MappingNodeFromString(testProxyName),
		"vpcId", core.MappingNodeFromString(testVPCID),
		"vpcSubnetIds", core.MappingNodeItems(
			core.MappingNodeFromString(testSubnetID),
		),
		"vpcSecurityGroupIds", core.MappingNodeItems(
			core.MappingNodeFromString(testProxySGID),
		),
	)
}

func proxyInfo(spec *core.MappingNode, annotations map[string]*core.MappingNode) *provider.ResourceInfo {
	info := &provider.ResourceInfo{
		ResourceName:         testProxyResource,
		InstanceID:           "instance-1",
		CurrentResourceState: &state.ResourceState{SpecData: spec},
	}
	if annotations != nil {
		info.ResourceWithResolvedSubs = &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: annotations},
			},
		}
	}
	return info
}

// The instance as its state carries it: the identifier registration needs, and every
// security group it holds, author-owned ones included.
func instanceSpec(extra ...*core.MappingNode) *core.MappingNode {
	spec := core.MappingNodeFields(
		"dbInstanceIdentifier", core.MappingNodeFromString(testInstanceID),
		"vpcSecurityGroups", core.MappingNodeItems(
			core.MappingNodeFromString(testAuthorSGID),
			core.MappingNodeFromString(testInstanceSGID),
		),
	)
	for _, fields := range extra {
		for name, value := range fields.Fields {
			spec.Fields[name] = value
		}
	}
	return spec
}

func instanceInfo(spec *core.MappingNode) *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName:         testInstanceName,
		InstanceID:           "instance-1",
		CurrentResourceState: &state.ResourceState{SpecData: spec},
	}
}

// The VPC as EC2 reports it, carrying the tag the flex VPC resource name is recovered
// from.
func flexVPCDescribeOutput() []*ec2.DescribeVpcsOutput {
	return []*ec2.DescribeVpcsOutput{
		{
			Vpcs: []ec2types.Vpc{
				{
					VpcId: aws.String(testVPCID),
					Tags: []ec2types.Tag{
						{
							Key:   aws.String(flex.TagFlexVPCName),
							Value: aws.String(testFlexVPCName),
						},
					},
				},
			},
		},
	}
}

func flexVPCState() *state.ResourceState {
	return &state.ResourceState{
		Name: testFlexVPCResource,
		SpecData: core.MappingNodeFields(
			"name", core.MappingNodeFromString(testFlexVPCName),
			"enableDNSSupport", core.MappingNodeFromBool(true),
			"enableDNSHostnames", core.MappingNodeFromBool(true),
			"securityGroupIdsByName", core.MappingNodeFields(
				"db", core.MappingNodeFromString(testInstanceSGID),
			),
		),
	}
}

// The EC2 calls an open path needs: the VPC lookup plus the two rules.
func pairingEC2Mock() *ec2mock.Ec2ServiceMock {
	return ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
}

// The EC2 calls the destroy path needs, with no rule of this link's own to revoke.
func revokingEC2Mock() *ec2mock.Ec2ServiceMock {
	return ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithDescribeSecurityGroupRulesOutput(&ec2.DescribeSecurityGroupRulesOutput{}),
	)
}

// Runs the link's intermediary resources phase, which is the only surface this link has:
// it opens the path and registers the target, in that order, and nothing else is public.
func runUpdate(
	updateType provider.LinkUpdateType,
	proxy *provider.ResourceInfo,
	instance *provider.ResourceInfo,
	ec2Svc *ec2mock.Ec2ServiceMock,
	rdsSvc rdsservice.Service,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	link := ProxyInstanceLink(
		func(*aws.Config, provider.Context) ec2service.Service { return ec2Svc },
		func(*aws.Config, provider.Context) rdsservice.Service { return rdsSvc },
	)(ProxyToInstanceLinkDeps{
		ResourceAService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
			ServiceFactory: noopCloudControlServiceFactory,
			ConfigStore:    testConfigStore(),
		},
		ResourceBService: pluginutils.ServiceWithConfigStore[*aws.Config, cloudcontrolservice.Service]{
			ServiceFactory: noopCloudControlServiceFactory,
			ConfigStore:    testConfigStore(),
		},
	})

	return link.UpdateIntermediaryResources(
		context.Background(),
		&provider.LinkUpdateIntermediaryResourcesInput{
			LinkID:         "link-1",
			InstanceName:   "test-instance",
			LinkUpdateType: updateType,
			LinkContext:    testLinkContext(),
			ResourceAInfo:  proxy,
			ResourceBInfo:  instance,
			ResourceService: resourceservicemock.Create(
				resourceservicemock.WithLookupResourceInState(flexVPCState()),
			),
		},
	)
}

func noopCloudControlServiceFactory(_ *aws.Config, _ provider.Context) cloudcontrolservice.Service {
	return nil
}
