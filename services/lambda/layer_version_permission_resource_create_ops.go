package lambda

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	lambdaservice "github.com/newstack-cloud/bluelink-provider-aws/services/lambda/service"
)

type layerVersionPermissionCreate struct {
	input *lambda.AddLayerVersionPermissionInput
}

func (u *layerVersionPermissionCreate) Name() string {
	return "add layer version permission"
}

func (u *layerVersionPermissionCreate) Prepare(
	saveOpCtx pluginutils.SaveOperationContext,
	specData *core.MappingNode,
	changes *provider.Changes,
) (bool, pluginutils.SaveOperationContext, error) {
	input, hasValues, err := changesToAddLayerVersionPermissionInput(
		specData,
	)
	if err != nil {
		return false, saveOpCtx, err
	}
	u.input = input
	return hasValues, saveOpCtx, nil
}

func (u *layerVersionPermissionCreate) Execute(
	ctx context.Context,
	saveOpCtx pluginutils.SaveOperationContext,
	lambdaService lambdaservice.Service,
) (pluginutils.SaveOperationContext, error) {
	newSaveOpCtx := pluginutils.SaveOperationContext{
		Data: saveOpCtx.Data,
	}

	addLayerVersionPermissionOutput, err := lambdaService.AddLayerVersionPermission(ctx, u.input)
	if err != nil {
		return saveOpCtx, err
	}

	combinedId := fmt.Sprintf("%s#%s",
		aws.ToString(u.input.LayerName)+":"+strconv.FormatInt(aws.ToInt64(u.input.VersionNumber), 10),
		aws.ToString(u.input.StatementId))

	newSaveOpCtx.ProviderUpstreamID = combinedId
	newSaveOpCtx.Data["addLayerVersionPermissionOutput"] = addLayerVersionPermissionOutput

	return newSaveOpCtx, err
}

func changesToAddLayerVersionPermissionInput(
	specData *core.MappingNode,
) (*lambda.AddLayerVersionPermissionInput, bool, error) {
	input := &lambda.AddLayerVersionPermissionInput{}

	layerVersionArn := core.StringValue(specData.Fields["layerVersionArn"])
	if layerVersionArn == "" {
		return nil, false, fmt.Errorf("layerVersionArn is required")
	}

	layerName, versionNumber, err := parseLayerVersionPermissionArn(layerVersionArn)
	if err != nil {
		return nil, false, fmt.Errorf("failed to parse layer version ARN: %w", err)
	}

	input.LayerName = aws.String(layerName)
	input.VersionNumber = aws.Int64(versionNumber)

	statementId := core.StringValue(specData.Fields["statementId"])
	if statementId == "" {
		return nil, false, fmt.Errorf("statementId is required")
	}
	input.StatementId = aws.String(statementId)

	valueSetters := []*pluginutils.ValueSetter[*lambda.AddLayerVersionPermissionInput]{
		pluginutils.NewValueSetter(
			"$.action",
			func(value *core.MappingNode, input *lambda.AddLayerVersionPermissionInput) {
				input.Action = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.principal",
			func(value *core.MappingNode, input *lambda.AddLayerVersionPermissionInput) {
				input.Principal = aws.String(core.StringValue(value))
			},
		),
		pluginutils.NewValueSetter(
			"$.organizationId",
			func(value *core.MappingNode, input *lambda.AddLayerVersionPermissionInput) {
				input.OrganizationId = aws.String(core.StringValue(value))
			},
		),
	}

	hasValuesToSave := true
	for _, valueSetter := range valueSetters {
		valueSetter.Set(specData, input)
		hasValuesToSave = hasValuesToSave || valueSetter.DidSet()
	}

	return input, hasValuesToSave, nil
}

// parseLayerVersionPermissionArn parses a layer version ARN or name to extract the layer name and version number.
func parseLayerVersionPermissionArn(layerVersionArn string) (layerName string, versionNumber int64, err error) {
	// Check if it's an ARN format: arn:aws:lambda:region:account:layer:layer-name:version
	if strings.HasPrefix(layerVersionArn, "arn:") {
		parts := strings.Split(layerVersionArn, ":")
		if len(parts) != 8 {
			return "", 0, fmt.Errorf("invalid layer version ARN format")
		}
		layerName = parts[6]
		version, parseErr := strconv.ParseInt(parts[7], 10, 64)
		if parseErr != nil {
			return "", 0, fmt.Errorf("invalid version number in ARN: %w", parseErr)
		}
		return layerName, version, nil
	}

	// If it's not an ARN, assume it's in the format layer-name:version
	parts := strings.Split(layerVersionArn, ":")
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("layer version must be in format 'layer-name:version' or full ARN")
	}

	layerName = parts[0]
	version, parseErr := strconv.ParseInt(parts[1], 10, 64)
	if parseErr != nil {
		return "", 0, fmt.Errorf("invalid version number: %w", parseErr)
	}

	return layerName, version, nil
}
