# THE ONLY CURRENT INSTRUCTION — TASK-261018-adogt7: publish round 8 (developer)

Round 8 was complete but could not publish: the ivmbp agent keys were unavailable. They are back (2026-10-20 20:00Z); `ssh-add -l` on the agent socket lists them, and `agent-git-env` is sourced by the launcher. Trunk has not moved: it is still `e88fdf2d`, your checkpoint.

## Do this

1. Change no source, test or documentation file. Confirm with `git status --short` that the candidate is the round-8 tree: compare against the source-scope SHA-256 `52db2b21…` recorded in the round-8 gaps section.
2. Run only the quick checks, to confirm the tree after the reboot: the round-8 numeric tests, the directory mutant table, `go build ./...` and `go vet ./...`. The round-8 full local suite stands for this unchanged tree; say so, with the scope hash, rather than rerunning it.
3. In `TASK-261018-adogt7_gaps.md`, add a short "Publication after key restore" section: the scope-hash comparison and the checks with their exit codes. Retitle the round-8 sentence that calls publication "currently blocked by unavailable required agent keys" as resolved.
4. Do not run `refresh-candidate` unless `worktree status` reports a state only it resolves. Do not add checklist items about the post-turn validation.

## Hand-off

Run `task-board handoff TASK-261018-adogt7 --role developer`; it must publish CR revision 4. If it refuses, quote the exact output in the gaps file and stop. Then END YOUR TURN.
