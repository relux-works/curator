# TASK-260923-xq4pjj: adopt-hand-written-task-board-shim

## Description
Last mile of TASK-260915-s4qkmr (accepted 2026-09-23 by claude-opus-5-5). ~/.local/bin/task-board is a hand-written shim (unquoted exec), so curator global install refuses it as an unmanaged conflict (internal/globalbins unmanagedConflict + stage.go) and the product never adds task-board to .curator-managed.json. Replace it LIVE with the Curator-shaped bytes and let the product adopt it. Operator decision 2026-09-22: no quiet window, do it surgically on the live machine.

## Scope
Host state only: ~/.local/bin/task-board, the Curator-managed marker via the product command, ~/.curator/backups. No repository change, no daemon restart, no removal of any binary a live process resolves.

## Acceptance Criteria
1) current ~/.local/bin/task-board and the marker backed up with cp -p and sha256; 2) the shim replaced by an atomic temp-file move with bytes equal (cmp) to runtimestore.UnixShimContent for the canonical target; 3) curator global install adds task-board to the managed marker through the product (no hand edit); 4) fresh login shell resolves task-board, live PIDs/paths unchanged before and after; 5) task-board-tui state reported (removed at ~19:59Z by the operator's global reinstall) without restoring it unasked
