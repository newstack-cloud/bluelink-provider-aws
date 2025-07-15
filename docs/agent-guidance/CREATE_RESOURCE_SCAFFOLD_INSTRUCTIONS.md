## Celerity AWS Resource Scaffolding Guide

### Overview

This guide describes how to scaffold a new `${resource}` resource for the `${service}` service in the `services/${service}` package directory. The goal is to generate only the resource schema, file structure, and method/test stubs—leaving all operation logic to be implemented later.

### Plan

1. **Create the resource schema definition file**
   - Follow patterns from existing schemas (e.g., `services/lambda/function_resource_schema.go`).
   - Define the resource fields, types, and validation, referencing the service definition schema in `definitions/services/${service}.yml` and `definitions/schema.yml`.
   - Add a placeholder for examples.

2. **Create the resource implementation files as stubs**
   - Main resource file: `${resource}_resource.go` with struct and method stubs for `Create`, `Update`, `Destroy`, `GetExternalState`, and `Stabilised`.
   - Operation files: `${resource}_resource_create.go`, `${resource}_resource_update.go`, `${resource}_resource_destroy.go`, `${resource}_resource_get_external_state.go`, `${resource}_resource_stabilised.go`—each containing only the function signature and a TODO comment.
   - If following the ops pattern, also create `${resource}_resource_create_ops.go` and `${resource}_resource_update_ops.go` as empty files or with TODOs.

3. **Create test scaffolding for each operation**
   - For each operation, create a corresponding test file (e.g., `${resource}_resource_create_test.go`).
   - Each test file should include a test function stub (e.g., `func Test${Resource}Create(t *testing.T) { /* TODO */ }`).
   - Import the testing package and add a TODO comment for future test logic.

4. **Set up example scaffolding**
   - In `services/${service}/examples/resources/`, create empty or placeholder markdown files for the resource (e.g., `${service}_${resource}_basic.md`).
   - In `examples_embed.go`, add a placeholder for embedding examples for the new resource.

5. **Register the resource**
   - Add a placeholder for registering the resource in `provider/provider.go`.

### File Structure

The resource scaffolding should include the following files:

- `${resource}_resource.go` - Main resource file with method stubs
- `${resource}_resource_schema.go` - Resource schema definition
- `${resource}_resource_create.go` - Create operation stub
- `${resource}_resource_create_test.go` - Create operation test stub
- `${resource}_resource_update.go` - Update operation stub
- `${resource}_resource_update_test.go` - Update operation test stub
- `${resource}_resource_destroy.go` - Destroy operation stub
- `${resource}_resource_destroy_test.go` - Destroy operation test stub
- `${resource}_resource_get_external_state.go` - Get external state operation stub
- `${resource}_resource_get_external_state_test.go` - Get external state operation test stub
- `${resource}_resource_stabilised.go` - Stabilised operation stub
- `${resource}_resource_stabilised_test.go` - Stabilised operation test stub
- `${resource}_resource_create_ops.go` - (if needed) Create ops stub
- `${resource}_resource_update_ops.go` - (if needed) Update ops stub
- `examples/resources/${service}_${resource}_basic.md` - Example placeholder
- `examples_embed.go` - Add placeholder for new resource

### Method and Test Stubs

- Each method in the main resource file and operation files should have the correct signature and a `TODO` comment.
- Each test file should import `testing` and have a single test function with a `TODO`.

### Examples

- Add a placeholder example markdown file in the examples directory.
- In the schema file, add a comment or placeholder for embedding examples.

### Notes

- Do not implement any operation logic at this stage—focus only on structure and signatures.
- Follow naming and file structure conventions from existing resources.
- Use TODO comments to indicate where logic will be added later.
- Ensure all files are created, even if empty or with only a stub.
- This approach enables iterative, logic-driven development after scaffolding is complete. 