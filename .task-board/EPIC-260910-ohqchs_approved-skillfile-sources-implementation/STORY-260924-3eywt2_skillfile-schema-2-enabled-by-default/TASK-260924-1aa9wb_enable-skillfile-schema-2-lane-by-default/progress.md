## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- TASK-260924-1r66o7
- TASK-260924-m28s6b
- TASK-260924-11burj

## Checklist
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Schema 2 is the default project reader with no source switch; real CLI resolve/install/status/help work with the variable unset, and the removed variable is ignored.
- [x] Schema 1 remains read-only and byte-identical in base-on versus candidate-unset CLI outputs; schema-1-to-schema-2 routing and renewed schema-2 refusal mutants are both killed.
- [x] README, CLI/troubleshooting docs, help, and CHANGELOG describe default schema-2 support without opt-in/refusal wording; tests prove the removed variable is ignored.
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6-luna/max","text":"operator decision 2026-09-24: schema 2 lane on by default; highest priority"}
spawn selection rationale for gpt-6-luna/max: operator decision 2026-09-24: schema 2 lane on by default; highest priority
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-260924-e700b3, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260924-e700b3)
spawn run RUN-260924-e700b3 cancelled by operator; operator action required; reason: no operator reason supplied
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6-luna/max","text":"operator handoff 2026-09-24: remove the switch; schema 2 default reader path"}
spawn selection rationale for gpt-6-luna/max: operator handoff 2026-09-24: remove the switch; schema 2 default reader path
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-260924-0d2460, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260924-0d2460)
agent completed: [implementer] developer (codex) (exit=-1)
spawn run completed: codex (run=RUN-260924-e700b3, pid=34447, exit=-1)
The attached skillfile-default-on-handoff-20260924.md and 1aa9wb-brief-v2.md supersede the original checklist wording: the switch and opt-out are removed, and CURATOR_DRAFT_SOURCES_V1 is tested as ignored. The original checklist items 1-3 describe the superseded =0 opt-out contract; equivalent current-scope items are added below. See TASK-260924-1aa9wb_results.md for implementation choices, test exit codes, byte-identity evidence, killed mutants, and broad-suite timeouts. No LOGBOOK.md edit per campaign rule; findings are recorded in the task-scoped result.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260924-0d2460, pid=42637, exit=0)
run write-boundary clearance for RUN-260924-0d2460: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-260924-e700b3: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
spawn selection rationale tuple: {"role":"reviewer","pair":"claude-opus-5-5/low","text":"operator policy 2026-09-23: opus-5-5 low full; schema 2 default-on review"}
spawn selection rationale for claude-opus-5-5/low: operator policy 2026-09-23: opus-5-5 low full; schema 2 default-on review
spawn agent resolution: Agent selection: claude via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (claude) (run=RUN-260924-1cb8ce, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260924-1cb8ce)
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260924-1cb8ce, pid=38224, exit=0)
run write-boundary clearance for RUN-260924-1cb8ce: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
spawn selection rationale tuple: {"role":"developer","pair":"muse-spark-1.3-contributor/xhigh","text":"bound 1aa9wb-checkpoint (lock queue); muse xhigh full"}
spawn selection rationale for muse-spark-1.3-contributor/xhigh: bound 1aa9wb-checkpoint (lock queue); muse xhigh full
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (muse) (run=RUN-260924-ba5d20, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260924-ba5d20)
spawn run child final message (run=RUN-260924-ba5d20, tools=3 patches=0 failed=0):
Checkpoint recorded as `41dd18dd9e941c6dd1d14e7fd82c3b52b753061e` on `task-board/story/STORY-260924-3eywt2` (status integrating, exit 0). Log `.temp/checkpoint-1aa9wb.log` attached as outcome `TASK-260924-1aa9wb_checkpoint-results.md`. Stopping per checkpoint instruction with no further board writes.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-260924-ba5d20, pid=60366, exit=0)

## Precondition Resources
- [1aa9wb-checkpoint-instruction.md](file://TASK-260924-1aa9wb/1aa9wb-checkpoint-instruction.md)

## Outcome Resources
- [TASK-260924-1aa9wb_spawn-log_-implementer--developer--codex-_RUN-260924-e700b3.log](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_spawn-log_-implementer--developer--codex-_RUN-260924-e700b3.log) — System spawn log captured by task-board
- [1aa9wb-brief.md](file://TASK-260924-1aa9wb/1aa9wb-brief.md)
- [TASK-260924-1aa9wb_spawn-log_-implementer--developer--codex-_RUN-260924-0d2460.log](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_spawn-log_-implementer--developer--codex-_RUN-260924-0d2460.log) — System spawn log captured by task-board
- [TASK-260924-1aa9wb_results.md](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_results.md) — Implementation, compatibility, mutation and bounded validation evidence
- [TASK-260924-1aa9wb_v1-cli-identity.json](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_v1-cli-identity.json) — Base versus candidate schema-1 CLI output comparison hashes
- [TASK-260924-1aa9wb_change-request_rev1.patch](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_change-request_rev1.patch) — Change Request CR-TASK-260924-1aa9wb-1 revision 1 candidate patch (repository_delta=present, 49 changed paths)
- [TASK-260924-1aa9wb_change-request_rev1-validation.log](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_change-request_rev1-validation.log) — Change Request CR-TASK-260924-1aa9wb-1 revision 1 bounded validation log
- [1aa9wb-brief-v2.md](file://TASK-260924-1aa9wb/1aa9wb-brief-v2.md)
- [TASK-260924-1aa9wb_spawn-log_-reviewer--reviewer--claude-_RUN-260924-1cb8ce.log](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_spawn-log_-reviewer--reviewer--claude-_RUN-260924-1cb8ce.log) — System spawn log captured by task-board
- [TASK-260924-1aa9wb_review-verdict-rev1.md](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_review-verdict-rev1.md) — Reviewer verdict rev1 accepted
- [1aa9wb-review-note.md](file://TASK-260924-1aa9wb/1aa9wb-review-note.md)
- [campaign-producer-rules.md](file://TASK-260924-1aa9wb/campaign-producer-rules.md)
- [skillfile-default-on-handoff-20260924.md](file://TASK-260924-1aa9wb/skillfile-default-on-handoff-20260924.md)
- [TASK-260924-1aa9wb_spawn-log_-implementer--developer--muse-_RUN-260924-ba5d20.log](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_spawn-log_-implementer--developer--muse-_RUN-260924-ba5d20.log) — System spawn log captured by task-board
- [TASK-260924-1aa9wb_checkpoint-results.md](file://TASK-260924-1aa9wb/TASK-260924-1aa9wb_checkpoint-results.md) — Checkpoint log for TASK-260924-1aa9wb on Story branch

## Created
2026-09-24T02:16:55Z

## Last Update
2026-09-25T13:39:49Z

## Assigned To
[implementer] developer (muse)
