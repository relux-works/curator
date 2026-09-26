# TASK-260922-cww1ov — refresh onto trunk and publish (THE ONLY CURRENT INSTRUCTION)

Your accepted 0017 credential-modes work (revision 5, carried forward as revision 7) is complete and sits uncommitted in the Story worktree (safety ref `refs/campaign/cww1ov-delta-20260924`,
see the ref). Publication refused `change_request_base_authority_mismatch`: the Story checkpoint `84e2fb27`
(F-C1 TASK-260922-1t551d and F-C2 TASK-260922-1t2w1q checkpoints on base `fad88136`) does not descend from trunk `origin/main` (now 1511b345: rc.12 union 48da2690,
2qvzwk, 1bfk8y, 2kqa77 — 54 non-board paths). `converge` cannot move a diverged Story; the sanctioned exit is
`worktree refresh-candidate` from a live producer (you).

1. `task-board m 'set_status(TASK-260922-cww1ov, status=development)'`.
2. Combine trunk's incoming content INTO the working candidate first: `git diff fad88136 origin/main -- . ':!.task-board'`
   applied with `git apply --3way` (never bring `.task-board/**` into the worktree). Resolve every overlap so BOTH sides
   survive: trunk's code/tests/ledgers plus your stateread migration — where trunk changed a reader you migrated, migrate
   trunk's version onto the seam too. Leave nothing staged.
3. `task-board worktree refresh-candidate TASK-260922-cww1ov`; on a checkpoint replay conflict (the F-C1/F-C2 commits) follow
   its own `--replay-resolutions` template exactly — never hand-commit the replay worktree.
4. Re-run your 0017 rows (hazards, migration/recovery, no-copy scan) + `go vet` on touched packages (bounded calls); list any trunk
   reader you additionally migrated.
5. Append "Revision 8 — refresh onto <trunk>" to your results, `task-board resource update` the results file (a NEW/UPDATED
   outcome is required for the Change Request to be built), then `task-board handoff TASK-260922-cww1ov --role developer`;
   stay in the turn while the gate runs. A `run_wrote_outside_worktree … policy warn` block is a warning — verify
   status `to-review` and that a new revision was published.
