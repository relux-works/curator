# TASK-260930-22sp8w — integration preconditions (bound developer run, rev 2)

Binding: `CR-TASK-260930-22sp8w-2` revision 2, role `developer`.
Per the integration assignment this run does NOT invoke `worktree integrate`
(the runner lands synchronously); it confirms preconditions, attaches this
evidence, and ends with the board left at `integrating` and no handoff call.

## Preconditions (all observed this run, read-only)

- Board: `TASK-260930-22sp8w` status `integrating`; `STORY-260930-12oimr`
  status `integrating` (via `task-board q 'get(...) { id status }'`).
- `worktree integrating` row:
  `TASK-260930-22sp8w  2  awaiting_landing  no  present  1h55m
  candidate tree is not carried by any post-base ancestor of
  refs/heads/main (landed_tree_not_on_trunk)`
  → accepted rev-2 candidate with a present delta, not yet on trunk.
  Protected trunk: `refs/heads/main at b4b08a1993d240b5dd24d6929cafe2b51e72637c`.
- `worktree obligations` row:
  `TASK-260930-22sp8w  2  accepted  checkpoint  1h55m  integrating
  STORY-260930-12oimr` → rev 2 accepted, checkpoint/land pending.
- Worktree: branch `task-board/story/STORY-260930-12oimr`, HEAD `bdb77413`
  (record base, no commits past checkpoint — work left UNCOMMITTED as required).
- Delta: 35 paths — 11 modified + 24 untracked — matching the accepted rev-2
  tree (35 paths). All tracked modifications are test/CI-gate files
  (`*_test.go`, `main_test.go`, `.github/ci/test-gate.sh`,
  `.github/ci/gate-selftest.sh`); untracked additions are the hostile-config
  regression test, per-package `gitenv_main_test.go` helpers, and
  `internal/testgitenv/testgitenv.go`. No production code changes.
- Spawn directives for this run: none (`No directives recorded`), so no
  pause/reroute to honor.
- Repo files changed by this run: none (read-only inspection only).

## Left for the runner

Synchronous `worktree integrate STORY-260930-12oimr --cr TASK-260930-22sp8w
--revision 2` landing and the `integrating` → `done` transaction.
