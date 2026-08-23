# TASK-260720-2jfnz6 review verdict — RUN-260730-a67852

## Verdict

CHANGES REQUESTED. Route to `to-dev`. The current Linux quarantine rework fixes the observed owner-control requirement for readable sealed directories, but it does not safely cover the explicit mode-`0000` candidate across the claimed generic Linux support surface.

## Reviewed candidate

- Product worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2jfnz6/worktree`
- Branch: `task/TASK-260720-2jfnz6-protected-posix-build-cache`
- Base: `495ad021847529ce5a544dba415ca2fe19949539`
- HEAD: `0d6ad16fce35c1bd8854511e13766cd236908e3b`
- Current unstaged source hash: `src/csk/builds/cache_posix.py` = `4acb413d2964296350e93e6be2e3ed193c063b95ff0ce6b619585b9d6a1e1112`
- Current unstaged test hash: `tests/test_build_cache_posix.py` = `cd797ddce979611887076f6214c9c747eb50a7ac5103dfda509e84e5e79f4874`
- Delta: 2 tracked files, 108 insertions, 6 deletions. Dependency `TASK-260720-2dnqw2` is `done`.

## Blocking finding R2-1 — mode-0000 recovery assumes an unavailable Linux chmod combination and leaks its reservation

`_unlock_owned_directory_for_move` falls back from an `openat` failure to `os.chmod(name, mode, dir_fd=parent_fd, follow_symlinks=False)` at `cache_posix.py:1315-1321`. Python documents both arguments as platform-optional. CPython 3.11 and 3.14 call `fchmodat` for this combination and translate `ENOTSUP` or `EOPNOTSUPP` into `ValueError: chmod: cannot use dir_fd and follow_symlinks together`. GNU documents that Linux with glibc 2.31 returns `ENOTSUP` even for a non-symlink. The package claims generic POSIX Linux and Python 3.11+, while `_protection_supported` at lines 511-533 does not gate or probe this primitive.

This matters for the acceptance-owned `0000` candidate: the initial directory open can fail, the no-follow chmod can raise `ValueError`, and `_move_aside` at lines 1238-1250 cleans the already-created quarantine reservation only for `OSError`. The exception therefore escapes without a controlled cache error, the reservation remains, and the candidate is neither quarantined nor rebuilt into fresh protected state. An independent seam probe reproduced exactly `exception=ValueError` and `reservation_count=1`.

The new tests at `tests/test_build_cache_posix.py:698-798` monkeypatch `sys.platform` and `os.rename`, but they still execute the Darwin host implementation of `os.chmod`; they do not exercise the Linux capability failure or its cleanup. The current unstaged bytes have no native Linux run. The prior GitHub run `30514706010` covers committed HEAD only and failed all four Ubuntu jobs.

Primary references:

- Python OS API portability contract: https://docs.python.org/3/library/os.html#os.chmod
- CPython 3.11 implementation and `ValueError` path: https://raw.githubusercontent.com/python/cpython/v3.11.15/Modules/posixmodule.c
- CPython 3.14 implementation and `ValueError` path: https://raw.githubusercontent.com/python/cpython/v3.14.4/Modules/posixmodule.c
- GNU fchmodat portability note for glibc 2.31: https://www.gnu.org/software/gnulib/manual/html_node/fchmodat.html
- Red baseline: https://github.com/ivanopcode/cocoaskills/actions/runs/30514706010

## Required rework

1. Make the mode-`0000` unlock capability-safe while preserving rooted no-follow identity checks, or fail closed as unsupported before publication; do not surface raw `ValueError` or `NotImplementedError`.
2. Remove the destination reservation and restore any temporary mode on every unlock failure path, not only `OSError`.
3. Add a regression that forces the initial open failure plus unavailable no-follow chmod and proves controlled failure with zero reservation leakage; retain success/restoration coverage for 0500, 0550, and 0000.
4. Run the current bytes on native Linux across the supported CI path and prove the original four Ubuntu failures are closed.

## Independent gates

- Focused POSIX pytest: `37 passed in 0.35s`.
- Strict mypy: `Success: no issues found in 63 source files`.
- Full repository pytest with accepted conformance root: `886 passed, 6 skipped in 82.74s`.
- `git diff --check`: exit 0.
- Git state remained unchanged: only the producer-owned modifications to `cache_posix.py` and `test_build_cache_posix.py`; the reviewer edited, staged, committed, pushed, or discarded no candidate bytes.

The standalone logbook command is unavailable. This outcome and the appended task notes are the durable reviewer logbook record.