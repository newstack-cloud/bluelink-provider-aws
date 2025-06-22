package iammock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	iamservice "github.com/newstack-cloud/celerity-provider-aws/services/iam/service"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
	"github.com/newstack-cloud/celerity/libs/plugin-framework/sdk/plugintestutils"
)

type iamServiceMock struct {
	plugintestutils.MockCalls

	// Role-related mock fields
	createRoleOutput *iam.CreateRoleOutput
	createRoleError  error
	getRoleOutput    *iam.GetRoleOutput
	getRoleError     error
	updateRoleOutput *iam.UpdateRoleOutput
	updateRoleError  error
	deleteRoleOutput *iam.DeleteRoleOutput
	deleteRoleError  error

	// Role policy document-related mock fields
	updateAssumeRolePolicyOutput *iam.UpdateAssumeRolePolicyOutput
	updateAssumeRolePolicyError  error

	// Role policy attachment-related mock fields
	attachRolePolicyOutput         *iam.AttachRolePolicyOutput
	attachRolePolicyError          error
	detachRolePolicyOutput         *iam.DetachRolePolicyOutput
	detachRolePolicyError          error
	listAttachedRolePoliciesOutput *iam.ListAttachedRolePoliciesOutput
	listAttachedRolePoliciesError  error

	// Inline policy-related mock fields
	putRolePolicyOutput    *iam.PutRolePolicyOutput
	putRolePolicyError     error
	deleteRolePolicyOutput *iam.DeleteRolePolicyOutput
	deleteRolePolicyError  error
	listRolePoliciesOutput *iam.ListRolePoliciesOutput
	listRolePoliciesError  error
	getRolePolicyOutput    *iam.GetRolePolicyOutput
	getRolePolicyError     error

	// Tag-related mock fields
	tagRoleOutput      *iam.TagRoleOutput
	tagRoleError       error
	untagRoleOutput    *iam.UntagRoleOutput
	untagRoleError     error
	listRoleTagsOutput *iam.ListRoleTagsOutput
	listRoleTagsError  error

	// Permissions boundary-related mock fields
	putRolePermissionsBoundaryOutput    *iam.PutRolePermissionsBoundaryOutput
	putRolePermissionsBoundaryError     error
	deleteRolePermissionsBoundaryOutput *iam.DeleteRolePermissionsBoundaryOutput
	deleteRolePermissionsBoundaryError  error

	// User-related mock fields
	createUserOutput *iam.CreateUserOutput
	createUserError  error
	getUserOutput    *iam.GetUserOutput
	getUserError     error
	updateUserOutput *iam.UpdateUserOutput
	updateUserError  error
	deleteUserOutput *iam.DeleteUserOutput
	deleteUserError  error

	// User policy attachment-related mock fields
	attachUserPolicyOutput         *iam.AttachUserPolicyOutput
	attachUserPolicyError          error
	detachUserPolicyOutput         *iam.DetachUserPolicyOutput
	detachUserPolicyError          error
	listAttachedUserPoliciesOutput *iam.ListAttachedUserPoliciesOutput
	listAttachedUserPoliciesError  error

	// User policy operations
	putUserPolicyOutput    *iam.PutUserPolicyOutput
	putUserPolicyError     error
	deleteUserPolicyOutput *iam.DeleteUserPolicyOutput
	deleteUserPolicyError  error
	listUserPoliciesOutput *iam.ListUserPoliciesOutput
	listUserPoliciesError  error

	// User tag operations
	tagUserOutput      *iam.TagUserOutput
	tagUserError       error
	untagUserOutput    *iam.UntagUserOutput
	untagUserError     error
	listUserTagsOutput *iam.ListUserTagsOutput
	listUserTagsError  error

	// User permissions boundary operations
	putUserPermissionsBoundaryOutput    *iam.PutUserPermissionsBoundaryOutput
	putUserPermissionsBoundaryError     error
	deleteUserPermissionsBoundaryOutput *iam.DeleteUserPermissionsBoundaryOutput
	deleteUserPermissionsBoundaryError  error

	// User policy operations
	getUserPolicyOutput *iam.GetUserPolicyOutput
	getUserPolicyError  error

	// Group operations
	addUserToGroupOutput      *iam.AddUserToGroupOutput
	addUserToGroupError       error
	removeUserFromGroupOutput *iam.RemoveUserFromGroupOutput
	removeUserFromGroupError  error
	listGroupsForUserOutput   *iam.ListGroupsForUserOutput
	listGroupsForUserError    error

	// Login profile operations
	createLoginProfileOutput *iam.CreateLoginProfileOutput
	createLoginProfileError  error
	getLoginProfileOutput    *iam.GetLoginProfileOutput
	getLoginProfileError     error
	updateLoginProfileOutput *iam.UpdateLoginProfileOutput
	updateLoginProfileError  error
	deleteLoginProfileOutput *iam.DeleteLoginProfileOutput
	deleteLoginProfileError  error
}

