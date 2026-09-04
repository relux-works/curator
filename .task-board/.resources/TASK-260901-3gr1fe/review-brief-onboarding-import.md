# Review brief: onboarding import + §9.1 ref selection

Subject: worktree `~/Developer/ReluxWorks/.worktrees/curator-spec-onboarding-import`, branch `draft/environments-onboarding-import`, head `f8d7e7a`, base `62e592a` (= main). Producer notes `onboarding-import-notes.md` on TASK-260901-3gr1fe — read fully; judge every recorded decision.

## Priority checks
1. **`release/1.0.0-rc.9.json` is in the delta (4 lines).** Released manifests are immutable records ("Release tags are immutable; a defective release is superseded"). Determine exactly what changed and whether ANY modification of a released manifest is admissible here. If it is a checksum/manifest side effect of vector regeneration tooling, that is a process bug — the fix must not rewrite released records. This is presumptively blocking unless proven benign AND consistent with GOVERNANCE/COMPATIBILITY.
2. **Path-in-revision-1 promotion** vs landed Decision 0010 D5 ("delivered as its own tracked story ... targeted between revisions 1 and 2") and D1 ("delivered with the onboarding import story, not in revision 1"). Producer argues that is delivery sequencing, not wire-revision mandate. Verdict: does the promotion contradict the landed decision text, and if so does the rationale justify amending prose vs reverting? Check every place the deferral was removed (environments.md §9.5, manager §12.3 MUST-reject) for consistency.
3. **Schema-1 in-place evolution** of agent-environment-marker-v1 (+12): rationale sound under COMPATIBILITY.md pre-release rules? Old vectors still valid? New pin branch closed and strict?
4. **Normative quality**: import flow (inventory reassembly, lossless/lossy definition per adapter surface list, consent gate, skill migration warnings, auth untouched), diagnostics closed and owned by exactly one table; §9.1 ref selection (install-level --tag|--branch|--revision, repo-wide, default tracking, strict-tags carryover) aligned with core §6 and Skillfile branch rules; syntactic path-recognition rule watertight (no filesystem probe, scp-form unshadowable).
5. **Vectors/validate**: rerun `make validate` and the generator twice yourself; determinism claims hold; new cases actually exercise the new branches.
6. Cross-references, house style, no manager/cli drift beyond the minimal noted sentences.

Verdict: `review-findings-onboarding-1.md` on TASK-260901-3gr1fe; blocking/major -> development; else ACCEPT + accept_cr. Read-only; ignore tools/__pycache__ and any .venv.
