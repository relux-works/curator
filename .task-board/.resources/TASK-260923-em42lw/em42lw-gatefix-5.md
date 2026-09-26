# TASK-260923-em42lw — Windows gate fix (THE ONLY CURRENT INSTRUCTION)

Revisions 6 and 7 (identical tree af66f713) fail ONLY on windows-latest: cmd/curator TestEnvResolveEmitsPermissionsV2AtProductionEntry
(env_permissions_test.go:250) — the emitted fragment fails the copied curator-spec launch-env-fragment-v2 schema:
  at '/env/CODEX_HOME': 'C:\Users\…\environments\acme\codex_cli' does not match pattern '^/[^\u0000]*$'
The schema's env values are POSIX absolute paths. Do NOT loosen the schema copy and do NOT fake a POSIX path on Windows.
1. Find how the EXISTING launch-env-fragment (v1) production-entry/schema tests treat Windows (skip with a declared platform class in
   .github/ci/platform-cases.tsv / skip-classes.tsv, or a spec-defined Windows form). Follow that precedent exactly for the v2 row: if v1
   skips schema validation on Windows with a declared class, do the same with the exact reason (and keep asserting the non-path members —
   permissions lattice — on Windows); if the spec defines a Windows form, emit it. State which and cite the precedent file:line.
2. If there is NO precedent and the spec's v2 schema cannot express Windows paths, record that as a spec gap in results (clause + schema
   pattern) for a follow-up leaf, and declare the Windows skip with that reason.
3. `GOOS=windows go vet ./cmd/curator`; bounded local runs of the permissions rows (darwin) + platform-case gate script. Real exit codes.
4. `task-board m 'set_status(TASK-260923-em42lw, status=development)'` first; before handoff `git fetch origin main` — if trunk moved, combine
   (keep both sides; em42lw's workspace base is aa093918 and trunk has moved: 20o9dk, 2elcdc, 306v4m …) and `task-board worktree
   refresh-candidate TASK-260923-em42lw` if publication would otherwise refuse. Append "Revision 8 — Windows fragment schema",
   `resource update`, `task-board handoff TASK-260923-em42lw --role developer`; stay in the turn. If the loop detector refuses, stop and report.
No CHANGELOG/LOGBOOK edit.
