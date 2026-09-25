# TASK-260924-m28s6b — rework 2: Windows gate failure (THE ONLY CURRENT INSTRUCTION)

Revision 2 (refresh) hosted gate run 35990838316: macOS/ubuntu green; windows-latest FAILS only your new test
`internal/install TestDraftFreshMachineReplaysDeclaredGitSourceWithoutBindings` — all four subtests (git-tag/install, git-tag/upgrade,
repository/install, repository/upgrade) at draftsources_test.go:436 with
    Status:failed Errors:[source_snapshot_unavailable: declared Git source for review cannot provide the locked revision]
while `TestDraftFreshMachineReplaysEverySourceThroughInstallAndUpgrade` (with bindings) passes on Windows.
1. `task-board m 'set_status(TASK-260924-m28s6b, status=development)'` if needed.
2. Diagnose the Windows-only difference in the no-bindings replay path (how the declared URL/endpoint is formed from the Skillfile entry
   on Windows — drive letters, backslashes, file:// form, path→URL normalisation — or how the exact locked revision is fetched there).
   Decide and STATE whether it is a production defect (a Windows user with a declared local/file source would hit it) or a fixture-only
   artefact. A production defect is fixed in production code with the test unchanged in intent; a fixture artefact is fixed in the
   fixture by constructing the declared source the way a real Windows user would write it. Never skip or ledger it; the addendum rule
   stands: `source_snapshot_unavailable` only for an unreachable source.
3. Artifacts: `gh run download 35990838316 -R relux-works/curator --dir "$TMPDIR/m28"` — never into the worktree.
4. Bounded local runs of the replay rows (darwin) + a cross-compiled `GOOS=windows go vet ./internal/install`. No CHANGELOG edit.
5. Append "Revision 3 — Windows no-bindings replay" (root cause, defect vs fixture, fix) to results, `resource update`,
   `task-board handoff TASK-260924-m28s6b --role developer`. A `run_wrote_outside_worktree … policy warn` block is a warning.
