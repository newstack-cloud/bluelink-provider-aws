//go:build unit

package lambdards

import (
	"fmt"
	"maps"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	cloudcontrolservice "github.com/newstack-cloud/bluelink-provider-aws/services/cloudcontrol/service"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/stretchr/testify/suite"
)

type FunctionClusterLinkStageChangesSuite struct {
	suite.Suite
}

const (
	fcStageHostPath   = "apiFunction.environmentVariables[\"" + fcPrefix + "_HOST\"]"
	fcStageReaderPath = "apiFunction.environmentVariables[\"" + fcPrefix + "_READER_HOST\"]"
)

func functionClusterStageLinkFactory() func(
	pluginutils.LinkServiceDeps[*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service],
) provider.Link {
	return functionClusterLinkFactory(iammock.CreateIamServiceMock(), noopEC2ServiceFactory())
}

func clusterCaller(authMode string, extra map[string]*core.MappingNode) provider.ResourceInfo {
	fields := map[string]*core.MappingNode{
		"aws.lambda.rds.ordersCluster.envVarPrefix": core.MappingNodeFromString(fcPrefix),
	}
	if authMode != "" {
		fields["aws.lambda.rds.ordersCluster.authMode"] = core.MappingNodeFromString(authMode)
	}
	maps.Copy(fields, extra)
	return provider.ResourceInfo{
		ResourceName: "apiFunction",
		ResourceWithResolvedSubs: &provider.ResolvedResource{
			Metadata: &provider.ResolvedResourceMetadata{
				Annotations: &core.MappingNode{Fields: fields},
			},
		},
	}
}

func clusterCallerVPC(authMode string) provider.ResourceInfo {
	caller := clusterCaller(authMode, nil)
	caller.ResourceWithResolvedSubs.Spec = core.MappingNodeFields(
		"vpcConfig", core.MappingNodeFields(
			"subnetIds", &core.MappingNode{Items: []*core.MappingNode{
				core.MappingNodeFromString("subnet-1"),
			}},
		),
	)
	return caller
}

func clusterResourceBChanges() *provider.Changes {
	return &provider.Changes{
		AppliedResourceInfo: provider.ResourceInfo{
			ResourceName: "ordersCluster",
			ResourceWithResolvedSubs: &provider.ResolvedResource{
				Spec: core.MappingNodeFields(
					"endpoint", core.MappingNodeFields(
						"address", core.MappingNodeFromString(testClusterEndpoint),
					),
					"readEndpoint", core.MappingNodeFields(
						"address", core.MappingNodeFromString(testClusterReaderEndpoint),
					),
				),
			},
		},
		NewFields: []provider.FieldChange{
			{FieldPath: "spec.endpoint.address", NewValue: core.MappingNodeFromString(testClusterEndpoint)},
		},
	}
}

func (s *FunctionClusterLinkStageChangesSuite) Test_stage_changes() {
	testCases := []plugintestutils.LinkChangeStagingTestCase[
		*aws.Config, lambdaservice.Service, *aws.Config, cloudcontrolservice.Service,
	]{
		{
			Name: "stages host env var change from the cluster writer endpoint",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: clusterCaller("password", nil)},
				ResourceBChanges: clusterResourceBChanges(),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: fcStageHostPath, NewValue: core.MappingNodeFromString(testClusterEndpoint)},
					},

					// The variables this link declares it will contribute.
					ResourceDataMappings: fcDeclaredContributions(),
				},
			},
		},
		{
			Name: "stages reader host env var when readerEndpoint is enabled",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: clusterCaller("password", map[string]*core.MappingNode{
					"aws.lambda.rds.ordersCluster.readerEndpoint": core.MappingNodeFromBool(true),
				})},
				ResourceBChanges: clusterResourceBChanges(),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: fcStageHostPath, NewValue: core.MappingNodeFromString(testClusterEndpoint)},
						{FieldPath: fcStageReaderPath, NewValue: core.MappingNodeFromString(testClusterReaderEndpoint)},
					},

					// The variables this link declares it will contribute.
					ResourceDataMappings: fcDeclaredContributions(),
				},
			},
		},
		{
			Name: "adds exec-role known-on-deploy for iam auth on a new cluster",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: clusterCaller("iam", nil)},
				ResourceBChanges: clusterResourceBChanges(),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: fcStageHostPath, NewValue: core.MappingNodeFromString(testClusterEndpoint)},
					},
					FieldChangesKnownOnDeploy: []string{"apiFunctionExecutionRole"},

					// The variables this link declares it will contribute.
					ResourceDataMappings: fcDeclaredContributions(),
				},
			},
		},
		{
			Name: "surfaces network access known-on-deploy for a VPC-attached function",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: clusterCallerVPC("password")},
				ResourceBChanges: clusterResourceBChanges(),
				CurrentLinkState: &state.LinkState{LinkID: "test-link", Data: map[string]*core.MappingNode{}},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					NewFields: []*provider.FieldChange{
						{FieldPath: fcStageHostPath, NewValue: core.MappingNodeFromString(testClusterEndpoint)},
					},
					FieldChangesKnownOnDeploy: []string{"apiFunctionNetworkAccess"},

					// The variables this link declares it will contribute.
					ResourceDataMappings: fcDeclaredContributions(),
				},
			},
		},
		{
			Name: "stages no change when the host env var already matches",
			Input: &provider.LinkStageChangesInput{
				ResourceAChanges: &provider.Changes{AppliedResourceInfo: clusterCaller("password", nil)},
				ResourceBChanges: clusterResourceBChanges(),
				CurrentLinkState: &state.LinkState{
					LinkID: "test-link",
					Data: map[string]*core.MappingNode{
						"apiFunction": {
							Fields: map[string]*core.MappingNode{
								"environmentVariables": {
									Fields: map[string]*core.MappingNode{
										fcPrefix + "_HOST": core.MappingNodeFromString(testClusterEndpoint),
									},
								},
							},
						},
					},
				},
			},
			ExpectedOutput: &provider.LinkStageChangesOutput{
				Changes: &provider.LinkChanges{
					UnchangedFields: []string{fcStageHostPath},

					// The variables this link declares it will contribute.
					ResourceDataMappings: fcDeclaredContributions(),
				},
			},
		},
	}

	plugintestutils.RunLinkChangeStagingTestCases(
		testCases,
		functionClusterStageLinkFactory(),
		&s.Suite,
	)
}

func TestFunctionClusterLinkStageChangesSuite(t *testing.T) {
	suite.Run(t, new(FunctionClusterLinkStageChangesSuite))
}

// The variables the cluster link declares it will contribute to the function, which staging
// records so the framework knows the function is waiting on this link.
func fcDeclaredContributions() map[string]string {
	declared := map[string]string{}
	for _, name := range clusterEnvVarNames(fcPrefix).all() {
		declared[fmt.Sprintf("apiFunction::spec.environment.variables[%q]", name)] =
			fmt.Sprintf("apiFunction.environmentVariables[%q]", name)
	}

	return declared
}
