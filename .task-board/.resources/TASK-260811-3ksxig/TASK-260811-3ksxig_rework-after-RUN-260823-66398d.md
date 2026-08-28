# Reviewer verdict for TASK-260811-3ksxig

Verdict: **changes requested -> to-dev**

## Review authority

- Reviewer run: `RUN-260823-66398d`
- `task-board spawn goal RUN-260823-66398d`: no active goal; the run is not goal-bound.
- Reviewed producer outcome: `TASK-260811-3ksxig_second-rework-report.md`.
- Reviewed implementation scope: `internal/pnpmsource`, the common portable-runner adjustment, tests, and pnpm profile documentation.
- No product or test code was modified by this reviewer. The reproducer used a Go overlay whose source and logs remain task-local under `.temp/TASK-260811-3ksxig/`.

## Rework confirmed

The three findings from `RUN-260823-7cb7d2` are materially addressed:

1. Workspace importer identities use `local:<path>` through parsing, selection, `buildNodeCapture`, and active-graph closure. The regression test asserts the workspace-only `b` dependency in both the common capture and active graph.
2. Root and workspace `node_modules` layouts now reject unclaimed packages/files, missing or wrong direct links, malformed manager metadata, unclaimed virtual-store entries, and wrong peer-context links.
3. The runtime gate admits only pnpm `10.33.0` before process start, and the real pinned-pnpm test exercised store derivation, offline frozen scripts-disabled materialization, and Node invocation successfully.

## Required change

### Lock-superset snapshot dependencies are rejected when the snapshot is pruned

`validateMaterializedTree` requires one virtual-store directory for every lock snapshot, including target-pruned and unreachable snapshots (`internal/pnpmsource/materialize.go`, lines 628-635 and 660-663). That matches the producer's observed pnpm 10.33.0 behavior. However, `validateSnapshotInstance` builds the allowed dependency-link set only from edges with `edge.Selected == true` (lines 809-823). It then rejects any additional link as unclaimed (lines 832-851).

Those rules conflict for a lock-superset snapshot. Real pnpm 10.33.0 materializes the declared dependency links of a target-pruned snapshot even though the snapshot and its edges are absent from the active selection.

The reviewer probe changed the existing target-pruned `optional@1.0.0` fixture so its admitted package and lock snapshot both declare `b@1.0.0`. The snapshot remained pruned for `os_mismatch`. The real pinned pnpm derived the private store and completed frozen offline installation, after which Curator rejected the valid layout:

```text
=== RUN   TestReviewerProbeRealPNPMPrunedSnapshotWithDependency
reviewer_pruned_snapshot_probe_test.go:47: pruned snapshot with dependency rejected:
closure_graph_incomplete: pnpm snapshot contains missing or unclaimed dependency links
--- FAIL: TestReviewerProbeRealPNPMPrunedSnapshotWithDependency
```

This violates target reconciliation/N10 and the stated lock-superset materialization support. Supported graphs can fail even though pnpm produced exactly the lock-declared layout from admitted bytes.

Required rework:

- Reconcile the physical dependency links for every materialized snapshot against that snapshot's exact declared lock edges, independently of active selection. Selection must still control the common active graph, importer direct links, returned active package set, and build/runtime reachability.
- Preserve fail-closed handling for unresolved/optional-missing edges and exact peer-context targets.
- Add a real pnpm 10.33.0 positive fixture for a target-pruned snapshot with a dependency, plus missing/swapped/unclaimed link negatives.
- Add coverage for an unreachable lock-superset snapshot with dependencies, or record and enforce a narrower unsupported boundary before materialization if pnpm's exact behavior differs for that shape.

## Validation evidence

| Command | Result |
| --- | --- |
| `PATH=<task-local pnpm 10.33.0> go test -count=1 -cover ./internal/pnpmsource` | pass; 80.1% statements |
| Targeted workspace/layout/version/real-pnpm tests | pass; real pnpm E2E ran, not skipped |
| `go vet ./internal/pnpmsource ./internal/closureexec ./internal/artifactpolicy ./internal/nodesource` | pass |
| `golangci-lint run ./internal/pnpmsource ./internal/closureexec ./internal/artifactpolicy ./internal/nodesource` | pass; 0 issues |
| `git diff --check` | pass |
| Real pinned-pnpm pruned-snapshot overlay probe | fail with `closure_graph_incomplete`, reproducing the defect |

Probe log: `TASK-260811-3ksxig_reviewer-pruned-snapshot-probe_RUN-260823-66398d.log`.

## Routing

This is ordinary implementation and conformance-test rework. There is no external or human-only blocker. Route to `to-dev`; another independent reviewer cycle is required after the fix.
