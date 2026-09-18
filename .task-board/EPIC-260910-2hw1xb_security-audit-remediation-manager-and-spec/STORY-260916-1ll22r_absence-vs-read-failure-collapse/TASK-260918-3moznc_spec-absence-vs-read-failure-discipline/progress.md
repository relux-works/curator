## Status
done

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(8))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] §8.4 general rule stated once; closed per-file-class table of absence vs unreadable outcomes with owning sections
- [x] Every read-site section (§1.3, §7.4, §8.2, §8.3, §9.5, §9.6, §10.1 + any found) references the rule; evidence lists each site with its verdict
- [x] New codes (if any) in §9.7 and every diagnostics list; closed sets stay closed
- [x] Vectors: unreadable-but-present for markers, seeds, locks, passthrough entries with a rule-7 validator gate; existing vectors byte-identical
- [x] make regenerate + validate + regeneration proof exit 0; CHANGELOG entry
- [x] TASK-260918-3moznc_spec-patch_rev1.patch = git diff HEAD (base recorded) with new files intent-to-added; EMPTY curator delta
- [x] Docs updated and consistent with current code
- [x] No discrepancies between code and description
- [x] Result linked as a new task-scoped outcome resource
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn selection rationale tuple: {"role":"doc-writer","pair":"muse-spark-1.3-contributor/max","text":"Normative curator-spec revision (closed tables/rules with rule-7 pinned vectors) for a wave-3 story; muse-spark-1.3-contributor:max is the operator's producer pair for this campaign; reviewer stays codex gpt-6-astra:low"}
spawn selection rationale for muse-spark-1.3-contributor/max: Normative curator-spec revision (closed tables/rules with rule-7 pinned vectors) for a wave-3 story; muse-spark-1.3-contributor:max is the operator's producer pair for this campaign; reviewer stays codex gpt-6-astra:low
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] doc-writer (muse) (run=RUN-260918-217db9, max_parallel=20)
spawn run started: [implementer] doc-writer (muse) (run=RUN-260918-217db9)
Design done; baselines green (validate.py exit 0, go test exit 0, unittest 538 pass). Inventory: 1 new code (environment_passthrough_unreadable); lock->store_untrusted; seed/marker/surface/inventory reuse existing; backup behavioral. Starting environments.md edits.
Key decisions: (1) lock read/parse failures join entry-class environment_store_untrusted — no environment_lock_invalid exists (only the unrelated mutation-lock timeout environment_lock_unavailable); no new lock code. (2) Single new code environment_passthrough_unreadable in owning table §7.7, NOT §9.7 (§9.7 lists only §9.x lifecycle diagnostics; inventories reuse loss-list+import_lossy) — rationale in evidence §1/§3. (3) Backup-record row is behavioral (fail closed + unknown status), no dedicated code exists and brief authorizes new codes only for seed/passthrough. (4) Found+filled pre-existing §10.4 gap: surface_unreadable at resolve was unlisted (resolve must report it, never stale+re-materialize). (5) New family file environments-read-failure.json (environments.json is generator-owned).
agent completed: [implementer] doc-writer (muse) (exit=0)
spawn run completed: muse (run=RUN-260918-217db9, pid=67219, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/low","text":"Round-1 independent review of the absence-vs-read-failure spec revision (closed table, references, rule-7 pinned vectors); codex gpt-6-astra:low is the operator's reviewer pair"}
spawn selection rationale for gpt-6-astra/low: Round-1 independent review of the absence-vs-read-failure spec revision (closed table, references, rule-7 pinned vectors); codex gpt-6-astra:low is the operator's reviewer pair
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-260918-7df759, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260918-7df759)
Review rev1 changes_requested; see TASK-260918-3moznc_review-verdict-rev1.md. Reconcile unreadable-lock entry-class rebuild with fail-closed rule; add required passthrough diagnostic in section 9.7; resolve unnamed backup unreadable diagnostic. Patch byte-identical, existing vectors 36/36 unchanged, focused tests 13/13 and absence probes 21/21 pass. Full reviewer suite interrupted; producer full-suite evidence not independently confirmed. Regeneration fixpoint identical. No LOGBOOK.md edit per campaign restriction; findings preserved in verdict and notes.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260918-7df759, pid=97077, exit=0)
spawn selection rationale tuple: {"role":"doc-writer","pair":"muse-spark-1.3-contributor/max","text":"Targeted rework of the absence-vs-read-failure spec revision (lock fail-closed precedence, §9.7 row, backup diagnostic) after changes_requested; muse-spark-1.3-contributor:max is the operator's producer pair"}
spawn selection rationale for muse-spark-1.3-contributor/max: Targeted rework of the absence-vs-read-failure spec revision (lock fail-closed precedence, §9.7 row, backup diagnostic) after changes_requested; muse-spark-1.3-contributor:max is the operator's producer pair
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [implementer] doc-writer (muse) (run=RUN-260918-109f6d, max_parallel=20)
spawn run started: [implementer] doc-writer (muse) (run=RUN-260918-109f6d)
Rev2 rework complete (F1/F2/F3 + migration note + coverage tightening). F1: lock no-rebuild precedence stated once in §8.4.1 lock row, referenced from §4/§1.3/§10.1/§10.4 + manager §12.5; repair+update no-rebuild vectors + rebuild-shaped negative. F2: §9.7 passthrough row (policy owned by §7.7). F3: new error-class environment_backup_record_unreadable in §8.4.1/§8.3/§8.5/§9.7/§12/§13 + manager mirrors, 5 backup vectors. Migration cites §8.4.1 inventory-candidate row, stops with environment_import_lossy (no consent flag on migration path). 39 cases, 18 focused tests, full suite 556 green, regenerate x2 exit 0 fixpoint-identical, 36 pre-existing vectors + 31 rev1 cases byte-identical, curator delta EMPTY. Evidence updated, patch rev2 attached (sha 7fff0900, base 5146c7b).
Item 14 verification for handoff gate: round-1 review did not accept (changes_requested); verdict evidence TASK-260918-3moznc_review-verdict-rev1.md is attached as a task outcome resource and status was routed to-dev per its explicit branches, then back to development for this rework run. Conditional satisfied by board state; checking item 14.
agent completed: [implementer] doc-writer (muse) (exit=0)
spawn run completed: muse (run=RUN-260918-109f6d, pid=10784, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"claude-opus-5/max","text":"Round-2 review of the absence-vs-read-failure spec revision (three targeted corrections over a conformant rev1); claude-opus-5:max is the operator's reviewer pair from 2026-09-18 (codex approaching usage limits)"}
spawn selection rationale for claude-opus-5/max: Round-2 review of the absence-vs-read-failure spec revision (three targeted corrections over a conformant rev1); claude-opus-5:max is the operator's reviewer pair from 2026-09-18 (codex approaching usage limits)
spawn agent resolution: Agent selection: claude via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (claude) (run=RUN-260918-41c0f7, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260918-41c0f7)
Review rev2 (RUN-260918-41c0f7, claude-opus-5): ACCEPTED via accept_cr revision=2 -> integrating. F1 (unreadable lock never rebuilt from; precedence in §8.4.1 lock row, referenced from §1.3/§4/§10.1/§10.4 + manager §12.5; repair/update no-rebuild vectors + rebuild-shaped negative), F2 (§9.7 passthrough row), F3 (environment_backup_record_unreadable, error-class, spelled identically in §8.3/§8.4.1/§8.5/§9.7/§12/§13 + manager §12.2/§12.7, 5 vectors) and the §9.4 migration note verified; rev1->rev2 interdiff traces only to F1-F3/§9.4. Patch == git diff HEAD (sha 7fff0900, base 5146c7b); 36/36 pre-existing vectors byte-identical; EMPTY curator delta is correct (spec lives in curator-spec). Independent gates on disposable clones: validate.py exit 0 (62 schemas/1118 vectors), go test exit 0, unittest 556 OK exit 0 (992 s), make regenerate-check exit 0 + double-regenerate fixpoint. Rule-7 replays: 135/135 in-memory and 31/31 through the real validate.py entry point refused. No worktree files touched. Evidence: TASK-260918-3moznc_review-verdict-rev2.md + _review-validation-rev2.log. Next: integration run bound to producer role doc-writer / archetype implementer.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260918-41c0f7, pid=84945, exit=0)

