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
- TASK-260916-vht714

## Checklist
- [x] Current-state matrix per environment x platform with file:line citations and the e11-1 observations
- [x] Credential-mode options analysed with knob shape, strategies, security implications and migration
- [x] Permission-interface options analysed with verified native flag table, refusal rules and ax-profile interplay
- [x] Recommended designs with spec and implementation touchpoints; open questions listed; report attached and handed off
- [x] Findings written to file
- [x] Key aspects highlighted
- [x] Fact-checking performed — claims verified, sources cited
- [x] Findings linked on the board as a new task-scoped outcome resource
- [x] All questions from task description answered
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn selection rationale tuple: {"role":"researcher","pair":"gpt-6-astra/low","text":"operator asked for an Astra report on credential/permission modes; astra:low"}
spawn selection rationale for gpt-6-astra/low: operator asked for an Astra report on credential/permission modes; astra:low
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=main-7a0a24c; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [analyst] researcher (codex) (run=RUN-260916-6d44e9, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-260916-6d44e9)
Research outcomes attached. Existing isolation knob and shared-only lock direction confirmed. Findings: Pi native default .pi/agent differs from Curator .pi and both auth paths exist; shared-to-isolated may retain old auth link and repair may unlink detached credential (inspection, not live reproduction); tracked native bypass forbidden by current ax contract; Pi --approve is project trust, not permission bypass. Recommend existing credential knob plus safe migration, explicit permissions native|yolo with unsupported/tracked refusal. Four native help probes exit 0; artifact checks exit 0; clean worktree. B5 accepted, not rerun: Claude 1, Codex 0, Pi 1. LOGBOOK.md edits prohibited; this note and report serve as logbook-equivalent.
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-260916-6d44e9, pid=55700, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"claude-fable-5-1/low","text":"operator process: Astra report reviewed by Fable (claude-fable-5-1:low admitted fallback pair)"}
Story STORY-260916-1on1d2 stayed on base 18f05497f6ad4db243279de643535716bbea6ec6: 1 published Change Request revision(s) are still measured from it — CR-TASK-260916-2timlf-1 revision 1 (ready, element TASK-260916-2timlf, base 18f05497f6ad4db243279de643535716bbea6ec6). Carry them forward and the next spawn converges. carry the listed revision(s) forward: review one still awaiting a verdict, or task-board worktree checkpoint <ELEMENT-ID> an accepted one; when trunk has already advanced on paths the revision also changes, task-board worktree converge STORY-260916-1on1d2 is the sanctioned convergence; inspect with task-board worktree status STORY-260916-1on1d2, or task-board worktree abort STORY-260916-1on1d2
spawn selection rationale for claude-fable-5-1/low: operator process: Astra report reviewed by Fable (claude-fable-5-1:low admitted fallback pair)
spawn agent resolution: Agent selection: claude via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=main-7a0a24c; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260916-786774, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260916-786774)
Reviewer (Fable) verdict: ACCEPT with corrections, evidence TASK-260916-2timlf_review-verdict.md. 26 claims fact-checked against worktree 18f05497, curator-spec a68854d, launcher b34e1e2, relux-agents-infra-main 459742e and installed claude 2.1.273 / codex 0.153.4 / pi 0.84.2 --help; all verified. Corrections to append to the drafts: (1) NEW host fact: managed claude home holds .credentials.json (819 B, 2026-09-16 04:30) while the login Keychain has only the single unsuffixed Claude Code-credentials item (mdat 2026-08-17) -> the managed-home /login wrote a JSON file, not a suffixed Keychain item; environments.md:1119-1123 residual answered negatively, cause unknown, keep DiagSharedUnsupported until a disposable-account experiment; (2) Pi native root must be ~/.pi/agent (switch.go:78 says .pi); (3) both migration hazards confirmed by inspection (managed.go:500-502, 936, 1670-1677); (4) codex conflict checks must include -c approval_policy/sandbox_mode. Empty repository delta is correct for a research-only task; Tests green ticked as not applicable (no code).
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260916-786774, pid=74510, exit=0)

## Precondition Resources
- [campaign-producer-rules.md](file://TASK-260916-2timlf/campaign-producer-rules.md)
- [2timlf-brief.md](file://TASK-260916-2timlf/2timlf-brief.md)
- [2timlf-review-brief.md](file://TASK-260916-2timlf/2timlf-review-brief.md)

## Outcome Resources
- [TASK-260916-2timlf_spawn-log_-analyst--researcher--codex-_RUN-260916-6d44e9.log](file://TASK-260916-2timlf/TASK-260916-2timlf_spawn-log_-analyst--researcher--codex-_RUN-260916-6d44e9.log) — System spawn log captured by task-board
- [TASK-260916-2timlf_report.md](file://TASK-260916-2timlf/TASK-260916-2timlf_report.md) — Five-part research report: current state, credential modes, native permissions, migration risks and recommendations
- [TASK-260916-2timlf_native-help.txt](file://TASK-260916-2timlf/TASK-260916-2timlf_native-help.txt) — Installed Claude, Codex interactive/exec and Pi help; four standalone commands, exit 0
- [TASK-260916-2timlf_change-request_rev1.patch](file://TASK-260916-2timlf/TASK-260916-2timlf_change-request_rev1.patch) — Change Request CR-TASK-260916-2timlf-1 revision 1 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260916-2timlf_spawn-log_-reviewer--reviewer--claude-_RUN-260916-786774.log](file://TASK-260916-2timlf/TASK-260916-2timlf_spawn-log_-reviewer--reviewer--claude-_RUN-260916-786774.log) — System spawn log captured by task-board
- [TASK-260916-2timlf_review-verdict.md](file://TASK-260916-2timlf/TASK-260916-2timlf_review-verdict.md) — Fable independent review verdict: ACCEPT with corrections (fact-check table, host Keychain/JSON observation, final designs, decision-draft contents)

## Created
2026-09-16T00:57:43Z

## Last Update
2026-09-16T01:44:50Z

## Assigned To
[reviewer] reviewer (claude)
