## Status
backlog

## Assigned To
[analyst] solution-architect (codex)

## Created
2026-07-19T22:10:05Z

## Last Update
2026-07-28T23:57:15Z

## Blocked By
- STORY-260720-1uv5gi

## Blocks
- (none)

## Checklist
- [x] Decompose cross-manager conformance, fixture, authoring guidance, language-matrix, and interoperability verification into atomic tasks with explicit dependencies; planning only, then leave the story at to-dev
- [x] Tasks created with description and AC
- [x] Dependencies linked
- [x] Tasks are atomic — one clear deliverable each
- [x] Completeness verified — nothing forgotten
- [x] Gaps closed with blocking tasks
- [x] Diagrams or planning artifacts linked as new task-scoped outcome resources
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Audit the existing interoperability decomposition for shared-suite ownership, parity coverage, dependency correctness, and non-fabricated release evidence; do not duplicate tasks, correct only evidenced defects

## Notes
spawn queued: [analyst] solution-architect (codex) (run=RUN-260720-04caa8, max_parallel=20)
spawn run started: [analyst] solution-architect (codex) (run=RUN-260720-04caa8)
Logbook 2026-07-20 — solution architecture handoff: accepted research TASK-260720-poa3ze_compile-only-build-drivers.md at SHA-256 6308d99d8bdad4445841bc1cbd230cadbed0020012d0e9d38d877b413348f681. Decomposed the story into 12 atomic tasks across 8 phases. curator-spec alone owns the shared executable oracle; Curator and csk remain native independent consumers, and a black-box runner compares raw observable results. Release ordering uses two evidence gates: qualify real manager releases against the exact candidate-suite digest, then pin those releases in curator-spec, qualify the actual protocol release, and only then audit manager CI pins. Candidate tests may use caller-supplied CURATOR_CONFORMANCE_ROOT; no unreleased ref or release claim may be committed. A task-to-sibling-story link briefly triggered hierarchy escalation to the parent epic; it was removed and replaced with task-to-task dependencies. Local Graphviz is missing libltdl.7.dylib, so diagrams use PlantUML Smetana; syntax, SVG/XML, hashes, and visual inspection passed. task-board validate names only pre-existing legacy issues, not this story. The inherited checklist says to-dev, but this assigned run explicitly requires the solution-architect handoff at to-review; that role lifecycle governs. Planning only: no product code, workflow pin, or release metadata changed.
Logbook 2026-07-20 — review-handoff lifecycle adjustment: task-board rejected the first to-review transition because it counts prerequisite stories already at to-review as unfinished implementation blockers. The three redundant story-level links were removed; no prerequisite was discarded. Direct child-task links still gate the shared protocol cases and docs, both native manager consumers, candidate verification, manager CI handoffs, evidence qualification, pin promotion, and released-suite audits. This preserves development ordering while allowing the planning-only story to enter its mandated solution-architecture review lifecycle.
agent completed: [analyst] solution-architect (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-04caa8, pid=77435, exit=0)
spawn queued: [analyst] solution-architect (codex) (run=RUN-260720-7b4c25, max_parallel=20)
spawn run started: [analyst] solution-architect (codex) (run=RUN-260720-7b4c25)
Logbook 2026-07-20 — interoperability decomposition audit: retained the existing 12 atomic tasks and 8 internal phases; no duplicate task was created. Closed three development-readiness defects: TASK-260720-2g7avf now owns the suite result transport and deterministic JSON Lines record contract required before both native consumers; TASK-260720-31zeo2 now actually waits for TASK-260720-3pemm6 real-Go E2E; and TASK-260720-3pvihp now waits for both manager integration handoffs TASK-260720-1pvfj5 and TASK-260720-3s27te. Aligned candidate-suite wording across those sibling handoffs and corrected TASK-260720-vs6den to precede manager released-suite pins. Candidate evidence uses an explicit immutable non-default input and never advances the default committed pin; evidence tasks block on absent public releases instead of inventing refs or claims. Updated and re-rendered both task-linked diagrams, refreshed TASK-260720-3nj1r6_interop-plan.md, and attached new outcomes STORY-260720-21bsr2_canonical-audit-plan.md plus STORY-260720-21bsr2_decomposition-audit.md. Validation: 12 tasks, no empty brief, no duplicate names, at least 3 checklist items each, acyclic 59-task related plan, diagram syntax/SVG/XML/visual checks pass, board resources match canonical bytes, and workflow pin diffs are empty. task-board validate reports only the same 13 unrelated legacy issues. New task links again caused coarse story auto-escalation; the redundant story blockers were removed while all exact child-task gates remain.
agent completed: [analyst] solution-architect (codex) (exit=0)
spawn run completed: codex (run=RUN-260720-7b4c25, pid=37449, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [STORY-260720-21bsr2_spawn-log_-analyst--solution-architect--codex-.log](file://STORY-260720-21bsr2/STORY-260720-21bsr2_spawn-log_-analyst--solution-architect--codex-.log) — System spawn log captured by task-board
- [STORY-260720-21bsr2_decomposition-plan.md](file://STORY-260720-21bsr2/STORY-260720-21bsr2_decomposition-plan.md) — Development-ready task map, acceptance coverage, architecture decisions, gaps, and release sequence
- [STORY-260720-21bsr2_canonical-board-plan.md](file://STORY-260720-21bsr2/STORY-260720-21bsr2_canonical-board-plan.md) — Canonical task-board phase plan and critical path snapshot
- [STORY-260720-21bsr2_planning-validation.md](file://STORY-260720-21bsr2/STORY-260720-21bsr2_planning-validation.md) — Board, completeness, dependency, diagram, release-pin, and resource validation evidence
- [STORY-260720-21bsr2_canonical-audit-plan.md](file://STORY-260720-21bsr2/STORY-260720-21bsr2_canonical-audit-plan.md) — Canonical post-audit task-board phase plan and critical path snapshot
- [STORY-260720-21bsr2_decomposition-audit.md](file://STORY-260720-21bsr2/STORY-260720-21bsr2_decomposition-audit.md) — Post-audit corrections, ownership matrix, dependency sequence, release-evidence boundary, and validation evidence
