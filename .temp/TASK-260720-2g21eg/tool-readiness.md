# Tool readiness

Checked 2026-07-30 before project work.

| Tool command | Exit | Output |
| --- | ---: | --- |
| `git --version` | 0 | `git version 2.50.1 (Apple Git-155)` |
| `rg --version` | 0 | `ripgrep 15.2.0 (rev e89fff89ac)`; PCRE2 and NEON available |
| `task-board --version` | 0 | `task-board version 0.23.0 (commit beec6e9, built 2026-07-27T23:36:53Z)` |
| `python --version` | 127 | `zsh:1: command not found: python` |
| `python -m pytest --version` | 127 | `zsh:1: command not found: python` |
| `python -m mypy --version` | 127 | `zsh:1: command not found: python` |
| `go version` | 0 | `go version go1.25.5 darwin/arm64` |

The unqualified `python` executable is unavailable. Project inspection must identify
the supported Python runner before any Python validation is attempted.

## Project runner resolution

The CocoaSkills task worktree contains its own executable `.venv/bin/python`.

| Tool command | Exit | Output |
| --- | ---: | --- |
| `python3 --version` | 0 | `Python 3.14.4` |
| `uv --version` | 0 | `uv 0.11.3 (45da18ac3 2026-04-01 aarch64-apple-darwin)` |
| `.venv/bin/python --version` | 0 | `Python 3.14.4` |
| `.venv/bin/python -m pytest --version` | 0 | `pytest 9.1.1` |
| `.venv/bin/python -m mypy --version` | 0 | `mypy 2.3.0 (compiled: yes)` |

All authoritative Python validation will use the task-worktree interpreter
explicitly as `.venv/bin/python`.

## Native-host tooling

| Tool command | Exit | Output |
| --- | ---: | --- |
| `ssh -V` | 0 | `OpenSSH_10.2p1, LibreSSL 3.3.6` |
| initial `ssh win ...` readiness probe | 255 | `connect to host 100.120.84.42 port 22: Operation timed out` |

The first Windows-host probe timed out before authentication or command
execution. It will be retried at later safe checkpoints; no Windows validation
is treated as having run yet.

Two later SSH probes also exited 255 before authentication because
`100.120.84.42` had no Tailscale route. The macOS VPN service remained in
`Connecting` with zero successful connections. Restarting the existing
Tailscale app and its configured VPN service did not restore the network
extension. The app's own failure detail reported:

```text
System Extension Setup Failed
Error: OSSystemExtensionErrorExtensionNotFound
```

The app also requires a fresh Tailscale login to add the VPN tunnel and advises
a Mac reboot for the missing extension. A reboot/login is external operator
authorization and has not been performed. Native Windows validation remains
unrun.
