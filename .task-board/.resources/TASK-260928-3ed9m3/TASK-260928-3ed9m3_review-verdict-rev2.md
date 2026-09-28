# TASK-260928-3ed9m3 review verdict — rev2 (CR-TASK-260928-3ed9m3-2): ACCEPTED

Candidate: base d8e87bac, tree 9f6f20f3, 7 paths (same as rev1). Reviewer: claude-opus-5-5 low. Scope per delta-review note: internal/envprofile/switch.go.

## Hunk classification: `git diff d8e87bac 9f6f20f3 -- internal/envprofile/switch.go`
- Package doc comment (the dropped trunk line "in-place homes only. The skills tree…"): (a)/(b). Rev1 adds the skills surface, so the comment now describes root context and profile skill directories. Not a regression.
- `profileSkillsSurfaceDir`, `resolveNativeHome`, `materializeScopeWithNativeHome`, `preflightInPlaceScope`, `inPlaceConflictPath`, the skills surface in `materializeOne`: (a) rev1 content. There are 31 matching lines in `git diff 97e85642 refs/campaign/lpnvkn-rev1-20260928`. Preflight keeps E6's `validateProfilePathSources` and routes through `managedPath`, `stateread.Lstat` and `stateread.Readlink`.
- Dropped `linkTarget, readErr := os.Readlink(full)` (switch.go ~L727): (b) replaced by `stateread.Readlink` (the §8.4.1 seam). The nofollow check still runs on `stateread.Lstat` plus Readlink. Manager-owned vs planted detection (`sameStoreTree`) is unchanged.
  - One behaviour change: an unreadable link now reports `environment_surface_unmanaged_conflict: inspect symlink …` instead of `environment_foreign_manager_detected … cannot be inspected; refusing takeover`. It still refuses unconditionally, takeover included, so this is not a relaxation.
  - No test pins the old text (`grep "cannot be inspected" internal/envprofile` only hits credential_link_test.go, which is unrelated).
- openBackup: (b) E5 discipline is kept.
  - `os.RemoveAll(generation)` became `removeManagedTree`, which never follows a link: it uses Lstat, recursion through `managedPath`, and `os.Remove`. This is stricter than trunk.
  - The inline symlink/regular copy became `copyManagedEntry`, which still publishes through `atomicManagedLink` and `atomicManagedFile` and reads via `readManagedRegular`. It adds directory support (skills dirs) through `managedDirectory`.
  - The generation-link guard at L963 is intact.
- The stale-surface removal changed from `removeManagedEntry` to `removeManagedTree`: (a), needed for skill directories, and nofollow-safe.
- No (c) regressions found.

## Runs (from a git-archive extract of 9f6f20f3, zsh, pipefail, real exit codes)
- `go test ./internal/envprofile -run 'Nofollow|Link|Switch|Global|Takeover|Path|Guarded' -count=1` → ok 143s, rc=0.
- `go test ./cmd/curator -run 'Global|Takeover|Profile'` → FAIL. It hit Go's default 600s `-timeout` under host load (timeout panic, not an assertion). Not counted as evidence.
- Narrowed to the four new production-entry rows: `go test ./cmd/curator -run 'TestGlobal(Add|Install)(Publishes|Conflict)' -v` → all 4 PASS, ok 79.8s, rc=0.
- The hosted gate for rev2 was reported green by the orchestrator; this review accepts that for the broad cmd/curator package.

## E5 mutant re-run (bound, not a rev2 regression)
- M1: Lstat→Stat in the parent walk of `managedPath` (nofollow.go:43). Against the E5 rows (`TestEnvironmentWriteNofollowVectors`, `TestResolveRenderedDocumentReplacesStoreSymlinkWithoutFollowing`) it SURVIVES on the candidate, rc=0. It survives identically on TRUNK d8e87bac (rc=0).
- M2: disabling the symlink refusal at nofollow.go:60 alone, and at 60+111 together, also survives. The rows are held by later layers: the nofollow open helpers and the atomic writers. rev2 does not touch nofollow.go or those rows.
- This is a pre-existing E5 residual, not caused by this candidate. It should be recorded against the E5 story: the parent-walk guard has no row that kills it on its own.

## Carried from rev1 verdict
Rev1 accepted lock-first ordering, lock kept on conflict, sync --takeover recovery, and killed the ordering/rollback mutants. The rev2 delta in the other 6 paths is clean (orchestrator line check). The rows switch to the vector driver when SPEC_PIN moves past 4ad8042.

Verdict: ACCEPT revision 2.
