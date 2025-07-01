# Link Implementation Checklist

This checklist is for use by background agents (or human contributors) to track progress and validate the quality of a new link implementation. Mark each item as complete as you progress.

---

## 1. File and Structure Validation
- [ ] All required files are present:
  - `*_link.go`
  - `*_link_update.go`
  - `*_link_update_test.go`
  - `*_link_stage_changes.go`
  - `*_link_stage_changes_test.go`
  - `*_link_annotations.go`
- [ ] Files are placed in the correct directory (`services/${service}/links/`)

## 2. Schema and Method Validation
- [ ] Link schema matches the service definition schema (`definitions/services/${service}.yml`).
- [ ] All required methods are implemented:
  - `UpdateResourceA`, `UpdateResourceB`, `UpdateIntermediaryResources`, `StageChanges`
- [ ] Method signatures match existing patterns.
- [ ] Uses `linkhelpers` and `pluginutils` helpers for value extraction and nil checks.

## 3. Annotation Handling
- [ ] Link annotations are defined and handled as per the schema and existing patterns.
- [ ] Any dynamic annotation keys (e.g., with <placeholder>) are implemented as required.
- [ ] Tests cover both static and dynamic annotation scenarios.

## 4. Test Coverage
- [ ] Test files exist for each method.
- [ ] Tests cover:
  - [ ] Basic use cases
  - [ ] Complex/edge cases
  - [ ] Error cases (e.g., missing required fields, service call failures)
- [ ] Tests use mocks and test utils as per existing patterns.
- [ ] All tests pass locally.

## 5. Rich Descriptions and Examples
- [ ] At least one rich description is added to `services/${service}/links/descriptions/`.
- [ ] Rich description is referenced in the main link definition file.
- [ ] Example uses correct code block formatting (e.g., ```javascript for JSONC).
- [ ] No unclosed code blocks.
- [ ] Description is present above the example.

## 6. Consistency and Pattern Adherence
- [ ] Implementation follows patterns from existing links in the same or similar service.
- [ ] No large blocks of repetitive nil checks—uses helpers instead.
- [ ] No deviation from naming conventions or file structure.
- [ ] No deprecated or discouraged patterns present.

## 7. Source of Truth and Reference Validation
- [ ] API calls used match those specified in the service definition schema and doc links.
- [ ] All required fields and annotations are handled as per the schema.
- [ ] No missing or extra fields in the link spec.

## 8. Regression and Integration
- [ ] All existing tests in the project pass (no regressions introduced).
- [ ] Link integrates cleanly with the rest of the provider (no import or dependency issues).

## 9. Linting and Formatting
- [ ] Code passes linting (e.g., `gofmt`, `golint`, or project-specific linter).
- [ ] No TODOs, commented-out code, or debug prints left in output.

## 10. Notes and Special Cases
- [ ] If any `notes` are present in the service definition, they are addressed in the implementation or tests.
- [ ] Any known quirks or edge cases are documented in comments or test cases.

---

**Instructions:**
- Use this checklist as a progress tracker during implementation.
- Mark each item as complete as you finish it.
- If any item cannot be completed, flag it for review or request clarification. 