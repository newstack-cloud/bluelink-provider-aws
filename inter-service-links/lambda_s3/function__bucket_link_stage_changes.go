package lambdas3

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *functionBucketLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getBucketLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)
	// The name the variable is actually given, which falls back to a derived one when
	// the annotation is not set. Staging used the annotation as it was given, so an unset
	// one recorded a path with an empty name in it. That was cosmetic while the link
	// wrote the function itself; the declaration below is read to decide which resource
	// waits for this link, so it has to name the field the link will contribute.
	envVarName := bucketEnvVarName(
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
		targetBucketFieldPath := fmt.Sprintf("$.%s", envVarFieldPath)
		err := linkhelpers.CollectChanges(
			"$.spec.bucketName",
			targetBucketFieldPath,
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
		// deployment, as the granted resource ARNs derive from the bucket name.
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
				envVarName,
			): envVarFieldPath,
		}
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
