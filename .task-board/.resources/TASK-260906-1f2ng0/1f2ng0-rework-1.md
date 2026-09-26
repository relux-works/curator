# TASK-260906-1f2ng0 — rework 1 (orchestrator, binding): Windows platform-case gate

Revision 1's hosted gate (run 35874163469) failed ONLY on Windows, in the platform-case gate, because three
tests you added/changed skip on Windows with a reason the gate does not recognise:

    FAIL  skip with an unrecognised reason on windows: cmd/curator :: TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable/unreadable_module_directory
    FAIL  skip with an unrecognised reason on windows: cmd/curator :: TestProfilePathOperandDiagnosticsDistinguishAbsenceAndUnreadable/unreadable_root
    FAIL  skip with an unrecognised reason on windows: internal/envprofile :: TestLoadMachinePolicyTreatsOnlyAbsentConfigAsDefault

Fix it the way the repository already governs platform skips: read `.github/ci/skip-classes.tsv`,
`.github/ci/platform-case-gate.sh` and `platform-cases.tsv`. Prefer making the case EXECUTE on Windows
(an unreadable path via a Windows-appropriate mechanism) where the behaviour is meaningful there; where it is
genuinely POSIX-permission-only, skip with a registered skip class and exact reason text the gate accepts —
never broaden a class or suppress the gate (`no-broad-suppression.sh` must stay green). Run
`sh .github/ci/gate-selftest.sh` and the focused packages locally (bounded calls). Append "Revision 2" to results,
then `task-board handoff TASK-260906-1f2ng0 --role developer`. A `run_wrote_outside_worktree … policy warn`
block is a warning — verify the status moved to `to-review`.
