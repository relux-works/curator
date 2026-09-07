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
- (none)

## Checklist
- [x] SPEC section 4.1: real fragment subprocess with repair, closed parsing and CCJ-1 digest.
- [x] Attach producer evidence and complete independent reviewer acceptance before closure.
- [x] Resolve subprocess preserves argv/profile, always repairs, forwards all stderr and never retries or falls back on failure.
- [x] Closed fragment parser validates required schema and channel shapes, rejecting malformed/unknown/duplicate fields and withdrawn composition.
- [x] CCJ-1 exact canonical bytes and digest match all three recorded A0 fragments and adversarial Unicode/escaping/order cases.
- [x] Production boundary tests and make check pass; SPEC E6 clarification is tied to installed evidence and the existing erratum Story.
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
- [x] F1 Unicode path limits and F2 pinned 49-row corpus pass focused behavioral tests and narrowing mutants

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"claude-fable-5-1/low","text":"Implement the isolated SPEC4.1 fragment boundary from verified installed evidence with strict parsing, canonicalization tests and independent review."}
spawn selection rationale for claude-fable-5-1/low: Implement the isolated SPEC4.1 fragment boundary from verified installed evidence with strict parsing, canonicalization tests and independent review.
spawn agent resolution: Agent selection: claude via explicit_override (preferred_agentic_system: exclusive[claude], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260908-d1701c, max_parallel=3)
spawn run started: [implementer] developer (claude) (run=RUN-260908-d1701c)
A1 §4.1 delivered in the Story worktree, uncommitted. internal/fragment: CCJ-1 reader, closed parser (conformance corpus 9 valid/40 invalid vendored from curator-spec 87a0d00 + adapter channel registry), digest from parsed object (3/3 A0 digests match, also reproduced on installed curator), subprocess resolver with --repair, verbatim stderr, resolve_* mapping, no retry/fallback. Wired into curator-run before the not_implemented refusal. SPEC §4.1 E6 paragraph only (STORY-260908-2utz8k). make check exit 0 twice; 26/26 fragment mutants + 14/14 cli mutants killed; AC 9/9 driven. Evidence: TASK-260908-ranc5y_a1-fragment-evidence.md. Item 2 left for the reviewer. Logbook entry attached for the parent to append (control root untouched).
A1 §4.1 delivered in the Story worktree, uncommitted. internal/fragment: CCJ-1 reader, closed parser (conformance corpus 9 valid/40 invalid vendored from curator-spec 87a0d00 + adapter channel registry), digest from parsed object (3/3 A0 digests match, also reproduced on installed curator), subprocess resolver with --repair, verbatim stderr, resolve_* mapping, no retry/fallback. Wired into curator-run before the not_implemented refusal. SPEC §4.1 E6 paragraph only (STORY-260908-2utz8k). make check exit 0 twice; 26/26 fragment mutants + 14/14 cli mutants killed; AC 9/9 driven. Evidence: TASK-260908-ranc5y_a1-fragment-evidence.md. Checklist item 2: producer evidence attached; the independent reviewer acceptance half is NOT done by the producer — it is the review this handoff requests. Logbook entry attached for the parent to append (control root untouched).
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260908-d1701c, pid=1446, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"claude-fable-5-1/low","text":"Independent review targets strict fragment, canonicalization and actual subprocess boundaries while reusing the successful exact-tree suite."}
spawn selection rationale for claude-fable-5-1/low: Independent review targets strict fragment, canonicalization and actual subprocess boundaries while reusing the successful exact-tree suite.
spawn agent resolution: Agent selection: claude via explicit_override (preferred_agentic_system: exclusive[claude], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260908-75e04c, max_parallel=3)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260908-75e04c)
agent completed: [reviewer] reviewer (claude) (exit=143)
spawn run RUN-260908-75e04c cancelled by operator; operator action required; reason: Operator urgently replaced the model policy: all producers and reviewers must use Codex gpt-6-astra at medium. Stop this Claude run; preserve WIP and existing evidence for supported continuation.
spawn run completed: claude (run=RUN-260908-75e04c, pid=61922, exit=143)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/medium","text":"Explicit urgent operator policy: resume preserved work exclusively on Codex gpt-6-astra medium."}
spawn selection rationale for gpt-6-astra/medium: Explicit urgent operator policy: resume preserved work exclusively on Codex gpt-6-astra medium.
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: exclusive[codex], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260908-a67608, max_parallel=3)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260908-a67608)
CR1 reviewer verdict: changes_requested; repeat-of: none. F1: checkAbsolutePath uses byte length instead of schema character length; reproduced at production binary with Unicode 4096-character path. F2: 49-row corpus coverage gate admits omission of a normative negative (48 rows still pass). See TASK-260908-ranc5y_review-verdict-rev1.md and review-evidence-rev1.tar.gz. Selected boundary tests green; duplicate-key and token-preserving diagnostic mutants killed; exact frozen candidate preserved. Parent may propagate these findings to logbook; no control-root edit performed.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260908-a67608, pid=95008, exit=0)
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6-astra/medium","text":"Latest operator policy: all producer/reviewer work on Codex gpt-6-astra medium; continue focused accepted-evidence workflow."}
spawn selection rationale for gpt-6-astra/medium: Latest operator policy: all producer/reviewer work on Codex gpt-6-astra medium; continue focused accepted-evidence workflow.
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: exclusive[codex], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260908-567fb1, max_parallel=3)
spawn run started: [implementer] developer (codex) (run=RUN-260908-567fb1)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260908-567fb1, pid=56800, exit=0)
spawn selection rationale tuple: {"role":"reviewer","pair":"gpt-6-astra/medium","text":"Focused F1/F2 review on the mandated Astra medium model, with independent boundary and corpus attacks."}
spawn selection rationale for gpt-6-astra/medium: Focused F1/F2 review on the mandated Astra medium model, with independent boundary and corpus attacks.
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: exclusive[codex], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260908-39de12, max_parallel=3)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260908-39de12)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260908-39de12, pid=99136, exit=0)
spawn selection rationale tuple: {"role":"developer","pair":"gpt-6-astra/medium","text":"Accepted fragment code is landed; close only its signed separate-board transaction with existing evidence."}
Story STORY-260908-15st55 stayed on base 25379f22245e3bd8d2b81b192287d04368bf42c7: 1 published Change Request revision(s) are still measured from it — CR-TASK-260908-ranc5y-2 revision 2 (accepted, element TASK-260908-ranc5y, base 25379f22245e3bd8d2b81b192287d04368bf42c7). Carry them forward and the next spawn converges. carry the listed revision(s) forward: review one still awaiting a verdict, or task-board worktree checkpoint <ELEMENT-ID> an accepted one; inspect with task-board worktree status STORY-260908-15st55, or task-board worktree abort STORY-260908-15st55
spawn selection rationale for gpt-6-astra/medium: Accepted fragment code is landed; close only its signed separate-board transaction with existing evidence.
spawn agent resolution: Agent selection: codex via explicit_override (preferred_agentic_system: exclusive[codex], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260908-61cb96, max_parallel=3)
spawn run started: [implementer] developer (codex) (run=RUN-260908-61cb96)

