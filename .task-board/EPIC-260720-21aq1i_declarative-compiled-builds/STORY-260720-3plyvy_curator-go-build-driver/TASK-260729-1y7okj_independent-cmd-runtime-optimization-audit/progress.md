## Status
done

## Review
required

## Task Class
code

## Estimate
notEstimated

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Verifier2 failure and current candidate source are analyzed without edits
- [x] Ranked optimization plan names exact files and functions with expected savings
- [x] Assertion-preservation and regression risks are mapped
- [x] Literal narrow producer command allowlist is provided in a task-scoped outcome
- [x] Board size is proportional to the spec and is the smallest decomposition that maps every requirement
- [x] Every story and task traces to a concrete spec requirement; justified-gap elements also carry a self-verified gap record
- [x] Beyond-literal-spec elements include a written justification naming the gap and the spec and out-of-scope checks performed before creation
- [x] Research tasks cite an exact question the spec genuinely leaves open
- [x] Dependencies linked
- [x] Tasks are atomic — one clear deliverable each
- [x] Completeness verified — nothing forgotten
- [x] Any planning artifacts actually produced are linked as new task-scoped outcome resources; diagrams are strictly optional, never a standing deliverable
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [ ] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [x] Verification delegated, not run: this read-only audit executed no Go test/build/vet/race/coverage/Windows command, per TASK-260729-1y7okj_audit-scope.md and TASK-260729-1y7okj_rework-cycle-1.md; the literal narrow producer allowlist in the audit artifact is the verification contract for the timing agent

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [analyst] solution-architect (claude) (run=RUN-260729-a235a6, max_parallel=20)
spawn run started: [analyst] solution-architect (claude) (run=RUN-260729-a235a6)
Independent read-only audit complete; no worktree file modified (status_test.go digest still 4c2339dd..., global_status_test.go still 25aebe85...). No Go test/build/vet/race/coverage/Windows command was run by this audit.

Cost model: all cmd/curator wall time is (a) planBuilds -> godriver.establish, which runs 3 go subprocesses plus TWO full 233MB/14502-file GOROOT SHA-256 fingerprints per compiled-scope invocation (Probe does not memoize, Close re-fingerprints), and (b) a per-operation-private empty GOCACHE, so every real compiled install compiles cold. Derived units: plan-only P ~3.2s, real rebuild B ~8-11s. Static inventory: ~24 real builds + ~55 plan-only invocations => predicted ~463s vs accepted clean baseline 465.944s (model reproduces reality within 1%).

Ranked test-only plan, all in cmd/curator/status_test.go:
R1 status_test.go:1216 replace t.Cleanup(reinstall) with snapshotBuildCacheAfter (all 4 moves touch only home/cache/build) -> ~31s
R2 status_test.go:1203-1208 hoist plan/facts/markerDigests out of the loop (valid only with R1) -> ~13s
R3 status_test.go:337-621 adopt the accepted acquire-once/replay pattern from global_status_test.go:386; planOne never takes the install marker, so the 11 marker/context cases may share one frozen plan; the 2 cache cases keep live plans -> ~36s
R4 status_test.go:1019/1031/1042/1050 combine --json and --check -> ~6.5s
Total ~86s clean, ~101-111s at ./... load (observed high-load inflation x1.17-1.29). Projects cmd/curator 600.6s -> ~490-500s, i.e. ~100s margin, above the 90s requirement. Reserve R5/R6 documented but NOT recommended (R6 loses real coverage).

Top regression trap: in a replayed case markerDigests must be taken AFTER the tamper, otherwise every marker case classifies as build-state-changed instead of its own code.

Second latent deadline flagged for the timing agent: internal/install/atomicity measured 543.719s and internal/install 411.094s against the same unchanged 600s deadline; atomicity sits 56s from its own timeout. No change proposed for those packages here.

No new board element created: rework is already owned by TASK-260720-jrrgw9 and the product-side fingerprint work closed under TASK-260729-1zex8r.

