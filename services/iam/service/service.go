package iamservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/celerity-provider-aws/utils"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
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
