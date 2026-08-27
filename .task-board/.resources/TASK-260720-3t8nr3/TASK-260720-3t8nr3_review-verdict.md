# TASK-260720-3t8nr3 review verdict — CHANGES REQUESTED (to-dev)

Reviewer run: `RUN-260730-c38a81`. Read-only review; no product code was
modified.

## Reviewed artifact

- Repository: `/Users/iv/Developer/intranet/cocoaskills`
- Worktree: `.temp/TASK-260720-3t8nr3/worktree`
- Branch: `task/TASK-260720-3t8nr3-transactional-project-hybrid`
- Worktree `HEAD` = `b3a5031ed551b27a298eef486a068b5175beaacc`, identical to the
  recorded base SHA and to the `main` SHA in the attached instruction. Work is
  uncommitted: `LOGBOOK.md`, `src/csk/adapters.py`, `src/csk/cli.py`,
  `src/csk/consumers.py`, `src/csk/installer.py`, `tests/test_install.py`
  modified, `tests/test_installer_transactions.py` added
  (1787 insertions / 226 deletions plus a 714-line new test module).

## Independent validation performed by the reviewer

Every command below was run by the reviewer against the worktree using
`/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python` (pytest resolves
`pythonpath = ["src"]` from the worktree, so the worktree sources are the ones
exercised).

| Gate | Command | Result |
| --- | --- | --- |
| Task vectors | `python -m pytest -q tests/test_installer_transactions.py` | 7 passed in 83.38s |
| Focused project/hybrid/adapter/closure | `python -m pytest -q tests/test_install.py tests/test_hybrid_scope.py tests/test_adapters.py tests/test_closure_install.py` | 80 passed in 732.73s |
| Transaction/locking/cache/planner/activation/gc/cli | `python -m pytest -q tests/test_transactions.py tests/test_locking.py tests/test_build_cache_posix.py tests/test_build_planner.py tests/test_build_activation.py tests/test_gc.py tests/test_cli.py` | 285 passed, 9 skipped in 170.09s |
| Decisive full suite | `python -m pytest -q` | 1127 passed, 98 skipped in 1311.10s |
| Strict typing | `python -m mypy` | Success: no issues found in 67 source files |
| Lint (as claimed) | `uvx ruff check src/csk/installer.py src/csk/adapters.py src/csk/cli.py src/csk/consumers.py tests/test_installer_transactions.py tests/test_install.py` | All checks passed |
| Concurrency vector, 12 consecutive runs | `python -m pytest -q tests/test_transactions.py::test_concurrent_project_transactions_preserve_both_consumers` | 12/12 passed, ~0.20s each (macOS) |

The implementer's reported evidence is reproducible and honest: the full-suite
figure (1127 passed / 98 skipped / 1311.10s) matched exactly. Note the repo has
no `[tool.ruff]` configuration and CI runs only pytest (12 OS/Python
combinations) plus strict mypy, so "lint clean" is an ad-hoc, file-scoped
signal, not a project gate.

## What is genuinely good

- The lock order is correct and matches what `src/csk/locking.py` enforces:
  project lock spans planning and private compilation, per-key `BuildLock`
  serializes cache misses, and the manager-home lock is held only for recovery,
  revalidation, cache publication and the durable commit.
- Commit-time revalidation is real, not decorative: generations
  (`_assert_generation_current`), target preimages
  (`_assert_target_preimages_current`), closure/ref/provider/collision state
  (`_revalidate_closure`) and cache winners (`_publish_planned_builds`) are all
  rechecked under the home lock before `engine.prepare`.
- Consumer-last ordering is structurally guaranteed rather than incidental: the
  engine sorts targets by `(target_class, identifier)` UTF-8, and `90-consumer`
  is the highest class.
- Marker v2, `build_roots` exclusion, and immutable-artifact shims are covered
  by executable vectors, including actually running the produced launchers.
- Adapter mirrors are modelled as per-entry transaction targets rather than
  aggregate byte trees. That is the right call — the transaction protocol
  deliberately rejects symlink descendants — and the rationale is recorded in
  `LOGBOOK.md`.
