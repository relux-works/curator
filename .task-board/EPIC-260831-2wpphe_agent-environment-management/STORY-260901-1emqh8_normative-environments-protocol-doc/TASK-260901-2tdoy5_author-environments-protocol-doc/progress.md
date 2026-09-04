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
- [x] protocol/environments.md drafted covering every rev-1 surface in the brief scope list
- [x] Carry-forwards resolved normatively: zero-applicable-modules output, referenced-form layout naming, exact chapter separator bytes
- [x] Every cited core/manager section verified to exist and match
- [x] Signed commit(s) on draft/environments-protocol from base 2a861e5+
- [x] Draft-notes resource attached and handoff to to-review
- [x] Code written per task description and AC
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Gate, refusal, validation, authorization, and attestation behavior attacked, not read — positive-path-only evidence is not accepted
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260901-91a4ab, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260901-91a4ab)
Draft 1 of protocol/environments.md authored and ready for review. Worktree ~/Developer/ReluxWorks/.worktrees/curator-spec-environments-normative, branch draft/environments-protocol, signed commit eddd509e88194f914ca0473a7453d4568e649c7f from base 2a861e5 (exact origin/main; ancestor check passed). Signature verified against maintainers.allowed_signers. Not pushed, no PR, no tag, CHANGELOG untouched, per brief. All 11 rev-1 scope items covered with per-section diagnostics tables; carry-forwards resolved normatively: (1) zero-applicable-modules = header-only output; (2) chapter separator = --- LF LF the Profile heading line; (3) referenced layout = .agent-context/modules/<profile>/<module-path> with the fixed modules/ literal as collision guard. Deliberate deviations flagged for reviewer in resource environments-doc-draft-notes.md: opencode skills kept at manager section-5 native surface per Decision 0010 open question 3 recommendation, not the table cell <home>/skills/; system-prompt output carries no header/chapters; -file flag spellings need pinned-release verification. Evidence: make validate exit 0 on the worktree (53 schemas, 691 vectors, 134 tool tests, go tests), docs-only change, run as tree health check; initial bare make validate failed on missing jsonschema module (environment issue, resolved via venv from requirements-dev.txt). Logbook entry committed on the story branch (389b6ff4).
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260901-91a4ab, pid=14291, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260901-b7cddd, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260901-b7cddd)
REVIEW (RUN-260901-b7cddd): CHANGES REQUESTED on CR-TASK-260901-2tdoy5-1 rev 1 (accept_cr NOT called). Full verdict: resource review-findings-environments-1.md. 2 MAJOR, both in the section-5 byte surface: (M1) managed opencode.json bytes undefined (member set + serialization) yet hashed into the 5.6 cross-platform MUST-equal conformance surface — false environment_surface_drift across conforming implementations; fix by pinning exact bytes, e.g. CCJ-1 of registry.md sec 1. (M2) referenced-layout and managed-home per-profile paths collide on case-insensitive filesystems (profiles Base vs base) — the core sec-2 platform-path collision guard is never invoked on the materialization write path, so the collision-free-by-construction claim is false and clobbering is silent; fix with one sentence extending that rule to materialization writes + a stable code. Minor: 7.2 unsupported-form config error has no diagnostic code; 7.3 -file flag spellings stated as closed fact without the 7.6-style pre-freeze verification caveat. Nits: unnumbered diagnostics tables in 6/11; 8.4 read-failure code naming. All 7 deliberate deviations judged SOUND (incl. no-header system output, flat composition, opencode skills at native surface). Coverage vs Decision 0010 rev-1 complete, no scope leakage; every core/manager citation verified to exist and match; validate.py rerun at eddd509: exit 0 (53 schemas, 691 vectors); head eddd509 signed by the same key as main 2a861e5, base exactly 2a861e5, delta environments.md only. Producer: fix M1+M2 (+m1/m2) on draft/environments-protocol, re-handoff to review.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260901-b7cddd, pid=38745, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260901-898f36, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260901-898f36)
Rework 1 handed to review. All 6 findings from review-findings-environments-1.md applied in one signed commit c3b29b1f7f37829fd4d0c50b2023efa2feb4c615 on draft/environments-protocol (protocol/environments.md only, +48/-16 from reviewed head eddd509). M1: opencode.json pinned to CCJ-1 bytes of the single-member instructions object plus exactly one trailing LF (registry.md §1 cited; §13 vector surface now producible). M2: new §5 Platform-path collisions rule — any materialization/provisioning write of two protocol paths mapping to one platform path fails with new code environment_path_collision before writing; §5.3 by-construction claim qualified to protocol-path space; §8.1 managed homes covered. m1: environment_form_unsupported named in §7.2 + §7.7 row. m2: §7.6-style pre-freeze verification caveat added to §7.3 flag spellings. n1: §6.1/§11.1 numbered diagnostics subsections. n2: new environment_surface_unreadable code in §8.4/§8.5, distinct from marker failure and absence. Evidence: make validate exit 0 at c3b29b1 (53 schemas, 691 vectors, 134 py tests, go tests ok; scratch venv — ambient python3 lacks jsonschema); git verify-commit vs maintainers.allowed_signers exit 0, same ECDSA key as main. Details in resource environments-rework-report-1.md.
agent completed: [implementer] developer (claude) (exit=0)
spawn completion blocked: no new or updated task-scoped outcome artifact was attached. Add or update an outcome resource named like TASK-260901-2tdoy5_results.md and then set status back to to-review.
spawn run completed: claude (run=RUN-260901-898f36, pid=75147, exit=0)
No Change Request revision was published for TASK-260901-2tdoy5 (handoff_unsatisfied): the board is not at to-review

