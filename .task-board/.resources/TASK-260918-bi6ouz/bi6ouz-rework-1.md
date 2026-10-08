# TASK-260918-bi6ouz — rework 1 (orchestrator, binding)

Revision 1 was CHANGES_REQUESTED (`TASK-260918-bi6ouz_review-verdict-rev1*`). Everything else verified OK. Fix exactly:

F1 — production entry unpinned: `os.Lstat` → `os.Stat` in the production call
`foreignManagerHintAt(home, runtime.GOOS, os.Getenv, os.Lstat)` survives every test. Add a row through the
PRODUCTION entry (the repair/use takeover path that emits the notice) with a symlink-to-directory at a resolved
table path (e.g. `~/.local/share/chezmoi` → a real dir elsewhere) asserting NO
`environment_foreign_manager_suspected`; apply the Stat mutant in a disposable copy and show this row kills it.
(On Windows, if symlink creation needs privilege, use the lane's existing symlink-capability probe/skip class — do
not invent a new class.)

F2 — spec deviation: a non-ENOENT lstat failure on one row must NOT abort the scan. §9.5 (802caee
environments.md:2500-2516): a failed inspection is "not present" (never recorded as absent) and "the manager scans
the table in row order and reports the first row whose location is present". Continue the scan; update
`TestForeignManagerHintLstatDiscipline` so EACCES on chezmoi + a real `~/.local/share/yadm` → the notice names
yadm, and add a row where the ONLY present-looking row is unreadable → no notice and no "absent" record. Narrowing
mutant: restore the abort → killed.

Bounded local runs of the package (+ `-race`). Append "Revision 2" to results, then
`task-board handoff TASK-260918-bi6ouz --role developer`. A `run_wrote_outside_worktree … policy warn` block is a
warning — verify the status moved to `to-review`. No LOGBOOK.md.
