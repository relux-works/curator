# TASK-260720-2g21eg cycle-5 tool readiness

Checked from the preserved CocoaSkills task worktree at
`/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2g21eg/worktree`.

- `task-board --version`: exit 0; 0.23.0.
- `git --version`: exit 0; 2.50.1 (Apple Git-155).
- `rg --version`: exit 0; 15.2.0 with PCRE2/JIT.
- `.venv/bin/python -VV`: exit 0; CPython 3.14.4.
- `.venv/bin/pytest --version`: exit 0; pytest 9.1.1.
- `.venv/bin/python -m mypy --version`: exit 0; mypy 2.3.0.
- `.venv/bin/python -m build --version`: exit 0; build 1.5.0.
- `.venv/bin/python -m twine --version`: exit 0; twine 7.0.0.
- `go version`: exit 0; go1.25.5 darwin/arm64.
- `ssh -V`: exit 0; OpenSSH_10.2p1.
- Initial remote command spelling `ssh ... win "cmd /c ver"`: exit 1
  because the remote shell treated the quoted text as one invalid command.
- Corrected `ssh ... win cmd.exe /c ver`: exit 0; Microsoft Windows
  10.0.19045.6456.

The initial SSH invocation failure was command syntax only; the corrected
standalone probe confirms the native Windows host is reachable.
