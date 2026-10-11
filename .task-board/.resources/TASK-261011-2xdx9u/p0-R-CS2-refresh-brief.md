# THE ONLY CURRENT INSTRUCTION — refresh TASK-261011-2xdx9u onto new trunk (developer)

Your revision 2 also edited `decisions/0021-sessions-enter-through-curator-run.md`. That file belongs to TASK-261011-kojgyg, whose amendment of Decision 0021 has now landed on curator-spec main (ddafa1ff). Your revision is therefore base-superseded.

Do this:
1. Run `task-board worktree refresh-candidate TASK-261011-2xdx9u` to move onto fresh trunk.
2. Drop your edits to `decisions/0021…`; trunk's version wins. If a CIP needs to refer to the amendment, link to it.
3. Keep all your other changes.
4. Hand off normally. The configured validation suite at handoff is sanctioned (compile-only Go). Do not run tests or builds yourself.

If refresh-candidate refuses, attach the exact output and stop. Then END YOUR TURN.
