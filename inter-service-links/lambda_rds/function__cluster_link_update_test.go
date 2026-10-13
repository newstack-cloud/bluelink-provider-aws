//go:build unit

package lambdards

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/newstack-cloud/bluelink-provider-aws/flex"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	resourceservicemock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/resourceservice_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	ec2service "github.com/newstack-cloud/bluelink-provider-aws/services/ec2/service"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

const (
	fcFunctionARN  = "arn:aws:lambda:us-west-2:123456789012:function:api"
	fcRoleARN      = "arn:aws:iam::123456789012:role/api-role"
	fcRoleName     = "api-role"
	fcRoleResource = "apiFunctionRole"
	fcExecRole     = "apiFunctionExecutionRole"
	fcConnectSID   = "RDSConnectordersCluster"
	fcPrefix       = "ORDERS_DB"

	testClusterEndpoint       = "orders-cluster.cluster-abc.us-west-2.rds.amazonaws.com"
	testClusterReaderEndpoint = "orders-cluster.cluster-ro-abc.us-west-2.rds.amazonaws.com"
	testClusterARN            = "arn:aws:rds:us-west-2:123456789012:cluster:orders-cluster"
	testClusterResourceID     = "cluster-ABCDEF0123456789"
	testClusterSGID           = "sg-db-cluster"
)

type FunctionClusterLinkUpdateSuite struct {
	suite.Suite
}

func functionClusterLinkFactory(
	iamSvc iamservice.Service,
	ec2Factory pluginutils.ServiceFactory[*aws.Config, ec2service.Service],
) func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	build := FunctionClusterLink(
		func(c *aws.Config, pc provider.Context) iamservice.Service { return iamSvc },
		ec2Factory,
	)
	return func(
		deps pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
	) provider.Link {
		return build(FunctionToClusterLinkDeps(deps))
	}
}

func fcFunctionInfo(annotations map[string]*core.MappingNode) *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "apiFunction",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields("arn", core.MappingNodeFromString(fcFunctionARN)),
		},
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: annotations},
			},
		},
	}
}

func fcEnvVarAnnotations() map[string]*core.MappingNode {
	return map[string]*core.MappingNode{
		"aws.lambda.rds.ordersCluster.envVarPrefix": core.MappingNodeFromString(fcPrefix),
		"aws.lambda.rds.ordersCluster.databaseName": core.MappingNodeFromString("orders"),
	}
}

func fcClusterInfo() *provider.ResourceInfo {
	return &provider.ResourceInfo{
		ResourceName: "ordersCluster",
		CurrentResourceState: &state.ResourceState{
			SpecData: core.MappingNodeFields(
				"endpoint", core.MappingNodeFields(
					"address", core.MappingNodeFromString(testClusterEndpoint),
					"port", core.MappingNodeFromString("5432"),
				),
				"readEndpoint", core.MappingNodeFields(
					"address", core.MappingNodeFromString(testClusterReaderEndpoint),
				),
				"dbClusterArn", core.MappingNodeFromString(testClusterARN),
				"dbClusterResourceId", core.MappingNodeFromString(testClusterResourceID),
				"vpcSecurityGroupIds", &core.MappingNode{Items: []*core.MappingNode{
					core.MappingNodeFromString(testClusterSGID),
				}},
			),
		},
	}
}

func fcGetFunctionOutput(vars map[string]string, vpcConfig *lambdatypes.VpcConfigResponse) *lambda.GetFunctionOutput {
	return &lambda.GetFunctionOutput{
		Configuration: &lambdatypes.FunctionConfiguration{
			FunctionArn: aws.String(fcFunctionARN),
			Role:        aws.String(fcRoleARN),
			Environment: &lambdatypes.EnvironmentResponse{Variables: vars},
			VpcConfig:   vpcConfig,
		},
	}
}

