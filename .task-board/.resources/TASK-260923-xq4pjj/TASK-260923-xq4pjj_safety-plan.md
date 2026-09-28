# TASK-260923-xq4pjj — prechange safety plan

Captured 2026-09-23 12:15:43Z, before any change to the live shim or marker.

## Current state

- `/Users/administrator/.local/bin/task-board`: regular executable, mode 711, 72 bytes, sha256 `5092fb50d34cdd27540531abdb71b3e68111866842becc8e117101fabf1a8cdb`.
- `/Users/administrator/.local/bin/.curator-managed.json`: mode 600, 64 bytes, sha256 `7a4efaacaf3d78c6a5dae6f838c67500b367ded96d1529593bb2e7e6ab55a3d5`; entries are only `tb-sessiond`.
- `lsof -nP /Users/administrator/.local/bin/task-board`: exit 1, no open reference.
- `task-board-tui` is absent from both `~/.local/bin` and `~/.curator/global/bin`, and is absent from the global Skillfile and managed marker. It will not be restored.
- Curator global manifest currently declares only `project-management`; the product dry-run lists `task-board` and `tb-sessiond`.

## Rollback line, written before the shim swap

If the atomic swap needs reverting, first confirm no live PID has the target open, then run:

`cp -p '/Users/administrator/.curator/backups/20260923T121543Z-shim-adoption/task-board' '/Users/administrator/.local/bin/.task-board.rollback.tmp' && mv '/Users/administrator/.local/bin/.task-board.rollback.tmp' '/Users/administrator/.local/bin/task-board'`

The swap will be in the same directory, so the final rename is atomic. The managed marker will not be hand-edited.

## Product adoption constraint found before mutation

The live marker has no `task-board` entry. Curator's `unmanagedConflict` requires both a ledger entry and a byte-identical owned target before it will accept an existing path. Therefore changing the shim bytes alone does not let `StageForwarding` add `task-board` to the marker.

The product dry-run returned exit 0 and said no files were modified. It listed cache hits for `task-board` and `tb-sessiond`, and planned commands `[task-board,tb-sessiond]`; it does not stage or preview target replacements. The real global install path stages the global runtime and skill, forwarding ledger, environment files, and adapter mirrors. That is broader than this task's host-state scope, so the real install will not be run unless a targeted supported adoption path is established.
