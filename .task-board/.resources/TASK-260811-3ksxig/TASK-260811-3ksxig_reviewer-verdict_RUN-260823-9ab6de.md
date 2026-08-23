# Reviewer verdict for TASK-260811-3ksxig

Verdict: **changes requested -> to-dev**

## Review authority

- Reviewer run: `RUN-260823-9ab6de`
- `task-board spawn goal RUN-260823-9ab6de`: no active goal; the run is not goal-bound.
- Reviewed producer outcome: `TASK-260811-3ksxig_third-rework-report.md`.
- Reviewed implementation scope: `internal/pnpmsource` and pnpm profile documentation.
- No product or test code was modified by this reviewer. The reproducer uses a Go overlay under `.temp/TASK-260811-3ksxig/`.

## Rework confirmed

The prior target-pruned snapshot-link defect is fixed for the covered reachable
shape. Physical dependency links are reconciled from the full resolved lock
snapshot rather than `edge.Selected`; target-pruned packages remain absent from
the active package result. Missing, swapped, and unclaimed links reject. A real
pinned pnpm 10.33.0 run successfully derived the private store and completed
frozen, offline, scripts-disabled materialization for that fixture.

The ordinary wholly unreachable fixture also rejects before install, and the
focused package/shared tests, vet, lint, diff check, and board validation pass.

## Required change

### Target-incompatible wholly unreachable snapshots bypass the pre-install boundary

`markSelection` stores only one `Snapshot.PruneReason`. A snapshot rejected by
its `os`, `cpu`, or `libc` selector retains `os_mismatch`, `cpu_mismatch`, or
`libc_mismatch` even when no importer or reachable snapshot references it.
`Materialize` detects the unsupported wholly unreachable lock shape only by
testing `snapshot.PruneReason == "unreachable"` (`internal/pnpmsource/materialize.go`,
lines 279-286). The target reason therefore masks graph unreachability.

The reviewer overlay created an unreferenced `dormant@1.0.0` snapshot with a
dependency and an incompatible `os` selector. Parsing produced
`Selected=false, PruneReason="os_mismatch"`. Materialization did not apply the
documented pre-install unsupported boundary: the fake pinned-contract runner
started install and accepted the snapshot.

```text
=== RUN   TestReviewerProbeUnreachableTargetPrunedRejectsBeforeInstall
reviewer_unreachable_target_pruned_probe_test.go:51:
    unreachable target-pruned snapshot was accepted; starts=2
--- FAIL: TestReviewerProbeUnreachableTargetPrunedRejectsBeforeInstall
```

This contradicts the profile's stated rule that wholly unreachable snapshots
are rejected before install and violates the negative-vector requirement that
no affected manager process starts after a graph/profile failure. The same gap
applies to `cpu_mismatch` and `libc_mismatch`.

Required rework:

1. Track graph reachability independently from target selection/prune reason,
   or derive it explicitly from importer/snapshot edges before materialization.
2. Reject every wholly unreachable lock snapshot before the install process,
   including snapshots that are also OS/CPU/libc-pruned.
3. Preserve supported target-pruned-but-lock-reachable physical reconciliation
   and active-graph separation.
4. Add OS/CPU/libc-overlap regression coverage with a zero-install-start
   assertion; exercise at least one overlap through the real pinned pnpm path
   if manager behavior is part of the boundary evidence.

## Validation evidence

| Command | Result |
| --- | --- |
| Pinned pnpm targeted lock-superset, link-negative, unreachable, and real E2E tests | pass |
| `PATH=<task-local-pnpm-10.33.0> go test -count=1 -cover ./internal/pnpmsource` | pass; 80.3% statements |
| Shared scoped Go tests | pass |
| `go vet` and `golangci-lint` for pnpm/shared packages | pass |
| `git diff --check` | pass |
| `task-board validate` | pass |
| Reviewer target-pruned + unreachable overlay probe | fail; snapshot accepted and install started (`starts=2`) |

Probe log: `TASK-260811-3ksxig_reviewer-unreachable-target-pruned-probe_RUN-260823-9ab6de.log`.

## Routing

This is ordinary implementation and conformance-test rework. There is no
external or human-only blocker. Route to `to-dev`; another independent reviewer
cycle is required after the fix.