- The directory-precreation unwind in `_commit_materialization`
  (`_make_missing_directories` / `_remove_created_directories`) is genuinely
  exercised: `test_second_project_consumer_failure_rolls_back_without_touching_first`
  watches `project_two/.claude`, which only exists because of that
  precreation, and asserts byte-identical restoration.

## Blocking findings

### B1 — The attached task instruction was not delivered and not acknowledged

`main-windows-concurrency-flake.md` is attached to this task as a precondition
resource and states, verbatim:

> While implementing the project transaction integration, preserve the vector
> and assess whether the test needs deterministic coordination or a
> platform-appropriate bounded wait. Do not merely skip/xfail it or remove its
> liveness assertion.

The orchestrator note on the board repeats it: "Preserve the attached flaky
liveness evidence and harden deterministic coverage in this transaction scope
rather than ignoring it."

Current state:

- `tests/test_transactions.py` is byte-for-byte unchanged. The vector still
  relies purely on wall-clock budgets — `threading.Barrier(2)` with
  `timeout=3`, `ProjectLock(..., timeout=3)`, `ManagerHomeLock(..., timeout=3)`,
  and a fixed `thread.join(timeout=5)` — and still asserts liveness at
  `tests/test_transactions.py:2170`.
- Neither `TASK-260720-3t8nr3_results.md` nor the `LOGBOOK.md` entry mentions
  the instruction, the Windows 3.14 failure, or any assessment of the vector.
  A full-text search of both artifacts for `concurren|windows|flake|liveness`
  returns nothing.
- The AC also names `concurrency` among the gates that must pass.

The vector was not skipped or weakened — that part of the instruction is
respected — but the required *assessment* was neither performed nor recorded,
and no deterministic coordination was added. On the failing Windows 3.14 job
both worker `errors` lists were empty while one thread was still alive, which is
exactly the shape a time-budget-only harness produces; 12 consecutive local
macOS runs at ~0.20s each confirm the macOS margin is enormous and therefore
tells us nothing about the Windows behaviour.

Required: either add deterministic coordination (e.g. explicit handoff events
instead of a bare barrier plus fixed join budget) or a defensible
platform-appropriate bounded wait, and record the assessment and its reasoning
in the task artifact and `LOGBOOK.md`. If the conclusion is that the vector is
already correct as written, that conclusion still has to be argued in writing
with the Windows evidence addressed.

### B2 — `csk install` / `csk upgrade` silently lost snapshot GC and part of orphan sweeping

`installer.install()` dropped its terminal `gc.collect_runtime(config, config.path.parent)`
call, and the `gc` import was removed from `src/csk/installer.py`. That call has
existed on the project-install path since the MVP commit (`4f5af6b`).
`gc.collect_runtime` is now reachable only from `csk gc` (`src/csk/cli.py:574`)
and `csk global install/upgrade` (`src/csk/cli.py:888`).

What was correctly reimplemented as journal targets:

- runtime GC → `80-removal` targets from `_runtime_references_for_plan`
- consumer pruning → `_desired_consumers` (drops registry entries whose checkout
  is gone or holds no markers)
- dead install temporaries under the project and hybrid skills roots →
  `_stale_entry_targets` / `_is_dead_install_orphan`

What was **not** reimplemented and is now simply gone from the main install
path:

- snapshot cache pruning (`gc._collect_snapshots`, `~/.csk/cache/<source>/<commit>/snapshot`).
  Every `csk install` / `csk upgrade` populates that cache via
  `closure.build_closure(use_cache=True)` and nothing prunes it any more, so a
  routine upgrade loop accumulates one snapshot tree per source per commit
  until the user happens to run `csk gc` or a global install.
- `gc.sweep_orphans` over `~/.csk/global/skills` and `~/.csk/runtime/<skill>`.

