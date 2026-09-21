# TASK-260919-2cmg0y re-apply after the trunk base refresh (orchestrator, binding)

Your previous runs (RUN-260921-f5253c, RUN-260921-897e9f) finished the work but the Change Request construction
failed with `change_request_base_authority_mismatch`: the Story branch (checkpoints on top of
c3f9eeea) no longer descends from trunk (main moved to d4fe8347 via PR #82/#83). The
orchestrator captured your exact candidate delta as
`TASK-260919-2cmg0y_rev1-candidate.patch` (precondition resource; 5 paths: CHANGELOG.md,
internal/transaction/{engine.go,journal.go,namespace.go,namespace_cache_test.go}) and cleaned
the workspace; the first replay attempt conflicted on CHANGELOG.md, which is now resolved by a
repository-local `merge=union` attribute for CHANGELOG.md, so the runtime replays the Story
checkpoints onto the fresh trunk at your spawn (simulated clean: 7 commits, +5 CHANGELOG lines). Your results.md is preserved on the task (rev1 snapshot).

Do exactly:
1. `git status` — confirm the workspace is clean and the Story tip descends from d4fe8347
   (`git merge-base --is-ancestor d4fe8347 HEAD`). If it does not, stop and attach the
   `worktree status` output as your outcome (do not try to rebase by hand).
2. `git apply --index .task-board/.resources/TASK-260919-2cmg0y/TASK-260919-2cmg0y_rev1-candidate.patch`
   (from the Story worktree root; if the patch is not at that path, fetch it with
   `task-board resource get TASK-260919-2cmg0y TASK-260919-2cmg0y_rev1-candidate.patch -o …`).
   The trunk refresh merged trunk's own CHANGELOG entries into the Story branch, so the
   CHANGELOG hunk may need `git apply --3way` (or, if it still refuses, re-add your single
   CHANGELOG entry by hand at the same place under `## Unreleased` → `### Changed`); every
   OTHER path must apply cleanly — verify `git diff --cached --binary HEAD -- . ':!CHANGELOG.md'
   | git patch-id --stable` equals `git patch-id --stable` of the patch filtered the same way
   (`git apply --exclude=CHANGELOG.md`-equivalent: use `filterdiff`-free approach: apply the
   patch fully, then compare per-file with `git diff --cached -- <path> | git patch-id`).
   Then unstage (`git reset`) so the tree holds the bytes.
3. Re-run the fast checks (`go build ./...`, `go vet ./internal/transaction/`, `go test
   ./internal/transaction/ -count=1`, the sweep locally once) and append "Re-apply on
   d4fe8347" to results.md with the commands and exit codes — no other content changes.
4. Publish the Change Request and hand off; the configured gate reruns the suite.
Rulings R1–R4 of 2cmg0y-brief.md unchanged; no new scope.
