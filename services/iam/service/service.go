package iamservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS IAM service
// used by the IAM resource implementations.
type Service interface {
	// CreateRole creates a new role for your AWS account.
	CreateRole(
		ctx context.Context,
		params *iam.CreateRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateRoleOutput, error)

	// GetRole retrieves information about the specified role, including the role's
	// path, GUID, ARN, and the role's trust policy that grants permission to assume the role.
	GetRole(
		ctx context.Context,
		params *iam.GetRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.GetRoleOutput, error)

	// UpdateRole updates the description or maximum session duration setting of a role.
	UpdateRole(
		ctx context.Context,
		params *iam.UpdateRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateRoleOutput, error)

	// UpdateAssumeRolePolicy updates the policy that grants an IAM entity permission
	// to assume a role.
	UpdateAssumeRolePolicy(
		ctx context.Context,
		params *iam.UpdateAssumeRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateAssumeRolePolicyOutput, error)

	// DeleteRole deletes the specified role. Unlike the AWS Management Console, when you
	// delete a role programmatically, you must delete the items attached to the role manually,
	// or the deletion fails.
	DeleteRole(
		ctx context.Context,
		params *iam.DeleteRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteRoleOutput, error)

	// AttachRolePolicy attaches the specified managed policy to the specified IAM role.
	AttachRolePolicy(
		ctx context.Context,
		params *iam.AttachRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.AttachRolePolicyOutput, error)

	// DetachRolePolicy removes the specified managed policy from the specified role.
	DetachRolePolicy(
		ctx context.Context,
		params *iam.DetachRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DetachRolePolicyOutput, error)

	// ListAttachedRolePolicies lists all managed policies that are attached to the specified IAM role.
	ListAttachedRolePolicies(
		ctx context.Context,
		params *iam.ListAttachedRolePoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListAttachedRolePoliciesOutput, error)

	// PutRolePolicy adds or updates an inline policy document that is embedded in the specified IAM role.
	PutRolePolicy(
		ctx context.Context,
		params *iam.PutRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.PutRolePolicyOutput, error)

	// DeleteRolePolicy deletes the specified inline policy that is embedded in the specified IAM role.
	DeleteRolePolicy(
		ctx context.Context,
		params *iam.DeleteRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteRolePolicyOutput, error)

	// ListRolePolicies lists the names of the inline policies that are embedded in the specified IAM role.
	ListRolePolicies(
		ctx context.Context,
		params *iam.ListRolePoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListRolePoliciesOutput, error)

	// TagRole adds one or more tags to an IAM role.
	TagRole(
		ctx context.Context,
		params *iam.TagRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.TagRoleOutput, error)

	// UntagRole removes the specified tags from the role.
	UntagRole(
		ctx context.Context,
		params *iam.UntagRoleInput,
		optFns ...func(*iam.Options),
	) (*iam.UntagRoleOutput, error)

	// ListRoleTags lists the tags that are attached to the specified role.
	ListRoleTags(
		ctx context.Context,
		params *iam.ListRoleTagsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListRoleTagsOutput, error)

	// PutRolePermissionsBoundary adds a permissions boundary to the IAM role.
	PutRolePermissionsBoundary(
		ctx context.Context,
		params *iam.PutRolePermissionsBoundaryInput,
		optFns ...func(*iam.Options),
	) (*iam.PutRolePermissionsBoundaryOutput, error)

	// DeleteRolePermissionsBoundary deletes the permissions boundary for the specified IAM role.
	DeleteRolePermissionsBoundary(
		ctx context.Context,
		params *iam.DeleteRolePermissionsBoundaryInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteRolePermissionsBoundaryOutput, error)

	// GetRolePolicy retrieves an inline policy document that is embedded in the specified IAM role.
	GetRolePolicy(
		ctx context.Context,
		params *iam.GetRolePolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetRolePolicyOutput, error)

	// CreateUser creates a new user for your AWS account.
	CreateUser(
		ctx context.Context,
		params *iam.CreateUserInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateUserOutput, error)

	// GetUser retrieves information about the specified user, including the user's
	// path, GUID, ARN, and the user's creation date.
	GetUser(
		ctx context.Context,
		params *iam.GetUserInput,
		optFns ...func(*iam.Options),
	) (*iam.GetUserOutput, error)

	// UpdateUser updates the name and/or the path of the specified IAM user.
	UpdateUser(
		ctx context.Context,
		params *iam.UpdateUserInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateUserOutput, error)

	// DeleteUser deletes the specified IAM user. Unlike the AWS Management Console, when you
	// delete a user programmatically, you must delete the items attached to the user manually,
	// or the deletion fails.
	DeleteUser(
		ctx context.Context,
		params *iam.DeleteUserInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteUserOutput, error)

	// AttachUserPolicy attaches the specified managed policy to the specified user.
	AttachUserPolicy(
		ctx context.Context,
		params *iam.AttachUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.AttachUserPolicyOutput, error)

	// DetachUserPolicy removes the specified managed policy from the specified user.
	DetachUserPolicy(
		ctx context.Context,
		params *iam.DetachUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DetachUserPolicyOutput, error)

	// ListAttachedUserPolicies lists all managed policies that are attached to the specified IAM user.
	ListAttachedUserPolicies(
		ctx context.Context,
		params *iam.ListAttachedUserPoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListAttachedUserPoliciesOutput, error)

	// PutUserPolicy adds or updates an inline policy document that is embedded in the specified IAM user.
	PutUserPolicy(
		ctx context.Context,
		params *iam.PutUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.PutUserPolicyOutput, error)

	// DeleteUserPolicy deletes the specified inline policy that is embedded in the specified IAM user.
	DeleteUserPolicy(
		ctx context.Context,
		params *iam.DeleteUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteUserPolicyOutput, error)

	// ListUserPolicies lists the names of the inline policies that are embedded in the specified IAM user.
	ListUserPolicies(
		ctx context.Context,
		params *iam.ListUserPoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListUserPoliciesOutput, error)

	// TagUser adds one or more tags to an IAM user.
	TagUser(
		ctx context.Context,
		params *iam.TagUserInput,
		optFns ...func(*iam.Options),
	) (*iam.TagUserOutput, error)

	// UntagUser removes the specified tags from the user.
	UntagUser(
		ctx context.Context,
		params *iam.UntagUserInput,
		optFns ...func(*iam.Options),
	) (*iam.UntagUserOutput, error)

	// ListUserTags lists the tags that are attached to the specified user.
	ListUserTags(
		ctx context.Context,
		params *iam.ListUserTagsInput,
		optFns ...func(*iam.Options),
	) (*iam.ListUserTagsOutput, error)

	// PutUserPermissionsBoundary adds a permissions boundary to the IAM user.
	PutUserPermissionsBoundary(
		ctx context.Context,
		params *iam.PutUserPermissionsBoundaryInput,
		optFns ...func(*iam.Options),
	) (*iam.PutUserPermissionsBoundaryOutput, error)

	// DeleteUserPermissionsBoundary deletes the permissions boundary for the specified IAM user.
	DeleteUserPermissionsBoundary(
		ctx context.Context,
		params *iam.DeleteUserPermissionsBoundaryInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteUserPermissionsBoundaryOutput, error)

	// GetUserPolicy retrieves an inline policy document that is embedded in the specified IAM user.
	GetUserPolicy(
		ctx context.Context,
		params *iam.GetUserPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetUserPolicyOutput, error)

	// AddUserToGroup adds the specified user to the specified group.
	AddUserToGroup(
		ctx context.Context,
		params *iam.AddUserToGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.AddUserToGroupOutput, error)

	// RemoveUserFromGroup removes the specified user from the specified group.
	RemoveUserFromGroup(
		ctx context.Context,
		params *iam.RemoveUserFromGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.RemoveUserFromGroupOutput, error)

	// ListGroupsForUser lists the IAM groups that the specified IAM user belongs to.
	ListGroupsForUser(
		ctx context.Context,
		params *iam.ListGroupsForUserInput,
		optFns ...func(*iam.Options),
	) (*iam.ListGroupsForUserOutput, error)

	// CreateLoginProfile creates a password for the specified IAM user.
	CreateLoginProfile(
		ctx context.Context,
		params *iam.CreateLoginProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateLoginProfileOutput, error)

	// GetLoginProfile retrieves the login profile for the specified IAM user.
	GetLoginProfile(
		ctx context.Context,
		params *iam.GetLoginProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.GetLoginProfileOutput, error)

	// UpdateLoginProfile changes the password for the specified IAM user.
	UpdateLoginProfile(
		ctx context.Context,
		params *iam.UpdateLoginProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateLoginProfileOutput, error)

	// DeleteLoginProfile deletes the password for the specified IAM user.
	DeleteLoginProfile(
		ctx context.Context,
		params *iam.DeleteLoginProfileInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteLoginProfileOutput, error)

	// CreateGroup creates a new group for your AWS account.
	CreateGroup(
		ctx context.Context,
		params *iam.CreateGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.CreateGroupOutput, error)

	// GetGroup retrieves information about the specified group, including the group's
	// path, GUID, ARN, and the group's creation date.
	GetGroup(
		ctx context.Context,
		params *iam.GetGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.GetGroupOutput, error)

	// UpdateGroup updates the name and/or the path of the specified IAM group.
	UpdateGroup(
		ctx context.Context,
		params *iam.UpdateGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.UpdateGroupOutput, error)

	// DeleteGroup deletes the specified IAM group. Unlike the AWS Management Console, when you
	// delete a group programmatically, you must delete the items attached to the group manually,
	// or the deletion fails.
	DeleteGroup(
		ctx context.Context,
		params *iam.DeleteGroupInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteGroupOutput, error)

	// AttachGroupPolicy attaches the specified managed policy to the specified group.
	AttachGroupPolicy(
		ctx context.Context,
		params *iam.AttachGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.AttachGroupPolicyOutput, error)

	// DetachGroupPolicy removes the specified managed policy from the specified group.
	DetachGroupPolicy(
		ctx context.Context,
		params *iam.DetachGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DetachGroupPolicyOutput, error)

	// ListAttachedGroupPolicies lists all managed policies that are attached to the specified IAM group.
	ListAttachedGroupPolicies(
		ctx context.Context,
		params *iam.ListAttachedGroupPoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListAttachedGroupPoliciesOutput, error)

	// PutGroupPolicy adds or updates an inline policy document that is embedded in the specified IAM group.
	PutGroupPolicy(
		ctx context.Context,
		params *iam.PutGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.PutGroupPolicyOutput, error)

	// DeleteGroupPolicy deletes the specified inline policy that is embedded in the specified IAM group.
	DeleteGroupPolicy(
		ctx context.Context,
		params *iam.DeleteGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.DeleteGroupPolicyOutput, error)

	// ListGroupPolicies lists the names of the inline policies that are embedded in the specified IAM group.
	ListGroupPolicies(
		ctx context.Context,
		params *iam.ListGroupPoliciesInput,
		optFns ...func(*iam.Options),
	) (*iam.ListGroupPoliciesOutput, error)

	// GetGroupPolicy retrieves an inline policy document that is embedded in the specified IAM group.
	GetGroupPolicy(
		ctx context.Context,
		params *iam.GetGroupPolicyInput,
		optFns ...func(*iam.Options),
	) (*iam.GetGroupPolicyOutput, error)
}

// NewService creates a new instance of the AWS IAM service
// based on the provided AWS configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return iam.NewFromConfig(
		*awsConfig,
		iam.WithEndpointResolverV2(
			&iamEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type iamEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *iamEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params iam.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	iamAliases := utils.Services["iam"]
	iamEndpoint, hasIamEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"iam",
		iamAliases,
	)
	if hasIamEndpoint && !core.IsScalarNil(iamEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(iamEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return iam.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
