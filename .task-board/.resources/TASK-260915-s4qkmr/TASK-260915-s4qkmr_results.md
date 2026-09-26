# TASK-260915-s4qkmr — Results (migrate-global-cli-ownership, live machine)

Role: developer. task_class=metadata: no Story workspace, no Change Request;
deliverable is host state + board evidence. All commands from zsh with
`set -o pipefail`; every cited command exited 0 unless noted.

## Outcome: partial, honest migration — stopped at managed publication with evidence

| Step | Result | Evidence resource |
|---|---|---|
| 1. Inventory | done — 15 entries in ~/.local/bin (3 shims, 10 standalone binaries, curator, marker) + 3 global shims, all hashed, all references traced | TASK-260915-s4qkmr_inventory.md |
| 2. Backup | done — 17 `cp -p` copies + MANIFEST.txt under `~/.curator/backups/20260922T165929Z-legacy-task-board/`, all sha256-verified, nothing moved/removed | TASK-260915-s4qkmr_backup-manifest.md |
| 3. Managed publication | STOPPED with evidence — product refuses the hand-written shim by design; real install would rewrite machine-wide state incl. forbidden `global/bin` in-place edits; zero progress for real risk. No write performed | TASK-260915-s4qkmr_managed-publication.md |
| 4. Verification without restart | done — fresh login resolves all 3 commands; all 7 PIDs/paths identical; all 18 hashes identical; hosted board call answers via campaign binary | TASK-260915-s4qkmr_verification.md |

Net host delta: one new backup dir (262M). Zero bytes modified or removed
anywhere else. `task-board-main-6cb09a23-curatorlike` untouched
(`18a02361…` before == after), recorded residual with retirement condition
(campaign finished + N days unreferenced; suggest N=7; re-backup then remove
in a quiet window only).

## Findings / deviations from the brief (logbook-substitute)

Per campaign-producer-rules producers do not edit LOGBOOK.md; these stand as
the task's logbook record in its outcome resources:

1. Brief says 7 legacy standalone binaries; host holds 10 (3 extra
   `task-board-main-e8205e78*`). All 10 backed up.
2. Brief's `~/.local/bin` pattern missed `tb-sessiond` (matches neither
   "task-board" nor "curator"). Inventoried, backed up, verified managed.
3. `~/.local/bin/task-board` is hand-written (unquoted exec, mtime Sep 17, one
   day newer than the Curator-generated pair) — the root cause of the product's
   unmanaged-conflict refusal.
4. The 3 live daemon images are open-but-unlinked mappings (original build dirs
   AND `.sweep-*` dirs absent from disk — Curator cache GC). Bytes unhashable;
   reported unknown with PID/lstart/argv/lsof identity proof. Pre-existing,
   unchanged by this task.
5. `curator global install --dry-run` returns before forwarding staging
   (`internal/install/global.go:302`), so dry-run cannot preview shim
   refusal; refusal proven by source + byte comparison instead.
6. No launchd job references any in-scope binary; no board/manifest config
   references binary paths — the campaign binding is live-argv absolute paths.

## Checklist rationale

- Items 1–4: done per resources above. Item 3 via the explicit stop-with-evidence branch.
- Item 5 ("code written"): task is metadata/host-state only per brief; no repo
  change permitted. Satisfied by the host-state + evidence deliverables.
- Item 6: satisfied by this + the 4 evidence resources (all `TASK-260915-s4qkmr_*`).
- Item 7: findings recorded above (LOGBOOK.md not producer-editable per campaign rules).
