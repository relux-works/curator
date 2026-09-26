# TASK-260915-s4qkmr — migrate global CLI ownership, surgically, on a LIVE machine

Operator decision 2026-09-22: **there will be no quiet window — do it live, surgically.** That makes
the safety rules below part of the acceptance criteria, not advice. `task_class=metadata`: you get
no Story workspace and publish no Change Request; the deliverable is host state plus evidence
resources on the board.

## What is already true (verify, do not assume)
- `~/.local/bin/task-board` is ALREADY a shim: `#!/bin/sh` + `exec ~/.curator/global/bin/task-board "$@"`.
- `~/.local/bin/.curator-managed.json` lists only `task-board-tui` and `tb-sessiond` as managed
  entries — `task-board` itself is not listed although it is shimmed.
- `~/.curator/global/bin/` holds three managed scripts (`task-board`, `task-board-tui`, `tb-sessiond`).
- Legacy STANDALONE binaries sit in `~/.local/bin`: `task-board-main-154956a6`,
  `task-board-main-260914`, `task-board-main-6cb09a23-curatorlike`,
  `task-board-main-91959c1a-curatorlike`, `task-board-main-91959c1a-nocgo`,
  `task-board-main-91959c1a-nocgo-nop`, `task-board-main-ac2ad9c0-curatorlike` (~27 MB each).
- LIVE processes right now resolve absolute paths under `~/.curator/cache/build/go-v1/<hash>/bin/`
  (two `tb-sessiond`, one `task-board codex`) and the campaign orchestrator runs
  `~/.local/bin/task-board-main-6cb09a23-curatorlike`.

## Hard safety rules (violating any of these fails the task)
1. **Never move, rename, delete, truncate or overwrite** any file that a live process resolves.
   Back up by COPY (`cp -p`) only. No `mv`, no `rm`, no `install -f` over a live path.
2. **`task-board-main-6cb09a23-curatorlike` is OFF LIMITS in this leaf.** The campaign orchestrator
   and every running spawn resolve it by absolute path. Record it as an explicit residual with the
   retirement plan; do not touch the bytes.
3. No `make install*`, no `scripts/setup.sh`, no `task-board self-update`, no daemon restart, no
   `launchctl` action, no editing of `~/.curator/global/bin/*` in place.
4. Do not stop, signal or re-exec `tb-sessiond` or any hosted session.
5. Every step is reversible with one command; write the rollback line for each step BEFORE running it.

## Deliverable
1. **Inventory** (`TASK-260915-s4qkmr_inventory.md`): every `task-board*` and `curator*` entry in
   `~/.local/bin` and `~/.curator/global/bin` with type (shim / standalone binary / symlink), size,
   mtime, sha256, and for each: is it referenced by a live process (`ps`/`lsof`), by
   `.curator-managed.json`, by a launchd plist, or by the campaign config. State which shim points
   where and which Curator cache build each resolves to.
2. **Backup** (`TASK-260915-s4qkmr_backup-manifest.md`): copy every legacy standalone binary and
   every current shim into `~/.curator/backups/<UTC-stamp>-legacy-task-board/` with a manifest
   (path → sha256 → size). Verify each copy's sha256 equals the source. Nothing is removed.
3. **Managed publication**: make the shimmed commands explicitly Curator-managed — i.e. bring
   `task-board` into `.curator-managed.json` the way Curator itself does it (find the supported
   command: read `curator --help` / `cmd/curator` for the global/managed-shim surface; use the
   product's own path, never hand-edit the marker JSON unless the product has no such command, and
   if you must, back the file up first and say so). If the product's command would rewrite a live
   shim, DO NOT run it — record what it would do and stop at that step with evidence.
4. **Verification without restarting anything** (`TASK-260915-s4qkmr_verification.md`):
   - a FRESH login shell resolves the managed command: `zsh -lc 'command -v task-board; task-board --version'`;
   - the same for `task-board-tui` and `tb-sessiond` (`--version`/`--help`, no side effects);
   - the live processes are byte-identical and still running: same PIDs, same absolute paths, same
     sha256 of the resolved binaries, captured BEFORE and AFTER;
   - one hosted session still answers: a bounded read-only board call through the campaign binary.
5. **Residuals**: the seven legacy standalone binaries stay on disk (≈190 MB) — name the retirement
   condition (campaign finished, no process resolves them for N days) and leave them.

## Evidence discipline
Every command with its real exit code (state the shell, `set -o pipefail`). Before/after snapshots
are file listings with hashes, not prose. If any step cannot be done safely live, stop at that step,
record exactly what is blocked and why, and hand off — a partial, honest migration is the correct
outcome; a broken session is not.

Hand off with `task-board handoff TASK-260915-s4qkmr --role developer` after attaching the four
resources above.
