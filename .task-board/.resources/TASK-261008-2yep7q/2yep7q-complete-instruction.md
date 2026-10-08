# Completion instruction — TASK-261008-2yep7q (bound developer run, curator-spec)

Revision 4 is accepted (CR-TASK-261008-2yep7q-4, exact candidate tree 7bb3b5d0e2819e73ff1e8aa18b7aa2e2fc759ca2) and that exact tree landed on the curator-spec protected default as the signed commit 7eaeb73fcf22cb8ce8e21325a8237accdfaf9646 (PR relux-works/curator-spec#135, all checks green, fast-forwarded by the orchestrator). No board writes before or during. From /Users/administrator/Developer/ReluxWorks/curator/curator-spec run exactly:

    task-board worktree complete STORY-260925-2nat7f --cr TASK-261008-2yep7q --revision 4 --landed-commit 7eaeb73fcf22cb8ce8e21325a8237accdfaf9646 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/complete-2yep7q-4.log

ONLY AFTER it exits: attach `.temp/complete-2yep7q-4.log` as outcome resource `TASK-261008-2yep7q_completion-results.md` and stop. If it refuses, the exact refusal is in the log; do not retry, do not handoff, no code changes.
