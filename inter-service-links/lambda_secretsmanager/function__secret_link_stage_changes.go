package lambdasecretsmanager

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionSecretLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getSecretLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)
	// Apply the same default the deploy path uses so the staged field path matches the
	// env var that is actually written when no custom name annotation is set.
	finalEnvVarName := secretEnvVarName(
		annotations.envVarName,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	envVarFieldPath := fmt.Sprintf(
		"%s.environmentVariables[\"%s\"]",
		pluginutils.GetResourceName(&input.ResourceAChanges.AppliedResourceInfo),
		finalEnvVarName,
	)
	_, linkDataHasEnvVar := pluginutils.GetValueByPath(
		fmt.Sprintf("$.%s", envVarFieldPath),
		currentLinkData,
	)

	if !annotations.populateEnvVars && linkDataHasEnvVar {
		changes.RemovedFields = append(changes.RemovedFields, envVarFieldPath)
	} else if annotations.populateEnvVars {
		targetSecretFieldPath := fmt.Sprintf("$.%s", envVarFieldPath)
		// The secret ARN is the secret's primary identifier ("id").
		err := linkhelpers.CollectChanges(
			"$.spec.id",
			targetSecretFieldPath,
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
		// deployment, as the granted secret ARN derives from the secret resource.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createLinkDataExecutionRoleName(&input.ResourceAChanges.AppliedResourceInfo),
		)
	}

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
				finalEnvVarName,
			): envVarFieldPath,
		}
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
