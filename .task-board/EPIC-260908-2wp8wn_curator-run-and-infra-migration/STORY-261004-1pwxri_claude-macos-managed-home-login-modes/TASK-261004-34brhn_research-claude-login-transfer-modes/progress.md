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
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn selection rationale tuple: {"role":"researcher","pair":"gpt-6-astra/max","text":"tb-R164/R139 wave 1b 2026-10-04: researcher gpt-6-astra max"}
spawn selection rationale for gpt-6-astra/max: tb-R164/R139 wave 1b 2026-10-04: researcher gpt-6-astra max
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (codex) (run=RUN-261004-36bb33, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-261004-36bb33)
CIP-0003 research for TASK-261004-34brhn — research-claude-login-transfer-modes. Decision: define token/helper/per-home controls and retain a gated shared-file mode. Budget: 60 minutes; two outcome Markdown files under 96 KiB; one research prerequisite. Grammar frozen: not applicable to research; proposed closed schema will be enacted by the spec leaf. First production slice: curator-run protected-store token injection after trusted mode resolution, with a real child-process sentinel test. Current findings: Claude Code 2.1.287 has per-config Keychain selection with file fallback; synthetic status probes do not prove authentication or refresh. Q1 disposable-account gate remains open. LOGBOOK.md will remain untouched as explicitly required by the brief; findings go into research outcomes and this note.
TASK-261004-34brhn — research-claude-login-transfer-modes: CIP-0003 draft and evidence written under .research/. Packaging is blocked by the local command interface: isolated pwd/true/echo and artifact verification stopped returning results. Stalled commands were interrupted; verifier exit 130, not a pass. Draft resource-add exit 130 with no receipt, so attachment state is unknown and must be queried before add/update. No product code or LOGBOOK.md changes were requested or made. Need execution-host restoration, then artifact verification, task-scoped outcome attachment and researcher handoff.
Correction and recovery for TASK-261004-34brhn — research-claude-login-transfer-modes: command execution recovered. The delayed artifact verifier returned exit 0; the interim note incorrectly predicted exit 130 while cancellation raced completion. Actual checks: 12/12 template headings, 2/2 JSON examples, 15/15 probe rows, public-path/whitespace checks and exactly two uncommitted research files; nested git status exit 0. Standalone git diff --check exit 0. LOGBOOK.md untouched. Resuming attachment and researcher handoff; no host action is needed.
TASK-261004-34brhn — research-claude-login-transfer-modes: both task-scoped CIP-0003 outcome resources are now attached (draft and evidence), with successful CLI exit 0 receipts. Recommendation: token/helper/per-home controls; per-home default; shared-file remains Q1-gated. Q7 allows only a proposed narrow dedicated-token enrollment/injection boundary, never native credential copying. Artifact verifier rerun after correction exited 0; 51,905 bytes, exactly two uncommitted research files. Generic checklist item 8 is N/A under the binding no-LOGBOOK-edit brief: its check records that explicit exemption, with important findings persisted in outcomes and board notes instead. No operator credential was read or printed by the research probes.
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-261004-36bb33, pid=61054, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/medium","text":"tb-R164 reviewer astra medium; same-provider review (operator rule)"}
spawn selection rationale for gpt-6-astra/medium: tb-R164 reviewer astra medium; same-provider review (operator rule)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261004-9707e0, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261004-9707e0)
spawn run RUN-261004-9707e0 cancelled by operator; operator action required; reason: tb-R136 host cap: parked by orchestrator; resume later
agent completed: [reviewer] reviewer (codex) (exit=-1)
spawn run completed: codex (run=RUN-261004-9707e0, pid=24925, exit=-1)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/medium","text":"tb-R164 reviewer gpt-6-astra medium; tb-R136 health-gated queue"}
spawn selection rationale for gpt-6-astra/medium: tb-R164 reviewer gpt-6-astra medium; tb-R136 health-gated queue
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261004-9bb73f, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261004-9bb73f)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261004-9bb73f, pid=60732, exit=0)
loop-detector rev1: S2/S3/S5 not evaluable — runtime-recorded verdict carries no stamped findings array
spawn selection rationale tuple: {"role":"researcher","pair":"muse-spark-1.3-contributor/max","text":"tb-R164 researcher muse-spark-1.3-contributor max; tb-R136 health-gated queue"}
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 researcher muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (muse) (run=RUN-261005-92163f, max_parallel=20)
spawn run started: [analyst] researcher (muse) (run=RUN-261005-92163f)
R1 rework 2026-10-05: evidence extended with reproducibility appendix (exact bodies of reconstructed probe.py, matrix.py, security shim, fake-helper, probe.sb template, verify_artifacts.py; 6/6 byte-match executed files). Rerun R1-R15 on pinned bundle 2.1.287: 11 child exits 0, 4 expected-red exits 1; 11/11 matrix assertions PASS. Verifier 3 runs all exit 0. Combined size 84970/98304. Git scope exactly the two .research files; CIP draft byte-identical; LOGBOOK untouched; no code changes; synthetic fixtures only.
Checklist 9-11 rationale (2026-10-05): (9) AC verified clause by clause - CIP draft 12/12 headings, 4 options plus recommendation plus 6 open questions, pinned citations plus reproducible R1-R15 probes; git scope exactly the two .research files; synthetic fixtures only; LOGBOOK untouched. (10) Design unchanged by rework; reviewer rev1 architecture Pass stands (Q7 boundary, post-serialization injection, v4 reservations, named Go call sites). (11) Applicable checks green: 11/11 matrix assertions PASS, 4/4 probe wrappers expected exits (R3 exit 1 is the expected-red no-login row), verifier 3 runs exit 0. No product suite exists or was requested; go build/test not run per hosted-evidence mode.
spawn run child final message (run=RUN-261005-92163f, tools=125 patches=24 failed=0):
R1 rework handed off to review: TASK-261004-34brhn — research-claude-login-transfer-modes is now `to-review`, checklist 12/12, handoff exit 0.