`tests/test_gc.py` calls `gc.collect_runtime` directly, so the suite stays green
and cannot see this. The removal is not mentioned in
`TASK-260720-3t8nr3_results.md` or in the `LOGBOOK.md` entry — the logbook only
covers the project/hybrid orphan case ("Dead legacy install temporaries are now
explicit stale-removal targets rather than an unjournaled post-install GC
mutation"), which suggests the snapshot half was dropped by accident rather than
by decision.

Nothing in the AC forbids running snapshot GC after a successful commit — the
global install path still does exactly that — so this is recoverable rework, not
a design conflict.

Required: either restore snapshot GC and the remaining orphan sweeps on the
project-install path (a post-commit `gc.collect_runtime`, as
`_cmd_global_install` already does, is the obvious shape), or, if the omission
is deliberate, state the decision and its disk-growth consequence explicitly in
`LOGBOOK.md` and in the task artifact, with the follow-up work tracked.

## Non-blocking findings (fix or record, reviewer will not re-block on these alone)

### N1 — The entire new vector module is skipped off POSIX

`tests/test_installer_transactions.py:27` applies a module-level
`pytestmark = pytest.mark.skipif(os.name != "posix", ...)`. All seven AC vectors
— including two-project shared-cache success and consumer-last rollback
isolation — therefore never execute on the Windows CI leg, which is precisely
the platform carrying the known transaction timing flake. The repo already
maintains a Windows counterpart for the cache layer
(`tests/test_build_cache_windows.py`), so a full module-level skip is a broader
exclusion than the stated reason ("Exercises the POSIX protected build cache and
launchers") requires. At minimum the platform-independent vectors (generation
restart, consumer-last rollback ordering) should be reachable on Windows.

### N2 — Staging copies the whole live world on every project install

`_stage_materialization` calls `_copy_live_directory` on the project's entire
`.agents` tree, on `~/.csk/hybrid`, and on `~/.csk/runtime` — unconditionally,
on every real install, including a no-op re-install where every digest matches
and zero transaction targets are produced. Install cost becomes
O(total installed runtime on the machine) rather than O(changed), and it lands
on the temp filesystem. The isolation goal is right; the whole-tree copy is a
heavier way to reach it than the target set requires. Worth a design note even
if it is not changed now.

### N3 — Unprotected `recover()` can escape as a traceback

`src/csk/installer.py:161` calls `_transaction_engine(csk_home).recover(home_lock)`
outside any `try`, before the retry loop. `TransactionEngine.recover` raises
`TransactionError` / `TransactionCorruptionError`
(`src/csk/transactions.py:400-403`), and neither type is in `cli.main`'s handled
exception tuple (`src/csk/cli.py:71-87`). A corrupt journal — exactly the
condition recovery exists for — therefore surfaces as an uncaught traceback
instead of a clean `error:` line and exit code. The second `recover()` call
inside `_install_project_once` is protected by that function's
`except Exception` boundary, so the two paths behave inconsistently.

### N4 — `auto` adapter mode probes the live project tree outside the transaction

`adapters._link_probe` creates and immediately unlinks
`.csk-symlink-probe-<pid>` inside `_nearest_existing_directory(live_path.parent)`.
For an adapter root such as `<project>/.claude/skills` that resolves to
`<project>/.claude` or, when the agent directory does not exist yet, to the
project root itself. The write is transient and cleaned up on both success and
`OSError`, but it is an unjournaled mutation of the user's tree during staging,
an abrupt kill leaves an artifact that no sweep recognises (the name does not
match `_INSTALL_ORPHAN_RE`), and the pid-only name is not unique across
concurrent same-process installs. The pre-change behaviour probed by attempting
the real symlink at the real target, so this is a new artifact class rather than
a preserved one.

## Verdict

**changes requested → `to-dev`.**

The transactional design is sound, the AC vectors that were written are real and
green, strict mypy is clean, and the full suite reproduces exactly. Two items
must be closed before acceptance:

1. **B1** — deliver and document the concurrency-vector assessment required by
   the attached instruction.
2. **B2** — restore install-time snapshot GC and the remaining orphan sweeps, or
   record the deliberate removal and its consequence.

N1–N4 should be addressed or explicitly recorded in the same cycle, but on their
own would not have blocked acceptance.
