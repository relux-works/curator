# TASK-260910-2vnjej — re-apply rev7 on fresh trunk, no content change (THE ONLY CURRENT INSTRUCTION; bound developer run)

Rev7 was ACCEPTED (identity review). Its landing was refused as base-moved only because trunk also touched .github/ci/platform-cases.tsv,
which merges cleanly; S5 has landed since.

Your rev7 content is saved as `refs/campaign/6bo7ej-rev7-20260929` (parent f30c2b34, tree 67626241, 26 paths). The orchestrator already
checked that it applies cleanly to current trunk. Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-260910-2vnjej, status=development)'`.
2. Apply it, and change nothing else:
   `git diff f30c2b34 refs/campaign/6bo7ej-rev7-20260929 -- . ':!.task-board' > $TMPDIR/s2.patch && git apply $TMPDIR/s2.patch`.
3. Verify, and paste the output into the results:
   - `git diff --name-only origin/main -- . ':!.task-board' | wc -l` = 26;
   - `diff <(git diff f30c2b34 refs/campaign/6bo7ej-rev7-20260929 -- . ':!.task-board' | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | sort) <(git diff origin/main -- . ':!.task-board' | grep -E '^[+-]' | grep -vE '^(\+\+\+|---)' | sort)`
     must be empty.
4. Append "Revision 8 — rev7 re-applied unchanged on <trunk sha>" to the results, run `resource update`, then
   `task-board handoff TASK-260910-2vnjej --role developer`, then END YOUR TURN. The runner publishes the CR and runs the gate; do not wait
   for it. No CHANGELOG/LOGBOOK edit.