Outcome: TASK-260729-1y7okj_runtime-optimization-audit.md
agent completed: [analyst] solution-architect (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-a235a6, pid=10366, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-567284, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-567284)
REVIEW CYCLE 1 — CHANGES REQUESTED → analysis. R1/R2/R4, timeout diagnosis, narrow allowlist, and initial digests are supported. R3 is not assertion-preserving as written: the protected-cache permission case at audited status_test.go:523-539 changes cache evidence, so a pre-tamper frozen plan is rechecked as build-state-changed instead of the required untrusted-build-cache. Treat all three protected-cache cases as live post-tamper plans and reconcile the 19→8 count by explicitly combining the clean JSON/check call, or report 19→9 and reduced savings. During review a separate TASK-260720-jrrgw9 producer changed status_test.go from 4c2339dd… to bb21c15c…; reviewer changed neither worktree. Full evidence: TASK-260729-1y7okj_review-verdict-cycle-1.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-567284, pid=23430, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [analyst] solution-architect (claude) (run=RUN-260729-545602, max_parallel=20)
spawn run started: [analyst] solution-architect (claude) (run=RUN-260729-545602)
Cycle-1 rework complete, ready for review.

Reviewer corrections applied and independently verified against production source:
- R1 (verdict): three protected-cache cases, not two, must classify from a live post-tamper plan. Confirmed via planOne->cache.Inspect (internal/install/plan.go:504-509), plannedEvidence/observedEvidence/recheckBuildCache (cmd/curator/builds.go:201-262), statusReport overwrite to build-state-changed (cmd/curator/main.go:795-808), and the 0o022 boundary test (internal/buildcache/protection_unix.go:239,259). The chmod 0777 case (status_test.go:523) moves inspected evidence, so a frozen cache-hit plan would publish build-state-changed instead of untrusted-build-cache.
- R2 (verdict): clean-phase consolidation is now an explicitly ranked step R3.1; audit carries a line-by-line plan-invocation ledger landing on 19->8 preferred (minimal 19->9 stated and rejected with reason). Savings recomputed from P = 3.26 s: R1 31 s, R2 13 s, R3 36 s, R4 6.5 s = 86 s clean, 101-112 s under the ./... load that produced the 600.591 s failure. Clean-load 86 s is below the 90 s bar and is stated as such; the failing-load figure clears it.

Design refinement over the verdict: for the three cache cases the plan IS the cost, so an in-process live plan + replay and a real CLI status --json --check both cost 1 P. The real CLI is recommended - identical price, and it additionally preserves the published document and the fail-closed process exit for those three. Exit-mapping representatives therefore go 14 -> 3 drift cases + 1 clean, better than rev. 1 claimed.

Provenance: reviewed snapshot 4c2339dd reconstructed byte-exactly into .temp/TASK-260729-1y7okj/reviewed-snapshot/ from the recorded pre-patch base c62d0791 plus board diff 03c637da. The live candidate moved twice since (4c2339dd -> bb21c15c -> 487b12bd) under TASK-260720-jrrgw9; it is not this audit baseline. That producer has already implemented R1 and deliberately retains the per-case plan R2 targets; R3 and R4 apply verbatim to its byte-identical lines 344-620. Incremental there: R3+R4 ~42.5 s clean, or R2+R3+R4 ~55.5 s if the owner accepts R2.

Not checked, truthfully: checklist 16 (Tests green). This task is forbidden by its own scope and rework instruction from running Go tests, builds, vet, race or coverage; measurement belongs to the timing agent via the narrow allowlist in the artifact. No source file was modified in either worktree and no board element of TASK-260720-jrrgw9 was touched.
HANDOFF GATE — checklist item 16 (Tests green) left UNCHECKED deliberately.

task-board handoff TASK-260729-1y7okj --role solution-architect fails closed on item 16. It is not checked because it cannot be checked truthfully: this task is a read-only audit whose own attached instructions forbid running Go tests, builds, vet, race, coverage or Windows commands (TASK-260729-1y7okj_audit-scope.md: "Do not run Go tests, builds, vet, race, coverage, or Windows commands; another timing agent owns measurements"; TASK-260729-1y7okj_rework-cycle-1.md: "Do not edit Curator source/tests/specs, run Go commands"). No test was observed green, and the last known full run of the parent candidate was RED (cmd/curator FAIL 600.591s under verifier 2), so asserting Tests green would be actively misleading.

