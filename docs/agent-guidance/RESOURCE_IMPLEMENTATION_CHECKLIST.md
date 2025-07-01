# Resource Implementation Checklist

This checklist is for use by background agents (or human contributors) to track progress and validate the quality of a new resource implementation. Mark each item as complete as you progress.

---

## 1. File and Structure Validation
- [ ] All required files are present:
  - `*_resource.go`
  - `*_resource_create.go` & `*_resource_create_test.go`
  - `*_resource_update.go` & `*_resource_update_test.go`
  - `*_resource_destroy.go` & `*_resource_destroy_test.go`
  - `*_resource_get_external_state.go` & `*_resource_get_external_state_test.go`
  - `*_resource_stabilised.go` & `*_resource_stabilised_test.go`
  - `*_resource_schema.go`
- [ ] Files are placed in the correct directory (`services/${service}/`)

## 2. Schema and Method Validation
- [ ] Resource schema matches the service definition schema (`definitions/services/${service}.yml`).
- [ ] All required methods are implemented:
  - `Create`, `Update`, `Destroy`, `GetExternalState`, `Stabilised`
- [ ] Method signatures match existing patterns.
- [ ] Uses `pluginutils` helpers for value extraction and nil checks.

## 3. Test Coverage
- [ ] Test files exist for each method.
- [ ] Tests cover:
  - [ ] Basic use cases
  - [ ] Complex/edge cases
  - [ ] Error cases (e.g., missing required fields, service call failures)
- [ ] Tests use mocks and test utils as per existing patterns.
- [ ] All tests pass locally.

## 4. Example and Documentation Validation
- [ ] At least one example is added to `services/${service}/examples/resources/`.
- [ ] Example is referenced in the resource schema.
- [ ] Example uses correct code block formatting (e.g., ```javascript for JSONC).
- [ ] No unclosed code blocks.
- [ ] Description is present above the example.

## 5. Consistency and Pattern Adherence
- [ ] Implementation follows patterns from existing resources in the same or similar service.
- [ ] No large blocks of repetitive nil checks—uses helpers instead.
- [ ] No deviation from naming conventions or file structure.
- [ ] No deprecated or discouraged patterns present.

## 6. Source of Truth and Reference Validation
- [ ] AWS API calls used match those specified in the service definition schema and doc links.
- [ ] All required fields are handled as per the schema.
- [ ] No missing or extra fields in the resource spec.

## 7. Regression and Integration
- [ ] All existing tests in the project pass (no regressions introduced).
- [ ] Resource integrates cleanly with the rest of the provider (no import or dependency issues).

## 8. Linting and Formatting
- [ ] Code passes linting (e.g., `gofmt`, `golint`, or project-specific linter).
- [ ] No TODOs, commented-out code, or debug prints left in output.

## 9. Notes and Special Cases
- [ ] If any `notes` are present in the service definition, they are addressed in the implementation or tests.
- [ ] Any known quirks or edge cases are documented in comments or test cases.

---

**Instructions:**
- Use this checklist as a progress tracker during implementation.
- Mark each item as complete as you finish it.
- If any item cannot be completed, flag it for review or request clarification. 