# TASK-260924-10d3l1 — repair the stale candidate and republish (THE ONLY CURRENT INSTRUCTION; bound developer run)

Revision 3 CHANGES_REQUESTED (F1, verdict rev3): the Story worktree HEAD is trunk ab34556e, but its WORKING FILES are the old
a48f584c snapshot plus your one test change, so the candidate reverts 60 trunk files. The accepted content is ONLY
internal/install/draftevidence_test.go (per-file `git patch-id --stable` c7ce917d…, identical in rev1/rev2/rev3).
Safety ref of the current dirty state: refs/campaign/10d3l1-delta-20260925 (orchestrator made it; do not delete).
In /Users/administrator/Developer/ReluxWorks/curator/curator/.temp/STORY-260924-1oyh2m/worktree:
1. `task-board m 'set_status(TASK-260924-10d3l1, status=development)'` if needed.
2. Save the accepted file: `git diff HEAD -- internal/install/draftevidence_test.go > $TMPDIR/10d3l1.patch` is NOT enough (the working file
   was written against a48f584c); instead take the accepted change itself: `git diff a48f584c -- internal/install/draftevidence_test.go`
   from the rev2 patch resource `TASK-260924-10d3l1_change-request_rev2.patch` (path-limited) into $TMPDIR.
3. Restore EVERY path to HEAD: `git restore --source=HEAD --staged --worktree -- .` (this is the one sanctioned full restore — the
   dirty state is a converge artefact, saved in the safety ref), then `git clean -n` and remove only untracked files that the
   rev2 patch does not add (list them first; never touch .task-board).
4. Apply the accepted change: `git apply --3way` of the path-limited rev2 patch onto HEAD's draftevidence_test.go; if trunk changed that
   file since a48f584c (m28s6b did), keep BOTH sides. Result: `git diff --name-only HEAD` == internal/install/draftevidence_test.go only;
   report its per-file patch-id vs c7ce917d… (equal, or explain the merge).
5. `go test ./internal/install -run '^TestDraftEvidenceExactMatch$' -count=1` (real exit code). No CHANGELOG edit.
6. Append "Revision 4 — stale-candidate repair" to results, `resource update`, `task-board handoff TASK-260924-10d3l1 --role developer`;
   stay in the turn while the gate runs. A write-boundary `policy warn` block is a warning.
