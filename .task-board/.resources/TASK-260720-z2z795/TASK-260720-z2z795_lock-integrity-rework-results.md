# TASK-260720-z2z795 lock-integrity rework outcome

## Scope and provenance

Reused the accepted task worktree without reset or byte loss:

`/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

- Original accepted base:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`
- Current committed head:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`
- Reviewed current-byte input:
  `TASK-260720-z2z795_lock-integrity-review-verdict.md`
- Current changes remain unstaged and limited to:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`

## Finding 1: lock namespace self-mutation

Closed.

Transaction namespace validation now treats the canonical manager-home lock
file plus the complete project and build lock directories as internal reserved
state during both plan construction and journal loading. The existing
ancestor/descendant, physical `samefile`, platform case, and macOS Unicode
normalization checks now apply to every live target, desired/backup/rollback
sidecar, and cleanup tomb against those lock namespaces.

Planning rejects the home lock, held project and build lock files, lock
directories, their parents and descendants, and physical aliases before a
journal or target mutation. Regressions assert that every active witness and
its bytes remain unchanged.

Transaction operations also carry the caller's home-lock witness through the
serialized mutation scope. Journal publication, staging, target swap,
rollback, cleanup, removal, and fault-hook boundaries assert the witness, so a
lost lock can no longer be returned as a successful commit or recovery.

## Finding 2: stale restore can overwrite a new owner

Closed.

Lock files are now stable manager state. Acquisition opens one canonical file
and takes a nonblocking operating-system exclusive lock:

- POSIX: `flock(LOCK_EX | LOCK_NB)`
- Windows: `LockFileEx(LOCKFILE_EXCLUSIVE_LOCK |
  LOCKFILE_FAIL_IMMEDIATELY)`

Release drops the OS lock and descriptor but never unlinks or renames the lock
path. Process crash releases ownership through the kernel. There is no stale
rename or restoration path, so a stale claimant cannot move or overwrite a
new owner's file. Stable v1 records are diagnostic; OS ownership is
authoritative. Legacy records remain fail-closed while their PID is live or
their bytes are corrupt, and a provably dead legacy record is adopted under
the OS lock.

The cross-process barrier regression kills an owner, proves ordinary
contention, starts two simultaneous consumers, verifies non-overlapping
critical sections and both witnesses, and reacquires the same persistent lock
file afterward. It is platform-neutral and will run unchanged on Windows CI.

## Validation

Current-byte final gates:

- New lock-integrity regression set: 14 passed, exit 0.
- Prior reviewer regression set: 14 passed, exit 0.
- Original contract-targeted set: 13 passed, exit 0.
- Focused pytest: 82 passed, 1 Windows-only skipped, exit 0.
- Strict `python -m mypy`: 57 source files clean, exit 0.
- Ruff lint: clean, exit 0.
- Ruff format check: 4 files already formatted, exit 0.
- Full pytest: 652 passed, 20 skipped, exit 0.
- Package build: sdist and wheel built, exit 0.
- `git diff --check`: clean, exit 0.

The expected-red lock-integrity command failed with exit 1 and 14 failures
before implementation, as recorded in the validation ledger.

Exact commands, real exit codes, hashes, diagnostic failures, and the platform
boundary are in
`TASK-260720-z2z795_lock-integrity-validation.log`. Tool readiness is in
`TASK-260720-z2z795_lock-integrity-tool-readiness.md`.

No current-byte real-Windows execution was available in this run; earlier
remote Windows evidence applies only to committed head `5f8bfbbd` and is not
claimed for these unstaged bytes. No SSH, host management, staging, commit,
push, tag, release, pin update, Go UX, or installer-policy work was performed.
