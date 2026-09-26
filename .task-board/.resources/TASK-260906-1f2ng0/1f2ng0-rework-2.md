# TASK-260906-1f2ng0 — rework 2 (orchestrator, binding). THIS IS THE ONLY CURRENT INSTRUCTION.

Revision 2's gate (run 35898607780) is green everywhere except the Windows platform-case gate. Your skip reason is
now recognised, but the LEDGER still requires the three unreadable subtests on Windows:

    FAIL ledger case skipped where the ledger does not tolerate it (windows): cmd/curator :: TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable/unreadable_module_directory
    FAIL ... unreadable_root
    FAIL ... internal/envprofile :: TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault/unreadable_config
    (+ the matching "required case skipped on windows")

Choose, per subtest, and say which in results:
(a) PREFERRED — make it EXECUTE on Windows by denying read through the DACL, the way
    `internal/godriver TestFingerprintReportsUnreadableDirectoryIdentically` already does (ledger row 251: "unix mode
    bits, Windows DACL") — reuse that helper; or
(b) if genuinely POSIX-permission-only, give the subtest its own ledger row `linux,darwin  windows  platform-control
    <behaviour>` exactly like `internal/gitops TestExtractPreservesExecutableBit` (row 86), so Windows tolerance is
    declared, not implicit.
Absent-vs-unreadable must stay proven on linux and darwin either way. Run `sh .github/ci/ledger-consistency.sh`,
`sh .github/ci/gate-selftest.sh`, and the focused packages (bounded). Append "Revision 3" to results, then
`task-board handoff TASK-260906-1f2ng0 --role developer`. A `run_wrote_outside_worktree … policy warn` block is a
warning — verify status `to-review`. No LOGBOOK.md.
