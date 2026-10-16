# Refresh your candidate onto current trunk (curator)

Your revision was built on an older curator trunk, and the trunk has since changed `README.md`: it added a pointer to the new agent logins reference, `docs/agent-logins.md`. Your uncommitted delta conflicts there, so the operator's converge refused.

Do this and nothing more:
1. Run `task-board worktree refresh-candidate` for this Story to move your candidate onto current curator main, carrying your delta.
2. Resolve the `README.md` conflict by keeping both: the trunk's new pointer and your own text, each once. Touch nothing else.
3. Re-run the focused tests and the full gate (`sh scripts/remote-gate.sh`) and list the exact commands and results.
4. Re-handoff for review as the board requires. The rules of rc7-code-common.md (attached) stay in force.
