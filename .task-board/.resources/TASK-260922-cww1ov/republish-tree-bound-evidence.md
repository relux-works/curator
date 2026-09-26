# Republish the SAME tree so the validation evidence is bound to it (bound developer run)

Why: the reviewer ACCEPTED this leaf's content, but `accept_cr` was refused by the new board binary
with `validation_not_bound_to_tree` — the Change Request was published by the previous binary, whose
validation evidence carries no source-tree identity. Nothing in the content is in question.

Do exactly this, and nothing else:
1. In the assigned Story worktree, confirm `git status --short` shows the same path set as the
   accepted revision (compare with the latest `*_change-request_rev*.patch` resource of this task).
   Change NO file.
2. Append one line to the task's results resource (or attach a small outcome resource if the
   results file is not yours to edit): "revision N+1 = revision N unchanged; republished under the
   new board binary so the validation evidence is tree-bound (validation_not_bound_to_tree)".
3. `task-board handoff <TASK-ID> --role developer` — the runtime reruns the landing suite once and
   publishes the new revision with tree-bound evidence.
If the handoff refuses for any reason, attach the exact refusal as an outcome resource and stop — do
not edit the checklist, do not check items by proxy, do not change code.

Note: a `run_wrote_outside_worktree … policy warn` block printed by handoff is a warning, not a refusal — check the board status and the new CR revision (see campaign-producer-rules.md).
