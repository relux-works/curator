# THE ONLY CURRENT INSTRUCTION — TASK-261018-adogt7: publish a fresh CR revision after orchestrator converge (developer)

Your candidate was not reviewed: CR revision 1 went stale because trunk advanced. PR #144 (BUG-261018-b2jnyg) landed on main at `69e71e38` and touched shared paths. The orchestrator has run `task-board worktree converge` on your Story; your uncommitted candidate is now carried onto fresh trunk `5ecb68da`.

## Do this

1. Run `task-board worktree status` for your Story. If it reports that only `task-board worktree refresh-candidate TASK-261018-adogt7` resolves the state, run it; otherwise do not.
2. Change no source, test or documentation file, except to resolve a real conflict with the newly landed code. If a carried path now overlaps #144 (for example `internal/envprofile/managed.go`), make sure the combined code builds and keeps both behaviours, and say so in the gaps file.
3. Check the carried tree: the tests that cover your changed paths, `go build ./...`, `go vet ./...`. The rc.7 wave rule holds: `CHANGELOG.md` and the root `README.md` must equal trunk.
4. In `TASK-261018-adogt7_gaps.md`, add a "Publication after converge onto 5ecb68da" section with the new base, the checks and their exit codes. If the file still has an open "Stop-The-Line" or "Blocker" section from an earlier round, retitle it "Resolved in an earlier round" and add one line on how it was resolved (the orchestrator removed the hosted-gate checklist rows; the handoff then succeeded). Nothing in the gaps file may read as a current blocker unless there is one.
5. Do not add checklist items about the post-turn validation: it runs after your handoff and is the gate itself.

## Hand-off

Run `task-board handoff TASK-261018-adogt7 --role developer`. It must publish a new CR revision on base `5ecb68da`; check `task-board worktree status` afterwards. If the handoff refuses, or no new revision appears, quote the exact output in the gaps file and stop. Then END YOUR TURN.
