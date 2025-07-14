package iam

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/newstack-cloud/bluelink-provider-aws/internal/testutils"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/schema"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
	"github.com/stretchr/testify/suite"
)

type IAMOidcProviderResourceUpdateSuite struct {
	suite.Suite
}

func (s *IAMOidcProviderResourceUpdateSuite) Test_update_iam_oidc_provider() {
	loader := &testutils.MockAWSConfigLoader{}
	providerCtx := plugintestutils.NewTestProviderContext(
		"aws",
		map[string]*core.ScalarValue{
			"region": core.ScalarFromString("us-west-2"),
		},
		map[string]*core.ScalarValue{
			"session_id": core.ScalarFromString("test-session-id"),
		},
	)

	testCases := []plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		updateOidcProviderClientIdsTestCase(providerCtx, loader),
		updateOidcProviderThumbprintsTestCase(providerCtx, loader),
		updateOidcProviderTagsTestCase(providerCtx, loader),
		updateOidcProviderNoChangesTestCase(providerCtx, loader),
		updateOidcProviderServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		OidcProviderResource,
		&s.Suite,
	)
}

func updateOidcProviderClientIdsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithAddClientIDToOpenIDConnectProviderOutput(&iam.AddClientIDToOpenIDConnectProviderOutput{}),
		iammock.WithRemoveClientIDFromOpenIDConnectProviderOutput(&iam.RemoveClientIDFromOpenIDConnectProviderOutput{}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
			"clientIdList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("old-client-id"),
					core.MappingNodeFromString("keep-client-id"),
				},
			},
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
			"clientIdList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("keep-client-id"),
					core.MappingNodeFromString("new-client-id"),
				},
			},
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider client IDs",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOidcProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.clientIdList",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{},
		SaveActionsCalled: []plugintestutils.SaveActionCalled{
			{
				SaveAction: "update OIDC provider client IDs",
			},
		},
	}
}

func updateOidcProviderThumbprintsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateOpenIDConnectProviderThumbprintOutput(&iam.UpdateOpenIDConnectProviderThumbprintOutput{}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
					core.MappingNodeFromString("9e99a48a9960b14926bb7f3b02e22da2b0ab7280"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider thumbprints",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOidcProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.thumbprintList",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{},
		SaveActionsCalled: []plugintestutils.SaveActionCalled{
			{
				SaveAction: "update OIDC provider thumbprints",
			},
		},
	}
}

func updateOidcProviderTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithTagOpenIDConnectProviderOutput(&iam.TagOpenIDConnectProviderOutput{}),
		iammock.WithUntagOpenIDConnectProviderOutput(&iam.UntagOpenIDConnectProviderOutput{}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
			"tags": &core.MappingNode{
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Development"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("ToRemove"),
							"value": core.MappingNodeFromString("OldValue"),
						},
					},
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
			"tags": &core.MappingNode{
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Production"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Service"),
							"value": core.MappingNodeFromString("Authentication"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider tags",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOidcProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.tags",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{},
		SaveActionsCalled: []plugintestutils.SaveActionCalled{
			{
				SaveAction: "update OIDC provider tags",
			},
		},
	}
}

func updateOidcProviderNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock()

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
		},
	}

	// Updated state - same as current
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider no changes",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOidcProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput:    &provider.ResourceDeployOutput{},
		SaveActionsCalled: []plugintestutils.SaveActionCalled{},
	}
}

func updateOidcProviderServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	oidcProviderArn := "arn:aws:iam::123456789012:oidc-provider/example.com"

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateOpenIDConnectProviderThumbprintError(fmt.Errorf("failed to update thumbprints")),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString(oidcProviderArn),
			"url": core.MappingNodeFromString("https://example.com"),
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("cf23df2207d99a74fbe169e3eba035e633b65d94"),
				},
			},
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"url": core.MappingNodeFromString("https://example.com"),
			"thumbprintList": &core.MappingNode{
				Items: []*core.MappingNode{
					core.MappingNodeFromString("9e99a48a9960b14926bb7f3b02e22da2b0ab7280"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM OIDC provider service error",
		ServiceFactory: func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
			return service
		},
		ServiceMockCalls: &service.MockCalls,
		ConfigStore: utils.NewAWSConfigStore(
			[]string{},
			utils.AWSConfigFromProviderContext,
			loader,
			utils.AWSConfigCacheKey,
		),
		Input: &provider.ResourceDeployInput{
			InstanceID: "test-instance-id",
			ResourceID: "test-oidc-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-oidc-provider-id",
					ResourceName: "TestOidcProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-oidc-provider-id",
						Name:       "TestOidcProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/oidcProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.thumbprintList",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedError: "failed to update thumbprints",
	}
}

func TestIAMOidcProviderResourceUpdate(t *testing.T) {
	suite.Run(t, new(IAMOidcProviderResourceUpdateSuite))
}