# Completion instruction — BUG-260921-3cgij4 (bound developer run, curator-spec)

Revision 2 is accepted (story_final of STORY-260921-atwkfi) and its exact candidate tree d7d83c7f landed on the curator-spec protected default as the signed commit 8e65374c5dcba2ae3ac9e0f869a2cf51361db52c (PR relux-works/curator-spec#76, all checks green, fast-forward push). The board lives in the separate repository relux-works/curator, so the Story closes through `worktree complete` run by YOU (the bound producer role). Rules: no board writes (no resources, notes, checklist, set_status) before or during. Run exactly, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator-spec:

    task-board worktree complete STORY-260921-atwkfi --cr BUG-260921-3cgij4 --revision 2 --landed-commit 8e65374c5dcba2ae3ac9e0f869a2cf51361db52c --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/complete-3cgij4-2.log

ONLY AFTER it exits: attach `.temp/complete-3cgij4-2.log` as outcome resource `BUG-260921-3cgij4_completion-results.md` and stop. If it refuses, the exact refusal is in the log; do not retry, do not handoff, no code changes.
