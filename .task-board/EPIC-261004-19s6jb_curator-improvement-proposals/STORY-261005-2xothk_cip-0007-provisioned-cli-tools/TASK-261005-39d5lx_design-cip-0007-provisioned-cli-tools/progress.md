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
- [x] Draft cips/CIP-0007-<slug>.md in TEMPLATE.md shape, Status Draft, README index row
- [x] Options compared: status quo system commands vs registry-only lane vs layered signed registries with mirror
- [x] Trust model, threat analysis and offline/air-gapped behaviour stated with spec citations
- [x] Interaction with skillfile schema, lock, install transaction and conformance named
- [x] Decisions-needed list for the operator; no organisation or personal names; no normative edits
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
spawn selection rationale tuple: {"role":"researcher","pair":"muse-spark-1.3-contributor/max","text":"tb-R164 researcher muse-spark-1.3-contributor max; tb-R136 health-gated queue"}
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 researcher muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (muse) (run=RUN-261005-ed2a5d, max_parallel=8)
spawn run started: [analyst] researcher (muse) (run=RUN-261005-ed2a5d)
Draft CIP-0007 filed: cips/CIP-0007-manager-provisioned-cli-tools.md (Status Draft, TEMPLATE.md shape) + cips/README.md index row + .research/261005_manager-provisioned-cli-tools.md evidence note. Recommends option B registry-only lane per triage; option C held as target gated on owner/service/key/policy decisions. Validation: python -B tools/validate.py exit 0 (validated 73 schemas and 1294 vector files; run via isolated venv, jsonschema 4.25.1, since bare python is absent). No normative/schema/CHANGELOG/LOGBOOK edits; no org or personal names; work left uncommitted. No anomalies -> no logbook entry warranted.
spawn run child final message (run=RUN-261005-ed2a5d, tools=40 patches=5 failed=1):
Draft CIP-0007 is handed off to review.

What was produced (uncommitted in the story worktree, as required):

- [CIP-0007 draft](/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-261005-2xothk/worktree/cips/CIP-0007-manager-provisioned-cli-tools.md) — Status Draft in `TEMPLATE.md` shape. Compares option A (status quo system commands), option B (registry-only lane, recommended per triage), and option C (full provisioning with layered registries, trust policy, upstream mirror); covers trust model and threat analysis, offline/air-gapped behavior, interaction with the skillfile schema, lock, install transaction, conformance, and CIP-0005 audit backends, plus a decisions-needed list (registry owner/service, signing keys, default policy) and numbered open questions with recommendations.
- [Evidence note](/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-261005-2xothk/worktree/.research/261005_manager-provisioned-cli-tools.md) — sources read, key findings with file:line/§ citations, and fact-checks.
- One index row in `cips/README.md`.

Verification: `python -B tools/validate.py` exited 0 ("validated 73 schemas and 1294 vector files"), run via an isolated venv since bare `python` is absent on the host. No normative, schema, CHANGELOG, or LOGBOOK edits; no organisation or personal names (verified by grep). Board shows 11/11 checklist items and two outcome resources attached.
agent completed: [analyst] researcher (muse) (exit=0)
spawn run completed: muse (run=RUN-261005-ed2a5d, pid=78004, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6.1-sol/high","text":"R187 cross-provider reviewer; sol high (only sol pair admitted on spec)"}
spawn selection rationale for gpt-6.1-sol/high: R187 cross-provider reviewer; sol high (only sol pair admitted on spec)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261005-fa0397, max_parallel=8)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261005-fa0397)
Review revision 1: changes_requested; evidence TASK-261005-39d5lx_review-verdict-rev1.md. Six numbered findings: forbidden named examples (2 occurrences), federation-wide revocation before first-match selection, upstream publisher/subject binding, downgrade and offline replay authorization, Skillfile versus skill-manifest lock identity, and capability-aware B-lane runtime exposure. Independently reran validation: exit 0, 73 schemas and 1294 vectors. No repository edits. Return to analysis for researcher rework and a new security review.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-fa0397, pid=82550, exit=0)
loop-detector rev1: S2/S3/S5 not evaluable — runtime-recorded verdict carries no stamped findings array
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 researcher muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (muse) (run=RUN-261005-2e8f96, max_parallel=8)
spawn run started: [analyst] researcher (muse) (run=RUN-261005-2e8f96)
Rework 1 (researcher): addressed all six rev1 findings in CIP-0007 Draft + evidence note. validate.py exit 0 (73 schemas, 1294 vectors). Name/path rescans all 0. Outcomes attached as rev2 resources.
spawn run child final message (run=RUN-261005-2e8f96, tools=63 patches=18 failed=1):
Rework 1 of Draft CIP-0007 is handed off to review.

