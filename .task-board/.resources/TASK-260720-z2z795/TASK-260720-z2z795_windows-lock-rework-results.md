# TASK-260720-z2z795 Windows stable-lock rework outcome

Run: `RUN-260730-bcc707`  
Date: `2026-07-30`  
Input commit: `108ae51e549566b0b7e1ff4a8a2424e8cb4fcc77`

## Outcome

The GitHub Windows matrix failure was caused by opening the stable lock
descriptor without `O_BINARY`. Windows CRT text translation changed the
trailing journal newline from LF to CRLF on disk and translated it back on
read. `_read_lock_fd_bytes()` bounds its read with the physical `fstat()` size,
so the translated read returned one logical byte fewer and the next read
reported EOF. The acquisition then correctly failed closed with
`lock publication witness was lost`.

`src/csk/locking.py` now adds the platform-optional `O_BINARY` flag to every
stable lock descriptor. The stable OS-lock handle, publication token,
path/descriptor identity checks, legacy-breaker fence, and repeated witness
checks remain intact.

Focused test rework also aligns with Windows byte-range lock semantics:

- a deterministic regression requires every stable lock descriptor to include
  `O_BINARY`;
- held lock records are inspected through the owning descriptor instead of a
  second handle, which Windows correctly denies for an exclusively locked
  range;
- the publication race test retains the POSIX rename/replacement attack and
  verifies on Windows that the stable open handle blocks rename until release;
- the witness-loss fault writes through the owning descriptor, so the
  mutation-boundary assertion remains exercised on every platform.

## Current-byte validation

- Expected-red binary-mode regression: `1 failed`, exit 1.
- Fixed binary-mode regression: `1 passed`, exit 0.
- Python 3.11 Windows-semantics selection: `21 passed`, exit 0.
- Python 3.14 Windows-semantics selection: `21 passed`, exit 0.
- Prescribed lock-integrity regressions: `14 passed`, exit 0.
- Contract-targeted transaction regressions: `13 passed`, exit 0.
- Focused locking/transaction pytest: `125 passed, 4 skipped`, exit 0.
- Strict mypy: no issues in `62 source files`, exit 0.
- Ruff lint and format checks: exit 0.
- Full pytest: `804 passed, 58 skipped`, exit 0.
- Package sdist and wheel build: exit 0.
- `git diff --check`: exit 0.

Exact commands, exit codes, hashes, and diagnostic boundaries are recorded in
`TASK-260720-z2z795_windows-lock-rework-validation.log`.

## GitHub/Windows boundary

Read-only inspection of GitHub Actions run `30515699781` at the input commit
confirmed that all eight macOS/Linux Python jobs passed and all four Windows
Python jobs failed. The supplied Windows/Python 3.14 log recorded
`787 passed, 42 skipped, 150 failed`, dominated by the text-mode witness read;
the separate path-rename adversarial case received `WinError 32`, confirming
the stable Windows handle already prevents rename.

The patched bytes remain unstaged and uncommitted under the explicit task
constraint. A current-byte GitHub Windows matrix therefore cannot be launched
from this developer run and is not claimed. The commit-owning mover must publish
the reviewed three-file patch and rerun the four Windows matrix jobs.

## Provenance and scope

- Worktree:
  `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`
- HEAD and remote task ref:
  `108ae51e549566b0b7e1ff4a8a2424e8cb4fcc77`
- Original accepted base
  `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12` remains an ancestor.
- Modified only:
  - `src/csk/locking.py`
  - `tests/test_locking.py`
  - `tests/test_transactions.py`
- No stage, commit, push, tag, release, pin, SSH, host-management, schema,
  compiler-policy, installer-policy, or Go UX changes were performed.

The standalone `logbook` executable remains unavailable (exit 1), so the
finding and its evidence are persisted in this outcome and on the task board.
