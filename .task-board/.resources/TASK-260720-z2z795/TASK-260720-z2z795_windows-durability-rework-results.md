# TASK-260720-z2z795 Windows durability rework

Run: `RUN-260729-a702af`  
Role: developer  
Accepted base: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`  
Worktree: `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

## Review finding addressed

The Windows journal lifecycle no longer relies on `os.link`, non-write-through
`os.replace`, or a no-op directory sync:

- new journals use atomic `MoveFileExW` without replacement and with
  `MOVEFILE_WRITE_THROUGH`;
- journal state replacement uses
  `MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH`;
- journal cleanup first moves the canonical journal to a durable `.delete`
  tomb, so recovery can safely resume cleanup after either committed success or
  reverse rollback;
- regular-file and directory durability use explicit Windows handles and
  `FlushFileBuffers`; the filesystem-dependent directory errors
  `ERROR_ACCESS_DENIED` and `ERROR_INVALID_HANDLE` are tolerated only because
  every journal namespace move is independently write-through;
- Win32 access, sharing, open, reparse-point, backup-semantics, move, and error
  values are named constants rather than opaque literals.

The implementation follows the accepted Go reference in
`internal/transaction/durability_windows.go`, `replace_windows.go`, and
`journal.go`.

## Focused coverage

- exact Windows flags for journal create, replace, and cleanup-tomb moves;
- exact regular-file and directory handle access/attribute routing;
- a Windows-only real file/tree flush test for the `windows-latest` run;
- restart after a crash immediately after journal tomb publication for both
  committed cleanup and rollback cleanup;
- all existing deterministic ordering, crash recovery, reverse rollback,
  concurrent-consumer, stale-preimage, atomic no-replace, absent-parent, and
  lock-order regressions remain green locally.

## Local validation

- Focused pytest: exit `0`; `32 passed, 1 skipped`.
- Strict mypy: exit `0`; no issues in `56` source files.
- Ruff check: exit `0`.
- Ruff format check: exit `0`.
- Full pytest: exit `0`; `516 passed, 19 skipped`.
- Package build: exit `0`; sdist and wheel built.
- Tracked diff whitespace check: exit `0`.
- The two untracked-file `git diff --no-index --check` diagnostics each exited
  `1` with zero output because Git reports the intentional file difference;
  they are not claimed as green gates.

The one focused skip is the intentionally Windows-only real handle-flush test.
Per the orchestrator boundary recorded on the task, this worker did not run SSH,
VM discovery, host management, publication, or GitHub Actions. The orchestrator
owns the required `windows-latest` focused-suite execution after handoff.

Exact commands, outputs, and exit codes are in
`TASK-260720-z2z795_windows-durability-validation.log`.

## Worktree state

The source remains unstaged and uncommitted, limited to the four task-owned
paths:

- `src/csk/locking.py`
- `src/csk/transactions.py`
- `tests/test_locking.py`
- `tests/test_transactions.py`

No branch publication, pin change, Go driver/toolchain UX, installer policy, or
compiler/schema integration was performed.

