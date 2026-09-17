# Integration instruction — TASK-260910-19w2aj (bound developer run)

Revision 11 is accepted (story_final; the 1a75qd checkpoint plus this leaf). This board and code are colocated (curator control root), so the Story lands through `worktree integrate` run by YOU (the bound producer role). Trunk moved by board-only commits since the review; the installed task-board build now tolerates the transaction's own revalidation writes, so let the integrate reparent and revalidate (the remote gate runs on GitHub, up to ~45 minutes; a single foreground shell call may stay open that long here). Rules: no board writes (no resources, notes, checklist, set_status) before or during the integrate. Run exactly, in the FOREGROUND, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260910-3vxe3y --cr TASK-260910-19w2aj --revision 11 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-19w2aj-2.log

ONLY AFTER it exits: attach `.temp/integrate-19w2aj-2.log` as outcome resource `TASK-260910-19w2aj_integration-results.md` and stop. If it refuses, the exact refusal is in the log; do not retry, do not handoff, no code changes.
