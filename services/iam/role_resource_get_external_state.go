package iam

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/smithy-go"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (i *iamRoleResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := i.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	arn := core.StringValue(
		input.CurrentResourceSpec.Fields["arn"],
	)
	if arn == "" {
		return nil, fmt.Errorf("ARN is required for get external state operation")
	}

	roleName, err := extractRoleNameFromARN(arn)
	if err != nil {
		return nil, fmt.Errorf("failed to extract role name from ARN %s: %w", arn, err)
	}

	roleOutput, err := iamService.GetRole(
		ctx,
		&iam.GetRoleInput{
			RoleName: &roleName,
		},
	)
	if err != nil {
		var apiError smithy.APIError
		if errors.As(err, &apiError) {
			// If the role doesn't exist, return empty state
			if apiError.ErrorCode() == "NoSuchEntity" {
				return &provider.ResourceGetExternalStateOutput{
					ResourceSpecState: &core.MappingNode{
						Fields: map[string]*core.MappingNode{},
					},
				}, nil
			}
		}
		return nil, err
	}

	role := roleOutput.Role

	// URL decode the policy document
	decodedPolicyDocument, err := url.QueryUnescape(aws.ToString(role.AssumeRolePolicyDocument))
	if err != nil {
		return nil, err
	}

	// Parse the JSON policy document into a structured format
	var policyDocument interface{}
	if err := json.Unmarshal([]byte(decodedPolicyDocument), &policyDocument); err != nil {
		return nil, err
	}

	// Convert back to MappingNode structure
	policyMappingNode, err := convertInterfaceToMappingNode(policyDocument)
	if err != nil {
		return nil, err
	}

	resourceSpecState := &core.MappingNode{
		Fields: map[string]*core.MappingNode{
			"roleName": core.MappingNodeFromString(
				aws.ToString(role.RoleName),
			),
			"assumeRolePolicyDocument": policyMappingNode,
			"arn": core.MappingNodeFromString(
				aws.ToString(role.Arn),
			),
			"roleId": core.MappingNodeFromString(
				aws.ToString(role.RoleId),
			),
		},
	}

	// Add optional fields if they exist
	if role.Description != nil {
		resourceSpecState.Fields["description"] = core.MappingNodeFromString(
			aws.ToString(role.Description),
		)
	}

	if role.MaxSessionDuration != nil {
		resourceSpecState.Fields["maxSessionDuration"] = core.MappingNodeFromInt(
			int(aws.ToInt32(role.MaxSessionDuration)),
		)
	}

	if role.Path != nil {
		resourceSpecState.Fields["path"] = core.MappingNodeFromString(
			aws.ToString(role.Path),
		)
	}

	// Fetch and add inline policies as structured objects
	listPoliciesOutput, err := iamService.ListRolePolicies(ctx, &iam.ListRolePoliciesInput{
		RoleName: role.RoleName,
	})
	if err != nil {
		return nil, err
	}
	if len(listPoliciesOutput.PolicyNames) > 0 {
		policies := make([]*core.MappingNode, 0, len(listPoliciesOutput.PolicyNames))
		for _, policyName := range listPoliciesOutput.PolicyNames {
			getPolicyOutput, err := iamService.GetRolePolicy(ctx, &iam.GetRolePolicyInput{
				RoleName:   role.RoleName,
				PolicyName: aws.String(policyName),
			})
			if err != nil {
				return nil, err
			}
			var policyDoc interface{}
			if err := json.Unmarshal([]byte(aws.ToString(getPolicyOutput.PolicyDocument)), &policyDoc); err != nil {
				return nil, err
			}
			policyDocNode, err := convertInterfaceToMappingNode(policyDoc)
			if err != nil {
				return nil, err
			}
			policies = append(policies, &core.MappingNode{
				Fields: map[string]*core.MappingNode{
					"policyName":     core.MappingNodeFromString(policyName),
					"policyDocument": policyDocNode,
				},
			})
		}
		resourceSpecState.Fields["policies"] = &core.MappingNode{Items: policies}
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: resourceSpecState,
	}, nil
}

// convertInterfaceToMappingNode converts a JSON interface{} to a MappingNode.
func convertInterfaceToMappingNode(data interface{}) (*core.MappingNode, error) {
	switch v := data.(type) {
	case map[string]interface{}:
		fields := make(map[string]*core.MappingNode)
		for key, value := range v {
			convertedValue, err := convertInterfaceToMappingNode(value)
			if err != nil {
				return nil, err
			}
			fields[key] = convertedValue
		}
		return &core.MappingNode{Fields: fields}, nil
	case []interface{}:
		items := make([]*core.MappingNode, len(v))
		for i, item := range v {
			convertedItem, err := convertInterfaceToMappingNode(item)
			if err != nil {
				return nil, err
			}
			items[i] = convertedItem
		}
		return &core.MappingNode{Items: items}, nil
	case string:
		return core.MappingNodeFromString(v), nil
	case float64:
		return core.MappingNodeFromInt(int(v)), nil
	case bool:
		return core.MappingNodeFromBool(v), nil
	case nil:
		return &core.MappingNode{}, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T", data)
	}
}