type iamServiceMockOption func(*iamServiceMock)

func CreateIamServiceMockFactory(
	opts ...iamServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
	mock := CreateIamServiceMock(opts...)
	return func(awsConfig *aws.Config, providerContext provider.Context) iamservice.Service {
		return mock
	}
}

func CreateIamServiceMock(
	opts ...iamServiceMockOption,
) *iamServiceMock {
	mock := &iamServiceMock{}

	for _, opt := range opts {
		opt(mock)
	}

	return mock
}

// NewMockService creates a new mock Service for testing.
func NewMockService(t any) *iamServiceMock {
	return CreateIamServiceMock()
}

// Mock configuration options for Role operations

func WithCreateRoleOutput(output *iam.CreateRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createRoleOutput = output
	}
}

func WithCreateRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createRoleError = err
	}
}

func WithGetRoleOutput(output *iam.GetRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getRoleOutput = output
	}
}

func WithGetRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getRoleError = err
	}
}

func WithUpdateRoleOutput(output *iam.UpdateRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateRoleOutput = output
	}
}

func WithUpdateRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateRoleError = err
	}
}

func WithDeleteRoleOutput(output *iam.DeleteRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRoleOutput = output
	}
}

func WithDeleteRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRoleError = err
	}
}

// Mock configuration options for Role Policy Document operations

func WithUpdateAssumeRolePolicyOutput(output *iam.UpdateAssumeRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateAssumeRolePolicyOutput = output
	}
}

func WithUpdateAssumeRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateAssumeRolePolicyError = err
	}
}

// Mock configuration options for Role Policy Attachment operations

func WithAttachRolePolicyOutput(output *iam.AttachRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachRolePolicyOutput = output
	}
}

func WithAttachRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachRolePolicyError = err
	}
}

func WithDetachRolePolicyOutput(output *iam.DetachRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachRolePolicyOutput = output
	}
}

func WithDetachRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachRolePolicyError = err
	}
}

func WithListAttachedRolePoliciesOutput(output *iam.ListAttachedRolePoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedRolePoliciesOutput = output
	}
}

func WithListAttachedRolePoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedRolePoliciesError = err
	}
}

// Mock configuration options for Inline Policy operations

func WithPutRolePolicyOutput(output *iam.PutRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putRolePolicyOutput = output
	}
}

func WithPutRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putRolePolicyError = err
	}
}

func WithDeleteRolePolicyOutput(output *iam.DeleteRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRolePolicyOutput = output
	}
}

func WithDeleteRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRolePolicyError = err
	}
}

func WithListRolePoliciesOutput(output *iam.ListRolePoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listRolePoliciesOutput = output
	}
}

func WithListRolePoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listRolePoliciesError = err
	}
}

func WithGetRolePolicyOutput(output *iam.GetRolePolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getRolePolicyOutput = output
	}
}

func WithGetRolePolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getRolePolicyError = err
	}
}

// Mock configuration options for Tag operations

func WithTagRoleOutput(output *iam.TagRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagRoleOutput = output
	}
}

func WithTagRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagRoleError = err
	}
}

