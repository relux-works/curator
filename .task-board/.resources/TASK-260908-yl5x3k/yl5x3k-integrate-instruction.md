# Integration instruction — TASK-260908-yl5x3k (bound developer run)

Revision 3 is accepted (empty repository delta; evidence-only). This board and code are colocated (curator control root), so the Story lands through `worktree integrate` run by YOU (the bound producer role). Rules: do NOT write anything to the board (no resources, notes, checklist ticks, set_status) before or during the integrate — a board write during the integration window is refused as board_delta_moved. Run exactly, in the FOREGROUND, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator, and wait for it to finish (the configured remote gate may run on GitHub, up to ~45 minutes; a single shell call may stay open that long here):

    task-board worktree integrate STORY-260908-2u6nly --cr TASK-260908-yl5x3k --revision 3 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-yl5x3k.log

ONLY AFTER it exits: attach `.temp/integrate-yl5x3k.log` as outcome resource `TASK-260908-yl5x3k_integration-results.md` and stop. If it refuses, the exact refusal is in the log; do not retry, do not handoff, no code changes.
