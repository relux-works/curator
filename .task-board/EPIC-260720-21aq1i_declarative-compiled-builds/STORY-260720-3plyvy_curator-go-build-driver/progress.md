## Status
reviewing

## Assigned To
[analyst] solution-architect (codex)

## Created
2026-07-19T22:10:05Z

## Last Update
2026-07-30T10:00:29Z

## Blocked By
- (none)

## Blocks
- STORY-260720-1uv5gi

## Checklist
- [x] Decompose the accepted protocol contract into atomic Curator Go implementation, lifecycle, cache, platform, documentation, and test tasks with explicit dependencies; planning only, then leave the story at to-dev
- [x] Tasks created with description and AC
- [x] Dependencies linked
- [x] Tasks are atomic — one clear deliverable each
- [x] Completeness verified — nothing forgotten
- [x] Gaps closed with blocking tasks
- [x] Diagrams or planning artifacts linked as new task-scoped outcome resources
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Audit the existing 18-task Curator decomposition for atomic package ownership, complete accepted-contract coverage, dependency correctness, and executable Go/platform gates; do not duplicate tasks, correct only evidenced defects

## Notes
spawn queued: [analyst] solution-architect (codex) (run=RUN-260720-04218f, max_parallel=20)
spawn run started: [analyst] solution-architect (codex) (run=RUN-260720-04218f)
Logbook 2026-07-20 — solution architecture: decomposed the accepted rc.4 contract at SHA-256 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681 into 18 atomic Curator tasks and 42 dependencies across 10 phases. The three foundation tasks are explicitly blocked by upstream integrated verification TASK-260720-3ag6pi. Decisions: trusted Go resolution is CURATOR_GO, then GOROOT/bin/go, then runtime.GOROOT()/bin/go, never PATH; physical cache is manager-local cache/build/go-v1/<key> while protocol identity remains logical; build-source identity remains separate from legacy installed-content hashing; real installs use project lock, private builds, manager-home revalidation, journaled target swaps, and consumer-last commit; dry-run takes no locks and performs no build or persistent mutation. Findings: shared checkout c06aa1a differs from required origin/main 17804cea, so baseline analysis and passing go test, race, and vet checks used a detached worktree without touching user work; existing install writes consumer state before all materialization, runtime reuse lacks required-path validation, lifecycle writes are piecemeal without lock or journal, and CI still pins rc.3 without an explicit race gate. No unresolved human decision or additional research task is needed. Board validation reports only 12 legacy EPIC-260712 broken references and one unrelated orphan resource. Component and lifecycle diagrams plus the completeness audit are attached. Role handoff target is to-review; this overrides the inherited checklist phrase saying to-dev.
Logbook 2026-07-20 — handoff state-model anomaly: task-board rejected to-review while the coarse story-level blocked_by STORY-260720-35dck7 remained at to-review. That link was removed as redundant after verifying the exact implementation gates on TASK-260720-2g0e3b, TASK-260720-256kj1, and TASK-260720-1zl1cj all remain blocked_by TASK-260720-3ag6pi. This preserves the real protocol dependency and prevents implementation from starting, while allowing the solution-architecture decomposition itself to enter review. Downstream story links remain unchanged.
agent completed: [analyst] solution-architect (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-04218f, pid=66993, exit=0)
spawn queued: [analyst] solution-architect (codex) (run=RUN-260720-010913, max_parallel=20)
spawn run started: [analyst] solution-architect (codex) (run=RUN-260720-010913)
Logbook 2026-07-20 — second solution-architecture audit: independently traced all 18 Curator tasks against accepted contract SHA-256 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681 and origin/main 17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8. Corrected six evidenced planning defects in place without adding or duplicating tasks: fixed exact ownership at internal/managerlock, internal/transaction, internal/protocoljson, internal/buildmeta, and internal/buildcache; made runtime/shim work staged-only before the atomic commit task; made install planning return immutable staged results; added read-only source/vendor, operation-private writes, network/module/VCS denial, unexpected-child rejection, and supported deadline/output/disk/memory/process gates to go-v1; and reconciled candidate CI with the two-stage release-evidence model. Decision: TASK-260720-1pvfj5 must keep the committed suite at the current qualified release and produce explicit candidate evidence; TASK-260720-38l1sy, already blocked by both TASK-260720-1pvfj5 and protocol qualification TASK-260720-25d05o, owns the official pin. Adding TASK-260720-25d05o as a prerequisite of candidate CI would recreate the manager/spec release cycle, so the existing 42-edge graph remains unchanged and acyclic across 10 phases. Added and visually validated task-scoped execution-boundary and release-gate diagrams plus a new decomposition audit and canonical plan outcome. No human clarification or new research task is needed. task-board validate still reports only the same 12 unrelated legacy EPIC-260712 prose links and one unrelated orphan resource; no product code was changed.
Logbook 2026-07-20 — concurrent dependency reconciliation: while final validation was running, the parallel interoperability audit linked manager-release qualification TASK-260720-3pvihp blocked_by candidate Curator CI TASK-260720-1pvfj5 and added explicit checklist/notes preserving the previous committed suite pin. This is the missing pre-release evidence edge and is consistent with this audit. The Curator story-local graph remains 42 prerequisites and 10 phases; the official pin audit TASK-260720-38l1sy remains the post-release join blocked by both TASK-260720-1pvfj5 and TASK-260720-25d05o. The audit outcome was updated rather than overwriting concurrent work.
agent completed: [analyst] solution-architect (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-010913, pid=37487, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [STORY-260720-3plyvy_spawn-log_-analyst--solution-architect--codex-.log](file://STORY-260720-3plyvy/STORY-260720-3plyvy_spawn-log_-analyst--solution-architect--codex-.log) — System spawn log captured by task-board
- [STORY-260720-3plyvy_decomposition-plan.md](file://STORY-260720-3plyvy/STORY-260720-3plyvy_decomposition-plan.md) — Development-ready task map, completeness audit, architecture decisions, and baseline findings
- [STORY-260720-3plyvy_canonical-board-plan.md](file://STORY-260720-3plyvy/STORY-260720-3plyvy_canonical-board-plan.md) — Canonical task-board dependency plan with phases and critical path
- [STORY-260720-3plyvy_origin-main-baseline.md](file://STORY-260720-3plyvy/STORY-260720-3plyvy_origin-main-baseline.md) — Detached origin/main Go test, race, and vet baseline evidence
- [STORY-260720-3plyvy_component-map.puml](file://STORY-260720-3plyvy/STORY-260720-3plyvy_component-map.puml) — PlantUML source for implementation ownership and package seams
- [STORY-260720-3plyvy_component-map.svg](file://STORY-260720-3plyvy/STORY-260720-3plyvy_component-map.svg) — Rendered implementation ownership diagram
- [STORY-260720-3plyvy_install-lifecycle.puml](file://STORY-260720-3plyvy/STORY-260720-3plyvy_install-lifecycle.puml) — PlantUML source for safe dry-run, build, commit, rollback, and GC ordering
- [STORY-260720-3plyvy_install-lifecycle.svg](file://STORY-260720-3plyvy/STORY-260720-3plyvy_install-lifecycle.svg) — Rendered safe install lifecycle diagram
- [STORY-260720-3plyvy_planning-validation.md](file://STORY-260720-3plyvy/STORY-260720-3plyvy_planning-validation.md) — Story-local completeness audit and repository-wide validation findings
- [STORY-260720-3plyvy_decomposition-audit-2.md](file://STORY-260720-3plyvy/STORY-260720-3plyvy_decomposition-audit-2.md) — Second-pass atomic ownership, accepted-contract coverage, dependency, platform-gate, and release-order audit
- [STORY-260720-3plyvy_canonical-board-plan-audit-2.md](file://STORY-260720-3plyvy/STORY-260720-3plyvy_canonical-board-plan-audit-2.md) — Canonical post-audit task-board phase plan and critical path
