# Review note for TASK-260915-s4qkmr (orchestrator, binding) — live CLI-ownership migration (metadata)

`task_class=metadata`: no Change Request exists, so there is no `accept_cr`. Review the EVIDENCE and
route the task with `set_status` (`done` if the brief's safety rules held and the stop was correct,
`to-dev` with findings otherwise). Read `s4qkmr-brief.md` and the four resources:
`TASK-260915-s4qkmr_inventory.md`, `_backup-manifest.md`, `_managed-publication.md`,
`_verification.md`, plus `_results.md`.

Judge, independently and read-only (do NOT modify anything under `~/.local/bin`, `~/.curator`, or any
live process):
1. **Safety rules held**: no move/rename/delete/overwrite of a live-resolved path; the campaign binary
   `task-board-main-6cb09a23-curatorlike` untouched; no install/setup/self-update/daemon action.
   Recompute sha256 of the in-scope files now and compare with the manifest and the before/after
   tables; confirm the live PIDs recorded still match `ps`.
2. **Backups are real**: the copies under `~/.curator/backups/<stamp>-legacy-task-board/` exist, and
   each copy's sha256 equals its source.
3. **The stop was right**: the producer stopped at managed publication because
   `~/.local/bin/task-board` is a hand-written shim that `curator global install` would refuse as an
   unmanaged conflict, and running the install would stage machine-wide global state on a live
   machine. Check that claim against the code it cites (`internal/install/global.go`,
   `internal/globalbins/stage.go`) — a wrong reason for stopping is a finding.
4. Verification: a fresh login shell resolves the three managed commands; nothing was restarted.

Name the remaining "last mile" precisely (replace the hand-written shim with a Curator-shaped
managed one and add the marker entry, with backup and the same before/after verification) so the
orchestrator can file it as the follow-up leaf. No LOGBOOK.md writes.
