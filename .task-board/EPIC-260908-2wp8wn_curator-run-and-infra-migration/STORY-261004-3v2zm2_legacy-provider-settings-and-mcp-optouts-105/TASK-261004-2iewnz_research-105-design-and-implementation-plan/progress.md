## Status
done

## Review
required

## Task Class
research

## Estimate
notEstimated

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Research document under .research/ with options, tradeoffs, recommendation and evidence (file:line or measured)
- [x] No code changes; no LOGBOOK.md edit; no secrets read or printed
- [x] Findings written to file
- [x] Key aspects highlighted
- [x] Fact-checking performed — claims verified, sources cited
- [x] Findings linked on the board as a new task-scoped outcome resource
- [x] All questions from task description answered
- [x] Important findings, decisions and anomalies recorded in CIP-0006, companion evidence and board notes; LOGBOOK.md untouched as required by the task brief
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [x] R1 reproduction scripts green: 18/18, 7/7, 11/11 checks, each exit 0 standalone on 2026-10-05 (product Go gates out of scope per verdict/hosted-evidence mode)
- [x] Tests green
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn selection rationale tuple: {"role":"researcher","pair":"gpt-6-astra/max","text":"tb-R164/R139 wave 1b 2026-10-04: researcher gpt-6-astra max"}
spawn selection rationale for gpt-6-astra/max: tb-R164/R139 wave 1b 2026-10-04: researcher gpt-6-astra max
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (codex) (run=RUN-261004-4589ce, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-261004-4589ce)
Research for TASK-261004-2iewnz — research-105-design-and-implementation-plan: pinned curator main ca1b776 and spec rc.14 43bf0a2; Decision 0014 remains proposed. Recommended direction is a closed machine-owned per-profile settings knob on env config plus MCP negative policy, without package authority expansion. Key findings: overlays are profile-scoped, empty MCP output omits the channel, old Codex seeds retain inherited servers. Five selected baseline tests exited 0; scratch settings/MCP package probes refused with exit 1 as expected. LOGBOOK.md remains untouched per the binding brief; important findings will be recorded in the CIP and companion evidence.
TASK-261004-2iewnz — research-105-design-and-implementation-plan: CIP-0006 draft and companion evidence are attached as two new task-scoped outcome resources. Recommendation: typed machine-owned profile_settings on existing config/resolve surfaces; keep Decision 0014 proposed; preserve source scopes and explicit MCP negatives using key ownership and a versioned fragment. Eight sized implementation leaves and five operator decisions are included. Current shipped Codex seed A and empty-MCP channel omission require explicit migration/enforcement coverage. Validation: build exit 0; 5/5 selected baseline tests exit 0; expected schema/config refusal probes exit 1 or 2 as recorded; artifact checker and git diff --check reruns exit 0. Initial artifact-check interruptions exited 130 and are recorded without claiming a pass. Artifact checks verified 10/10 template sections, 38/38 pinned source locations, 5/5 JSON examples, and only two untracked research files. No product code, LOGBOOK.md, commits, or real provider credentials were touched. Checklist item 8 is intentionally unchecked: the binding brief forbids a LOGBOOK.md edit; important findings are recorded in the CIP, evidence, and these board notes instead. Ready for researcher review handoff.
TASK-261004-2iewnz — research-105-design-and-implementation-plan: the first researcher handoff exited 1 because generic checklist item 8 required a logbook entry despite the binding task brief prohibiting a LOGBOOK.md edit. Reconciled that obsolete item with the explicit brief: findings must be recorded in the attached CIP-0006, companion evidence and board notes, while LOGBOOK.md remains untouched. The replacement item is verified against the two attached artifacts and existing notes. No artifact or product-code changes were needed. The checklist-repair preview exited 0; handoff will be retried with all eight applicable items checked.
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-261004-4589ce, pid=69348, exit=0)
spawn run RUN-261004-4589ce cancelled by operator; operator action required; reason: tb-R136 host cap: parked by orchestrator; resume later
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/medium","text":"tb-R164 reviewer gpt-6-astra medium; tb-R136 health-gated queue"}
spawn selection rationale for gpt-6-astra/medium: tb-R164 reviewer gpt-6-astra medium; tb-R136 health-gated queue
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261004-9bf8df, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261004-9bf8df)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261004-9bf8df, pid=33338, exit=0)
spawn selection rationale tuple: {"role":"researcher","pair":"muse-spark-1.3-contributor/max","text":"tb-R164 researcher muse-spark-1.3-contributor max; tb-R136 health-gated queue"}
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 researcher muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (muse) (run=RUN-261005-4b0ac5, max_parallel=20)
spawn run started: [analyst] researcher (muse) (run=RUN-261005-4b0ac5)
R1 rework (researcher, 2026-10-05): verdict blocking finding R1 addressed in evidence E8. Original inline Python bodies were unrecoverable, so three labeled replacement reproductions were written (install/publication assertions 18 checks, config-refusal assertions 7 checks, artifact checker 11 gated checks), each rerun once standalone with real exit 0 on 2026-10-05 against pinned curator ca1b776 and spec v1.0.0-rc.14. Observed: probe exits 1/1/0 and 2/1 reproduced; 10/10 sections, 38/38 links, 10/10 JSON, 2/2 files, 0 pattern matches. Historical E5/E7 results preserved as historical. New finding F-R1-size: combined artifacts now 96520 bytes vs the 81920 historical budget; overrun is forced by the mandated script bodies, recommendation to accept is in E8.5. Evidence outcome resource updated; CIP draft unchanged; no product code, LOGBOOK.md, secrets, or logins touched. Ready for review.
Checklist reconciliation (researcher, 2026-10-05, same precedent as the item-8 repair): checked 9 (deliverables match AC: 4 options, tradeoffs, recommendation, 38 resolved citations, rerun probes, no code/secrets/LOGBOOK changes), 10 (option-A design reuses existing env-config/resolver/marker/fragment surfaces verified at the pinned commit), 12 (verdict outcome exists on board, route analysis taken, R1 rework executed). Removed obsolete 13 (logbook edit conflicts with the binding brief; already covered by item 8) and generic 11 (product-suite Tests green; verdict scopes Go gates out of hosted-evidence rework), replacing 11 with the verified reproduction-script gate, checked. All 12 current items green.
spawn run child final message (run=RUN-261005-4b0ac5, tools=92 patches=10 failed=2):
R1 is resolved and the task is handed off to review (`to-review`, checklist 12/12).

