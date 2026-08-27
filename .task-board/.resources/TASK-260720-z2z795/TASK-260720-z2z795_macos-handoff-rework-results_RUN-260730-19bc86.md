# TASK-260720-z2z795 macOS manager-home handoff rework

Run: `RUN-260730-19bc86`

## Outcome

The current-byte macOS manager-home canonicalization stampede is repaired
without changing the accepted transaction engine or Windows permission
behavior.

- Darwin stored spelling now comes from the kernel `F_GETPATH` witness on an
  `O_EVTONLY`/no-follow descriptor. Canonicalization no longer enumerates every
  entry in every path component for each contender.
- The stable-lock deadline starts before manager-home creation,
  recanonicalization, and lock-directory preparation. Preparation consumes the
  same timeout budget as filesystem lock contention.
- New deterministic tests reject parent-directory enumeration, synchronize
  twelve same-home preparation calls at a barrier, prove every consumer
  finishes, and prove slow preparation expires the lock deadline without
  leaking process reservations.
- Existing case aliases, Unicode normalization aliases, symlink aliases,
  first-use recanonicalization, lock order, release handoff, stable witnesses,
  transaction ordering/recovery/rollback, and Windows permission-identity
  behavior remain covered and green.

The current run changed only:

- `src/csk/locking.py`
- `tests/test_locking.py`

The existing unstaged `tests/test_transactions.py` delta is the preserved,
independently accepted Windows permission-stimulus rework from
`RUN-260730-1faeaf`; this run did not alter it. `src/csk/transactions.py`
remains byte-identical to signed HEAD.

## Final validation

All required final gates exited `0`:

- deterministic repaired boundary: `4 passed`;
- repeated fresh-process handoff stress: `20/20` processes passed;
- prior lock-integrity selection: `14 passed`;
- transaction contract selection: `13 passed`;
- focused locking/transaction pytest: `132 passed, 4 skipped`;
- full Python 3.14 pytest: `811 passed, 58 skipped`;
- strict mypy 2.1 and 2.3: no issues in `62 source files`;
- Ruff lint and format: clean;
- preserved forced-Windows permission identity: `2 passed`;
- package build: sdist and wheel built successfully;
- `git diff --check`, cached diff check, accepted-base ancestry, signed-HEAD
  verification, and nothing-staged checks: green.

The exact commands, real exits, diagnostic failures, hashes, and provenance are
in
`TASK-260720-z2z795_macos-handoff-validation_RUN-260730-19bc86.log`.

## Provenance and boundary

- signed task HEAD:
  `a5ec015c0b5b51149c81a679ce259920a3523374`;
- accepted base:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` remains an ancestor;
- local `origin/main` advanced independently to
  `138ab82a9a8cda8de3ca260a535c25c4f7dc054f`; its merge base with task HEAD is
  `495ad021847529ce5a544dba415ca2fe19949539`.

The current run did not rebase, stage, commit, push, use SSH, tag, release, or
modify any pin. Real Windows CI cannot run against these unstaged bytes, so no
current-byte Windows matrix success is claimed. The commit-owning orchestrator
retains publication and remote-matrix responsibility after independent review.
