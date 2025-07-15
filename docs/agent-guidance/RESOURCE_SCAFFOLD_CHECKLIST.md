# Resource Scaffolding Progress Checklist

Use this checklist to track progress when generating scaffolding for any new resource in any service. Replace `<service>` and `<resource>` with the appropriate names for your context.

## Planning & Structure
- [ ] Review existing resource implementations in `services/<service>/` for naming, structure, and conventions.
- [ ] Reference the service definition in `definitions/services/<service>.yml` and `definitions/schema.yml` for schema and field types.

## Schema
- [ ] Create the resource schema definition file: `services/<service>/<resource>_resource_schema.go`.
  - [ ] Define all resource fields, types, and validation.
  - [ ] Add a placeholder for examples.
  - [ ] Do not use convenience helpers for `core.MappingNode` arrays/maps; use `Items` and `Fields` directly.

## Implementation Stubs
- [ ] Create the main resource implementation file: `services/<service>/<resource>_resource.go`.
  - [ ] Implement a factory function returning the resource definition.
  - [ ] Define an actions struct `<resource>ResourceActions` with methods for each action.
  - [ ] Reference the shared examples embed.FS variable (do not embed each file individually).
  - [ ] Import the service package as `<service>service`.
  - [ ] Add stubs for `GetExternalStateFunc`, `CreateFunc`, `UpdateFunc`, `DestroyFunc`, `StabilisedFunc` that point to the action methods.
- [ ] Create action stub files:
  - [ ] `<resource>_resource_create.go`
  - [ ] `<resource>_resource_update.go`
  - [ ] `<resource>_resource_destroy.go`
  - [ ] `<resource>_resource_get_external_state.go`
  - [ ] `<resource>_resource_stabilised.go`
  - [ ] Each should have the correct method signature (matching existing resources), return empty values, and include a TODO comment.
- [ ] If following the ops pattern, create ops stub files:
  - [ ] `<resource>_resource_create_ops.go`
  - [ ] `<resource>_resource_update_ops.go`
  - [ ] These can be empty or contain TODOs.

## Test Stubs
- [ ] For each operation, create a test file:
  - [ ] `<resource>_resource_create_test.go`
  - [ ] `<resource>_resource_update_test.go`
  - [ ] `<resource>_resource_destroy_test.go`
  - [ ] `<resource>_resource_get_external_state_test.go`
  - [ ] `<resource>_resource_stabilised_test.go`
  - [ ] Each should:
    - [ ] Import `testing` and `testify`.
    - [ ] Define a test suite struct named `<resource>Resource<Action>Suite`.
    - [ ] Define a suite registration function named `Test<resource>ResourceActionsSuite`.
    - [ ] Include a TODO comment for future test logic.

## Examples
- [ ] In `services/<service>/examples/resources/`, create example markdown files:
  - [ ] `<service>_<resource>_basic.md` (YAML)
  - [ ] `<service>_<resource>_complete.md` (YAML)
  - [ ] `<service>_<resource>_jsonc.md` (JavaScript/JSONC)
  - [ ] Each example:
    - [ ] Has a description above the code block, no explanation section at the bottom.
    - [ ] Uses correct code block syntax (`yaml` for YAML, `javascript` for JSONC).
    - [ ] All code blocks are properly closed.
    - [ ] Follows structure and style of existing examples.
  - [ ] Ensure examples are embedded in the resource implementation file via the shared embed.FS variable.
  - [ ] Include the full resources section of blueprint file, see existing examples for reference.

## Registration
- [ ] Register the `<resource>` resource using the factory function in `provider/provider.go`.

## Final Checks
- [ ] All files are created, even if only a stub or with a TODO.
- [ ] All method and function signatures match those of equivalent actions in existing resources.
- [ ] All TODO comments are present where logic is to be implemented later.
- [ ] No operation logic is implemented at this stage—focus is only on structure and signatures.
- [ ] All placeholders, types, and references are correct for the new resource. 