# TASK-260720-31nl14 review cycle 5 rework validation

## Resolution

The staged-sidecar durability gap is closed in `internal/transaction` without changing install orchestration or any consumer API.

- `Engine` now binds the platform-native directory durability primitive when constructed.
- `Prepare` syncs the staged sidecar's containing directory after the staged file or directory tree has been synced and re-digested, and before it durably records `PhasePrepared`.
- The sync is performed for every staged target, including regular-file and directory targets.
- A failed staged-parent sync enters the existing deterministic preparation-abort path; it never publishes a prepared journal, never mutates the live target, and removes the incomplete staged sidecar.

## Regression coverage

`internal/transaction/preparation_durability_test.go` adds file and directory cases that:

- observe the exact staged-parent durability primitive;
- prove the durable journal is still `preparing` when the primitive runs;
- prove the staged digest already matches the desired digest before the parent sync;
- run the native platform directory sync before allowing preparation to return `prepared`; and
- inject failure from the durability primitive and prove no prepared journal or staged sidecar is exposed and the live preimage is unchanged.

All earlier transaction fault, rollback, recovery, cleanup-manifest, namespace, subprocess, and race coverage remains green.

## Provenance and scope

- Worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/worktree`
- Exact base: `origin/main` commit `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`
- Review-cycle-5 product delta: `internal/transaction/engine.go` plus `internal/transaction/preparation_durability_test.go`
- No file was staged or committed.
- The reviewer-confirmed manager-lock candidate provenance is unchanged.
- `install.Project`, `install.Global`, CLI orchestration, GC, and build execution were not refactored.

## Verification

Passed on native Darwin/arm64:

- focused file/directory durability regressions, repeated 10 times;
- `go test ./internal/transaction -count=1`;
- `go test -race ./internal/transaction -count=1`;
- `go vet ./internal/transaction`;
- focused coverage: 77.3% of statements;
- subprocess crash recovery at swap boundaries, repeated 5 times;
- `make check` across the complete imported product graph using the established task-local module overlay;
- `go test -race ./... -count=1` across the complete imported product graph using the same overlay;
- native `make build` and `curator --version` runtime smoke (`curator dev`);
- complete Linux/amd64 compile graph using `go test -exec=true ./...`;
- complete Windows/amd64 compile graph using `go test -exec=true ./...`;
- golangci-lint v2.4.0 on `internal/transaction`: 0 issues;
- `gofmt -l internal/transaction`, `git diff --check`, and staged-file checks: clean.

The detached exact-base worktree intentionally leaves the tracked testing-tool submodule unmaterialized. Plain repository-wide `make check` and `go test -race ./...` therefore cannot resolve `internal/ui`; the successful full gates used `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-31nl14/go.test.mod`, whose only semantic difference is an absolute replacement to the canonical project checkout of that same tracked testing module.

Native Windows runtime execution is unavailable on this Darwin host. Windows cross-compilation is not runtime evidence, and the inherited `TASK-260720-1zl1cj` native Windows qualification gate remains preserved.

`task-board validate` still reports the same 13 inherited board-wide issues: 12 legacy `EPIC-260712-*` broken dependency links and orphan `.resources/TASK-260713-7a9c1e/review.md`. No issue belongs to `TASK-260720-31nl14`.
