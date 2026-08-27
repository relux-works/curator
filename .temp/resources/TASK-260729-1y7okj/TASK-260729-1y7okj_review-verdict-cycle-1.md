# TASK-260729-1y7okj — review verdict, cycle 1

Date: 2026-07-29  
Role: reviewer  
Verdict: **changes requested → analysis**

## Summary

The independent audit correctly identifies the verifier-2 failure as cumulative
`cmd/curator` package runtime, preserves the read-only/no-measurement boundary,
accounts for the dominant compiler/plan fixtures, and supplies exact files,
functions, risks, and a literal narrow producer allowlist. R1, R2, and R4 are
supported by the current candidate source.

R3 is not assertion-preserving as written. One of the eleven cases assigned to
the frozen-plan replay changes protected-cache evidence and therefore requires a
plan acquired after its tamper. The resulting ranked plan and call-count
arithmetic must be corrected before the audit can be accepted.

No candidate or accepted-worktree file was modified by this review. No Go test,
build, vet, race, coverage, or Windows command was run. A separate producer
changed the candidate during this review; that provenance anomaly is recorded
below.

## Required correction

### R1 — frozen plan misclassifies the permission-boundary case

The R3 text says the eleven marker/context cases may share one frozen plan and
only the two cache-damage cases keep live plans. That partition is incorrect.

- `cmd/curator/status_test.go:523-539`,
  `TestStatusReportsCompiledCurrentnessAndFailsCheck`, changes every protected
  cache entry from mode `0700` to `0777` and expects
  `buildUntrustedCache` / `would-rebuild-untrusted-cache`.
- Planning reads cache evidence in
  `internal/install/plan.go:469-529` (`planOne` →
  `cache.Inspect`) and embeds the observed outcome/expectation in `buildFacts`.
- `cmd/curator/builds.go:228-261` (`recheckBuildCache`) compares that planned
  evidence with a new inspection.
- `cmd/curator/main.go:795-808` (`statusReport`) overwrites the classified state
  with `buildStateChanged` whenever the evidence moved.
- The candidate's own
  `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck`,
  `status_test.go:1183-1188,1228-1239`, explicitly proves that changing these
  same permissions after planning must produce `buildStateChanged`.
- The accepted global replay precedent states the governing rule at
  `global_status_test.go:338-343`: protected-cache state needs a plan acquired
  while that state is live.

Consequently, the proposed order “acquire frozen plan → chmod tamper →
`statusReport`” cannot preserve the expected `untrusted-build-cache` assertion.
It would fail as `build-state-changed`; changing the expected result would be a
semantic coverage regression.

Revise R3 so **three** cases acquire a live post-tamper plan:

1. protected cache entry cannot be interpreted;
2. protected cache holds no entry for the recorded key;
3. protected cache boundary is no longer provable.

The remaining eleven marker/context cases may use the frozen plan, with
`markerDigests` still taken after each tamper.

### R2 — reconcile the plan count and savings

The audit says R3 reduces 19 plan invocations to 8, but its written steps retain
the two separate clean-phase calls and count only two live cache cases. With the
required third live-cache plan, those steps produce 9 invocations:

- two clean-phase calls;
- one frozen acquisition;
- three post-tamper cache-plan acquisitions;
- three representative human CLI calls.

There are two valid corrections:

- Preferred: also combine the clean phase's `status --json` and `--check`
  calls (the duplicate already identified in section 4 item 5). The corrected
  total remains 8 and saves 11 plan units, approximately 36 seconds clean.
- Minimal: retain the split clean phase. The corrected total is 9 and saves
  10 plan units, approximately 33 seconds clean.

The artifact currently says the clean consolidation is “folded into item 6,”
but the ranked plan has only R1-R4. Name the consolidation in the actual ranked
patch steps and update the inventory, assertion matrix, and total consistently.

With the preferred correction, the original aggregate projection remains
approximately 86 seconds clean and 101-111 seconds under the verifier-2
`./...` load. With the minimal correction it is approximately 83 seconds clean
and 97-107 seconds under that load. Both still clear the task's 90-second
expected saving in the failing-load model; the producer must measure using the
existing narrow allowlist.

## Findings that remain supported

- Verifier 2 failed at `cmd/curator` 600.591 seconds while
  `TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands` had been active
  for only one second. Its stack shows ordinary transaction commit progress,
  not a compiled-command hang.
- At review start, candidate source digests matched the audit:
  `status_test.go` `4c2339dd...`, `global_status_test.go` `25aebe85...`, and
  `lifecycle_conformance_test.go` `5e832811...`.
- R1's cache snapshot replacement is confined to test fixture state and reuses
  the accepted `snapshotBuildCacheAfter` helper.
- R2 is valid when coupled to R1 because the marker and restored cache bytes/modes
  remain stable between sequential subtests.
- R4 combines document and exit-code assertions without reducing behavior
  coverage.
- The literal producer allowlist is narrow, has no timeout override or broad
  `./...` command, and covers every proposed change plus the unchanged
  lifecycle/repair regressions.
- No new board element is justified; the audit is a task-scoped input to the
  existing timing rework owner.

## Concurrent candidate mutation

The reviewer rechecked the three digests after writing this verdict. At
15:54:27 +0400, a separate in-flight producer changed candidate
`cmd/curator/status_test.go` from `4c2339dd...` to `bb21c15c...` while
`global_status_test.go` and `lifecycle_conformance_test.go` remained unchanged.
The new source begins a broader shared-fixture rework under
`TASK-260720-jrrgw9`; it was not written by this reviewer and was not the source
snapshot audited in the producer artifact.

This reinforces the `analysis` route: the revised audit must distinguish its
immutable reviewed snapshot from the now-moving primary rework, and must not
claim that `4c2339dd...` is still the current candidate. The R3 defect remains
a defect in the submitted patch plan independently of whether the primary
producer ultimately chooses that plan.

## Re-review gate

Return the research task through `analysis → to-review → reviewing` with a
revised task-scoped audit artifact that:

1. treats all three protected-cache mutations as live-plan cases;
2. names the clean-phase consolidation explicitly if it is used;
3. makes the 19→8 (or 19→9) accounting, assertion matrix, and savings agree;
4. preserves the current literal narrow producer allowlist and read-only audit
   boundary.
