package lambdalinks

import "github.com/newstack-cloud/celerity/libs/blueprint/provider"

func lambdaFunctionCodeSigningConfigLinkAnnotations() map[string]*provider.LinkAnnotationDefinition {
	// The relationship between a lambda function and a code signing config
	// does not have any annotations, so an empty map is returned.
	return map[string]*provider.LinkAnnotationDefinition{}
}
