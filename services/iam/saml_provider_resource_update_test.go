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

type IAMSAMLProviderResourceUpdateSuite struct {
	suite.Suite
}

func (s *IAMSAMLProviderResourceUpdateSuite) Test_update_iam_saml_provider() {
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
		updateSAMLProviderMetadataTestCase(providerCtx, loader),
		updateSAMLProviderTagsTestCase(providerCtx, loader),
		updateSAMLProviderNoChangesTestCase(providerCtx, loader),
		updateSAMLProviderServiceErrorTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		SAMLProviderResource,
		&s.Suite,
	)
}

func updateSAMLProviderMetadataTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"

	oldMetadata := `<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" 
                  entityID="http://old.example.com/saml">
    <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
                           Location="https://old.example.com/saml/sso"/>
    </IDPSSODescriptor>
</EntityDescriptor>`

	newMetadata := `<?xml version="1.0"?>
<EntityDescriptor xmlns="urn:oasis:names:tc:SAML:2.0:metadata" 
                  entityID="http://new.example.com/saml">
    <IDPSSODescriptor protocolSupportEnumeration="urn:oasis:names:tc:SAML:2.0:protocol">
        <SingleSignOnService Binding="urn:oasis:names:tc:SAML:2.0:bindings:HTTP-POST" 
                           Location="https://new.example.com/saml/sso"/>
    </IDPSSODescriptor>
</EntityDescriptor>`

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateSAMLProviderOutput(&iam.UpdateSAMLProviderOutput{
			SAMLProviderArn: aws.String(samlProviderArn),
		}),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                  core.MappingNodeFromString(samlProviderArn),
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(oldMetadata),
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(newMetadata),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM SAML provider metadata",
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
			ResourceID: "test-saml-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-saml-provider-id",
					ResourceName: "TestSAMLProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-saml-provider-id",
						Name:       "TestSAMLProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/samlProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.samlMetadataDocument",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{},
		},
		SaveActionsCalled: map[string]any{
			"UpdateSAMLProvider": &iam.UpdateSAMLProviderInput{
				SAMLProviderArn:      aws.String(samlProviderArn),
				SAMLMetadataDocument: aws.String(newMetadata),
			},
		},
	}
}

func updateSAMLProviderTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"
	metadata := `<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`

	service := iammock.CreateIamServiceMock(
		iammock.WithTagSAMLProviderOutput(&iam.TagSAMLProviderOutput{}),
		iammock.WithUntagSAMLProviderOutput(&iam.UntagSAMLProviderOutput{}),
	)

	// Current state with tags
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                  core.MappingNodeFromString(samlProviderArn),
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(metadata),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Development"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("OldTag"),
							"value": core.MappingNodeFromString("OldValue"),
						},
					},
				},
			},
		},
	}

	// Updated state with different tags
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(metadata),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("Production"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("NewTag"),
							"value": core.MappingNodeFromString("NewValue"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM SAML provider tags",
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
			ResourceID: "test-saml-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-saml-provider-id",
					ResourceName: "TestSAMLProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-saml-provider-id",
						Name:       "TestSAMLProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/samlProvider",
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
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{},
		},
		SaveActionsCalled: map[string]any{
			"UntagSAMLProvider": &iam.UntagSAMLProviderInput{
				SAMLProviderArn: aws.String(samlProviderArn),
				TagKeys:         []string{"OldTag"},
			},
			"TagSAMLProvider": &iam.TagSAMLProviderInput{
				SAMLProviderArn: aws.String(samlProviderArn),
				Tags: []types.Tag{
					{
						Key:   aws.String("Environment"),
						Value: aws.String("Production"),
					},
					{
						Key:   aws.String("NewTag"),
						Value: aws.String("NewValue"),
					},
				},
			},
		},
	}
}

func updateSAMLProviderNoChangesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"
	metadata := `<?xml version="1.0"?><EntityDescriptor>...</EntityDescriptor>`

	service := iammock.CreateIamServiceMock()

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                  core.MappingNodeFromString(samlProviderArn),
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(metadata),
		},
	}

	// No changes in updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(metadata),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM SAML provider with no changes",
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
			ResourceID: "test-saml-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-saml-provider-id",
					ResourceName: "TestSAMLProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-saml-provider-id",
						Name:       "TestSAMLProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/samlProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
	}
}

func updateSAMLProviderServiceErrorTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	samlProviderArn := "arn:aws:iam::123456789012:saml-provider/MySAMLProvider"
	oldMetadata := `<?xml version="1.0"?><EntityDescriptor>old</EntityDescriptor>`
	newMetadata := `<?xml version="1.0"?><EntityDescriptor>new</EntityDescriptor>`

	serviceError := fmt.Errorf("AWS service error")

	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateSAMLProviderError(serviceError),
	)

	// Current state
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                  core.MappingNodeFromString(samlProviderArn),
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(oldMetadata),
		},
	}

	// Updated state
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"name":                 core.MappingNodeFromString("MySAMLProvider"),
			"samlMetadataDocument": core.MappingNodeFromString(newMetadata),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "Update IAM SAML provider with service error",
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
			ResourceID: "test-saml-provider-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-saml-provider-id",
					ResourceName: "TestSAMLProvider",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-saml-provider-id",
						Name:       "TestSAMLProvider",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/samlProvider",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.samlMetadataDocument",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		ExpectError: true,
		SaveActionsCalled: map[string]any{
			"UpdateSAMLProvider": &iam.UpdateSAMLProviderInput{
				SAMLProviderArn:      aws.String(samlProviderArn),
				SAMLMetadataDocument: aws.String(newMetadata),
			},
		},
	}
}

func TestIAMSAMLProviderResourceUpdateSuite(t *testing.T) {
	suite.Run(t, new(IAMSAMLProviderResourceUpdateSuite))
}
