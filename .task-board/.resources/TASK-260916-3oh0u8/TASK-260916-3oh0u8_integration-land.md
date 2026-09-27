# TASK-260916-3oh0u8 — integration landing preconditions (rev3)

CR: `CR-TASK-260916-3oh0u8-3`, revision 3, ACCEPTED on content (rev2 accepted; rev3 = re-apply on trunk eca2bf27, gate green per delta review note).
Board status: `integrating` (verified via `task-board q`, left untouched).

Worktree: `task-board/story/STORY-260916-2otjbn`, HEAD = `eca2bf27` (no own commit; changes uncommitted for handoff snapshot).

Changed paths (`git diff --name-only HEAD -- . ':!.task-board'`), exactly the 6 rev3 paths, no untracked strays:
- `.github/ci/platform-cases.tsv`
- `cmd/curator/envstatus.go`
- `cmd/curator/main.go`
- `cmd/curator/status_test.go`
- `cmd/curator/umbrella.go`
- `cmd/curator/umbrella_test.go`

No file changed, no commit, no `worktree integrate`/`checkpoint` executed, no status/handoff command run — per the integration assignment the runner performs the bound landing synchronously. No CHANGELOG/LOGBOOK edits.
