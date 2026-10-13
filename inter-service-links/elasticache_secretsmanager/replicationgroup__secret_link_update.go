package elasticachesecretsmanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	elasticachetypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/blueprint/state"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// UpdateLinkedResources reads the Redis AUTH token from the linked Secrets Manager secret at
// deploy time and applies it to the ElastiCache replication group via ModifyReplicationGroup.
// The token value is never written into blueprint state or a computed field.
//
// This link writes the replication group rather than contributing to it, which is
// deliberate and the only one of its kind left in this provider. A contribution's value is
// held in the link's data and persisted, so contributing the AUTH token would put a secret
// into blueprint state, which is exactly what this link exists to avoid. It is not exposed
// to the revocation that contributions solve either as authToken is a write-only field, and
// the resource strips write-only fields from both sides before diffing, so a later
// deployment of the replication group cannot clear the token by omitting it.
func (l *replicationGroupSecretLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	// On destroy, we do not remove the AUTH token: deleting the token (DELETE strategy) is
	// destructive to live clients, and the replication group is typically being torn down anyway.
	if input.LinkUpdateType == provider.LinkUpdateTypeDestroy {
		return &provider.LinkUpdateLinkedResourcesOutput{
			LinkData: core.MappingNodeFields(),
		}, nil
	}

	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")

	replicationGroupID, hasReplicationGroupID := extractReplicationGroupID(input.ResourceAInfo)
	if !hasReplicationGroupID {
		return nil, fmt.Errorf(
			"replication group id could not be retrieved from the ElastiCache replication group %q",
			pluginutils.GetResourceName(input.ResourceAInfo),
		)
	}

	// The Secrets Manager secret's primary identifier (its "id" field) is the secret ARN.
	secretARN, hasSecretARN := extractSecretID(input.ResourceBInfo)
	if !hasSecretARN {
		return nil, fmt.Errorf(
			"secret ARN could not be retrieved from the linked Secrets Manager secret %q",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	authToken, err := l.readAuthToken(ctx, providerCtx, secretARN)
	if err != nil {
		return nil, err
	}

	strategy := resolveAuthTokenUpdateStrategy(input)

	elastiCacheService, err := l.getElastiCacheService(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	_, err = elastiCacheService.ModifyReplicationGroup(ctx, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:      aws.String(replicationGroupID),
		AuthToken:               aws.String(authToken),
		AuthTokenUpdateStrategy: strategy,
		ApplyImmediately:        aws.Bool(true),
	})
	if err != nil {
		return nil, fmt.Errorf(
			"failed to apply Redis AUTH token to ElastiCache replication group %q: %w",
			replicationGroupID,
			err,
		)
	}

	// Record only that the token was applied and from which secret (an ARN, not a secret
	// value) plus the strategy used. The AUTH token value itself is never persisted.
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData: core.MappingNodeFields(
			"authToken",
			core.MappingNodeFields(
				"configured", core.MappingNodeFromBool(true),
				"secretArn", core.MappingNodeFromString(secretARN),
				"updateStrategy", core.MappingNodeFromString(string(strategy)),
			),
		),
	}, nil
}

// Reads the secret's string value to use as the Redis AUTH token. The value is held only in
// memory and passed straight to ModifyReplicationGroup; it is never returned in link data.
func (l *replicationGroupSecretLinkActions) readAuthToken(
	ctx context.Context,
	providerCtx provider.Context,
	secretARN string,
) (string, error) {
	secretsManagerService, err := l.getSecretsManagerService(ctx, providerCtx)
	if err != nil {
		return "", err
	}

	secretValueOutput, err := secretsManagerService.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(secretARN),
	})
	if err != nil {
		return "", fmt.Errorf(
			"failed to read the Redis AUTH token from Secrets Manager secret %q: %w",
			secretARN,
			err,
		)
	}

	authToken := aws.ToString(secretValueOutput.SecretString)
	if authToken == "" {
		return "", fmt.Errorf(
			"secrets Manager secret %q has no string value to use as a Redis AUTH token",
			secretARN,
		)
	}

	if err := validateAuthToken(authToken, secretARN); err != nil {
		return "", err
	}

	return authToken, nil
}

