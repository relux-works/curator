# TASK-260720-z2z795 final re-review verdict

Date: 2026-07-30  
Role: reviewer  
Run: `RUN-260730-7576f5`

## Verdict

**Changes requested → `to-dev`.**

This is ordinary implementation rework. No external blocker, product decision,
approval, or human-only architecture decision is required.

The reviewer run is not goal-bound: the required pre-verdict
`task-board spawn goal RUN-260730-7576f5` query returned `Active Goal: none`.
The only run directive is acknowledged and forbids SSH, product-code changes,
commits, and publication; this review complied.

## Reviewed provenance and scope

- Accepted prerequisite `TASK-260720-1pvfj5` is `done`.
- Original task base
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` remains an ancestor of the
  reviewed task branch.
- The clean canonical main worktree is now at
  `d5d16bfcaa2fe43dc994b819c2659512c4fd8f0a`, matching `origin/main`.
- The task worktree is
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`.
- Its reviewed committed head is
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`; its parent and current
  merge-base with `origin/main` are
  `dd76b570f88339fd1d659c02950e68b17f6ba834`.
- Current rework remains unstaged and uncommitted and changes exactly:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`
- Current SHA-256:
  - `locking.py`:
    `c8ad9d12b64b014ed4069b94248d9ef3d18a1d54b2c112624eefef62eced174d`
  - `transactions.py`:
    `6eed62050966181ffc25350d00a0ad80eeca2ab8c6af814b75d5fd8da6f25239`
  - `test_locking.py`:
    `8bb3deed100cd4906b3ce29701226b093f172af3157f83366fae38bed013bdf1`
  - `test_transactions.py`:
    `88298f3aee8ca705e641555ffad69e83b60cb9de1c61dd7e305d52193739f0f5`

The prior target/journal/sidecar namespace-overlap cases are closed by the
current tests and implementation. The journal-owned sidecar cleanup state and
deterministic Windows `MoveFileExW` routing are also present locally. The
current uncommitted bytes have not run on real Windows; the older green remote
matrix covers only `5f8bfbbd...` and is not current-byte evidence.

## Material finding 1: a failed same-instance acquisition erases a held lock reservation

`_ExclusiveFileLock.__enter__` calls `_record_acquire_failed()` for every
failure, including failures from the order precheck before a new reservation
exists (`src/csk/locking.py:88-133`). The manager-home implementation then
clears `home_held` when it points at the same object
(`src/csk/locking.py:448-455`). Project and build locks have the same
held-set-discard pattern at `src/csk/locking.py:259-265` and
`src/csk/locking.py:380-386`.

A deterministic probe held one `ManagerHomeLock`, attempted to re-enter that
same object, caught the expected `LockOrderError`, and then attempted a project
lock from another thread while the home filesystem lock was still held.

Actual output:

```text
reentrant_error=LockOrderError:the manager-home lock is already held
reentrant_worker_alive=False
reentrant_worker_result=['PROJECT_ACQUIRED_UNDER_HELD_HOME']
```

The failed acquisition weakens the process registry and permits the exact
project-under-home inversion the hierarchy must reject.

Required rework:

1. Associate cleanup with the specific pending acquisition attempt; a
   precheck failure must never remove an already-held reservation.
2. Keep held and pending state separate across every success/failure path for
   project, build, and home locks.
3. Add same-instance re-entry regressions that catch the error and then prove
   cross-thread project/build/home order remains enforced until real release.

## Material finding 2: the filesystem/process release handoff rejects a valid concurrent home consumer

`_ExclusiveFileLock.__exit__` removes the filesystem lock at
`src/csk/locking.py:145-147` and only afterwards clears process state at
`src/csk/locking.py:148-150`. A waiting home consumer can therefore create the
now-absent filesystem lock while the first object is still recorded in
`home_held`. Its `_record_acquired()` then raises at
`src/csk/locking.py:432-445` instead of completing the required serialized
handoff.

A deterministic scheduler probe paused the first consumer after filesystem
unlink and before `_record_released`, then ran the second consumer.

Actual output:

```text
handoff_threads_alive=[False, False]
handoff_results=['SECOND_ERROR:LockOrderError:the manager-home lock is already held']
```

This makes a valid concurrent home consumer fail at the release boundary,
contrary to the concurrent-success requirement. The accepted Go reference
uses counted process-wide pending/held home state so this serialized OS-lock
handoff does not become an order error.

Required rework:

1. Model home ownership so an OS-serialized release/acquire handoff cannot
   spuriously fail (for example, counted/set state with per-attempt ownership,
   or equivalent atomic coordination).
2. Preserve the prohibition on new project/build acquisitions throughout the
   transition.
3. Add a deterministic release-window regression plus a repeated concurrent
   consumer stress case.

## Material finding 3: preparing recovery adopts and deletes unjournaled foreign staging bytes

`prepare()` durably writes only the generic `preparing` journal and then calls
`_copy_target` without journaling staging-entry or byte progress
(`src/csk/transactions.py:136-157`). During recovery,
`_prepare_sidecar_cleanup()` explicitly bypasses the allowed-digest check for
any preparing staged sidecar and records whatever bytes are observed after the
crash (`src/csk/transactions.py:855-894`, especially lines 871-875).

A deterministic probe crashed after writing a partial staged file, replaced
that sidecar with different safe bytes after the crash, and invoked recovery.

Actual output:

```text
before_recovery_sidecar=FOREIGN BYTES AFTER CRASH
before_recovery_journal_exists=True
recovery_error=None
after_recovery_sidecar_exists=False
after_recovery_journal_exists=False
after_recovery_live_exists=False
```

Recovery silently claimed and deleted bytes that no durable journal state
proved the transaction had written. This violates fail-closed recovery and the
accepted behavior reference, whose `internal/transaction/staging.go` journals
entry creation plus write-ahead byte/digest progress and validates the durable
source prefix before cleanup.

Required rework:

1. Journal the desired-source manifest and per-entry staging ownership before
   creating bytes, including write-ahead byte/digest progress for partial
   files (or an equivalently strong durable ownership design).
2. Recovery must validate observed partial staging strictly against that
   durable progress and preserve/reject changed or unrecorded bytes.
3. Add file and directory crash regressions that mutate, replace, and add
   staging bytes after the crash; recovery must fail closed and retain the
   journal/evidence.

## Independent validation ledger

All standard gates are green on the exact current bytes:

- Contract-targeted pytest: `13 passed`.
- Focused pytest:
  `57 passed, 1 skipped in 2.25s`; the skip is the real-Windows
  file/tree-flush test.
- Strict `python -m mypy`: no issues in `57 source files`.
- Changed-file Ruff lint: `All checks passed!`.
- Changed-file Ruff format: `4 files already formatted`.
- Full pytest: `627 passed, 20 skipped in 86.95s`.
- Package build: sdist and wheel built successfully.
- `git diff --check`: exit `0`, no output.

These gates establish that the current suite is stable but lacks coverage for
the three reproduced boundaries above. Exact commands, exits, probe outputs,
hashes, and log digests are attached separately as
`TASK-260720-z2z795_final-rereview-validation.log`.

After rework, rerun focused pytest, strict mypy, Ruff, full pytest, build, the
three deterministic regressions, and a real Windows Python matrix on the exact
committed bytes before the next independent review.

The standalone `logbook` command and a task-board logbook subcommand are not
available in this environment. Findings are persisted in this task-scoped
verdict, validation resource, and task notes.
