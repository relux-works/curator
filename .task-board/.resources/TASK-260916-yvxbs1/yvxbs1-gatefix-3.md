# TASK-260916-yvxbs1 — gate fix 3 (THE ONLY CURRENT INSTRUCTION, with yvxbs1-sec-brief.md, yvxbs1-decision-1.md)

Revision 3 (tree 9aee7903) is green on ubuntu/macos; Windows fails TWO tests (run 36337460311):
internal/envprofile TestInstallSurfacesSystemModuleWarning and cmd/curator TestProfileInstallWarnsOnSystemModule —
`environment_store_untrusted: path "…\Temp\…\003" fails permissions boundary check … DACL grants mutation rights to another identity`.
1. Do NOT weaken the DACL check. rc.13 environments.md §4 "`path` source directories": "no identity other than the operator may mutate the
   directory or any component below it" — a Windows temp directory with inherited SYSTEM/Administrators ACEs correctly fails. Fix the two
   tests' fixtures: create the path source directory through your Windows protected-tree helper (pathboundary_test_helpers_windows_test.go /
   ProtectTree), as the other path-source tests do.
2. CORRECTION of the orchestrator's gatefix-1 note: it said the boundary check belongs only at read points. The spec (same section) says the
   manager MUST verify the path source directory "on every `env resolve`, and again under the manager-home mutation lock for every mutating
   profile operation — install, update, use, sync, repair, garbage collection"; a failing directory is entry-class
   environment_store_untrusted for that profile (no fragment, non-current, no rebuild, dry-run reports environment_store_untrusted). Make sure
   the implementation follows THE SPEC (your preflightCurrentPathSources looks like it does — verify env resolve, use, sync, repair, gc too),
   and that test fixtures referencing a non-existent source directory were fixed by creating real protected directories, not by skipping the
   check. State the call sites in the results.
3. Base: rev3's recorded base is eca2bf27 but trunk is 6bd98d49 (E4, E3 landed) and your tree differs from trunk in 30 paths (from eca2bf27 in
   41) — `git fetch origin main` and combine properly so `git diff --name-only origin/main -- . ':!.task-board'` lists ONLY your own paths.
Run the two tests' packages split by -run groups with real exit codes (Windows cannot run locally — reason about the helper). Set status
development; append "Revision 4 — Windows fixtures protected; spec call sites confirmed"; handoff and WAIT for the gate; hand off only green.
No CHANGELOG/LOGBOOK edit.
