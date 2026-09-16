# Review verdict — TASK-260916-3gcc00 rev 1 (CR-TASK-260916-3gcc00-1)

Verdict: **accepted**. Reviewer: claude-fable-5-1, 2026-09-16, Darwin, zsh.

## Candidate identity
- Base 559447efe4a9d6f0c5c0f2a9e254cd3cedea883d, candidate tree 291505ce1c1f9e60ec0f8f7354372ccb460caf48.
- Delta: exactly one new file `.research/TASK-260916-3gcc00_reconciliation.md` (81 lines). Worktree file sha256 94b2a408… equals the candidate-tree blob. No production code touched, matching the "no code changes" AC.

## Fact-checks performed (all confirmed at 559447ef)
- `internal/scriptpolicy/scriptpolicy.go` `Admit` returns `PolicyUnsupported` for any enforced command (lines 83–107). Confirmed.
- `internal/scriptpolicy/conformance_test.go` classifies 5 sections consumed, 6 `refusedBeforeReached`, `audit_label_cases` as "not implemented … owned by STORY-260822-2h0v9j". Confirmed (lines 36–61).
- `cmd/curator/main.go:93` dispatches only the go-v1 build worker; `auditTarget` at :1755. Confirmed.
- `skillcheck.go:61` `executionPolicyIssues` calls `Admit`; `install.go:918` calls `skillcheck.Validate`; `targets.go:278` second guard. Confirmed.
- `.github/ci/platform-cases.tsv:313–316` four scriptpolicy tests on linux/darwin/windows; `ci.yml:140` gate step; `SPEC_PIN` 87a0d006… publishing protocol 1.0.0-rc.9 via tag v1.0.0-rc.11; manifest sha256 0e195ecd… recomputed from the pin. Confirmed.
- Production grep for `script-command-declared-only`, `script_execution_control_unavailable`, `script-capability-evidence` in non-test Go: zero hits. Confirmed missing.
- rc.9 vector at the pin: opt_in 6, audit 4, derivation 4, evidence 14, preflight 5, mandatory_controls 11, native inventory `controls` 8; 66 schema cases each in agent-skill-v8 and csk-skill-v8. Counts in the table match.
- PR merge commits 62d578c5, 77aafa09, a3abcf34 are ancestors of 559447ef; `git log -- internal/scriptpolicy` is exactly a74f1f5 + b902023. Confirmed.
- Board Story description lists exactly the clauses mapped in the table; AC is the ubuntu/macos/windows conformance row; scope field is the placeholder. Confirmed via `task-board q`.

## Independent reruns (exit 0, `set -o pipefail`, zsh)
```
env CURATOR_CONFORMANCE_ROOT="$PWD/.temp/reconciliation-spec/conformance/v1" go test -count=1 ./internal/scriptpolicy ./internal/skillspec ./internal/skillcheck   # ok / ok / ok, rc=0
go test -count=1 ./internal/install -run 'Test(EnforcedScriptCommandIsRefusedAtInstall|DeclaredOnlySchema8ScriptCommandInstallsUnchanged|ActiveScriptCommandsRefusesEnforcedCommand)$'   # ok, rc=0
```
CI run 35012468253 job status accepted from the document, not replayed.

## Assessment
- Every Story clause is mapped with file/test/vector or an exact gap; recommendation (keep Story open, residual tasks R1–R5) is stated and consistent with evidence.
- Coverage is reported as ratios (6/33 behavioral cases consumed; 0/11, 0/8, 0/14, 0/5, 0/4) with stated bounds. Correct.
- Minor, non-blocking: the Story's board note still cites the older spec candidate 859727b1 / manifest 782d6868, while CI pins 87a0d006 / 0e195ecd. The reconciliation correctly uses the CI pin; the orchestrator may want to refresh the Story note when creating R1–R5.
