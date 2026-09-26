# TASK-260915-s4qkmr — Verification without restart (AFTER)

All checks from zsh, `set -o pipefail`, real exit codes. NOTHING was restarted,
signalled, or re-execed. BEFORE snapshots: see TASK-260915-s4qkmr_inventory.md.

## 1. Fresh login shell resolves the managed commands (exit 0)

`zsh -lc 'command -v task-board; task-board --version; command -v task-board-tui; task-board-tui --version; command -v tb-sessiond'`:

```
/Users/administrator/.local/bin/task-board
task-board version dev
/Users/administrator/.local/bin/task-board-tui
task-board-tui version dev
/Users/administrator/.local/bin/tb-sessiond
```

All three resolve via the `~/.local/bin` shims (fresh login PATH does not
contain `~/.curator/global/bin`; verified in inventory). No side effects.

## 2. Live processes: same PIDs, paths, hashes — BEFORE == AFTER

AFTER `ps` (all 7 rows; etime advanced, lstart identical):

| PID | started (unchanged) | command (unchanged) |
|---|---|---|
| 35709 | Thu Sep 10 18:18:44 2026 | cache/go-v1/3b57ed0a…/bin/tb-sessiond --board-dir …/skill-project-management/.task-board |
| 46192 | Tue Sep 8 15:56:43 2026 | cache/go-v1/e2029bed…/bin/tb-sessiond --board-dir ~/.task-board |
| 54120 | Thu Sep 10 19:32:21 2026 | cache/go-v1/10f46bb6…/bin/task-board codex --model gpt-6-astra … |
| 39451 | Tue Sep 22 20:53:10 2026 | ~/.local/bin/task-board-main-6cb09a23-curatorlike spawn-runner …/RUN-260922-839c83/… |
| 39477 | Tue Sep 22 20:53:11 2026 | ~/.local/bin/task-board-main-6cb09a23-curatorlike spawn wait RUN-260922-839c83 |
| 41409 | Tue Sep 22 20:56:46 2026 | ~/.local/bin/task-board-main-6cb09a23-curatorlike spawn-runner …/RUN-260922-17ade7/… |
| 41430 | Tue Sep 22 20:56:46 2026 | ~/.local/bin/task-board-main-6cb09a23-curatorlike spawn wait RUN-260922-17ade7 |

AFTER `lsof -Fn` txt mappings: the same 4 paths as BEFORE (3×
`cache/build/go-v1/.sweep-<hash>-<id>/bin/…`, 1×
`~/.local/bin/task-board-main-6cb09a23-curatorlike`).

AFTER sha256 of every in-scope file — all 18 match BEFORE exactly
(exit 0; `task-board-main-6cb09a23-curatorlike` = `18a02361…`, marker =
`01a049df…`, all shims and binaries identical; full AFTER table in run
transcript). No in-scope byte changed other than the newly created backup dir.

Honest bound on "same sha256 of the resolved binaries": the three daemon
images (PIDs 35709/46192/54120) are open-but-UNLINKED mappings — neither the
`sweep-*` paths lsof reports nor the original build dirs exist on disk
(verified ABSENT/no-match, exit 0), so their bytes cannot be hashed from disk
and are reported UNKNOWN, not inferred. Their identity is proven instead by
same-PID + same-lstart + same-argv + same-lsof-txt before/after. The campaign
binary (4 live processes) IS on disk and its hash is identical before/after.
This unlinked-mapping state predates the task (Curator cache GC) and was not
caused or changed by it.

## 3. One hosted board call still answers (exit 0)

Bounded read-only call through the campaign binary (absolute path, no shim):

`~/.local/bin/task-board-main-6cb09a23-curatorlike q 'get(TASK-260915-s4qkmr) { status }'`
→ `{"status":"development"}`, EXIT=0.

## 4. Net host-state delta of the whole task

- ADDED: `/Users/administrator/.curator/backups/20260922T165929Z-legacy-task-board/`
  (17 `cp -p` copies + MANIFEST.txt, 262M). Nothing else written.
- MODIFIED/REMOVED: nothing. All BEFORE hashes == AFTER hashes; all PIDs alive.
