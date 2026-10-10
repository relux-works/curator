# THE ONLY CURRENT INSTRUCTION — rework TASK-261010-1992si to rev2: candidate scope only (researcher, mechanical)

The rev1 review (`TASK-261010-1992si_review-verdict-rev1.md`) accepted the study's content and found one blocking defect. The candidate delta also contains the sibling study `.research/261010_modular-instructions-design.md` of TASK-261010-2uqd3t. That study is already checkpointed on this Story branch by its own task.

Do exactly this:
1. Rebuild this task's candidate so that its delta, measured from the Story checkpoint that already carries the sibling study, contains only `.research/261010_project-surfaces-coverage.md`.
2. Keep that file byte-identical to rev1, and keep the sibling study where its own task put it. Do not delete shared work to narrow this candidate.
3. Republish as rev2 through the normal producer flow, then `task-board handoff TASK-261010-1992si --role researcher` and END YOUR TURN.

No content changes, no `LOGBOOK.md` edits, no tests or builds on this host, no hand edits of board files. If the board's commands cannot express a checkpoint-relative candidate, stop and record the exact refusal.
