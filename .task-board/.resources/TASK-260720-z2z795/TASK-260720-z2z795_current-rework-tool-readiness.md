# TASK-260720-z2z795 current rework tool readiness

Run: `RUN-260730-b00eb8`  
Date: `2026-07-30`

| Command | Exit | Version |
| --- | ---: | --- |
| `task-board --version` | 0 | `task-board version 0.23.0 (commit beec6e9, built 2026-07-27T23:36:53Z)` |
| `git --version` | 0 | `git version 2.50.1 (Apple Git-155)` |
| `rg --version` | 0 | `ripgrep 15.2.0` |
| `/opt/homebrew/bin/python3.11 --version` | 0 | `Python 3.11.14` |
| `/opt/homebrew/bin/python3.11 -m pytest --version` | 0 | `pytest 8.3.4` |
| task mypy venv `python -m mypy --version` | 0 | `mypy 2.3.0 (compiled: yes)` |
| `uvx ruff --version` | 0 | `ruff 0.16.0` |
| `/opt/homebrew/bin/python3.11 -m build --version` | 0 | `build 1.2.2.post1` |

The unqualified `python3` is Python 3.14.4 and does not provide mypy. All
project tests/builds use the explicitly qualified Python 3.11 interpreter;
strict mypy uses the existing task-scoped mypy virtual environment.
