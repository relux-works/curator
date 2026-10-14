## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(2))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Both migration audit call sites run under hashing.WriteVersion(); no path exemption
- [x] Migrated lock records hash version 2 under a v2 write version
- [x] Hermetic tests: gzip+binary skill migrates under v2; forced v1 still refuses with opaque-file
- [x] network-undeclared follow-up tested and recorded in the gaps file
- [x] Gaps file BUG-261014-351fcr_gaps.md attached; remote-gate result cited
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
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6.1-sol/high","text":"Wave 7 BF1 producer: docs-only, codex gpt-6.1-sol high (R80), owner go 2026-10-13"}
spawn selection rationale for gpt-6.1-sol/high: Wave 7 BF1 producer: docs-only, codex gpt-6.1-sol high (R80), owner go 2026-10-13
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [implementer] developer (codex) (run=RUN-261014-b4034a, max_parallel=4)
spawn run started: [implementer] developer (codex) (run=RUN-261014-b4034a)
BF1 development at base 84cf54b94670ae9881f91b1c147feb4b75afddee, matching freshly fetched origin/main. Scope: two migration version selections, hermetic gzip/binary and forced-v1 tests, URL follow-up characterization, focused local checks and hosted remote gate. Existing branch declarations remain outside migrateGlobalSkills; fixture pins legacy branch to revision before migration. No surface table supplied: handoff will declare that brief gap and map all four AC rows. Candidate remains uncommitted in the managed Story worktree.
BF1 candidate changes only two audit-version selections plus hermetic migration tests. Local go build ./..., go vet ./..., and full go test ./internal/envprofile/... -count=1 -v exited 0 (package 214.281s); conformance-root/platform skips declared in gaps. Four mutants all exited 1 with named failing regressions; originals restored. Confirmed separate-leaf follow-up: TestMigratedGlobalSkillNetworkUndeclaredFollowUp drives EnsureDefaultWithPolicy and gets high audit.capability.network-undeclared at scripts/fetch.sh because strictAuditMember uses capabilities.ImplicitNone(). Hosted snapshot f34a0db8be5b13282f7bb4742b5d8a21ead874cf, run https://github.com/relux-works/curator/actions/runs/38517800551 still running; no handoff until terminal evidence is attached.
BF1 ready for review at unchanged managed checkpoint 84cf54b94670ae9881f91b1c147feb4b75afddee. Both migration audit selections now use hashing.WriteVersion(); v2 locks record version 2 with binary/gzip bytes preserved; forced v1 retains opaque-file refusal without any directory exemption. AC coverage: 4 of 4 rows driven through EnsureDefaultWithPolicy -> migrateGlobalSkills. Five mutants were killed (two independent V1 audit regressions, directory exemption, exact gzip-file exemption, exact undeclared-host exemption), each with a named failing test and real exit 1. Local build, vet, full envprofile suite, formatting and diff checks exit 0. Hosted gate https://github.com/relux-works/curator/actions/runs/38517800551 passed, script exit 0, snapshot f34a0db8be5b13282f7bb4742b5d8a21ead874cf; 20 applicable jobs green, two workflow-conditional jobs explicitly skipped. Gaps, raw mutant evidence and raw remote result attached. Follow-up for separate leaf: strictAuditMember uses capabilities.ImplicitNone(); TestMigratedGlobalSkillNetworkUndeclaredFollowUp confirms high audit.capability.network-undeclared at scripts/fetch.sh. No source-text checker added; host mutant preserves detector tokens and fails the behavioral suite. No control-root logbook edit; the task-scoped gaps outcome carries the logbook entry. No surface table supplied (brief gap declared), and live-branch source selection remains outside the audit-only BF1 scope. Candidate remains uncommitted, existing reviewer tests untouched.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-261014-b4034a, pid=93135, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/medium","text":"R138 reviewer A, record-only (codex gpt-6-astra medium)"}
spawn selection rationale for gpt-6-astra/medium: R138 reviewer A, record-only (codex gpt-6-astra medium)
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (codex) (run=RUN-261014-06fa0f, max_parallel=4)
spawn run started: [reviewer] reviewer (codex) (run=RUN-261014-06fa0f)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-261014-06fa0f, pid=26432, exit=0)
spawn autonomous recovery: run RUN-261014-06fa0f queued successor RUN-261014-6f656f (attempt 1/3, model=gpt-6-astra): reviewer run RUN-261014-06fa0f remains unsatisfied: reviewer run has no verdict branch while BUG-261014-351fcr is reviewing
spawn run started: [reviewer] reviewer (codex) (run=RUN-261014-6f656f)
spawn run RUN-261014-6f656f cancelled by operator; operator action required; reason: R225 addendum: auto-recovery successor of record-only reviewer A RUN-261014-06fa0f
agent completed: [reviewer] reviewer (codex) (exit=-1)
spawn run completed: codex (run=RUN-261014-6f656f, pid=71923, exit=-1)
spawn selection rationale tuple: {"role":"reviewer","pair":"claude-sonnet-5-5/high","text":"R138 reviewer B, cross-provider, deciding review"}
spawn selection rationale for claude-sonnet-5-5/high: R138 reviewer B, cross-provider, deciding review
spawn agent resolution: Agent selection: claude via explicit_override (preferred_agentic_system: mixed[codex,claude], config: spawn.preferred_agentic_system)
spawn queued: [reviewer] reviewer (claude) (run=RUN-261014-52a5b6, max_parallel=4)
spawn run started: [reviewer] reviewer (claude) (run=RUN-261014-52a5b6)
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-261014-52a5b6, pid=90088, exit=0)

