# TASK-260720-3t8nr3 review verdict rev 2 — CHANGES REQUESTED (to-dev)

Reviewer run: `RUN-260730-bbd2a8`. Read-only review; no product code was modified.
Supersedes `TASK-260720-3t8nr3_review-verdict.md` (run `RUN-260730-c38a81`), which is
retained for the audit trail.

## Reviewed artifact

- Repository: `/Users/iv/Developer/Wildberries/cocoaskills`
- Worktree: `.temp/TASK-260720-3t8nr3/worktree`
- Branch: `task/TASK-260720-3t8nr3-transactional-project-hybrid`
- Worktree `HEAD` = `b3a5031ed551b27a298eef486a068b5175beaacc`. Re-fetched this cycle:
  `main` == `origin/main` == recorded base SHA.
- Uncommitted: 9 modified files (2097 insertions / 256 deletions) plus the untracked
  `tests/test_installer_transactions.py`.

## Gates independently re-run by the reviewer

Interpreter `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python`; pytest
resolves `pythonpath = ["src"]` from the worktree.

| Gate | Command | Result |
| --- | --- | --- |
| Decisive full suite | `python -m pytest -q` | **1131 passed, 98 skipped in 1354.52s** (implementer claimed 1131/98 in 1328.73s — exact count match) |
| Task vectors + gc + adapters | `python -m pytest -q tests/test_installer_transactions.py tests/test_gc.py tests/test_adapters.py` | 23 passed in 195.90s |
| Concurrency vector, 20 consecutive runs **under concurrent full-suite CPU load** | `python -m pytest -q tests/test_transactions.py::test_concurrent_project_transactions_preserve_both_consumers` | pass=20 fail=0 |
| Strict typing | `python -m mypy` | Success: no issues found in 67 source files |

Every reported figure is reproducible. The implementer's evidence is honest.

## Prior-cycle findings — verified closure

**B1 (concurrency assessment, attached instruction) — CLOSED.**
`tests/test_transactions.py` now replaces `threading.Barrier(2, timeout=3)` + fixed
`join(timeout=5)` with explicit handoff events: both project locks are held before the
parent releases `start_first_home`; A takes the manager-home lock and signals
`first_home_acquired`; B waits for that, signals `second_home_attempt`, then contends;
A commits only after `second_home_attempt`. The intended serialization is now the thing
under test rather than a timing coincidence. Waits are bounded (10s POSIX / 30s Windows,
completion deadline `2×coordination + 5s` off `time.monotonic()`), worker exceptions are
collected under a lock, and the terminal
`assert all(not thread.is_alive() for thread in threads)` liveness assertion is retained
verbatim. Not skipped, not xfailed. The assessment and the Windows 3.14 evidence are
written into both `LOGBOOK.md` and the results artifact. The Windows branch remains
CI-only, and the artifact says so.

**B2 (lost snapshot GC / orphan sweeps) — CLOSED.**
The `gc` import and a terminal `gc.collect_runtime` are restored on the project install
path (`src/csk/installer.py:130`–`:137`), under `ManagerHomeLock`,
gated on non-dry-run + at least one `ok` result + no failed result. I confirmed the gate
actually fires on repeat installs: `ProjectResult.status` is `"ok"` for every successful
project (`src/csk/installer.py:271`) — `"up-to-date"` at `:2174`/`:2241` is a per-skill
status, not a project status — so the accumulate-forever scenario is genuinely closed.
Two real regression vectors:
`tests/test_gc.py::test_successful_install_runs_snapshot_and_remaining_orphan_gc`
(seeds an unreferenced snapshot plus dead-pid orphans under `global/skills` and
`runtime/<skill>`, asserts all three collected) and
`tests/test_install.py::test_post_commit_gc_runs_only_after_successful_real_install`
(asserts GC runs exactly once — not on dry-run, not on failure — and that the home lock
is held while it runs). No deadlock: `gc.collect_runtime` and `consumers.replace_consumers`
take no locks, and neither `_cmd_update` nor `global_install.install` re-enters
`installer.install`.

