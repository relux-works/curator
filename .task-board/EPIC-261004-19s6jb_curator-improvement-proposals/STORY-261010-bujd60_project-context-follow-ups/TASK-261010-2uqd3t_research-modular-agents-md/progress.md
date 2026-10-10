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
- [x] Research resource attached with every claim cited
- [x] Options compared with trade-offs and one recommendation
- [x] No implementation, no tests or builds on this host
- [x] No secrets, personal paths or host names
- [x] Findings written to file
- [x] Key aspects highlighted
- [x] Fact-checking performed — claims verified, sources cited
- [x] Findings linked on the board as a new task-scoped outcome resource
- [x] All questions from task description answered
- [x] Important findings recorded in research and task notes; LOGBOOK.md untouched as required by modular-instructions-brief.md
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn selection rationale tuple: {"role":"researcher","pair":"gpt-6-astra/max","text":"owner ask 2026-10-10 (issue 114): preliminary design research, astra max as for the earlier CIP research"}
spawn selection rationale for gpt-6-astra/max: owner ask 2026-10-10 (issue 114): preliminary design research, astra max as for the earlier CIP research
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (codex) (run=RUN-261010-ac6189, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-261010-ac6189)
Research evidence attached as TASK-261010-2uqd3t/modular-instructions-design.md; source file .research/261010_modular-instructions-design.md. Recommendation: derived project chapter plan, protected per-environment output, native discovery only as explicit legacy compatibility; first bounded slice is preview plus protected Codex materialization, with strict launch unavailable until adapter qualification. Covers all six brief questions and five harness inventories. Key findings: existing profile chapter composition is reusable; pure-umbrella roots need explicit project-surface handling; Pi no-context-files removes global context too; Codex has separate home/project loaders; Claude project-rule exclusion fixes predate the pinned release; OpenCode has no verified pin and V2 ignores the referenced instructions array; Muse has no admitted root target. Used updated PR 136 head 2f0531b4edcc99c6118c277deb00e8736392c04d including profile alternatives/stacks. No implementation, harness execution, tests or builds. Editorial check exit 0: 31/31 reference IDs, 21/21 fully pinned blob links, 13/13 local source locators (git show exit 0 each), formatting/hygiene and sole-research-file delta. New-file diff command exited 1 with no diagnostics, expected nonempty-file versus empty comparison; not counted as a passing gate. Artifact 38607 bytes, SHA-256 819b52420512f160a46de07d9d12cc7bb04855fb401571741e081405934aa80d. Campaign brief explicitly forbids LOGBOOK.md edits: findings are preserved here and in the study; inherited logbook checklist item is replaced by that applicable campaign requirement.
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-261010-ac6189, pid=13632, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/low","text":"R138 reviewer (codex gpt-6-astra low) for the issue-114 design study"}
spawn selection rationale for gpt-6-astra/low: R138 reviewer (codex gpt-6-astra low) for the issue-114 design study
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261010-e69e1a, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261010-e69e1a)
Reviewer accepts rev1: 6/6 questions covered; 14/14 sampled claims supported across Curator composition/materialization, CIP admission and all five harness inventories. Exact candidate delta is only the research file; LOGBOOK untouched. No harness, tests or builds run. Tests-green checklist is N/A under the explicit research-only brief, not a test attestation; nonacceptance-routing checklist is N/A because accepted. Verdict attached as TASK-261010-2uqd3t_review-verdict-rev1.md. Native runtime qualification remains 0/5.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261010-e69e1a, pid=80297, exit=0)
spawn selection rationale tuple: {"role":"researcher","pair":"gpt-6-astra/low","text":"bound 2uqd3t-checkpoint (land queue); codex gpt-6-astra low"}
spawn selection rationale for gpt-6-astra/low: bound 2uqd3t-checkpoint (land queue); codex gpt-6-astra low
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (codex) (run=RUN-261010-41a644, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-261010-41a644)
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-261010-41a644, pid=3495, exit=0)

## Precondition Resources
- [modular-instructions-brief.md](file://TASK-261010-2uqd3t/modular-instructions-brief.md)
- [modular-md-review-brief.md](file://TASK-261010-2uqd3t/modular-md-review-brief.md)
- [2uqd3t-integrate-land.md](file://TASK-261010-2uqd3t/2uqd3t-integrate-land.md)

## Outcome Resources
- [TASK-261010-2uqd3t_spawn-log_-analyst--researcher--codex-_RUN-261010-ac6189.log](file://TASK-261010-2uqd3t/TASK-261010-2uqd3t_spawn-log_-analyst--researcher--codex-_RUN-261010-ac6189.log) — System spawn log captured by task-board
- [TASK-261010-2uqd3t/modular-instructions-design.md](file://TASK-261010-2uqd3t/TASK-261010-2uqd3t/modular-instructions-design.md) — Preliminary modular-instruction design: cited five-harness inventory, chapter model, admission, migration, options and recommended first slice.
- [TASK-261010-2uqd3t_change-request_rev1.patch](file://TASK-261010-2uqd3t/TASK-261010-2uqd3t_change-request_rev1.patch) — Change Request CR-TASK-261010-2uqd3t-1 revision 1 candidate patch (repository_delta=present, 1 changed paths)
- [TASK-261010-2uqd3t_change-request_rev1-validation.log](file://TASK-261010-2uqd3t/TASK-261010-2uqd3t_change-request_rev1-validation.log) — Change Request CR-TASK-261010-2uqd3t-1 revision 1 bounded validation log
- [TASK-261010-2uqd3t_spawn-log_-reviewer--reviewer--codex-_RUN-261010-e69e1a.log](file://TASK-261010-2uqd3t/TASK-261010-2uqd3t_spawn-log_-reviewer--reviewer--codex-_RUN-261010-e69e1a.log) — System spawn log captured by task-board
- [TASK-261010-2uqd3t_review-verdict-rev1.md](file://TASK-261010-2uqd3t/TASK-261010-2uqd3t_review-verdict-rev1.md) — Accepted rev1: six-question coverage, 14 source checks, exact delta and read-only boundary
- [TASK-261010-2uqd3t_spawn-log_-analyst--researcher--codex-_RUN-261010-41a644.log](file://TASK-261010-2uqd3t/TASK-261010-2uqd3t_spawn-log_-analyst--researcher--codex-_RUN-261010-41a644.log) — System spawn log captured by task-board
- [TASK-261010-2uqd3t_integration-land.md](file://TASK-261010-2uqd3t/TASK-261010-2uqd3t_integration-land.md) — Fresh accepted-revision and byte-identity checks for runner-owned landing; no producer landing attempted

## Created
2026-10-10T02:26:48Z

## Last Update
2026-10-10T10:20:28Z

## Assigned To
[analyst] researcher (codex)
