package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/celerity/libs/blueprint/core"
	"github.com/newstack-cloud/celerity/libs/blueprint/provider"
)

func (i *iamRoleResourceActions) Stabilised(
	ctx context.Context,
	input *provider.ResourceHasStabilisedInput,
) (*provider.ResourceHasStabilisedOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	roleArn := core.StringValue(
		input.ResourceSpec.Fields["arn"],
	)
	if roleArn == "" {
		return nil, fmt.Errorf("ARN is required for stabilised check")
	}

	// Extract role name from ARN
	roleName, err := extractRoleNameFromARN(roleArn)
	if err != nil {
		return nil, fmt.Errorf("failed to extract role name from ARN %s: %w", roleArn, err)
	}

	_, err = iamService.GetRole(
		ctx,
		&iam.GetRoleInput{
			RoleName: &roleName,
		},
	)
	if err != nil {
		return nil, err
	}

	// IAM roles are typically available immediately after creation
	// Unlike Lambda functions which have states, IAM roles are stable once they exist
	hasStabilised := true
	return &provider.ResourceHasStabilisedOutput{
		Stabilised: hasStabilised,
	}, nil
}
