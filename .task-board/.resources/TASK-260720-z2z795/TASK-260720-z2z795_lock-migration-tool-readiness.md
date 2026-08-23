# TASK-260720-z2z795 lock-migration tool readiness

Checked before technical changes on 2026-07-30:

- `git --version` — exit 0; `git version 2.50.1 (Apple Git-155)`.
- `rg --version` — exit 0; `ripgrep 15.2.0` with PCRE2.
- `/opt/homebrew/bin/python3.11 -m pytest --version` — exit 0;
  `pytest 8.3.4`.
- Task mypy venv `python -m mypy --version` — exit 0;
  `mypy 2.3.0 (compiled: yes)`.
- `uvx ruff --version` — exit 0; `ruff 0.16.0`.
- `/opt/homebrew/bin/python3.11 -m build --version` — exit 0;
  `build 1.2.2.post1`.

The intended project platform is Python 3.11+ on macOS, Linux, and Windows.
This run is on macOS; platform-neutral and platform-routing regressions cover
the cross-platform migration behavior locally.
