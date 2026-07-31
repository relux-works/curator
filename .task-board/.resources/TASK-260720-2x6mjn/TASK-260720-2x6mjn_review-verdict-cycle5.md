# TASK-260720-2x6mjn — reviewer verdict (cycle 5, Windows CI rework)

**Verdict: accepted**

Reviewer: reviewer role, RUN-260730-b310b4 (not goal-bound — `task-board spawn goal` reports
"run is not goal-bound").
Scope of this cycle: the Windows CI rework mandated by the attached instruction
`windows-ci-rework.md`, on top of the already-accepted cycle-4 content.
Reviewed tree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2x6mjn/worktree`,
branch `task/TASK-260720-2x6mjn-pure-build-planner`, HEAD `b3a5031ed551b27a298eef486a068b5175beaacc`,
clean working tree, `HEAD == @{u}`.
Accepted base for this cycle: `323ea47db821641bc68f2b9054b50e82642a6df2`.
Nothing was staged, committed, pushed, or modified by this review. No `commit_ack` supplied
(reviewer archetype). Review scratch lives under the gitignored `.temp/review5/`.

Prior verdicts: `..._review-verdict.md` (c1), `..._review-verdict-cycle2.md`,
`..._review-verdict-cycle3.md`, `..._review-verdict-cycle4.md` (accepted).

Provenance: board is `/Users/iv/Developer/ReluxWorks/curator/.task-board`; code is
`/Users/iv/Developer/Wildberries/cocoaskills`. Paths below are relative to the code repo.

## Rework diff

`323ea47..b3a5031` — 6 files, +152/−19. No production file other than the planner.

```
LOGBOOK.md                   | 18 +
src/csk/builds/planner.py    | 24 +-
tests/conftest.py            | 21 +
tests/test_build_planner.py  | 67 +
tests/test_global_install.py | 21 +-
tests/test_install.py        | 20 +-
```

`git diff 323ea47..b3a5031 --stat -- .github` is empty, and the diff adds no
`skipif` / `xfail` / `pytest.skip`. Both explicit prohibitions in the brief are honored.

## Instruction-by-instruction

| Requirement | Status | Evidence |
|---|---|---|
| Preserve accepted planner behavior and the stable error contract | met | Error codes and message strings untouched (`concurrent_state_change`, `generation_unreadable`). Cycle-2 probe archive re-run: byte-identical observable outputs (below). |
| Deterministic read-only generation probe on Windows | met | `_stable_open_stat` (`planner.py:634`) replaces the whole-tuple cross-API compare in `_hash_regular_file_noatime` (`planner.py:555`). |
| …without weakening concurrent-change detection | met | Compensated by a newly mandatory post-read same-API recheck for regular files: `_require_stable_lstat(path, before)` at `planner.py:497`, which the pre-rework code did **not** have in the `S_ISREG` branch. Pinned by a new regression test. |
| Portable no-Go real-install tests, minimum trusted executables preserved, Go still proven absent | met | `set_path_with_git_without_go` (`tests/conftest.py:98`) symlinks the **native** basename (`git.exe` on Windows) and self-asserts both `which("git") is not None` and `which("go") is None`. |
| Focused regression coverage for the Windows semantics | met | Two new tests in `tests/test_build_planner.py`; both independently verified red against pre-fix product code. |
| No workflow-matrix / skip / xfail changes | met | See diff checks above. |
| Focused tests + strict mypy locally | met | Independently reproduced below. |
| Push only to `origin`, never `wb` | met | `git ls-remote origin` → `b3a5031` on the task branch. `wb` (`gitlab.wildberries.ru`) does not resolve from this host, so it could not have been pushed. |
| Do not merge or land | met | PR #15 `state=OPEN`, `mergedAt=null`, `mergeable=MERGEABLE`, `mergeStateStatus=CLEAN`, 0 behind / 2 ahead of `origin/main` (`0be99ba`). |

## The fix is correct, and the detection is not weaker

`_stable_open_stat` drops `st_nlink`, `st_uid`, `st_gid`, permission bits, and `st_ctime_ns`
from the *path-lstat vs. fstat* comparison, keeping `os.path.samestat` (dev+ino), file type,
`st_size`, and `st_mtime_ns`. Every dropped field is still compared — later, and in the
same API — by the post-read `_require_stable_lstat`, which diffs the full `_stat_fields`
tuple (`planner.py:652`, includes `st_ctime_ns`) of `before` against a fresh `path.lstat()`.

Attack/race walk-through against the current code:

| Mutation inside the window | Caught by |
|---|---|
| File replaced (new inode) before `open` | `samestat` → `concurrent_state_change` |
| Truncate / append before `open` | `st_size`, `st_mtime_ns` |
| In-place content edit with `mtime` restored via `utime` | post-read `lstat`: `st_ctime_ns` bumped by the write and by the `utime`, not restorable unprivileged |
| `chmod` / `chown` / hardlink-count change | post-read `lstat`: `st_mode` / `st_uid` / `st_gid` / `st_nlink` + `st_ctime_ns` |
| Metadata change then restore | post-read `lstat`: `st_ctime_ns` bumped twice |
| Replaced *after* the read with identical bytes and restored `mtime` | post-read `lstat`: new inode → **newly detected**, was not detected before this rework |
| Content changes during the read | unchanged: `_stable_stat(opened, after)` fd-to-fd compare (`planner.py:580`) |

The last row is a strict improvement, not a regression, and it is exactly what the second
new test pins.

## Independently reproduced evidence

Canonical interpreter `/Users/iv/Developer/Wildberries/cocoaskills/.venv/bin/python` (3.14.4),
rootdir = the task worktree.

| Gate | Result |
|---|---|
| `python -m pytest -q` (full suite) | exit 0 — **1120 passed, 98 skipped** in 86.69s |
| `python -m mypy` | exit 0 — **Success: no issues found in 67 source files** |
| Go-less focused suite (`test_build_planner`, `test_install`, `test_global_install`, `test_cli`) under `PATH=nogo-bin` | exit 0 — **151 passed** in 37.95s |
| Cycle-2 probe archive, extracted fresh, run unmodified | exit 1 — **1 failed, 5 passed** (same known non-finding as cycles 2–4) |
| CI run 30554363746 on `b3a5031` | **14/14 success**, incl. Windows Python 3.11/3.12/3.13/3.14, Linux+macOS matrix, strict mypy, artifact build |

`nogo-bin` is the 3081-entry symlink farm of the full `PATH` minus `go`/`gofmt`
(`PATH=$NB command -v go gofmt` exits 1). Counts match the implementer's report exactly
(1120/98, 151), so the reported numbers are reproducible, not asserted.

### Test-first claim verified, and it is stronger than reported

`results.md` claimed one expected-red test. I reconstructed the pre-fix tree
(`git archive 323ea47 | tar -x`) and overlaid **only** the post-fix
`tests/test_build_planner.py`:

```
FAILED test_filesystem_generation_probe_accepts_windows_path_fd_ctime_split
FAILED test_filesystem_generation_probe_detects_replacement_after_file_read
2 failed, 5 passed
```

Both new tests are non-vacuous. The second one specifically fails pre-fix with
`DID NOT RAISE BuildPlanningError`, proving the newly added `_require_stable_lstat` in the
`S_ISREG` branch is load-bearing and not decorative.

### The fix is causally tied to the reported Windows failures

I did not take green CI at face value. I wrote an out-of-tree pytest plugin
(`.temp/review5/winsim.py`) that shims **only** `planner.os.fstat` to return a descriptor
stat whose `st_ctime_ns` is offset by 1 — the exact divergence diagnosed — leaving every
other stdlib consumer on real `os.stat_result` objects.

| Tree | Result under the shim |
|---|---|
| pre-fix (`323ea47`) | **5 failed** — `test_filesystem_generation_probe_is_deterministic_and_read_only` and the four `test_global_install.py` dry-run tests, all with `concurrent_state_change: shared planning file changed while opening: …/Skillfile.json` |
| post-fix (`b3a5031`) | **5 passed** |

That reproduces the reported Windows failure mode on macOS and confirms the fix addresses
the cause, not a symptom.

### Accepted behavior preserved end-to-end

Cycle-2 probes, re-run against `b3a5031`, are byte-identical to the cycle-3/cycle-4 accepted
table:

| probe | accepted (c3/c4) | cycle 5 |
|---|---|---|
| `test_probe_global_transitive_leak` | `['consumer']`, `global/bin=[]` | same |
| `test_probe_global_inactive_shim_collision` | `['consumer-one','consumer-two']`, `global/bin=[]` | same |
| `test_probe_global_requires_go` | `STATUS=ok`, `['skill-build','skill-plain']` | same |
| `test_probe_real_install_requires_go` | pass | pass |

The single probe failure remains `test_probe_global_install_with_skill_missing_skill_md`,
documented as a pre-existing non-finding in cycle 2 and re-confirmed against detached base
`82d1cfc` in cycle 3.

## Non-blocking findings — new this cycle

None of these block acceptance.

11. **Redundant `S_IFMT` compare.** `_stable_open_stat` (`planner.py:646`) re-checks the file
    type that the caller already established via the `S_ISREG(before.st_mode)` branch and the
    adjacent `stat.S_ISREG(opened.st_mode)` guard at `planner.py:556`. Harmless; free to drop.
12. **Inconsistent recheck placement.** In the `S_ISREG` branch `_require_stable_lstat` runs
    *before* `_generation_record` (`planner.py:497`); in the link/dir/other branches it runs
    *after*. Semantically irrelevant (the digest is process-local) but reads as an oversight.
13. **`samestat` degrades where `st_ino == 0`.** Some Windows network filesystems report
    `st_ino == 0`, which makes `os.path.samestat` trivially true and leaves the cross-API
    identity check resting on size + `mtime_ns`. The pre-rework code was no better (it also
    compared `ino == 0` against `ino == 0`), and the post-read full `lstat` still applies, so
    this is a platform limit rather than a regression — but it is now the sole identity gate
    in that window and deserves to be known.
14. **Probe cost compounds.** The added `lstat` makes it 3 `lstat`/`fstat` calls + 1 `open`
    per regular file per capture, twice per planning attempt. This sharpens carried-forward
    item 9 below; worth a cheap-stat fast path before planning lands on real installs in
    TASK-260720-3t8nr3.
15. **`symlink_to` on Windows needs Developer Mode or admin.** `set_path_with_git_without_go`
    (`tests/conftest.py:108`) is green on GitHub-hosted runners, but a Windows contributor
    without Developer Mode gets a raw `OSError` instead of the helper's clean assertion.
    `shutil.copy2` is not a drop-in (`git.exe` needs sibling DLLs); `os.link` or a `.cmd`
    shim would be.
16. **The isolated PATH narrowed silently.** The old fixture symlinked `git`, `sh`, `env`,
    `uname`; the helper now exposes `git` only. Everything is green on all three OSes so the
    other three are demonstrably unused by these paths, but the brief's "minimum trusted
    executable set" is now an implicit claim with no test pinning it.
17. **Instruction-file detail drift (informational, no action).** The attached brief says all
    four Windows jobs failed the same seven tests. The actual logs of run 30551952750 show
    Python 3.11 failing **only** the two PATH-fixture tests, while the `concurrent_state_change`
    set starts at 3.12 — which independently corroborates the "Python 3.12+" attribution in
    the code comment and LOGBOOK. The 3.12 ctime set also included
    `test_cli.py::test_cli_install_dot_dry_run_does_not_save_config` and
    `test_install.py::test_audit_dry_run_does_not_write_verdict_cache` rather than two of the
    `test_global_install.py` tests named in the brief. All are the same root cause and all are
    green on `b3a5031`.

## Non-blocking findings carried forward from cycle 4

Unchanged and still open: (1) duplicated activation logic in `installer._active_build_command_names`;
(2) tautological `observed_argv` assertion; (3) dry-run vs install trust divergence over
`migrate_snapshot_states`; (4) late-binding `provider` closure in `planner._plan_once`;
(5) read-only registry state stricter than the install + `check_snapshots` `cache_dir`
misnomer; (6) `result.builds` has no consumer on the real path (owner: TASK-260720-3t8nr3);
(7) global dry-run stricter than global install on MCP/registry gates; (8) `csk update --dry-run`
now validates the skills root, untested either way; (9) generation-probe cost; (10) Go is an
undeclared test dependency in CI.

## Definition-of-Done status

| DoD item | Status |
|---|---|
| Base SHA recorded, worktree from fast-forwarded main + dependency handoff | met — base `323ea47`, rework `b3a5031`, branch 0 behind `origin/main` (`0be99ba`) |
| Before/after purity tests for forbidden dry-run paths and audit-before-build ordering | met — unchanged from cycle 4, re-verified green |
| Focused pytest + strict mypy with task-scoped evidence | met — independently reproduced |
| Code per description and AC | met |
| Tests for new/changed behavior and passing | met — 2 new regression tests, both proven red pre-fix |
| Lint clean | met — strict mypy clean; `git diff --check` and `git diff --cached --check` exit 0 |
| Build/validation run, build not broken | met — full suite green, CI artifact-build job green |
| Outcome artifact attached with task-scoped name | met — `TASK-260720-2x6mjn_windows-ci-rework-results.md`; this verdict adds cycle-5 evidence |
| Important findings recorded in logbook | met — `LOGBOOK.md:31-47` records both Windows semantics and the compensating recheck |
| Implementation matches AC | met |
| Solution fits project architecture | met — planner stays read-only and pure; the added recheck is the same primitive already used by the link/dir branches |
| Tests green | met — green on macOS with and without Go, and on all 14 CI jobs |

## Reproducing

```bash
REPO=/Users/iv/Developer/Wildberries/cocoaskills
VENV=$REPO/.venv/bin/python
WT=$REPO/.temp/TASK-260720-2x6mjn/worktree

