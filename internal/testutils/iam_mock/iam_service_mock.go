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
