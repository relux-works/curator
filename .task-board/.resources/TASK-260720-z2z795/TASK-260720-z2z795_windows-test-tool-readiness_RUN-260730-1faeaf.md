# TASK-260720-z2z795 Windows test rework tool readiness

Run: `RUN-260730-1faeaf`

- `task-board --help`: exit 0
- `task-board --version`: exit 0, version 0.23.0
- `git --version`: exit 0, version 2.50.1
- `rg --version`: exit 0, version 15.2.0
- `python --version`: exit 127, unqualified `python` is unavailable
- `python3 --version`: exit 0, version 3.14.4
- `pytest --version`: exit 0, version 8.3.4
- `mypy --version`: exit 127, unqualified `mypy` is unavailable
- `ruff --version`: exit 127, unqualified `ruff` is unavailable
- `.venv/bin/python --version`: exit 127, the task worktree has no local virtual environment
- `/opt/homebrew/bin/python3.11 --version`: exit 0, version 3.11.14
- `/opt/homebrew/bin/python3.11 -m pytest --version`: exit 0, pytest 8.3.4
- `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python --version`: exit 0, version 3.14.4
- `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m pytest --version`: exit 0, pytest 9.0.3
- `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m mypy --version`: exit 0, mypy 2.1.0
- `/Users/iv/Developer/ReluxWorks/cocoaskills-production/.temp/TASK-260720-z2z795/mypy-venv/bin/python -m mypy --version`: exit 0, mypy 2.3.0
- `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python -m build --version`: exit 0, build 1.5.0
- `uvx ruff --version`: exit 0, Ruff 0.16.0

Qualified Python and module entry points above are the gate routes for this run.
