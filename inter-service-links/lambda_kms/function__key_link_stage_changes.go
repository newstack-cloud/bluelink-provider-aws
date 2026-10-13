package lambdakms

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionKeyLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getKeyLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)
	// The name the variable is actually given, which falls back to a derived one when
	// the annotation is not set. Staging used the annotation as it was given, so an unset
	// one recorded a path with an empty name in it. That was cosmetic while the link
	// wrote the function itself; the declaration below is read to decide which resource
	// waits for this link, so it has to name the field the link will contribute.
	envVarName := keyEnvVarName(
		annotations.envVarName,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	envVarFieldPath := fmt.Sprintf(
		"%s.environmentVariables[\"%s\"]",
		pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo),
		envVarName,
	)
	_, linkDataHasEnvVar := pluginutils.GetValueByPath(
		fmt.Sprintf("$.%s", envVarFieldPath),
		currentLinkData,
	)

	if !annotations.populateEnvVars && linkDataHasEnvVar {
		changes.RemovedFields = append(changes.RemovedFields, envVarFieldPath)
	} else if annotations.populateEnvVars {
		targetKeyFieldPath := fmt.Sprintf("$.%s", envVarFieldPath)
		// The environment variable holds the key ARN.
		err := linkhelpers.CollectChanges(
			"$.spec.arn",
			targetKeyFieldPath,
			linkhelpers.GetLinkDataFromState(input.CurrentLinkState),
			input.ResourceBChanges,
			changes,
		)
		if err != nil {
			return nil, err
		}
	}

	if pluginutils.IsResourceNew(input.ResourceAChanges) ||
		pluginutils.IsResourceNew(input.ResourceBChanges) {
		// When either resource will be (re)created, the execution-role permission data
		// specific to this link will change but the value will not be known until
		// deployment, as the granted key ARN derives from the key resource.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createLinkDataExecutionRoleName(&input.ResourceAChanges.AppliedResourceInfo),
		)
	}

	stageKeyGrantChanges(input, annotations, currentLinkData, changes)

	// Networking activation for a VPC-attached caller is configured at deploy; surface it as a
	// best-effort known-on-deploy signal.
	linkutils.StageNetworkAccessKnownOnDeploy(input.ResourceAChanges, changes)

	if annotations.populateEnvVars {
		// What this link will contribute to the function, declared before it runs so the
		// framework knows the function is waiting on it. The value itself is produced at
		// deploy time; this says only where it will land.
		changes.ResourceDataMappings = map[string]string{
			fmt.Sprintf(
				"%s::spec.environment.variables[\"%s\"]",
				pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo),
				envVarName,
			): envVarFieldPath,
		}
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}

// This surfaces the managed KMS grant side effect in staged changes: creating
// or updating the grant when manageKeyGrant is enabled, or revoking it when disabled. The
// grant's name and operations are known at plan time, so the change carries a concrete value.
func stageKeyGrantChanges(
	input *provider.LinkStageChangesInput,
	annotations *keyLinkAnnotations,
	currentLinkData *core.MappingNode,
	changes *provider.LinkChanges,
) {
	currentGrant, hasCurrentGrant := pluginutils.GetValueByPath(
		fmt.Sprintf("$.%s", keyGrantLinkDataField),
		currentLinkData,
	)

	if !annotations.manageKeyGrant {
		if hasCurrentGrant {
			changes.RemovedFields = append(changes.RemovedFields, keyGrantLinkDataField)
		}
		return
	}

	desiredGrant := keyGrantLinkDataNode(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
		annotations.accessLevel,
	)

	switch {
	case !hasCurrentGrant:
		changes.NewFields = append(changes.NewFields, &provider.FieldChange{
			FieldPath: keyGrantLinkDataField,
			NewValue:  desiredGrant,
		})
	case !core.MappingNodeEqual(currentGrant, desiredGrant):
		changes.ModifiedFields = append(changes.ModifiedFields, &provider.FieldChange{
			FieldPath: keyGrantLinkDataField,
			PrevValue: currentGrant,
			NewValue:  desiredGrant,
		})
	default:
		changes.UnchangedFields = append(changes.UnchangedFields, keyGrantLinkDataField)
	}
}
