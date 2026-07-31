# TASK-260720-3c0ss2 — review acceptance cycle 4

## Verdict

Accepted. Verdict branch: `done`. This run is not goal-bound, and no `commit_ack` was supplied.

This verdict covers the preserved cycle-4 current bytes only. It authorizes the commit-owning mover to commit the exact reviewed scope and run origin Windows CI. It does not claim native Windows execution; the exact CI-backed commit must still receive the orchestrator-planned final review before main landing.

## Reviewed identity and scope

- Signed base and HEAD: `2734beff1a0c93d725c00b1c66ef6ad22c3a780a`.
- Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-3c0ss2/worktree`.
- Current product scope: `src/csk/builds/source.py`, new `src/csk/builds/_windows.py`, and `tests/test_build_source.py`.
- `task-board.config.json` is run-routing state and is excluded from product scope.
- Candidate hashes: source.py `e3de0817da12c6157cbc35e1ddce725c1d5b3543e2b352bda6bb8a6b3ec08aaa`; _windows.py `cd38c295c1c718baa5e4d7fbf899ef9a6653393eda5846515297a10ea2987b17`; test_build_source.py `e3489729c6d5186a63178b431b6563c9303baf3c9468d6d9cb51c58e03deba2a`.
- Canonical main was clean at `d5d16bfcaa2fe43dc994b819c2659512c4fd8f0a`, equal to local main and origin/main; the signed candidate base is an ancestor.

## Findings closed

1. `FrozenSnapshot.recheck()` now rejects named streams on the snapshot root before each descendant scan, while the path walker freshly inspects and stream-checks descendants. Persistent root and descendant ADS additions during `use()` fail as `SnapshotMutationError`.
2. `named_data_streams()` checks `FindClose`, captures last error immediately, raises on close-only failure, and preserves enumeration failure as the primary exception with cleanup failure as the explicit cause.
3. The path fallback no longer relies on cached `DirEntry.stat()` identity, and portable path validation occurs before Windows can reinterpret a colon path as an ADS.

## Independent evidence

- Accepted manifest SHA-256: `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`.
- Prior cycle-3 seam probe on current bytes: `root_stream_mutation_failed_closed=True`; `find_close_failure_failed_closed=True`.
- Root/descendant mutation, close failure, next-enumeration cleanup, and primary-error precedence selection: 6 passed; 3 native-Windows cases skipped on macOS.
- Accepted-root task-focused pytest: 207 passed, 4 native-Windows skips.
- Strict `python -m mypy --no-incremental`: no issues in 58 source files.
- Exact full accepted-root pytest with both `CURATOR_CONFORMANCE_ROOT` and `CURATOR_SCHEMA_V6_ROOT`: 716 passed, 5 platform skips.
- An earlier reviewer full run omitted `CURATOR_SCHEMA_V6_ROOT` and therefore reported 715 passed, 6 skipped; the exact corrected command above supersedes it.
- Tracked diff check and untracked Windows-module whitespace check are clean. Hashes remained unchanged after validation.

No product file was modified, staged, committed, pushed, tagged, or released during review. No Go command, SSH, or native Windows execution was used.