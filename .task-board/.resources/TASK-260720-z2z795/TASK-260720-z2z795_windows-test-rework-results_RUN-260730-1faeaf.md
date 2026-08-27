# TASK-260720-z2z795 Windows test rework outcome

Run: `RUN-260730-1faeaf`

## Outcome

The two post-commit Windows CI failures are narrowed to invalid test stimuli,
not a transaction-engine defect. The accepted engine implementation remains
unchanged.

Only `tests/test_transactions.py` changed:

- staging recovery now tampers with a later entry whose durable state permits
  only the writable construction identity, then makes it read-only;
- cleanup recovery now advances beyond one file, tampers with that durably
  writable prior file, and makes it read-only;
- both cases cross Windows' actual `S_IWRITE` identity boundary while remaining
  exact-mode mismatches on POSIX;
- each test verifies that recovery rejects the foreign mode, preserves the
  tampered entry and journal, and does not advance live state incorrectly.

`src/csk/locking.py`, `src/csk/transactions.py`, and `tests/test_locking.py`
remain byte-for-byte identical to signed commit
`a5ec015c0b5b51149c81a679ce259920a3523374`.

## Validation

Every required local gate exited 0:

- repaired regressions: 2 passed;
- deterministic Windows one-bit identity probe: 2 passed;
- Windows-sensitive selection: 15 passed on Python 3.11 and 15 passed on
  Python 3.14;
- prior lock-integrity regressions: 14 passed;
- transaction contract selection: 13 passed;
- focused locking/transaction pytest: 129 passed, 4 platform skips;
- strict mypy 2.1 and 2.3: clean across 62 source files;
- Ruff lint and format check: clean;
- full Python 3.14 pytest: 808 passed, 58 platform skips;
- package build: sdist and wheel built successfully;
- `git diff --check`, base ancestry, signed-head verification, product-code
  unchanged check, and nothing-staged check: all exit 0.

The exact command ledger, outputs, real exit codes, intermediate diagnostic
failures, hashes, and platform boundary are in
`TASK-260720-z2z795_windows-test-rework-validation_RUN-260730-1faeaf.log`.

## Handoff boundary

No files were staged, committed, or pushed. The current test-only patch could
not be run by GitHub Actions on real Windows because this role is prohibited
from publishing bytes. The commit-owning orchestrator must publish the reviewed
patch and rerun the Windows Python 3.11–3.14 matrix. No current-byte real
Windows success is claimed here.
