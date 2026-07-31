# TASK-260720-z2z795 — rework review verdict

## Verdict

**Changes requested → `to-dev`.** This is ordinary implementation rework. No external blocker, product decision, or human-only architecture decision is required.

Review run: `RUN-260729-5be9e5`  
Accepted base reviewed: `6fc2fd97dbdf40f5b0e46f846eaa0d78a1b33d12`  
Worktree: `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/worktree`

Current source hashes match the producer handoff: `locking.py` feae9c197a62bce3c320774cd46d5c2fdfd6f694b6f66e9bd15c129c2d29bf5a; `transactions.py` cd1f3344151014f299561535d90f9337f45d7198fbf10457ae1545966f3454aa; `test_locking.py` 110db19650592a690804e5904d945f79a67a3f163fc378cefe758f77b9976f47; `test_transactions.py` c0ddd310cb7643b969a0c54d7e9a9ac28221c6eb5570dbddfd0f50ba0d3ce21b.

## Prior findings closed

The rework closes both findings from `TASK-260720-z2z795_review-verdict.md`. Native no-replace mutation is now routed through `renamex_np(RENAME_EXCL)` on macOS, `renameat2(RENAME_NOREPLACE)` on Linux, and `MoveFileExW` without replacement on Windows. The deterministic file and directory boundary tests preserve both source and competing destination. Preparation now refuses an absent live parent and removes already prepared sidecars and the journal without creating namespace residue. Commit and rollback collision recovery tests pass.

## Material finding: Windows journal transitions are not durable

`src/csk/transactions.py:488-519` fsyncs the temporary journal bytes, then publishes a new journal with `os.link` or updates one with `os.replace`. `src/csk/transactions.py:855-857` makes `_sync_directory` a no-op on Windows, and `src/csk/transactions.py:706-709` removes the journal with ordinary unlink. By contrast, target no-replace moves at `src/csk/transactions.py:941-950` explicitly use `MoveFileExW(..., MOVEFILE_WRITE_THROUGH)`.

CPython implements Windows `os.replace` with `MoveFileExW(..., MOVEFILE_REPLACE_EXISTING)` and does not add `MOVEFILE_WRITE_THROUGH`. Microsoft documents `MOVEFILE_WRITE_THROUGH` as the flag that waits for the move to reach disk: https://learn.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-movefileexw . The accepted project reference implements journal creation via durable no-replace, journal updates via `MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH`, and Windows directory flushing in `internal/transaction/journal.go`, `replace_windows.go`, and `durability_windows.go`.

This asymmetry permits a Windows power-loss window in which a target rename is durable but the latest journal phase or target state is not. In the rollback path, for example, target restoration can reach disk while the durable journal still says `committing`; recovery can then reject the restored preimage as a changed committed target instead of resuming rollback. That violates the acceptance requirement for durable commit state and crash recovery. Passing process-crash tests on macOS cannot establish this Windows persistence property.

Required rework:

1. Route new-journal publication through an atomic Windows no-replace move with write-through, and journal replacement through `MOVEFILE_REPLACE_EXISTING | MOVEFILE_WRITE_THROUGH`; keep equivalent fsync plus directory-fsync behavior on POSIX.
2. Make journal removal/recovery metadata transitions durably restartable on Windows, using a write-through tomb/rename or an equivalently strong scheme; do not rely on the current no-op directory sync.
3. Add focused Windows routing tests that pin the exact flags and state-transition helpers, and run the focused transaction suite on a Windows runner in addition to the existing macOS gates.
4. Rerun focused pytest, strict `python -m mypy`, changed-file Ruff/format, full pytest, build, and diff checks; attach exact evidence.

## Independent validation ledger

- Focused pytest: `/opt/homebrew/bin/python3.11 -m pytest -p no:cacheprovider tests/test_locking.py tests/test_transactions.py -q` → exit 0, `28 passed in 0.86s`.
- Strict mypy: task-local Python `-m mypy --cache-dir=/tmp/TASK-260720-z2z795-review2-mypy-cache` → exit 0, `Success: no issues found in 56 source files`.
- Changed-file Ruff check → exit 0, `All checks passed!`.
- Changed-file Ruff format check → exit 0, `4 files already formatted`.
- Full pytest → exit 0, `512 passed, 18 skipped in 77.75s`.
- `python -m build` → exit 0, sdist and wheel built successfully.
- `git diff --check` → exit 0.
- Worktree remained limited to the four task-owned source/test paths and source hashes were unchanged after validation.

The green gates confirm the implemented macOS behavior and close the prior rework. They do not close the Windows durable-journal gap above.