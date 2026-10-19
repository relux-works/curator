# THE ONLY CURRENT INSTRUCTION — BUG-261018-dbeghc: publish after orchestrator converge (developer)

Your previous run stopped correctly: `refresh-candidate` was refused. The orchestrator has now run `task-board worktree converge` on your Story. Your uncommitted candidate is carried onto fresh trunk `892dab09`, which includes BUG-261018-bisu6n (the runner tmp self-heal) and the latest board state.

## Do this

1. Do not run `refresh-candidate` unless `task-board worktree status` still reports a refusal that only it resolves. Change no source, test or documentation file.
2. Check the carried tree: run the checks that cover your changed paths, plus `go build ./...`. The rc.7 wave rule holds: `CHANGELOG.md` and the root `README.md` must equal trunk.
3. Add a short "Publication after converge" section to the gaps file: the new base, the checks and their exit codes.
4. Do not add checklist items about the post-turn validation: it runs after your handoff and is the gate itself.

## Hand-off

Run `task-board handoff BUG-261018-dbeghc --role developer`, then END YOUR TURN. The configured validation reruns automatically.

Round note: the previous attempt prepared the evidence but did not complete the handoff. The orchestrator has removed the checklist row that asked for the hosted gate to be green before handoff: that gate is the post-turn validation, which runs after your handoff. Tick the remaining rows only where they are true now: "Implementation matches AC", "Tests green" and similar rows from your own local evidence. "If review does not accept the work" is a branch instruction for the review phase; it is satisfied when you route through the normal handoff. Then run the handoff. If the handoff still refuses, quote the exact refusal in the gaps file and stop.