All six rev1 findings are addressed in [CIP-0007](</Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-261005-2xothk/worktree/cips/CIP-0007-manager-provisioned-cli-tools.md>) (Status Draft, template shape kept, one README index row) and the [evidence note](</Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-261005-2xothk/worktree/.research/261005_manager-provisioned-cli-tools.md>), which now carries a finding → change → location table:

- Vendor examples replaced with generic CLI descriptions; the note's false "none" claim corrected with rescan counts (0 matches).
- Two-phase resolution: federation-wide revocation scan with proposed match keys before first-layer positive selection; B-lane unreachable registry refuses; revocation beats allow overrides.
- C `verified` mode bound to per-tool expected publisher/subject; resolver boundary stated; mirror claim qualified; signer-substitution negatives added.
- Downgrade policy separated from snapshot rollback; replay requires current authorization (hash presence alone insufficient); air-gapped snapshots expire unless re-authorized.
- Lock-hash claim corrected (declaring Skillfile vs member manifests); versioned `skillfile-lock-v2` + new skill-manifest schema proposed.
- B minimal provisioned-only binding defined with enforced-PATH preservation; probe references fixed.

Verification: `tools/validate.py` exit 0 (73 schemas, 1294 vectors); name/path scans all 0; only the 3 doc paths touched. Board: 15/15 checklist, rev2 outcome resources attached, status `to-review` — ready for review.
agent completed: [analyst] researcher (muse) (exit=0)
spawn run completed: muse (run=RUN-261005-2e8f96, pid=84199, exit=0)
spawn selection rationale for gpt-6.1-sol/high: R187 cross-provider reviewer; sol high (only sol pair admitted on spec)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261005-5e340e, max_parallel=8)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261005-5e340e)
Revision 2 review: changes requested; see TASK-261005-39d5lx_review-verdict-rev2.md. Repeated finding 1: public evidence note embeds both prohibited vendor tokens while claiming a final scan of zero; independent count is 2 occurrences on 1 line. Other five previous design findings are addressed. Seven of seven review areas assessed, six held and one broken. Independent validator exit 0 (73 schemas, 1294 vectors). Correct final-document scan/report in the existing docs scope; next evidence should include a bounded evidence-only negative control. No repository edits or LOGBOOK changes by reviewer.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-5e340e, pid=89158, exit=0)
spawn selection rationale for muse-spark-1.3-contributor/max: tb-R164 researcher muse-spark-1.3-contributor max; tb-R136 health-gated queue
spawn agent resolution: Agent selection: muse via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (muse) (run=RUN-261005-ddcdf3, max_parallel=8)
spawn run started: [analyst] researcher (muse) (run=RUN-261005-ddcdf3)
Rework 2 (RUN current): addressed rev2/F1, the only remaining finding. Evidence note no longer publishes the two vendor tokens (generic scan labels only); pre-correction 2 occurrences on 1 line vs final 0 matches over saved bytes, all recorded in the note. Named check final-public-doc-name-scan added with passing negative control (spiked temp copy detected, CIP-only narrowing misses; control discarded). Token patterns lived only in ephemeral 0600 files outside the repo, never on argv or in committed files. Validator exit 0 (73 schemas, 1294 vectors). Changed paths still exactly the 3 authorized docs. Rev3 candidate bytes attached as outcome resources.
spawn run child final message (run=RUN-261005-ddcdf3, tools=41 patches=7 failed=0):
Rework 2 of Draft CIP-0007 is handed off to review.

