# TASK-260720-3c0ss2 — review verdict cycle 3

## Verdict

Changes requested. Verdict branch: `to-dev`.

This run is not goal-bound: `task-board spawn goal RUN-260730-6463af`
reported `Active Goal: none`. No `commit_ack` was supplied.

The cached `DirEntry.stat()` identity defect is correctly removed: the fallback
walker retains only entry names and obtains current physical identity with
no-follow `os.lstat()`. The exact accepted conformance/task suite and strict
mypy pass. Acceptance is nevertheless blocked by two fail-closed gaps in the
new Windows named-stream boundary.

## Findings

### P1 — persistent named-stream mutation on the snapshot root escapes recheck

`FrozenSnapshot.recheck()` in `src/csk/builds/source.py:97` performs a raw
`os.lstat()` of the root and then scans descendants. It does not invoke
`_reject_windows_named_streams()` for the root. Root stream enumeration occurs
only through `_validated_root_stat()` during initial `freeze_snapshot()`;
descendant enumeration occurs in `_collect_path_entries()`.

The independent seam probe forced the Windows path walker, armed a simulated
root named stream after the initial freeze, and called `recheck()`. Its exact
trace was:

```text
root_checks_during_freeze=('.', 'file', '.')
checks_during_recheck=('file',)
root_stream_mutation_failed_closed=False
```

This is a real Windows state, not a synthetic path shape. Microsoft documents
that directories have no unnamed `$DATA` stream by default but may have named
data streams:
https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-findfirststreamw

Therefore a build child can add a persistent named stream to the snapshot root
and the final `FrozenSnapshot.use()` recheck can accept it. That contradicts
the task requirement that mutation before the last build child exits fail
closed.

Required rework:

- enumerate/reject root named streams on every `recheck()`, translating the
  invalid-root condition to `SnapshotMutationError`;
- add a platform-independent call-boundary regression for root enumeration on
  recheck/use;
- add native Windows regressions for a root-directory ADS present before
  freezing and one added during `FrozenSnapshot.use()`.

### P2 — `FindClose` failure is silently accepted

`src/csk/builds/_windows.py:64-65` calls `FindClose(handle)` in `finally` but
does not inspect its Boolean result. The independent Win32 seam returned zero
from `FindClose`, set error 5, and observed:

```text
find_close_calls=[42]
find_close_failure_result=()
find_close_failure_failed_closed=False
```

Microsoft documents that zero means `FindClose` failed and that callers must
obtain the extended error with `GetLastError`:
https://learn.microsoft.com/en-us/windows/win32/api/fileapi/nf-fileapi-findclose

Required rework:

- check the close result, capture `ctypes.get_last_error()` immediately, and
  raise `OSError` so the snapshot wrapper fails closed;
- preserve deterministic chaining when enumeration and cleanup both fail;
- extend the platform-independent seam for documented first-call EOF and
  unsupported-filesystem results, unexpected `FindFirstStreamW` failure,
  unexpected `FindNextStreamW` failure with cleanup, and `FindClose` failure.

The existing `ERROR_HANDLE_EOF` and `ERROR_INVALID_PARAMETER` handling matches
the documented `FindFirstStreamW` contract, and unexpected first/next
enumeration errors already raise.

## Independent validation

- Candidate base/HEAD:
  `2734beff1a0c93d725c00b1c66ef6ad22c3a780a`, with a valid Git signature.
- Candidate product diff:
  `src/csk/builds/source.py`, new `src/csk/builds/_windows.py`, and
  `tests/test_build_source.py`.
- `task-board.config.json` is an untracked run-routing file created during this
  reviewer run; it is not candidate scope and must not be committed.
- Accepted manifest SHA-256:
  `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
- Accepted conformance/task-focused pytest: `198 passed, 1 skipped`, exit 0.
- Strict `python -m mypy`: 58 source files clean, exit 0.
- Tracked diff check and new-module patch whitespace check: exit 0.
- Candidate file SHA-256:
  - `source.py`: `ee9eedff47686798820ab95e47aab8c53c814a28da8489415d9f78dd44caf977`
  - `_windows.py`: `1065eac7cdc6fa7240f1d1d5f8d55cc7a5a4e1c7fae20923bc1b678b4696c14d`
  - `test_build_source.py`: `126ce5c0b2fe56bfed39dcdfecdf344ccf3d880b12c0b4f51a229bd6b423fa70`

Native origin Windows CI remains intentionally sequenced after a reviewed fix
commit. This verdict requests ordinary implementation rework; it is not a
stop-the-line blocker.