Item 18 was added and checked instead, recording the delegation truthfully. The role end status to-review has already been applied via set_status. What is needed is a coordinator decision to waive item 16 for this read-only analysis task (or to record the timing agent measurement against it once the narrow allowlist is run). No other work remains.
agent completed: [analyst] solution-architect (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-545602, pid=27276, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-eaea05, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-eaea05)
REVIEW CYCLE 2 — ACCEPTED. Revision 2 corrects the protected-cache partition: corrupt, missing, and untrusted-permission cases each retain a real post-tamper CLI plan; eleven marker/context cases use the frozen plan with markerDigests taken after tamper. R3.1 explicitly consolidates clean JSON/check; ledger is 19→8 and the assertion matrix/savings agree. Immutable reviewed snapshot 4c2339dd… is distinguished from live primary candidate 487b12bd…. Literal narrow producer allowlist and read-only/no-Go-command boundary remain intact. Tests-green is N/A for this research artifact by attached instruction; last full parent verifier remains red. Evidence: TASK-260729-1y7okj_review-verdict-cycle-2.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-eaea05, pid=42910, exit=0)

## Precondition Resources
- [TASK-260729-1y7okj_audit-scope.md](file://TASK-260729-1y7okj/TASK-260729-1y7okj_audit-scope.md) — Exact candidate and verifier2 evidence for independent runtime audit
- [TASK-260729-1y7okj_rework-cycle-1.md](file://TASK-260729-1y7okj/TASK-260729-1y7okj_rework-cycle-1.md) — Reviewer-requested correction to cache-plan partition and savings accounting
- [TASK-260729-1y7okj_review-instructions-cycle-2.md](file://TASK-260729-1y7okj/TASK-260729-1y7okj_review-instructions-cycle-2.md) — Independent cycle-2 review of corrected runtime optimization audit

## Outcome Resources
- [TASK-260729-1y7okj_spawn-log_-analyst--solution-architect--claude-_RUN-260729-a235a6.log](file://TASK-260729-1y7okj/TASK-260729-1y7okj_spawn-log_-analyst--solution-architect--claude-_RUN-260729-a235a6.log) — System spawn log captured by task-board
- [TASK-260729-1y7okj_runtime-optimization-audit.md](file://TASK-260729-1y7okj/TASK-260729-1y7okj_runtime-optimization-audit.md) — Rev. 2 independent cmd/curator runtime audit: corrected 3-case live-plan partition, named clean-phase consolidation, 19->8 plan ledger, provenance of the immutable reviewed snapshot vs the moving candidate
- [TASK-260729-1y7okj_spawn-log_-reviewer--reviewer--codex-_RUN-260729-567284.log](file://TASK-260729-1y7okj/TASK-260729-1y7okj_spawn-log_-reviewer--reviewer--codex-_RUN-260729-567284.log) — System spawn log captured by task-board
- [TASK-260729-1y7okj_review-verdict-cycle-1.md](file://TASK-260729-1y7okj/TASK-260729-1y7okj_review-verdict-cycle-1.md) — Reviewer changes-requested evidence: frozen-plan cache-permission misclassification, corrected plan accounting, concurrent candidate provenance anomaly
- [TASK-260729-1y7okj_spawn-log_-analyst--solution-architect--claude-_RUN-260729-545602.log](file://TASK-260729-1y7okj/TASK-260729-1y7okj_spawn-log_-analyst--solution-architect--claude-_RUN-260729-545602.log) — System spawn log captured by task-board
- [TASK-260729-1y7okj_reviewed-snapshot-provenance.md](file://TASK-260729-1y7okj/TASK-260729-1y7okj_reviewed-snapshot-provenance.md) — Immutable reviewed-snapshot provenance: byte-exact reconstruction of 4c2339dd, unchanged product digests, and the moving TASK-260720-jrrgw9 candidate distinguished
- [TASK-260729-1y7okj_spawn-log_-reviewer--reviewer--codex-_RUN-260729-eaea05.log](file://TASK-260729-1y7okj/TASK-260729-1y7okj_spawn-log_-reviewer--reviewer--codex-_RUN-260729-eaea05.log) — System spawn log captured by task-board
- [TASK-260729-1y7okj_review-verdict-cycle-2.md](file://TASK-260729-1y7okj/TASK-260729-1y7okj_review-verdict-cycle-2.md) — Cycle-2 accepted verdict: corrected live-plan partition, 19-to-8 ledger, assertion/savings/provenance/boundary evidence

## Created
2026-07-29T11:34:08Z

## Last Update
2026-07-29T12:16:14Z

## Assigned To
[reviewer] reviewer (codex)
