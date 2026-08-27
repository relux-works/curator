# TASK-260720-3t8nr3 reviewer evidence — RUN-260730-bbd2a8

Repository: /Users/iv/Developer/intranet/cocoaskills
Worktree:   .temp/TASK-260720-3t8nr3/worktree
Branch:     task/TASK-260720-3t8nr3-transactional-project-hybrid
HEAD:       b3a5031ed551b27a298eef486a068b5175beaacc (== main == origin/main, re-fetched this cycle)
Working tree: 9 modified files + tests/test_installer_transactions.py (untracked), uncommitted.

## Gates re-run independently by the reviewer

Interpreter: /Users/iv/Developer/intranet/cocoaskills/.venv/bin/python (pytest pythonpath = ["src"] resolves to the worktree).

| Gate | Command | Result |
| --- | --- | --- |
| Strict typing | `python -m mypy` | Success: no issues found in 67 source files |
| Task vectors + gc + adapters | `python -m pytest -q tests/test_installer_transactions.py tests/test_gc.py tests/test_adapters.py` | 23 passed in 195.90s |
| Concurrency vector, 20 consecutive runs *under concurrent full-suite CPU load* | `python -m pytest -q tests/test_transactions.py::test_concurrent_project_transactions_preserve_both_consumers` | pass=20 fail=0 |
| Decisive full suite | `python -m pytest -q` | 1131 passed, 98 skipped in 1354.52s |

## Cross-device adapter-mode probe (new reviewer evidence for B3)

macOS RAM disk used as an alternate filesystem:
`hdiutil attach -nomount ram://262144` + `diskutil erasevolume HFS+ CSKRAM <dev>` -> /Volumes/CSKRAM

### Matrix A — fresh install, one variable (TMPDIR device)

| code under test | TMPDIR | temp st_dev | project st_dev | `.claude/skills/skill-a` |
| --- | --- | --- | --- | --- |
| worktree (this task) | default | 16777232 | 16777232 | symlink |
| worktree (this task) | /Volumes/CSKRAM | 16777236 | 16777232 | **plain directory copy** |
| pre-task main b3a5031 | /Volumes/CSKRAM | 16777236 | 16777232 | symlink |

### Matrix B — same project, same config, three consecutive installs, only $TMPDIR changes

```
install#1 default TMPDIR : status=ok errors=[] tempdev=16777232 .claude/skills/skill-a symlink=True
install#2 TMPDIR=ramdisk : status=ok errors=[] tempdev=16777236 .claude/skills/skill-a symlink=False
install#3 default TMPDIR : status=ok errors=[] tempdev=16777232 .claude/skills/skill-a symlink=True
```

Scripts: `.temp/review-bbd2a8/xdev_probe.py`, `.temp/review-bbd2a8/flip_probe.py` in the worktree.
