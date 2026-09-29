# TASK-260928-2s0jsc: curator-global-adopt-command

## Description
Add `curator global adopt <command>`: take an existing forwarding shim in the user-bin directory under management when its bytes equal the canonical Curator shim for the canonical target (runtimestore.UnixShimContent / Windows .cmd equivalent) — record it in the ownership marker (.curator-managed.json) with a cp -p backup; refuse (exit non-zero, name the path and the reason) when bytes differ, the entry is a symlink or special file, or the command is not a Curator global command. Never overwrite an unmanaged entry (manager §3 MUST). profiles/manager.md §3 makes forwarding locations and ownership records implementation-specific, so no spec change is needed; document the command in docs/cli.md and troubleshooting.

## Scope
(define task scope)

## Acceptance Criteria
adopt accepts a byte-identical canonical shim and records it; refuses differing bytes/symlink/unknown command with no write; global install then treats the adopted entry as managed; tests on unix and windows; hosted gate green
