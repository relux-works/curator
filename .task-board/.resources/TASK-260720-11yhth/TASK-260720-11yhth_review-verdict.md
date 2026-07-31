# TASK-260720-11yhth review verdict — accepted

Reviewer run `RUN-260730-1a975d` (not goal-bound). Read-only review; no product,
test, or board file was modified outside the board CLI.

## Reviewed delta

- Worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-11yhth/worktree`
- Branch: `task/TASK-260720-11yhth-command-runtime-activation`
- Base SHA: `11160f642d65a8daf3fbcca5401dca5ec80440f9`
- Files: `src/csk/shims.py` (rewritten), `src/csk/global_bins.py`,
  `src/csk/installer.py`, `tests/test_shims.py`, new `tests/test_build_activation.py`
  (1218 insertions / 103 deletions, no other file touched)

### Provenance re-verified by the reviewer

| Check | Result |
|---|---|
| `git verify-commit 11160f6` | good ECDSA signature for `oparin@me.com` |
| `git merge-base --is-ancestor main 11160f6` | exit 0 — base is a descendant of local `main` |
| dependency `TASK-260720-8nxlgx` (`protected-build-cache-windows`) | `done` |
| `git status --short` in the worktree | exactly the 4 modified + 1 new test file |

Drift note: `origin/main` has since advanced to `82d1cfc` with two additive
commits (`src/csk/builds/go_v1.py`, `tests/test_builds_go_v1*.py`, `+13` in
`cli.py`). The file sets are disjoint from this delta, so the rebase surface is
clean and the `go-v1` driver identity this task validates against is unchanged.

## Gates re-run by the reviewer (macOS arm64, worktree `.venv`, CPython 3.12.13)

| Command | Exit | Result |
|---|---:|---|
| `python -m pytest -q tests/test_shims.py tests/test_build_activation.py` | 0 | 100 passed, 6 skipped (Windows-execution tests) |
| `python -m pytest -q -p no:randomly` (full suite) | 0 | 950 passed, 86 skipped |
| `python -m mypy` | 0 | `Success: no issues found in 65 source files` |
| `uvx ruff check src/csk/shims.py tests/test_shims.py tests/test_build_activation.py` | 0 | `All checks passed!` |
| `python -m compileall -q src` | 0 | — |
| `git diff --check` | 0 | — |

The reviewer's full-suite count (950/86) is lower than the implementer's
(1116/38) because `CURATOR_CONFORMANCE_ROOT` was not exported in this run, so
the pinned conformance suite skips. Both runs are green; the delta is
environment, not behavior. CI gates are `python -m pytest -v` and
`python -m mypy` only — the repository configures no ruff gate, so the
pre-existing `global_bins.py` / `installer.py` ruff findings the implementer
disclosed do not gate anything.

Native Windows evidence was accepted as attached (`TASK-260720-11yhth_native-windows-runs.log`):
7 passed / 6 skipped on the explicit launcher-execution selection, 852 passed on
the full suite with only the unrelated PowerShell-execution-policy test red, and
zero mypy errors in `shims.py`, `global_bins.py`, `installer.py`. The two red
gates are reproduced on the unmodified base on the same host, so they are
pre-existing and correctly disclosed rather than masked.

## Acceptance-criteria verdict

| Criterion | Verdict |
|---|---|
| Reuse only when every required active script path is regular, contained, present | **met** — `runtime_state_is_complete` + `_contained_entry` reject a link at every path component below the runtime dir; covered by the reuse/replacement family incl. directory, dangling-link, linked-directory, non-directory-runtime and missing-root cases |
| Missing or wrong-type paths trigger staged replacement | **met** — `_publish_runtime_directory` stages into a sibling, moves the broken tree aside, renames in, discards the superseded tree, and yields to a concurrent writer that published complete state first |
| Unix project and global activation targets the immutable cache artifact | **met** — symlink (no PATH entries) or `exec`ing `/bin/sh` wrapper straight at the cache artifact; asserted byte-exact, and the artifact keeps its `0o500` cache mode |
| Unix activation forwards argv and exit status without profile setup | **met** — launcher body asserted in full (no profile read/write), and executed for real: argv preserved through quoting, exit 9 propagated |
| Windows `.cmd` quotes the executable, forwards `%*`, preserves ERRORLEVEL | **met** — `set "ERRORLEVEL="` inside `setlocal DisableDelayedExpansion` is the right fix for an inherited ERRORLEVEL *variable* shadowing the dynamic value; proven natively with a poisoned `ERRORLEVEL=0` and exit 9 |
| Rejects command-name or path injection | **met** — rejection, not escaping: portable-identifier command names, absolute targets free of `"` CR LF NUL, PATH entries free of the platform separator, literal `%` doubled; 16 + 6 injection cases |
| Mixed script and build commands share collision and stale-shim rules | **met at this layer** — `shim_path` is the single "one name owns one launcher path" rule for project, global, and user-bin publication, and stale removal is kind-agnostic. Closure-level export of build commands is explicitly downstream (see scope note) |
| Install-time tests prove no built artifact is launched | **met** — a runnable artifact writes a sentinel; absent after activation, present only after an explicit run; both POSIX and native Windows, project and global |
| Explicit post-install tests prove argv and nonzero exit codes propagate | **met** — exit 9 with and without PATH entries on both platforms |
| POSIX and Windows-focused pytest plus strict mypy pass | **met** — re-run above; strict mypy green on the authoritative configuration |

