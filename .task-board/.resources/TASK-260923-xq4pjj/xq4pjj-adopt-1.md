# TASK-260923-xq4pjj — adopt the task-board shim through `curator global adopt` (THE ONLY CURRENT INSTRUCTION; host-only)

Operator decision 2026-09-28: option (a). `curator global adopt <command>` landed on curator main (35472926, TASK-260928-2s0jsc). Your
earlier run already replaced ~/.local/bin/task-board with bytes equal to the canonical shim (backups under
~/.curator/backups/20260923T121543Z-shim-adoption). Do ONLY this, on the live host, touching nothing else:
1. Build curator from current origin/main into $TMPDIR (a fresh `git worktree add --detach $TMPDIR/cur origin/main` or `git archive`
   export; `go build -o $TMPDIR/curator ./cmd/curator`). Do NOT install it, do NOT replace ~/.curator binaries, do NOT run make install,
   setup.sh, task-board self-update or any daemon restart.
2. Re-verify first: `cmp` ~/.local/bin/task-board against the canonical shim bytes (as in your inventory); record sha256 of the shim and
   of the ownership marker; `cp -p` the marker to a new backup directory under ~/.curator/backups/.
3. `$TMPDIR/curator global adopt task-board --dry-run` (exit code + output), then the real `$TMPDIR/curator global adopt task-board`.
   If it refuses, STOP and report the exact diagnostic (do not hand-edit the marker, do not rewrite the shim).
4. Verify: the marker now records task-board (show the entry); the shim bytes are unchanged (cmp + sha256 identical to step 2); a fresh
   login shell resolves `task-board` and `task-board version` exits 0; live task-board/tb-sessiond PIDs, paths and hashes unchanged
   before/after (the process snapshots you used before); `$TMPDIR/curator global install --dry-run` (or status) no longer reports an
   unmanaged conflict for task-board. task-board-tui state reported only, not restored.
5. Check the remaining DoD items with evidence in the results (item 2's "via curator global install" is superseded by the operator decision:
   the product command is `global adopt`), update the results resource, handoff, END YOUR TURN. No repository change. No LOGBOOK.md.
