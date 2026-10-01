# TASK-261001-1klixs — carrier: land the accepted second-operator guide (THE ONLY CURRENT INSTRUCTION)

TASK-260930-o5uu7a rev1 was ACCEPTED: base 5ed5c4e1, tree ca2d4a1a, 2 paths, docs/second-operator.md and the README.md link. The review
spot-ran the guide literally on a clean HOME. It cannot land: `validation_suite_changed` after a task-board binary upgrade, and
invalidate-acceptance refuses because "no drift proven". Your Story worktree is fresh on trunk.
1. `task-board m 'set_status(TASK-261001-1klixs, status=development)'`.
2. Apply it unchanged:
   `git diff 5ed5c4e1 ca2d4a1a -- docs/second-operator.md README.md > $TMPDIR/so.patch && git apply --3way $TMPDIR/so.patch`.
   Show that `git diff origin/main --stat` lists exactly those 2 paths and that the +/- lines equal the accepted diff.
3. Change nothing else. No CHANGELOG/LOGBOOK edit. Never spell any employer name.
Update the results, then run `task-board handoff TASK-261001-1klixs --role developer`, then END YOUR TURN.
