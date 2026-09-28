# TASK-260923-xq4pjj — results and blocker

Role: developer. Scope: live host state only; no product source edits.

## Host work performed

1. Backed up the current task-board shim and .curator-managed.json with cp -p under /Users/administrator/.curator/backups/20260923T121543Z-shim-adoption. Both backup SHA-256 values match their sources; both cmp checks exited 0.
2. Atomically replaced /Users/administrator/.local/bin/task-board from a same-directory temp file. Its bytes match the product's UnixShimContent output for the canonical target; cmp exited 0. Mode remains 711.
3. Verified a fresh login shell resolves the command through /Users/administrator/.local/bin/task-board and prints task-board version dev (exit 0).
4. Confirmed task-board-tui remains absent and task-board-main-6cb09a23-curatorlike remains byte-identical. No processes were stopped and no daemon actions were performed.

## Adoption blocker

The installed product is curator main-04550e2, matching /Users/administrator/.curator/curator-main-install commit 04550e2. Its globalbins.unmanagedConflict implementation requires both an existing .curator-managed.json entry and a byte-identical Curator-owned target before accepting an existing user-bin shim. The live marker has only tb-sessiond. Therefore, after canonicalizing the existing shim, curator global install still treats task-board as unmanaged and will not add it to the marker.

The product dry-run exited 0 and listed cache hits for task-board and tb-sessiond, with planned commands [task-board,tb-sessiond]. The installed source returns from dry-run before stageGlobalTargets, so it cannot preview forwarding or other target writes. The real global path stages global runtime shims, the global skill and stale removals, forwarding ledger, environment files, and adapter mirrors. That exceeds this task's host-state scope. The real install was not run, and the marker remains unchanged. AC3 is not met.

task-board-tui is absent at both local and global bin paths, absent from the global Skillfile and marker, and was not restored.

## Exact decision needed

Either Curator must provide a targeted supported operation that adopts this exact existing shim and records it in the marker, or the task owner must expand the scope to allow a separately reviewed full global-install plan and its machine-wide target changes. The targeted product operation is the recommended route. No marker was hand-edited.

## Evidence and limits

- Fresh login check: exit 0.
- Backup copies and both backup cmp checks: exit 0.
- Atomic mv: exit 0.
- Function-output comparison and final live shim cmp: exit 0.
- Dry-run: exit 0; real install: not run because installed product code cannot adopt the existing unledgered target and the real operation stages broader global state.
- Process diff: exit 1 due unrelated board activity: 41 prior task-board executable PIDs remained, one short-lived PID exited, and two unrelated task-board command PIDs appeared. The shared cached task-board binary hash stayed unchanged. Daemon hashes are unknown because lsof maps their open images to absent .sweep-* paths.
- curator global status --json: exit 130 after manual interrupt following about 70 seconds with no output; not a gate.
- No repository tests/build were run because the assigned work is host-state-only and no product code changed. The temporary generator was removed.
- Relevant findings and this blocker are recorded here and in the task outcome resources. The campaign rules prohibit editing LOGBOOK.md from this producer run.