External integration evidence: curator-spec PR #73 landed on main as 23be89e94237d921d14ae325c6bc41ccfbe17103 (ff push after independent acceptance at revision 2 by claude-opus-5:max, TASK-260918-3moznc_review-verdict-rev2.md; clean apply of the accepted patch 16755b45 on its base 5146c7b; regenerate-check and validate exit 0; 9/9 checks green). Curator delta empty by design (spec-only task).

## Precondition Resources
- [TASK-260918-3moznc_brief.md](file://TASK-260918-3moznc/TASK-260918-3moznc_brief.md) — Producer brief
- [remediation-spec-producer-rules.md](file://TASK-260918-3moznc/remediation-spec-producer-rules.md) — Campaign rules for spec tasks (rule 7 pinning; patch = git diff HEAD)
- [TASK-260918-3moznc_review-brief.md](file://TASK-260918-3moznc/TASK-260918-3moznc_review-brief.md) — Reviewer brief, round 1
- [TASK-260918-3moznc_rework-rev2.md](file://TASK-260918-3moznc/TASK-260918-3moznc_rework-rev2.md) — Rework brief rev2: F1 unreadable-lock fail-closed precedence + vector, F2 §9.7 row, F3 backup-record diagnostic, §9.4 citation
- [TASK-260918-3moznc_review-brief-rev2.md](file://TASK-260918-3moznc/TASK-260918-3moznc_review-brief-rev2.md) — Reviewer brief, round 2 (F1 lock fail-closed precedence, F2 §9.7 row, F3 backup diagnostic)

## Outcome Resources
- [TASK-260918-3moznc_spawn-log_-implementer--doc-writer--muse-_RUN-260918-217db9.log](file://TASK-260918-3moznc/TASK-260918-3moznc_spawn-log_-implementer--doc-writer--muse-_RUN-260918-217db9.log) — System spawn log captured by task-board
- [TASK-260918-3moznc_evidence.md](file://TASK-260918-3moznc/TASK-260918-3moznc_evidence.md) — Evidence rev2: F1/F2/F3 rework, migration note, coverage tightening, full validation transcript
- [TASK-260918-3moznc_spec-patch_rev1.patch](file://TASK-260918-3moznc/TASK-260918-3moznc_spec-patch_rev1.patch) — Spec patch rev1: git diff HEAD (base 5146c7b9ed4b0c07b840ab58f9908f197d667478) with new vector file intent-to-added; curator delta EMPTY
- [TASK-260918-3moznc_change-request_rev1.patch](file://TASK-260918-3moznc/TASK-260918-3moznc_change-request_rev1.patch) — Change Request CR-TASK-260918-3moznc-1 revision 1 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260918-3moznc_change-request_rev1-validation.log](file://TASK-260918-3moznc/TASK-260918-3moznc_change-request_rev1-validation.log) — Change Request CR-TASK-260918-3moznc-1 revision 1 bounded validation log
- [TASK-260918-3moznc_spawn-log_-reviewer--reviewer--codex-_RUN-260918-7df759.log](file://TASK-260918-3moznc/TASK-260918-3moznc_spawn-log_-reviewer--reviewer--codex-_RUN-260918-7df759.log) — System spawn log captured by task-board
- [TASK-260918-3moznc_review-validation-rev1.log](file://TASK-260918-3moznc/TASK-260918-3moznc_review-validation-rev1.log) — Independent validate attempt: schema/vector gate passes; full unittest interrupted, with bounded reruns documented in verdict
- [TASK-260918-3moznc_review-verdict-rev1.md](file://TASK-260918-3moznc/TASK-260918-3moznc_review-verdict-rev1.md) — Changes requested: lock recovery contradiction, missing required diagnostic admission and backup diagnostic; independent validation and negative probes
- [TASK-260918-3moznc_spawn-log_-implementer--doc-writer--muse-_RUN-260918-109f6d.log](file://TASK-260918-3moznc/TASK-260918-3moznc_spawn-log_-implementer--doc-writer--muse-_RUN-260918-109f6d.log) — System spawn log captured by task-board
- [TASK-260918-3moznc_spec-patch_rev2.patch](file://TASK-260918-3moznc/TASK-260918-3moznc_spec-patch_rev2.patch) — Spec patch rev2: git diff HEAD (base 5146c7b9ed4b0c07b840ab58f9908f197d667478), 8 files, sha 7fff0900; curator delta EMPTY
- [TASK-260918-3moznc_change-request_rev2.patch](file://TASK-260918-3moznc/TASK-260918-3moznc_change-request_rev2.patch) — Change Request CR-TASK-260918-3moznc-2 revision 2 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260918-3moznc_change-request_rev2-validation.log](file://TASK-260918-3moznc/TASK-260918-3moznc_change-request_rev2-validation.log) — Change Request CR-TASK-260918-3moznc-2 revision 2 bounded validation log
- [TASK-260918-3moznc_spawn-log_-reviewer--reviewer--claude-_RUN-260918-41c0f7.log](file://TASK-260918-3moznc/TASK-260918-3moznc_spawn-log_-reviewer--reviewer--claude-_RUN-260918-41c0f7.log) — System spawn log captured by task-board
- [TASK-260918-3moznc_review-verdict-rev2.md](file://TASK-260918-3moznc/TASK-260918-3moznc_review-verdict-rev2.md) — Round-2 review verdict: accepted — F1 lock no-rebuild precedence, F2 §9.7 row, F3 backup-record diagnostic, §9.4 note verified; independent gates green; 135 in-memory + 31 through-main rule-7 replays refused
- [TASK-260918-3moznc_review-validation-rev2.log](file://TASK-260918-3moznc/TASK-260918-3moznc_review-validation-rev2.log) — Round-2 reviewer validation transcript: validate.py / go test / 556-test unittest (exit 0), regenerate-check + fixpoint (exit 0), rule-7 probe logs

## Created
2026-09-18T04:44:45Z

## Last Update
2026-09-18T17:25:37Z

## Assigned To
[reviewer] reviewer (claude)
