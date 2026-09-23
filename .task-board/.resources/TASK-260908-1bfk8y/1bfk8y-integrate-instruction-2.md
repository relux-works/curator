# Integration instruction 2 — TASK-260908-1bfk8y (bound developer run, curator)

The first integrate refused with `integration_base_moved` (head 09b25ef6, protected 48da2690): the
orchestrator landed the rc.12 pin promotion on trunk while the run was in flight. The accepted
revision is a ONE-FILE comment-only change to `.github/ci/platform-exclusions.tsv`, uncommitted in
the Story workspace, with no checkpoints on the Story branch — so the workspace can be re-parented
onto fresh trunk carrying that delta.

No board writes (no resources, notes, checklist, set_status) before or during. Run exactly, from the
control root /Users/administrator/Developer/ReluxWorks/curator/curator, in the foreground, in order:

    task-board worktree converge STORY-260907-2bddfc 2>&1 | tee .temp/converge-2bddfc.log
    task-board worktree integrate STORY-260907-2bddfc --cr TASK-260908-1bfk8y --revision 1 --commit-time "$(date -u +%Y-%m-%dT%H:%M:%SZ)" 2>&1 | tee .temp/integrate-1bfk8y-2.log

If `converge` refuses, run the integrate anyway and capture its refusal too. ONLY AFTER both exit:
attach both logs as outcome resources (`TASK-260908-1bfk8y_converge-results.md` and
`TASK-260908-1bfk8y_integration-results.md`) and stop. Do not retry, do not handoff, change no file.
If the integrate refuses again because the acceptance no longer holds against the moved base, attach
the exact refusal and stop — the orchestrator will route a re-review or a PR landing.
