# TASK-260720-3j8pp5 — trusted Go toolchain identity evidence

## Provenance and scope

- Product repository: `/Users/iv/Developer/Wildberries/cocoaskills`.
- Clean, fast-forwarded `main` and `origin/main` base:
  `dd76b570f88339fd1d659c02950e68b17f6ba834`.
- Accepted dependency `TASK-260720-z9j4c9` was `done` with handoff outcomes
  before the task worktree was created.
- Task branch: `task/TASK-260720-3j8pp5-toolchain-identity`.
- Task worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3j8pp5/worktree`.
- Product delta is exactly:
  - `src/csk/builds/toolchain.py`
  - `tests/test_builds_toolchain.py`
- File SHA-256 values:
  - `a2284025742cfbfe53b914924c3c7289172a52b8c190ef6c0b577e7975851077`
    — `src/csk/builds/toolchain.py`
  - `1c8c0faa5fb826d7b34c87bb6093c54cf57b8e18b1aabb10cc0b851ec2ea4d8f`
    — `tests/test_builds_toolchain.py`

The hashes above were captured before the final validation-only commands. The
final status still listed exactly the same two task-owned untracked paths.

## Implemented contract

- An immutable `OperatorSearchPath` is captured before project shim
  augmentation; resolution never re-reads ambient `PATH`.
- Search-order candidates, explicit launchers, and configured roots resolve to
  the real native `<GOROOT>/bin/go[.exe]`; wrappers, relative/empty search
  entries, repository/project-managed launchers (including outward symlinks),
  missing executables, and mismatched reported roots fail closed.
- One operation-private root owns the empty working directory, empty `PATH`,
  home/config/cache/temp roots, and telemetry state. Cleanup runs on probe
  failures, context exit, successful close, and mutation failures.
- Direct execution has closed stdin, no shell, a deadline, and a shared
  streaming stdout/stderr byte budget.
- Exactly these package-independent forms run once:
  1. `go telemetry off`
  2. `go version`
  3. the fixed `go env -json` field vector from the task acceptance criteria
- Probe JSON requires the exact unique string field set. `GOROOT`, host/target,
  version target, `GOTELEMETRY=off`, private `GOTELEMETRYDIR`, Go `1.23` floor,
  manager-tested family `1.25`, and exactly one native tuning value are
  validated and frozen.
- `curator-go-toolchain-v1` walks the complete real GOROOT without following
  links, sorts unsigned UTF-8 protocol paths, frames `D/F/L/V` records with
  uint64 big-endian lengths, normalizes one terminal LF/CRLF version line,
  rejects duplicate/invalid paths, special files/reparse points and unsafe
  links, and re-fingerprints through the last child boundary.
- No package/manager schema field was added. No `go list` or `go build` was
  invoked.

## Authoritative validation ledger

Every command below ran directly as a standalone process. No gate was piped
through `tee`.

| Command | Exit | Exact result |
|---|---:|---|
| `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python -m pytest tests/test_builds_toolchain.py -q` | 0 | `59 passed in 0.22s` |
| `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python -m mypy` | 0 | `Success: no issues found in 57 source files` |
| `PYTHONPATH=src /Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python -m compileall -q src/csk` | 0 | no output |
| task-file 119-column/trailing-whitespace `awk` gate | 0 | no findings |
| `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python -m pytest -q` | 0 | `633 passed, 19 skipped in 77.21s` |
| `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python -m build --wheel` | 0 | built `dist/cocoaskills-0.12.6.dev1+gdd76b570f-py3-none-any.whl`, including `csk/builds/toolchain.py` |
| `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python -m twine check dist/*.whl` | 0 | wheel `PASSED` |
| task-scoped real `probe_toolchain` smoke with `PYTHONPATH=src` | 0 | resolved Go 1.25.5 Darwin/arm64, froze `GOARM64=v8.0`, fingerprinted the complete Homebrew GOROOT as `sha256:69f6b3484a10b288561c7fc66be60945e48b7628978c7baafbaa2ca5c823da0b`, and reported `private_children []` |
| `git status --short --branch` | 0 | only the two task-owned product paths above |
| `git ls-files --others --exclude-standard` | 0 | only the two task-owned product paths above |

The project declares pytest and strict mypy but no Ruff, Black, Pyflakes, or
other lint command. Strict mypy, compileall, and the task-file
line-length/trailing-whitespace gate are green. Native Windows execution was
not available on this macOS host; the module uses only cross-platform stdlib
imports, Windows behavior is isolated through runtime branches, strict mypy is
green, and focused tests cover Windows-shaped executable/layout semantics
where they do not require privileged native symlink creation.

## Recoverable preflight/iteration results

These non-green commands are reported rather than hidden:

- `python --version` exited 127 because this host has no unqualified `python`;
  all authoritative gates used the repository virtualenv interpreter.
- `python3 -m pytest --version` exited 1 because system Python has no pytest.
- The first scoped mypy attempt exited 1 on the pre-existing generated
  `csk._version` import because this fresh worktree had not generated the
  ignored SCM version module. The wheel build generated it; the required full
  `python -m mypy` gate then exited 0 over all 57 source files.
- The first real-smoke attempt exited 1 with
  `ModuleNotFoundError: csk.builds.toolchain` because the shared editable
  virtualenv points at the canonical clone. Re-running against this worktree
  with `PYTHONPATH=src` exited 0; the final smoke repeated that green result.
- `python -m coverage --version`, `python -m pyflakes --version`, and
  `python -m black --version` each exited 1 because those undeclared tools are
  not installed. `command -v ruff` also exited 1. No coverage or undeclared
  formatter/linter claim is made.

Earlier iterative focused runs also exited 0 at 56 and 57 tests; the final
authoritative focused result is the 59-test run above. An earlier full suite
exited 0 with 630 passed / 19 skipped before the final three runner/framing
regressions were added; the final authoritative repository result is
633 passed / 19 skipped.