## Precondition Resources
- [producer-brief-environments-doc.md](file://TASK-260901-2tdoy5/producer-brief-environments-doc.md) — Producer brief: normative protocol/environments.md draft 1, worktree from main 2a861e5, rev-1 scope, carry-forward resolutions
- [review-brief-environments-doc.md](file://TASK-260901-2tdoy5/review-brief-environments-doc.md) — Reviewer brief: normative doc draft 1 — coverage vs Decision 0010, byte-rule soundness, deviations verdicts, citations, style, diagnostics
- [producer-brief-environments-rework-1.md](file://TASK-260901-2tdoy5/producer-brief-environments-rework-1.md) — Rework 1: apply all 6 findings (CCJ-1 opencode bytes, platform-collision guard on write path, minor codes/caveats, nits)
- [review-brief-environments-cycle-2.md](file://TASK-260901-2tdoy5/review-brief-environments-cycle-2.md) — Cycle-2: verify 6 findings resolved at c3b29b1, CCJ-1 producibility, collision rule reach, table consistency

## Outcome Resources
- [TASK-260901-2tdoy5_spawn-log_-implementer--developer--claude-_RUN-260901-91a4ab.log](file://TASK-260901-2tdoy5/TASK-260901-2tdoy5_spawn-log_-implementer--developer--claude-_RUN-260901-91a4ab.log) — System spawn log captured by task-board
- [environments-doc-draft-notes.md](file://TASK-260901-2tdoy5/environments-doc-draft-notes.md) — TASK-260901-2tdoy5: carry-forward resolutions, deviations, and reviewer open items for protocol/environments.md draft 1 (commit eddd509 on draft/environments-protocol)
- [TASK-260901-2tdoy5_results.md](file://TASK-260901-2tdoy5/TASK-260901-2tdoy5_results.md) — Results summary: environments.md draft 1, commit eddd509, evidence and scope coverage
- [TASK-260901-2tdoy5_change-request_rev1.patch](file://TASK-260901-2tdoy5/TASK-260901-2tdoy5_change-request_rev1.patch) — Change Request CR-TASK-260901-2tdoy5-1 revision 1 candidate patch (repository_delta=present, 1 changed paths)
- [TASK-260901-2tdoy5_change-request_rev1-validation.log](file://TASK-260901-2tdoy5/TASK-260901-2tdoy5_change-request_rev1-validation.log) — Change Request CR-TASK-260901-2tdoy5-1 revision 1 bounded validation log
- [TASK-260901-2tdoy5_spawn-log_-reviewer--reviewer--claude-_RUN-260901-b7cddd.log](file://TASK-260901-2tdoy5/TASK-260901-2tdoy5_spawn-log_-reviewer--reviewer--claude-_RUN-260901-b7cddd.log) — System spawn log captured by task-board
- [review-findings-environments-1.md](file://TASK-260901-2tdoy5/review-findings-environments-1.md) — Review verdict draft 1: CHANGES REQUESTED — 2 major (opencode.json canonical bytes; case-insensitive platform-path collision in referenced layout), 2 minor, 2 nits; all deliberate deviations judged sound
- [TASK-260901-2tdoy5_spawn-log_-implementer--developer--claude-_RUN-260901-898f36.log](file://TASK-260901-2tdoy5/TASK-260901-2tdoy5_spawn-log_-implementer--developer--claude-_RUN-260901-898f36.log) — System spawn log captured by task-board
- [environments-rework-report-1.md](file://TASK-260901-2tdoy5/environments-rework-report-1.md) — Rework 1: all 6 review findings applied (2 major, 2 minor, 2 nit), signed commit c3b29b1, make validate exit 0

## Created
2026-09-01T11:10:37Z

## Last Update
2026-09-01T12:02:03Z

## Assigned To
[implementer] developer (claude)