func fcRoleState() *state.ResourceState {
	return &state.ResourceState{
		Name: fcRoleResource,
		SpecData: core.MappingNodeFields(
			"roleName", core.MappingNodeFromString(fcRoleName),
			"arn", core.MappingNodeFromString(fcRoleARN),
		),
	}
}

// Test_update_intermediary_resources_iam_grant: authMode=iam on a non-VPC function grants
// rds-db:connect scoped to the cluster resource id (networking is a no-op when not VPC-attached).
func (s *FunctionClusterLinkUpdateSuite) Test_update_intermediary_resources_iam_grant() {
	loader := &testutils.MockAWSConfigLoader{}

	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fcGetFunctionOutput(map[string]string{}, nil)),
	)
	rs := resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(fcRoleState()))

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "grants rds-db:connect scoped to the cluster and db user",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &iamSvc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType: provider.LinkUpdateTypeCreate,
			InstanceName:   "test-instance",
			ResourceAInfo: fcFunctionInfo(map[string]*core.MappingNode{
				"aws.lambda.rds.ordersCluster.authMode": core.MappingNodeFromString("iam"),
				"aws.lambda.rds.ordersCluster.dbUser":   core.MappingNodeFromString("orders_app"),
			}),
			ResourceBInfo:    fcClusterInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  rs,
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutputMatcher: fcMatchConnectOutput,
		UpdateActionsCalled: map[string]any{
			"PutRolePolicy": func(arg any) bool { return fcMatchConnectPolicy(arg, "orders_app") },
		},
		UpdateActionsNotCalled: []string{"DeleteRolePolicy"},
	}

	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionClusterLinkFactory(iamSvc, noopEC2ServiceFactory()),
		&s.Suite,
	)
}

// Test_update_intermediary_resources_sg_pair: authMode=password on a VPC-attached function
// opens the SG-pair rule (no IAM grant).
func (s *FunctionClusterLinkUpdateSuite) Test_update_intermediary_resources_sg_pair() {
	loader := &testutils.MockAWSConfigLoader{}

	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeVpcsOutputs(flexVPCDescribeOutput()),
		ec2mock.WithAuthorizeSecurityGroupIngressOutput(&ec2.AuthorizeSecurityGroupIngressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(fcGetFunctionOutput(map[string]string{}, &lambdatypes.VpcConfigResponse{
			VpcId:            aws.String("vpc-1"),
			SubnetIds:        []string{"subnet-1"},
			SecurityGroupIds: []string{"sg-caller"},
		})),
	)
	// Password mode does not look up the role, so the single lookup resolves the flex VPC.
	rs := resourceservicemock.Create(resourceservicemock.WithLookupResourceInState(&state.ResourceState{
		Name: "appVpc",
		SpecData: core.MappingNodeFields(
			"name", core.MappingNodeFromString("orders-vpc"),
			// The group the VPC minted for the target. A link pairs against one of
			// these rather than against whatever group the target lists first.
			"securityGroupIdsByName", core.MappingNodeFields(
				"db", core.MappingNodeFromString(testClusterSGID),
			),
			"enableDNSSupport", core.MappingNodeFromBool(true),
			"enableDNSHostnames", core.MappingNodeFromBool(true),
		),
	}))

	testCase := plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		Name:                           "opens the SG-pair rule for a VPC-attached function",
		ServiceFactoryA:                func(c *aws.Config, pc provider.Context) lambdaservice.Service { return lambdaSvc },
		ConfigStoreA:                   testConfigStore(loader),
		ServiceFactoryB:                noopCloudControlServiceFactory,
		ConfigStoreB:                   testConfigStore(loader),
		IntermediariesServiceMockCalls: &ec2Svc.MockCalls,
		Input: &provider.LinkUpdateIntermediaryResourcesInput{
			LinkUpdateType:   provider.LinkUpdateTypeCreate,
			InstanceName:     "test-instance",
			ResourceAInfo:    fcFunctionInfo(map[string]*core.MappingNode{}),
			ResourceBInfo:    fcClusterInfo(),
			LinkContext:      testLinkContext(),
			ResourceService:  rs,
			CurrentLinkState: &state.LinkState{},
		},
		ExpectedOutput: &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()},
		UpdateActionsCalled: map[string]any{
			"AuthorizeSecurityGroupIngress": func(arg any) bool {
				in, ok := arg.(*ec2.AuthorizeSecurityGroupIngressInput)
				return ok && aws.ToString(in.GroupId) == testClusterSGID && len(in.IpPermissions) == 1 &&
					aws.ToInt32(in.IpPermissions[0].FromPort) == 5432
			},
			"AuthorizeSecurityGroupEgress": func(arg any) bool {
				in, ok := arg.(*ec2.AuthorizeSecurityGroupEgressInput)
				return ok && aws.ToString(in.GroupId) == "sg-caller"
			},
		},
		UpdateActionsNotCalled: []string{"PutRolePolicy"},
	}

	// The asserted service here is ec2 (SG-pair); the iam service is unused in password mode.
	plugintestutils.RunLinkUpdateIntermediaryResourcesTestCases(
		[]plugintestutils.LinkUpdateIntermediaryResourcesTestCase[
			*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
		]{testCase},
		functionClusterLinkFactory(
			iammock.CreateIamServiceMock(),
			func(c *aws.Config, pc provider.Context) ec2service.Service { return ec2Svc },
		),
		&s.Suite,
	)
}

