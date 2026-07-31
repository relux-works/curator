# TASK-260720-3c0ss2 — cycle-4 Windows boundary rework

## Candidate

- Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-3c0ss2/worktree`
- Branch: `task/TASK-260720-3c0ss2-build-source-identity`
- Signed published base/HEAD: `2734beff1a0c93d725c00b1c66ef6ad22c3a780a`
- Accepted conformance root:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
- Accepted manifest SHA-256:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`

The candidate remains unstaged and uncommitted. The untracked
`task-board.config.json` is run-routing state and is not product scope. No SSH,
Go command, `wb`, commit, push, PR, tag, or release was used.

## Reviewer findings closed

1. `FrozenSnapshot.recheck()` now enumerates/rejects Windows named data streams
   on the snapshot root before every descendant rescan. Root and descendant
   stream additions during `FrozenSnapshot.use()` are translated to
   `SnapshotMutationError`.
2. `_windows.named_data_streams()` checks the Boolean `FindClose` result and
   captures `ctypes.get_last_error()` immediately on failure. A close-only
   failure raises `OSError`. If enumeration and cleanup both fail, enumeration
   remains the primary `OSError` and the close `OSError` is its explicit cause.

This follows the Microsoft contracts for
[`FindFirstStreamW`](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-findfirststreamw),
[`FindNextStreamW`](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-findnextstreamw),
and
[`FindClose`](https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-findclose).

## Regression coverage

- Platform-independent root recheck and post-callback root mutation seams.
- Platform-independent post-callback descendant mutation seam using the
  Windows path walker.
- Native Windows root-directory ADS rejection before freeze.
- Native Windows root and descendant ADS additions during
  `FrozenSnapshot.use()`.
- Documented first-call `ERROR_HANDLE_EOF` and `ERROR_INVALID_PARAMETER`.
- Unexpected first and next enumeration failures.
- Successful cleanup after next-enumeration failure.
- `FindClose` failure and deterministic enumeration/cleanup error chaining.
- Prior cached-`DirEntry.stat()` physical-identity regression remains covered.

## Exact gate ledger

| Gate | Result | Exit |
|---|---:|---:|
| Expected-red `tests/test_build_source.py` before production fix | 4 failed, 23 passed, 3 skipped | 1 |
| Final `tests/test_build_source.py` | 28 passed, 4 native-Windows skips | 0 |
| Reviewer cycle-3 seam probe | root mutation and close failure both fail closed | 0 |
| Accepted-root task-focused pytest | 207 passed, 4 skips | 0 |
| Strict `python -m mypy` | 58 source files clean | 0 |
| Full accepted-root pytest | 716 passed, 5 skips | 0 |
| `python -m build` | wheel and sdist built | 0 |
| `python -m twine check dist/*` | all six present distributions passed | 0 |
| First distribution inventory scratch assertion | checker used wheel-relative paths for sdist; corrected below | 1 |
| Corrected wheel/sdist source inventory | `source.py` and `_windows.py` present | 0 |
| `python -m compileall -q src/csk tests` | no diagnostics | 0 |
| `git diff --check` | no diagnostics | 0 |
| Untracked `_windows.py` whitespace assertion | clean | 0 |

All repository Python gates used the existing project environment via
`PATH=/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin:$PATH`.

The native Windows tests are skipped on this macOS host. Origin Windows CI was
not run because the orchestrator explicitly sequenced it after this
pre-publication review: the current candidate must remain unstaged/uncommitted,
and the workflow has no dispatch path for an uncommitted diff. No native
Windows result is claimed.

## Candidate file SHA-256

- `src/csk/builds/source.py`:
  `e3de0817da12c6157cbc35e1ddce725c1d5b3543e2b352bda6bb8a6b3ec08aaa`
- `src/csk/builds/_windows.py`:
  `cd38c295c1c718baa5e4d7fbf899ef9a6653393eda5846515297a10ea2987b17`
- `tests/test_build_source.py`:
  `e3489729c6d5186a63178b431b6563c9303baf3c9468d6d9cb51c58e03deba2a`
