# TASK-260720-2g21eg tool readiness — cycle 4

- `zsh --version`: exit 0, zsh 5.9.
- `git --version`: exit 0, git 2.50.1.
- `rg --version`: exit 0, ripgrep 15.2.0.
- `go version`: exit 0, Go 1.25.5 darwin/arm64.
- `.venv/bin/python --version`: exit 127; the prior task-local virtual
  environment is absent.
- `python3 --version`: exit 0, Python 3.14.4.
- `python3 -m pytest --version`: exit 1; pytest is not installed in the system
  interpreter.
- `python3 -m mypy --version`: exit 1; mypy is not installed in the system
  interpreter.
- `python3 -m build --version`: exit 1; the importable `build` name is not the
  packaging build frontend.

Remediation: create the repository-documented task-local `.venv`, install the
`dev` extra, and re-run readiness checks before using pytest, mypy, or build.

## Remediation result

- `python3 -m venv .venv`: exit 0.
- `.venv/bin/python -m pip install -e '.[dev]'`: exit 0.
- `.venv/bin/python --version`: exit 0, Python 3.14.4.
- `.venv/bin/python -m pytest --version`: exit 0, pytest 9.1.1.
- `.venv/bin/python -m mypy --version`: exit 0, mypy 2.3.0.
- `.venv/bin/python -m build --version`: exit 0, build 1.5.0.
- `.venv/bin/python -m twine --version`: exit 0, twine 7.0.0.
