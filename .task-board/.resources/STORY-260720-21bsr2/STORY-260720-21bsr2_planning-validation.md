# STORY-260720-21bsr2 planning validation

## Board readiness

- Child tasks: 12.
- Tasks with placeholder or empty description, scope, or acceptance criteria: 0.
- Tasks with fewer than three task-specific checklist items: 0.
- Tasks with no dependency: 0. The first internal phase is intentionally held
  behind protocol tasks, and implementation ordering across the protocol,
  Curator, and csk stories is preserved by task-to-task links.
- Canonical internal plan: 8 phases.
- Critical path: TASK-260720-2g7avf -> TASK-260720-1673lr ->
  TASK-260720-3nj1r6 -> TASK-260720-3pvihp -> TASK-260720-vs6den ->
  TASK-260720-25d05o -> TASK-260720-1utsx8 -> TASK-260720-22ynoi.
- Cross-story task handoffs are explicit for the protocol fixture/lifecycle/docs,
  Curator conformance, csk conformance, and both manager CI audits.

`task-board validate` exited 0. It reported only the existing twelve
EPIC-260712 prose-style dependency references and the unrelated orphan
`.resources/TASK-260713-7a9c1e/review.md`. No issue names this story, its twelve
tasks, or their resources.

## Diagram validation

PlantUML syntax checking and SVG rendering exited 0 for both focused diagrams.
Both SVG files passed `xmllint --noout` and both PNG renders were visually
inspected after iteration.

| Artifact | SHA-256 |
|---|---|
| TASK-260720-2g7avf_independent-consumers.puml | e139142e285b71bbb79e7927de49e617f8897fc0e45352826fe94c3792e3d5e2 |
| TASK-260720-2g7avf_independent-consumers.svg | 828d0ca3beb1e446c8efff35c287f1ce29f9ae5c434725e486a926327798f63c |
| TASK-260720-3pvihp_release-evidence-gates.puml | 51ab901d49f851ce4cceb667549bfbd76106533ecfec22b9c7e3ec661afec688 |
| TASK-260720-3pvihp_release-evidence-gates.svg | 02503f8cf40a888a3c80a2703b29f5bbd42501c35ac631c49a8b6d308295fb11 |

The installed Graphviz `dot` binary is currently unusable because its Homebrew
binary cannot load `libltdl.7.dylib`. The component diagram therefore declares
PlantUML Smetana explicitly. This is a local rendering-tool anomaly, not a
project architecture blocker; PlantUML check, render, XML validation, and visual
inspection all succeeded.

## Release and scope checks

- Git diffs for Curator `.github/workflows/ci.yml`, curator-spec
  `.github/workflows/implementations.yml`, and csk `.github/workflows/ci.yml`
  are empty. No release or implementation pin changed during decomposition.
- Cross-story notes state that candidate tests use a caller-supplied
  `CURATOR_CONFORMANCE_ROOT`; committed pins wait for immutable published
  release qualification.
- No product, test, documentation, workflow, release metadata, or package
  implementation was changed. This run created board tasks, dependencies,
  notes, plans, and diagrams only.

## Recorded anomaly and resolution

An attempted task-to-sibling-story dependency caused task-board hierarchy
escalation to infer that STORY-260720-21bsr2 was blocked by its own parent epic.
Both cross-level task links were immediately removed and replaced with direct
task-to-task handoffs. The first review-handoff attempt then showed that
task-board also treats story prerequisites in `to-review` as unfinished
implementation blockers. The three redundant story-to-story links were removed;
the more precise child-task dependencies continue to enforce every protocol,
Curator, and csk prerequisite without blocking this planning-only lifecycle.
