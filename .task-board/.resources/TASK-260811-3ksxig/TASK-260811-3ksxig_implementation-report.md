# TASK-260811-3ksxig implementation report

Date: 2026-08-23

## Outcome

Implemented `pnpm-source-v1` for pinned `pnpm-lock.yaml` schema `9.0`.
The adapter now:

- parses a closed lock/configuration grammar and reconciles importers,
  workspaces, snapshots, exact peer contexts, overrides, declared patches,
  target selectors, and root/workspace manifests;
- captures and admits raw registry tarballs, contained local roots, and patch
  files independently;
- derives a task-private pnpm store solely from admitted tarballs, rejects
  side-effects state, receipts the store, and freezes it read-only;
- materializes with frozen lock, offline mode, lifecycle scripts disabled,
  side-effects disabled, and copy-only imports;
- rechecks the lock, local roots, store receipt, installed snapshot graph,
  package metadata, exact owned file inventory, and recursive artifact policy
  before admitting the materialized tree; and
- invokes only the exact Node runtime binding through the protected executor.

The profile rejects `.pnpmfile.*`, unknown/custom resolver or fetcher config,
undeclared patches, escaping or uncaptured `file:` roots, stale locks,
integrity/origin/metadata drift, lifecycle/native build requirements, compiled
or opaque dependency payloads, side-effects cache data, and ambient-store
fallback.

Shared artifact admission was extended with closed plain-text declarations for
`.npmrc`, `.patch`, and `.diff`. README tooling and profile documentation were
updated.

## Test coverage

The new conformance suite covers shared S01-S08 semantics applicable to this
leaf and pnpm N01-N13 variants, including canonical ordering, target pruning,
peer resolution, independently captured local sources and patches, extension
rejection, missing/tampered artifacts, compiled nested payloads, protected
store creation, exact install flags, zero-start failures, side-effect output,
poisoned ambient state, materialized graph reconciliation, and local-root
drift. Package statement coverage is 79.3%.

## Validation evidence

Every command below was run directly as a standalone gate; no pipeline masked
its status.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 -cover ./internal/pnpmsource` | 0 | pass, 79.3% statement coverage |
| `go test -count=1 ./internal/pnpmsource` | 0 | pass |
| `go test -race -count=1 ./internal/pnpmsource` | 0 | pass |
| `go test -count=1 ./internal/artifactpolicy ./internal/nodesource ./internal/pnpmsource` | 0 | pass |
| `go vet ./internal/pnpmsource ./internal/artifactpolicy ./internal/nodesource` | 0 | pass |
| `golangci-lint run ./internal/pnpmsource ./internal/artifactpolicy` | 0 | `0 issues.` |
| `go build ./...` | 0 | pass |
| `go test -count=1 ./...` | 0 | pass; `cmd/curator` 421.804s |
| `git diff --check` | 0 | pass |
| `pnpm --version` | 127 | expected environment limitation: pnpm is not installed |

The tests use the protected executor with a deterministic pnpm runner that
materializes only from admitted tarballs and records exact permits, input
mounts, work copies, argv, writes, and process starts. A real pnpm executable
smoke was not run because no pnpm or Corepack binary is installed locally; no
claim of that unavailable gate is made.

## Files in scope

- `internal/pnpmsource/{errors,lock,capture,materialize}.go`
- `internal/pnpmsource/conformance_test.go`
- `internal/artifactpolicy/text.go`
- `internal/artifactpolicy/text_metadata_test.go`
- `go.mod`, `go.sum`, and `README.md`

No files were staged or committed.
