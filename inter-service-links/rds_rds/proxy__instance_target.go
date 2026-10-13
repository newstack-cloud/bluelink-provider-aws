package rdsrds

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdstypes "github.com/aws/aws-sdk-go-v2/service/rds/types"
	rdsservice "github.com/newstack-cloud/bluelink-provider-aws/services/rds/service"
)

// defaultTargetGroupName is the only target group a proxy has. RDS creates it with
// the proxy and does not support additional ones.
const defaultTargetGroupName = "default"

// This puts the DB instance behind the proxy, which is what makes the proxy
// usable.
//
// This is done here rather than through an AWS::RDS::DBProxyTargetGroup resource
// because that type's create blocks until the registered target reports healthy, and
// the target cannot become healthy until the security group rule written moments ago
// in this same link exists. As a resource it ran in an earlier phase than the rule it
// depended on and spent its whole timeout waiting for something that had not happened
// yet.
//
// This is idempotent as a target already registered is left alone, so a re-deploy or a retry of
// this link does not fail on a target it put there itself.
func registerTarget(
	ctx context.Context,
	rdsService rdsservice.Service,
	proxyName string,
	instanceIdentifier string,
) error {
	registered, err := targetIsRegistered(ctx, rdsService, proxyName, instanceIdentifier)
	if err != nil {
		return err
	}

	if registered {
		return nil
	}

	_, err = rdsService.RegisterDBProxyTargets(
		ctx,
		&rds.RegisterDBProxyTargetsInput{
			DBProxyName:           aws.String(proxyName),
			TargetGroupName:       aws.String(defaultTargetGroupName),
			DBInstanceIdentifiers: []string{instanceIdentifier},
		},
	)
	if err != nil {
		// Two links racing to register the same target both succeed as far as the
		// caller is concerned; the second one simply finds the work already done.
		var alreadyRegistered *rdstypes.DBProxyTargetAlreadyRegisteredFault
		if errors.As(err, &alreadyRegistered) {
			return nil
		}
		return fmt.Errorf(
			"registering DB instance %q with RDS proxy %q: %w",
			instanceIdentifier, proxyName, err,
		)
	}

	return nil
}

// This takes the instance back out of the proxy's target group, so a
// destroyed link does not leave registrations behind.
//
// A target that is already gone is not an error as the instance may have been destroyed
// before the link, which is the ordinary case when a whole blueprint comes down.
func deregisterTarget(
	ctx context.Context,
	rdsService rdsservice.Service,
	proxyName string,
	instanceIdentifier string,
) error {
	_, err := rdsService.DeregisterDBProxyTargets(
		ctx,
		&rds.DeregisterDBProxyTargetsInput{
			DBProxyName:           aws.String(proxyName),
			TargetGroupName:       aws.String(defaultTargetGroupName),
			DBInstanceIdentifiers: []string{instanceIdentifier},
		},
	)
	if err != nil {
		if isTargetOrProxyGone(err) {
			return nil
		}

		return fmt.Errorf(
			"deregistering DB instance %q from RDS proxy %q: %w",
			instanceIdentifier, proxyName, err,
		)
	}

	return nil
}

func targetIsRegistered(
	ctx context.Context,
	rdsService rdsservice.Service,
	proxyName string,
	instanceIdentifier string,
) (bool, error) {
	output, err := rdsService.DescribeDBProxyTargets(
		ctx,
		&rds.DescribeDBProxyTargetsInput{
			DBProxyName:     aws.String(proxyName),
			TargetGroupName: aws.String(defaultTargetGroupName),
		},
	)
	if err != nil {
		// A proxy with no target group yet has nothing registered, which is the state
		// this link exists to change rather than a failure.
		if isTargetOrProxyGone(err) {
			return false, nil
		}

		return false, fmt.Errorf(
			"reading the targets registered with RDS proxy %q: %w", proxyName, err,
		)
	}

	for _, target := range output.Targets {
		if aws.ToString(target.RdsResourceId) == instanceIdentifier {
			return true, nil
		}
	}

	return false, nil
}

func isTargetOrProxyGone(err error) bool {
	var proxyNotFound *rdstypes.DBProxyNotFoundFault
	var targetGroupNotFound *rdstypes.DBProxyTargetGroupNotFoundFault
	var targetNotFound *rdstypes.DBProxyTargetNotFoundFault
	return errors.As(err, &proxyNotFound) ||
		errors.As(err, &targetGroupNotFound) ||
		errors.As(err, &targetNotFound)
}
