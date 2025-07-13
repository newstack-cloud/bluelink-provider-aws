package iam

import (
	"net/url"
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

// Helper to compare two []types.Tag slices as sets (ignoring order)
func assertTagsEqualIgnoreOrder(t *testing.T, expected, actual []types.Tag) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("tag slices have different lengths: expected %d, got %d", len(expected), len(actual))
	}
	tagMap := make(map[string]string, len(expected))
	for _, tag := range expected {
		tagMap[*tag.Key] = *tag.Value
	}
	for _, tag := range actual {
		val, ok := tagMap[*tag.Key]
		if !ok {
			t.Fatalf("unexpected tag key: %s", *tag.Key)
		}
		if val != *tag.Value {
			t.Fatalf("tag value mismatch for key %s: expected %s, got %s", *tag.Key, val, *tag.Value)
		}
	}
}

type IamRoleResourceUpdateSuite struct {
	suite.Suite
}

func (s *IamRoleResourceUpdateSuite) Test_update_iam_role() {
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
		updateRoleDescriptionTestCase(providerCtx, loader),
		updateRoleAssumeRolePolicyTestCase(providerCtx, loader),
		updateRoleMaxSessionDurationTestCase(providerCtx, loader),
		updateRoleInlinePoliciesTestCase(providerCtx, loader),
		updateRoleManagedPoliciesTestCase(providerCtx, loader),
		updateRoleTagsTestCase(providerCtx, loader, s),
		updateRoleRemoveInlinePoliciesTestCase(providerCtx, loader),
		updateRoleDetachManagedPoliciesTestCase(providerCtx, loader),
		updateRolePermissionsBoundaryTestCase(providerCtx, loader),
		updateRoleRemovePermissionsBoundaryTestCase(providerCtx, loader),
	}

	plugintestutils.RunResourceDeployTestCases(
		testCases,
		RoleResource,
		&s.Suite,
	)
}

func updateRoleDescriptionTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateRoleOutput(&iam.UpdateRoleOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName:    aws.String("TestRole"),
				Arn:         aws.String("arn:aws:iam::123456789012:role/TestRole"),
				RoleId:      aws.String("AROA1234567890123456"),
				Description: aws.String("Updated test role description"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Create current state data
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
			"description": core.MappingNodeFromString("Original test role description"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create updated spec data (changing description)
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName":    core.MappingNodeFromString("TestRole"),
			"description": core.MappingNodeFromString("Updated test role description"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role description",
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
			ResourceID: "TestRole",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "TestRole",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "TestRole",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.description",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"UpdateRole": &iam.UpdateRoleInput{
				RoleName:    aws.String("TestRole"),
				Description: aws.String("Updated test role description"),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRoleAssumeRolePolicyTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateAssumeRolePolicyOutput(&iam.UpdateAssumeRolePolicyOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("TestRole"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "ec2.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Create current state data
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":         core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
			"description": core.MappingNodeFromString("Original test role description"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create updated spec data (changing assume role policy)
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("TestRole"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("ec2.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role assume role policy",
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
			ResourceID: "TestRole",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "TestRole",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "TestRole",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.assumeRolePolicyDocument",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"UpdateAssumeRolePolicy": &iam.UpdateAssumeRolePolicyInput{
				RoleName:       aws.String("TestRole"),
				PolicyDocument: aws.String(url.QueryEscape(`{"Statement":[{"Action":"sts:AssumeRole","Effect":"Allow","Principal":{"Service":"ec2.amazonaws.com"}}],"Version":"2012-10-17"}`)),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRoleMaxSessionDurationTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithUpdateRoleOutput(&iam.UpdateRoleOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName:           aws.String("TestRole"),
				Arn:                aws.String("arn:aws:iam::123456789012:role/TestRole"),
				RoleId:             aws.String("AROA1234567890123456"),
				MaxSessionDuration: aws.Int32(7200),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Create current state data
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
			"maxSessionDuration": core.MappingNodeFromInt(3600),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
		},
	}

	// Create updated spec data (changing max session duration)
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName":           core.MappingNodeFromString("TestRole"),
			"maxSessionDuration": core.MappingNodeFromInt(7200),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role max session duration",
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
			ResourceID: "TestRole",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "TestRole",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "TestRole",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.maxSessionDuration",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"UpdateRole": &iam.UpdateRoleInput{
				RoleName:           aws.String("TestRole"),
				MaxSessionDuration: aws.Int32(7200),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRoleInlinePoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("TestRole"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Create current state data with original policies
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
			"policies": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("OriginalPolicy"),
							"policyDocument": {
								Fields: map[string]*core.MappingNode{
									"Version": core.MappingNodeFromString("2012-10-17"),
									"Statement": {
										Items: []*core.MappingNode{
											{
												Fields: map[string]*core.MappingNode{
													"Effect": core.MappingNodeFromString("Allow"),
													"Action": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("s3:GetObject"),
														},
													},
													"Resource": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("arn:aws:s3:::original-bucket/*"),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Create updated spec data with modified policies
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("TestRole"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
			"policies": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("UpdatedPolicy"),
							"policyDocument": {
								Fields: map[string]*core.MappingNode{
									"Version": core.MappingNodeFromString("2012-10-17"),
									"Statement": {
										Items: []*core.MappingNode{
											{
												Fields: map[string]*core.MappingNode{
													"Effect": core.MappingNodeFromString("Allow"),
													"Action": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("s3:GetObject"),
															core.MappingNodeFromString("s3:PutObject"),
														},
													},
													"Resource": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("arn:aws:s3:::updated-bucket/*"),
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role inline policies",
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
			ResourceID: "TestRole",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "TestRole",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "TestRole",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.policies",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"PutRolePolicy": &iam.PutRolePolicyInput{
				RoleName:       aws.String("TestRole"),
				PolicyName:     aws.String("UpdatedPolicy"),
				PolicyDocument: aws.String(`{"Statement":[{"Action":["s3:GetObject","s3:PutObject"],"Effect":"Allow","Resource":["arn:aws:s3:::updated-bucket/*"]}],"Version":"2012-10-17"}`),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRoleManagedPoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithAttachRolePolicyOutput(&iam.AttachRolePolicyOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("TestRole"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/TestRole"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Create current state data with original managed policies
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn": core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
				},
			},
		},
	}

	// Create updated spec data with modified managed policies
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("TestRole"),
			"assumeRolePolicyDocument": {
				Fields: map[string]*core.MappingNode{
					"Version": core.MappingNodeFromString("2012-10-17"),
					"Statement": {
						Items: []*core.MappingNode{
							{
								Fields: map[string]*core.MappingNode{
									"Effect": core.MappingNodeFromString("Allow"),
									"Principal": {
										Fields: map[string]*core.MappingNode{
											"Service": core.MappingNodeFromString("lambda.amazonaws.com"),
										},
									},
									"Action": core.MappingNodeFromString("sts:AssumeRole"),
								},
							},
						},
					},
				},
			},
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
					core.MappingNodeFromString("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role managed policies",
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
			ResourceID: "TestRole",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "TestRole",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "TestRole",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.managedPolicyArns",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"AttachRolePolicy": &iam.AttachRolePolicyInput{
				RoleName:  aws.String("TestRole"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/service-role/AWSLambdaVPCAccessExecutionRole"),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("TestRole"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/TestRole"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRoleTagsTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
	suite *IamRoleResourceUpdateSuite,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithUntagRoleOutput(&iam.UntagRoleOutput{}),
		iammock.WithTagRoleOutput(&iam.TagRoleOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("test-role"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/test-role"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Current state with existing tags
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"roleName": core.MappingNodeFromString("test-role"),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("test"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("OldTag"),
							"value": core.MappingNodeFromString("to-remove"),
						},
					},
				},
			},
		},
	}

	// Updated state with modified tags
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role"),
			"tags": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("Environment"),
							"value": core.MappingNodeFromString("production"),
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"key":   core.MappingNodeFromString("NewTag"),
							"value": core.MappingNodeFromString("added"),
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role tags",
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
			ResourceID: "test-role-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-id",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-role-id",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
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
		SaveActionsCalled: map[string]any{
			"UntagRole": &iam.UntagRoleInput{
				RoleName: aws.String("test-role"),
				TagKeys:  []string{"OldTag"},
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("test-role"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRoleRemoveInlinePoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteRolePolicyOutput(&iam.DeleteRolePolicyOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("test-role"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/test-role"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Current state with multiple policies
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"roleName": core.MappingNodeFromString("test-role"),
			"policies": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("KeepPolicy"),
							"policyDocument": {
								Fields: map[string]*core.MappingNode{
									"Version": core.MappingNodeFromString("2012-10-17"),
									"Statement": {
										Items: []*core.MappingNode{
											{
												Fields: map[string]*core.MappingNode{
													"Effect": core.MappingNodeFromString("Allow"),
													"Action": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("s3:GetObject"),
														},
													},
													"Resource": core.MappingNodeFromString("*"),
												},
											},
										},
									},
								},
							},
						},
					},
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("RemovePolicy"),
							"policyDocument": {
								Fields: map[string]*core.MappingNode{
									"Version": core.MappingNodeFromString("2012-10-17"),
									"Statement": {
										Items: []*core.MappingNode{
											{
												Fields: map[string]*core.MappingNode{
													"Effect": core.MappingNodeFromString("Allow"),
													"Action": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("ec2:DescribeInstances"),
														},
													},
													"Resource": core.MappingNodeFromString("*"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	// Updated state with only one policy (removed the other)
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role"),
			"policies": {
				Items: []*core.MappingNode{
					{
						Fields: map[string]*core.MappingNode{
							"policyName": core.MappingNodeFromString("KeepPolicy"),
							"policyDocument": {
								Fields: map[string]*core.MappingNode{
									"Version": core.MappingNodeFromString("2012-10-17"),
									"Statement": {
										Items: []*core.MappingNode{
											{
												Fields: map[string]*core.MappingNode{
													"Effect": core.MappingNodeFromString("Allow"),
													"Action": {
														Items: []*core.MappingNode{
															core.MappingNodeFromString("s3:GetObject"),
														},
													},
													"Resource": core.MappingNodeFromString("*"),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role remove inline policies",
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
			ResourceID: "test-role-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-id",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-role-id",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.policies",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"DeleteRolePolicy": &iam.DeleteRolePolicyInput{
				RoleName:   aws.String("test-role"),
				PolicyName: aws.String("RemovePolicy"),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("test-role"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRoleDetachManagedPoliciesTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDetachRolePolicyOutput(&iam.DetachRolePolicyOutput{}),
		iammock.WithAttachRolePolicyOutput(&iam.AttachRolePolicyOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("test-role"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/test-role"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Current state with multiple managed policies
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"roleName": core.MappingNodeFromString("test-role"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					core.MappingNodeFromString("arn:aws:iam::aws:policy/PowerUserAccess"),
				},
			},
		},
	}

	// Updated state with different managed policies
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role"),
			"managedPolicyArns": {
				Items: []*core.MappingNode{
					core.MappingNodeFromString("arn:aws:iam::aws:policy/ReadOnlyAccess"),
					core.MappingNodeFromString("arn:aws:iam::aws:policy/AdministratorAccess"),
				},
			},
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role detach managed policies",
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
			ResourceID: "test-role-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-id",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-role-id",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.managedPolicyArns",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"DetachRolePolicy": &iam.DetachRolePolicyInput{
				RoleName:  aws.String("test-role"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/PowerUserAccess"),
			},
			"AttachRolePolicy": &iam.AttachRolePolicyInput{
				RoleName:  aws.String("test-role"),
				PolicyArn: aws.String("arn:aws:iam::aws:policy/AdministratorAccess"),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("test-role"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRolePermissionsBoundaryTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithPutRolePermissionsBoundaryOutput(&iam.PutRolePermissionsBoundaryOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("test-role"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/test-role"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Current state with no permissions boundary
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":      core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"roleName": core.MappingNodeFromString("test-role"),
		},
	}

	// Updated state with permissions boundary
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName":            core.MappingNodeFromString("test-role"),
			"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/test-boundary"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role permissions boundary",
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
			ResourceID: "test-role-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-id",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-role-id",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.permissionsBoundary",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"PutRolePermissionsBoundary": &iam.PutRolePermissionsBoundaryInput{
				RoleName:            aws.String("test-role"),
				PermissionsBoundary: aws.String("arn:aws:iam::123456789012:policy/test-boundary"),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("test-role"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func updateRoleRemovePermissionsBoundaryTestCase(
	providerCtx provider.Context,
	loader *testutils.MockAWSConfigLoader,
) plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service] {
	service := iammock.CreateIamServiceMock(
		iammock.WithDeleteRolePermissionsBoundaryOutput(&iam.DeleteRolePermissionsBoundaryOutput{}),
		iammock.WithGetRoleOutput(&iam.GetRoleOutput{
			Role: &types.Role{
				RoleName: aws.String("test-role"),
				Arn:      aws.String("arn:aws:iam::123456789012:role/test-role"),
				RoleId:   aws.String("AROA1234567890123456"),
				AssumeRolePolicyDocument: aws.String(`{
					"Version": "2012-10-17",
					"Statement": [{
						"Effect": "Allow",
						"Principal": {"Service": "lambda.amazonaws.com"},
						"Action": "sts:AssumeRole"
					}]
				}`),
			},
		}),
	)

	// Current state with permissions boundary
	currentStateSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"arn":                 core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
			"roleName":            core.MappingNodeFromString("test-role"),
			"permissionsBoundary": core.MappingNodeFromString("arn:aws:iam::123456789012:policy/test-boundary"),
		},
	}

	// Updated state with no permissions boundary
	updatedSpecData := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString("test-role"),
		},
	}

	return plugintestutils.ResourceDeployTestCase[*aws.Config, iamservice.Service]{
		Name: "update role remove permissions boundary",
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
			ResourceID: "test-role-id",
			Changes: &provider.Changes{
				AppliedResourceInfo: provider.ResourceInfo{
					ResourceID:   "test-role-id",
					ResourceName: "TestRole",
					InstanceID:   "test-instance-id",
					CurrentResourceState: &state.ResourceState{
						ResourceID: "test-role-id",
						Name:       "TestRole",
						InstanceID: "test-instance-id",
						SpecData:   currentStateSpecData,
					},
					ResourceWithResolvedSubs: &provider.ResolvedResource{
						Type: &schema.ResourceTypeWrapper{
							Value: "aws/iam/role",
						},
						Spec: updatedSpecData,
					},
				},
				ModifiedFields: []provider.FieldChange{
					{
						FieldPath: "spec.permissionsBoundary",
					},
				},
			},
			ProviderContext: providerCtx,
		},
		SaveActionsCalled: map[string]any{
			"DeleteRolePermissionsBoundary": &iam.DeleteRolePermissionsBoundaryInput{
				RoleName: aws.String("test-role"),
			},
			"GetRole": &iam.GetRoleInput{
				RoleName: aws.String("test-role"),
			},
		},
		ExpectedOutput: &provider.ResourceDeployOutput{
			ComputedFieldValues: map[string]*core.MappingNode{
				"spec.arn":    core.MappingNodeFromString("arn:aws:iam::123456789012:role/test-role"),
				"spec.roleId": core.MappingNodeFromString("AROA1234567890123456"),
			},
		},
	}
}

func (s *IamRoleResourceUpdateSuite) AfterTest(suiteName, testName string) {
	// Remove the problematic AfterTest method that tries to access unexported fields
}

func TestIamRoleResourceUpdate(t *testing.T) {
	suite.Run(t, new(IamRoleResourceUpdateSuite))
}