What changed: only [.research/261004_CIP-0003-claude-managed-home-credential-modes_evidence.md](/Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-261004-1pwxri/worktree/.research/261004_CIP-0003-claude-managed-home-credential-modes_evidence.md) gained an R1 reproducibility appendix (17.8 → 50.9 KB; combined 84,970 ≤ 98,304 cap). It embeds the exact bodies of the reconstructed probe.py, matrix.py, Keychain shim, fake helper, sandbox template, and artifact verifier; per-script pinned inputs and expected counts; the run recipe; a 15-row rerun ledger (R1–R15, all key fields agreeing with the historical P-rows, which stay labelled historical); a verbatim R8 envelope; and bring-up corrections. The CIP draft is byte-identical; no product code changed; LOGBOOK.md untouched; all fixtures synthetic.

Verification observed this session on the pinned 2.1.287 bundle (SHA-256 matched before every run): 11/11 matrix assertions PASS, 4/4 probe wrappers returned expected exits (11 child exits 0, 4 expected-red exits 1), embedded bodies 6/6 byte-match the executed files, artifact verifier 3 runs all exit 0 with the final quote self-consistent, `git status` shows exactly the two research files and `git diff --check` exits 0. Board evidence resource updated (exit 0). No real credential read or printed; no login/logout; no `go build/test` per hosted-evidence mode.
agent completed: [analyst] researcher (muse) (exit=0)
spawn run completed: muse (run=RUN-261005-92163f, pid=23053, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6.1-sol/medium","text":"R187 reviewer sol medium (cross-provider)"}
spawn selection rationale for gpt-6.1-sol/medium: R187 reviewer sol medium (cross-provider)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261005-e37a26, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261005-e37a26)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-e37a26, pid=19126, exit=0)
run write-boundary clearance for RUN-261004-36bb33: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261004-9707e0: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261004-9bb73f: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261005-92163f: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
run write-boundary clearance for RUN-261005-e37a26: Operator review 2026-09-23: every flagged path is board state under .task-board/ (activity streams, journals and element files of other elements) or inside OTHER Stories' managed worktrees under .temp/STORY-*/worktree, written concurrently by the orchestrator's own board mutations and by other tracked runs during this run's lifetime; no source path outside the Story worktree was written.
spawn selection rationale tuple: {"role":"researcher","pair":"gpt-6-astra/low","text":"bound 34brhn-land (land queue); codex gpt-6-astra low"}
spawn selection rationale for gpt-6-astra/low: bound 34brhn-land (land queue); codex gpt-6-astra low
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (codex) (run=RUN-261005-38ed02, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-261005-38ed02)
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-38ed02, pid=42345, exit=0)

