# TASK-260919-2cmg0y re-apply after the trunk base refresh (orchestrator, binding)

Your previous runs (RUN-260921-f5253c) finished the work but the Change Request construction
failed with `change_request_base_authority_mismatch`: the Story branch (checkpoints on top of
c3f9eeea) no longer descends from trunk (main moved to d4fe8347 via PR #82/#83). The
orchestrator captured your exact candidate delta as
`TASK-260919-2cmg0y_rev1-candidate.patch` (precondition resource; 5 paths: CHANGELOG.md,
internal/transaction/{engine.go,journal.go,namespace.go,namespace_cache_test.go}) and cleaned
the workspace so the runtime can replay the Story checkpoints onto the fresh trunk at your
spawn. Your results.md is preserved on the task (rev1 snapshot).

Do exactly:
1. `git status` — confirm the workspace is clean and the Story tip descends from d4fe8347
   (`git merge-base --is-ancestor d4fe8347 HEAD`). If it does not, stop and attach the
   `worktree status` output as your outcome (do not try to rebase by hand).
2. `git apply --index .task-board/.resources/TASK-260919-2cmg0y/TASK-260919-2cmg0y_rev1-candidate.patch`
   (from the Story worktree root; if the patch is not at that path, fetch it with
   `task-board resource get TASK-260919-2cmg0y TASK-260919-2cmg0y_rev1-candidate.patch -o …`).
   Verify `git diff --cached --binary HEAD | git patch-id --stable` equals the patch's
   `git patch-id --stable`; then unstage (`git reset`) so the tree holds the bytes.
3. Re-run the fast checks (`go build ./...`, `go vet ./internal/transaction/`, `go test
   ./internal/transaction/ -count=1`, the sweep locally once) and append "Re-apply on
   d4fe8347" to results.md with the commands and exit codes — no other content changes.
4. Publish the Change Request and hand off; the configured gate reruns the suite.
Rulings R1–R4 of 2cmg0y-brief.md unchanged; no new scope.
