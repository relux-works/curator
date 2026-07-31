# TASK-260720-z2z795 lock-integrity review verdict

Date: 2026-07-30  
Role: reviewer  
Run: `RUN-260730-540109`

## Verdict

**Changes requested → `to-dev`.**

This is ordinary implementation rework. No external blocker, product decision,
approval, or human-only architecture decision is required.

The required pre-verdict goal query returned:

```text
Active Goal: none (run is not goal-bound)
```

The active directive was acknowledged and this review did not modify product
files, stage, commit, push, publish, use SSH, or rely on the old PR #7 CI as
current-byte evidence.

## Reviewed provenance and scope

- Prerequisite `TASK-260720-1pvfj5` is `done`.
- The clean canonical main checkout is `main...origin/main` at
  `d5d16bfcaa2fe43dc994b819c2659512c4fd8f0a`.
- Original task base
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` remains an ancestor of the
  task worktree.
- Worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`
- Committed head:
  `5f8bfbbd6e0c3a00dd0a3de0295e74f05990a0ec`
- Current rework remains unstaged and changes exactly:
  - `src/csk/locking.py`
  - `src/csk/transactions.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`
- Current SHA-256:
  - `locking.py`:
    `7ffd507172f869efd72740de02e2b6ab0d8feb9693826e5da25cd616b9b0aa0b`
  - `transactions.py`:
    `04ac5063eba13d64a3ce520778778cb2ce3f6100dd28e3d98faaae37041ac5b3`
  - `test_locking.py`:
    `f314a7aed5d6d6e691a0139578bbc242b21baf316f7c79c4b5ab5ac107ac9b41`
  - `test_transactions.py`:
    `6938285ac7a377be37ce62c19edeaf9915ef740a885e3f9e7f3e38fd297729cd`

## Prior findings closed

The exact prior reviewer regressions independently pass on current bytes:

```text
14 passed in 0.64s
```

This covers same-instance failed project/build/home re-entry reservation
cleanup, the filesystem-unlink/process-state home handoff window, repeated
concurrent home consumers, preparing entry intent, valid partial-staging
cleanup, completed staging after source loss, and file/directory
mutation/replacement/addition rejection for foreign post-crash staging bytes.

## Material finding 1: transactions can replace the lock files that serialize them

`ManagerHomeLock` uses `<home>/.lock`
(`src/csk/locking.py:404-407`); project and build locks live below
`<home>/locks/projects` and `<home>/locks/builds`
(`src/csk/locking.py:207-217`, `333-340`). Transaction namespace validation,
however, reserves only the journal tree plus target live/sidecar/tomb paths
(`src/csk/transactions.py:652-657`, `2439-2476`). It does not reserve any lock
namespace.

Both `prepare()` and `commit()` check the lock witness only at method entry
(`src/csk/transactions.py:157-160`, `307-311`). A target may therefore be the
active lock file itself. The transaction backs it up, installs arbitrary
desired bytes, reports success, removes the original lock backup, and leaves
the witness invalid.

Deterministic home-lock target probe:

```text
commit_succeeded=True
lock_witness_acquired_flag=True
lock_witness_after_commit=LockError:lock ownership was lost: <temporary>/home/.lock
live_lock_contents=replacement lock bytes
corrupt_lock_left_after_release=True
subsequent_home_acquire=LockError:another csk process holds lock at <temporary>/home/.lock; remove it only after verifying the process is stale
```

Deterministic held-project-lock target probe:

```text
commit_succeeded=True
project_witness_after_commit=LockError:lock ownership was lost: <temporary>/home/locks/projects/<digest>.lock
corrupt_project_lock_left_after_release=True
subsequent_project_acquire=LockError:another csk process holds lock at <temporary>/home/locks/projects/<digest>.lock; remove it only after verifying the process is stale
```

Impact: the transaction can remove the filesystem evidence for its own active
home lock or an outer project lock. While that path is absent, another process
can acquire the same logical lock and enter a supposedly serialized critical
section. Normal release then leaves corrupt lock bytes, permanently blocking
later operations. This violates the project → optional build → home hierarchy
and the single manager-home mutation lock requirement.

