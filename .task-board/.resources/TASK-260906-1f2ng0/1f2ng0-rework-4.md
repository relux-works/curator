# TASK-260906-1f2ng0 — rework 4 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 4 (run 35927109604) failed on Windows, and it broke a test that PASSES on main:
    internal/godriver TestFingerprintReportsUnreadableDirectoryIdentically: fingerprint_equivalence_test.go:1179:
      current-user DACL deny did not prevent reading test path … (plus the three subtests of this task, same message)
You changed the shared DACL helper that godriver relied on. Decision (no more DACL experiments on the hosted runner):
1. RESTORE `internal/godriver` (and any helper it uses) byte-identical to the Story base — `git diff <base> --
   internal/godriver` must be empty. Do not share or modify its helper.
2. For the three subtests (`cmd/curator TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable/unreadable_root`,
   `/unreadable_module_directory`, `internal/envprofile TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault/unreadable_config`):
   on Windows skip them with a registered reason, and give EACH its own ledger row `linux,darwin  windows  platform-control
   <behaviour: the unreadable case uses POSIX mode bits; Windows ACL unreadability is not reproducible for the runner account>`
   exactly like `internal/gitops TestExtractPreservesExecutableBit` (ledger row ~86). Remove any Windows-only helper file you
   added for them.
3. Absent-vs-unreadable stays proven on linux and darwin. Bounded local runs + cross-compile the Windows test binary;
   `sh .github/ci/ledger-consistency.sh`, `sh .github/ci/gate-selftest.sh`, `sh .github/ci/no-broad-suppression.sh`.
Append "Revision 5" to results, `task-board resource update` it, then `task-board handoff TASK-260906-1f2ng0 --role developer`.
A `run_wrote_outside_worktree … policy warn` block is a warning — verify status `to-review`.
