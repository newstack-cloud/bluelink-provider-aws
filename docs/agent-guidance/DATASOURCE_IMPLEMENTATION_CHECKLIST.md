# Data Source Implementation Checklist

This checklist is for use by background agents (or human contributors) to track progress and validate the quality of a new data source implementation. Mark each item as complete as you progress.

---

## 1. File and Structure Validation
- [ ] All required files are present:
  - `*_data_source.go`
  - `*_data_source_test.go`
  - `*_data_source_schema.go`
- [ ] Files are placed in the correct directory (`services/${service}/`)

## 2. Schema and Method Validation
- [ ] Data source schema matches the service definition schema (`definitions/services/${service}.yml`).
- [ ] The `Fetch` method is implemented with the correct signature and follows existing patterns.
- [ ] Uses `pluginutils` helpers for value extraction and nil checks.

## 3. Output and Filterable Fields
- [ ] Output fields are determined by reviewing the doc links provided in the schema and match the authoritative documentation.
- [ ] All `filterableFields` defined in the schema are supported in the implementation.

## 4. Test Coverage
- [ ] Test files exist for the data source.
- [ ] Tests cover:
  - [ ] Basic use cases
  - [ ] Complex/edge cases
  - [ ] Error cases (e.g., missing required filter fields, service call failures)
- [ ] Tests use mocks and test utils as per existing patterns.
- [ ] All tests pass locally.

## 5. Example and Documentation Validation
- [ ] At least one example is added to `services/${service}/examples/datasources/`.
- [ ] Example is referenced in the data source schema.
- [ ] Example uses correct code block formatting (e.g., ```javascript for JSONC).
- [ ] No unclosed code blocks.
- [ ] Description is present above the example.

## 6. Consistency and Pattern Adherence
- [ ] Implementation follows patterns from existing data sources in the same or similar service.
- [ ] No large blocks of repetitive nil checks—uses helpers instead.
- [ ] No deviation from naming conventions or file structure.
- [ ] No deprecated or discouraged patterns present.

## 7. Source of Truth and Reference Validation
- [ ] API calls used match those specified in the service definition schema and doc links.
- [ ] All required and filterable fields are handled as per the schema.
- [ ] No missing or extra fields in the data source spec.

## 8. Regression and Integration
- [ ] All existing tests in the project pass (no regressions introduced).
- [ ] Data source integrates cleanly with the rest of the provider (no import or dependency issues).

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