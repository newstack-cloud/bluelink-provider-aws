# Bluelink AWS Provider

The Bluelink provider plugin for AWS: resources, data sources, links and functions for AWS
services. bluelink-transformer-celerity builds on it, the Celerity CLI installs it, and the
Bluelink Registry publishes its docs. The GitHub repo is `newstack-cloud/bluelink-provider-aws`.

## Layout

| Path | Holds |
|---|---|
| `services/cloudcontrol` | Generic Cloud Control engine that serves most resources and data sources |
| `services/cloudcontrol/gen` | Generated per-type resources and data sources; don't edit by hand |
| `services/cloudcontrol/overlays` | Hand-written per-type adjustments, re-applied on each generation |
| `tools/awsgen` | Generator for `gen/`, with vendored CloudFormation schemas in `schemas/` |
| `services/<service>` | Service clients, hand-written resources and data sources (`ssm`, `lambda` data sources), intra-service links (`links/`), examples |
| `flex` | Flex resources: higher-level abstractions with presets (`aws/flex/vpc`) |
| `inter-service-links/<a>_<b>` | Links between resources of different services |
| `definitions` | YAML definitions for hand-written resources and links, against `definitions/schema.yml` |
| `provider/provider.go` | Registration of every resource, data source and link |
| `internal/testutils` | Service mocks for unit tests |
| `tests/e2e` | End-to-end suite: blueprints in `testdata/` deployed to a real AWS account |
| `docs/agent-guidance` | Step-by-step guides and checklists for resources, data sources and links |

How generation works, and how to add a Cloud Control type: `docs/CLOUD_CONTROL_RESOURCES.md`.

## Adding a resource, data source or link

- Prefer a Cloud Control type over a hand-written resource, unless Cloud Control can't serve
  it (`docs/CLOUD_CONTROL_RESOURCES.md`).
- For a hand-written resource, data source or link, write its YAML definition first, and
  present it for approval before writing any Go.
- Then follow the guide and checklist in `docs/agent-guidance/` for its kind. Link design
  rules (annotation placement and naming, intermediary resources, resource data mappings) are
  in `CREATE_LINK_INSTRUCTIONS.md` and `LINK_DESIGN_GUIDANCE.md`.
- Register it in `provider/provider.go`, and add examples in `services/<service>/examples/`.

## Checks

- **Tools:** Go 1.26, `golangci-lint` v2, and Node with Yarn 4 for the commit-msg hook (run
  `yarn install` first).
- **Lint:** `golangci-lint run ./...`: the standard linters plus those `.golangci.yml` enables.
- **Build tags:** test files are tagged `unit` or `integration`, except `tools/awsgen`'s. Plain
  `go test ./...` runs only those; pass `-tags=unit`. A focused run:
  `go test -tags=unit -count=1 ./services/sqs/...`.
- **`scripts/run-tests.sh` runs unit and integration tests by default,** and integration
  tests deploy to a real AWS account. The script exports `.env.test` (from
  `.env.test.example`) if it exists, but access keys already exported in the shell
  (`AWS_ACCESS_KEY_ID`, `AWS_SECRET_ACCESS_KEY`) take precedence over its `AWS_PROFILE`.
  Without `.env.test`, the shell's credentials are used; without `AWS_REGION`, the e2e tests
  skip. Pass `--unit` for unit tests only.
  - `--integration` runs the e2e suite in `tests/e2e` (resources, data sources, links, the
    flex VPC) and a Lambda layer version data source test. It creates and destroys real
    resources, takes up to an hour, and costs money. `E2E_SLOW=1` adds an Aurora cluster.
    When the test process finishes normally, it then deletes resources the run created and
    left behind, and fails if it found any. A timeout or a killed process skips that sweep
    and each test's teardown, and can leave resources in AWS to delete by hand.
  - `--run <pattern>` passes `-run` to `go test`. Combine it with `--unit`, or it also runs
    integration tests.
- **Generated code:** `go run ./tools/awsgen` regenerates `services/cloudcontrol/gen`, which is
  committed. `go run ./tools/awsgen -sync` re-vendors schemas from AWS and needs credentials.
- **Gates:**
  - With `git config core.hooksPath .githooks`, pre-commit runs `golangci-lint run --fix` over
    the whole module, which rewrites files without staging them, and commit-msg runs
    commitlint.
  - CI (`ci.yaml`) runs unit tests, lint and SonarQube on pull requests and pushes to `main`.
  - `integration-tests.yaml` runs the integration tests every Monday and by hand, in
    NewStack's AWS account. Nothing runs them on a pull request.

## Running locally

The provider runs as a plugin of the deploy engine. To try a change in a running engine, the
plugin binary is built into the engine's plugin path
(`BLUELINK_DEPLOY_ENGINE_PLUGIN_PATH`, under `providers/newstack-cloud/aws/<version>/plugin`).
See Safety.

## Compatibility

- **Plugin framework:** built against the `libs/plugin-framework` version in `go.mod`. A
  framework interface change reaches this repo by Renovate pin bump (`cross-repo`).
- **Blueprint surface:** resource types, spec fields, data source filters and exports, link
  annotation names and their defaults are what blueprints are written against. Provisional
  (2026-11-22): changes are additive; a rename or removal is a breaking change for blueprints
  and needs the person's agreement.
- **Persisted state:** resource specs and link data are stored by the deploy engine. No
  deployed engines exist yet (as of 2026-10-21), so their shape can change without a
  migration; say in the pull request what changed.
- **Registry docs:** releases publish `docs.json` from `bluelink-plugin-docgen` (pinned in
  `release.yaml`), which the Bluelink Registry reads.

## Conventions

- Commits: `docs/COMMIT_GUIDELINES.md`. Scopes are free-form (`fix(links): ...`).
- Releases: svu sets the version and git-cliff writes the changelog, through a release pull
  request on `main` (`docs/SOURCE_CONTROL_RELEASE_STRATEGY.md`). Don't edit `CHANGELOG.md`
  or create tags.
- Examples: each example file holds the same blueprint in `blueprintlang`, then `yaml`, then
  `javascript` (JSONC) blocks, as `docs/agent-guidance/CREATE_RESOURCE_INSTRUCTIONS.md` sets
  out.
- Tags: the Cloud Control engine handles Bluelink's provenance tags for generated types. A
  hand-written resource sends user tags merged with them in the create call itself
  (`utils.MergeBluelinkTagsWithUserTags`), and leaves them out of `GetExternalState`
  (`utils.IsBluelinkTag`, `FilterTags`, `FilterTagsMap`). `utils/bluelink_tags.go` also builds
  a lookup by provenance tags for when there's no external ID.

## Safety

- Don't run `scripts/run-tests.sh` without `--unit`, or anything with the `integration` tag,
  `go run ./tools/awsgen -sync`, or the integration workflow, unless asked. Each reaches a
  real AWS account.
- Don't build the plugin into a deploy engine's plugin path or restart an engine: it may be
  serving end-to-end runs. Say when a change needs a rebuild, and wait.
- Don't create tags or releases, or merge release pull requests.

## Known mistakes

<!-- One line per mistake an agent has made here, phrased as the rule that prevents it. -->
