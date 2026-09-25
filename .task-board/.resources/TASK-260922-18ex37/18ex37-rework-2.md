# TASK-260922-18ex37 — rework 2 (THE ONLY CURRENT INSTRUCTION)

Revision 3 (refreshed onto 948ae7c9, owners fixed) failed ONLY on Windows (run 35977701956):
    internal/scriptworker TestExecutableIdentityCasesAtProductionEntry: case windows-exec-uncaptured-systemroot-hardlinks is not listed in the
    gap ledger: production resolver accepted = true, want false (reason systemroot-not-manager-captured; platform_owned=true)
That is a real Curator defect (owned by BUG-260924-5p8b0z), not yours to fix here. Add ONE gap-ledger row for exactly that case, owner BUG-260924-5p8b0z, reason
"production resolver accepts a multiply-linked System32 exec whose SystemRoot is not manager-captured (erratum TASK-260924-mcmova)", on the
lane(s) where it fails (windows). Nothing else changes. Bounded: the conformance test on darwin/linux, ledger scripts. Append "Revision 4",
`resource update`, `task-board handoff TASK-260922-18ex37 --role developer`. Do NOT run refresh-candidate unless publication refuses a
base mismatch (then follow the refresh recipe from rework 1). A `run_wrote_outside_worktree … policy warn` block is a warning.
