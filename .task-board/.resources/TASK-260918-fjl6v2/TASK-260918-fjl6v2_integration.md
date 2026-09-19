# TASK-260918-fjl6v2 integration — NOT LANDED (local gate-poller SIGKILLed twice)

Story `STORY-260918-2yvd86` (conformance-pin-rc12-landing), final leaf
`TASK-260918-fjl6v2` revision 1 (accepted). Board-owner path:
`worktree integrate`. Run at 2026-09-18 ~20:15–21:15 UTC.

## Outcome

**No landing occurred.** `task-board worktree integrate
STORY-260918-2yvd86 --cr TASK-260918-fjl6v2 --revision 1` was run twice
(initial + the one sanctioned re-run); both refused with the identical
typed error `revalidation_failed`, exit 1, `exit_status: 137`. Per the
run instruction ("re-run once first, then report") no further attempts
were made. No squash commit, no board commit, no status change performed
by this run (task and story remain `integrating`).

## Failure signature (verbatim, attempt 1 — attempt 2 identical modulo branch/run id)

```text
revalidation_failed: the validation suite failed on the tree 094dfcc5c8edba97f0841a0ca074766d4819d5c3 that would land; trunk is unchanged, board status is unchanged and no integration phase was entered
  candidate_tree_oid: 094dfcc5c8edba97f0841a0ca074766d4819d5c3
  element_id: TASK-260918-fjl6v2
  exit_status: 137
  log: $ sh scripts/remote-gate.sh
remote gate: attempt 1/3: pushing ad91a9f0128818281279786e95ab2b2f2aa5792a as gate/STORY-260918-2yvd86/260918-201615-17306-1
...
remote gate: run 35390532501 (https://github.com/relux-works/curator/actions/runs/35390532501)
scripts/remote-gate.sh: line 102: 31762 Killed: 9               sleep "$POLL"

exit 137
```

Attempt 2 (re-run, PID 33383): pushed `37fe58417a45e9823888a195a61e5f4f605ed811`
as `gate/STORY-260918-2yvd86/260918-205149-33742-1`, run
`35393742345`; `sleep "$POLL"` was `Killed: 9` three times
(script lines 102, 57, 57), exit 137. Same
`candidate_tree_oid: 094dfcc5c8edba97f0841a0ca074766d4819d5c3`.

Cause class: **local infrastructure, not a gate verdict.** The
`remote-gate.sh` poll loop's `sleep` children are being SIGKILLed by
something on this host (host heavily loaded: many concurrent
spawn-runners). The integrate process itself held
`.temp/integration/repository.lock` and was otherwise idle (~3 s CPU
over 30+ min). This is not `integration_base_moved` and not a red
gate — no transaction phase was entered, so per the run instruction
no `--rollback` applies.

## Gate evidence: the landing tree is green

- Attempt-2 gate branch tree verified locally:
  `git rev-parse origin/gate/STORY-260918-2yvd86/260918-205149-33742-1^{tree}`
  = `094dfcc5c8edba97f0841a0ca074766d4819d5c3` — exactly the
  `candidate_tree_oid`, i.e. the runs test precisely the tree that
  would land. (Attempt-1 gate branch has since been deleted from
  origin; both attempts named the same candidate tree with trunk
  unmoved throughout, so both runs tested the same tree.)
- Run `35390532501` (attempt 1): **completed / success** — all lanes
  green on the landing tree (Test ubuntu/macos/windows, Race
  ubuntu/macos, Lint, Naming, Interop, gate self-tests; rose-air +
  candidate-suite skipped as usual).
- Run `35393742345` (attempt 2): still `in_progress` when the poller
  died (fast lanes already green; Windows lane pending, the known
  slow lane).

## Trunk / board state after both attempts (nothing landed)

- `git log --oneline -4` (control root): head still `6f17377`
  ("Record board state: 1ll22r and 12lbww spec revisions landed,
  manager leaves created"); no new commits.
- `git status --short` excluding `.task-board`: empty — no
  non-board changes.
- `task-board worktree obligations`: row
  `TASK-260918-fjl6v2  1  accepted  checkpoint  integrating  STORY-260918-2yvd86`
  still present (final leaf awaiting integration).
- `get(STORY-260918-2yvd86) { status children }` =
  `{"children":["TASK-260918-fjl6v2"],"status":"integrating"}`;
  `get(TASK-260918-fjl6v2) { status }` = `{"status":"integrating"}`.

## Hand-back to orchestrator

1. Re-run `task-board worktree integrate STORY-260918-2yvd86 --cr
   TASK-260918-fjl6v2 --revision 1` from the control root once the
   host stops SIGKILLing short-lived `sleep` processes (or from a
   quieter host). No code, board, or revision action is needed:
   revision 1 is accepted and its tree is gate-green
   (run 35390532501, completed/success).
2. Optional: run `35393742345` may complete green on its own; it
   cannot be reused by a later integrate (each integrate pushes a
   fresh gate branch), but its verdict corroborates the tree.
3. Suggested host follow-up (out of this run's scope): identify what
   sends SIGKILL to `sleep` children (memory-pressure jetsam,
   reaper, or a concurrent cleanup) — it will strike any long
   `remote-gate.sh` poll, not just this story.

## What this run did not do (per the run instruction)

No repair, no reset, no manual commit, no board status change, no
`--rollback` (no `prepared` phase), no push, no `reconcile-trunk`,
no code changes.
