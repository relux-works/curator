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
- [x] environments.md: path kind normative + import flow + classification + diagnostics
- [x] 9.1 install ref selection specified, core-6 aligned
- [x] Marker schema/vector delta with recorded evolution rationale; make validate green
- [x] Signed commits; notes; handoff
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
spawn queued: [implementer] developer (claude) (run=RUN-260901-f0e642, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260901-f0e642)
Delivered as one coherent delta on draft/environments-onboarding-import (worktree ~/Developer/ReluxWorks/.worktrees/curator-spec-onboarding-import, base origin/main=62e592a, signed commit f8d7e7a, verifies against maintainers.allowed_signers; not pushed per brief). (1) path source kind normative in environments.md §1: repo-shaped operand, copy-on-install snapshot, state-hash pin, §6.2 tree discipline, root .git excluded/nested rejected, missing vs unreadable operand kept as distinct diagnostics. (2) New §9.6 onboarding import: closed detected-surface list (native root-context + unledgered global skills entries), lossless/lossy classification with named loss list, per-operation consent gate (environment_import_lossy), content-preserving CRLF->LF reassembly normalization, skill migration via recovered exact declarations with environment_import_skill_foreign; §9.5 completed, diagnostics moved to §9.7. (3) §9.1 ref selection: install-level --tag|--branch|--revision, exactly one, whole-repository snapshot, default = track remote HEAD default branch, resolved commit = effective pin, --strict-tags unchanged, profile_install_ref_conflict for mixed refs; core §6 grammar aligned. (4) Marker schema 1 evolved IN PLACE (path branch: source_path + state_sha256 + optional imported_from_native const true) — recorded rationale: surface postdates rc.9 pin, no tag/claim/reader binds its bytes; fragment schema and determinism vectors deliberately unchanged (pin grammar closed at commit/state spellings). 7 schema-cases added via the generator; mutation-tested (loosen const -> validate exit 1; delete branch -> exit 1). make validate exit 0 (57 schemas, 773 vectors, 147 unittests, go tests); generator twice byte-identical. Minimal consistency edits noted: manager.md §12.3/§12.6 + cli/curator.md examples. Full decisions in TASK-260901-3gr1fe_onboarding-import-notes.md. Ready for review.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260901-f0e642, pid=51804, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260901-8f5709, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260901-8f5709)
REVIEW ACCEPT (RUN-260901-8f5709): CR-TASK-260901-3gr1fe-1 rev 1 accepted; element parked at to-review for orchestrator checkpoint/integration + done with commit_ack=scope_committed. Full verdict in TASK-260901-3gr1fe_review-findings-onboarding-1.md. Highlights: (1) rc.9 manifest 4-line delta proven benign — generator-owned candidate pin (main.go:2109, regenerate-check diffs it), tag v1.0.0-rc.9 bytes untouched, cef93fb precedent on main, regenerated manifest sha independently equals committed pin 90ee8047. (2) Path-in-rev-1 promotion justified: no tag contains the marker schema (verified absent from v1.0.0-rc.10 tag tree), no claims; deferral text fully removed everywhere, grep-clean. (3) Schema-1 in-place evolution sound; gate attacked with three mutants incl. a reviewer-run additionalProperties-widening mutant the producer did not run — all caught, validate exit 1 each, exit 0 restored. (4) make validate reproduced on clean f8d7e7a export: exit 0, 57 schemas/773 vectors/147 unittests/go tests; generator twice byte-identical. (5) Signature f8d7e7a Good against maintainers.allowed_signers. Non-blocking minors M1-M3 recorded in verdict (relative path operand outside a project unnamed outcome; divergent-secondary loss rule presupposes a detected primary root file; Decision 0010 retains superseded sequencing sentence).
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260901-8f5709, pid=24758, exit=0)

## Precondition Resources
- [producer-brief-onboarding-import.md](file://TASK-260901-3gr1fe/producer-brief-onboarding-import.md) — Onboarding import + 9.1 ref selection brief
- [review-brief-onboarding-import.md](file://TASK-260901-3gr1fe/review-brief-onboarding-import.md) — Reviewer brief: release-manifest delta flag, rev-1 promotion verdict, schema evolution, normative quality, vectors

## Outcome Resources
- [TASK-260901-3gr1fe_spawn-log_-implementer--developer--claude-_RUN-260901-f0e642.log](file://TASK-260901-3gr1fe/TASK-260901-3gr1fe_spawn-log_-implementer--developer--claude-_RUN-260901-f0e642.log) — System spawn log captured by task-board
- [TASK-260901-3gr1fe_onboarding-import-notes.md](file://TASK-260901-3gr1fe/TASK-260901-3gr1fe_onboarding-import-notes.md) — Onboarding import + 9.1 ref selection: decisions incl. schema-evolution rationale, loss-list definition, diagnostics, validation evidence
- [TASK-260901-3gr1fe_change-request_rev1.patch](file://TASK-260901-3gr1fe/TASK-260901-3gr1fe_change-request_rev1.patch) — Change Request CR-TASK-260901-3gr1fe-1 revision 1 candidate patch (repository_delta=present, 1 changed paths)
- [TASK-260901-3gr1fe_change-request_rev1-validation.log](file://TASK-260901-3gr1fe/TASK-260901-3gr1fe_change-request_rev1-validation.log) — Change Request CR-TASK-260901-3gr1fe-1 revision 1 bounded validation log
- [TASK-260901-3gr1fe_spawn-log_-reviewer--reviewer--claude-_RUN-260901-8f5709.log](file://TASK-260901-3gr1fe/TASK-260901-3gr1fe_spawn-log_-reviewer--reviewer--claude-_RUN-260901-8f5709.log) — System spawn log captured by task-board
- [TASK-260901-3gr1fe_review-findings-onboarding-1.md](file://TASK-260901-3gr1fe/TASK-260901-3gr1fe_review-findings-onboarding-1.md) — Review verdict ACCEPT: all 6 priority checks pass (rc.9 delta proven generator-forced, schema gate held under reviewer-run mutants, validate+determinism reproduced independently); minors M1-M3 recorded

## Created
2026-09-01T17:42:56Z

## Last Update
2026-09-01T18:22:34Z

## Assigned To
[reviewer] reviewer (claude)