**N1 (module-wide non-POSIX skip) — CLOSED.** Module-level `pytestmark` removed;
`POSIX_BUILD_VECTOR` applied to 5 of 7 vectors. `test_commit_generation_change_restarts_complete_project_plan`
and `test_second_project_consumer_failure_rolls_back_without_touching_first` now run on Windows.

**N3 (unprotected `recover()`) — CLOSED.** The installer now wraps the
pre-loop `engine.recover(home_lock)` (`src/csk/installer.py:170`) in `except transactions.TransactionError` (`:171`);
`TransactionCorruptionError` subclasses it, so a corrupt journal becomes a failed
`ProjectResult` instead of an uncaught CLI traceback.

**N2 (whole-live-world staging copy) — RECORDED.** Accepted as written, but see B3 and
N6: staging *location* is now load-bearing, so the two are one fix.

**N4 (live-tree probe) — the live-tree write is gone, but the replacement introduced B3.**

## Blocking finding

### B3 — default `adapter_mode: auto` is now decided by `$TMPDIR`, not by the destination

`adapters._transaction_links_supported` (`src/csk/adapters.py:326`–`:338`) replaced the
live-tree probe with:

```python
same_device = stage_root.stat().st_dev == live_directory.stat().st_dev
...
return same_device and _link_probe(stage_root)
```

`stage_root` is `staging_root / "adapters"`, and `staging_root` is
`tempfile.TemporaryDirectory(prefix="csk-materialization-plan-")` in
`_commit_materialization` (`src/csk/installer.py:1225`) with no `dir=` — i.e. the
**system temp directory**. So whenever `$TMPDIR` is on a different filesystem from the
project, `same_device` is false and default `auto` mode (`config.py:263`) silently
selects `copy` for every managed adapter mirror.

This is not an edge case. It is the default configuration on Linux hosts with a tmpfs
`/tmp` (systemd default; Fedora, Arch, Ubuntu 24.10+) and for any project on an external
or network volume on macOS. CI would not catch it: GitHub Actions runners keep `/tmp` on
the same root filesystem.

Two problems, in order of severity:

1. **The probe now tests the wrong filesystem.** "Can a symlink be created at
   `<project>/.claude/skills/`" is a property of the *destination*. Probing the staging
   filesystem is only ever correct by coincidence — when it happens to be the same
   device. The `same_device` guard is an acknowledgement of that, but the conclusion it
   draws (fall back to copy) discards a capability the destination actually has.
2. **Materialization output is no longer deterministic.** It is now a function of an
   environment variable that has nothing to do with the project. Two installs of the same
   project from two shells commit different content, and each flip rewrites every adapter
   mirror target. That is in direct tension with the property this task exists to
   establish — stable preimages and deterministic target derivation.

Measured, with a macOS RAM disk as the alternate filesystem
(`hdiutil attach -nomount ram://262144` + `diskutil erasevolume HFS+ CSKRAM <dev>`):

Matrix A — fresh install, one variable:

| code under test | TMPDIR | temp `st_dev` | project `st_dev` | `.claude/skills/skill-a` |
| --- | --- | --- | --- | --- |
| worktree (this task) | default | 16777232 | 16777232 | symlink |
| worktree (this task) | `/Volumes/CSKRAM` | 16777236 | 16777232 | **plain directory copy** |
| pre-task main `b3a5031` | `/Volumes/CSKRAM` | 16777236 | 16777232 | symlink |

Matrix B — same project, same config, three consecutive installs, only `$TMPDIR` changes:

```
install#1 default TMPDIR : status=ok errors=[] tempdev=16777232 .claude/skills/skill-a symlink=True
install#2 TMPDIR=ramdisk : status=ok errors=[] tempdev=16777236 .claude/skills/skill-a symlink=False
install#3 default TMPDIR : status=ok errors=[] tempdev=16777232 .claude/skills/skill-a symlink=True
```

