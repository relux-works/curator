# TASK-260720-z2z795 lock-integrity re-review verdict

Date: 2026-07-30  
Role: reviewer  
Run: `RUN-260730-6f206f`

## Verdict

**Changes requested → `to-dev`.**

This is ordinary implementation rework. No external blocker, product decision,
approval, or human-only architecture decision is required.

The required pre-verdict goal query returned:

```text
Active Goal: none (run is not goal-bound)
```

The active progress directive was observed. This review did not modify product
files, stage, commit, push, publish, use SSH, or claim the old PR #7 CI as
current-byte evidence.

## Reviewed provenance and scope

- Prerequisite `TASK-260720-1pvfj5` is accepted and `done`.
- Original task base
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` remains an ancestor.
- Worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`
- Committed task head:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`
- Current canonical `main...origin/main` is clean at
  `1d28910f5bb276ff58e2a102e06968bd7640abe3`.
- Current rework remains unstaged and limited to:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`
- Current SHA-256:
  - `locking.py`:
    `389b8e5e748dd4fba8657fa40b20c45d4c576a5405a10b7ba11b37f7402bbbfd`
  - `transactions.py`:
    `76975b1c682d5f82834fce3ab4a2174546b1e22301935564d1528b7de5b1e3ab`
  - `test_locking.py`:
    `81dddf366405965a1861cefa665410bfa7c21e33f1b69d140f04994e5aa9086c`
  - `test_transactions.py`:
    `0608671f9c9ee6ff37d85a368eeb0b69a0e8b9c834421b2d72a562ebd64d6cc1`

## Prior lock-namespace finding closed

Planning and journal loading now reserve the canonical home lock file and the
complete project/build lock directories. The existing ancestor/descendant,
physical-alias, platform-case, and macOS normalization checks apply before
journal or target mutation. Held home/project/build witnesses are preserved,
and mutation boundaries assert the caller's home-lock witness.

The new 14-case lock-integrity set, prior 14 reviewer regressions, and 13
contract-targeted regressions all pass independently.

## Material finding: legacy stale breaker can still split the stable lock

The new OS-lock protocol reuses the legacy canonical path and automatically
adopts a dead legacy record:

- `src/csk/locking.py:115-136` checks that the opened descriptor still matches
  the path only **before** legacy takeover, then writes the stable record and
  returns without a second path check.
- `src/csk/locking.py:581-594` accepts any non-v1 legacy record whose PID is
  dead.
- The pre-rework implementation at
  `6fc2fd9:src/csk/locking.py:49-83` can already be resident in another process;
  its stale breaker renames the canonical path and restores with
  replacement-capable `Path.rename` on POSIX.

A deterministic barrier probe reproduced this sequence on current macOS bytes:

1. A current owner opens and OS-locks a dead legacy `.lock`, passes the only
   `_path_matches_fd` check, and pauses before takeover.
2. A simulated already-running legacy breaker renames that inode aside.
3. A second current owner creates and acquires the now-absent canonical path.
4. The first current owner writes its v1 record to the moved inode and reports
   acquired.
5. The legacy mismatch restore replaces the second owner's canonical path.

Observed output:

```text
first_acquired=True
second_acquired=True
second_token_before_restore=cec2a52dea3449a585699a902641a012
captured_token=89a8b118721d4e288a09a77f458e62a8
final_token=89a8b118721d4e288a09a77f458e62a8
second_witness_after_restore=LockError:lock ownership was lost: <temporary>/home/.lock
first_thread_error=[]
```

Both current lock objects have already returned from acquisition, so callers
that rely on the context manager alone can enter one supposedly serialized
critical section concurrently. Existing CLI `GlobalLock` call sites do exactly
that. A later witness failure does not undo already-overlapping mutation.

The added `test_crashed_owner_and_barrier_contenders_share_one_stable_lock`
exercises only current-protocol owners after kernel crash release. It never
places a legacy stale breaker against a new owner, so it does not satisfy the
explicit stale-breaker-vs-new-owner regression boundary.

This violates the single manager-home mutation lock and leaves the POSIX stale
restore race reachable during an in-place protocol upgrade.

## Required rework

1. Make legacy-to-v1 transition fail closed or provide a provably serialized
   migration fence. Automatically adopting a dead legacy record at the same
   path is unsafe while a pre-upgrade breaker may already have read it.
2. Add a post-publication canonical-path/descriptor witness check before
   acquisition is recorded as defense in depth. This alone is not sufficient;
   the migration design must also prevent a legacy restore from replacing a
   later canonical owner.
3. Add a deterministic cross-process/barrier regression that includes the
   legacy stale-breaker sequence above and proves that at most one current owner
   can enter. Cover the supported POSIX platforms and exercise the chosen
   Windows migration behavior.
4. Keep the stable-file invariant after migration: release must never unlink or
   rename the canonical v1 lock file. Do not instruct operators to remove a v1
   stable lock as ordinary timeout recovery.
5. Re-run the new migration regression, the existing 14 + 14 + 13 regression
   groups, focused/full pytest, strict mypy, Ruff lint/format, package build, and
   `git diff --check`.

## Independent validation ledger

- New lock-integrity regressions: `14 passed`, exit 0.
- Prior reviewer regressions: `14 passed`, exit 0.
- Contract-targeted regressions: `13 passed`, exit 0.
- Focused pytest: `82 passed, 1 skipped`, exit 0.
- Strict mypy: no issues in `57 source files`, exit 0.
- Ruff lint: clean, exit 0.
- Ruff format: four files already formatted, exit 0.
- Full pytest: `652 passed, 20 skipped`, exit 0.
- Package build: sdist and wheel built, exit 0.
- `git diff --check`: exit 0.
- Post-validation scope and all four hashes are unchanged.

The focused skip is the real-Windows durability test. No current-byte real
Windows run was performed; earlier remote CI covers only committed head
`5f8bfbbd...` and is not claimed for this rework.
