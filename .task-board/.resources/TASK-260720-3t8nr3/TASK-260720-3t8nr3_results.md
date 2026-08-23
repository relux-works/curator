# TASK-260720-3t8nr3 implementation and validation evidence

Revision 3. This artifact covers the original integration, the rework demanded
by review verdict `RUN-260730-c38a81` (B1, B2, N1–N4), and the subsequent
rework demanded by `RUN-260730-bbd2a8` (B3 and N5–N7).

## Workspace

- Repository: `/Users/iv/Developer/Wildberries/cocoaskills`
- Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-3t8nr3/worktree`
- Branch: `task/TASK-260720-3t8nr3-transactional-project-hybrid`
- Recorded base SHA: `b3a5031ed551b27a298eef486a068b5175beaacc`
- `main` in the canonical clean clone equals `origin/main` equals the recorded
  base SHA (verified again this cycle with `git fetch origin main` +
  `git rev-parse main origin/main`). The task worktree was created only after
  that fast-forward and after both dependency handoffs were present.
- Work is intentionally uncommitted for reviewer inspection:
  9 files changed, 2097 insertions(+), 256 deletions(-), plus the new
  `tests/test_installer_transactions.py`.

## Delivered behavior

- Real project and applicable hybrid installs plan active build providers, run
  `go-v1` misses provider-first and command-lexically in one operation-private
  toolchain staging area outside the manager-home lock, and publish only
  verified cache winners under that lock.
- A project lock spans the operation. Under the manager-home lock the installer
  recovers durable transactions; revalidates generations, target preimages,
  closure refs, provider snapshots, collisions, and cache winners; then prepares
  and commits one durable target plan
  (`src/csk/installer.py:460`–`src/csk/installer.py:497`).
- The target plan contains granular project/hybrid context and marker entries,
  script runtime entries, compiled and script shims, environment files, adapter
  mirrors and ledgers, stale removals, and the consumer registry. Transaction
  class ordering makes `90-consumer` last; rollback restores committed targets
  in reverse order while the home lock remains held.
- New or changed installs use typed marker v2 with build source and receipt
  identities. Build roots are excluded from context/runtime materialization, and
  compiled shims resolve immutable protected-cache artifacts.
- Adapter symlink/copy modes use per-entry transaction targets plus a later
  ledger target. Aggregate byte-tree targets were rejected because the
  transaction protocol intentionally rejects symlink descendants.
- CLI project install/upgrade no longer takes the legacy outer home lock;
  installer-owned project/build/home lock ordering keeps compilation outside the
  home lock. `update` and global operations keep their existing lock ownership.
- Seven integration vectors cover audit/build/commit ordering, marker v2 and
  build-root isolation, build-failure preservation, publication-failure
  preservation, stale-generation restart, hybrid compiled activation, two-project
  shared-cache success, and second-project consumer-last rollback isolation.

## Review rework — how each finding was closed

### B1 — concurrency-vector assessment (attached instruction) — CLOSED

The attached instruction (`main-windows-concurrency-flake.md`) required that
`test_concurrent_project_transactions_preserve_both_consumers` be preserved,
assessed, and either given deterministic coordination or a platform-appropriate
bounded wait — and explicitly not skipped, xfailed, or stripped of its liveness
assertion.

Assessment. The Windows 3.14 failure on main SHA `b3a5031`
(run `30556125542`, job `90916913692`) had both worker `errors` lists empty with
one thread still alive at the fixed `thread.join(timeout=5)` budget. That is the
signature of a wall-clock-budget harness, not of a transaction-protocol defect:
the original vector used `threading.Barrier(2, timeout=3)`, `ProjectLock(...,
timeout=3)`, `ManagerHomeLock(..., timeout=3)` and a five-second join, so the
whole test asserted "this machine is fast enough" as much as it asserted
serialization. Twelve consecutive local macOS runs at ~0.2s each say nothing
about the Windows margin, so a rerun-passes result is not evidence of
robustness.

Change (`tests/test_transactions.py`). The barrier and the fixed join were
replaced with explicit handoff events:

1. both workers take their project locks and signal `projects_ready`;
2. the parent releases `start_first_home` only after both project locks are held;
3. project A takes the manager-home lock and signals `first_home_acquired`;
4. project B waits for that, signals `second_home_attempt`, then contends for the
   manager-home lock;
5. project A commits only after `second_home_attempt`, then releases.

This makes the intended serialization — B blocks on the home lock while A
commits — the thing under test, instead of a timing coincidence. Every wait is
bounded: 10s on POSIX, 30s on Windows (the production lock default), with a
`(2 x coordination) + 5s` completion deadline derived from `time.monotonic()`.
Worker exceptions are collected under a lock, per-worker `completed` events were
added, and the terminal `assert all(not thread.is_alive() ...)` liveness
assertion is retained verbatim. The vector is neither skipped nor xfailed.

Local evidence: 20/20 consecutive green runs, ~0.19s each (see
`concurrency-stress-01.log`). Honest limitation: this machine is macOS/arm64, so
the Windows 30s budget and the `os.name == "nt"` branch are exercised only by CI.
The assessment is also recorded in `LOGBOOK.md`.

### B2 — install-time snapshot GC and orphan sweeps — CLOSED (restored)

`gc.collect_runtime` is restored on the project install/upgrade path
(`src/csk/installer.py:131`–`src/csk/installer.py:137`), together with the `gc`
import. Snapshot cache pruning (`gc._collect_snapshots`) and
`gc.sweep_orphans` over `~/.csk/global/skills`, `~/.csk/hybrid/skills` and
`~/.csk/runtime/<skill>` therefore run again on every successful `csk install` /
`csk upgrade`.

Two deliberate, documented differences from the pre-change behavior:

- GC runs only when the batch is a real install with at least one `ok` result and
  no failed result. The old code ran it unconditionally after any non-dry-run
  batch, which would have violated this task's AC that a build or target failure
  leaves live cache, runtime, markers and consumers unchanged.
- GC holds `ManagerHomeLock` while pruning, so consumer-ledger maintenance is
  serialized against concurrent installs. `gc.collect_runtime` and
  `consumers.replace_consumers` take no locks themselves, so the nesting is safe
  and matches the existing `csk gc` shape (`src/csk/cli.py:573`).

Regression vector: `tests/test_gc.py::test_successful_install_runs_snapshot_and_remaining_orphan_gc`
seeds an unreferenced `~/.csk/cache/<source>/<commit>/snapshot`, a dead-pid
orphan under `~/.csk/global/skills`, and a dead-pid orphan under
`~/.csk/runtime/skill-a`, runs a real install, and asserts all three are gone.

### N1 — module-wide non-POSIX skip — CLOSED

`tests/test_installer_transactions.py` no longer carries a module-level
`pytestmark`. `POSIX_BUILD_VECTOR` is now applied per test, only to the five
vectors that execute POSIX protected-cache artifacts and launchers. The two
platform-independent vectors — `test_commit_generation_change_restarts_complete_project_plan`
and `test_second_project_consumer_failure_rolls_back_without_touching_first` —
run on Windows. The rollback vector was switched to a context-only skill so it
does not smuggle a POSIX shell dependency into that claim.

### N2 — whole-live-world staging copy — RECORDED, not changed

`_stage_materialization` still copies the complete live project `.agents`, shared
hybrid, and runtime trees, so install cost is O(total installed state) rather
than O(changed). This is recorded in `LOGBOOK.md` as a deliberate performance
tradeoff: complete isolated preimages keep transaction target derivation
deterministic, and replacing it with target-scoped copy-on-write staging is a
second state model that this task will not introduce late in its cycle. It
should be revisited if install scale makes the cost material.

### N3 — unprotected `recover()` — CLOSED

The pre-loop `_transaction_engine(csk_home).recover(home_lock)` is now inside
`try/except transactions.TransactionError`
(`src/csk/installer.py:167`–`src/csk/installer.py:178`).
`TransactionCorruptionError` subclasses `TransactionError`
(`src/csk/transactions.py:61`), so a corrupt journal now becomes the same
per-project failed `ProjectResult` that a recovery error inside planning
produces, instead of an uncaught CLI traceback.

### N4 — auto-mode adapter probe writing into the live tree — CLOSED

`adapters._transaction_links_supported` no longer probes the user's project. It
compares `st_dev` of the operation-private staging root and the nearest existing
live parent, and only when they are on the same filesystem does it run
`_link_probe` **against the staging root**; otherwise auto mode selects copy.
Explicit symlink mode is unchanged. Vector:
`tests/test_adapters.py::test_transaction_auto_mode_probes_only_private_staging`
asserts the probe target is staging and that the live project tree stays empty.

## Validation evidence (this cycle)

Every command ran directly as a standalone process, no `tee`, no pipe chain.
Interpreter: `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python`
(pytest resolves `pythonpath = ["src"]` from the worktree). Raw output is
attached as `TASK-260720-3t8nr3_gate-evidence.log`.

| Gate | Command | Result | Exit |
| --- | --- | --- | --- |
| Decisive full suite | `python -m pytest -q` | 1131 passed, 98 skipped in 1328.73s | 0 |
| Task vectors | `python -m pytest -q tests/test_installer_transactions.py` | 7 passed in 74.04s | 0 |
| Transactions / locking / gc / adapters / cache / planner / activation / cli | `python -m pytest -q tests/test_transactions.py tests/test_locking.py tests/test_gc.py tests/test_adapters.py tests/test_build_cache_posix.py tests/test_build_planner.py tests/test_build_activation.py tests/test_cli.py` | 294 passed, 9 skipped in 163.57s | 0 |
| Focused project / hybrid / closure | `python -m pytest -q tests/test_install.py tests/test_hybrid_scope.py tests/test_closure_install.py` | 75 passed in 716.50s | 0 |
| Concurrency vector, 20 consecutive runs | `python -m pytest -q tests/test_transactions.py::test_concurrent_project_transactions_preserve_both_consumers` | 20/20 passed, ~0.19s each | 0 (all 20) |
| Strict typing | `python -m mypy` | Success: no issues found in 67 source files | 0 |
| Lint | `uvx ruff check src/csk/installer.py src/csk/adapters.py src/csk/cli.py src/csk/consumers.py tests/test_installer_transactions.py tests/test_install.py tests/test_transactions.py tests/test_gc.py tests/test_adapters.py` | All checks passed | 0 |
| Syntax | `python -m py_compile` over all changed modules and tests | clean | 0 |
| Patch hygiene | `git diff --check` | clean | 0 |

Dry-run coverage is part of the focused project gate
(`tests/test_install.py::test_dry_run_does_not_modify_project_or_cache`,
`::test_schema_v6_build_root_stays_out_of_dry_run_real_and_up_to_date_context`,
`::test_project_dry_run_missing_go_fails_whole_plan_without_mutation`), all green.

### What was not run, and why

- No Windows execution. This host is macOS/arm64; there is no Windows runner
  available to this session. The Windows-specific parts of the rework — the 30s
  `os.name == "nt"` coordination budget and the two now-unskipped
  platform-independent vectors on the Windows leg — are validated only by CI.
  This is the one claim in the B1 fix that local evidence cannot close.
- `uvx ruff format --check` was not treated as a gate. The repository has no
  `[tool.ruff]` configuration and CI runs only pytest plus strict mypy, so
  file-scoped ruff remains an ad-hoc signal rather than a project gate; the
  pre-existing files are not ruff-format canonical and were not mass-reformatted.

## Notes for review

- No product/API forced fit was required. The adapter representation follows the
  accepted transaction `entry` target model rather than weakening byte-tree
  validation.
- The GC narrowing in B2 is a behavior change relative to pre-task `main`
  (unconditional post-batch GC → success-only, home-lock-held GC). It is
  intentional and required by this task's AC; it is called out here and in
  `LOGBOOK.md` so it is reviewed as a decision, not discovered as a diff.
- N2 is knowingly left as-is and recorded as a tradeoff, not silently dropped.

## Review rework revision 3 — B3 and N5–N7

### B3 — destination-correct private staging — CLOSED

The reviewer proved that materialization's unqualified
`tempfile.TemporaryDirectory` made default adapter auto mode depend on
`$TMPDIR`: a process temp directory on another device caused symlink mirrors to
become copies. `_commit_materialization` now anchors its hidden private stage
beside the physical project, outside the checkout but on the adapter destination
filesystem. If that parent cannot host the stage, it falls back to manager home
and retains the conservative same-device copy decision. The system temporary
directory is no longer part of materialization target derivation.

Two vectors close the gap:

- `test_materialization_staging_is_private_and_anchored_to_project_filesystem`
  sets the process temp root to an unrelated location and proves the real stage
  is a cleaned-up sibling of the physical project, outside both checkout and
  process temp.
- `test_transaction_auto_mode_rejects_cross_device_staging_without_probe`
  supplies deterministic distinct device identities and proves the
  cross-device branch never attempts the symlink probe.

### N5 — post-commit GC contention — CLOSED

A post-commit `ManagerHomeLock` `LockError` now reports a visible
`post-install garbage collection skipped` message without turning already
committed project results into failure. `LockOrderError` is re-raised so an
internal locking defect remains loud. The regression proves the live marker,
successful result, and maintenance message after forced post-commit contention.

### N6 — staging placement — CLOSED as part of B3

The known O(total installed state) copy remains, but it now lands beside the
physical project (manager-home fallback) rather than in a potentially
RAM-backed system temp directory. The residual copy-on-write optimization is
recorded in `LOGBOOK.md`.

### N7 — consumer path canonicalization — RECORDED

Consumer-ledger encoding intentionally resolves paths before sorting and
serializing them. That stabilizes transaction digests across lexical symlink
routes, while allowing the first successful install to rewrite a legacy
unresolved entry. The behavior is recorded in `LOGBOOK.md`.

## Validation evidence — revision 3

Every command ran directly as a standalone process; no gate used `tee` or a
pipe chain. Raw revision-3 evidence is attached as
`TASK-260720-3t8nr3_rework-evidence-rev3.md`.

| Gate | Result | Exit |
| --- | --- | --- |
| Expected-red B3/N5 pair before implementation | 2 intended failures | 1 (honest expected red) |
| Exact B3/N5 post-fix regressions | 3 passed in 25.27s | 0 |
| Task transaction vectors | 7 passed in 73.88s | 0 |
| Project / hybrid / closure / adapter / GC | 94 passed in 845.71s | 0 |
| Transactions / locking / cache / planner / activation / CLI | 295 passed, 9 skipped in 162.21s | 0 |
| decisive full suite | 1134 passed, 98 skipped in 1362.58s | 0 |
| `python -m mypy` | clean on 67 source files | 0 |
| scoped Ruff | all checks passed | 0 |
| changed-file `py_compile` | clean | 0 |
| package build | sdist + wheel built | 0 |
| `git diff --check` | clean | 0 |

The orchestrator initially directed this bounded rework to rely on the focused
gates rather than repeat the 22-minute suite. The developer-added checklist
nevertheless named the full suite explicitly; the item was unchecked when that
evidence mismatch was detected, the exact suite ran green, and only then was
the item checked.

No Windows host or real second filesystem was available in revision 3. The
Windows concurrency budget remains CI-only. The reviewer already reproduced B3
on a macOS RAM disk; this revision deterministically simulates distinct device
identities and separately proves process-temp independence.