## Precondition Resources
- [fragment-producer-brief.md](file://TASK-260908-ranc5y/fragment-producer-brief.md)
- [verified-fragment-claude_code.json](file://TASK-260908-ranc5y/verified-fragment-claude_code.json)
- [verified-fragment-codex_cli.json](file://TASK-260908-ranc5y/verified-fragment-codex_cli.json)
- [verified-fragment-pi.json](file://TASK-260908-ranc5y/verified-fragment-pi.json)
- [verified-fragment-digests.txt](file://TASK-260908-ranc5y/verified-fragment-digests.txt)
- [fragment-reviewer-brief.md](file://TASK-260908-ranc5y/fragment-reviewer-brief.md)
- [operator-astra-medium-policy.md](file://TASK-260908-ranc5y/operator-astra-medium-policy.md)
- [fragment-rework.md](file://TASK-260908-ranc5y/fragment-rework.md)
- [fragment-review-rev2.md](file://TASK-260908-ranc5y/fragment-review-rev2.md)
- [fragment-complete.md](file://TASK-260908-ranc5y/fragment-complete.md)

## Outcome Resources
- [TASK-260908-ranc5y_spawn-log_-implementer--developer--claude-_RUN-260908-d1701c.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_spawn-log_-implementer--developer--claude-_RUN-260908-d1701c.log) — System spawn log captured by task-board
- [TASK-260908-ranc5y_a1-fragment-evidence.md](file://TASK-260908-ranc5y/TASK-260908-ranc5y_a1-fragment-evidence.md) — A1 §4.1 producer evidence: changed files, exact commands/exits, AC coverage 9/9, mutant table 26/26, source mapping, stated bounds
- [TASK-260908-ranc5y_logbook-entry.md](file://TASK-260908-ranc5y/TASK-260908-ranc5y_logbook-entry.md) — Logbook entry for the parent to append to the launcher control-root LOGBOOK (worker must not edit control root)
- [TASK-260908-ranc5y_fragment-mutants-summary.tsv](file://TASK-260908-ranc5y/TASK-260908-ranc5y_fragment-mutants-summary.tsv) — fragment-mutants.sh summary: 26/26 KILLED
- [TASK-260908-ranc5y_cli-mutants-summary.tsv](file://TASK-260908-ranc5y/TASK-260908-ranc5y_cli-mutants-summary.tsv) — cli-mutants.sh re-run after §4.1 wiring: 14/14 KILLED
- [TASK-260908-ranc5y_make-check.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_make-check.log) — make check final run, exit 0
- [TASK-260908-ranc5y_change-request_rev1.patch](file://TASK-260908-ranc5y/TASK-260908-ranc5y_change-request_rev1.patch) — Change Request CR-TASK-260908-ranc5y-1 revision 1 candidate patch (repository_delta=present, 67 changed paths)
- [TASK-260908-ranc5y_change-request_rev1-validation.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_change-request_rev1-validation.log) — Change Request CR-TASK-260908-ranc5y-1 revision 1 bounded validation log
- [TASK-260908-ranc5y_spawn-log_-reviewer--reviewer--claude-_RUN-260908-75e04c.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_spawn-log_-reviewer--reviewer--claude-_RUN-260908-75e04c.log) — System spawn log captured by task-board
- [TASK-260908-ranc5y_spawn-log_-reviewer--reviewer--codex-_RUN-260908-a67608.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_spawn-log_-reviewer--reviewer--codex-_RUN-260908-a67608.log) — System spawn log captured by task-board
- [TASK-260908-ranc5y_review-verdict-rev1.md](file://TASK-260908-ranc5y/TASK-260908-ranc5y_review-verdict-rev1.md) — CR1 changes requested: Unicode path bound and corpus coverage gate; independent production probes and mutants
- [TASK-260908-ranc5y_review-evidence-rev1.tar.gz](file://TASK-260908-ranc5y/TASK-260908-ranc5y_review-evidence-rev1.tar.gz) — Reproducible reviewer probes, selected mutant outcomes, narrow test logs, frozen-tree verification
- [TASK-260908-ranc5y_spawn-log_-implementer--developer--codex-_RUN-260908-567fb1.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_spawn-log_-implementer--developer--codex-_RUN-260908-567fb1.log) — System spawn log captured by task-board
- [TASK-260908-ranc5y_f1-f2-evidence.md](file://TASK-260908-ranc5y/TASK-260908-ranc5y_f1-f2-evidence.md) — F1/F2 evidence including directly executed make check exit 0
- [TASK-260908-ranc5y_f1-f2-logs.tar.gz](file://TASK-260908-ranc5y/TASK-260908-ranc5y_f1-f2-logs.tar.gz) — Focused tests and mutant exits, upstream manifest, readiness and restored-source checks
- [TASK-260908-ranc5y_f1-f2-make-check.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_f1-f2-make-check.log) — Direct make check exit 0 after F1/F2: build fmt vet test race
- [TASK-260908-ranc5y_change-request_rev2.patch](file://TASK-260908-ranc5y/TASK-260908-ranc5y_change-request_rev2.patch) — Change Request CR-TASK-260908-ranc5y-2 revision 2 candidate patch (repository_delta=present, 67 changed paths)
- [TASK-260908-ranc5y_change-request_rev2-validation.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_change-request_rev2-validation.log) — Change Request CR-TASK-260908-ranc5y-2 revision 2 bounded validation log
- [TASK-260908-ranc5y_spawn-log_-reviewer--reviewer--codex-_RUN-260908-39de12.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_spawn-log_-reviewer--reviewer--codex-_RUN-260908-39de12.log) — System spawn log captured by task-board
- [TASK-260908-ranc5y_review-verdict-rev2.md](file://TASK-260908-ranc5y/TASK-260908-ranc5y_review-verdict-rev2.md) — Independent CR2 acceptance: F1/F2 closed, 49/49 upstream correspondence, 2/2 narrowing mutants killed
- [TASK-260908-ranc5y_review-evidence-rev2.tar.gz](file://TASK-260908-ranc5y/TASK-260908-ranc5y_review-evidence-rev2.tar.gz) — Independent derivation, focused behavioral tests and actual mutant logs
- [TASK-260908-ranc5y_spawn-log_-implementer--developer--codex-_RUN-260908-61cb96.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_spawn-log_-implementer--developer--codex-_RUN-260908-61cb96.log) — System spawn log captured by task-board

## Created
2026-09-07T23:10:50Z

## Last Update
2026-09-07T18:30:00Z

## Assigned To
[implementer] developer (codex)
