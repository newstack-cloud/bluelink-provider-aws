package rdsmock

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	rdsservice "github.com/newstack-cloud/bluelink-provider-aws/services/rds/service"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/plugintestutils"
)

type rdsServiceMock struct {
	plugintestutils.MockCalls

	registerTargetsOutput *rds.RegisterDBProxyTargetsOutput
	registerTargetsError  error

	deregisterTargetsOutput *rds.DeregisterDBProxyTargetsOutput
	deregisterTargetsError  error

	describeTargetsOutput *rds.DescribeDBProxyTargetsOutput
	describeTargetsError  error

	modifyTargetGroupOutput *rds.ModifyDBProxyTargetGroupOutput
	modifyTargetGroupError  error
}

// CreateRDSServiceMock creates a new instance of the RDS service mock with the provided options.
func CreateRDSServiceMock(options ...RDSServiceMockOption) *rdsServiceMock {
	mock := &rdsServiceMock{}
	for _, option := range options {
		option(mock)
	}
	return mock
}

// CreateRDSServiceMockFactory returns a service factory bound to a single mock instance.
func CreateRDSServiceMockFactory(
	options ...RDSServiceMockOption,
) func(awsConfig *aws.Config, providerContext provider.Context) rdsservice.Service {
	mock := CreateRDSServiceMock(options...)
	return func(awsConfig *aws.Config, providerContext provider.Context) rdsservice.Service {
		return mock
	}
}

// RDSServiceMockOption is a function type for configuring the RDS service mock.
type RDSServiceMockOption func(*rdsServiceMock)

func WithRegisterDBProxyTargetsOutput(output *rds.RegisterDBProxyTargetsOutput) RDSServiceMockOption {
	return func(mock *rdsServiceMock) { mock.registerTargetsOutput = output }
}

func WithRegisterDBProxyTargetsError(err error) RDSServiceMockOption {
	return func(mock *rdsServiceMock) { mock.registerTargetsError = err }
}

func WithDeregisterDBProxyTargetsOutput(output *rds.DeregisterDBProxyTargetsOutput) RDSServiceMockOption {
	return func(mock *rdsServiceMock) { mock.deregisterTargetsOutput = output }
}

func WithDeregisterDBProxyTargetsError(err error) RDSServiceMockOption {
	return func(mock *rdsServiceMock) { mock.deregisterTargetsError = err }
}

func WithDescribeDBProxyTargetsOutput(output *rds.DescribeDBProxyTargetsOutput) RDSServiceMockOption {
	return func(mock *rdsServiceMock) { mock.describeTargetsOutput = output }
}

func WithDescribeDBProxyTargetsError(err error) RDSServiceMockOption {
	return func(mock *rdsServiceMock) { mock.describeTargetsError = err }
}

func WithModifyDBProxyTargetGroupOutput(output *rds.ModifyDBProxyTargetGroupOutput) RDSServiceMockOption {
	return func(mock *rdsServiceMock) { mock.modifyTargetGroupOutput = output }
}

func WithModifyDBProxyTargetGroupError(err error) RDSServiceMockOption {
	return func(mock *rdsServiceMock) { mock.modifyTargetGroupError = err }
}

func (m *rdsServiceMock) RegisterDBProxyTargets(
	ctx context.Context,
	params *rds.RegisterDBProxyTargetsInput,
	optFns ...func(*rds.Options),
) (*rds.RegisterDBProxyTargetsOutput, error) {
	m.RegisterCall(ctx, params)
	if m.registerTargetsError != nil {
		return nil, m.registerTargetsError
	}
	return m.registerTargetsOutput, nil
}

func (m *rdsServiceMock) DeregisterDBProxyTargets(
	ctx context.Context,
	params *rds.DeregisterDBProxyTargetsInput,
	optFns ...func(*rds.Options),
) (*rds.DeregisterDBProxyTargetsOutput, error) {
	m.RegisterCall(ctx, params)
	if m.deregisterTargetsError != nil {
		return nil, m.deregisterTargetsError
	}
	return m.deregisterTargetsOutput, nil
}

func (m *rdsServiceMock) DescribeDBProxyTargets(
	ctx context.Context,
	params *rds.DescribeDBProxyTargetsInput,
	optFns ...func(*rds.Options),
) (*rds.DescribeDBProxyTargetsOutput, error) {
	m.RegisterCall(ctx, params)
	if m.describeTargetsError != nil {
		return nil, m.describeTargetsError
	}
	return m.describeTargetsOutput, nil
}

func (m *rdsServiceMock) ModifyDBProxyTargetGroup(
	ctx context.Context,
	params *rds.ModifyDBProxyTargetGroupInput,
	optFns ...func(*rds.Options),
) (*rds.ModifyDBProxyTargetGroupOutput, error) {
	m.RegisterCall(ctx, params)
	if m.modifyTargetGroupError != nil {
		return nil, m.modifyTargetGroupError
	}
	return m.modifyTargetGroupOutput, nil
}
