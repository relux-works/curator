# TASK-260720-z2z795 Windows CI rework outcome

Run: `RUN-260730-fbe9eb`  
Date: 2026-07-30  
Input commit: signed `d0d91788bc01c6590dec49a0dd296a78e95fd44b`  
Input CI evidence: `TASK-260720-z2z795_windows-ci-d0d9178-py314.log`

## Scope

This rework closes the 11 failures from the Windows/Python 3.14 job at signed commit `d0d9178` while preserving the stable no-rename lock protocol and its `O_BINARY` repair.

Only these current-byte paths changed:

- `src/csk/transactions.py`
- `tests/test_transactions.py`

No files were staged, committed, pushed, tagged, or released.

## Changes

1. Invalid manager-home, project, and build lock targets are now constructed in tests without reopening or digesting the held `LockFileEx` path. The transaction engine therefore reaches its existing canonical namespace rejection, and witness verification continues through the owning descriptor.
2. Intermediate staging and cleanup mode validation now uses the only permission identity Windows `chmod` can set: the read-only/writable state represented by `stat.S_IWRITE`. POSIX mode validation remains exact.
3. The Windows identity is applied consistently to ordinary staging validation, active crash-recovery staging validation, cleanup-tree validation, and cleanup-entry validation.
4. Regressions exercise files and directories in both directions: ignored writable POSIX-bit differences are accepted, while a read-only/writable transition is rejected as foreign mutation.
5. Cross-platform read-only assertions now test Windows read-only identity rather than unsupported exact POSIX bits. The foreign-staging mutation test toggles the read-only state on Windows so the fail-closed branch remains meaningful.
6. The canonical mypy 2.1 environment exposed an existing literal-narrowing error in `_require_phase`; the validated string is now explicitly cast to `Phase` without changing runtime behavior.

The platform rule follows Python's documented Windows contract: `os.chmod()` can set only the file read-only flag and ignores other permission bits:
https://docs.python.org/3/library/os.html#os.chmod

## Validation summary

- Expected-red Windows permission regression: exit 1, `4 failed`.
- Repaired Windows CI cluster selection:
  - Python 3.11: exit 0, `15 passed`.
  - Python 3.14: exit 0, `15 passed`.
- Prescribed prior lock-integrity regressions: exit 0, `14 passed`.
- Contract regressions: exit 0, `13 passed`.
- Focused locking/transaction pytest: exit 0, `129 passed, 4 skipped`.
- Strict mypy:
  - canonical Python 3.14 / mypy 2.1: exit 0, 62 source files;
  - mypy 2.3 compatibility route: exit 0, 62 source files.
- Ruff lint and format: exit 0.
- Full Python 3.14 pytest: exit 0, `808 passed, 58 skipped`.
- Package build: exit 0; wheel and sdist produced.
- `git diff --check`: exit 0.
- Signed input commit verification: exit 0.

Exact commands, real exits, diagnostic failures, hashes, and the remote-Windows boundary are in `TASK-260720-z2z795_windows-ci-rework-validation_RUN-260730-fbe9eb.log`.

## Handoff boundary

The current two-file patch is unstaged and uncommitted at signed HEAD `d0d9178`. This developer run did not publish bytes. The commit-owning orchestrator must publish reviewed bytes and rerun the Windows Python 3.11, 3.12, 3.13, and 3.14 jobs before any current-byte Windows success claim.

