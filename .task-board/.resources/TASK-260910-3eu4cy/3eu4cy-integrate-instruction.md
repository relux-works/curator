# Integration instruction — TASK-260910-3eu4cy (bound developer run)

Revision 3 is accepted and it is the story_final revision of STORY-260910-1s75e1 (the sibling checkpoint 1xs0pj rev3 plus this leaf). This board and code are colocated (curator control root), so the Story lands through `worktree integrate` run by YOU (the bound producer role). If trunk moved by board-only commits since the review, let the integrate reparent and revalidate (the remote gate runs on GitHub, up to ~45 minutes; the shell call may stay open that long). Rules: no board writes (no resources, notes, checklist, set_status) before or during the integrate. Run exactly, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260910-1s75e1 --cr TASK-260910-3eu4cy --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3eu4cy-3.log

ONLY AFTER it exits: attach `.temp/integrate-3eu4cy-3.log` as outcome resource `TASK-260910-3eu4cy_integration-results.md` and stop. If it refuses, the exact refusal is in the log; do not retry, do not handoff, no code changes.
