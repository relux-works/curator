# TASK-260923-em42lw — rework 1 (orchestrator, binding): ledger row names a test that is not compiled

Revision 1's hosted gate (run 35877968471) failed on every OS in "Ledger consistency" and in the macOS race
platform-case gate, for one reason:

    FAIL  required on linux|darwin|windows but not compiled into that build: internal/envfragment :: TestFragmentEmissionMatchesReference
    FAIL  required case never ran on darwin: internal/envfragment :: TestFragmentEmissionMatchesReference

The row you added to `.github/ci/platform-cases.tsv` (or the ledger) names a test that does not exist under that
exact name in `internal/envfragment` in the candidate (renamed, subtest-only, other package, build tag, or
`_test` file not in the package). Make the ledger row and the real test agree: `go test ./internal/envfragment
-list '.*'` must print the exact name the row requires, and it must run on all three OSes. Run
`sh .github/ci/ledger-consistency.sh` and `sh .github/ci/gate-selftest.sh` locally (bounded calls), plus the
focused package with `-race`. Append "Revision 2" to results, then `task-board handoff TASK-260923-em42lw --role
developer`. A `run_wrote_outside_worktree … policy warn` block is a warning — verify the status moved to `to-review`.
