## Status
development

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
- [ ] SPEC section 4.1: real fragment subprocess with repair, closed parsing and CCJ-1 digest.
- [ ] Attach producer evidence and complete independent reviewer acceptance before closure.
- [ ] Resolve subprocess preserves argv/profile, always repairs, forwards all stderr and never retries or falls back on failure.
- [ ] Closed fragment parser validates required schema and channel shapes, rejecting malformed/unknown/duplicate fields and withdrawn composition.
- [ ] CCJ-1 exact canonical bytes and digest match all three recorded A0 fragments and adversarial Unicode/escaping/order cases.
- [ ] Production boundary tests and make check pass; SPEC E6 clarification is tied to installed evidence and the existing erratum Story.
- [ ] Code written per task description and AC
- [ ] Relevant tests written for new or changed behavior and passing
- [ ] In a managed Story worktree the candidate is left UNCOMMITTED in the worktree for the handoff to snapshot — never commit on the Story branch. A producer commit moves the branch tip off the recorded checkpoint and the handoff refuses with change_request_candidate_committed_past_checkpoint; repair with `git reset --soft <checkpoint_oid>` before completing again.
- [ ] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [ ] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [ ] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [ ] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.
- [ ] Lint clean
- [ ] Relevant build/validation commands run after changes and build not broken
- [ ] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [ ] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant

## Notes
spawn selection rationale tuple: {"role":"developer","pair":"claude-fable-5-1/low","text":"Implement the isolated SPEC4.1 fragment boundary from verified installed evidence with strict parsing, canonicalization tests and independent review."}
spawn selection rationale for claude-fable-5-1/low: Implement the isolated SPEC4.1 fragment boundary from verified installed evidence with strict parsing, canonicalization tests and independent review.
spawn agent resolution: Agent selection: claude via explicit_override (preferred_agentic_system: exclusive[claude], config: spawn.preferred_agentic_system)
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260908-d1701c, max_parallel=3)
spawn run started: [implementer] developer (claude) (run=RUN-260908-d1701c)

## Precondition Resources
- [fragment-producer-brief.md](file://TASK-260908-ranc5y/fragment-producer-brief.md)
- [verified-fragment-claude_code.json](file://TASK-260908-ranc5y/verified-fragment-claude_code.json)
- [verified-fragment-codex_cli.json](file://TASK-260908-ranc5y/verified-fragment-codex_cli.json)
- [verified-fragment-pi.json](file://TASK-260908-ranc5y/verified-fragment-pi.json)
- [verified-fragment-digests.txt](file://TASK-260908-ranc5y/verified-fragment-digests.txt)

## Outcome Resources
- [TASK-260908-ranc5y_spawn-log_-implementer--developer--claude-_RUN-260908-d1701c.log](file://TASK-260908-ranc5y/TASK-260908-ranc5y_spawn-log_-implementer--developer--claude-_RUN-260908-d1701c.log) — System spawn log captured by task-board

## Created
2026-09-07T23:10:50Z

## Last Update
2026-09-08T15:05:37Z

## Assigned To
[implementer] developer (claude)
