# TASK-260910-32gki6 — re-apply rev7 on fresh trunk, no content change (THE ONLY CURRENT INSTRUCTION; bound developer run)

Rev7 was green on every lane except the Naming gate. That failure came from a trunk-side problem, now fixed on trunk. The workspace
record kept a stale upstream anchor, so it was discarded.

Your rev7 content is saved as `refs/campaign/148pj1-rev7-20260929` (parent fc499a96, tree 0bab9fd8, 31 paths). The orchestrator already
checked that it applies cleanly to current trunk. Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260910-32gki6, status=development)'`.
2. Apply it, and change nothing else:
   `git diff fc499a96 refs/campaign/148pj1-rev7-20260929 -- . ':!.task-board' > $TMPDIR/s5.patch && git apply $TMPDIR/s5.patch`.
3. Verify, and paste the output into the results:
   - `git diff --name-only origin/main -- . ':!.task-board' | wc -l` = 31;
   - `diff <(git diff fc499a96 refs/campaign/148pj1-rev7-20260929 -- . ':!.task-board' | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | sort) <(git diff origin/main -- . ':!.task-board' | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | sort)`
     must be empty.
4. Append "Revision 8 — rev7 re-applied unchanged on <trunk sha>" to the results, run `resource update`, then
   `task-board handoff TASK-260910-32gki6 --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate; do not wait
   for it. No CHANGELOG/LOGBOOK edit.
