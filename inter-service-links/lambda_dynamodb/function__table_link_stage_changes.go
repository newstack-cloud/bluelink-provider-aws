package lambdadynamodb

import (
	"context"
	"fmt"

	"github.com/newstack-cloud/bluelink/libs/blueprint/linkhelpers"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (l *lambdaFunctionDynamoDBTableLinkActions) StageChanges(
	ctx context.Context,
	input *provider.LinkStageChangesInput,
) (*provider.LinkStageChangesOutput, error) {
	changes := &provider.LinkChanges{}

	annotations := getDynamoDBTableLinkAnnotations(
		&input.ResourceAChanges.AppliedResourceInfo,
		&input.ResourceBChanges.AppliedResourceInfo,
	)

	currentLinkData := linkhelpers.GetLinkDataFromState(input.CurrentLinkState)
	functionResourceName := pluginutils.GetResourceName(
		&input.ResourceAChanges.AppliedResourceInfo,
	)
	// The name the variable is actually given, which falls back to one derived from the
	// table when the annotation is not set. Staging used the annotation as it was given,
	// so an unset one recorded a path with an empty name in it. That was cosmetic while
	// the link wrote the function itself; the declaration below is read to decide which
	// resource waits for this link, so it has to name the field the link will contribute.
	envVarName := dynamoDBTableEnvVarName(
		annotations.envVarName,
		&input.ResourceBChanges.AppliedResourceInfo,
	)
	envVarFieldPath := fmt.Sprintf(
		"%s.environmentVariables[\"%s\"]",
		functionResourceName,
		envVarName,
	)
	_, linkDataHasEnvVar := pluginutils.GetValueByPath(
		fmt.Sprintf("$.%s", envVarFieldPath),
		currentLinkData,
	)

	if !annotations.populateEnvVars && linkDataHasEnvVar {
		changes.RemovedFields = append(changes.RemovedFields, envVarFieldPath)
	} else if annotations.populateEnvVars {
		targetTableFieldPath := fmt.Sprintf(
			"$.%s",
			envVarFieldPath,
		)
		err := linkhelpers.CollectChanges(
			"$.spec.tableName",
			targetTableFieldPath,
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
		// Indicate that when either of the resources will be (re)created,
		// the execution role permission data specific to this link will change
		// but the value will not be known until deployment.
		// As the permissions are to access the DynamoDB table,
		// the ARN has to change to trigger a change to the permissions.
		changes.FieldChangesKnownOnDeploy = append(
			changes.FieldChangesKnownOnDeploy,
			createLinkDataExecutionRoleName(&input.ResourceAChanges.AppliedResourceInfo),
		)
	}

	if annotations.populateEnvVars {
		// What this link will contribute to the function, declared before it runs so the
		// framework knows the function is waiting on it. The value itself is produced at
		// deploy time; this says only where it will land.
		changes.ResourceDataMappings = map[string]string{
			fmt.Sprintf(
				"%s::spec.environment.variables[\"%s\"]",
				functionResourceName,
				envVarName,
			): envVarFieldPath,
		}
	}

	return &provider.LinkStageChangesOutput{
		Changes: changes,
	}, nil
}
