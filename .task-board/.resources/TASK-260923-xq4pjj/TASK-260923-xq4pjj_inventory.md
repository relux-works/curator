# TASK-260923-xq4pjj — host inventory

Scope: live host state only. The live shim and marker are under /Users/administrator/.local/bin; the product-managed canonical command files are under /Users/administrator/.curator/global/bin.

## File inventory

| Path | Before SHA-256 | After SHA-256 | State |
|---|---|---|---|
| /Users/administrator/.local/bin/task-board | 5092fb50d34cdd27540531abdb71b3e68111866842becc8e117101fabf1a8cdb | dc096f769e38e06110fb0f8b25a7a8c99742541982ce51c131b325e381c45faa | Atomically replaced; mode 711; size 72 → 74 bytes |
| /Users/administrator/.local/bin/.curator-managed.json | 7a4efaacaf3d78c6a5dae6f838c67500b367ded96d1529593bb2e7e6ab55a3d5 | same | Unchanged; mode 600; entries remain only tb-sessiond |
| /Users/administrator/.curator/global/bin/task-board | 250e845551336fe2e2f721027289fab0a202ab829e90daf5cbed245961719b0a | same | Unchanged |
| /Users/administrator/.curator/global/bin/tb-sessiond | 6517e6d732f86bef0a45095b206f2f09fba50291f8fc64f77ef267507887fa17 | same | Unchanged |
| /Users/administrator/.local/bin/task-board-main-6cb09a23-curatorlike | 18a0236197ac1f9c3e65e1d82f3812047ec83fd63a7b67fa68a520120d39da5d | same | Off-limits binary untouched |
| task-board-tui in ~/.local/bin and ~/.curator/global/bin | absent | absent | Not restored; absent from the global Skillfile and managed marker |

The shim mode remains 711. Its final bytes are exactly:
#!/bin/sh
exec '/Users/administrator/.curator/global/bin/task-board' "$@"

The managed marker was never edited. It remains byte-identical to the backup and still excludes task-board.

## Live process inventory

The filtered executable process snapshot had 42 rows before the swap and 43 at final capture. The sorted diff command exited 1 because campaign board activity changed one short-lived command: PID 27755 exited; PIDs 25907 and 27709 appeared for unrelated task-board handoff commands. The other 41 baseline PIDs remain with the same start times and executable paths. No process was signalled, restarted, or re-execed by this task.

All task-board CLI process argv paths in the snapshots use the Curator cache binary ending in f26342570d2a865b7b5b18125575f0afc9f7ea17605bffb646fb2e3ad1422a1d/bin/task-board. Its SHA-256 was 6cf0588d73dbdc214fe560f0bd8522df7f5a3ea6279b96d195589ffd2d85b042 before and after. The stable tb-sessiond PIDs 35709 and 46192 retain their same process starts and lsof executable mappings to the pre-existing .sweep-* cache paths. Those mapped files are absent, so their binary hashes are unknown; no hash is inferred.

The live process snapshots are attached as TASK-260923-xq4pjj_processes_before.txt and TASK-260923-xq4pjj_processes_after.txt. lsof on /Users/administrator/.local/bin/task-board returned exit 1 with no open references both before and after the rename. PID 25779, PID 35709, and PID 46192 had the same executable mappings before and after.

The off-limits task-board-main-6cb09a23-curatorlike file hash was identical before and after. No live process command path referenced it in either snapshot.
