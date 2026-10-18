## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- BUG-261017-8v470p

## Blocks
- (none)

## Checklist
- [x] Audit classifies loopback and last labels test, example, invalid, localhost as non-network in every audit path (member, migration, cache)
- [x] Lookalikes, bare labels and example.com/.net/.org stay contacts; canary still blocks; mixed file still reports the real host
- [x] All 40 curator-spec network-contact vectors mapped to executable Go tests
- [x] Named mutants go red for their stated reason; cached findings cannot mask the new classification
- [x] Full local suite and full hosted gate green on the exact signed candidate
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] In a managed Story worktree the candidate is left UNCOMMITTED in the worktree for the handoff to snapshot — never commit on the Story branch. A producer commit moves the branch tip off the recorded checkpoint and the handoff refuses with change_request_candidate_committed_past_checkpoint; repair with `git reset --soft <checkpoint_oid>` before completing again.
- [x] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [x] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [x] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [x] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Gate, refusal, validation, authorization, and attestation behavior attacked, not read — positive-path-only evidence is not accepted
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Blocked on an owner decision: approve treating RFC 6761 special-use TLD hosts (.test, .example, .invalid, .localhost) as non-network in the static audit, with example.com/net/org still flagged and the canary unchanged. Recommendation: approve; the hosts never resolve publicly and .localhost is loopback by RFC. Interim workaround for the consumer, no contract change: use a localhost URL in the fixture.
The requesting consumer adopted the localhost workaround and no longer needs the exemption. Kept for an owner decision at low priority; close if no other consumer needs it.
Owner approved 2026-10-17 about 15:10Z: treat hosts under the special-use TLDs .test, .example, .invalid and .localhost as non-network in the static audit, exactly in the requested shape; example.com, example.net and example.org stay flagged so the static canary keeps firing; lookalikes such as x.test.evil.com and a bare test stay flagged; a mixed file still reports the real host. Order: curator-spec contract text first, then this code change; it follows BUG-261017-8v470p because both touch internal/audit.
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6.1-sol/high","text":"Wave 7 TLDCODE producer: docs-only, codex gpt-6.1-sol high (R80), owner go 2026-10-13"}
spawn selection rationale for gpt-6.1-sol/high: Wave 7 TLDCODE producer: docs-only, codex gpt-6.1-sol high (R80), owner go 2026-10-13
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261017-28b76f, max_parallel=4)
spawn run started: [implementer] developer (codex) (run=RUN-261017-28b76f)
Ready for developer handoff. Signed detached candidate 528618358696a3d498ac456667ad19b87f6fc8bd matches the uncommitted Story tree; checkpoint unchanged. All 40 vectors cover Gate/GateReadOnly, Install and migration (120 terminals); six narrowing mutants killed. Full local: 111 commands, 85 packages, 3445 pass / 31 skip top-level; all static gates green. Hosted run 39077229686: 20 success, 2 conditional skips, controller exit 0. AC coverage is 4 of 5 rows driven through product entry points; row 5 is validation-only. Outcome records the missing surface-table brief gap, bounds, full mapping, original diagnostic failures and their passing replays. Seven sanitized plaintext evidence bundles and their digest index are attached. No Story branch commit; integration belongs to the parent.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-261017-28b76f, pid=82968, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/medium","text":"Deciding review B on codex gpt-6-astra medium: owner spawn policy 2026-10-16 (Codex only; reviewers astra medium)"}
spawn selection rationale for gpt-6-astra/medium: Deciding review B on codex gpt-6-astra medium: owner spawn policy 2026-10-16 (Codex only; reviewers astra medium)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261018-99e225, max_parallel=4)
spawn run RUN-261018-99e225 failed; operator action required; failure: queued spawn preparation failed: revision_base_superseded: protected trunk advanced on Change Request TASK-261017-sxydvh revision 1 paths that the candidate changes (README.md); reviewer spawn refused (element_id=TASK-261017-sxydvh, overlapping_paths=README.md, protected_authority_oid=402fdbc9b77803cb2d6b11f4f1cfd83961e6897b, refusal_reason=overlap, remedy=converge the Story workspace onto fresh protected authority, then retry reviewer spawn, remedy_command=task-board worktree converge STORY-261017-7eqh0t --reason "trunk advanced on changed candidate paths", revision=1)
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6.1-sol/high","text":"TLDCODE round 2: combination check on trunk 402fdbc9 (codex gpt-6.1-sol high per owner spawn policy)"}
spawn selection rationale for gpt-6.1-sol/high: TLDCODE round 2: combination check on trunk 402fdbc9 (codex gpt-6.1-sol high per owner spawn policy)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261018-08b11a, max_parallel=4)
spawn run started: [implementer] developer (codex) (run=RUN-261018-08b11a)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-261018-08b11a, pid=90023, exit=0)
spawn selection rationale for gpt-6-astra/medium: Deciding review B on codex gpt-6-astra medium: owner spawn policy 2026-10-16 (Codex only; reviewers astra medium)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261018-1d57da, max_parallel=4)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261018-1d57da)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261018-1d57da, pid=19240, exit=0)