# gates
(cd $WT && $VENV -m pytest -q && $VENV -m mypy)

# Go-less focused suite
(cd $WT && PATH=$REPO/.temp/TASK-260720-2x6mjn/nogo-bin $VENV -m pytest -q \
   tests/test_build_planner.py tests/test_install.py tests/test_global_install.py tests/test_cli.py)

# test-first check: new tests against pre-fix product code
(cd $WT && mkdir -p .temp/review5/prefix && git archive 323ea47 | tar -x -C .temp/review5/prefix \
   && git show b3a5031:tests/test_build_planner.py > .temp/review5/prefix/tests/test_build_planner.py)
(cd $WT/.temp/review5/prefix && $VENV -m pytest -q -p no:cacheprovider tests/test_build_planner.py)   # 2 failed

# Windows ctime-split simulation (winsim.py is in the evidence archive)
(cd $WT && PYTHONPATH=.temp/review5 $VENV -m pytest -q -p no:cacheprovider -p winsim \
   tests/test_build_planner.py::test_filesystem_generation_probe_is_deterministic_and_read_only \
   tests/test_global_install.py::test_global_install_dry_run_does_not_write_anywhere)               # passes
(cd $WT/.temp/review5/prefix && PYTHONPATH=.. $VENV -m pytest -q -p no:cacheprovider -p winsim \
   <same node ids>)                                                                                # 5 failed
```

## Handoff

The rework is accepted on content and the branch is already committed and pushed to `origin`
(`b3a5031`, PR #15 open and mergeable, CI fully green). This review is read-only and supplied
no `commit_ack`; the commit-owning mover records it on the final `done` transition.

Reviewer scratch to delete when the task closes: `.temp/TASK-260720-2x6mjn/review5/`, plus the
earlier `probe/`, `probe2/`, `review3/`, `review4/`, `nogo-bin/`, `review2-*.log`, and the
baseline worktree (`git worktree remove .temp/TASK-260720-2x6mjn/review-baseline`).
