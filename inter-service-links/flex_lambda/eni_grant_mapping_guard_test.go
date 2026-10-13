//go:build unit

package flexlambda

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	ec2mock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/ec2_mock"
	iammock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/iam_mock"
	lambdamock "github.com/newstack-cloud/bluelink-provider-aws/internal/testutils/lambda_mock"
	"github.com/newstack-cloud/bluelink-provider-aws/linkutils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/stretchr/testify/suite"
)

// Guards the promise the ENI grant makes about its own contribution: the mapping path it
// records selects the statement by @.sid, so the statement it puts in the link data has to
// carry a sid.
//
// The grant recorded it with IAM's JSON casing, "Sid", which was inert until composition
// began reading the mappings and then failed staging for every instance that had placed a
// function in a VPC.
type ENIGrantMappingGuardSuite struct {
	suite.Suite
}

// A placement, run for its link data and mappings rather than for the attachment.
func (s *ENIGrantMappingGuardSuite) placementOutput() *provider.LinkUpdateLinkedResourcesOutput {
	ec2Svc := ec2mock.CreateEc2ServiceMock(
		ec2mock.WithDescribeSecurityGroupsOutput(&ec2.DescribeSecurityGroupsOutput{}),
		ec2mock.WithCreateSecurityGroupOutput(&ec2.CreateSecurityGroupOutput{
			GroupId: aws.String("sg-workload"),
		}),
		ec2mock.WithRevokeSecurityGroupEgressOutput(&ec2.RevokeSecurityGroupEgressOutput{}),
		ec2mock.WithAuthorizeSecurityGroupEgressOutput(&ec2.AuthorizeSecurityGroupEgressOutput{}),
	)
	lambdaSvc := lambdamock.CreateLambdaServiceMock(
		lambdamock.WithGetFunctionOutput(vfGetFunctionOutput()),
		lambdamock.WithUpdateFunctionConfigurationOutput(nil),
	)
	iamSvc := iammock.CreateIamServiceMock(
		iammock.WithListRolePoliciesOutput(&iam.ListRolePoliciesOutput{PolicyNames: []string{}}),
		iammock.WithListAttachedRolePoliciesOutput(&iam.ListAttachedRolePoliciesOutput{}),
		iammock.WithPutRolePolicyOutput(&iam.PutRolePolicyOutput{}),
	)

	output, err := placementLinkWithEC2(iamSvc, lambdaSvc, ec2Svc).UpdateLinkedResources(
		context.Background(),
		&provider.LinkUpdateLinkedResourcesInput{
			LinkUpdateType:  provider.LinkUpdateTypeCreate,
			ResourceAInfo:   flexVPCResourceInfoAWithGateways("igw-1", "nat-1"),
			ResourceBInfo:   functionResourceInfoB(),
			LinkContext:     testLinkContext(),
			LinkID:          vfLinkID,
			ResourceService: vfRoleService(),
		},
	)
	s.Require().NoError(err)
	s.Require().NotNil(output)

	return output
}

func (s *ENIGrantMappingGuardSuite) Test_every_mapping_selects_the_data_it_points_at() {
	output := s.placementOutput()

	s.Require().NotEmpty(output.ResourceDataMappings, "the grant must record a mapping")
	s.Assert().Empty(linkutils.ValidateResourceDataMappings(
		output.ResourceDataMappings,
		output.LinkData,
	))
}

func (s *ENIGrantMappingGuardSuite) Test_the_statement_uses_the_resource_spec_casing() {
	output := s.placementOutput()

	roleLinkData := output.LinkData.Fields[fmt.Sprintf(
		"%sExecutionRole",
		functionResourceInfoB().ResourceName,
	)]
	s.Require().NotNil(roleLinkData, "the grant must record the statement under the role")
	statement := roleLinkData.Fields[linkutils.PermissionFieldName]
	s.Require().NotNil(statement)

	// Field names spelled out literally rather than taken from the builder, so putting the
	// capital back fails here. The IAM role spec schema declares these lowercase; the
	// capitalised form belongs only in the policy JSON sent to AWS.
	s.Require().NotNil(statement.Fields)
	for _, field := range []string{"sid", "effect", "action", "resource"} {
		s.Assert().Contains(statement.Fields, field)
	}
	for _, field := range []string{"Sid", "Effect", "Action", "Resource"} {
		s.Assert().NotContains(statement.Fields, field)
	}
	s.Assert().NotEmpty(core.StringValue(statement.Fields["sid"]))
}

func TestENIGrantMappingGuardSuite(t *testing.T) {
	suite.Run(t, new(ENIGrantMappingGuardSuite))
}
