## Status
backlog

## Assigned To
[analyst] solution-architect (codex)

## Created
2026-07-19T22:10:05Z

## Last Update
2026-08-27T03:32:10Z

## Blocked By
- STORY-260720-3plyvy

## Blocks
- STORY-260720-21bsr2

## Checklist
- [x] Decompose the accepted protocol contract into atomic cocoaskills Python implementation, lifecycle, cache, platform, documentation, and test tasks with explicit dependencies; planning only, then leave the story at to-dev
- [x] Tasks created with description and AC
- [x] Dependencies linked
- [x] Tasks are atomic — one clear deliverable each
- [x] Completeness verified — nothing forgotten
- [x] Gaps closed with blocking tasks
- [x] Diagrams or planning artifacts linked as new task-scoped outcome resources
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Audit the existing 17-task csk decomposition for atomic module ownership, complete accepted-contract coverage, dependency correctness, and executable Python/platform gates; do not duplicate tasks, correct only evidenced defects

## Notes
spawn queued: [analyst] solution-architect (codex) (run=RUN-260720-3b3d1f, max_parallel=20)
spawn run started: [analyst] solution-architect (codex) (run=RUN-260720-3b3d1f)
Logbook 2026-07-20 — verified the accepted compile-only driver contract at SHA-256 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681 and cocoaskills origin/main 6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12, also confirmed remotely. The local cocoaskills main clone is clean but two commits behind, so every implementation task requires a clean fast-forward before its task worktree and inclusion of dependency handoffs. Decision: 17 atomic tasks in 12 phases, with parallel manifest/transaction foundations, separate source/toolchain identities, csk-specific top-level manager-home builds storage, explicit POSIX and Windows protected-cache backends, audit-first pure planning, serialized project then global integration, maintenance, shared vectors, cross-platform E2E, docs, and integrated verification. Existing implementation anomalies are assigned explicitly: consumer-before-materialization, unverified runtime reuse, non-transactional target replacement, dry-run lock creation, and registry cache/state mutation. No separate lint command exists at the inspected baseline; strict mypy is the declared typing gate and final verification will run any lint gate present then. No new clarification or research task is needed because the accepted contract closes product and architecture choices; existing Curator and protocol dependencies remain blocking. PlantUML source and SVG diagrams validated with Smetana. task-board validate reports only 12 unrelated legacy EPIC-260712 broken prose references and one unrelated orphan resource. Lifecycle note: inherited checklist item 1 says to-dev, but this spawned solution-architect assignment explicitly mandates the final board command to-review; the decomposition will be handed to review and will not mark implementation accepted.
Board lifecycle anomaly 2026-07-20: creating backlog child tasks auto-derived the parent story back to backlog after the required analysis start status. This did not affect decomposition data; the explicit role handoff command remains authoritative.
Logbook 2026-07-20 — handoff state-model audit: task-board rejected to-review while coarse story-level blockers STORY-260720-3plyvy and STORY-260720-35dck7 were active. These links are removed at planning handoff because the Curator link is non-technical and conflicts with the story requirement for an independent Python implementation, while the protocol-story link duplicates exact executable gates. TASK-260720-12r55p remains blocked by protocol verification TASK-260720-3ag6pi, TASK-260720-akf5kh remains blocked by protocol documentation TASK-260720-3lo9jc, and downstream interoperability STORY-260720-21bsr2 remains blocked by both manager stories and the protocol story. This preserves implementation and integration sequencing while allowing solution architecture to enter review.
agent completed: [analyst] solution-architect (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-3b3d1f, pid=77326, exit=0)
spawn queued: [analyst] solution-architect (codex) (run=RUN-260720-a4c1dc, max_parallel=20)
spawn run started: [analyst] solution-architect (codex) (run=RUN-260720-a4c1dc)
Logbook 2026-07-20 — independent 17-task audit retained the decomposition and dependency graph, found no missing contract area and created no duplicate or clarification task. Corrected two evidenced release-boundary contradictions: TASK-260720-12r55p and TASK-260720-3pemm6 now consume the exact candidate suite through CURATOR_CONFORMANCE_ROOT while preserving the committed released-suite pin for TASK-260720-1utsx8 after TASK-260720-25d05o. Added STORY-260720-1uv5gi_17-task-audit.md and TASK-260720-3pemm6_release-boundary-plan.md. PlantUML sources and attached bytes revalidated; canonical plan remains 17 tasks in 12 acyclic phases. Tool anomaly: task-board 0.20.1 persisted both set_details mutations despite --dry-run; a follow-up query proved the write. The intended corrections were valid, so no rollback was needed.
agent completed: [analyst] solution-architect (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-a4c1dc, pid=37452, exit=0)
RELEASE GATE 2026-07-30: Human requires an immediate CocoaSkills release-candidate publication after all 17 core Go parity delivery tasks in this story are independently accepted and landed to ivanopcode/cocoaskills main. Do not wait for the separate release-grade interop/platform layer. Planned semver tag is v0.13.0-rc.1 unless a conflicting tag exists at publication time. No RC may be created before the 17-task gate is complete.

## Precondition Resources
(none)

## Outcome Resources
- [STORY-260720-1uv5gi_spawn-log_-analyst--solution-architect--codex-.log](file://STORY-260720-1uv5gi/STORY-260720-1uv5gi_spawn-log_-analyst--solution-architect--codex-.log) — System spawn log captured by task-board
- [STORY-260720-1uv5gi_decomposition-plan.md](file://STORY-260720-1uv5gi/STORY-260720-1uv5gi_decomposition-plan.md) — Development-ready decomposition, completeness matrix, decisions, risks, and integration gates
- [STORY-260720-1uv5gi_canonical-board-plan.md](file://STORY-260720-1uv5gi/STORY-260720-1uv5gi_canonical-board-plan.md) — Canonical task-board phase plan and critical path snapshot
- [STORY-260720-1uv5gi_planning-validation.md](file://STORY-260720-1uv5gi/STORY-260720-1uv5gi_planning-validation.md) — Board, dependency, completeness, diagram, baseline, and anomaly validation evidence
- [STORY-260720-1uv5gi_17-task-audit.md](file://STORY-260720-1uv5gi/STORY-260720-1uv5gi_17-task-audit.md) — Evidence-backed audit of the existing 17-task csk decomposition and the two corrected release-boundary briefs
