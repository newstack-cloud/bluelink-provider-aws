package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/aws/aws-sdk-go-v2/service/lambda/types"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

type tagsUpdate struct {
	saveTagsInput   *lambda.TagResourceInput
	removeTagsInput *lambda.UntagResourceInput
	pathRoot        string
}

func (u *tagsUpdate) Name() string {
	return "tags"
}

func (u *tagsUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	currentResourceStateSpecData := pluginutils.GetCurrentResourceStateSpecData(changes)
	newTagsNode, _ := pluginutils.GetValueByPath(u.pathRoot, specData)
	currentTagsNode, _ := pluginutils.GetValueByPath(
		u.pathRoot,
		currentResourceStateSpecData,
	)
	input, hasUpdates := changesToResourceTagUpdatesInput(
		saveOpCtx.ProviderUpstreamID,
		newTagsNode,
		currentTagsNode,
	)
	u.saveTagsInput = input.saveTagsInput
	u.removeTagsInput = input.removeTagsInput
	return hasUpdates, saveOpCtx, nil
}

func (u *tagsUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	if len(u.saveTagsInput.Tags) > 0 {
		_, err := lambdaService.TagResource(ctx, u.saveTagsInput)
		if err != nil {
			return saveOpCtx, err
		}
	}

	if len(u.removeTagsInput.TagKeys) > 0 {
		_, err := lambdaService.UntagResource(ctx, u.removeTagsInput)
		if err != nil {
			return saveOpCtx, err
		}
	}

	return saveOpCtx, nil
}

type tagUpdatesInput struct {
	saveTagsInput   *lambda.TagResourceInput
	removeTagsInput *lambda.UntagResourceInput
}

func changesToResourceTagUpdatesInput(
	arn string,
	newTagsNode *core.MappingNode,
	currentTagsNode *core.MappingNode,
) (*tagUpdatesInput, bool) {
	removedTags := []string{}
	addTags := map[string]string{}

	newTagsNodeItems := getItems(newTagsNode)
	currentTagsNodeItems := getItems(currentTagsNode)
	for _, item := range newTagsNodeItems {
		key := core.StringValue(item.Fields["key"])
		value := core.StringValue(item.Fields["value"])
		addTags[key] = value
	}

	for _, item := range currentTagsNodeItems {
		key := core.StringValue(item.Fields["key"])
		if _, inNewTags := addTags[key]; !inNewTags {
			removedTags = append(removedTags, key)
		}
	}

	hasUpdates := len(addTags) > 0 || len(removedTags) > 0

	return &tagUpdatesInput{
		saveTagsInput: &lambda.TagResourceInput{
			Resource: aws.String(arn),
			Tags:     addTags,
		},
		removeTagsInput: &lambda.UntagResourceInput{
			Resource: aws.String(arn),
			TagKeys:  removedTags,
		},
	}, hasUpdates
}

// Shared between function and function version resources.
type functionRuntimeManagementConfigUpdate struct {
	path                 string
	fieldChangesPathRoot string
	input                *lambda.PutRuntimeManagementConfigInput
}

func (u *functionRuntimeManagementConfigUpdate) Name() string {
	return "runtime management config"
}

func (u *functionRuntimeManagementConfigUpdate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	runtimeMgmtConfigData, _ := pluginutils.GetValueByPath(
		u.path,
		specData,
	)
	functionARN, version := extractFunctionARNAndVersion(saveOpCtx)
	input, hasUpdates := changesToPutRuntimeMgmtConfigInput(
		functionARN,
		version,
		runtimeMgmtConfigData,
		changes,
		u.fieldChangesPathRoot,
	)
	u.input = input
	return hasUpdates, saveOpCtx, nil
}

func (u *functionRuntimeManagementConfigUpdate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	_, err := lambdaService.PutRuntimeManagementConfig(ctx, u.input)
	return saveOpCtx, err
}

func getItems(node *core.MappingNode) []*core.MappingNode {
	if node == nil {
		return nil
	}

	return node.Items
}

// Image Config Setters.
func setImageConfigCommand(
	value *core.MappingNode,
	imageConfig *types.ImageConfig,
) {
	imageConfig.Command = core.StringSliceValue(value)
}

func setImageConfigEntrypoint(
	value *core.MappingNode,
	imageConfig *types.ImageConfig,
) {
	imageConfig.EntryPoint = core.StringSliceValue(value)
}

func setImageConfigWorkingDirectory(
	value *core.MappingNode,
	imageConfig *types.ImageConfig,
) {
	imageConfig.WorkingDirectory = aws.String(core.StringValue(value))
}

// File System Config Setters.
func setFileSystemConfigARN(
	value *core.MappingNode,
	fileSystemConfig *types.FileSystemConfig,
) {
	fileSystemConfig.Arn = aws.String(core.StringValue(value))
}

func setFileSystemConfigLocalMountPath(
	value *core.MappingNode,
	fileSystemConfig *types.FileSystemConfig,
) {
	fileSystemConfig.LocalMountPath = aws.String(core.StringValue(value))
}

// Logging Config Setters.
func setLoggingConfigApplicationLogLevel(
	value *core.MappingNode,
	loggingConfig *types.LoggingConfig,
) {
	loggingConfig.ApplicationLogLevel = types.ApplicationLogLevel(core.StringValue(value))
}

func setLoggingConfigLogFormat(
	value *core.MappingNode,
	loggingConfig *types.LoggingConfig,
) {
	loggingConfig.LogFormat = types.LogFormat(core.StringValue(value))
}