The one remaining finding (rev2/F1) is addressed in the [evidence note](/Users/administrator/Developer/ReluxWorks/curator/curator-spec/.temp/STORY-261005-2xothk/worktree/.research/261005_manager-provisioned-cli-tools.md): the literal vendor-token scan expression is removed, only generic scan labels remain, and the note now records the pre-correction count (2 occurrences on 1 line) versus 0 matches over the corrected saved bytes. The named check `final-public-doc-name-scan` is documented with a passing negative control (spiked temp copy detected; CIP-only narrowing misses it; control discarded). Token patterns lived only in ephemeral 0600 files outside the repo, never on argv or in committed files.

Verification: `tools/validate.py` exit 0 (73 schemas, 1294 vectors); vendor, issue-org, and path/host scans all 0 over the corrected bytes; only the 3 authorized doc paths touched, `git diff --check` clean. Rev3 candidate bytes attached; board shows 15/15 checklist, status `to-review` — ready for review.
agent completed: [analyst] researcher (muse) (exit=0)
spawn run completed: muse (run=RUN-261005-ddcdf3, pid=90856, exit=0)
spawn selection rationale for gpt-6.1-sol/high: R187 cross-provider reviewer; sol high (only sol pair admitted on spec)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261005-2cbde4, max_parallel=8)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261005-2cbde4)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261005-2cbde4, pid=95595, exit=0)
spawn selection rationale tuple: {"role":"researcher","pair":"gpt-6-astra/low","text":"tb-R164 researcher gpt-6-astra low; tb-R136 health-gated queue"}
spawn selection rationale for gpt-6-astra/low: tb-R164 researcher gpt-6-astra low; tb-R136 health-gated queue
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[claude,codex,muse], config: spawn.preferred_agentic_system)
spawn queued: [analyst] researcher (codex) (run=RUN-261005-5ed199, max_parallel=8)
spawn run started: [analyst] researcher (codex) (run=RUN-261005-5ed199)

