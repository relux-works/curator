# TASK-260915-s4qkmr — Reviewer verdict: ACCEPTED (metadata task → done)

Reviewer: claude-opus-5-5, read-only. Shell zsh, pipefail. Reviewed at 2026-09-23. No writes to ~/.local/bin, ~/.curator or any process.

## 1. Backups are real — CONFIRMED
- `~/.curator/backups/20260922T165929Z-legacy-task-board/`: 18 files (17 copies + MANIFEST.txt).
- Recomputed sha256 of all 17 copies against the manifest table: 17/17 match, bad=0.
- All 10 legacy standalone binaries: source sha256 today == backup == manifest (including `task-board-main-6cb09a23-curatorlike` = 18a02361…). Bytes untouched, nothing removed.

## 2. Safety rules held during the producer run — CONFIRMED (with a timeline bound)
- The producer's activity ended with handoff at 2026-09-22T17:10Z. Every in-scope mtime that changed is later: 19:51–19:59Z (23:51–23:59 +0400).
- Today, 5 manifest rows differ from the current host. None of these changes came from the producer run (timeline above):
  - `task-board-tui` is gone from both `~/.local/bin` and `~/.curator/global/bin`;
  - the marker now lists only `tb-sessiond`;
  - global `task-board`/`tb-sessiond` shims were regenerated (task-board now execs cache build f2634257…);
  - `~/.local/bin/task-board` was rewritten with identical bytes (mtime 19:51Z, still the hand-written unquoted form).
  This matches a later `curator global` install/remove by another actor at ~19:59Z (consumers.json, env.sh and global/ also changed 19:59Z). Bound: I cannot attribute that actor from the activity log. The orchestrator should confirm it was intended. The backup is still the valid pre-image for rollback.
- PIDs: 35709 and 46192 (tb-sessiond) are still running with the same lstart. 54120 and the four campaign runner PIDs have exited since (normal spawn completion). No `6cb09a23` process is live now. The producer's AFTER table was taken while they were alive.

## 3. The stop was right — CONFIRMED against source
- `internal/install/global.go:302-308`: dry-run returns before staging. Its claim that the dry run cannot preview the refusal is correct.
- `internal/globalbins/globalbins.go` `unmanagedConflict`: an existing target is a conflict unless it is BOTH in the ledger AND passes `ownedTarget`. The code comment says a ledger entry is not enough.
- `internal/globalbins/stage.go:~103`: a conflict emits "not managed by Curator" and `continue`s, so the name never enters `nextManaged`.
- `runtimestore.UnixShimContent` emits `exec '<quoted>'`. The live `~/.local/bin/task-board` is `exec /Users/...` (unquoted), so it is non-canonical. The producer's refusal reason is correct. Hand-editing the marker would give a false ledger entry that is still refused. Rejecting that was correct.

## 4. Verification — CONFIRMED (current state)
`zsh -lc 'command -v task-board tb-sessiond task-board-tui; task-board --version'` gives:
- `~/.local/bin/task-board`
- `~/.local/bin/tb-sessiond`
- `task-board version dev`
- rc=0

`task-board-tui` no longer resolves because of the later external removal described in §2. It resolved in the producer's run.

## Last mile (follow-up leaf)
1. Back up the current `~/.local/bin/task-board` and marker (cp -p, sha256).
2. Replace the hand-written `~/.local/bin/task-board` with the Curator-shaped bytes, written to a temp file and moved atomically:
   `#!/bin/sh\nexec '/Users/administrator/.curator/global/bin/task-board' "$@"\n`
   Check it with `cmp` against `UnixShimContent(canonical, nil)`, then run `curator global install` so the product adds the `task-board` marker entry. Do not hand-edit the marker.
3. Decide whether `task-board-tui` should come back. It was removed at ~19:59Z by an unattributed global operation.
4. Use the same before/after capture: hashes, `ps` lstart, and a fresh-login `command -v`.
5. Retire the 10 legacy binaries (≈270 MB) once the campaign has finished and no process has resolved them for 7 days, re-backing them up first.