func WithUntagRoleOutput(output *iam.UntagRoleOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagRoleOutput = output
	}
}

func WithUntagRoleError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagRoleError = err
	}
}

func WithListRoleTagsOutput(output *iam.ListRoleTagsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listRoleTagsOutput = output
	}
}

func WithListRoleTagsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listRoleTagsError = err
	}
}

// Mock configuration options for Permissions Boundary operations

func WithPutRolePermissionsBoundaryOutput(output *iam.PutRolePermissionsBoundaryOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putRolePermissionsBoundaryOutput = output
	}
}

func WithPutRolePermissionsBoundaryError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putRolePermissionsBoundaryError = err
	}
}

func WithDeleteRolePermissionsBoundaryOutput(output *iam.DeleteRolePermissionsBoundaryOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRolePermissionsBoundaryOutput = output
	}
}

func WithDeleteRolePermissionsBoundaryError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteRolePermissionsBoundaryError = err
	}
}

// Mock configuration options for User operations

func WithCreateUserOutput(output *iam.CreateUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createUserOutput = output
	}
}

func WithCreateUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createUserError = err
	}
}

func WithGetUserOutput(output *iam.GetUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getUserOutput = output
	}
}

func WithGetUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getUserError = err
	}
}

func WithUpdateUserOutput(output *iam.UpdateUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateUserOutput = output
	}
}

func WithUpdateUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateUserError = err
	}
}

func WithDeleteUserOutput(output *iam.DeleteUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserOutput = output
	}
}

func WithDeleteUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserError = err
	}
}

// Mock configuration options for User Policy Attachment operations

func WithAttachUserPolicyOutput(output *iam.AttachUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachUserPolicyOutput = output
	}
}

func WithAttachUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.attachUserPolicyError = err
	}
}

func WithDetachUserPolicyOutput(output *iam.DetachUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachUserPolicyOutput = output
	}
}

func WithDetachUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.detachUserPolicyError = err
	}
}

func WithListAttachedUserPoliciesOutput(output *iam.ListAttachedUserPoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedUserPoliciesOutput = output
	}
}

func WithListAttachedUserPoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listAttachedUserPoliciesError = err
	}
}

// Mock configuration options for User Policy operations

func WithPutUserPolicyOutput(output *iam.PutUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putUserPolicyOutput = output
	}
}

func WithPutUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putUserPolicyError = err
	}
}

func WithDeleteUserPolicyOutput(output *iam.DeleteUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserPolicyOutput = output
	}
}

func WithDeleteUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserPolicyError = err
	}
}

func WithListUserPoliciesOutput(output *iam.ListUserPoliciesOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listUserPoliciesOutput = output
	}
}

func WithListUserPoliciesError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listUserPoliciesError = err
	}
}

// Mock configuration options for User Tag operations

func WithTagUserOutput(output *iam.TagUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagUserOutput = output
	}
}

func WithTagUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.tagUserError = err
	}
}

func WithUntagUserOutput(output *iam.UntagUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagUserOutput = output
	}
}

func WithUntagUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.untagUserError = err
	}
}

func WithListUserTagsOutput(output *iam.ListUserTagsOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listUserTagsOutput = output
	}
}

func WithListUserTagsError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listUserTagsError = err
	}
}

// Mock configuration options for User Permissions Boundary operations

func WithPutUserPermissionsBoundaryOutput(output *iam.PutUserPermissionsBoundaryOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putUserPermissionsBoundaryOutput = output
	}
}

func WithPutUserPermissionsBoundaryError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.putUserPermissionsBoundaryError = err
	}
}

func WithDeleteUserPermissionsBoundaryOutput(output *iam.DeleteUserPermissionsBoundaryOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserPermissionsBoundaryOutput = output
	}
}

func WithDeleteUserPermissionsBoundaryError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteUserPermissionsBoundaryError = err
	}
}

// Mock configuration options for User Policy operations

func WithGetUserPolicyOutput(output *iam.GetUserPolicyOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getUserPolicyOutput = output
	}
}

func WithGetUserPolicyError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getUserPolicyError = err
	}
}

