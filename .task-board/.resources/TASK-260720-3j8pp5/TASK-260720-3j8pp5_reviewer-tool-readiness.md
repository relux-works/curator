# TASK-260720-3j8pp5 tool readiness

- Working directory: `/Users/iv/Developer/ReluxWorks/curator`
- Spawn run: `RUN-260729-688e43`
- `task-board`: `/Users/iv/.local/bin/task-board`, version 0.23.0
- `git`: `/usr/bin/git`, version 2.50.1
- `rg`: available, version 15.2.0
- `pytest`: `/opt/homebrew/bin/pytest`, version 8.3.4
- `python`: unavailable on PATH (`command not found`)
- `mypy`: unavailable on PATH (`command not found`)
- `python3`: `/opt/homebrew/bin/python3`, version 3.14.4
- `uv`: `/opt/homebrew/bin/uv`, version 0.11.3
- Project-local `.venv`: absent

The bare `python` and `mypy` commands failed readiness. Do not assume those
entry points work. Inspect the project configuration and use the repository's
established development interpreter before running Python or mypy validation.

The product repository's established shared development interpreter was then
verified successfully:

- `/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python`: Python 3.14.4
- `python -m pytest`: pytest 9.0.3
- `python -m mypy`: mypy 2.1.0 (compiled)

All review validation uses that interpreter against the task worktree.
