package iam

import (
	"context"
	"embed"

	"github.com/aws/aws-sdk-go-v2/aws"

	iamservice "github.com/newstack-cloud/bluelink-provider-aws/services/iam/service"
	"github.com/newstack-cloud/bluelink-provider-aws/utils"
	"github.com/newstack-cloud/bluelink/libs/blueprint/provider"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/pluginutils"
	"github.com/newstack-cloud/bluelink/libs/plugin-framework/sdk/providerv1"
)

//go:embed examples/resources/*.md
var managedPolicyExamples embed.FS

// ManagedPolicyResource returns a resource implementation for an AWS IAM Managed Policy.
func ManagedPolicyResource(
	iamServiceFactory pluginutils.ServiceFactory[*aws.Config, iamservice.Service],
	awsConfigStore pluginutils.ServiceConfigStore[*aws.Config],
) provider.Resource {
	basicExample, _ := managedPolicyExamples.ReadFile("examples/resources/iam_managed_policy_basic.md")
	completeExample, _ := managedPolicyExamples.ReadFile("examples/resources/iam_managed_policy_complete.md")
	jsoncExample, _ := managedPolicyExamples.ReadFile("examples/resources/iam_managed_policy_jsonc.md")

	iamManagedPolicyActions := &iamManagedPolicyResourceActions{
		iamServiceFactory:   iamServiceFactory,
		awsConfigStore:      awsConfigStore,
		uniqueNameGenerator: utils.IAMPolicyNameGenerator,
	}
	return &providerv1.ResourceDefinition{
		Type:             "aws/iam/managedPolicy",
		Label:            "AWS IAM Managed Policy",
		PlainTextSummary: "A resource for managing an AWS IAM managed policy.",
		FormattedDescription: "The resource type used to define an [IAM managed policy](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_managed-vs-inline.html#aws-managed-policies) " +
			"that is deployed to AWS.",
		Schema:  iamManagedPolicyResourceSchema(),
		IDField: "arn",
		// A managed policy is commonly used by other resources that need to attach policies
		CommonTerminal: false,
		FormattedExamples: []string{
			string(basicExample),
			string(completeExample),
			string(jsoncExample),
		},
		ResourceCanLinkTo:    []string{},
		GetExternalStateFunc: iamManagedPolicyActions.GetExternalState,
		CreateFunc:           iamManagedPolicyActions.Create,
		UpdateFunc:           iamManagedPolicyActions.Update,
		DestroyFunc:          iamManagedPolicyActions.Destroy,
		StabilisedFunc:       iamManagedPolicyActions.Stabilised,
	}
}

type iamManagedPolicyResourceActions struct {
	iamServiceFactory   pluginutils.ServiceFactory[*aws.Config, iamservice.Service]
	awsConfigStore      pluginutils.ServiceConfigStore[*aws.Config]
	uniqueNameGenerator utils.UniqueNameGenerator
}

func (i *iamManagedPolicyResourceActions) getIamService(
	ctx context.Context,
	providerContext provider.Context,
) (iamservice.Service, error) {
	awsConfig, err := i.awsConfigStore.FromProviderContext(
		ctx,
		providerContext,
		nil,
	)
	if err != nil {
		return nil, err
	}

	return i.iamServiceFactory(awsConfig, providerContext), nil
}