// Mock configuration options for Group operations

func WithAddUserToGroupOutput(output *iam.AddUserToGroupOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.addUserToGroupOutput = output
	}
}

func WithAddUserToGroupError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.addUserToGroupError = err
	}
}

func WithRemoveUserFromGroupOutput(output *iam.RemoveUserFromGroupOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.removeUserFromGroupOutput = output
	}
}

func WithRemoveUserFromGroupError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.removeUserFromGroupError = err
	}
}

func WithListGroupsForUserOutput(output *iam.ListGroupsForUserOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listGroupsForUserOutput = output
	}
}

func WithListGroupsForUserError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.listGroupsForUserError = err
	}
}

// Mock configuration options for Login Profile operations

func WithCreateLoginProfileOutput(output *iam.CreateLoginProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createLoginProfileOutput = output
	}
}

func WithCreateLoginProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.createLoginProfileError = err
	}
}

func WithGetLoginProfileOutput(output *iam.GetLoginProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getLoginProfileOutput = output
	}
}

func WithGetLoginProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.getLoginProfileError = err
	}
}

func WithUpdateLoginProfileOutput(output *iam.UpdateLoginProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateLoginProfileOutput = output
	}
}

func WithUpdateLoginProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.updateLoginProfileError = err
	}
}

func WithDeleteLoginProfileOutput(output *iam.DeleteLoginProfileOutput) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteLoginProfileOutput = output
	}
}

func WithDeleteLoginProfileError(err error) iamServiceMockOption {
	return func(m *iamServiceMock) {
		m.deleteLoginProfileError = err
	}
}

// Service interface implementation methods

func (m *iamServiceMock) CreateRole(
	ctx context.Context,
	params *iam.CreateRoleInput,
	optFns ...func(*iam.Options),
) (*iam.CreateRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createRoleOutput, m.createRoleError
}

func (m *iamServiceMock) GetRole(
	ctx context.Context,
	params *iam.GetRoleInput,
	optFns ...func(*iam.Options),
) (*iam.GetRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getRoleOutput, m.getRoleError
}

func (m *iamServiceMock) UpdateRole(
	ctx context.Context,
	params *iam.UpdateRoleInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateRoleOutput, m.updateRoleError
}

func (m *iamServiceMock) DeleteRole(
	ctx context.Context,
	params *iam.DeleteRoleInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteRoleOutput, m.deleteRoleError
}

func (m *iamServiceMock) UpdateAssumeRolePolicy(
	ctx context.Context,
	params *iam.UpdateAssumeRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateAssumeRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateAssumeRolePolicyOutput, m.updateAssumeRolePolicyError
}

func (m *iamServiceMock) AttachRolePolicy(
	ctx context.Context,
	params *iam.AttachRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.AttachRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.attachRolePolicyOutput, m.attachRolePolicyError
}

func (m *iamServiceMock) DetachRolePolicy(
	ctx context.Context,
	params *iam.DetachRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DetachRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.detachRolePolicyOutput, m.detachRolePolicyError
}

func (m *iamServiceMock) ListAttachedRolePolicies(
	ctx context.Context,
	params *iam.ListAttachedRolePoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListAttachedRolePoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listAttachedRolePoliciesOutput, m.listAttachedRolePoliciesError
}

func (m *iamServiceMock) PutRolePolicy(
	ctx context.Context,
	params *iam.PutRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.PutRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putRolePolicyOutput, m.putRolePolicyError
}

func (m *iamServiceMock) DeleteRolePolicy(
	ctx context.Context,
	params *iam.DeleteRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteRolePolicyOutput, m.deleteRolePolicyError
}

func (m *iamServiceMock) ListRolePolicies(
	ctx context.Context,
	params *iam.ListRolePoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListRolePoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listRolePoliciesOutput, m.listRolePoliciesError
}

func (m *iamServiceMock) GetRolePolicy(
	ctx context.Context,
	params *iam.GetRolePolicyInput,
	optFns ...func(*iam.Options),
) (*iam.GetRolePolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getRolePolicyOutput, m.getRolePolicyError
}

