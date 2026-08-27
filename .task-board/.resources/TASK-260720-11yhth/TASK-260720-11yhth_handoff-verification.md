# TASK-260720-11yhth handoff verification

## Preserved implementation

- Worktree: `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-11yhth/worktree`
- Branch: `task/TASK-260720-11yhth-command-runtime-activation`
- Recorded base SHA: `11160f642d65a8daf3fbcca5401dca5ec80440f9`
- Recovery respected the orchestrator instruction to preserve the existing
  product and test delta. No CocoaSkills product or test file was changed,
  staged, committed, or pushed during handoff verification.
- The task delta remains:
  `src/csk/shims.py`, `src/csk/global_bins.py`, `src/csk/installer.py`,
  `tests/test_shims.py`, and new `tests/test_build_activation.py`.

## Standalone local gates

| Command | Exit | Result |
|---|---:|---|
| `.venv/bin/python -m pytest -q tests/test_shims.py tests/test_build_activation.py` | 0 | 100 passed, 6 Windows-only skips |
| `.venv/bin/python -m mypy` | 0 | Success: no issues found in 65 source files |
| `uvx ruff check src/csk/shims.py tests/test_shims.py tests/test_build_activation.py` | 0 | All checks passed |
| `.venv/bin/python -m compileall -q src` | 0 | Passed |
| `.venv/bin/python -m build` | 0 | sdist and wheel built |
| `.venv/bin/python -m twine check dist/*` | 0 | sdist and wheel passed |
| `git diff --check` | 0 | Passed |
| `test ! -e uv.lock` | 0 | Passed |

An additional helper-only Ruff diagnostic,
`uvx ruff check src/csk/global_bins.py src/csk/installer.py --ignore
SIM103,BLE001,I001,UP017`, exited 1 on `ISC004` at the untouched
`src/csk/global_bins.py:54`. That line is outside the task diff and the recovery
directive forbids unrelated product cleanup. This expected-red diagnostic is
not presented as a passing gate; the rewritten shim implementation and focused
tests are clean under the task-scoped Ruff command above.

## Native Windows evidence

The existing board attachment
`TASK-260720-11yhth_native-windows-runs.log` was materialized and verified at
SHA-256 `9f0c0593b51cb2c21962b04694f5b21d7c3c6eb5f06b26a016cef52c9937e430`.
It records:

- focused shim and build activation pytest: exit 0, 65 passed / 41 skipped;
- explicit launcher execution selection: exit 0, 7 passed / 6 skipped /
  91 deselected, including argument forwarding, no install-time execution,
  and nonzero exit propagation;
- full pytest excluding the unrelated PowerShell-policy test: exit 0,
  852 passed / 183 skipped / 1 deselected;
- the unexcluded full pytest is honestly red at exit 1 solely because the host
  CurrentUser execution policy denies the unrelated generated `.ps1` file;
- native mypy is honestly red at exit 1 with pre-existing platform-stub errors
  outside `shims.py`, `global_bins.py`, and `installer.py`; strict configured
  mypy is green on macOS as shown above.

## Cleanup limitation

A read-only `ssh -o BatchMode=yes -o ConnectTimeout=5 win ...` probe exited 255
with a connection timeout. The Windows validation host was therefore not
reachable, so the task-scoped remote directory
`C:\Users\admin\csk-11yhth` was not removed.

## Handoff conclusion

The existing implementation and tests cover incomplete runtime reuse,
project/global immutable built-command activation on Unix and Windows,
injection rejection, mixed command collisions and stale removal, install-time
non-execution, and explicit argv/nonzero-exit propagation. Focused pytest,
strict mypy, lint, compilation, package build, metadata, and diff hygiene are
green for review.
