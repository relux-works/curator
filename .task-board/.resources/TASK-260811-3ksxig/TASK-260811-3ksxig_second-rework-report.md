# TASK-260811-3ksxig second rework report

Status: prepared for review

## Reviewer findings addressed

1. Workspace importer identities now use one canonical `local:<path>` key from lock parsing through selection and the common Node capture. Tests prove a workspace-only dependency is present in both the common capture records and active reachability.
2. Materialization now owns the complete root and workspace `node_modules` layouts. It admits only exact direct links, exact virtual-store snapshot and peer links, and the pinned closed metadata set. Metadata-less packages, valid undeclared packages, regular files, malformed metadata, missing links, and wrong links fail with `closure_graph_incomplete`.
3. `pnpm-source-v1` now admits only pnpm `10.33.0`; another manager version fails with `closure_runtime_identity_changed` before any process starts. A task-local real `pnpm@10.33.0` integration test exercises private-store derivation, frozen/offline/scripts-disabled materialization, exact layout reconciliation, and C0-bound Node invocation.

## Real-manager corrections

- Pinned pnpm does not accept `--offline` or `--ignore-scripts` as `store add` options. Store derivation supplies those controls through the closed environment and uses only admitted local tarballs.
- `store add` indexes local tarballs by their file locator. The adapter deterministically reconciles those indexes to the lock package IDs using pnpm 10.33.0's integrity-prefix and package-ID filename algorithm before freezing the store.
- pnpm writes a project registration under `v10/projects` even during offline install. The receipted store remains read-only authority; installation uses a declared ephemeral writable copy. The adapter verifies the exact project-registration symlink, removes it, rejects every other overlay member, and requires all frozen store bytes to remain unchanged.
- pnpm materializes lock-superset snapshot directories even when a target condition prunes a direct link. The adapter owns and byte-validates the complete lock snapshot set while returning only selected snapshots as the active materialized package set.
- Dereferencing pnpm links into retained source previously broke Node dependency resolution. The retained link-free representation now deterministically nests exact snapshot-context dependencies and handles runtime cycles through Node's ancestor lookup.
- Shared portable replay verification now sorts observed files by canonical full path before comparison. This fixes false mutation reports for a `.pnpm/` directory beside `.pnpm-workspace-state-v1.json`; a focused regression test covers the case.

## Files changed in this rework

- `internal/pnpmsource/lock.go`
- `internal/pnpmsource/capture.go`
- `internal/pnpmsource/materialize.go`
- `internal/pnpmsource/errors.go`
- `internal/pnpmsource/conformance_test.go`
- `internal/closureexec/portable_runner.go`
- `internal/closureexec/closureexec_test.go`
- `README.md`

## Validation

Every command ran as a standalone process and reported its real exit code:

- focused pnpm coverage: exit 0, 80.1%
- pnpm race test: exit 0
- artifact/closureexec/nodesource/pnpmsource suite: exit 0
- scoped `go vet`: exit 0
- scoped `golangci-lint`: exit 0, 0 issues
- `go build ./...`: exit 0
- uncached `go test -count=1 ./...` with pinned pnpm on `PATH`: exit 0
- `git diff --check`: exit 0
- `task-board validate`: exit 0

Detailed command evidence is in `.temp/TASK-260811-3ksxig/validation-02.log`.