func (m *iamServiceMock) TagRole(
	ctx context.Context,
	params *iam.TagRoleInput,
	optFns ...func(*iam.Options),
) (*iam.TagRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagRoleOutput, m.tagRoleError
}

func (m *iamServiceMock) UntagRole(
	ctx context.Context,
	params *iam.UntagRoleInput,
	optFns ...func(*iam.Options),
) (*iam.UntagRoleOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagRoleOutput, m.untagRoleError
}

func (m *iamServiceMock) ListRoleTags(
	ctx context.Context,
	params *iam.ListRoleTagsInput,
	optFns ...func(*iam.Options),
) (*iam.ListRoleTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listRoleTagsOutput, m.listRoleTagsError
}

func (m *iamServiceMock) PutRolePermissionsBoundary(
	ctx context.Context,
	params *iam.PutRolePermissionsBoundaryInput,
	optFns ...func(*iam.Options),
) (*iam.PutRolePermissionsBoundaryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putRolePermissionsBoundaryOutput, m.putRolePermissionsBoundaryError
}

func (m *iamServiceMock) DeleteRolePermissionsBoundary(
	ctx context.Context,
	params *iam.DeleteRolePermissionsBoundaryInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteRolePermissionsBoundaryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteRolePermissionsBoundaryOutput, m.deleteRolePermissionsBoundaryError
}

func (m *iamServiceMock) CreateUser(
	ctx context.Context,
	params *iam.CreateUserInput,
	optFns ...func(*iam.Options),
) (*iam.CreateUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createUserOutput, m.createUserError
}

func (m *iamServiceMock) GetUser(
	ctx context.Context,
	params *iam.GetUserInput,
	optFns ...func(*iam.Options),
) (*iam.GetUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getUserOutput, m.getUserError
}

func (m *iamServiceMock) UpdateUser(
	ctx context.Context,
	params *iam.UpdateUserInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateUserOutput, m.updateUserError
}

func (m *iamServiceMock) DeleteUser(
	ctx context.Context,
	params *iam.DeleteUserInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteUserOutput, m.deleteUserError
}

func (m *iamServiceMock) AttachUserPolicy(
	ctx context.Context,
	params *iam.AttachUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.AttachUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.attachUserPolicyOutput, m.attachUserPolicyError
}

func (m *iamServiceMock) DetachUserPolicy(
	ctx context.Context,
	params *iam.DetachUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DetachUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.detachUserPolicyOutput, m.detachUserPolicyError
}

func (m *iamServiceMock) ListAttachedUserPolicies(
	ctx context.Context,
	params *iam.ListAttachedUserPoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListAttachedUserPoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listAttachedUserPoliciesOutput, m.listAttachedUserPoliciesError
}

func (m *iamServiceMock) PutUserPolicy(
	ctx context.Context,
	params *iam.PutUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.PutUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putUserPolicyOutput, m.putUserPolicyError
}

func (m *iamServiceMock) DeleteUserPolicy(
	ctx context.Context,
	params *iam.DeleteUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteUserPolicyOutput, m.deleteUserPolicyError
}

func (m *iamServiceMock) ListUserPolicies(
	ctx context.Context,
	params *iam.ListUserPoliciesInput,
	optFns ...func(*iam.Options),
) (*iam.ListUserPoliciesOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listUserPoliciesOutput, m.listUserPoliciesError
}

func (m *iamServiceMock) TagUser(
	ctx context.Context,
	params *iam.TagUserInput,
	optFns ...func(*iam.Options),
) (*iam.TagUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.tagUserOutput, m.tagUserError
}

func (m *iamServiceMock) UntagUser(
	ctx context.Context,
	params *iam.UntagUserInput,
	optFns ...func(*iam.Options),
) (*iam.UntagUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.untagUserOutput, m.untagUserError
}

func (m *iamServiceMock) ListUserTags(
	ctx context.Context,
	params *iam.ListUserTagsInput,
	optFns ...func(*iam.Options),
) (*iam.ListUserTagsOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listUserTagsOutput, m.listUserTagsError
}