The review verdict's only blocking finding was reproducibility: the original inline Python bodies were unrecoverable. I wrote three labeled replacement reproductions — not claimed to be the originals — with pinned inputs (curator `ca1b776f`, spec `v1.0.0-rc.14`), fixture construction, and configuration snapshots, all recorded in evidence section E8. Each ran once standalone on 2026-10-05 with real exit 0: install/publication assertions 18/18, config-refusal assertions 7/7, artifact checker 11/11 (10/10 sections, 38/38 links resolved in range, 10/10 JSON, 2/2 files, 0 pattern matches). Historical E5/E7 results stay labeled historical.

Two things for the next reviewer/operator. First, new finding F-R1-size: the mandated script bodies push the combined artifacts to 96,520 bytes over the historical 81,920 budget; E8.5 recommends accepting the one-time overrun. Second, the original internal-hostname sentinel value was never published, so the replacement publication check covers documented generic pattern families instead — stated as a bound in E8.5.

Changed files: only [.research/261004_CIP-0006-legacy-provider-settings-and-mcp-optouts_evidence.md](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261004-3v2zm2/worktree/.research/261004_CIP-0006-legacy-provider-settings-and-mcp-optouts_evidence.md) (E8 added); the CIP draft is untouched, as are product code and LOGBOOK.md. The evidence outcome resource, board notes, and reconciled checklist are updated; the handoff's outside-worktree warning covers only the board CLI's own bookkeeping paths.
agent completed: [analyst] researcher (muse) (exit=0)
spawn run completed: muse (run=RUN-261005-4b0ac5, pid=5228, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6.1-sol/medium","text":"R187 reviewer sol medium (cross-provider)"}
spawn selection rationale for gpt-6.1-sol/medium: R187 reviewer sol medium (cross-provider)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261005-6515e5, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261005-6515e5)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-6515e5, pid=16208, exit=0)
run write-boundary clearance for RUN-261004-4589ce: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261004-9bf8df: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261005-4b0ac5: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261005-6515e5: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
spawn selection rationale tuple: {"role":"researcher","pair":"gpt-6-astra/low","text":"bound 2iewnz-land (land queue); codex gpt-6-astra low"}
spawn selection rationale for gpt-6-astra/low: bound 2iewnz-land (land queue); codex gpt-6-astra low
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (codex) (run=RUN-261005-1bebb1, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-261005-1bebb1)
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-1bebb1, pid=32472, exit=0)

