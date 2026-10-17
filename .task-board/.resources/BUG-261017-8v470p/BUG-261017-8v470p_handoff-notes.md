# BUG-261017-8v470p — developer handoff advisory

The developer handoff command exited 0. A separate compact read verified status `to-review`.

The CLI also emitted `run_wrote_outside_worktree` with `policy warn`, `violated=true`. Its report includes unattributed board activity/resources and changes reported as foreign worktrees. The report was extremely large and the tool response was truncated; this worker does not claim to have audited the complete report or cleared the attribution warning.

This worker made no direct control-root file edits: repository changes were made only in the managed Story worktree; board status, checklist, notes and resources were written only through task-board CLI. Temporary outcome upload files were written outside the worktree as required by the assignment. The warning did not prevent the `to-review` transition under the current warn policy. The orchestrator should inspect attribution before any stricter-policy acceptance or integration; the worker has not changed policy or cleared another operator's state.

The advisory changes no product source, tests or configuration. All 5 AC rows remain evidenced on the exact signed green gate tree. Full gate: https://github.com/relux-works/curator/actions/runs/39049991046, exit 0, 20 successful jobs, 2 configured skips.
