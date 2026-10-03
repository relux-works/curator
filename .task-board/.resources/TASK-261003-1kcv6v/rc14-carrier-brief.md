# THE ONLY CURRENT INSTRUCTION — TASK-261003-1kcv6v: carry the accepted rc.14 lockstep onto the rewritten main

Under ruling tb-R162, curator main was rewritten: dfaa557f became f17ea733, with the rewrite starting at 5ce440c2. TASK-261002-2ipeqa rev4 was ACCEPTED (astra; candidate lane run 37017428885 green on 3 OSes), but its base 64345d71 is pre-rewrite history and can no longer be integrated.

Its exact delta is `refs/campaign/2ipeqa-rev4-20261003` (92ae3c19) in the control repo, on base 64345d71. It has 16 non-board files and 0 restricted-pattern hits.

Do:
1. In your fresh workspace on main f17ea733, run `git diff 64345d71 92ae3c19 -- . ':!.task-board' | git apply --3way`. Expect no conflicts: the rewrite touched only board and research files that this delta does not change.
2. Prove identity: for every one of the 16 paths, `git hash-object` of your file equals `git rev-parse 92ae3c19:<path>`. Record the table. If CHANGELOG.md differs only because the trunk context changed, explain it. Nothing else changes.
3. Never push anything yourself. Never reference or cherry-pick any pre-rewrite commit; apply the diff only.
4. Follow the attached host-rules.md (-work). Never edit LOGBOOK.md.

Then run `task-board handoff TASK-261003-1kcv6v --role developer` and END YOUR TURN.
