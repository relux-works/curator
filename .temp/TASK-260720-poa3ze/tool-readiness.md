# Tool readiness — TASK-260720-poa3ze

Verified 2026-07-20 before repository research.

- `task-board`: `/Users/iv/.local/bin/task-board`, version `0.20.1-15-g5e2c927`
- `git`: `/usr/bin/git`, version `2.50.1 (Apple Git-155)`
- `rg`: Codex-vendored binary, version `15.1.0`, PCRE2 available
- `go`: `/opt/homebrew/bin/go`, version `go1.25.5 darwin/arm64`; `GOROOT=/Users/iv/.goenv/versions/1.25.5`
- `python3`: `/opt/homebrew/bin/python3`, version `3.14.4`
- `jq`: `/opt/homebrew/bin/jq`, version `1.8.1`
- `file`: `/usr/bin/file`, version `5.41`
- `otool`: `/usr/bin/otool`, verified by listing `/bin/ls` dependencies
- `curl`: `/usr/bin/curl`, version `8.7.1`; used only for read-only primary-source link checks
- Python test environment: `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python`, with `pytest 9.0.3` and `jsonschema 4.25.1`

## Corrected probe

The first `go list -f` smoke attempted an unavailable template helper (`json`) and exited before compilation. The rerun uses direct list fields; no workflow depends on the failed form. The first spec/cocoaskills regression commands used the system Python, which lacked `jsonschema`/`pytest` and stopped before collection; both were rerun through the existing project virtual environment without installing or changing dependencies.

All required tools executed successfully and produced expected version output.
