## Celerity AWS Link Implementation Guide

### Overview

You need to create a new `${resourceTypeA}::${resourceTypeB}` link for the `${service}` service in the `services/${service}` package directory.
A resource enables users of Celerity to connect resources using link selectors and labels in blueprint files, links provide a powerful abstraction that allows users to define both simple and complex relationships between resources through declarative link selectors.

The links needs to be implemented following existing patterns and conventions in link implementations so far either in the `services/${service}/links` package or in another service implementation under the `services` directory in the `links` subdirectory.


### Plan

1. Create the link annotations definition file, carefully following existing link annotation definitions such as `services/lambda/function__function_link_annotations.go`. Annotations will be defined in the definitions schema file as mentioned in [sources](#sources).
2. Create the link implementation files, carefully following existing link implementations such as `services/lambda/links/function__function_link.go`.
3. Create the functionality for the `UpdateResourceA` method for the link, carefully following existing links, studying and reusing existing patterns and utils. See `services/lambda/links/function__function_link_update.go` for a guide. You must use the `pluginutils` package helpers to extract values from the spec data.
4. Implement an extensive test suite for the `UpdateResourceA` method, carefully following existing test suites for the `UpdateResourceA` method such as `services/lambda/links/function__function_link_update_test.go`. Study and reuse test utils and patterns from existing tests.
5. Create the functionality for the `UpdateResourceB` method for the link, carefully following existing links, studying and reusing existing patterns and utils. See `services/lambda/links/function__function_link_update.go` for a guide. You must use the `pluginutils` package helpers to extract values from the spec data.
6. Implement an extensive test suite for the `UpdateResourceB` method, carefully following existing test suites for the `UpdateResourceB` method such as `services/lambda/links/function__function_link_update_test.go`. Study and reuse test utils and patterns from existing tests.
7. Create the functionality for the `UpdateIntermediaryResources` method for the link, carefully following existing links, studying and reusing existing patterns and utils. See `services/lambda/links/function__function_link_update.go` for a guide. You must use the `pluginutils` package helpers to extract values from the spec data.
8. Implement an extensive test suite for the `UpdateIntermediaryResources` method, carefully following existing test suites for the `UpdateIntermediaryResources` method such as `services/lambda/links/function__function_link_update_test.go`. Study and reuse test utils and patterns from existing tests.
9. Create the functionality for the `StageChanges` method for the resource, carefully following existing links, studying and reusing existing patterns and utils. See `services/lambda/links/function__code_signing_config_link_stage_changes.go` for a guide. You must use the `linkhelpers` and `pluginutils` package helpers to extract values from the spec data and collect derived changes for the link data projection as a link is a projection of a subset of the state in resource A and resource B along with the full state of any intermediary resources required for the link.
10. Implement an extensive test suite for the `StageChanges` method, carefully following existing test suites for the `StageChanges` method such as `services/lambda/links/function__code_signing_config_link_stage_changes_test.go`. Study and reuse test utils and patterns from existing tests.
11. Validate the implementations and tests are complete and correct. Ensure that the tests are covering failures and edge cases as well as basic and complex use cases.
12. Check for any deviation from the patterns used in existing link implementations, study multiple existing link implementations.
13. Apply any corrections as a result of the analysis from step 11 and 12.
14. Study existing rich descriptions in the `services/${service}/links/descriptions` directory or in other `services/*/links/descriptions` directories.
15. Add a rich description for the new link to the `services/${service}/links/descriptions` directory and integrate the rich description into the main link definition file.

### Sources

You should use the service definition schema as a source of truth to guide the implementation of the link.
The service definition schema is located in the `definitions/services/${service}.yml` file and the structure of the schema is defined in the `definitions/schema.yml` file.

To determine the service call actions that are required in the link implementation, you should use the AWS API Reference docs for the `${service}` service. You can also use the AWS SDK v2 for Go docs to accurately determine the service call actions required.

You should thoroughly review the existing link implementations, taking note of patterns for value extraction, collecting changes for the link data projection and methods to access fields in `*core.MappingNode` objects or `map[string]*core.MappingNode` maps.
You should also thoroughly review the existing tests for link implementations to understand how to implement the tests for the new link, using the `plugintestutils` package helpers where possible.

### Setting up new service

If `${service}` does not exist in the `services` directory, you need to create a new directory for it in the `services` directory.
You should create a new service interface in the `services/${service}/service.go` file, using existing service interfaces as a guide.
You should prepare a mock for the service to be used in tests in the `internal/testutils/${service}_mock/${service}_service_mock.go` file, using existing service mock implementations as a guide.

### Link File Structure

The link implementation should be structured as follows:

- `*_link.go` - The main link implementation file.
- `*_link_update.go` - The update resource A, resource B and intermediary resources implementation.
- `*_link_update_test.go` - The update resource A, resource B and intermediary resources test implementation.
- `*_link_stage_changes.go` - The stage changes implementation.
- `*_link_stage_changes_test.go` - The stage changes test implementation.
- `*_link_annotations.go` - The link annotations definition file.

### Link Methods

Across the files mentioned in the previous section, you should implement the following methods:

- `UpdateResourceA` - The update resource A operation implementation.
- `UpdateResourceB` - The update resource B operation implementation.
- `UpdateIntermediaryResources` - The update intermediary resources operation implementation.
- `StageChanges` - The stage changes operation implementation.

### Tests

You should implement thorough tests that cover both basic and complex uses of the links along with error cases for missing IDs (or other required fields) and when the service method call returns an error.

You must provide tests for all the methods mentioned in the [Link Methods](#link-methods) section using the files defined in the [Link File Structure](#link-file-structure) section.

### Rich Descriptions

You should include a rich description that includes examples in the main link definition file.
Examples are defined in the `services/${service}/links/descriptions` directory and should be markdown files. You can use existing link examples as a guide.

Be sure to use the "```javascript ... ```" code block syntax for JSONC examples.

There is no need to add an explanation section at the bottom of the examples, only a description above the example code block(s).

Make sure you always close any code blocks that open with "\`\`\`" (usually followed by a language identifier) with "\`\`\`" on a new line.

You should inspect existing rich descriptions closely in the `services/${service}/links/descriptions` directory and other `services/*/links/descriptions` directories to understand how to structure the rich descriptions.

### About Link Annotations

Dynamic link annotation keys are defined in the form `aws.lambda.function.<targetFunction>.populateEnvVars` where `<targetFunction>` is the placeholder for the target function name.
There must only be one `<..>` placeholder in an annotation key and it can be located anywhere in the key.
The contents of the `<..>` placeholder can be anything but it must always point to a unique resource defined in the same blueprint file that is linked to the resource where the annotation is defined. This is then used to target a specific resource in the link that the given annotation should be applied to.

This must be considered when creating link implementations and using the annotations in the change staging and update operations.

See `services/lambda/links/function__function_link_annotations.go` for an example of how to define link annotations with dynamic keys.

### Extra notes

- You should avoid large code blocks with more than half a dozen nil checks for fields, instead, use the `linkhelpers`  and `pluginutils` package helpers to break down value extraction, as this will make the code more readable and maintainable.
- You must run the tests to ensure they are all passing before considering the task as complete.
- You must run existing tests in this project to ensure that regressions have not been introduced.
