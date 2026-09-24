# TASK-260916-2ok97n (R5) — publish revision 3 onto fresh trunk (THE ONLY CURRENT INSTRUCTION). HIGHEST PRIORITY.

Your revision-3 work (hard-linked System32 cmd.exe identity, nil ExecSearchDirs → default list, typed
`script_execution_declared_exec_unresolved` refusal, rows + 3 killed mutants) is DONE and sits uncommitted in the
Story worktree (safety ref `refs/campaign/r5-rev3-delta-20260923`, tree 450f8419). Publication was refused:
`change_request_base_authority_mismatch` — the Story checkpoint d239c342 (four checkpointed leaves on 48da2690)
does not descend from trunk `fad88136` (2qvzwk + 1bfk8y landed since). `worktree converge` cannot move a diverged
Story; the sanctioned exit is `worktree refresh-candidate` from a live producer (you).

1. `task-board m 'set_status(TASK-260916-2ok97n, status=development)'`.
2. Combine trunk's incoming content INTO your working candidate first (refresh-candidate requires it):
   `git diff 48da2690 fad88136` touches `.github/ci/platform-cases.tsv`, `.github/ci/platform-exclusions.tsv`,
   `CHANGELOG.md`, `internal/gitops/gitops.go`, `internal/gitops/dirfold_test.go`. Apply that incoming delta to the
   worktree (e.g. `git diff 48da2690 fad88136 | git apply --3way`), resolve overlaps keeping BOTH sides (ledger
   rows from both; CHANGELOG entries from both), leave nothing staged.
3. `task-board worktree refresh-candidate TASK-260916-2ok97n`. If a checkpoint replay conflicts, follow its own
   `--replay-resolutions` template instructions exactly — never hand-commit the replay worktree.
4. Prove the refreshed candidate = your rev-3 work + trunk's incoming content, nothing else (per-file summary).
   Re-run bounded: `go test ./internal/scriptworker/... ./internal/gitops/...`, `sh .github/ci/ledger-consistency.sh`,
   `sh .github/ci/gate-selftest.sh`.
5. Append "Revision 3 — refresh onto fad88136" to results, then `task-board handoff TASK-260916-2ok97n --role
   developer`. A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
No product change beyond the combination. No LOGBOOK.md.