## Precondition Resources
- [r105-brief.md](file://TASK-261004-2iewnz/r105-brief.md)
- [cip-template.md](file://TASK-261004-2iewnz/cip-template.md)
- [research-review-note.md](file://TASK-261004-2iewnz/research-review-note.md)
- [r105-rework.md](file://TASK-261004-2iewnz/r105-rework.md)
- [host-rules.md](file://TASK-261004-2iewnz/host-rules.md)
- [2iewnz-integrate-land.md](file://TASK-261004-2iewnz/2iewnz-integrate-land.md)

## Outcome Resources
- [TASK-261004-2iewnz_spawn-log_-analyst--researcher--codex-_RUN-261004-4589ce.log](file://TASK-261004-2iewnz/TASK-261004-2iewnz_spawn-log_-analyst--researcher--codex-_RUN-261004-4589ce.log) — System spawn log captured by task-board
- [TASK-261004-2iewnz_CIP-0006-legacy-provider-settings-and-mcp-optouts.md](file://TASK-261004-2iewnz/TASK-261004-2iewnz_CIP-0006-legacy-provider-settings-and-mcp-optouts.md) — CIP-0006 draft: minimal profile settings and MCP opt-out design, migration, specification impact, implementation leaves, and tests; ready for review.
- [TASK-261004-2iewnz_CIP-0006-legacy-provider-settings-and-mcp-optouts_evidence.md](file://TASK-261004-2iewnz/TASK-261004-2iewnz_CIP-0006-legacy-provider-settings-and-mcp-optouts_evidence.md)
- [TASK-261004-2iewnz_spawn-log_-reviewer--reviewer--codex-_RUN-261004-9bf8df.log](file://TASK-261004-2iewnz/TASK-261004-2iewnz_spawn-log_-reviewer--reviewer--codex-_RUN-261004-9bf8df.log) — System spawn log captured by task-board
- [TASK-261004-2iewnz_review-verdict.md](file://TASK-261004-2iewnz/TASK-261004-2iewnz_review-verdict.md) — Changes requested: reproduce Python evidence checks; 14 source/spec citation groups checked
- [TASK-261004-2iewnz_spawn-log_-analyst--researcher--muse-_RUN-261005-4b0ac5.log](file://TASK-261004-2iewnz/TASK-261004-2iewnz_spawn-log_-analyst--researcher--muse-_RUN-261005-4b0ac5.log) — System spawn log captured by task-board
- [TASK-261004-2iewnz_change-request_rev1.patch](file://TASK-261004-2iewnz/TASK-261004-2iewnz_change-request_rev1.patch) — Change Request CR-TASK-261004-2iewnz-1 revision 1 candidate patch (repository_delta=present, 2 changed paths)
- [TASK-261004-2iewnz_change-request_rev1-validation.log](file://TASK-261004-2iewnz/TASK-261004-2iewnz_change-request_rev1-validation.log) — Change Request CR-TASK-261004-2iewnz-1 revision 1 bounded validation log
- [TASK-261004-2iewnz_spawn-log_-reviewer--reviewer--codex-_RUN-261005-6515e5.log](file://TASK-261004-2iewnz/TASK-261004-2iewnz_spawn-log_-reviewer--reviewer--codex-_RUN-261005-6515e5.log) — System spawn log captured by task-board
- [TASK-261004-2iewnz_review-verdict-rev1.md](file://TASK-261004-2iewnz/TASK-261004-2iewnz_review-verdict-rev1.md)
- [TASK-261004-2iewnz_spawn-log_-analyst--researcher--codex-_RUN-261005-1bebb1.log](file://TASK-261004-2iewnz/TASK-261004-2iewnz_spawn-log_-analyst--researcher--codex-_RUN-261005-1bebb1.log) — System spawn log captured by task-board
- [TASK-261004-2iewnz_integration-preconditions.md](file://TASK-261004-2iewnz/TASK-261004-2iewnz_integration-preconditions.md) — Fresh accepted-revision identity and landing preconditions for bound runner

## Created
2026-10-04T01:53:13Z

## Last Update
2026-10-05T05:34:21Z

## Assigned To
[analyst] researcher (codex)