func setLoggingConfigLogGroup(
	value *core.MappingNode,
	loggingConfig *types.LoggingConfig,
) {
	loggingConfig.LogGroup = aws.String(core.StringValue(value))
}

func setLoggingConfigSystemLogLevel(
	value *core.MappingNode,
	loggingConfig *types.LoggingConfig,
) {
	loggingConfig.SystemLogLevel = types.SystemLogLevel(core.StringValue(value))
}

// VPC Config Setters.
func setVPCConfigSecurityGroupIds(
	value *core.MappingNode,
	vpcConfig *types.VpcConfig,
) {
	vpcConfig.SecurityGroupIds = core.StringSliceValue(value)
}

func setVPCConfigSubnetIds(
	value *core.MappingNode,
	vpcConfig *types.VpcConfig,
) {
	vpcConfig.SubnetIds = core.StringSliceValue(value)
}

func setVPCConfigIPv6AllowedForDualStack(
	value *core.MappingNode,
	vpcConfig *types.VpcConfig,
) {
	vpcConfig.Ipv6AllowedForDualStack = aws.Bool(core.BoolValue(value))
}

func extractComputedFieldsFromFunctionConfig(
	functionConfiguration *types.FunctionConfiguration,
) map[string]*core.MappingNode {
	fields := map[string]*core.MappingNode{}
	if functionConfiguration != nil {
		fields["spec.arn"] = core.MappingNodeFromString(
			aws.ToString(functionConfiguration.FunctionArn),
		)

		if functionConfiguration.SnapStart != nil {
			fields["spec.snapStartResponseApplyOn"] = core.MappingNodeFromString(
				string(functionConfiguration.SnapStart.ApplyOn),
			)
			fields["spec.snapStartResponseOptimizationStatus"] = core.MappingNodeFromString(
				string(functionConfiguration.SnapStart.OptimizationStatus),
			)
		}
	}
	return fields
}

func changesToPutRuntimeMgmtConfigInput(
	functionARN string,
	version string,
	putRuntimeMgmtConfigData *core.MappingNode,
	changes *provider.Changes,
	pathRoot string,
) (*lambda.PutRuntimeManagementConfigInput, bool) {
	modifiedFields := pluginutils.MergeFieldChanges(
		changes.ModifiedFields,
		changes.NewFields,
	)

	input := &lambda.PutRuntimeManagementConfigInput{
		FunctionName: &functionARN,
	}
	if version != "" {
		input.Qualifier = &version
	}

	valueSetters := []*pluginutils.ValueSetter[*lambda.PutRuntimeManagementConfigInput]{
		pluginutils.NewValueSetter(
			"$.runtimeVersionArn",
			setUpdateFunctionConfigRuntimeVersionARN,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.PutRuntimeManagementConfigInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.PutRuntimeManagementConfigInput](
				modifiedFields,
				pathRoot,
			),
		),
		pluginutils.NewValueSetter(
			"$.updateRuntimeOn",
			setUpdateFunctionConfigUpdateRuntimeOn,
			pluginutils.WithValueSetterCheckIfChanged[*lambda.PutRuntimeManagementConfigInput](true),
			pluginutils.WithValueSetterModifiedFields[*lambda.PutRuntimeManagementConfigInput](
				modifiedFields,
				pathRoot,
			),
		),
	}

	hasUpdates := false
	for _, valueSetter := range valueSetters {
		valueSetter.Set(putRuntimeMgmtConfigData, input)
		hasUpdates = hasUpdates || valueSetter.DidSet()
	}

	return input, hasUpdates
}

// setCorsValue is a reusable function that sets CORS configuration on a Lambda function URL input.
// It can be used with both CreateFunctionUrlConfigInput and UpdateFunctionUrlConfigInput.
func setCorsValue(value *core.MappingNode, corsSetter func(*types.Cors)) {
	cors := &types.Cors{}
	if allowCredentials, exists := pluginutils.GetValueByPath("$.allowCredentials", value); exists {
		cors.AllowCredentials = aws.Bool(core.BoolValue(allowCredentials))
	}
	if allowHeaders, exists := pluginutils.GetValueByPath("$.allowHeaders", value); exists {
		headers := make([]string, len(allowHeaders.Items))
		for i, header := range allowHeaders.Items {
			headers[i] = core.StringValue(header)
		}
		cors.AllowHeaders = headers
	}
	if allowMethods, exists := pluginutils.GetValueByPath("$.allowMethods", value); exists {
		methods := make([]string, len(allowMethods.Items))
		for i, method := range allowMethods.Items {
			methods[i] = core.StringValue(method)
		}
		cors.AllowMethods = methods
	}
	if allowOrigins, exists := pluginutils.GetValueByPath("$.allowOrigins", value); exists {
		origins := make([]string, len(allowOrigins.Items))
		for i, origin := range allowOrigins.Items {
			origins[i] = core.StringValue(origin)
		}
		cors.AllowOrigins = origins
	}
	if exposeHeaders, exists := pluginutils.GetValueByPath("$.exposeHeaders", value); exists {
		headers := make([]string, len(exposeHeaders.Items))
		for i, header := range exposeHeaders.Items {
			headers[i] = core.StringValue(header)
		}
		cors.ExposeHeaders = headers
	}
	if maxAge, exists := pluginutils.GetValueByPath("$.maxAge", value); exists {
		cors.MaxAge = aws.Int32(int32(core.IntValue(maxAge)))
	}
	corsSetter(cors)
}
