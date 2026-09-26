# TASK-260907-2as5sx — rework 4 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 5 was CHANGES_REQUESTED for F1 (`TASK-260907-2as5sx_review-verdict-rev5*`): the guard is a fixed named list (37 functions in
`internal/envprofile/state_read_sites_test.go:518-640`, and `internal/marker/read_state_test.go:94`), so a NEW collapse site survives —
the reviewer added `readNewManagerState` doing `os.ReadFile(…)` + `if os.IsNotExist(err) { return "default" }; return "default"` and
both guard tests stayed green. This task exists to make a sixth site impossible.
Make the guard DENY-BY-DEFAULT: a test that scans every non-test .go file under `internal/` and `cmd/` (outside `internal/stateread`)
with go/ast for filesystem reads (os.ReadFile/ReadDir/Stat/Lstat/Open, fs.* equivalents) whose error handling tests not-exist
(os.IsNotExist, errors.Is(…, fs.ErrNotExist)) and fails unless the function is (a) routed through the seam or (b) on an explicit,
reviewed allowlist (file:function + one-line reason: operator-selected/source-tree/cache read etc.). Report the ratio "N guarded via
seam / A allowlisted / M scanned" in the test log and results. Kill the reviewer's mutant with the named test (run it in a disposable
copy). Keep everything else of revision 5 byte-identical. If a repo-wide scan is genuinely infeasible for some construct, state it with
alternatives weighed (AC clause) — but the default must be deny.
Bounded runs; gofmt; no stray files (results are board resources). Append "Revision 6" to the results resource, `resource update`, then
`task-board handoff TASK-260907-2as5sx --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
