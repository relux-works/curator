# TASK-260910-32gki6 — Windows gate fix (THE ONLY CURRENT INSTRUCTION, with 32gki6-sec-brief.md, 32gki6-decision-1.md)

Rev2 (tree 36d8eefe) is green on linux/macOS; Windows fails 132 envprofile tests (run 36473999406) with
`environment_store_untrusted: profile store root failed permissions check: permissions: C:\…\Temp\…\001 …` (and the profile lock).
environments.md §4: the environments root, profile store root, lock and store entries are "manager-created, manager-protected" state. On
unix Curator already creates them private (0700/0600), which is why those lanes pass; on Windows it creates them WITHOUT an owner-only DACL,
so your (correct) DACL check refuses every Windows home. Fix the PRODUCT: every place the manager creates the environments root, profile
store root, store entries, lock files and markers must, on Windows, apply the owner-only DACL through the repo's existing helper
(internal/privatedir or the buildcache Windows protection used for other protected state) — at creation and when publishing a rebuilt
entry. Do not relax the check. Test fixtures that create these paths directly must go through the same product helper. An existing
unprotected root created by an older Curator is refused per §4 class (a) with a diagnostic that names the path and the repair; state that
upgrade consequence in the results (the orchestrator will decide whether a repair command is needed). `GOOS=windows go vet ./...`; unix
`go test ./internal/envprofile` split by -run groups with real exit codes. Set status development; update results (`git diff 36d8eefe`
non-empty); handoff; END YOUR TURN. No CHANGELOG/LOGBOOK edit.
