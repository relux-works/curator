# TASK-260811-3ksxig third rework report

Date: 2026-08-23

Role: developer

Review input: `TASK-260811-3ksxig_rework-after-RUN-260823-66398d.md`

## Outcome

The pnpm 10.33.0 materialized-layout validator now separates the physical
lock-superset graph from active selection. Every resolved dependency link of a
materialized snapshot is reconciled against the exact lock snapshot and peer
context even when that snapshot is target-pruned. Importer links, common Node
active graph reachability, returned active packages, and portable runtime
hydration remain selection-controlled.

The deterministic fake runner now emits the same snapshot-link superset as the
pinned manager. New fixtures prove:

- a target-pruned snapshot with a declared dependency materializes successfully
  without entering the active package set;
- missing, swapped, and unclaimed links below that target-pruned snapshot fail
  with `closure_graph_incomplete`; and
- an unreachable lock snapshot is rejected before the install process starts.

A real pinned pnpm 10.33.0 E2E derives the private store and performs frozen,
offline, scripts-disabled materialization for the target-pruned dependency
shape. The real manager omits wholly unreachable snapshots from the virtual
store, so the profile documents and enforces that narrower unsupported boundary
before install instead of accepting an unreconciled physical graph.

README profile documentation records the target-pruned behavior and unreachable
boundary.

## Validation

Every green gate below ran as a standalone process and exited 0.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 -cover ./internal/pnpmsource` | 0 | `go-test-cover-03.log`; 80.3% statements |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 -race ./internal/pnpmsource` | 0 | `go-test-race-03.log` |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 ./internal/pnpmsource ./internal/nodesource ./internal/artifactpolicy ./internal/closureexec` | 0 | `go-test-scoped-03.log` |
| `go vet ./internal/pnpmsource ./internal/nodesource ./internal/artifactpolicy ./internal/closureexec` | 0 | `go-vet-03.log` |
| `golangci-lint run ./internal/pnpmsource ./internal/nodesource ./internal/artifactpolicy ./internal/closureexec` | 0 | `golangci-lint-03.log`; pinned lint 2.12.2 |
| `go build ./...` | 0 | `go-build-03.log` |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 ./...` | 0 | `go-test-full-03.log` |
| `git diff --check` | 0 | `git-diff-check-03.log` |
| `task-board validate` | 0 | `task-board-validate-03.log`; board valid |

The closed-loop investigation also had three expected exploratory failures,
reported truthfully as failures rather than passes:

1. The first real target-pruned probe exited 1 because a direct optional
   importer fixture caused pnpm to retain an additional importer member; the
   fixture was narrowed to isolate snapshot-link behavior.
2. The second real target-pruned probe exited 1 because the selected parent
   snapshot retained its target-pruned optional link, proving physical snapshot
   links are a lock-superset rather than active-edge projection.
3. The combined real target-pruned/unreachable probe exited 1 only for the
   unreachable branch because pnpm omitted `dormant@1.0.0`; this supplied the
   evidence for the explicit pre-install unsupported boundary. The target-pruned
   branch passed in that same run.

## Scope and repository state

Changed for this rework:

- `internal/pnpmsource/materialize.go`
- `internal/pnpmsource/conformance_test.go`
- `README.md`

The worktree already contained extensive accepted changes from this delivery
story. They were preserved; nothing was staged or committed.

No forced-fit or external blocker remains. The task is ready for independent
review.