## Precondition Resources
- [gw-TLDCODE-brief.md](file://TASK-261017-sxydvh/gw-TLDCODE-brief.md)
- [gw-TLDCODE-review-B-brief.md](file://TASK-261017-sxydvh/gw-TLDCODE-review-B-brief.md)
- [gw-TLDCODE-r2-brief.md](file://TASK-261017-sxydvh/gw-TLDCODE-r2-brief.md)

## Outcome Resources
- [TASK-261017-sxydvh_spawn-log_-implementer--developer--codex-_RUN-261017-28b76f.log](file://TASK-261017-sxydvh/TASK-261017-sxydvh_spawn-log_-implementer--developer--codex-_RUN-261017-28b76f.log) — System spawn log captured by task-board
- [TASK-261017-sxydvh_outcome.md](file://TASK-261017-sxydvh/TASK-261017-sxydvh_outcome.md) — Ready for review after refresh onto 402fdbc9: README combination, full local and hosted green, exact signed candidate and complete evidence
- [TASK-261017-sxydvh_evidence_network.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_evidence_network.jsonl) — Corrected candidate 5286183: full audit and profile vector logs, cache controls and six narrowing mutants
- [TASK-261017-sxydvh_evidence_cli.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_evidence_cli.jsonl) — Corrected signed candidate 5286183: complete local cli logs and command receipts
- [TASK-261017-sxydvh_evidence_install.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_evidence_install.jsonl) — Corrected signed candidate 5286183: complete local install logs and command receipts
- [TASK-261017-sxydvh_evidence_godriver.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_evidence_godriver.jsonl) — Corrected signed candidate 5286183: complete local godriver logs and command receipts
- [TASK-261017-sxydvh_evidence_envprofile.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_evidence_envprofile.jsonl) — Corrected signed candidate 5286183: complete local envprofile logs and command receipts
- [TASK-261017-sxydvh_evidence_packages.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_evidence_packages.jsonl) — Corrected signed candidate 5286183: complete local packages logs and command receipts
- [TASK-261017-sxydvh_authority_regression.md](file://TASK-261017-sxydvh/TASK-261017-sxydvh_authority_regression.md) — Authority/path regression: original red probe, six killed mutants, corrected full local and hosted green results
- [TASK-261017-sxydvh_evidence_validation.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_evidence_validation.jsonl) — Exact candidate: hosted provider verdict, static gates, complete inventory, identity checks, redacted diagnostics and evidence controls
- [TASK-261017-sxydvh_evidence_index.json](file://TASK-261017-sxydvh/TASK-261017-sxydvh_evidence_index.json) — Seven sanitized plaintext bundle digests and full 51582-event preservation verification
- [TASK-261017-sxydvh_change-request_rev1.patch](file://TASK-261017-sxydvh/TASK-261017-sxydvh_change-request_rev1.patch) — Change Request CR-TASK-261017-sxydvh-1 revision 1 candidate patch (repository_delta=present, 11 changed paths)
- [TASK-261017-sxydvh_change-request_rev1-validation.log](file://TASK-261017-sxydvh/TASK-261017-sxydvh_change-request_rev1-validation.log) — Change Request CR-TASK-261017-sxydvh-1 revision 1 bounded validation log
- [TASK-261017-sxydvh_spawn-log_-reviewer--reviewer--codex-_RUN-261018-99e225.log](file://TASK-261017-sxydvh/TASK-261017-sxydvh_spawn-log_-reviewer--reviewer--codex-_RUN-261018-99e225.log) — System spawn log captured by task-board
- [TASK-261017-sxydvh_spawn-log_-implementer--developer--codex-_RUN-261018-08b11a.log](file://TASK-261017-sxydvh/TASK-261017-sxydvh_spawn-log_-implementer--developer--codex-_RUN-261018-08b11a.log) — System spawn log captured by task-board
- [TASK-261017-sxydvh_refresh-logbook.md](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh-logbook.md) — Refresh onto 402fdbc9: README-only rework, all local and hosted validation green, diagnostic retained
- [TASK-261017-sxydvh_refresh_network.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh_network.jsonl) — Refresh onto 402fdbc9: fresh network logs and standalone command receipts; signed candidate 5ceef5a6
- [TASK-261017-sxydvh_refresh_cli.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh_cli.jsonl) — Refresh onto 402fdbc9: fresh cli logs and standalone command receipts; signed candidate 5ceef5a6
- [TASK-261017-sxydvh_refresh_install.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh_install.jsonl) — Refresh onto 402fdbc9: fresh install logs and standalone command receipts; signed candidate 5ceef5a6
- [TASK-261017-sxydvh_refresh_godriver.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh_godriver.jsonl) — Refresh onto 402fdbc9: fresh godriver logs and standalone command receipts; signed candidate 5ceef5a6
- [TASK-261017-sxydvh_refresh_envprofile.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh_envprofile.jsonl) — Refresh onto 402fdbc9: fresh envprofile logs and standalone command receipts; signed candidate 5ceef5a6
- [TASK-261017-sxydvh_refresh_packages.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh_packages.jsonl) — Refresh onto 402fdbc9: fresh packages logs and standalone command receipts; signed candidate 5ceef5a6
- [TASK-261017-sxydvh_refresh_validation.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh_validation.jsonl) — Refresh onto 402fdbc9: all validation receipts, complete hosted verdict, source identity, coverage and evidence rejection controls
- [TASK-261017-sxydvh_refresh_index.json](file://TASK-261017-sxydvh/TASK-261017-sxydvh_refresh_index.json) — Refresh evidence digests; all 54256 accepted Go events preserved in seven sanitized bundles
- [TASK-261017-sxydvh_change-request_rev2.patch](file://TASK-261017-sxydvh/TASK-261017-sxydvh_change-request_rev2.patch) — Change Request CR-TASK-261017-sxydvh-2 revision 2 candidate patch (repository_delta=present, 11 changed paths)
- [TASK-261017-sxydvh_change-request_rev2-validation.log](file://TASK-261017-sxydvh/TASK-261017-sxydvh_change-request_rev2-validation.log) — Change Request CR-TASK-261017-sxydvh-2 revision 2 bounded validation log
- [TASK-261017-sxydvh_spawn-log_-reviewer--reviewer--codex-_RUN-261018-1d57da.log](file://TASK-261017-sxydvh/TASK-261017-sxydvh_spawn-log_-reviewer--reviewer--codex-_RUN-261018-1d57da.log) — System spawn log captured by task-board
- [TASK-261017-sxydvh_review-evidence-rev2.jsonl](file://TASK-261017-sxydvh/TASK-261017-sxydvh_review-evidence-rev2.jsonl) — Independent revision-2 surface sweep, behavioral mutants, free-hunt probes and provider evidence
- [TASK-261017-sxydvh_review-verdict-rev2.md](file://TASK-261017-sxydvh/TASK-261017-sxydvh_review-verdict-rev2.md) — Accepted revision 2: ten held surfaces, empty findings array, exact-tree validation and logbook
- [TASK-261017-sxydvh_review-acceptance-note-rev2.md](file://TASK-261017-sxydvh/TASK-261017-sxydvh_review-acceptance-note-rev2.md) — Acceptance readback and write-boundary attribution warning

## Created
2026-10-17T13:29:54Z

## Last Update
2026-10-18T13:52:59Z

## Assigned To
[reviewer] reviewer (codex)
