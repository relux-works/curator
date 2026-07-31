# TASK-260720-3t8nr3 rework evidence — revision 3

Developer run: `RUN-260730-6a96ec`

## Workspace

- Repository: `/Users/iv/Developer/Wildberries/cocoaskills`
- Worktree: `.temp/TASK-260720-3t8nr3/worktree`
- Branch: `task/TASK-260720-3t8nr3-transactional-project-hybrid`
- Recorded task base / worktree `HEAD`: `b3a5031ed551b27a298eef486a068b5175beaacc`
- Canonical clean clone: `main == origin/main == b3a5031ed551b27a298eef486a068b5175beaacc`
- Work remains intentionally uncommitted and unstaged for reviewer inspection.

## Reviewer findings closed

### B3 — `$TMPDIR` changed default adapter materialization

`_commit_materialization` now creates its hidden operation-private stage beside
the physical project, outside the checkout but on the adapter destination
filesystem. If that parent cannot host the stage, it falls back to the manager
home; the existing same-device check then conservatively chooses copies when
the fallback is on another filesystem. The process temporary directory is no
longer consulted for materialization, so changing `$TMPDIR` cannot flip
auto-mode output.

Coverage:

- `test_materialization_staging_is_private_and_anchored_to_project_filesystem`
  points Python's process temp root at an unrelated location, captures the real
  materialization stage, and proves that it is a cleaned-up sibling of the
  physical project, outside both the checkout and the process temp root.
- `test_transaction_auto_mode_rejects_cross_device_staging_without_probe`
  supplies deterministic distinct device identities and proves that the
  cross-device branch does not run the symlink probe and chooses the safe
  fallback.
- The prior same-device vector still proves that the probe writes only into
  private staging and never into the live project.

### N5 — post-commit GC lock contention reported a failed install

The success-only post-commit GC remains guarded by `ManagerHomeLock`. A
`LockError` acquiring that maintenance lock now adds a visible
`post-install garbage collection skipped` message to the successful result,
instead of changing a fully committed install into a non-zero CLI outcome.
`LockOrderError` is explicitly re-raised so an internal lock-order defect is
never hidden.

`test_post_commit_gc_lock_contention_does_not_fail_committed_install` makes only
the post-commit acquisition fail, then proves the installed marker is live,
the result has no errors, and the skipped-maintenance message is present.

### N6 — whole-world staging in system temp / RAM

The known O(total installed state) staging copy remains a deliberate
correctness tradeoff, but it no longer lands in the system temp directory.
It lands beside the physical project, with manager-home fallback. The decision
and residual copy-on-write optimization opportunity are recorded in
`LOGBOOK.md`.

### N7 — consumer ledger canonicalizes symlinked paths

`consumers.encode_consumers` intentionally resolves project paths before
sorting and encoding them so transaction preimages do not vary by lexical
symlink route. The first successful install may rewrite a legacy unresolved
entry. This behavior change is now recorded in `LOGBOOK.md` and here.

## Validation

Every gate below ran directly as a standalone process. No command was piped
through `tee` or a pipe chain.

| Gate | Command | Result | Exit |
| --- | --- | --- | --- |
| Expected-red regressions before implementation | `python -m pytest -q tests/test_install.py::test_materialization_staging_is_private_and_anchored_to_project_filesystem tests/test_install.py::test_post_commit_gc_lock_contention_does_not_fail_committed_install` | 2 failed for the intended system-temp and escaping-lock-error reasons | 1 (expected red, honestly failing) |
| Exact B3/N5 regressions after implementation | `python -m pytest -q tests/test_install.py::test_materialization_staging_is_private_and_anchored_to_project_filesystem tests/test_install.py::test_post_commit_gc_lock_contention_does_not_fail_committed_install tests/test_adapters.py::test_transaction_auto_mode_rejects_cross_device_staging_without_probe` | 3 passed in 25.27s | 0 |
| Task transaction vectors | `python -m pytest -q tests/test_installer_transactions.py` | 7 passed in 73.88s | 0 |
| Project / hybrid / closure / adapter / GC | `python -m pytest -q tests/test_install.py tests/test_hybrid_scope.py tests/test_closure_install.py tests/test_adapters.py tests/test_gc.py` | 94 passed in 845.71s | 0 |
| Transactions / locking / cache / planner / activation / CLI | `python -m pytest -q tests/test_transactions.py tests/test_locking.py tests/test_gc.py tests/test_adapters.py tests/test_build_cache_posix.py tests/test_build_planner.py tests/test_build_activation.py tests/test_cli.py` | 295 passed, 9 skipped in 162.21s | 0 |
| Decisive full suite | `python -m pytest -q` | 1134 passed, 98 skipped in 1362.58s | 0 |
| Strict typing | `python -m mypy` | Success: no issues found in 67 source files | 0 |
| Lint | `uvx ruff check` over every changed Python source and test file | All checks passed | 0 |
| Syntax | `python -m py_compile` over every changed Python source and test file | clean | 0 |
| Distribution build | `python -m build --outdir .temp/TASK-260720-3t8nr3/dist` | sdist and wheel built successfully | 0 |
| Patch hygiene | `git diff --check` | clean | 0 |

The orchestrator initially directed this bounded rework to rely on the focused
gates instead of repeating the 22-minute suite. The developer-added checklist
nevertheless named the full suite explicitly; after detecting that evidence
mismatch, the item was unchecked, the exact suite ran green, and only then was
the item checked.

## Platform limits

- No Windows host was available. The existing deterministic concurrency vector
  ran green in the systems gate, but its Windows-specific 30-second branch
  remains CI-only.
- No real second filesystem was mounted in this run. The prior reviewer
  reproduced B3 with a macOS RAM disk; this revision uses deterministic
  synthetic device identities for the cross-device branch and separately proves
  that materialization ignores the process temp root.