func (m *iamServiceMock) PutUserPermissionsBoundary(
	ctx context.Context,
	params *iam.PutUserPermissionsBoundaryInput,
	optFns ...func(*iam.Options),
) (*iam.PutUserPermissionsBoundaryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.putUserPermissionsBoundaryOutput, m.putUserPermissionsBoundaryError
}

func (m *iamServiceMock) DeleteUserPermissionsBoundary(
	ctx context.Context,
	params *iam.DeleteUserPermissionsBoundaryInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteUserPermissionsBoundaryOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteUserPermissionsBoundaryOutput, m.deleteUserPermissionsBoundaryError
}

func (m *iamServiceMock) GetUserPolicy(
	ctx context.Context,
	params *iam.GetUserPolicyInput,
	optFns ...func(*iam.Options),
) (*iam.GetUserPolicyOutput, error) {
	m.RegisterCall(ctx, params)

	// Handle different policy names
	if params.PolicyName != nil {
		switch aws.ToString(params.PolicyName) {
		case "S3Access":
			return &iam.GetUserPolicyOutput{
				UserName:       params.UserName,
				PolicyName:     aws.String("S3Access"),
				PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["s3:GetObject","s3:PutObject"],"Resource":["arn:aws:s3:::my-bucket/*"]}]}`),
			}, nil
		case "DynamoDBAccess":
			return &iam.GetUserPolicyOutput{
				UserName:       params.UserName,
				PolicyName:     aws.String("DynamoDBAccess"),
				PolicyDocument: aws.String(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Action":["dynamodb:PutItem","dynamodb:GetItem"],"Resource":["arn:aws:dynamodb:::table/my-table"]}]}`),
			}, nil
		}
	}

	// Fallback to the default mock response
	return m.getUserPolicyOutput, m.getUserPolicyError
}

func (m *iamServiceMock) AddUserToGroup(
	ctx context.Context,
	params *iam.AddUserToGroupInput,
	optFns ...func(*iam.Options),
) (*iam.AddUserToGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.addUserToGroupOutput, m.addUserToGroupError
}

func (m *iamServiceMock) RemoveUserFromGroup(
	ctx context.Context,
	params *iam.RemoveUserFromGroupInput,
	optFns ...func(*iam.Options),
) (*iam.RemoveUserFromGroupOutput, error) {
	m.RegisterCall(ctx, params)
	return m.removeUserFromGroupOutput, m.removeUserFromGroupError
}

func (m *iamServiceMock) ListGroupsForUser(
	ctx context.Context,
	params *iam.ListGroupsForUserInput,
	optFns ...func(*iam.Options),
) (*iam.ListGroupsForUserOutput, error) {
	m.RegisterCall(ctx, params)
	return m.listGroupsForUserOutput, m.listGroupsForUserError
}

func (m *iamServiceMock) CreateLoginProfile(
	ctx context.Context,
	params *iam.CreateLoginProfileInput,
	optFns ...func(*iam.Options),
) (*iam.CreateLoginProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.createLoginProfileOutput, m.createLoginProfileError
}

func (m *iamServiceMock) GetLoginProfile(
	ctx context.Context,
	params *iam.GetLoginProfileInput,
	optFns ...func(*iam.Options),
) (*iam.GetLoginProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.getLoginProfileOutput, m.getLoginProfileError
}

func (m *iamServiceMock) UpdateLoginProfile(
	ctx context.Context,
	params *iam.UpdateLoginProfileInput,
	optFns ...func(*iam.Options),
) (*iam.UpdateLoginProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.updateLoginProfileOutput, m.updateLoginProfileError
}

func (m *iamServiceMock) DeleteLoginProfile(
	ctx context.Context,
	params *iam.DeleteLoginProfileInput,
	optFns ...func(*iam.Options),
) (*iam.DeleteLoginProfileOutput, error) {
	m.RegisterCall(ctx, params)
	return m.deleteLoginProfileOutput, m.deleteLoginProfileError
}
