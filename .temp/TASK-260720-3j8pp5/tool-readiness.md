# TASK-260720-3j8pp5 tool readiness

- `git --version`: exit 0 (`git version 2.50.1 (Apple Git-155)`)
- `task-board --help`: exit 0 (CLI help rendered)
- `rg --version`: exit 0 (`ripgrep 15.2.0`)
- `python3 --version`: exit 0 (`Python 3.14.4`)
- `pytest --version`: exit 0 (`pytest 8.3.4`)
- `python3 -m mypy --version`: exit 1
  - Error: `/opt/homebrew/opt/python@3.14/bin/python3.14: No module named mypy`
  - Decision: do not use the system interpreter for the strict mypy gate; identify and verify the project-declared environment first.
