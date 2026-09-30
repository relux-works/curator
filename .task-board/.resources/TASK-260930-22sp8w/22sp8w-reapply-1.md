# TASK-260930-22sp8w — re-apply rev2 on fresh trunk, no content change (THE ONLY CURRENT INSTRUCTION; bound developer run)

Rev2 was ACCEPTED. Its landing was refused as base-moved only because trunk (TASK-260930-38fjt0) also touched .github/ci/gate-selftest.sh,
which merges cleanly. Use `git apply --3way` for that file.

Your rev2 content is saved as `refs/campaign/12oimr-rev2-20260930` (parent bdb77413, tree 6fc60498, 35 paths). The orchestrator already
checked that it applies cleanly to current trunk. Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260930-22sp8w, status=development)'`.
2. Apply it, and change nothing else:
   `git diff bdb77413 refs/campaign/12oimr-rev2-20260930 -- . ':!.task-board' > $TMPDIR/gitiso.patch && git apply --3way $TMPDIR/gitiso.patch`.
3. Verify, and paste the output into the results:
   - `git diff --name-only origin/main -- . ':!.task-board' | wc -l` = 35;
   - `diff <(git diff bdb77413 refs/campaign/12oimr-rev2-20260930 -- . ':!.task-board' | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | sort) <(git diff origin/main -- . ':!.task-board' | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | sort)`
     must be empty.
4. Append "Revision 3 — rev2 re-applied unchanged on <trunk sha>" to the results, run `resource update`, then
   `task-board handoff TASK-260930-22sp8w --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate; do not wait
   for it. No CHANGELOG/LOGBOOK edit.
