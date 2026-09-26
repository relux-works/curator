# Integration instruction — STORY-260922-1cenbr via TASK-260922-cww1ov (bound developer run, curator)

Revision 5 (tree-bound republish, identity-reviewed) is accepted and it is the story_final revision of STORY-260922-1cenbr (F-C1
TASK-260922-1t551d and F-C2 TASK-260922-1t2w1q are already checkpointed on the Story branch), so
this integrate lands the whole 0017 credential-modes implementation. Control root and board are both
the curator repository. Trunk is now 6c19e5ee (or later: 1bfk8y may land first) since the workspace was
provisioned — let the integrate reparent and revalidate; the gate is `sh scripts/remote-gate.sh`
(hosted, ~40 min). No board writes (no resources, notes, checklist, set_status) before or during.

Run exactly, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260922-1cenbr --cr TASK-260922-cww1ov --revision 5 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1cenbr-3.log

ONLY AFTER it exits: attach `.temp/integrate-1cenbr-3.log` as outcome resource
`TASK-260922-cww1ov_integration-results.md` and stop. If it refuses, the exact refusal is in the
log; do not retry, do not handoff, change no file.