Required rework:

1. Treat the canonical home lock and the complete project/build lock state as
   internal reserved namespaces during plan construction and journal loading.
   Reject live targets, sidecars, cleanup tombs, parents, children, physical
   aliases, platform-case aliases, and macOS normalization aliases that
   overlap them.
2. Add deterministic regressions for `<home>/.lock`, held project and build
   lock files, and their lock directories. Rejection must occur before journal
   or target mutation and must preserve every held witness.
3. Keep witness assertions around critical mutation boundaries as defense in
   depth; losing a witness must never be reported as a successful commit or
   recovery.

## Material finding 2: stale-lock restoration can overwrite a new live lock

`_break_stale_lock()` reads a stale record and renames the current lock path to
a unique side path. If the moved bytes no longer match the stale preimage, it
restores them with `stale.rename(self.path)`
(`src/csk/locking.py:178-199`). On POSIX, `Path.rename` replaces an existing
destination.

The race is:

1. reviewer A reads dead lock D;
2. another breaker removes D and live owner C creates the lock path;
3. A renames C's lock to its stale side path;
4. competing owner E acquires the now-absent lock path;
5. A sees that the captured bytes are not D and restores C with
   replacement-capable rename, overwriting E's live lock.

A deterministic mutation-boundary probe on the current macOS bytes produced:

```text
break_stale_result=False
competing_token_before_restore=competing-live
final_lock_token=captured-live
competing_live_lock_overwritten=True
```

Owner E can already be inside the critical section when its lock record is
overwritten. The accepted Go reference deliberately uses stable lock files and
operating-system locks, and states that release never removes a path another
process may already have opened
(`internal/managerlock/managerlock.go:1-3`,
`internal/managerlock/filelock.go:22-65`, `80-99`).

Required rework:

1. Replace create/unlink/stale-rename ownership with stable-file OS locking as
   in the accepted reference, or provide an equivalently race-proof
   cross-platform protocol. A stale claimant must never rename or overwrite a
   lock acquired after its preimage read.
2. If a restoration path remains, it must be atomic no-replace and preserve
   both claimants for fail-closed diagnosis; it must not silently discard a
   live owner.
3. Add deterministic cross-process/barrier regressions for stale-breaker vs
   new-owner acquisition on macOS/Linux and Windows, plus ordinary timeout,
   crash release, and concurrent-consumer coverage.

## Independent validation ledger

- Tool readiness:
  - Git `2.50.1`, ripgrep `15.2.0`, Python `3.11.14`, uvx `0.11.3`,
    task-board `0.23.0`.
  - `task-board version` was an invalid subcommand (exit 1); the supported
    `task-board --version` succeeded (exit 0).
  - The first standalone probe lacked `PYTHONPATH=src` and failed import
    before creating probe state (exit 1); the corrected invocation succeeded.
- Prior-review regression set: `14 passed`, exit 0.
- Contract-targeted deterministic ordering/recovery/reverse rollback/
  concurrency/stale-preimage/lock-order set: `13 passed`, exit 0.
- Focused pytest: `69 passed, 1 skipped in 4.11s`, exit 0. The skip is the
  real-Windows file/tree durability test.
- Strict `python -m mypy`: no issues in `57 source files`, exit 0.
- Ruff lint: `All checks passed!`, exit 0.
- Ruff format: `4 files already formatted`, exit 0.
- Full pytest: `639 passed, 20 skipped in 85.10s`, exit 0.
- Package build: sdist and wheel built successfully, exit 0.
- `git diff --check`: exit 0, no output.
- Post-validation status and all four source hashes are unchanged.

The earlier GitHub Windows matrix covers committed head `5f8bfbbd...`, not the
current unstaged rework. After the two defects above are fixed, rerun focused
pytest, strict mypy, lint/format, full pytest, build, both deterministic
adversarial regressions, and the real cross-platform matrix on the exact
committed bytes before another independent review.
