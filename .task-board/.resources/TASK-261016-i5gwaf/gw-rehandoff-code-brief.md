# Re-handoff after trunk convergence (curator)

Your revision was built on an older curator trunk. The operator ran `task-board worktree converge` on this Story: the workspace now sits on current curator main and your uncommitted delta was carried over by a three-way merge. Trunk gained the agent logins reference: `docs/agent-logins.md`, four PlantUML diagrams, and pointer edits in `README.md` and `docs/environment-config.md`.

Do this and nothing more:
1. Read the merged result in every file that both the trunk change and your delta touch (at least `README.md`). Make sure your text still reads correctly next to the new trunk text: no duplicated or contradicting lines, links resolve. Fix only what the merge made wrong.
2. Re-run the focused tests and the full gate from your previous handoff (`sh scripts/remote-gate.sh`), and list the exact commands and results.
3. Re-handoff the task for review as the board requires. The rules of rc7-code-common.md (attached) stay in force.
