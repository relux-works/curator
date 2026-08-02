## Status
reviewing

## Assigned To
(none)

## Created
2026-07-19T22:04:38Z

## Last Update
2026-07-31T23:11:17Z

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
ORCHESTRATOR CAPACITY POLICY 2026-07-29 04:33 +04: Until 08:00 Asia/Tbilisi, when any Claude Opus 5 run reaches a terminal state, leave that released Opus slot unfilled for at least 30 minutes before spawning another Opus. Do not cancel already-running Opus runs. The currently free slot from the 04:30 Swift handoff remains on cooldown until 05:00 +04. Codex capacity is independent and may be increased for useful non-duplicative work.
Infrastructure update 2026-07-29 04:xx Asia/Tbilisi: Linux validation host is available for the next several hours via ssh lev. Treat Linux as an executable validation surface when it advances the critical path, prioritizing Curator Go and then CocoaSkills Go parity/qualification. Inventory first and keep all remote work non-destructive in private temporary directories. Do not start Rust/Swift/Kotlin implementations; those remain backlog after their specifications.
Opus cooldown ledger update 2026-07-29: RUN-260729-375d85 reached terminal at 04:39:20 +04 without a valid handoff; its slot stays unfilled through 05:09:20 +04. Preserve TASK-260729-2kaopg worktree and retry only after that time. The earlier Swift-released slot remains unfilled through 05:00 +04. Every later Opus terminal event receives its own 30-minute no-refill window until 08:00.
Routing queue at 04:45 +04: when the Swift-derived Opus cooldown expires at 05:00, assign TASK-260728-1yhuqi rework cycle 2 first. The global-status slot remains unavailable until 05:09:20 and is reserved for its short validation/handoff continuation unless a stronger critical-path failure appears. TASK-260728-12pnm1 Rust cycle 4 stays queued for the next eligible Opus slot after those assignments. Do not fill any earlier.
Dependency-cycle remediation applied 2026-07-29 after accepted TASK-260729-iuepxx review: moved Rust/Swift/Kotlin design tasks plus two toolchain wire-design tasks into STORY-260728-2mnlp0, moved shared Linux preflight qualification into STORY-260728-1eye8p, and broadened that story wording to local plus external compiled validation. No prerequisite links were removed. Active child plan is now acyclic with six waves. After the structural move, linked Swift design TASK-260728-1yhuqi to completed measured-evidence prerequisite TASK-260729-rhjxtx; this link is intra-story and preserves acyclicity.
Opus cooldown ledger update: RUN-260728-5fd800 reached terminal without valid handoff at 04:53:58 +04; leave that slot unfilled through 05:23:58 +04. TASK-260720-1nlmvv continuation is validation/handoff only in its preserved worktree. Shared Go build cache was removed inside that run after ENOSPC and affected concurrent evidence; no continuation may clear shared caches. Current Opus slots: three active before 05:00 (Kotlin, CocoaSkills parity, plus none for completed CLI? actually two active), one eligible at 05:00 for Swift, one at 05:09:20 for global-status continuation, and the CLI slot eligible at 05:23:58; any new terminal Opus creates its own 30-minute window until 08:00.
Cooldown ledger correction: at 04:54 +04 exactly two Opus runs remain active (Kotlin cycle-3 and CocoaSkills parity rework). Three slots are cooling/idle: Swift-eligible 05:00, global-status-eligible 05:09:20, Curator CLI-eligible 05:23:58. This sentence supersedes the ambiguous active-count parenthetical in the preceding note.
OPUS COOLDOWN LEDGER 2026-07-29: Kotlin cycle-3 run RUN-260729-4d0a9a reached terminal at 05:14:11 Asia/Tbilisi. Its specific Opus slot must remain unfilled until 05:44:11 under the operator limit-protection rule; reviewer work may proceed on Codex.
OPUS COOLDOWN LEDGER 2026-07-29: global-status continuation RUN-260729-57b1c2 reached terminal at 05:20:14 Asia/Tbilisi without handoff; leave its specific Opus slot unfilled until 05:50:14. A separate Codex tester may validate the preserved worktree during cooldown.
Opus cooldown ledger: Swift specification rework run RUN-260729-f9b5c2 reached terminal success at 2026-07-29 05:41:22 +0400. This exact Opus slot MUST remain unfilled until 2026-07-29 06:11:22 +0400 under the operator policy active before 08:00 Asia/Tbilisi.
Opus cooldown ledger: Curator currentness/repair continuation run RUN-260729-39026a reached terminal success at 2026-07-29 05:55:25 +0400. This exact Opus slot MUST remain unfilled until 2026-07-29 06:25:25 +0400 under the operator policy active before 08:00 Asia/Tbilisi.
Opus cooldown ledger: RUN-260729-5cd42e (Kotlin/Native spec cycle 4) completed 2026-07-29 05:56:49 +04. Its exact producer slot must remain unfilled until 2026-07-29 06:26:49 +04; do not cancel other running Opus jobs.
Opus cooldown ledger: RUN-260729-ce59e2 (Rust specification rework cycle 4) completed 2026-07-29 06:06:07 +04. Its exact producer slot must remain unfilled until 2026-07-29 06:36:07 +04; do not cancel other running Opus jobs.
Opus cooldown ledger: RUN-260729-3bd9db completed 2026-07-29 06:46:50 +04; exact slot must remain unfilled until 2026-07-29 07:16:50 +04.
Opus cooldown ledger: RUN-260729-5f494f (rc.5 build-driver goldens) completed 2026-07-29 06:50:59 +04; exact slot must remain unfilled until 2026-07-29 07:20:59 +04.
Opus cooldown ledger: RUN-260729-abbd3c (Swift specification rework cycle 4) completed 2026-07-29 06:54:51 +04; exact slot must remain unfilled until 2026-07-29 07:24:51 +04.
Opus cooldown ledger: RUN-260729-b1885e (Go fingerprint optimization producer) completed 2026-07-29 07:10:30 +04; exact slot must remain unfilled until 2026-07-29 07:40:30 +04.
Opus cooldown ledger: RUN-260729-0e838f (Curator currentness/repair cycle 3) completed 2026-07-29 07:39:49 +04; exact slot must remain unfilled until 2026-07-29 08:09:49 +04.
OPUS SLOT COOLDOWN: RUN-260729-19d434 (fingerprint rework) reached terminal at 2026-07-29 07:56:05 +04. Do not reuse that Claude Opus 5 slot before 2026-07-29 08:26:05 +04. The run left TASK-260729-1zex8r in development; route evidence completion separately.
OPUS SLOT COOLDOWN: RUN-260729-dfd79e (Swift specification cycle-5 rework) reached terminal at 2026-07-29 07:57:41 +04. Do not reuse that Claude Opus 5 slot before 2026-07-29 08:27:41 +04. TASK-260728-1yhuqi handed off to independent Codex review at 18/18.

## Precondition Resources
- [EPIC-260720-21aq1i_protocol-candidate-handoff.md](file://EPIC-260720-21aq1i/EPIC-260720-21aq1i_protocol-candidate-handoff.md) — Candidate-suite handoff and release-only blocker boundary

## Outcome Resources
(none)
