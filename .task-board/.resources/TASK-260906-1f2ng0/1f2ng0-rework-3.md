# TASK-260906-1f2ng0 — rework 3 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 3's gate (run 35913351710) failed ONLY on Windows:
    internal/envprofile TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault/unreadable_config:
    path_permissions_windows_test.go:12: exclusive read handle did not prevent reading the file
An exclusive/no-share handle does not make a file unreadable to the same process on Windows. Use the DACL the way the
repository already does for unreadable paths on Windows (find the existing helper: grep for SetNamedSecurityInfo /
DACL in `internal/**/*_test.go`, e.g. the godriver fingerprint unreadable-directory row) — deny FILE_READ_DATA /
FILE_LIST_DIRECTORY to the current user, restore in Cleanup. If a subtest still cannot be made unreadable on the hosted
Windows runner, fall back to option (b) of rework 2 for THAT subtest only (its own ledger row `linux,darwin  windows
platform-control <behaviour>`), and say which. Absent-vs-unreadable stays proven on linux/darwin. Bounded local runs
(+ cross-compile the Windows test binary), ledger-consistency and gate-selftest. Append "Revision 4" to results, then
`task-board handoff TASK-260906-1f2ng0 --role developer`. A `run_wrote_outside_worktree … policy warn` block is a
warning — verify status `to-review`.