## Precondition Resources
- [login-research-brief.md](file://TASK-261004-34brhn/login-research-brief.md)
- [cip-template.md](file://TASK-261004-34brhn/cip-template.md)
- [research-review-note.md](file://TASK-261004-34brhn/research-review-note.md)
- [login-rework.md](file://TASK-261004-34brhn/login-rework.md)
- [host-rules.md](file://TASK-261004-34brhn/host-rules.md)
- [34brhn-integrate-land.md](file://TASK-261004-34brhn/34brhn-integrate-land.md)

## Outcome Resources
- [TASK-261004-34brhn_spawn-log_-analyst--researcher--codex-_RUN-261004-36bb33.log](file://TASK-261004-34brhn/TASK-261004-34brhn_spawn-log_-analyst--researcher--codex-_RUN-261004-36bb33.log) — System spawn log captured by task-board
- [TASK-261004-34brhn_CIP-0003-draft.md](file://TASK-261004-34brhn/TASK-261004-34brhn_CIP-0003-draft.md) — CIP-0003 draft: token, helper, per-home and gated shared-file modes; Q1 and Q7 amendments
- [TASK-261004-34brhn_CIP-0003-evidence.md](file://TASK-261004-34brhn/TASK-261004-34brhn_CIP-0003-evidence.md) — R1 rework: reconstructed harness bodies, rerun ledger R1-R15, verifier exit 0 (2026-10-05)
- [TASK-261004-34brhn_change-request_rev1.patch](file://TASK-261004-34brhn/TASK-261004-34brhn_change-request_rev1.patch) — Change Request CR-TASK-261004-34brhn-1 revision 1 candidate patch (repository_delta=present, 2 changed paths)
- [TASK-261004-34brhn_change-request_rev1-validation.log](file://TASK-261004-34brhn/TASK-261004-34brhn_change-request_rev1-validation.log) — Change Request CR-TASK-261004-34brhn-1 revision 1 bounded validation log
- [TASK-261004-34brhn_spawn-log_-reviewer--reviewer--codex-_RUN-261004-9707e0.log](file://TASK-261004-34brhn/TASK-261004-34brhn_spawn-log_-reviewer--reviewer--codex-_RUN-261004-9707e0.log) — System spawn log captured by task-board
- [TASK-261004-34brhn_spawn-log_-reviewer--reviewer--codex-_RUN-261004-9bb73f.log](file://TASK-261004-34brhn/TASK-261004-34brhn_spawn-log_-reviewer--reviewer--codex-_RUN-261004-9bb73f.log) — System spawn log captured by task-board
- [TASK-261004-34brhn_review-verdict-rev1.md](file://TASK-261004-34brhn/TASK-261004-34brhn_review-verdict-rev1.md) — Review revision 1: changes requested for reproducible probe evidence
- [TASK-261004-34brhn_spawn-log_-analyst--researcher--muse-_RUN-261005-92163f.log](file://TASK-261004-34brhn/TASK-261004-34brhn_spawn-log_-analyst--researcher--muse-_RUN-261005-92163f.log) — System spawn log captured by task-board
- [TASK-261004-34brhn_change-request_rev2.patch](file://TASK-261004-34brhn/TASK-261004-34brhn_change-request_rev2.patch) — Change Request CR-TASK-261004-34brhn-2 revision 2 candidate patch (repository_delta=present, 2 changed paths)
- [TASK-261004-34brhn_change-request_rev2-validation.log](file://TASK-261004-34brhn/TASK-261004-34brhn_change-request_rev2-validation.log) — Change Request CR-TASK-261004-34brhn-2 revision 2 bounded validation log
- [TASK-261004-34brhn_spawn-log_-reviewer--reviewer--codex-_RUN-261005-e37a26.log](file://TASK-261004-34brhn/TASK-261004-34brhn_spawn-log_-reviewer--reviewer--codex-_RUN-261005-e37a26.log) — System spawn log captured by task-board
- [TASK-261004-34brhn_review-verdict-rev2.md](file://TASK-261004-34brhn/TASK-261004-34brhn_review-verdict-rev2.md) — Revision 2 review: R1 reproducibility resolved, sixteen citation groups checked, artifact checker exit 0; accepted for producer integration
- [TASK-261004-34brhn_spawn-log_-analyst--researcher--codex-_RUN-261005-38ed02.log](file://TASK-261004-34brhn/TASK-261004-34brhn_spawn-log_-analyst--researcher--codex-_RUN-261005-38ed02.log) — System spawn log captured by task-board
- [TASK-261004-34brhn_integration-preconditions.md](file://TASK-261004-34brhn/TASK-261004-34brhn_integration-preconditions.md) — Fresh revision 2 integration precondition observations; authoritative landing delegated to bound runner

## Created
2026-10-04T01:52:41Z

## Last Update
2026-10-05T05:45:07Z

## Assigned To
[analyst] researcher (codex)
