# Review verdict: ACCEPTED

Task: `TASK-260729-iuepxx`  
Reviewer run: `RUN-260729-bd3a14`  
Date: 2026-07-29

## Verdict

Accepted. The audit reconstructs the live cycle correctly, distinguishes the
semantic dependency classes, proves that no link/unlink-only repair can retain
all mandatory phase orderings, and supplies the smallest phase-aligned
structural alternative for orchestrator review. No live dependency or parent
mutation was applied.

Acceptance here is for the audit outcome, not authorization to apply the
proposed structural mutations. The orchestrator should review the acknowledged
wording/status side effects, especially broadening the current external-only
Linux qualification story if the shared local-and-repository qualification
task is moved there.

## Independent graph reconstruction

Compact projections of all tasks in:

- `STORY-260728-16spsm` (Kotlin)
- `STORY-260728-2abmzr` (Swift)
- `STORY-260728-3tymqm` (Rust)
- `STORY-260728-2fsqtv` (toolchain preflight)
- `STORY-260728-2mnlp0` (additional-language protocol)

reproduced exactly 37 cross-story `blockedBy` links and these 14 derived edges
(dependent story → prerequisite story):

| Edge | Source links | Semantic classes present |
|---|---:|---|
| Kotlin → protocol | 3 | hard prerequisite, implementation ordering |
| Kotlin → toolchain | 6 | hard prerequisite, implementation ordering, advisory sequencing |
| Swift → protocol | 3 | hard prerequisite, implementation ordering |
| Swift → toolchain | 6 | hard prerequisite, implementation ordering, advisory sequencing |
| Rust → protocol | 3 | hard prerequisite, implementation ordering |
| Rust → toolchain | 6 | hard prerequisite, implementation ordering, advisory sequencing |
| toolchain → Kotlin | 1 | qualification ordering |
| toolchain → Swift | 1 | qualification ordering |
| toolchain → Rust | 1 | qualification ordering |
| toolchain → protocol | 3 | hard prerequisite, implementation ordering |
| protocol → Kotlin | 1 | hard prerequisite |
| protocol → Swift | 1 | hard prerequisite |
| protocol → Rust | 1 | hard prerequisite |
| protocol → toolchain | 1 | hard prerequisite |

This gives the seven reciprocal pairs reported by the producer:
Kotlin↔protocol, Swift↔protocol, Rust↔protocol, Kotlin↔toolchain,
Swift↔toolchain, Rust↔toolchain, and protocol↔toolchain.

The classifications are supported by the current task descriptions/scopes:

- the three language design tasks consume the accepted protocol boundary and
  shared toolchain contract, while protocol integration consumes those accepted
  language designs;
- Curator language implementations consume protocol qualification and Curator
  preflight; CocoaSkills implementations consume Curator work and CocoaSkills
  preflight;
- `TASK-260728-1e6811` explicitly defers shared Linux preflight qualification
  until all three language Linux qualification tasks;
- the three language documentation tasks wait on official guidance only for
  documentation consistency, so those links are advisory in intent even though
  the board stores them as ordinary hard blockers.

## Minimality and semantic impossibility

The attached producer verifier exits 0 and confirms:

- 37 source links and 14 story edges;
- current cyclicity;
- a minimum weighted feedback cut of seven task unlinks over all 120 story
  orders;
- acyclicity of that seven-unlink cut;
- the mandatory `protocol → language → toolchain → protocol` phase cycle for
  each language.

The mathematical seven-unlink cut is correctly rejected because it removes the
four inputs of `TASK-260728-251p01` plus the three Linux qualification gates of
`TASK-260728-1e6811`. A compensating hard link would recreate the same derived
story direction. The board exposes no advisory link kind. Therefore there is
no semantically correct link/unlink-only solution.

For the structural alternative, an independent exhaustive check used all 126
tasks under the epic's 15 stories and tested all 64 current/proposed parent
assignments for the six phase-boundary tasks. Exactly one assignment was
acyclic, and it moved all six tasks. Thus no smaller subset of the semantically
coherent phase-alignment choices works.

## Exact safe graph proposal for orchestrator review

Do not unlink or link any blocker. The graph-preserving proposal is:

```bash
task-board m --dry-run 'set_parent(TASK-260728-12pnm1, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-1yhuqi, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-168smo, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-1g0z69, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-2jaw7h, parent=STORY-260728-2mnlp0); set_parent(TASK-260728-1e6811, parent=STORY-260728-1eye8p)'
```

All six dry runs parsed and resolved on the live board. Follow-up projections
confirmed the live parents and every blocker stayed unchanged.

## Mechanical verification

| Check | Exit/result |
|---|---|
| Live active child plan before mutation | exit 1; names exactly the five audited cycle stories |
| Producer `TASK-260729-iuepxx_verify-cycle.js` | exit 0 |
| Seven-unlink live dry run | exit 0; no write |
| Six-parent live dry run | exit 0; no write |
| Six parent moves on a fresh isolated board copy | all six accepted |
| Isolated active child plan | exit 0; protocol wave 0, toolchain wave 2, language implementations wave 4, Linux qualification wave 5 |
| Isolated `task-board validate` | exit 0; 43 pre-existing unrelated issues, no target-scoped structural defect |
| Exhaustive semantically allowed parent search | 64/64 tested; exactly one acyclic assignment; minimum six moves |
| Final live parent/blocker projection | unchanged |

The isolated plan preserves product priority: language/protocol/toolchain
specification work precedes Curator implementation, CocoaSkills remains ordered
after Curator through existing task blockers, all Rust/Swift/Kotlin
implementation stories remain backlog after their design tasks move, and shared
Linux qualification remains last. Hardened work is not introduced as a gate.

## Sources

- Live task/story compact projections queried during `RUN-260729-bd3a14`.
- `TASK-260729-iuepxx_dependency-cycle-audit.md`.
- `TASK-260729-iuepxx_verify-cycle.js`.
- Current descriptions/scopes of every source and prerequisite task in the 37
  cross-story links.
- Fresh isolated copy:
  `.temp/TASK-260729-iuepxx/review-board.bcMZVw/` (review scratch only).

No code, product artifact, release state, worktree, live dependency, or live
parent was modified by this review.
