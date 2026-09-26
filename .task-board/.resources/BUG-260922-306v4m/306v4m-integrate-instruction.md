# Integration instruction (THE ONLY CURRENT INSTRUCTION — run integrate, NOT checkpoint) — STORY-260915-3w11un via BUG-260922-306v4m (bound developer run, curator)

Revision 3 is accepted; BUG-260922-306v4m is the last open leaf of STORY-260915-3w11un (its siblings were moved to their own Stories), so the
Story lands through integrate (checkpoint refused with change_request_final_leaf_checkpoint and said so).
Trunk has moved since the workspace was provisioned; let integrate reparent and revalidate. No board writes
before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260915-3w11un --cr BUG-260922-306v4m --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-306v4m.log

Attach the log as `BUG-260922-306v4m_integration-results.md` and stop. If it refuses (stale CR, write boundary, anything),
attach the exact refusal and stop — the orchestrator routes it. Change no file.
