# TASK-260720-2x6mjn Windows CI rework evidence

## Review handoff

- Accepted base SHA: `323ea47db821641bc68f2b9054b50e82642a6df2`
- Rework commit: `b3a5031ed551b27a298eef486a068b5175beaacc`
- Branch: `task/TASK-260720-2x6mjn-pure-build-planner`
- Pull request: <https://github.com/ivanopcode/cocoaskills/pull/15>
- Green CI run: <https://github.com/ivanopcode/cocoaskills/actions/runs/30554363746>
- Push target: `origin` (`ivanopcode/cocoaskills`) only; `intranet` was not pushed.

## Diagnosis and implementation

Windows Python 3.12+ can expose different `st_ctime` meanings for path-based
`stat` and descriptor-based `fstat`. Comparing their entire metadata tuples
therefore reported a false `concurrent_state_change` even when the opened file
was unchanged.

The cross-API opening check now proves same-file identity with
`os.path.samestat` and compares the portable invariants needed by the read:
file type, size, and nanosecond modification time. Full metadata comparisons,
including `ctime`, remain in both same-API before/after checks. A post-read
`lstat` recheck also remains mandatory, so replacement with identical bytes
and restored `mtime` is still rejected as a concurrent change.

The two real-install/no-Go fixtures now expose the existing Git executable
under its native basename (`git.exe` on Windows) in an isolated `PATH`, while
asserting that Go remains absent. This preserves the minimum executable needed
by real install behavior without weakening the test premise.

No workflow matrix, skip, or xfail was changed.

## Changed files

- `src/csk/builds/planner.py`
- `tests/conftest.py`
- `tests/test_build_planner.py`
- `tests/test_install.py`
- `tests/test_global_install.py`
- `LOGBOOK.md`

Diff summary: 6 files changed, 152 insertions, 19 deletions.

## Test-first evidence

- Expected-red regression:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q tests/test_build_planner.py::test_filesystem_generation_probe_accepts_windows_path_fd_ctime_split`
  exited `1`. Expected rationale: before the product fix, simulated Windows
  path/fd `ctime` divergence raised
  `BuildPlanningError: concurrent_state_change`.
- Portable no-Go fixture tests:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q tests/test_install.py::test_real_install_does_not_plan_builds_or_require_go tests/test_global_install.py::test_global_real_install_does_not_plan_builds_or_require_go`
  exited `0`: 2 passed.

## Local validation

Every command below was run as a standalone process.

- Focused affected modules:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q tests/test_build_planner.py tests/test_install.py tests/test_global_install.py tests/test_cli.py`
  exited `0`: 151 passed in 39.64s.
- Full suite:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest -q`
  exited `0`: 1120 passed, 98 skipped in 97.28s.
- Strict typing:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m mypy`
  exited `0`: no issues in 67 source files.
- Bytecode validation:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m compileall -q src tests`
  exited `0`.
- Indentation validation:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m tabnanny src tests`
  exited `0`.
- Worktree whitespace gate:
  `git diff --check`
  exited `0`.
- Staged whitespace gate:
  `git diff --cached --check`
  exited `0`.
- Post-commit package build:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m build`
  exited `0`; built the sdist and wheel for
  `cocoaskills-0.12.6.dev19+gb3a5031ed`.
- Post-commit package metadata:
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m twine check dist/*`
  exited `0`; all six local distributions passed.

## Remote validation

- `git push origin task/TASK-260720-2x6mjn-pure-build-planner`
  exited `0`, updating `323ea47..b3a5031`.
- Interactive `gh run watch 30554363746 --repo ivanopcode/cocoaskills --exit-status`
  was manually interrupted after noisy TTY polling and exited `1`; it was not
  treated as a passing gate.
- The subsequent standalone terminal-status query
  `gh run view 30554363746 --repo ivanopcode/cocoaskills --json status,conclusion,url,headSha,jobs`
  exited `0` and reported `status=completed`, `conclusion=success`, head SHA
  `b3a5031ed551b27a298eef486a068b5175beaacc`.
- All CI jobs passed: Windows Python 3.11, 3.12, 3.13, and 3.14; Linux and
  macOS Python matrix jobs; strict mypy; and artifact build.

The pushed worktree is clean and synchronized with its `origin` tracking
branch.