## Defects the implementer found and fixed in the touched surface

All four are real and correctly fixed: `Path.write_text` newline translation
producing `\r\r\n` in a `.cmd` on a Windows host; Windows stale-shim mapping
deleting the live launcher of a command whose declared name ends in `.cmd`;
`write_bin_shim` and `shutil.copy2` writing *through* a planted symlink at the
destination; and `.cmd` suffixing disagreeing between `shims` and `global_bins`
(`tool.cmd.cmd` vs `tool.cmd`). Unifying the suffix rule in `shims.shim_path`
and deleting the `global_bins` duplicate is the right call.

## Architecture fit

- No import cycle introduced: `shims` now depends on `builds.cache`,
  `builds.metadata`, `install_marker`, and `identifiers`; none of those import
  `shims`. Dependency direction stays downward.
- `derived_artifact_path` only special-cases `goos == "windows"`, so the
  activation cross-check against a `linux`/`darwin` receipt agrees with the
  `goos="unix"` form `install_marker` validates. No latent mismatch.
- `BuildCommandActivation` being constructible only through
  `select_build_activation` (which re-verifies marker vs receipt vs on-disk
  artifact) is a sound way to make "a launcher cannot be published from an
  unvalidated physical path" a type-level property.

## Scope boundary — checked, not assumed

`closure.active_commands` still exports script commands only, so no product
path calls `select_build_activation` yet. That is a documented boundary, not a
dead end: the wiring is owned by `TASK-260720-2x6mjn` (`development`),
`TASK-260720-3t8nr3` and `TASK-260720-g7kgox` (`backlog`), all of which exist on
the board under the same story. This task delivers the typed API and proves the
launcher, collision, and stale-shim rules are already kind-agnostic. Accepted.

## Non-blocking findings (for downstream tasks, not rework here)

1. **Relative `csk_home` now hard-fails.** `write_bin_shim` requires an absolute
   launcher target, so a relative `CSK_CONFIG` (⇒ `csk_home = Path(".")`) raises
   `ShimError: ... must be an absolute path`. Reproduced by the reviewer. The
   base accepted it, but only by resolving through the process cwd, so the old
   behavior was cwd-dependent rather than correct — failing closed is the better
   contract. Recommended follow-up (outside this task's scope): absolutize
   `config.config_path()` at load. Docs only ever show an absolute `CSK_CONFIG`.
2. **`call` on Windows is now unconditional**, including for compiled `.exe`
   artifacts that never need it (the base used a direct invocation when no PATH
   entries were present). `call` performs an extra percent-expansion pass, so an
   argument containing a literal `%VAR%` may be expanded before it reaches the
   target. Not verified here — the Windows validation host was unreachable — and
   `call` is the correct universal choice because a `.cmd` target would
   otherwise never return to `exit /b %ERRORLEVEL%`. Flagged as an argument-
   fidelity case worth an explicit assertion in
   `TASK-260720-3pemm6` (`cross-platform-go-build-e2e`).
3. **GC orphan-name hygiene.** The new staging names
   `.{name}.tmp-{pid}-{index}` and `.{commit}.stale-{pid}-{index}` do not match
   `gc._ORPHAN_RE` (`^\..+\.(tmp|backup)-(\d+)$`) because of the trailing index.
   No leak survives GC — a `.stale-` directory under `runtime/<skill>/` is
   removed by the unreferenced-commit sweep, and a staged command file lives
   inside the commit directory that is removed wholesale — but `sweep_orphans`
   no longer recognizes them by name. Cheap fix for
   `TASK-260720-th0jdi` (`build-currentness-repair-gc`).

None of the three is an AC violation or a regression in the tested surface.

## Verdict

**Accepted.** Implementation matches the AC, fits the project architecture,
tests are green on the reviewer's own re-run, and the Windows-side red gates are
pre-existing and honestly disclosed rather than masked.

Per the reviewer constraint, no `commit_ack` is supplied: this artifact is the
acceptance evidence for the commit-owning mover, which commits the delta and
performs the final `done` transition with `commit_ack=scope_committed`.
