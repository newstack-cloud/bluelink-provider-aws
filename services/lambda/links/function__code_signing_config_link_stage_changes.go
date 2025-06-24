package lambdalinks

import (
	"context"
	"fmt"

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

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
