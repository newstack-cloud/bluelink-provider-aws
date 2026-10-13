package eventslambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// The PutEvents grant on the function's execution role is applied in
// UpdateIntermediaryResources; neither the function nor the event bus is written by
// this link.
func (l *functionEventBusLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *functionEventBusLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, err := l.getLambdaService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceAInfo,
			LoadRoleInfo:           true,
		},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	// The role is locked around each grant below rather than for this whole
	// phase, which goes on to do slow work the role has no part in.
	iamService, err := l.getIamService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	sid := createPutEventsSID(input.ResourceBInfo)

	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		_, err := linkutils.GrantRoleAccessUnderLock(
			ctx,
			linkutils.RoleAccessLockFor(
				input.ResourceService,
				input.LinkID,
				input.ResourceAInfo,
				setupCtx,
				providerCtx,
			),
			iamService,
			linkutils.RoleAccessGrant{
				RoleName: setupCtx.RoleName,
				SID:      sid,
			})
		if err != nil {
			return nil, err
		}
		return &provider.LinkUpdateIntermediaryResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	eventBusARN, hasEventBusARN := utils.ExtractARNFromResourceInfo(input.ResourceBInfo)
	if !hasEventBusARN {
		return nil, fmt.Errorf(
			"event bus ARN could not be retrieved from the linked to %q event bus resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	result, err := linkutils.GrantRoleAccessUnderLock(
		ctx,
		linkutils.RoleAccessLockFor(
			input.ResourceService,
			input.LinkID,
			input.ResourceAInfo,
			setupCtx,
			providerCtx,
		),
		iamService,
		linkutils.RoleAccessGrant{
			RoleName:  setupCtx.RoleName,
			SID:       sid,
			Statement: putEventsStatement(sid, eventBusARN),
		})
	if err != nil {
		if errors.Is(err, linkutils.ErrAccessPolicyBudgetExhausted) {
			return nil, fmt.Errorf(
				"cannot grant Lambda %q permission to publish events to event bus %q: %w",
				pluginutils.GetResourceName(input.ResourceAInfo),
				pluginutils.GetResourceName(input.ResourceBInfo),
				err,
			)
		}
		return nil, err
	}

	return putEventsLinkOutput(input, setupCtx.RoleResourceName, sid, eventBusARN, result), nil
}

func putEventsStatement(sid, eventBusARN string) map[string]any {
	return map[string]any{
		"Sid":      sid,
		"Effect":   "Allow",
		"Action":   []any{"events:PutEvents"},
		"Resource": eventBusARN,
	}
}

// Records the granted statement in link data and, for inline
// placements, maps it onto the role's spec so the framework attributes the
// statement to this link and does not treat it as drift / strip it on redeploy.
func putEventsLinkOutput(
	input *provider.LinkUpdateIntermediaryResourcesInput,
	roleResourceName, sid, eventBusARN string,
	result linkutils.RoleAccessResult,
) *provider.LinkUpdateIntermediaryResourcesOutput {
	linkDataKey := createLinkDataExecutionRoleName(input.ResourceAInfo)
	roleLinkData := core.MappingNodeFields(
		linkutils.PermissionFieldName,
		specPutEventsStatementNode(sid, eventBusARN),
	)

	// Attribute the grant to this link so the role's drift/deploy does not strip it:
	// inline placements map the statement by Sid. Managed (overflow) placements map
	// the attached managed policy ARN.
	mappings := map[string]string{}
	linkutils.AppendRoleAccessMapping(mappings, roleLinkData, roleResourceName, linkDataKey, sid, result)

	return &provider.LinkUpdateIntermediaryResourcesOutput{
		LinkData:             core.MappingNodeFields(linkDataKey, roleLinkData),
		ResourceDataMappings: mappings,
	}
}

func specPutEventsStatementNode(sid, eventBusARN string) *core.MappingNode {
	return core.MappingNodeFields(
		"sid", core.MappingNodeFromString(sid),
		"effect", core.MappingNodeFromString("Allow"),
		"action", &core.MappingNode{Items: []*core.MappingNode{
			core.MappingNodeFromString("events:PutEvents"),
		}},
		"resource", core.MappingNodeFromString(eventBusARN),
	)
}

func eventBusEnvVarName(
	userDefinedEnvVarName string,
	resourceInfo *provider.ResourceInfo,
) string {
	if userDefinedEnvVarName != "" {
		return userDefinedEnvVarName
	}
	return fmt.Sprintf("EVENT_BUS_%s", resourceInfo.ResourceName)
}

func createPutEventsSID(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf(
		"EventBridgePutEvents%s",
		pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName),
	)
}

type eventBusLinkAnnotations struct {
	populateEnvVars bool
	envVarName      string
}

func getEventBusLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *eventBusLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key: fmt.Sprintf(
				"aws.lambda.events.%s.populateEnvVars",
				otherResourceInfo.ResourceName,
			),
			FallbackKey: "aws.lambda.events.populateEnvVars",
			Default:     true,
		},
	)

	envVarName, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf(
				"aws.lambda.events.%s.envVarName",
				otherResourceInfo.ResourceName,
			),
		},
	)

	return &eventBusLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarName:      envVarName,
	}
}

func createLinkDataExecutionRoleName(resourceInfo *provider.ResourceInfo) string {
	return fmt.Sprintf("%sExecutionRole", resourceInfo.ResourceName)
}

func extractEventBusNameFromResourceInfo(resourceInfo *provider.ResourceInfo) (string, bool) {
	eventBusName, hasEventBusName := pluginutils.GetValueByPath(
		"$.name",
		resourceInfo.CurrentResourceState.SpecData,
	)
	if !hasEventBusName {
		return "", false
	}
	return core.StringValue(eventBusName), true
}
