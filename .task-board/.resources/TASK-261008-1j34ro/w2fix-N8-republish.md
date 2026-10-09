# THE ONLY CURRENT INSTRUCTION — republish TASK-261008-1j34ro (N8) on current trunk (developer; no code changes)

Your rework-1 candidate has been carried forward twice without review: trunk advanced through the N9 landing (22cf8b4f) and then the 4mzun5 landing (37c19160, board record d7001974), which also change CHANGELOG.md, cmd/curator/main.go and internal/audit/audit.go. The orchestrator ran `task-board worktree converge STORY-261009-33rpbn`: the workspace now sits on d7001974 with your whole delta carried over; CHANGELOG.md adds only your entry, and cmd/curator/main.go and internal/audit/audit.go are clean three-way merges of your change with N9 and 4mzun5 (verified against dry-run merges). internal/hashing/hashing.go and your two test files are byte-identical. Revisions 2 and 3 are stale and were not reviewed.

Do exactly this:
1. In the Story worktree check `git status` and `git diff HEAD -- CHANGELOG.md cmd/curator/main.go internal/audit/audit.go`. Do not edit any file except to repair an obvious merge defect (there should be none).
2. Run compile-only checks: `go vet ./...` and `go build ./...` must pass with the merged files. Do NOT run `go test`, compiled test binaries or `go run` on this host (refused; they harm it).
3. Run the hosted green check for the merged candidate (a scratch branch from a disposable `git clone --shared`, CI must pass) and record the run URL in your results resource; your earlier red and ordering-mutant evidence stays valid because the production fix and tests are unchanged.
4. Republish with `task-board handoff TASK-261008-1j34ro --role developer`. If the handoff requires the task to be in development first, run `set_status(TASK-261008-1j34ro, status="development")` and retry once. If it refuses for any other reason, record the exact refusal in the task notes and stop.
5. END YOUR TURN. Do not edit LOGBOOK.md.
