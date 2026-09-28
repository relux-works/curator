# TASK-260921-3qcjsy re-apply after the trunk base refresh (orchestrator, binding) — curator-spec

Revision 3 (rework 1 for the rev1 verdict) exists as CR-TASK-260921-3qcjsy-3 but its gate was killed
by the host three times (fresh `go test` binary SIGKILLed; the gate now retries that step). Meanwhile
curator-spec trunk moved to 8e65374c (BUG-260921-3cgij4, draft-sources-v1 fixtures — path-disjoint
from this leaf), so a story_final handoff from the old base 802caee would be refused
(`change_request_base_authority_mismatch`). The orchestrator captured your exact revision-3 delta as
`TASK-260921-3qcjsy_rev3-candidate.patch` (precondition; patch-id == the rev3 CR patch) and cleaned
the workspace so the runtime replays the (commit-less) Story branch onto trunk at your spawn.

Do exactly:
1. `git status` — clean; `git merge-base --is-ancestor 8e65374c HEAD` must succeed (if not, stop and
   attach `task-board worktree status STORY-260921-3z0fgr` as your outcome).
2. `git apply --index .task-board/.resources/TASK-260921-3qcjsy/TASK-260921-3qcjsy_rev3-candidate.patch`
   from the Story worktree root (fetch it via `task-board resource get` if absent); use `--3way` only
   if a hunk refuses; verify `git diff --cached --binary HEAD | git patch-id --stable` equals the
   patch's `git patch-id --stable` (report any per-file deviation); `git reset` to unstage.
3. `make validate` locally once (the configured gate wrapper reruns it), append "Revision 4 =
   revision 3 re-applied on trunk 8e65374c" to results.md, then handoff
   (`task-board handoff TASK-260921-3qcjsy --role developer`). No other changes.