## Precondition Resources
- [gw-BF1-brief.md](file://BUG-261014-351fcr/gw-BF1-brief.md)
- [gw-BF1-review-A-brief.md](file://BUG-261014-351fcr/gw-BF1-review-A-brief.md)
- [gw-BF1-review-B-brief.md](file://BUG-261014-351fcr/gw-BF1-review-B-brief.md)

## Outcome Resources
- [BUG-261014-351fcr_spawn-log_-implementer--developer--codex-_RUN-261014-b4034a.log](file://BUG-261014-351fcr/BUG-261014-351fcr_spawn-log_-implementer--developer--codex-_RUN-261014-b4034a.log) — System spawn log captured by task-board
- [BUG-261014-351fcr_gaps.md](file://BUG-261014-351fcr/BUG-261014-351fcr_gaps.md) — BF1 ready-for-review evidence: four AC rows, five killed mutants, confirmed capability follow-up, local checks and hosted gate exit 0.
- [BUG-261014-351fcr_test-evidence.txt](file://BUG-261014-351fcr/BUG-261014-351fcr_test-evidence.txt) — Raw local regression and five expected-red mutant results with observed exit codes and log digests.
- [BUG-261014-351fcr_remote-gate.log](file://BUG-261014-351fcr/BUG-261014-351fcr_remote-gate.log) — Hosted landing gate raw terminal result: run 38517800551 success, script exit 0, exact candidate snapshot.
- [BUG-261014-351fcr_change-request_rev1.patch](file://BUG-261014-351fcr/BUG-261014-351fcr_change-request_rev1.patch) — Change Request CR-BUG-261014-351fcr-1 revision 1 candidate patch (repository_delta=present, 2 changed paths)
- [BUG-261014-351fcr_change-request_rev1-validation.log](file://BUG-261014-351fcr/BUG-261014-351fcr_change-request_rev1-validation.log) — Change Request CR-BUG-261014-351fcr-1 revision 1 bounded validation log
- [BUG-261014-351fcr_spawn-log_-reviewer--reviewer--codex-_RUN-261014-06fa0f.log](file://BUG-261014-351fcr/BUG-261014-351fcr_spawn-log_-reviewer--reviewer--codex-_RUN-261014-06fa0f.log) — System spawn log captured by task-board
- [BUG-261014-351fcr_review-A-evidence.txt](file://BUG-261014-351fcr/BUG-261014-351fcr_review-A-evidence.txt) — Reviewer A exact-tree baseline, five killed mutants, tag free hunt and independently read hosted result.
- [gw-BF1-review-A.md](file://BUG-261014-351fcr/gw-BF1-review-A.md) — Record-only reviewer A: recommend acceptance, findings empty; reviewer B owns the CR verdict.
- [BUG-261014-351fcr_spawn-log_-reviewer--reviewer--codex-_RUN-261014-6f656f.log](file://BUG-261014-351fcr/BUG-261014-351fcr_spawn-log_-reviewer--reviewer--codex-_RUN-261014-6f656f.log) — System spawn log captured by task-board
- [BUG-261014-351fcr_spawn-log_-reviewer--reviewer--claude-_RUN-261014-52a5b6.log](file://BUG-261014-351fcr/BUG-261014-351fcr_spawn-log_-reviewer--reviewer--claude-_RUN-261014-52a5b6.log) — System spawn log captured by task-board
- [BUG-261014-351fcr_review-verdict-rev1.md](file://BUG-261014-351fcr/BUG-261014-351fcr_review-verdict-rev1.md) — Reviewer B verdict rev1: accepted

## Created
2026-10-14T01:33:11Z

## Last Update
2026-10-14T08:16:25Z

## Assigned To
[reviewer] reviewer (claude)
