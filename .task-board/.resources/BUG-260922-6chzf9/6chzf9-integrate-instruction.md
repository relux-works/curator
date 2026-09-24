# Integration instruction (THE ONLY CURRENT INSTRUCTION — run integrate, NOT checkpoint) — STORY-260923-vkxt08 via BUG-260922-6chzf9 (bound developer run, curator)

Revision 1 is accepted; BUG-260922-6chzf9 is the last open leaf of STORY-260923-vkxt08 (its siblings were moved to their own Stories), so the
Story lands through integrate (checkpoint refused with change_request_final_leaf_checkpoint and said so).
Trunk has moved since the workspace was provisioned; let integrate reparent and revalidate. No board writes
before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260923-vkxt08 --cr BUG-260922-6chzf9 --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-6chzf9.log

Attach the log as `BUG-260922-6chzf9_integration-results.md` and stop. If it refuses (stale CR, write boundary, anything),
attach the exact refusal and stop — the orchestrator routes it. Change no file.
