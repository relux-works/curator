# TASK-260907-187z6x review verdict — revision 2: ACCEPTED

Reviewer: claude (independent). Candidate tree a7414535 (base 09b25ef6); worktree verified byte-identical (tracked diff empty; untracked test blob == candidate blob).

## Rev1 R1 fixed
- cmd/curator/profile.go:128-138 honours `updated` on the partial-activation branch; exit code/diagnostics unchanged.
- Row 2 (`not-current/use without takeover`) now asserts `updatedLine`.
- Mutant (reviewer-run): `if updated` -> `if false` (restores unconditional `installed`). Row 2 FAILS: `profile_git_reinstall_test.go:141: operator line "installed profile groot (...)", want the reinstall reported as an update`. KILLED. Restored; diff vs candidate empty.
- R2 evidence wording correction accepted.

## Reruns (zsh, set -o pipefail)
- `go test ./cmd/curator -run 'GitReinstall|Reinstall' -count=1` → ok 203.7s, exit 0 (all 8 rows + path siblings).
- `go test ./internal/envprofile -run Reinstall -count=1` → ok 58.9s, exit 0.
- `go vet ./cmd/curator ./internal/envprofile` → exit 0.
- Full landing suite: not rerun (hosted gate evidence reused per rules).

## Bound / observation (non-blocking)
The printing fix is shared: `activateReinstall` (envprofile.go:~881) returns updated=true on switch error for path reinstalls too, so a PATH same-source reinstall whose activation fails partially now prints `updated profile` instead of `installed profile`. This is a stdout wording change on the path root's error branch (exit/diagnostics identical), consistent with §9.1 reinstall reporting and with the rework-1 brief's instruction; no existing test pins the path partial line either way (unpinned, stated bound). Path success-branch reporting and reinstallPathLocked/reinstallActivation logic are byte-unchanged (only doc comment edited).

Rev1-verified items (parity citations, takeover §8.3.1 safety, activation mutants, ledger) stand per orchestrator note; not re-opened. Refreshed base 48da2690 not in this candidate's base (09b25ef6); rows unchanged at this base.

## accept_cr refused by runtime (stop-the-line, external)
`accept_cr(TASK-260907-187z6x, revision=2, ...)` → `validation_not_bound_to_tree: ... the validation evidence carries no source tree identity ... revalidation is required (candidate_tree_oid=a7414535..., evidence_tree_oid=a7414535...)`.
Code verdict is ACCEPT; the board refuses because the hosted-gate validation evidence for rev2 is not tree-bound. A reviewer cannot produce landing-gate evidence (rules: runtime runs the suite once). Needed: orchestrator/runtime re-runs or re-binds the validation for tree a7414535, then respawn a reviewer (or this verdict) to call accept_cr revision=2.
