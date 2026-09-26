# TASK-260907-2as5sx — refresh onto trunk and publish (THE ONLY CURRENT INSTRUCTION)

Your §8.4 work is complete and sits uncommitted in the Story worktree (safety ref `refs/campaign/2as5sx-delta-20260924`,
commit 851dac33). Publication refused `change_request_base_authority_mismatch`: the Story checkpoint `50fad31f`
(TASK-260907-187z6x on base `09b25ef6`) does not descend from trunk `origin/main` (now 1511b345: rc.12 union 48da2690,
2qvzwk, 1bfk8y, 2kqa77 — 54 non-board paths). `converge` cannot move a diverged Story; the sanctioned exit is
`worktree refresh-candidate` from a live producer (you).

1. `task-board m 'set_status(TASK-260907-2as5sx, status=development)'`.
2. Combine trunk's incoming content INTO the working candidate first: `git diff 09b25ef6 origin/main -- . ':!.task-board'`
   applied with `git apply --3way` (never bring `.task-board/**` into the worktree). Resolve every overlap so BOTH sides
   survive: trunk's code/tests/ledgers plus your stateread migration — where trunk changed a reader you migrated, migrate
   trunk's version onto the seam too. Leave nothing staged.
3. `task-board worktree refresh-candidate TASK-260907-2as5sx`; on a checkpoint replay conflict (187z6x's commit) follow
   its own `--replay-resolutions` template exactly — never hand-commit the replay worktree.
4. Re-run your focused rows + the AST inventory test + `go vet` on touched packages (bounded calls); list any trunk
   reader you additionally migrated.
5. Append "Revision 1 — refresh onto <trunk>" to your results, `task-board resource update` the results file (a NEW/UPDATED
   outcome is required for the Change Request to be built), then `task-board handoff TASK-260907-2as5sx --role developer`;
   stay in the turn while the gate runs. A `run_wrote_outside_worktree … policy warn` block is a warning — verify
   status `to-review` and that revision 1 was published.
