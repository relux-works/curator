# TASK-260720-8nxlgx tool readiness

Checked 2026-07-30 from the preserved task worktree at base
`15860e3f309888845b9271a257fb95f7c2825b56`.

- `task-board m set_status(...)`: exit 0.
- `git status`, `git rev-parse`, `git diff --check`: exit 0.
- Bare `python` and `python -m ...`: exit 127 because no global `python`
  command is installed. This is expected for this checkout and is not used for
  validation.
- Shared project interpreter
  `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python`: Python 3.14.4,
  exit 0.
- Shared interpreter `pytest --version`: pytest 9.0.3, exit 0.
- Shared interpreter `mypy --version`: mypy 2.1.0, exit 0.
- Shared interpreter `build --version`: build 1.5.0, exit 0.
- Shared interpreter `twine --version`: twine 6.2.0, exit 0.
- Shared interpreter `python -m ruff --version`: exit 1 because Ruff is not
  installed in the shared environment.
- Bare `ruff --version`: exit 127 because no global Ruff executable is
  installed.
- `uvx ruff --version`: Ruff 0.16.0, exit 0; use this exact entry point.
- `uv --version`: uv 0.11.3, exit 0.
- Native `ssh win` readiness: exit 0; remote reports PowerShell 5.1.19041.6456
  and Python 3.14.4.

No validation workflow relies on a failed entry point.
