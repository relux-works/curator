# TASK-260729-1y7okj — review verdict, cycle 2

Date: 2026-07-29
Role: reviewer
Verdict: **accepted → done**

## Summary

Revision 2 satisfies the cycle-1 correction and the task acceptance criteria. It is a concrete,
read-only optimization audit against an immutable reconstruction of the verifier-2 candidate, while
clearly separating that snapshot from the concurrently moving primary candidate. No Curator source,
test, spec, candidate, or accepted-worktree file was modified by this review.

## Cycle-1 correction verified

1. **All three protected-cache cases use live post-tamper plans.**
   Section 5 R3.3 keeps the corrupt-entry, missing-entry, and untrusted-permission cases as real
   `status app --json --check` invocations after each tamper. This agrees with
   `internal/install/plan.go:469-530` (`planOne` reads `cache.Inspect`),
   `cmd/curator/builds.go:201-262` (planned/live evidence comparison), and
   `cmd/curator/main.go:791-808` (moved evidence becomes `build-state-changed`). The permission case
   is no longer assigned to the frozen-plan group.
2. **Clean-phase consolidation is explicit.**
   R3.1 explicitly combines the clean `status --json` and `status --check` calls into one
   `status --json --check` call while retaining the decoded row, skill-map, identity, private-path,
   and zero-exit assertions.
3. **The invocation ledger reconciles to 19→8.**
   The reviewed snapshot has two clean calls, fourteen drift calls, and three extra human calls:
   19 plans. The preferred plan has one combined clean call, one frozen acquisition for eleven
   marker/context cases, three live post-tamper cache calls, and three human calls: 8 plans.
   Therefore R3 saves 11 plan units, approximately 36 seconds clean.
4. **Assertions and savings agree with that partition.**
   All fourteen drift state/cause/outcome/detail and skill-demotion assertions remain. Eleven cases
   invoke the same production classifier in-process; three cache cases retain real JSON/check
   process runs. Real drift exit mapping is truthfully reduced from fourteen to three representatives,
   plus the clean success process; per-case JSON assembly is reduced to four documents. The artifact
   names both reductions and maps the regression risks.

## Acceptance evidence

- Exact ranked locations are all in `cmd/curator/status_test.go`:
  `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck`,
  `TestStatusReportsCompiledCurrentnessAndFailsCheck`, and
  `TestStatusReportsATransitivelyResolvedCompiledCommand`.
- Duplicate work and projected savings are explicit: R1 ≈31 s, R2 ≈13 s, R3 ≈36 s, and R4
  ≈6.5 s clean. The document truthfully states that the clean projection is ≈86 s, while the
  verifier-2 failing-load projection is ≈101–112 s and therefore exceeds the required 90-second
  saving under the load where the unchanged ten-minute deadline failed. Measurement remains assigned
  to the timing producer.
- The literal producer allowlist remains narrow: no timeout override, no broad `./...`, no race,
  coverage, Windows, cache-clearing, installation, staging, commit, or publication command. The
  renamed shared-fixture selectors have a separate re-pointing rule and an executed-test-count guard.
- The assertion-preservation matrix covers compiled drift classification, skill demotion, operator
  detail and text, check behavior, cache-move detection, repair/rollback, transitive reporting,
  lifecycle vectors, and global status.
- The plan fits the project architecture: it is test-only, reuses the accepted global-status
  acquire/replay and cache-snapshot patterns, and calls production classification functions rather
  than introducing test-only product seams.
- Board decomposition is intentionally unchanged because implementation/timing ownership already
  exists under `TASK-260720-jrrgw9`; creating another element would duplicate active scope.

## Provenance and boundary

- Immutable reviewed snapshot:
  `.temp/TASK-260729-1y7okj/reviewed-snapshot/status_test.go` =
  `4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe`.
- Current moving primary candidate observed during this review:
  `.temp/TASK-260720-jrrgw9/worktree/cmd/curator/status_test.go` =
  `487b12bdf531e4714983eab83b804de7b4604513e435256e550f60391ee0d32e`.
- Stable supporting files still match the provenance record, including
  `global_status_test.go` = `25aebe85…`, `builds.go` = `6aab0a0f…`,
  `main.go` = `a0798e7d…`, and `internal/install/plan.go` = `b9684fde…`.
- Verifier-2 evidence independently confirms `cmd/curator` failed at 600.591 s, with a legacy
  status test active for only one second and ordinary transaction `Lstat` work on its stack;
  `internal/install` and `internal/install/atomicity` completed at 411.094 s and 543.719 s.
- This review executed no Go test, build, vet, race, coverage, or Windows command. Tests-green is
  not applicable to this research deliverable because both attached audit instructions prohibit
  those commands and assign measurement to the timing agent. The last known full verifier remains
  red; this verdict accepts the audit artifact, not the unmeasured producer patch or parent delivery.

