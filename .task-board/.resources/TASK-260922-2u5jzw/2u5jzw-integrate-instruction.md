# Integration instruction (THE ONLY CURRENT INSTRUCTION) — STORY-260922-39hxog via TASK-260922-2u5jzw (bound developer run, launcher)

Revision 1 (story_final) is accepted; TASK-260922-1zfqq0 (F-L1a, SPEC 0.5.0-draft) is already checkpointed on the Story branch. This
integrate lands the whole 0018 launcher permission interface. No board writes before or during. From
/Users/administrator/Developer/ReluxWorks/curator/curator-agent-launcher:

    task-board worktree integrate STORY-260922-39hxog --cr TASK-260922-2u5jzw --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-39hxog.log

Attach the log as `TASK-260922-2u5jzw_integration-results.md` and stop. If it refuses, attach the exact refusal and stop. Change no file.
