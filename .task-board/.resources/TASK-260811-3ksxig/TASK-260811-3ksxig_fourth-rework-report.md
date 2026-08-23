# TASK-260811-3ksxig fourth rework report

Date: 2026-08-23

Role: developer

Review input: `TASK-260811-3ksxig_rework-after-RUN-260823-9ab6de.md`

## Outcome

Pnpm snapshot graph reachability is now independent from active target
selection and its visible prune reason. Every snapshot records whether it is
reachable from the complete importer/snapshot lock graph before OS, CPU, libc,
development, or active-root pruning is applied.

Materialization rejects every wholly unreachable snapshot before the install
process starts by checking this independent reachability fact. A simultaneous
`os_mismatch`, `cpu_mismatch`, or `libc_mismatch` can no longer mask the
unsupported unreachable shape. Target-pruned snapshots that remain lock-graph
reachable continue to materialize and reconcile their complete physical
dependency and peer links without entering the common active graph.

Regression coverage proves all three target-selector overlaps with a fake
pinned-contract runner and a zero-install-start assertion. A real pnpm 10.33.0
fixture exercises the OS-overlap path: private-store derivation runs, the
pre-install reachability gate rejects, and no manager install launch occurs.
README profile documentation now states that lock reachability and target
selection are separate authorities.

## Closed-loop evidence

The first focused test was intentionally red and exited 1 because `Snapshot`
did not yet expose independent reachability. Its log is
`red-overlap-01.log`. After the implementation, the focused fake/real pinned
tests exited 0.

Every green gate below ran as a standalone process and returned its real exit
code.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 ./internal/pnpmsource -run 'Test(RealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall\|TargetPrunedUnreachableSnapshotRejectsBeforeInstall\|LockSupersetSnapshotDependencyLinksAreReconciledIndependentlyOfSelection)'` | 0 | `real-overlap-01.log` |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 -cover ./internal/pnpmsource` | 0 | `go-test-cover-04.log`; 80.5% statements |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 -race ./internal/pnpmsource` | 0 | `go-test-race-04b.log` |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 ./internal/pnpmsource ./internal/nodesource ./internal/artifactpolicy ./internal/closureexec` | 0 | `go-test-scoped-04.log` |
| `go vet ./internal/pnpmsource ./internal/nodesource ./internal/artifactpolicy ./internal/closureexec` | 0 | `go-vet-04.log` |
| `golangci-lint run ./internal/pnpmsource ./internal/nodesource ./internal/artifactpolicy ./internal/closureexec` | 0 | `golangci-lint-04.log` |
| `go build ./...` | 0 | `go-build-04.log` |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 ./...` | 0 | `go-test-full-04.log` |
| `git diff --check` | 0 | `git-diff-check-04.log` |
| `task-board validate` | 0 | `task-board-validate-04.log` |

## Scope and repository state

Changed in this rework:

- `internal/pnpmsource/lock.go`
- `internal/pnpmsource/materialize.go`
- `internal/pnpmsource/conformance_test.go`
- `README.md`

The existing dirty worktree and all unrelated delivery-story changes were
preserved. No files were staged or committed. No forced-fit or external
blocker remains; this rework is ready for independent review.
