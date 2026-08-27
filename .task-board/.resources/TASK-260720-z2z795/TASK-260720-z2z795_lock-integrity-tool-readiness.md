# Tool readiness

Checked 2026-07-30 before lock-integrity rework:

- `task-board --version` — exit 0; task-board 0.23.0 (commit beec6e9).
- `git --version` — exit 0; git 2.50.1.
- `rg --version` — exit 0; ripgrep 15.2.0 with PCRE2.
- `python --version` — exit 127; the unqualified `python` executable is not
  installed, so no workflow below assumes it is available.
- `python3 --version` — exit 0; Python 3.14.4.
- `/opt/homebrew/bin/python3.11 --version` — exit 0; Python 3.11.14.
- `/opt/homebrew/bin/python3.11 -m pytest --version` — exit 0; pytest 8.3.4.
- task mypy venv `python -m mypy --version` — exit 0; mypy 2.3.0.
- `uvx --version` — exit 0; uvx 0.11.3.
- `uvx ruff --version` — exit 0; Ruff 0.16.0.
- `/opt/homebrew/bin/python3.11 -m build --version` — exit 0; build
  1.2.2.post1.

