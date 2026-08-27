## Status
done

## Assigned To
[reviewer] reviewer (codex)

## Created
2026-07-20T01:44:31Z

## Last Update
2026-07-20T19:00:55Z

## Blocked By
- TASK-260720-1nvomm

## Blocks
- TASK-260720-12iigs

## Checklist
- [x] Specify the five fixed Go argv forms, clean environment, and allowed process graph exactly
- [x] Specify audit-before-build, compiler-free dry-run, locked commit, rollback, recovery, repair, and GC ordering
- [x] Run the scoped profile validation command and record the result
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn queued: [implementer] developer (codex) (run=RUN-260720-355fb0, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260720-355fb0)
Implementation logbook 2026-07-20: The spawned cwd was curator, while owned work belongs to sibling curator-spec. Verified curator-spec HEAD and origin/main both equal required 57c1f56846d221ecc55786bd3c2467ec32f11730. Treated accepted prerequisite edits in protocol/core.md, SECURITY.md, and decisions/0004 as shared read-only inputs and changed only profiles/manager.md. The lifecycle resolves the recovery-before-private-build nuance by permitting only mandatory recovery of an earlier journal before current-operation builds, while requiring every current-operation miss to build and verify before mutation. Scoped and combined validation gates pass; evidence is attached in TASK-260720-17llva_results.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-355fb0, pid=46738, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260720-5b678e, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260720-5b678e)
Review verdict 2026-07-20: changes requested. `profiles/manager.md` lines 263-270 performs journal recovery and shared mutation before private builds, contrary to the attached normative activity model and the AC requiring build misses before mutation and byte-for-byte preservation on build failure. Required rework: move recovery to the single manager-home locked publication phase after all private builds succeed; revalidate and restart from the earliest affected read/build when recovery changes assumptions. Full evidence and green validation results are attached in TASK-260720-17llva_review-verdict.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-5b678e, pid=51979, exit=0)
spawn queued: [implementer] developer (codex) (run=RUN-260720-383dfd, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260720-383dfd)
Rework logbook 2026-07-20: Removed the pre-build recovery pass identified by review. Install and repair now stage and verify every miss before any shared-state recovery or mutation; recovery occurs once under the post-build manager-home publication lock, and recovery-induced drift restarts from the earliest affected read or build. Failed private builds preserve installation, consumers, and live caches byte-for-byte from operation entry. Scoped validation passed 30 schemas and 93 vectors; make validate passed 8 Python tests and go test ./tools/...; git diff --check passed. Evidence attached as TASK-260720-17llva_rework-results.md and TASK-260720-17llva_rework-validation.log.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-383dfd, pid=53939, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260720-207e08, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260720-207e08)
Review verdict cycle 2 on 2026-07-20: changes requested. The recovery-ordering rework is correct and all scoped/full validations pass. Remaining rework: profiles/manager.md lines 574-576 must mark valid marker-v1 references as well as marker-v2 references during locked GC, otherwise a still-current schema 1 through 5 installation can lose runtime or snapshot entries after the grace period; line 463 must update the adapter-ledger cross-reference from Protocol Core section 10 to section 11. Full evidence is attached in TASK-260720-17llva_review-verdict-cycle-2.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-207e08, pid=57644, exit=0)
spawn queued: [implementer] developer (codex) (run=RUN-260720-4cb61c, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260720-4cb61c)
Rework logbook cycle 3 on 2026-07-20: Updated the adapter-ledger cross-reference to Protocol Core section 11 and made locked GC preserve runtime and snapshot entries referenced by every supported valid marker schema, including marker v1, while only marker v2 can supply compiled-cache references. Scoped validation passed 30 schemas and 93 vectors; make validate passed 8 Python tests and go test ./tools/...; diff check passed. Evidence attached as TASK-260720-17llva_rework-cycle-3-results.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-4cb61c, pid=61158, exit=0)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260720-ccb86e, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260720-ccb86e)
Review cycle 3 verdict 2026-07-20: accepted. The marker-v1 GC compatibility and Protocol Core section 11 cross-reference findings are resolved. Full AC and architecture audit found no remaining issues. Independent scoped validation passed 30 schemas and 93 vectors; make validate passed 8 Python tests and go test ./tools/...; owned-file diff check passed. Verdict and logs are attached as TASK-260720-17llva_review-verdict-cycle-3.md, TASK-260720-17llva_review-cycle-3-scoped-validation.log, and TASK-260720-17llva_review-cycle-3-full-validation.log.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-ccb86e, pid=63090, exit=0)

## Precondition Resources
- [TASK-260720-17llva_install-lifecycle.puml](file://TASK-260720-17llva/TASK-260720-17llva_install-lifecycle.puml) — Normative install and dry-run activity model for manager-profile implementation

## Outcome Resources
- [TASK-260720-17llva_install-lifecycle.svg](file://TASK-260720-17llva/TASK-260720-17llva_install-lifecycle.svg) — Rendered lifecycle diagram linked to the manager-profile task
- [TASK-260720-17llva_spawn-log_-implementer--developer--codex-.log](file://TASK-260720-17llva/TASK-260720-17llva_spawn-log_-implementer--developer--codex-.log) — System spawn log captured by task-board
- [TASK-260720-17llva_results.md](file://TASK-260720-17llva/TASK-260720-17llva_results.md) — Manager lifecycle implementation coverage and passing validation evidence
- [TASK-260720-17llva_spawn-log_-reviewer--reviewer--codex-.log](file://TASK-260720-17llva/TASK-260720-17llva_spawn-log_-reviewer--reviewer--codex-.log) — System spawn log captured by task-board
- [TASK-260720-17llva_review-verdict.md](file://TASK-260720-17llva/TASK-260720-17llva_review-verdict.md) — Reviewer verdict, lifecycle-ordering finding, required rework, and validation evidence
- [TASK-260720-17llva_rework-results.md](file://TASK-260720-17llva/TASK-260720-17llva_rework-results.md) — Post-review lifecycle rework summary and green validation evidence
- [TASK-260720-17llva_rework-validation.log](file://TASK-260720-17llva/TASK-260720-17llva_rework-validation.log) — Scoped and full validation log for lifecycle-ordering rework
- [TASK-260720-17llva_review-verdict-cycle-2.md](file://TASK-260720-17llva/TASK-260720-17llva_review-verdict-cycle-2.md) — Second reviewer-cycle verdict with GC compatibility and cross-reference findings
- [TASK-260720-17llva_rework-cycle-3-results.md](file://TASK-260720-17llva/TASK-260720-17llva_rework-cycle-3-results.md) — Third rework-cycle changes and green validation evidence
- [TASK-260720-17llva_review-verdict-cycle-3.md](file://TASK-260720-17llva/TASK-260720-17llva_review-verdict-cycle-3.md) — Third-cycle accepted reviewer verdict with AC, architecture, and validation evidence
- [TASK-260720-17llva_review-cycle-3-scoped-validation.log](file://TASK-260720-17llva/TASK-260720-17llva_review-cycle-3-scoped-validation.log) — Independent scoped profile validation for accepted review cycle 3
- [TASK-260720-17llva_review-cycle-3-full-validation.log](file://TASK-260720-17llva/TASK-260720-17llva_review-cycle-3-full-validation.log) — Independent full repository validation for accepted review cycle 3
