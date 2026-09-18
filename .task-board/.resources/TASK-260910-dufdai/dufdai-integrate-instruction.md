# Integration instruction — TASK-260910-dufdai (bound developer run)

Revision 6 is accepted and it is the story_final revision of STORY-260910-20sx61 (the two sibling checkpoints hwxr26 rev10 and 17ps6u rev4 plus this leaf). This board and code are colocated (curator control root), so the Story lands through `worktree integrate` run by YOU (the bound producer role). If trunk moved by board-only commits since the review, let the integrate reparent and revalidate (the remote gate runs on GitHub, up to ~45 minutes; the shell call may stay open that long). Rules: no board writes (no resources, notes, checklist, set_status) before or during the integrate. Run exactly, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260910-20sx61 --cr TASK-260910-dufdai --revision 6 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-dufdai-6.log

ONLY AFTER it exits: attach `.temp/integrate-dufdai-6.log` as outcome resource `TASK-260910-dufdai_integration-results.md` and stop. If it refuses, the exact refusal is in the log; do not retry, do not handoff, no code changes.
