# Review verdict — TASK-260916-3gcc00 (reconcile-script-worker-v1-delivery)

Verdict: **accepted** (CR-TASK-260916-3gcc00-1 revision 1). Reviewer: claude-fable-5-1, RUN-260915-97f2af, 2026-09-16, Darwin, zsh.

## What was reviewed
Candidate tree: worktree HEAD `559447efe4a9d6f0c5c0f2a9e254cd3cedea883d`; only delta is the untracked research file `.research/TASK-260916-3gcc00_reconciliation.md` (1 path, no code changes — matches the AC "no code changes").

## Independent fact checks (all reproduced)
- `internal/scriptpolicy/scriptpolicy.go` `Admit` returns `script_execution_policy_unsupported` for every enforced command; `cmd/curator/main.go` dispatches only the go-v1 build worker (`godriver.WorkerMode`). No script worker dispatch exists.
- Production grep (`internal`, `cmd`, non-test) for `declared-only` warning labels, `unfiltered-declared`, `control_unavailable`: only `build_execution_control_unavailable` in godriver and comments in scriptpolicy/skillspec. No script audit labels or `script_execution_control_unavailable` implementation. Confirms rows "Audit warning class" and "preflight" = missing.
- `internal/scriptpolicy/conformance_test.go` lines 44/47 carry the "unreachable" and "not implemented: audit warning classes, owned by STORY-260822-2h0v9j" classifications as cited.
- Platform ledger `.github/ci/platform-cases.tsv:313-316` lists the four scriptpolicy tests on linux,darwin,windows; `ci.yml` `SPEC_PIN` = `87a0d0060bad…` as cited.
- PR merge commits 62d578c5 (#33), 77aafa09 (#34), a3abcf34 (#37) are all ancestors of 559447ef (`git merge-base --is-ancestor` exit 0 each).
- Spec pin vector `script-host-execution-policy.json` at 87a0d006: 12 top-level keys; case counts 4 derivation / 14 evidence / 5 preflight / 4 audit / 6 opt-in / 11 mandatory controls / 8 inventory controls — matches every ratio in the table.
- GitHub run 35012468253: headSha 559447ef, conclusion success; Test/Gate self-test/Race lanes success; `Test (rose-air)` and `Candidate suite` skipped — matches the document.
- Install negative test `TestEnforcedScriptCommandIsRefusedAtInstall` drives `e.install(Options{})` and asserts no shim is published (negative shape at the production entry point), as the document classifies.

## Tests rerun by me (zsh, `set -o pipefail`, spec pin extracted via `git archive` to /tmp)
```
go test ./internal/scriptpolicy/ ./internal/skillspec/ ./internal/skillcheck/   -> ok, ok, ok, exit=0
go test ./internal/install/ -run 'TestEnforcedScriptCommandIsRefusedAtInstall|DeclaredOnlySchema8ScriptCommandInstallsUnchanged|ActiveScriptCommandsRefusesEnforcedCommand' -> ok, exit=0
```
These are narrow reruns, not the landing suite. No mutants run (research task, no gate under review).

## Minor notes (no rework required)
- The document says "Curator HEAD and local main: 559447ef". Local `main` has since moved to `b0e905d` via board-state commits; `git diff --name-only 559447ef main` touches only `.task-board/` paths, so the pinned reconciliation remains valid for code.
- `native_control_inventory` has 9 JSON keys but 8 `controls`; the document's "eight controls" is the correct reading.

## Conclusion
Every Story description clause plus the platform AC is mapped to landed evidence or a named gap with vector ratios; recommendation (keep Story open, residual R1–R5) is evidence-backed and consistent with the tree. Accept.

## Board recording anomaly (orchestrator action needed)
- `accept_cr(TASK-260916-3gcc00, revision=1, evidence=...)` refused: `change_request_acceptance_unauthorized` — "this reviewer run was handed Change Request revision 0 and is accepting revision 1". `revision=0` is rejected as non-positive.
- `set_status(..., status=done)` refused: task is under a Change Request; done is produced only by the integration transaction.
- `handoff --role reviewer` refused: reviewer role has no end_status.
- Net: the review verdict is **accepted** but this run (RUN-260915-97f2af) was spawned without a CR revision binding, so it cannot persist the acceptance. The task remains in `reviewing` for that reason only. Fix: respawn a reviewer run handed CR-TASK-260916-3gcc00-1 revision 1 (it can cite this verdict and reproduce the checks above), or accept from the orchestrator path.
