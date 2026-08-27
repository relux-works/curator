# Review verdict — TASK-260720-th0jdi

Verdict: **changes requested** (`to-dev`). The currentness and normal-install repair paths are substantially complete and the declared gates pass, but the GC implementation can delete state that it has not safely classified. These are ordinary implementation/test defects; no human-only decision or external blocker is present.

## Reviewed provenance

- Canonical clean `main` and `origin/main`: `07655553cebcf867bbe58629de98e77644606c85`.
- Exact task/PR head: `f3c1254e4fe7958c720cbef096a4ef00103d43a2`; merge-base is the recorded base above; task worktree was clean before and after review.
- Commit signature: good ECDSA signature for `oparin@me.com`.
- Dependency handoff `TASK-260720-g7kgox` is accepted/done and is present in the base.
- [PR #18](https://github.com/ivanopcode/cocoaskills/pull/18) is open, non-draft, mergeable/CLEAN, and selects the exact reviewed head.
- [CI run 30676739989](https://github.com/ivanopcode/cocoaskills/actions/runs/30676739989) completed successfully for the exact reviewed head.

## Findings

### P1 — A valid in-flight journal can lose every marker generation without making the mark phase uncertain

`TransactionEngine.referenced_install_marker_paths()` flattens all possible generations for each context target into an undifferentiated path set (`src/csk/transactions.py:421-455`). GC then scans each path independently (`src/csk/gc.py:166-177`), while a missing directory is explicitly treated as success/no warning (`src/csk/gc.py:308-318`). It therefore cannot prove that at least one valid marker generation remains for a journaled context target.

Independent reproduction prepared a valid new-install transaction, removed its only marker-bearing desired sidecar, aged the referenced build, and ran GC under the home lock. Observed result:

```text
{'builds_removed': 1, 'warnings': [], 'entry_exists': False}
```

The journal remained valid and in flight, but the referenced build was swept. This contradicts the requirement to mark in-flight journals and fail safe for unreadable/uncertain journal state (`curator-spec@f5d7673039226ab81de2f4f87e2155ae995c4df3`, `profiles/manager.md:673-684`). The existing intact-sidecar test (`tests/test_build_gc.py:138-185`) does not exercise the disappeared-generation branch.

There is a related error-boundary gap: `_journal_ids()` performs `lstat()`, `iterdir()`, and file probes without converting general `OSError` to `TransactionError` (`src/csk/transactions.py:1505-1536`), while `_collect_locked()` catches only `TransactionError` (`src/csk/gc.py:166-181`). An unreadable journal root can therefore abort collection instead of returning a reported fail-safe maintenance result.

Required correction: preserve per-journal/per-context-target generation groups; require at least one safely read, valid marker generation for every relevant target, and mark the whole pass uncertain if none can be proved. Convert journal filesystem uncertainty into a reported retain-all result. Add vanished-sidecar, unreadable-root, and generation-race tests.

### P1 — Protected cache GC classifies one entry identity and retires a later pathname occupant

On POSIX, GC opens and validates an entry, checks age, and closes the handle (`src/csk/builds/cache_posix.py:609-648`); it then calls pathname-based `_move_aside()` (`src/csk/builds/cache_posix.py:649-656`). `_move_aside()` reopens whatever currently occupies that name and receives no expected identity from the classified handle (`src/csk/builds/cache_posix.py:1559-1645`).

A deterministic exchange harness replaced the selected old entry with a valid young entry immediately after the final stable-state check. Observed result:

```text
{'swapped': True, 'builds_removed': 1, 'warnings': [],
 'young_selected_entry_exists': False, 'old_verified_entry_exists': True}
```

The grace/provenance decision applied to object A, but object B was moved and deleted. That violates “sweeps only unreferenced ... entries older than ... grace” and the protected-boundary requirement (`profiles/manager.md:673-684`).

The Windows implementation has the same identity discontinuity: `_inspect_gc_entry()` returns a verified identity (`src/csk/builds/cache_windows.py:2374-2426`), but `collect()` discards it before calling `_move_aside()` (`src/csk/builds/cache_windows.py:850-890`); `_move_aside()` validates only the newly opened occupant (`src/csk/builds/cache_windows.py:2544-2584`).

Required correction: bind retirement to the exact classified object/handle or pass and atomically enforce its expected identity through quarantine on both backends. Add deterministic root/entry exchange tests proving a replacement—especially a young replacement—is never retired under the previous occupant’s decision.

### P2 — Runtime orphan GC does not recognize the names emitted by the current staging code

The collector recognizes only `.<name>.tmp-<pid>` and `.<name>.backup-<pid>` (`src/csk/gc.py:19-21`). Current staging always appends another numeric suffix through `_unique_path()` (`src/csk/shims.py:835-840`), producing names such as `.<name>.tmp-<pid>-0` (`src/csk/shims.py:111`), and runtime replacement creates `.<commit>.stale-<pid>-0` (`src/csk/shims.py:643`). Neither shape matches the collector.

Independent reproduction left dead-PID examples and called `sweep_orphans`; both remained:

```text
{'.deadbeef.tmp-99999999-0': True,
 '.deadbeef.stale-99999999-0': True}
```

This exact naming mismatch was previously routed to this task by accepted task `TASK-260720-11yhth`, so it is also a missed dependency handoff and breaks the requested runtime-GC compatibility.

Required correction: make orphan recognition/traversal match every current runtime staging, backup, and stale name (including allocation indexes), without broadening deletion to unrelated dotfiles. Add exact-name and live/dead-PID tests.

### Coverage gap — the new Windows maintenance paths are not independently exercised

Every new protected build-cache GC case is guarded by `POSIX_BUILD_GC` (`tests/test_build_gc.py:16-19,48-262`), and the v2 build-currentness cases use `POSIX_BUILD_VECTOR` (`tests/test_installer_transactions.py:27-30`; usages in `tests/test_build_currentness.py`). The successful Windows CI jobs therefore do not execute the new Windows `collect()` path or the v2 currentness vectors. No deterministic collector exchange/fault test exposed the P1 gap above.

Required correction: add native Windows status/GC coverage and backend-specific fault/race coverage, then rerun the focused maintenance suite, strict mypy, and cross-platform CI.

## Acceptance-criteria assessment

The review found the currentness side aligned with the contract: marker v1 schemas 1–5 remain supported; v2 replans from the raw snapshot and static context; complete input/key comparison covers target, toolchain, and `policy.execution_policy`; legacy/reserved policy keys miss; protected receipt/artifact bytes and managed shims are checked; capability evidence is surfaced but excluded from the verdict; and normal install rebuilds untrusted/corrupt state. This matches `profiles/manager.md:329-340,444-463,657-671` and `decisions/0004-compile-only-build-drivers.md:53-75` at the cited spec commit.

GC does acquire the manager-home lock, includes project/global/hybrid stores, preserves marker-v1 runtime/snapshot behavior, and applies an age gate. It does **not** yet satisfy the active-journal fail-safe, exact protected-object retirement, runtime-orphan compatibility, or required cross-platform/fault coverage branches. The implementation is therefore not accepted as a whole.

## Independent validation

```text
python -m pytest tests/test_build_currentness.py tests/test_build_gc.py \
  tests/test_status.py tests/test_gc.py tests/test_installer_transactions.py \
  tests/test_global_install_transactions.py -q
61 passed in 80.01s

python -m mypy
Success: no issues found in 68 source files

python -m compileall -q src tests        # passed
git diff --check 0765555..f3c1254        # passed
```

No dedicated formatter/linter is declared in `pyproject.toml`; strict mypy is the declared static gate. No product code was modified during review.
