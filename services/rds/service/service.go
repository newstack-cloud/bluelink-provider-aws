package rdsservice

import (
	"context"
	"net/url"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	smithyendpoints "github.com/aws/smithy-go/endpoints"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
)

// Service is an interface that represents the functionality of the AWS RDS service used
// by the RDS link implementations.
//
// Only the proxy target operations are here. RDS instances, clusters and proxies
// themselves are Cloud Control–backed; this interface covers registering an instance
// behind a proxy, which a link performs directly rather than through a
// Cloud Control resource. The AWS::RDS::DBProxyTargetGroup type cannot serve that
// purpose as its creation blocks until the registered target reports healthy, and the
// security group rule that makes it healthy is written by the same link, so as a
// resource it would always be waiting on work that had not run yet.
type Service interface {
	// RegisterDBProxyTargets adds DB instances or clusters to a proxy's target group.
	RegisterDBProxyTargets(ctx context.Context, params *rds.RegisterDBProxyTargetsInput, optFns ...func(*rds.Options)) (*rds.RegisterDBProxyTargetsOutput, error)
	// DeregisterDBProxyTargets removes DB instances or clusters from a proxy's target group.
	DeregisterDBProxyTargets(ctx context.Context, params *rds.DeregisterDBProxyTargetsInput, optFns ...func(*rds.Options)) (*rds.DeregisterDBProxyTargetsOutput, error)
	// DescribeDBProxyTargets lists the targets currently registered with a proxy, so a
	// registration can be skipped when the target is already present.
	DescribeDBProxyTargets(ctx context.Context, params *rds.DescribeDBProxyTargetsInput, optFns ...func(*rds.Options)) (*rds.DescribeDBProxyTargetsOutput, error)
	// ModifyDBProxyTargetGroup applies connection pool settings to a proxy's target group.
	ModifyDBProxyTargetGroup(ctx context.Context, params *rds.ModifyDBProxyTargetGroupInput, optFns ...func(*rds.Options)) (*rds.ModifyDBProxyTargetGroupOutput, error)
}

// NewService creates a new instance of the AWS RDS service based on the provided AWS
// configuration.
func NewService(awsConfig *aws.Config, providerContext provider.Context) Service {
	return rds.NewFromConfig(
		*awsConfig,
		rds.WithEndpointResolverV2(
			&rdsEndpointResolverV2{
				providerContext,
			},
		),
	)
}

type rdsEndpointResolverV2 struct {
	providerContext provider.Context
}

func (i *rdsEndpointResolverV2) ResolveEndpoint(
	ctx context.Context,
	params rds.EndpointParameters,
) (smithyendpoints.Endpoint, error) {
	rdsAliases := utils.Services["rds"]
	rdsEndpoint, hasRDSEndpoint := utils.GetEndpointFromProviderConfig(
		i.providerContext,
		"rds",
		rdsAliases,
	)
	if hasRDSEndpoint && !core.IsScalarNil(rdsEndpoint) {
		u, err := url.Parse(core.StringValueFromScalar(rdsEndpoint))
		if err != nil {
			return smithyendpoints.Endpoint{}, err
		}
		return smithyendpoints.Endpoint{
			URI: *u,
		}, nil
	}

	return rds.NewDefaultEndpointResolverV2().ResolveEndpoint(ctx, params)
}
