package lambdaelasticache

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

// The cache endpoint is read from state when the function's environment is built,
// so neither resource is written by this link.
func (l *functionCacheLinkActions) UpdateLinkedResources(
	ctx context.Context,
	input *provider.LinkUpdateLinkedResourcesInput,
) (*provider.LinkUpdateLinkedResourcesOutput, error) {
	return &provider.LinkUpdateLinkedResourcesOutput{
		LinkData: core.MappingNodeFields(),
	}, nil
}

func (l *functionCacheLinkActions) UpdateIntermediaryResources(
	ctx context.Context,
	input *provider.LinkUpdateIntermediaryResourcesInput,
) (*provider.LinkUpdateIntermediaryResourcesOutput, error) {
	providerCtx := provider.NewProviderContextFromLinkContext(input.LinkContext, "aws")
	lambdaService, region, err := l.getLambdaServiceWithRegion(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	annotations := getCacheLinkAnnotations(input.ResourceAInfo, input.ResourceBInfo)

	// The execution role is only needed for the IAM auth grant; skip the role lookup for password
	// auth (which grants no IAM permission).
	setupCtx, err := linkutils.SetupLinkFromLambdaFunction(
		ctx,
		&linkutils.LambdaLinkSetupData{
			LambdaFuncResourceInfo: input.ResourceAInfo,
			LoadRoleInfo:           annotations.authMode == "iam",
		},
		lambdaService,
		input.ResourceService,
		providerCtx,
	)
	if err != nil {
		return nil, err
	}

	output := &provider.LinkUpdateIntermediaryResourcesOutput{LinkData: core.MappingNodeFields()}

	// IAM elasticache:Connect grant (only for token-based auth).
	if annotations.authMode == "iam" {
		output, err = l.reconcileConnectGrant(ctx, providerCtx, input, setupCtx, annotations, region)
		if err != nil {
			return nil, err
		}
	}

	// Open the security-group path from the function to the cache on the cache port.
	targetSecurityGroupIDs, hasTargetSG := extractCacheSecurityGroupIDs(input.ResourceBInfo)
	if !hasTargetSG {
		return nil, fmt.Errorf(
			"security group could not be retrieved from the linked to %q ElastiCache replication group resource",
			pluginutils.GetResourceName(input.ResourceBInfo),
		)
	}

	ec2Service, err := l.getEC2Service(ctx, providerCtx)
	if err != nil {
		return nil, err
	}

	return linkutils.ReconcileLinkNetworking(
		ctx,
		ec2Service,
		input,
		linkutils.NetworkingActivation{
			Caller:                 linkutils.CallerNetworkingFromLambdaVPCConfig(setupCtx.LambdaOutput.VpcConfig),
			Region:                 region,
			TargetSecurityGroupIDs: targetSecurityGroupIDs,
			TargetPort:             annotations.port,
		},
		output,
	)
}

func extractCacheEndpoint(cacheInfo *provider.ResourceInfo) (string, bool) {
	spec := pluginutils.GetCurrentStateSpecDataFromResourceInfo(cacheInfo)
	for _, path := range []string{"$.configurationEndPoint.address", "$.primaryEndPoint.address"} {
		if node, has := pluginutils.GetValueByPath(path, spec); has && node != nil && core.StringValue(node) != "" {
			return core.StringValue(node), true
		}
	}
	return "", false
}

// Reads the cache's security group ids. securityGroupIds is
// write-only on the ElastiCache replication group (not returned in external state), so it
// is read from the resolved blueprint spec, where it is set as a reference to a flex VPC
// security group.
//
// Every group is returned rather than the first. Which one the link should pair against is
// decided by matching them to the groups the flex VPC prepared, not by their order here.
func extractCacheSecurityGroupIDs(cacheInfo *provider.ResourceInfo) ([]string, bool) {
	specs := []*core.MappingNode{
		resolvedSpecData(cacheInfo),
		pluginutils.GetCurrentStateSpecDataFromResourceInfo(cacheInfo),
	}
	for _, spec := range specs {
		node, has := pluginutils.GetValueByPath("$.securityGroupIds", spec)
		if !has || node == nil || len(node.Items) == 0 {
			continue
		}

		groupIDs := make([]string, 0, len(node.Items))
		for _, item := range node.Items {
			if groupID := core.StringValue(item); groupID != "" {
				groupIDs = append(groupIDs, groupID)
			}
		}
		if len(groupIDs) > 0 {
			return groupIDs, true
		}
	}

	return nil, false
}

func resolvedSpecData(resourceInfo *provider.ResourceInfo) *core.MappingNode {
	if resourceInfo == nil || resourceInfo.ResourceWithResolvedSubs == nil {
		return &core.MappingNode{Fields: map[string]*core.MappingNode{}}
	}
	return resourceInfo.ResourceWithResolvedSubs.Spec
}

type cacheEnvVarNameSet struct {
	host string
	port string
}

func (s cacheEnvVarNameSet) all() []string {
	return []string{s.host, s.port}
}

func cacheEnvVarNames(prefix string) cacheEnvVarNameSet {
	return cacheEnvVarNameSet{
		host: prefix + "_HOST",
		port: prefix + "_PORT",
	}
}

func cacheEnvVarPrefix(userDefinedPrefix string, resourceInfo *provider.ResourceInfo) string {
	if userDefinedPrefix != "" {
		return userDefinedPrefix
	}
	return "CACHE_" + strings.ToUpper(pluginutils.StripNonAlphaNumericChars(resourceInfo.ResourceName))
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for key := range m {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

type cacheLinkAnnotations struct {
	populateEnvVars bool
	envVarPrefix    string
	authMode        string
	userId          string
	port            int32
}

func getCacheLinkAnnotations(
	resourceInfo *provider.ResourceInfo,
	otherResourceInfo *provider.ResourceInfo,
) *cacheLinkAnnotations {
	populateEnvVars, _ := pluginutils.GetBoolAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[bool]{
			Key:         fmt.Sprintf("aws.lambda.elasticache.%s.populateEnvVars", otherResourceInfo.ResourceName),
			FallbackKey: "aws.lambda.elasticache.populateEnvVars",
			Default:     true,
		},
	)
	envVarPrefix, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key: fmt.Sprintf("aws.lambda.elasticache.%s.envVarPrefix", otherResourceInfo.ResourceName),
		},
	)
	authMode, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.elasticache.%s.authMode", otherResourceInfo.ResourceName),
			Default: "password",
		},
	)
	userId, _ := pluginutils.GetStringAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[string]{
			Key:     fmt.Sprintf("aws.lambda.elasticache.%s.userId", otherResourceInfo.ResourceName),
			Default: "*",
		},
	)
	port, _ := pluginutils.GetIntAnnotation(
		resourceInfo,
		&pluginutils.AnnotationQuery[int]{
			Key:     fmt.Sprintf("aws.lambda.elasticache.%s.port", otherResourceInfo.ResourceName),
			Default: 6379,
		},
	)

	return &cacheLinkAnnotations{
		populateEnvVars: populateEnvVars,
		envVarPrefix:    envVarPrefix,
		authMode:        authMode,
		userId:          userId,
		port:            int32(port),
	}
}
