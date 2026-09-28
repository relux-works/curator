# TASK-260923-xq4pjj — adopt the hand-written task-board shim, LIVE (metadata, host state)

Last mile of TASK-260915-s4qkmr (accepted). Same safety rules as `s4qkmr-brief.md` (attached): no `mv`/`rm`
over a path a live process resolves, `task-board-main-6cb09a23-curatorlike` untouched, no install/setup/
self-update/daemon action, before/after PID + path + sha256 snapshots, rollback line written first.
Recipe from the s4qkmr reviewer:
1. `cp -p` the current `~/.local/bin/task-board` and `~/.local/bin/.curator-managed.json` into
   `~/.curator/backups/<UTC-stamp>-shim-adoption/` with sha256.
2. Write the Curator-shaped bytes to a temp file in `~/.local/bin/` and `mv` it over the shim ATOMICALLY
   (a rename is atomic; readers see old or new, never partial):
   `#!/bin/sh\nexec '/Users/administrator/.curator/global/bin/task-board' "$@"\n`
   Verify with `cmp` against what `runtimestore.UnixShimContent` produces for that canonical target
   (read `internal/runtimestore` in the curator repo to get the exact format; quote rules matter).
3. Run `curator global install` so the PRODUCT adds `task-board` to the managed marker — never hand-edit it.
   If the command would do anything beyond adopting this one shim (read its dry-run and the code path first),
   stop and report instead.
4. Verify in a fresh login shell (`zsh -lc 'command -v task-board; task-board --version'`); live PIDs/paths
   unchanged. `task-board-tui` was removed ~19:59Z 22.09 by the operator's reinstall — report, do not restore.
Attach inventory/backup/verification resources and hand off with `task-board handoff TASK-260923-xq4pjj --role developer`.
