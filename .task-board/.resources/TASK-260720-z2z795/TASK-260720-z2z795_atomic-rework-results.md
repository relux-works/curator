# TASK-260720-z2z795 atomic transaction rework

Date: 2026-07-30  
Role: developer  
Run: `RUN-260729-4ad5fe`

## Provenance

- Accepted dependency: `TASK-260720-1pvfj5`, board status `done`, verdict
  `TASK-260720-1pvfj5_candidate-input-final-review-verdict.md`.
- Accepted base and current task worktree `HEAD`:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Canonical clean worktree and its recorded `origin/main` ref were verified at
  the same SHA. This rework reused the already accepted task worktree and did
  not fetch again, reset, recreate, stage, commit, or publish.
- Worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`
- Branch: `task/TASK-260720-z2z795-transaction-engine`
- Normative contract: `curator-spec/profiles/manager.md` sections 2.5–2.6.
- Accepted behavior reference:
  `curator/.temp/TASK-260720-1pvfj5/rework/composite/internal/transaction`.

## Review findings closed

### Atomic no-replace mutation boundary

`src/csk/transactions.py` now routes every transaction rename through a native
atomic no-replace primitive:

- macOS: `renamex_np(..., RENAME_EXCL)`;
- Linux: `renameat2(..., RENAME_NOREPLACE)`;
- Windows: `MoveFileExW(..., MOVEFILE_WRITE_THROUGH)` without
  `MOVEFILE_REPLACE_EXISTING`;
- any unsupported platform or unavailable Linux kernel primitive fails closed
  before an unsafe fallback can run.

The old destination precheck is gone. Destination-exists errors are normalized
to `TransactionCorruptionError`, the competing destination and source are
preserved, and the renamed directory entries are synced after success. The
same helper covers live-to-backup, staged-to-live, desired-to-rollback, and
backup-to-live transitions for regular files and directories.

Rollback handling for a pending target no longer adopts an unexpected backup
sidecar. It requires the live path to be absent and verifies the journaled
preimage or generation at the backup before advancing state. A collision
therefore leaves the journal recoverable instead of turning foreign bytes into
the recorded backup.

### Exact namespace rollback

Preparation no longer calls `mkdir(parents=True)` for a live target's sibling
staging path. It requires the resolved live parent to already be a real,
non-symlink directory and fails closed otherwise. When a later target has an
absent parent, already prepared sidecars and the journal are removed while the
absent namespace remains absent.

This is the existing-parent option explicitly allowed by the review verdict
and matches the accepted Go reference, whose same-directory staging creation
also fails when the live parent does not exist.

## Added regression coverage

`tests/test_transactions.py` now proves:

- an atomic-boundary competitor is preserved for both a regular file and a
  directory using the real native primitive on the test platform;
- a staged-to-live commit collision preserves the competing success, backup,
  staging, and journal, then restores the preimage after the unknown
  destination is removed and recovery resumes;
- a backup-to-live reverse-rollback collision preserves a competing directory,
  backup, rollback sidecar, and journal, then completes recovery safely;
- a later preparation target with an absent live parent leaves no new parent,
  journal, or sidecar;
- all earlier deterministic ordering, crash recovery, reverse rollback,
  concurrent-consumer, stale-preimage/generation, removal, and lock-order
  cases remain green.

## Current source identities

```text
feae9c197a62bce3c320774cd46d5c2fdfd6f694b6f66e9bd15c129c2d29bf5a  src/csk/locking.py
cd1f3344151014f299561535d90f9337f45d7198fbf10457ae1545966f3454aa  src/csk/transactions.py
110db19650592a690804e5904d945f79a67a3f163fc378cefe758f77b9976f47  tests/test_locking.py
c0ddd310cb7643b969a0c54d7e9a9ac28221c6eb5570dbddfd0f50ba0d3ce21b  tests/test_transactions.py
```

## Final validation ledger

All green handoff gates were run directly as standalone processes against the
current worktree bytes:

1. Focused pytest:
   `/opt/homebrew/bin/python3.11 -m pytest -p no:cacheprovider tests/test_locking.py tests/test_transactions.py -q`
   → exit `0`; `28 passed in 0.98s`.
2. Strict mypy:
   `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/mypy-venv/bin/python -m mypy --cache-dir=/tmp/TASK-260720-z2z795-mypy-cache-final`
   → exit `0`; `Success: no issues found in 56 source files`.
3. Changed-file lint:
   `uvx ruff check src/csk/locking.py src/csk/transactions.py tests/test_locking.py tests/test_transactions.py`
   → exit `0`; `All checks passed!`.
4. Changed-file formatting:
   `uvx ruff format --check src/csk/locking.py src/csk/transactions.py tests/test_locking.py tests/test_transactions.py`
   → exit `0`; `4 files already formatted`.
5. Full pytest:
   `/opt/homebrew/bin/python3.11 -m pytest -p no:cacheprovider -q`
   → exit `0`; `512 passed, 18 skipped in 77.06s`.
6. Package build:
   `/opt/homebrew/bin/python3.11 -m build`
   → exit `0`; built
   `cocoaskills-0.12.6.dev0+g6fc2fd97d.d20260729.tar.gz` and
   `cocoaskills-0.12.6.dev0+g6fc2fd97d.d20260729-py3-none-any.whl`.
7. Tracked-diff whitespace check:
   `git diff --check`
   → exit `0`; no output.

Development-cycle red/repair evidence is retained rather than presented as
green:

- The new focused regressions first exited `1` with five expected failures
  against the reviewed implementation: four missing native-helper failures and
  one absent-parent non-failure. The same selection passed `5` tests after the
  implementation.
- The first changed-file Ruff lint exited `1` with twelve findings; they were
  repaired, and the current exact command exits `0`.
- The first Ruff format check exited `1` for two files; the mechanical format
  changes were applied, and the current exact command exits `0`.
- `git diff --no-index --check /dev/null src/csk/transactions.py` and the same
  command for `tests/test_transactions.py` each exit `1` with no output because
  Git's no-index mode reports the intentional whole-file content difference.
  These are not claimed as passing gates. Ruff lint and format cover both
  untracked files directly; the ordinary tracked `git diff --check` exits `0`.

## Scope and handoff state

The worktree contains only the four task-owned source/test paths as modified or
untracked task work. Build outputs remain ignored. No compiler or installer
policy, dry-run routing, Go UX, pins, staging, commit, publication, or release
operation was added.

