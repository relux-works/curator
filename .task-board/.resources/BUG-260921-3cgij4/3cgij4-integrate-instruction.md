# Integration instruction — BUG-260921-3cgij4 (bound developer run, curator-spec)

Revision 2 is accepted and it is the story_final revision of STORY-260921-atwkfi (its only leaf). The board lives in the curator repository and the code in curator-spec (this control root), so the Story lands through `worktree integrate` run by YOU (the bound producer role). If trunk moved since the review, let the integrate reparent and revalidate (the gate is local `make validate`, ~25 minutes). Rules: no board writes (no resources, notes, checklist, set_status) before or during the integrate. Run exactly, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator-spec:

    task-board worktree integrate STORY-260921-atwkfi --cr BUG-260921-3cgij4 --revision 2 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-3cgij4-2.log

ONLY AFTER it exits: attach `.temp/integrate-3cgij4-2.log` as outcome resource `BUG-260921-3cgij4_integration-results.md` and stop. If it refuses, the exact refusal is in the log; do not retry, do not handoff, no code changes.