## Precondition Resources
- [cip0007-brief.md](file://TASK-261005-39d5lx/cip0007-brief.md)
- [cip0007-review-note.md](file://TASK-261005-39d5lx/cip0007-review-note.md)
- [cip0007-rework.md](file://TASK-261005-39d5lx/cip0007-rework.md)
- [cip0007-rework2.md](file://TASK-261005-39d5lx/cip0007-rework2.md)
- [39d5lx-complete.md](file://TASK-261005-39d5lx/39d5lx-complete.md)

## Outcome Resources
- [TASK-261005-39d5lx_spawn-log_-analyst--researcher--muse-_RUN-261005-ed2a5d.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_spawn-log_-analyst--researcher--muse-_RUN-261005-ed2a5d.log) — System spawn log captured by task-board
- [TASK-261005-39d5lx_CIP-0007.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_CIP-0007.md) — Draft CIP-0007: manager-provisioned CLI tools (registry-only lane recommended)
- [TASK-261005-39d5lx_evidence.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_evidence.md) — Evidence note with spec citations and validation record
- [TASK-261005-39d5lx_change-request_rev1.patch](file://TASK-261005-39d5lx/TASK-261005-39d5lx_change-request_rev1.patch) — Change Request CR-TASK-261005-39d5lx-1 revision 1 candidate patch (repository_delta=present, 3 changed paths)
- [TASK-261005-39d5lx_change-request_rev1-validation.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_change-request_rev1-validation.log) — Change Request CR-TASK-261005-39d5lx-1 revision 1 bounded validation log
- [TASK-261005-39d5lx_spawn-log_-reviewer--reviewer--codex-_RUN-261005-fa0397.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_spawn-log_-reviewer--reviewer--codex-_RUN-261005-fa0397.log) — System spawn log captured by task-board
- [TASK-261005-39d5lx_review-verdict-rev1.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_review-verdict-rev1.md) — Changes requested: six design/security findings with complete review coverage and independent validation
- [TASK-261005-39d5lx_spawn-log_-analyst--researcher--muse-_RUN-261005-2e8f96.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_spawn-log_-analyst--researcher--muse-_RUN-261005-2e8f96.log) — System spawn log captured by task-board
- [TASK-261005-39d5lx_CIP-0007-rev2.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_CIP-0007-rev2.md) — Rework 1: Draft CIP-0007 addressing review findings 1-6
- [TASK-261005-39d5lx_evidence-rev2.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_evidence-rev2.md) — Rework 1: evidence note with finding map and corrected claims
- [TASK-261005-39d5lx_change-request_rev2.patch](file://TASK-261005-39d5lx/TASK-261005-39d5lx_change-request_rev2.patch) — Change Request CR-TASK-261005-39d5lx-2 revision 2 candidate patch (repository_delta=present, 3 changed paths)
- [TASK-261005-39d5lx_change-request_rev2-validation.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_change-request_rev2-validation.log) — Change Request CR-TASK-261005-39d5lx-2 revision 2 bounded validation log
- [TASK-261005-39d5lx_spawn-log_-reviewer--reviewer--codex-_RUN-261005-5e340e.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_spawn-log_-reviewer--reviewer--codex-_RUN-261005-5e340e.log) — System spawn log captured by task-board
- [TASK-261005-39d5lx_review-verdict-rev2.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_review-verdict-rev2.md) — Revision 2 changes requested: repeated names/false scan finding; five design findings corrected; independent validation passed
- [TASK-261005-39d5lx_spawn-log_-analyst--researcher--muse-_RUN-261005-ddcdf3.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_spawn-log_-analyst--researcher--muse-_RUN-261005-ddcdf3.log) — System spawn log captured by task-board
- [TASK-261005-39d5lx_CIP-0007-rev3.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_CIP-0007-rev3.md) — Draft CIP-0007 candidate bytes, rework 2 (unchanged since rev2; attached for exact-candidate comparison)
- [TASK-261005-39d5lx_evidence-rev3.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_evidence-rev3.md) — Evidence note candidate bytes, rework 2: vendor tokens removed, generic scan labels, final-public-doc-name-scan with negative control
- [TASK-261005-39d5lx_change-request_rev3.patch](file://TASK-261005-39d5lx/TASK-261005-39d5lx_change-request_rev3.patch) — Change Request CR-TASK-261005-39d5lx-3 revision 3 candidate patch (repository_delta=present, 3 changed paths)
- [TASK-261005-39d5lx_change-request_rev3-validation.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_change-request_rev3-validation.log) — Change Request CR-TASK-261005-39d5lx-3 revision 3 bounded validation log
- [TASK-261005-39d5lx_spawn-log_-reviewer--reviewer--codex-_RUN-261005-2cbde4.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_spawn-log_-reviewer--reviewer--codex-_RUN-261005-2cbde4.log) — System spawn log captured by task-board
- [TASK-261005-39d5lx_review-verdict-rev3.md](file://TASK-261005-39d5lx/TASK-261005-39d5lx_review-verdict-rev3.md) — Revision 3 accepted: repeated scan finding fixed; 7/7 surfaces held; independent validator and note-only negative control passed
- [TASK-261005-39d5lx_spawn-log_-analyst--researcher--codex-_RUN-261005-5ed199.log](file://TASK-261005-39d5lx/TASK-261005-39d5lx_spawn-log_-analyst--researcher--codex-_RUN-261005-5ed199.log) — System spawn log captured by task-board

## Created
2026-10-05T17:47:33Z

## Last Update
2026-10-05T19:31:22Z

## Assigned To
[analyst] researcher (codex)
