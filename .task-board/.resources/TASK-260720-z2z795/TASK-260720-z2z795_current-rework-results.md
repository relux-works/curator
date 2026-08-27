# TASK-260720-z2z795 current rework outcome

Run: `RUN-260730-b00eb8`  
Date: `2026-07-30`  
Reviewed-input verdict:
`TASK-260720-z2z795_rereview-verdict_RUN-260730-25e35f.md`

## Outcome

The four material findings in the current re-review verdict are addressed
within the task-owned transaction infrastructure:

- Every `TransactionEngine` operation now binds its held lock witness to the
  engine's canonical manager-home identity before journal, sidecar, lock
  namespace, generation, or mutable-target access. The read-only witness is
  checked again at mutation boundaries.
- Generic mutable targets now distinguish byte trees from stable filesystem
  entries. Entry targets journal and digest the exact symlink destination,
  replace or unlink the owned entry without following it, recover and roll back
  the exact link, and keep an explicitly managed referent in an independent
  namespace.
- Recovery treats a cleanup record durably marked `removed` as consumed
  ownership. Any canonical path or tomb that reappears is rejected and
  preserved, including byte-for-byte identical files and directories.
- Home, project, and build locks recanonicalize the manager home after
  first-use directory creation and before process reservation or physical lock
  selection. This closes the provisional Windows identity split while
  retaining deterministic lock ordering.

Existing durable journals, deterministic target ordering, consumer-last
durability, reverse rollback, stale-preimage defense, namespace rejection,
cross-project recovery, and concurrent-consumer behavior remain covered.

## Tests added

Focused regressions cover:

- wrong-home witnesses for prepare, commit, recover, and referenced generation
  reads, including absence of wrong-home residue;
- stable symlink entry digest, commit, removal, rollback, crash recovery,
  stale-preimage preservation, and referent independence;
- file and directory cleanup tombs, with both exact and foreign bytes,
  reappearing after durable `removed`;
- deterministic Windows provisional-to-physical identity changes,
  recanonicalization for all three lock classes, and real Windows
  per-directory-case-sensitive first-use contention.

The real Windows contention cases are collected but skipped on this macOS
host. No current-byte Windows runner result is claimed.

## Current-byte validation

- New boundary regressions: `19 passed, 3 skipped`, exit 0.
- Prescribed prior lock-integrity regressions: `14 passed`, exit 0.
- Contract-targeted regressions: `13 passed`, exit 0.
- Focused locking/transaction pytest: `124 passed, 4 skipped`, exit 0.
- Full pytest: `694 passed, 23 skipped`, exit 0.
- Strict mypy: no issues in 57 source files, exit 0.
- Ruff lint and format checks: exit 0.
- Package build: sdist and wheel built, exit 0.
- `git diff --check`: exit 0.

Exact commands, real exit codes (including diagnostic failures), hashes, and
skip boundaries are in
`TASK-260720-z2z795_current-rework-validation.log`. Tool versions are in
`TASK-260720-z2z795_current-rework-tool-readiness.md`.

## Provenance and scope

- Reused instructed worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`
- Accepted worktree HEAD remained:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`
- Modified only:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`
- No fetch, reset, rebase, stage, commit, SSH, push, tag, release, pin,
  compiler policy, installer policy, or Go UX work was performed.

The `logbook` executable is unavailable on this host (exit 1), so important
findings and evidence are persisted in this outcome and the task board.