func fcMatchConnectPolicy(arg any, dbUser string) bool {
	input, ok := arg.(*iam.PutRolePolicyInput)
	if !ok {
		return false
	}
	if aws.ToString(input.RoleName) != fcRoleName ||
		aws.ToString(input.PolicyName) != linkutils.InlineAccessPolicyName() {
		return false
	}
	var doc struct {
		Statement []struct {
			Sid      string
			Action   []string
			Resource []string
		}
	}
	if err := json.Unmarshal([]byte(aws.ToString(input.PolicyDocument)), &doc); err != nil {
		return false
	}
	wantResource := "arn:aws:rds-db:us-west-2:123456789012:dbuser:" + testClusterResourceID + "/" + dbUser
	for _, statement := range doc.Statement {
		if statement.Sid != fcConnectSID {
			continue
		}
		return len(statement.Action) == 1 && statement.Action[0] == "rds-db:connect" &&
			len(statement.Resource) == 1 && statement.Resource[0] == wantResource
	}
	return false
}

func fcMatchConnectOutput(
	actual *provider.LinkUpdateIntermediaryResourcesOutput,
) (plugintestutils.EqualityCheckValues, error) {
	summary := map[string]any{}
	if actual != nil {
		summary["hasStatement"] = actual.LinkData != nil &&
			actual.LinkData.Fields[fcExecRole] != nil &&
			actual.LinkData.Fields[fcExecRole].Fields[linkutils.PermissionFieldName] != nil
	}
	return plugintestutils.EqualityCheckValues{
		Expected: map[string]any{"hasStatement": true},
		Actual:   summary,
	}, nil
}

func TestFunctionClusterLinkUpdateSuite(t *testing.T) {
	suite.Run(t, new(FunctionClusterLinkUpdateSuite))
}

// The networking activation resolves the flex VPC's Bluelink name from the AWS VPC's
// tag before it can find the resource in state, so the mock has to answer it.
func flexVPCDescribeOutput() []*ec2.DescribeVpcsOutput {
	return []*ec2.DescribeVpcsOutput{
		{
			Vpcs: []ec2types.Vpc{
				{
					VpcId: aws.String("vpc-1"),
					Tags: []ec2types.Tag{
						{
							Key:   aws.String(flex.TagFlexVPCName),
							Value: aws.String("orders-vpc"),
						},
					},
				},
			},
		},
	}
}
