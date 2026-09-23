# Integration instruction — STORY-260905-2qvzwk via TASK-260906-2b3nar (bound developer run, curator)

Revision 1 is accepted and it is the last live leaf of STORY-260905-2qvzwk (its sibling
TASK-260905-3r30t1 is done), so this integrate closes the Story. Control root and board are both the
curator repository; the control root's trunk was reconciled to 48da2690 (the rc.12 pin promotion),
so the workspace will reparent onto it and revalidate through `sh scripts/remote-gate.sh` (hosted,
~40 min). No board writes (no resources, notes, checklist, set_status) before or during.

Run exactly, from the control root /Users/administrator/Developer/ReluxWorks/curator/curator:

    task-board worktree integrate STORY-260905-2qvzwk --cr TASK-260906-2b3nar --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-2b3nar.log

ONLY AFTER it exits: attach `.temp/integrate-2b3nar.log` as outcome resource
`TASK-260906-2b3nar_integration-results.md` and stop. If it refuses, the exact refusal is in the log;
do not retry, do not handoff, change no file. (A `CHANGELOG.md` collision with trunk is the known
stale-CR refusal — report it, the orchestrator handles the release-and-refresh.)
