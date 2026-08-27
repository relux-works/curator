# TASK-260720-2dnqw2 tool readiness

Checked from the task worktree on 2026-07-30.

- `git --version`: exit 0, git 2.50.1
- `rg --version`: exit 0, ripgrep 15.2.0 with PCRE2
- `task-board --version`: exit 0, task-board 0.23.0
- task venv `python --version`: exit 0, Python 3.14.4
- task venv `python -m pytest --version`: exit 0, pytest 9.1.1
- task venv `python -m mypy --version`: exit 0, mypy 2.3.0
- task venv `python -m build --version`: exit 0, build 1.5.0
- task venv `python -m twine --version`: exit 0, twine 7.0.0

The project metadata declares Python 3.11 through 3.14 on macOS, Linux, and
Windows. This local implementation run uses the supported Python 3.14/macOS
combination.