Pre-task `main` had no such dependency: auto mode attempted the real symlink at the real
target and fell back to copy only on `OSError`
(`git show HEAD:src/csk/adapters.py`, lines 158-166).

No test covers the cross-device branch. The new vector
`tests/test_adapters.py::test_transaction_auto_mode_probes_only_private_staging` places
`stage_root` and `live_root` under the same `tmp_path`, so it can only ever exercise the
same-device path, and it monkeypatches `_link_probe` away entirely.

`LOGBOOK.md` does record "accepts that witness only when staging and the nearest live
parent are on the same filesystem; otherwise auto mode chooses a copy" — but it does not
state that staging lives in the system temp dir, which is what turns a stated edge-case
fallback into the common case on Linux.

Recoverable rework, not a design conflict. The obvious shape: anchor the operation-private
staging root on the destination filesystem (e.g. under the manager home) instead of the
system temp dir. That keeps N4's requirement intact — nothing is written into the user's
project — restores a destination-correct capability witness, removes cross-device copying
of the staged world, and directly relieves N2/N6. Whichever shape is chosen, the
cross-device branch needs a vector, and the decision needs recording.

Reproduction scripts left in the worktree: `.temp/review-bbd2a8/xdev_probe.py`,
`.temp/review-bbd2a8/flip_probe.py`.

## Non-blocking findings

### N5 — a successful install can now exit non-zero because post-commit GC could not take the home lock

`install()` acquires `ManagerHomeLock` for GC *after* every project transaction has
committed. `ManagerHomeLock` acquisition raises `LockError` on timeout
(`_timeout_from_env()` default), and `cli.main` maps `LockError` to `EXIT_LOCK`
(`src/csk/cli.py:85`-`:87`). Before this task the CLI took `GlobalLock` around the whole
install, so contention failed cleanly *before* any work. Now a fully successful,
fully committed install can report failure because a maintenance step could not get a
lock. Low probability, but the fix is small: treat post-commit GC lock contention as a
skipped-maintenance message rather than an operation failure.

### N6 — staging the whole live world in the system temp dir means tmpfs/RAM on the platforms from B3

This sharpens the recorded N2 tradeoff rather than reopening it. `_stage_materialization`
copies the entire project `.agents`, `~/.csk/hybrid` and `~/.csk/runtime` trees into
`tempfile.TemporaryDirectory(...)`. On the same hosts B3 affects (tmpfs `/tmp`), that
copy lands in RAM, and its size is O(total installed state on the machine), not
O(changed). The recorded tradeoff should say where the copy lands, not only how big it
is. Moving staging off the system temp dir (see B3) resolves both.

### N7 — consumer ledger paths are now `resolve()`d; benign but unrecorded

`consumers.encode_consumers` (`src/csk/consumers.py:48`-`:56`) applies
`Path(path).resolve()`, and `_write` now routes through it. Pre-task `_write` stored
`str(path)` unresolved. Making the ledger bytes canonical is the right call — the
transaction target digest depends on it — but it silently rewrites existing ledger
entries on the first install after upgrade for anyone whose project path traverses a
symlink. Worth one line in the artifact.

## Verdict

**changes requested → `to-dev`.**

The rework is good work. B1 and B2 are genuinely closed, not papered over: the
concurrency vector now tests the serialization it claims to test and keeps its liveness
assertion; snapshot GC and the remaining orphan sweeps are back with two vectors that
would actually catch their removal. N1 and N3 are closed. Every claimed gate reproduces
exactly, including the full suite count.

One item blocks acceptance:

1. **B3** — the N4 fix moved the auto-mode capability probe onto the staging filesystem,
   which is the system temp dir. Default `adapter_mode: auto` now silently becomes `copy`
   whenever `$TMPDIR` is on another filesystem, and committed materialization varies with
   an unrelated environment variable. Fix the probe's anchor (or the staging location),
   add a cross-device vector, and record the decision.

N5–N7 should be addressed or explicitly recorded in the same cycle; on their own they
would not have blocked acceptance.
