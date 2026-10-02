# THE ONLY CURRENT INSTRUCTION — BUG-260923-2afgyq base refresh → revision 2 (orchestrator brief, binding)

Revision 1 was ACCEPTED, but its landing refused with integration_base_moved: trunk f40b77c1 (TASK-261002-9w4wy3, Windows hard links) changed `.github/ci/conformance-gaps.tsv` and `CHANGELOG.md`, which rev1 also changes. The acceptance is released. This is a base refresh only.

1. Run `task-board worktree refresh-candidate BUG-260923-2afgyq`.
   - On a conflict, follow `--replay-resolutions`; never hand-commit.
   - The gaps tsv must have trunk's two hardlink rows removed AND your five marker rows removed, with exact counts.
   - CHANGELOG keeps both lines.
2. Prove the refreshed candidate equals rev1 except for that merge. Run `go test ./internal/marker ./internal/conformancecoverage -count=1` and record the real exit code.
3. Add a "Revision 2 (refresh)" section to the results. Then run `task-board handoff BUG-260923-2afgyq --role developer` and END YOUR TURN.

No content change. Never edit LOGBOOK.md. Never spell any employer name.
