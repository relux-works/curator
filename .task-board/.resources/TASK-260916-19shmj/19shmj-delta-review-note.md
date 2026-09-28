# Delta review — TASK-260916-19shmj rev6 (orchestrator, binding; THE ONLY CURRENT INSTRUCTION)

Rev5 ACCEPTED (your verdict, delta eca2bf27..9c4311a5). Rev6 re-applies it on trunk 97e85642 (E1, E3, E4, ryh3kw landed): base = trunk,
tree f3d3e94c, gate green, 8 paths (state_read_guard_test.go no longer differs from trunk — confirm that is correct, i.e. trunk already
carries the needed rows). Orchestrator line check vs (trunk ∪ rev5): 6 paths clean. Review ONLY:
1. internal/envprofile/managed.go — 1 trunk line no longer present: per the re-apply brief, writes trunk added since eca2bf27 (E3 seed
   strip/record, ryh3kw, E1) must also go through the nofollow helpers (atomicManagedFile/atomicManagedLink/managedPath). Confirm the removed
   line is such a direct write replaced by the helper, that no trunk behaviour changed, and that every managed write in the file now follows
   §8.3.1 (grep for os.WriteFile/os.Symlink/os.Create/os.Rename on managed paths).
2. .github/ci/root-artifacts.tsv — the internal/envprofile row merged: every vector named by trunk and rev5 present once.
Run `go test ./internal/envprofile -run 'Nofollow|Seed|Codex|Credential|Link|Drift|Guarded'` and `GOOS=windows go vet ./internal/envprofile`
with real exit codes. accept_cr or changes requested with file:line. No LOGBOOK.md.
