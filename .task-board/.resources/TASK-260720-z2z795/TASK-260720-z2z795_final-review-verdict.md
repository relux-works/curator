# TASK-260720-z2z795 final independent review verdict

Date: 2026-07-30  
Role: reviewer  
Run: `RUN-260729-216f1f`

## Verdict

**Changes requested → `to-dev`.**

This is ordinary implementation rework. No external blocker, product decision,
approval, or human-only architecture decision is required.

The reviewer run is not goal-bound (`task-board spawn goal` returned
`Active Goal: none`) and has no recorded directives.

## Reviewed provenance

- Accepted predecessor `TASK-260720-1pvfj5` is `done`.
- Original accepted task base:
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`.
- Exact reviewed and published commit:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`.
- Current commit parent / PR base:
  `dd76b570f88339fd1d659c02950e68b17f6ba834`.
- `git verify-commit HEAD` reports a good ECDSA signature for
  `oparin@me.com`, key
  `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM`.
- `git ls-remote` confirms remote `main` at `dd76b570...` and remote task
  branch at `5f8bfbbd...`.
- PR #7 is open, mergeable, and clean against `main`:
  https://github.com/ivanopcode/cocoaskills/pull/7
- The task branch is exactly one commit ahead of `origin/main` and changes
  only:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`
- The task worktree is clean and matches its remote branch.
- Current task source hashes match the latest producer handoff:
  - `locking.py`
    `feae9c197a62bce3c320774cd46d5c2fdfd6f694b6f66e9bd15c129c2d29bf5a`
  - `transactions.py`
    `f50b1d4addb62532155b537303146468965ea9ab2a81ff55b82f3534ef606c49`
  - `test_locking.py`
    `110db19650592a690804e5904d945f79a67a3f163fc378cefe758f77b9976f47`
  - `test_transactions.py`
    `1ae89cd2352d754c972d104366ea96969ab0ddd3ede51aa43a2baac051647c00`

## Prior findings closed

The current bytes close the two findings from
`TASK-260720-z2z795_review-verdict.md`: target moves use native atomic
no-replace primitives, and preparation refuses an absent live parent without
creating unjournaled namespace.

The current bytes also close the journal-publication part of
`TASK-260720-z2z795_windows-durability-review-verdict.md`: Windows journal
create, replace, and canonical-journal-to-removal-tomb transitions use
`MoveFileExW` with the required write-through flags. Exact Windows Python
3.11–3.14 jobs are green.

## Material finding 1: the process-wide build/home lock order is not enforced

`src/csk/locking.py:22-31` stores project/build ownership in thread-local
state and stores only the home witness process-wide. `BuildLock` records the
key only in `_STATE.build` (`src/csk/locking.py:280-298`), while
`ManagerHomeLock._check_order_before_acquire` checks only the acquiring
thread's `_STATE.build` (`src/csk/locking.py:305-311`).

A deterministic two-thread probe held `ProjectLock + BuildLock` in one thread
and acquired `ManagerHomeLock` in another. Actual output:

```text
HOME_ACQUIRED_WHILE_BUILD_HELD
worker_errors=[]
worker_alive=False
```

This violates the acceptance requirement that an optional per-key build lock
is released before the home lock. It also leaves acquisition-pending races
unreserved. The accepted reference prevents this with per-home process-wide
pending/held state in
`internal/managerlock/managerlock.go:389-468`.

Required rework:

1. Track build-key pending/held state and home pending/held state
   process-wide per canonical manager home.
2. Reserve/check the class transition before filesystem lock acquisition and
   release the reservation on every success/failure path.
3. Add deterministic cross-thread regressions for a held build key versus home
   acquisition, pending project/build acquisition versus home acquisition, and
   project/build acquisition versus a pending/held home lock.

## Material finding 2: targets can overlap transaction state and invalidate their own journal

`TransactionEngine` reserves its journal tree at
`src/csk/transactions.py:105-108`, but `_build_journal` rejects only duplicate
exact live paths (`src/csk/transactions.py:428-449`) and journal validation
again checks only exact duplicates (`src/csk/transactions.py:638-676`).
There is no containment, alias, case/normalization, sidecar, or journal-root
independence check.

A deterministic probe used absent target `live = <home>/state`, which contains
the engine's journal root. Actual output:

```text
before_prepare=absent
after_prepare=sha256:e54e8c4293c97b53b8ecb51e68739a833bf771dceb517639ae38ae0bc365f33d
commit_error=stale preimage for 10-context/overlap: got sha256:ce282b8ebe563576c5bacfd322a7340323d70b6e908b7cb1752984f7efbe60fe, expected absent
after_failed_commit=sha256:46ad4d926f01073473a1e17b2e7188568c1de3b9ce7a5810d7d121025726a3f2
journal_exists=False
```

Creating the durable journal mutated the target's own absent preimage.
Rollback removed the journal but left the target present with residue. This
breaks stale-preimage defense, exact rollback, and journal/target isolation.
The accepted reference performs pairwise target/live/sidecar checks and
reserves the journal namespace in
`internal/transaction/namespace.go:87-174` and
`internal/transaction/journal.go:350-360`.

Required rework:

1. Fail before journal or staging mutation when any live/staged/backup/rollback
   namespace overlaps another target namespace or the engine journal tree.
2. Resolve physical aliases and apply platform case/normalization behavior so
   textual variants cannot bypass the check.
3. Add regressions for journal ancestor/descendant targets, target
   parent/child overlap, sidecar/live overlap, symlink aliases, and supported
   case/normalization aliases; every refusal must leave no journal, sidecar, or
   live-target residue.

## Material finding 3: Windows backup cleanup is still not durably ordered before journal removal

`_remove_journal` durably renames the journal to `.json.delete`, then calls
`_discard_sidecars`, which directly removes staging, backup, and rollback paths
with `unlink`/`shutil.rmtree`, and finally unlinks the journal tomb
(`src/csk/transactions.py:737-777`). `_remove_path` relies only on
`_sync_directory` (`src/csk/transactions.py:1226-1235`), while Windows
directory flushing deliberately tolerates `ERROR_ACCESS_DENIED` and
`ERROR_INVALID_HANDLE` (`src/csk/transactions.py:1124-1130`).

A deterministic Windows-routing probe completed a successful transaction with
directory flushes modeled as the allowed no-op result. Actual output:

```text
backup_removed=True
sidecar_cleanup_write_through_moves=[]
journal_tomb_moves=[('txn-cleanup-routing.json', 'txn-cleanup-routing.json.delete', 8)]
```

The canonical backup disappears only through ordinary deletion; there is no
write-through sidecar-to-cleanup-tomb transition or journaled partial-removal
state. A power loss can therefore persist final journal disappearance without
persisting backup deletion, leaving an unowned backup after "successful"
cleanup. Process-crash tests and a green Windows matrix do not establish this
durability ordering. This remains contrary to the requirement to keep and then
durably remove backups only after consumer-last durability.

Required rework:

1. Journal cleanup ownership/progress for each sidecar before deleting it.
2. On Windows, durably move each canonical sidecar to an owned cleanup tomb
   with `MOVEFILE_WRITE_THROUGH` before recursive removal; retain recoverable
   journal state until cleanup ordering is durable.
3. Verify recorded digests/entries and fail closed on changed or unrecorded
   cleanup bytes.
4. Add commit and rollback fault coverage after sidecar tomb publication,
   during partial directory removal, and around final journal removal, plus
   exact Windows flag-routing tests and a real Windows run.

The accepted reference's reusable pattern is in
`internal/transaction/engine.go:754-838` and
`internal/transaction/journal.go:698-778`.

## Independent validation ledger

All commands ran against exact commit `5f8bfbbd...`:

- Focused pytest:
  `/opt/homebrew/bin/python3.11 -m pytest -p no:cacheprovider tests/test_locking.py tests/test_transactions.py -q`
  → exit `0`; `32 passed, 1 skipped in 0.98s`.
- Strict mypy:
  task-local Python `-m mypy --cache-dir=/tmp/TASK-260720-z2z795-final-review-mypy-cache`
  → exit `0`; `Success: no issues found in 57 source files`.
- Changed-file Ruff:
  → exit `0`; `All checks passed!`.
- Changed-file Ruff format:
  → exit `0`; `4 files already formatted`.
- Full pytest:
  `/opt/homebrew/bin/python3.11 -m pytest -p no:cacheprovider -q`
  → exit `0`; `602 passed, 20 skipped in 88.28s`.
- Package build:
  `/opt/homebrew/bin/python3.11 -m build`
  → exit `0`; sdist and wheel built successfully.
- `git diff --check origin/main...HEAD`
  → exit `0`; no output.
- GitHub workflow `30498379391`:
  `14` successful, `0` failed/cancelled/skipped on exact head
  `5f8bfbbd...`; Ubuntu, macOS, and Windows Python 3.11–3.14, strict mypy,
  and build artifacts all passed.
- Final `git status --short --branch`:
  clean task branch tracking the matching remote branch.

The local green gates and exact remote matrix are accepted evidence. They do
not cover or negate the three deterministic contract failures above.

The standalone `logbook` command and a task-board logbook subcommand are not
available in this environment. Findings are persisted in this outcome, the
task notes, and the separate task-scoped validation ledger.
