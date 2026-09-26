# TASK-260925-h4syhu — lockedNetworkRepository read-failure rows (split leaf; THE ONLY CURRENT INSTRUCTION)

Read `campaign-producer-rules.md`, the task description (surface table, coverage obligation, budget 2), and from TASK-260907-2as5sx:
`_review-verdict-rev9.md` (F1/M1) and `_loop-response.md`. 2as5sx is CLOSED (split); its delta (rev9 = accepted rev7 guard + rev8 refresh
+ the rev9 Windows row rename) sits uncommitted in the Story worktree STORY-260906-1a2i5a — keep every path of it byte-identical EXCEPT
internal/install/draftsources_test.go (and a ledger row if needed), which this leaf owns.
1. `task-board m 'set_status(TASK-260925-h4syhu, status=development)'`.
2. Rows: absent → fallback; Lstat failure → typed manager_state_unreadable, no fallback (POSIX `blocked/child`, runtime.GOOS guard);
   present-but-unusable → stateread.UnusableError (own subtest); Windows: real Lstat failure input (reserved device name / invalid name /
   deny-ACL parent) or a platform-cases/skip-classes ledger row with the exact reason. State whether ERROR_PATH_NOT_FOUND under a regular
   file counting as absent is correct and why.
3. Kill M1 (insert `continue` as the first statement of the Lstat `if err != nil {` at draftsources.go:228) — show survive→killed with
   real exit codes on darwin; `GOOS=windows go vet ./internal/install`. Focused tests only (host memory is tight).
4. No CHANGELOG edit (2as5sx's entry text goes in your results under "## CHANGELOG entry (for release prep)"); no stray files.
5. Attach results, check DoD, `task-board handoff TASK-260925-h4syhu --role developer`; stay in the turn while the gate runs. A
   write-boundary `policy warn` block is a warning.
