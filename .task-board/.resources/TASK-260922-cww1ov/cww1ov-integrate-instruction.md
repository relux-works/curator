# Integration instruction — STORY-260922-1cenbr via TASK-260922-cww1ov (bound developer run, curator)

Revision 2 is accepted and it is the story_final revision of STORY-260922-1cenbr (F-C1
TASK-260922-1t551d and F-C2 TASK-260922-1t2w1q are already checkpointed on the Story branch), so
this integrate lands the whole 0017 credential-modes implementation. Control root and board are both
the curator repository. Trunk has moved to 48da2690 (the rc.12 pin promotion) since the workspace was
provisioned — let the integrate reparent and revalidate; the gate is `sh scripts/remote-gate.sh`
(hosted, ~40 min). No board writes (no resources, notes, checklist, set_status) before or during.

Run exactly, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260922-1cenbr --cr TASK-260922-cww1ov --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1cenbr.log

ONLY AFTER it exits: attach `.temp/integrate-1cenbr.log` as outcome resource
`TASK-260922-cww1ov_integration-results.md` and stop. If it refuses, the exact refusal is in the
log; do not retry, do not handoff, change no file.