const (
	minAuthTokenLength = 16
	maxAuthTokenLength = 128
)

// Enforces the ElastiCache AUTH token constraints before the value is sent to
// ModifyReplicationGroup, so an invalid secret value fails with a clear error instead of an
// opaque API rejection. Error messages never include the secret value itself.
func validateAuthToken(authToken string, secretARN string) error {
	if len(authToken) < minAuthTokenLength || len(authToken) > maxAuthTokenLength {
		return fmt.Errorf(
			"the value of Secrets Manager secret %q is not a valid Redis AUTH token: "+
				"it must be between %d and %d characters in length",
			secretARN,
			minAuthTokenLength,
			maxAuthTokenLength,
		)
	}

	for _, char := range authToken {
		if char < ' ' || char > '~' {
			return fmt.Errorf(
				"the value of Secrets Manager secret %q is not a valid Redis AUTH token: "+
					"it must contain only printable ASCII characters",
				secretARN,
			)
		}
		if !isAllowedAuthTokenChar(char) {
			return fmt.Errorf(
				"the value of Secrets Manager secret %q is not a valid Redis AUTH token: "+
					"it may contain only letters, digits and the characters "+
					"'!', '&', '#', '$', '^', '<', '>' and '-'",
				secretARN,
			)
		}
	}

	return nil
}

// ElastiCache accepts alphanumerics and eight printable specials in a Redis AUTH token,
// and rejects everything else with an opaque "Invalid AUTH token provided". Checking the
// allowed set rather than a list of known-bad delimiters is what makes that rejection
// reachable here, where the message can say which secret and what the rule is.
func isAllowedAuthTokenChar(char rune) bool {
	switch {
	case char >= 'a' && char <= 'z',
		char >= 'A' && char <= 'Z',
		char >= '0' && char <= '9':
		return true
	}

	return strings.ContainsRune("!&#$^<>-", char)
}

// UpdateIntermediaryResources is a no-op: this link does not manage intermediary resources, the AUTH
// token is applied directly to the replication group by UpdateLinkedResources.
func (l *replicationGroupSecretLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	return &provider.LinkUpdateIntermediaryResourcesOutput{
		IntermediaryResourceStates: []*state.LinkIntermediaryResourceState{},
		LinkData:                   core.MappingNodeFields(),
	}, nil
}

// Determines the AuthTokenUpdateStrategy for ModifyReplicationGroup. ElastiCache only accepts
// SET after a previous ROTATE, so an explicit SET annotation is interpreted as "retire the
// previous tokens when updating": on first configuration there is nothing to retire and the
// link falls back to ROTATE, which is also the default for both create and update. This keeps
// the annotation safe as a static value regardless of whether a deployment creates or updates
// the link.
func resolveAuthTokenUpdateStrategy(
	input *provider.LinkUpdateLinkedResourcesInput,
) elasticachetypes.AuthTokenUpdateStrategyType {
	strategy, hasStrategy := pluginutils.GetStringAnnotation(
		input.ResourceAInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: "aws.elasticache.secretsmanager.authTokenUpdateStrategy",
		},
	)
	if !hasStrategy || strategy == "" {
		return elasticachetypes.AuthTokenUpdateStrategyTypeRotate
	}

	selected := elasticachetypes.AuthTokenUpdateStrategyType(strategy)
	if selected == elasticachetypes.AuthTokenUpdateStrategyTypeSet &&
		input.LinkUpdateType == provider.LinkUpdateTypeCreate {
		return elasticachetypes.AuthTokenUpdateStrategyTypeRotate
	}
	return selected
}

func extractReplicationGroupID(resourceInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(resourceInfo)
	idNode, has := pluginutils.GetValueByPath("$.replicationGroupId", spec)
	if !has || idNode == nil {
		return "", false
	}
	value := core.StringValue(idNode)
	return value, value != ""
}

func extractSecretID(secretInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(secretInfo)
	arnNode, has := pluginutils.GetValueByPath("$.id", spec)
	if !has || arnNode == nil {
		return "", false
	}
	value := core.StringValue(arnNode)
	return value, value != ""
}
