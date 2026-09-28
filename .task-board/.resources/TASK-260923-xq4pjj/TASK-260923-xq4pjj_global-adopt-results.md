# TASK-260923-xq4pjj results — adopt task-board shim via `curator global adopt` (host-only)

Curator built from origin/main e4f4fe86 (contains 35472926 global adopt) into $TMPDIR via detached worktree (removed after); `go build` exit 0. Not installed.

## Pre (2026-09-28T04:36Z)
- `cmp ~/.local/bin/task-board canonical-shim` exit 0; canonical bytes `#!/bin/sh\nexec '/Users/administrator/.curator/global/bin/task-board' "$@"\n` (= runtimestore.UnixShimContent, verified in prior run).
- shim sha256 dc096f769e38e06110fb0f8b25a7a8c99742541982ce51c131b325e381c45faa
- marker sha256 7a4efaacaf3d78c6a5dae6f838c67500b367ded96d1529593bb2e7e6ab55a3d5 (entries: [tb-sessiond])
- Backup (cp -p): ~/.curator/backups/20260928T043627Z-shim-global-adopt/{task-board,.curator-managed.json}, hashes identical to above. Earlier backup 20260923T121543Z-shim-adoption retained.

## Adopt
- `curator global adopt task-board --dry-run` exit 0: `would adopt command "task-board" at ~/.local/bin/task-board`
- `curator global adopt task-board` exit 0: `adopted ... (backup: ~/.curator/backups/global-bins/task-board-352120411.bak)`

## Post
- marker: `{"schema_version":1,"entries":["task-board","tb-sessiond"]}` sha256 fcd6f51cc091960ccdd8ea92601a37b383ca7cad2dcb6bd0ceb4303f666389bd (written by product, not hand-edited)
- shim: cmp exit 0, sha256 dc096f76… unchanged.
- `zsh -lc 'command -v task-board; task-board --version'` → /Users/administrator/.local/bin/task-board, `task-board version dev`, exit 0. (`task-board version` subcommand does not exist — exit 1 "unknown command"; `--version` is the flag form.)
- Live processes (PIDs 18162, 33834, 67488 task-board spawn-runners; 35709, 46192 tb-sessiond): pre/post `ps` diff empty (exit 0). task-board cache binary sha256 f28a9245…. Anomaly (pre-existing, not caused here): the two running tb-sessiond executables' cache paths (go-v1/3b57ed0a…, go-v1/e2029bed…) no longer exist on disk — processes still run from unlinked images; adopt touched no cache path.
- `curator global install --dry-run` exit 0; 0 lines matching "conflict"; plans commands=[task-board,tb-sessiond], no files modified.
- task-board-tui: absent from ~/.local/bin (removed ~19:59Z 2026-09-22 by operator reinstall); reported, not restored.

No repository change, no install, no daemon restart.
