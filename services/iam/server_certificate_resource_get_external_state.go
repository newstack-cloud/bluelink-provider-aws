package iam

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/newstack-cloud/bluelink/libs/blueprint/core"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
)

func (a *iamServerCertificateResourceActions) GetExternalState(
	ctx context.Context,
	input *provider.ResourceGetExternalStateInput,
) (*provider.ResourceGetExternalStateOutput, error) {
	iamService, err := a.getIamService(ctx, input.ProviderContext)
	if err != nil {
		return nil, err
	}

	serverCertName, hasServerCertName := pluginutils.GetValueByPath(
		"$.serverCertificateName",
		input.CurrentResourceSpec,
	)
	if !hasServerCertName {
		return nil, fmt.Errorf("serverCertificateName is required for get external state operation")
	}

	serverCert, err := iamService.GetServerCertificate(ctx, &iam.GetServerCertificateInput{
		ServerCertificateName: aws.String(core.StringValue(serverCertName)),
	})
	if err != nil {
		return nil, err
	}

	// The private key is not accessible via the AWS service call for security reasons,
	// once you upload the private key to AWS, it can no longer be retrieved via the API.
	// Private keys are stored in state but are marked as sensitive in the schema to prevent them
	// from being displayed in tool UIs and logs.
	privateKey, hasPrivateKey := pluginutils.GetValueByPath(
		"$.privateKey",
		input.CurrentResourceSpec,
	)
	if !hasPrivateKey {
		return nil, fmt.Errorf(
			"privateKey is expected to be in the current state of the server certificate resource",
		)
	}

	return &provider.ResourceGetExternalStateOutput{
		ResourceSpecState: &core.MappingNode{
			Fields: map[string]*core.MappingNode{
				"arn": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.ServerCertificateMetadata.Arn),
				),
				"certificateBody": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.CertificateBody),
				),
				"certificateChain": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.CertificateChain),
				),
				"path": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.ServerCertificateMetadata.Path),
				),
				"privateKey": privateKey,
				"serverCertificateName": core.MappingNodeFromString(
					aws.ToString(serverCert.ServerCertificate.ServerCertificateMetadata.ServerCertificateName),
				),
				"tags": extractIAMTags(
					serverCert.ServerCertificate.Tags,
				),
			},
		},
	}, nil
}
