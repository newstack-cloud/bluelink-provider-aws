package lambdalinks

import (
	"context"
	"fmt"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"

	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

func (l *lambdaFunctionCodeSigningConfigLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	functionResourceName := linkhelpers.GetResourceNameFromChanges(input.ResourceAChanges)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)

	codeSigningConfigArnWriteToPath := fmt.Sprintf(
		"$[%q].codeSigningConfigArn",
		functionResourceName,
	)
	err := linkhelpers.CollectChanges(
		"$.spec.codeSigningConfigArn",
		codeSigningConfigArnWriteToPath,
		currentLinkData,
		input.ResourceBChanges,
		changes,
	)
	if err != nil {
		return nil, err
	}

	// What this link will contribute, declared before it runs so the framework knows the
	// resource is waiting on it. The value names the contribution rather than a path built
	// from the resource's name: a name with a hyphen in it is not a valid path segment, and
	// this is the form the framework records once the link has produced it.
	changes.ResourceDataMappings = map[string]string{
		fmt.Sprintf(
			"%s::spec.codeSigningConfigArn",
			pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo),
		): "contributions[0]",
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
