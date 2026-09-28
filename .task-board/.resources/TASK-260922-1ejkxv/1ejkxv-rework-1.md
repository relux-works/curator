# TASK-260922-1ejkxv — continuation 1: finish the aggregate gate in bounded parts

The orchestrator has reviewed your Stop-The-Line packet. It is not an external blocker: the
work is done and the only missing evidence is the aggregate `make validate`, which does not fit
one ~10-minute shell call on this host (load average ~100). The orchestrator accepts the
aggregate as the **ordered composition of its three recipe lines**, each run in its own bounded
call, with the unittest line split per test file. This is the gate, not a substitute for it.

## Do
1. `task-board m 'set_status(TASK-260922-1ejkxv, status=development)'` (leave `blocked`).
2. Leave the staged/unstaged tree exactly as it is (do not reset, do not re-stage broadly).
3. Run, each as its own shell call, in the Story worktree, with the same interpreter set-up you
   used before (`PATH=.temp/task-260922-1ejkxv-venv/bin:$PATH`):
   - `python3 tools/validate.py`
   - for every `tools/test_*.py` file F, one call:
     `python3 -B -m unittest discover -s tools -p "$F"` — if a single file still exceeds the
     bound, split it further with `-k <TestClass>` per class (list classes with
     `grep -n '^class .*TestCase' tools/$F`), and record the split.
   - `go test ./tools/...`
   - `make regenerate-check`
4. A call that is interrupted (exit 130) is re-run once in smaller pieces; never counted green.
5. Record every call with command, exit code and the tail line (`OK` / `Ran N tests`) in a table
   in `TASK-260922-1ejkxv_results.md`; state that together they are exactly the three lines of
   the `validate:` recipe (quote the recipe) plus regenerate-check.
6. Attach the updated results, check the DoD items citing that resource, and hand off:
   `task-board handoff TASK-260922-1ejkxv --role developer`.

## Boundaries
Unchanged from the original brief. No new product work. Do not run the aggregate
`make validate` in one call again. Do not background anything past your turn.

## Addendum (orchestrator, before this spawn) — base moved
The Story workspace was converged onto spec main `eadb1c0` (TASK-260906-1xbrz6 landed: it touched
CHANGELOG, profiles/manager.md, protocol/environments.md, tools/validate.py, tools/test_validate.py).
Your delta was carried by a clean three-way merge (tree `029bdd27`); a safety ref of your pre-merge
delta is `refs/campaign/1ejkxv-delta-20260923`. The index was UNSTAGED to let the converge run, so
your four new case files are now untracked. Before the gate: read the merged hunks in the five
shared files, `make regenerate`, then `git add` exactly your task's paths (list them in results;
never `git add -A` over `conformance/v1/fixtures/byte-exact/`), then the bounded gate above
including `make regenerate-check` (it compares against the index).
