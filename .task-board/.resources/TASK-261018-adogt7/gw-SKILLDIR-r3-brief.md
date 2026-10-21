# THE ONLY CURRENT INSTRUCTION — TASK-261018-adogt7: retry publication (developer)

Your previous run stopped correctly: the hosted gate failed only in "Complete runner" on the self-hosted runner relux-rise1-curator-4. The host cleanup hook could not remove a stale read-only pnpm store that an earlier job had left. That was infrastructure, not your candidate. The runner is healed: BUG-261018-bisu6n restored the tree's permissions, and on 2026-10-18 a job on curator-4 passed "Complete runner".

## Do this

1. Change no source, test or documentation file, unless the board reports that trunk moved under a path you change. In that case run `task-board worktree refresh-candidate TASK-261018-adogt7`, resolve only the combination, and re-run the affected checks. The rc.7 wave rule still holds: do not edit `CHANGELOG.md` or the root `README.md`; changelog text goes in the gaps file.
2. Confirm the tree equals your last signed candidate, or name the refresh delta.
3. Add a short "Publication retry" section to the gaps file: the candidate identity and what changed (or that nothing changed).
4. Do not add checklist items about the post-turn validation: it runs after your handoff and is the gate itself.

## Hand-off

Run `task-board handoff TASK-261018-adogt7 --role developer`, then END YOUR TURN. The configured validation reruns automatically.
