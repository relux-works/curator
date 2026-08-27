## Status
development

## Assigned To
(none)

## Created
2026-07-19T22:04:38Z

## Last Update
2026-07-28T23:33:14Z

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
(empty)

## Notes
Logbook 2026-07-20 — compiled-build interoperability release-order decision: TASK-260720-3pvihp qualifies immutable Curator and csk releases against the exact candidate-suite digest; TASK-260720-vs6den then pins those real manager releases in curator-spec; TASK-260720-25d05o qualifies the actual protocol release; TASK-260720-38l1sy and TASK-260720-1utsx8 audit the managers’ committed suite pins afterward. Candidate testing uses explicit CURATOR_CONFORMANCE_ROOT and must not masquerade as published release evidence. This breaks the manager/spec pin cycle without unreleased commits.
Execution policy set by the human on 2026-07-20: orchestrate only through the project-management skill; use one parent Codex orchestrator and at most one sequential child run at a time; every producer, reviewer, and rework run must use agent=codex, model=gpt-5.6-sol, reasoning-effort=high; do not use Claude Code, Claude agents, other agent environments, or any other model. Producer and reviewer remain separate sequential runs.
MODEL ROUTING OVERRIDE (2026-07-28, human-authorized): For remaining declarative compiled-build and external-build-repository delivery, launch every producer/implementer/rework child with --agent claude --model claude-opus-5 (Claude models do not take --reasoning-effort). Keep the active orchestrator and independent reviewer children on Codex gpt-5.6-sol with --reasoning-effort high unless the human changes reviewer routing separately. Enforce one active child at a time. This supersedes the earlier all-Sol implementer preference but does not waive any producer -> independent reviewer acceptance cycle or human architecture boundary.
PORTABLE/HARDENED SCOPE DECISION (human-authorized 2026-07-28): current rc.5, Curator and csk delivery targets the maximum autonomously enforceable portable manager-worker-v1 profile on macOS and Windows. Complete fail-closed host isolation is not a release/development dependency and is tracked in separate EPIC-260728-2m6dqo / STORY-260728-327soo. TASK-260728-zb2s4z owns the unreleased spec/vector amendment before manager implementation resumes. Portable outputs must identify their execution policy and cannot alias future hardened cache/receipt/marker/claim state. No dependency may point from this epic to the hardened follow-up epic.
PARALLELISM OVERRIDE (human-authorized 2026-07-28): accelerate remaining work with up to two simultaneous Claude Opus 5 producer/implementer/rework children when tasks are independent and use separate task-owned worktrees. Never run two producers on the same task or overlap a task producer with its reviewer. Keep independent reviewers on Codex gpt-5.6-sol/high and preserve producer -> reviewer acceptance before counting done. Integration tasks own convergence of parallel spec branches; predecessor accepted worktrees remain read-only.
PARALLELISM OVERRIDE 2 (human-authorized 2026-07-28): raise the producer ceiling from two to three simultaneous Claude Opus 5 producer/implementer/rework children when tasks are independent and use separate task-owned worktrees. Reviewers remain independent Codex gpt-5.6-sol/high runs and do not consume the three-producer ceiling. Never overlap producer and reviewer on the same task; preserve producer -> reviewer acceptance and integration ownership.
PRIORITY DIRECTIVE 2026-07-29: finish and independently accept the Rust, Swift and Kotlin driver specifications, but keep all Rust/Swift/Kotlin Curator and CocoaSkills implementation tasks in backlog. Complete the remaining Curator Go implementation, diagnostics, documentation, conformance and cross-platform CI first. Immediately after Curator Go acceptance, make CocoaSkills Go parity the main implementation priority. Separate hardened-execution work may use surplus agents in parallel but remains non-gating; Linux qualification remains later.
PARALLELISM OVERRIDE 3 (human-authorized 2026-07-29): raise the independent Claude Opus 5 producer/implementer/rework ceiling from three to five. Keep each task in a separate task-owned worktree, never overlap producer and reviewer on the same task, and keep independent acceptance reviewers on Codex gpt-5.6-sol/high. One additional Sol/high analysis or review child may run concurrently when it does not race an existing verdict.

## Precondition Resources
- [EPIC-260720-21aq1i_protocol-candidate-handoff.md](file://EPIC-260720-21aq1i/EPIC-260720-21aq1i_protocol-candidate-handoff.md) — Candidate-suite handoff and release-only blocker boundary

## Outcome Resources
(none)
