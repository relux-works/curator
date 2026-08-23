# TASK-260720-11yhth implementation evidence

Harden runtime reuse and activate built commands.

## Provenance

- Clean local main clone: `/Users/iv/Developer/Wildberries/cocoaskills`,
  fast-forward checked: `main == origin/main == 15860e3f309888845b9271a257fb95f7c2825b56`
  (`git rev-list --left-right --count main...origin/main` → `0 0`).
- Accepted dependency `TASK-260720-8nxlgx` (status `done`) handoff commit
  `11160f642d65a8daf3fbcca5401dca5ec80440f9`:
  - `git verify-commit 11160f6` → exit `0`, good signature for `oparin@me.com`;
  - `git merge-base --is-ancestor main 11160f6` → exit `0` (direct descendant of main).
- **Recorded base SHA: `11160f642d65a8daf3fbcca5401dca5ec80440f9`.**
- Task worktree: `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-11yhth/worktree`
- Branch: `task/TASK-260720-11yhth-command-runtime-activation`
- Nothing was staged or committed. No `uv.lock` was created.

## Changed files

| File | Change |
|---|---|
| `src/csk/shims.py` | rewritten: validated runtime reuse, typed compiled activation, hardened launchers |
| `src/csk/global_bins.py` | narrow: reuse `shims.shim_path` / `shims.global_bin_dir` instead of a duplicate local path rule |
| `src/csk/installer.py` | narrow call-site wiring: pass the active script commands to `install_runtime_roots` |
| `tests/test_shims.py` | extended: runtime reuse, launcher hardening, injection, stale-shim rules, execution |
| `tests/test_build_activation.py` | new: compiled activation selection matrix, no-launch, argv/exit propagation |

## Implemented behavior

### Runtime reuse (existing script runtime state)

- `install_runtime_roots` now takes `required_commands` and reuses an existing
  commit-keyed directory only when `runtime_state_is_complete` holds: the runtime
  directory is a real directory, every declared runtime root resolves to a real
  directory, and every required active command path resolves to a **regular**
  file. Containment is proven by rejecting a link at every path component, so no
  entry below the runtime directory can name a file outside it.
- Incomplete or wrong-type state triggers staged replacement: the new tree is
  built in a private sibling, the broken tree is moved aside, the staged tree is
  renamed in, and the superseded tree is discarded. A concurrent writer that
  published complete state first wins and the staged duplicate is dropped.
- `runtime_root_command_path` resolves through the same no-follow walk and adds
  execute bits through an `O_NOFOLLOW` descriptor instead of a path `chmod`.
- `install_runtime_command` stages and `os.replace`s, so a stale symlink at the
  destination can no longer redirect the copy outside the runtime tree.
- `runtime_directory` rejects a skill name or commit that is not a portable
  identifier, so neither can inject a path component.

### Compiled command activation

- New frozen `BuildCommandActivation` binds a command name to a verified
  immutable cache artifact plus its cache key, receipt hash, artifact hash, and
  size. `activate_build_command` refuses anything else, so a launcher cannot be
  published from an unvalidated physical path.
- `select_build_activation` is the only constructor path. It requires a `build`
  command with driver `go-v1`, a marker record with driver `go-v1`, and a
  `CacheEntryStatus.HIT` inspection carrying receipt, receipt hash, and artifact
  path, then cross-checks: receipt command identity, receipt target `goos`
  against the activation platform, `derived_artifact_path` against both the
  receipt and the marker, marker cache key against the receipt cache key, marker
  receipt hash against the inspected receipt hash, and marker artifact hash
  against the receipt artifact hash.
- The physical artifact must be an absolute path ending in the derived relative
  path, below the manager home, **not** below `<home>/runtime` (compiled outputs
  never use the script runtime identity), a regular file by `lstat`, of exactly
  the receipt's size, and owner-executable on Unix. Activation verifies rather
  than relaxes permissions because the published entry is immutable.
- Nothing on this path executes the artifact.

### Launchers

- One writer serves both command kinds. Unix: a relative symlink to the target,
  or a `#!/bin/sh` wrapper that prepends PATH entries and `exec`s the target with
  `"$@"`. Windows: `@echo off` / `setlocal DisableDelayedExpansion` /
  `set "ERRORLEVEL="` / optional `set "PATH=..."` / `call "<target>" %*` /
  `exit /b %ERRORLEVEL%`. No shell profile is read or written.
- `set "ERRORLEVEL="` is required: an inherited ERRORLEVEL *variable* shadows
  cmd's dynamic exit status and would mask the real exit code.
- Injection is rejected, not escaped away: a command name must be a portable
  identifier; a launcher target must be absolute and free of `"`, CR, LF, and
  NUL; a PATH entry must be absolute and must not contain the platform's PATH
  separator. Literal `%` in a Windows target or PATH entry is doubled.
- Launchers are written with explicit newline handling, and an existing file,
  symlink, or dangling symlink at the launcher path is removed first, so a write
  can never pass through a planted link. A directory at the launcher path fails
  closed.
- `shim_path` is now the single rule for "one command name owns one launcher
  path", shared by project shims, global shims, and user-bin publication.

### Mixed script and build commands

- Both kinds resolve to the same launcher path for a given name, so a build
  command replaces a same-named script launcher instead of creating a second
  entry, and stale removal keeps exactly the expected names of either kind.
- Windows stale removal now maps a `<name>.cmd` file to command `<name>` **or**
  to a command literally named `<name>.cmd`; previously a command whose declared
  name ended in `.cmd` had its own live launcher deleted as stale.
- Global user-bin publication and the unmanaged-conflict rule are name-based and
  were verified to treat a compiled command exactly like a script command.

## Defects found and fixed in the touched surface

1. `Path.write_text` newline translation turned the intended `\r\n` in a `.cmd`
   launcher into `\r\r\n` on a Windows host. Launchers are now written with
   explicit newline handling and asserted byte-exact.
2. Windows stale-shim mapping deleted the live launcher of a command whose name
   ends in `.cmd`.
3. `write_bin_shim` on Windows and `shutil.copy2` in `install_runtime_command`
   wrote *through* an existing symlink at the destination.
4. `write_bin_shim` appended `.cmd` unconditionally while `global_bins` did not,
   so a command named `tool.cmd` produced `tool.cmd.cmd` in one place and
   `tool.cmd` in the other.

## Validation gates

Every command was run directly as a standalone process; exit codes are the real
codes reported by the shell.

### Local (macOS `darwin` arm64, CPython 3.12.13, worktree `.venv`)

Conformance suite pinned to the CI ref `f5d7673039226ab81de2f4f87e2155ae995c4df3`
of `relux-works/curator-spec` and exported through `CURATOR_CONFORMANCE_ROOT`.

| # | Command | Exit | Result |
|---|---|---|---|
| 1 | `python -m pytest -q` (base `11160f6`, same env) | `0` | 1020 passed, 32 skipped (baseline) |
| 2 | `python -m pytest -q` | `0` | **1116 passed, 38 skipped** |
| 3 | `python -m pytest -q -rs` over shims, build activation, install, global install, closure install, hybrid scope, activation modes, e2e | `0` | 205 passed, 6 skipped |
| 4 | `python -m mypy` | `0` | `Success: no issues found in 65 source files` |
| 5 | `uvx ruff check src/csk/shims.py tests/test_shims.py tests/test_build_activation.py` | `0` | `All checks passed!` |
| 6 | `python -m compileall -q src` | `0` | — |
| 7 | `python -m build` | `0` | sdist + wheel built |
| 8 | `python -m twine check dist/*` | `0` | wheel `PASSED`, sdist `PASSED` |
| 9 | `git diff --check` | `0` | — |
| 10 | `test ! -e uv.lock` | `0` | — |

The 6 local skips are the Windows-execution tests, which require a Windows host;
they are covered natively below. The repository configures no lint gate, so ruff
was run with the ambient default configuration. `src/csk/global_bins.py` and
`src/csk/installer.py` report the **same 8 pre-existing** ruff findings
(`SIM103`, `BLE001` ×3, `I001`, `UP017`) before and after this change; the new
and rewritten files are clean.

### Native Windows (`Microsoft Windows [Version 10.0.19045.6456]`, CPython 3.14.4)

Source synced to the host and run there; the local `.venv` was not reused.

| # | Command | Exit | Result |
|---|---|---|---|
| 11 | `python -m pytest -q tests/test_shims.py tests/test_build_activation.py` | `0` | 65 passed, 41 skipped |
| 12 | `python -m pytest -v -k "forwards_arg or errorlevel or never_launches or preserves"` | `0` | 7 passed, 6 skipped, 91 deselected |
| 13 | `python -m pytest -q` (full) | `1` | 1 failed, 852 passed, 183 skipped — see below |
| 14 | `python -m pytest -q --deselect tests/test_shell_init.py::test_powershell_hook_activates_and_restores_on_every_prompt` | `0` | 852 passed, 183 skipped, 1 deselected |
| 15 | `python -m mypy` | `1` | 48 errors in 6 files — see below |

Gate 12 confirms the Windows launcher contract executed for real on Windows:
`test_windows_launcher_forwards_arguments_and_preserves_errorlevel`,
`test_windows_launcher_forwards_arguments_and_nonzero_exit_status[False|True]`,
`test_windows_activation_quotes_the_executable_and_forwards_argv[project|global]`,
`test_activated_windows_command_forwards_argv_and_nonzero_exit_status[False|True]`
and `test_windows_activation_never_launches_the_built_artifact[project|global]`
all pass, including the case where a poisoned `ERRORLEVEL` environment variable
is present and exit code 7 / 9 must still propagate.

**Gate 13 is honestly red.** The single failure is
`tests/test_shell_init.py::test_powershell_hook_activates_and_restores_on_every_prompt`,
which cannot launch its `.ps1` because this host's PowerShell execution policy
denies it (`UnauthorizedAccess`). It is pre-existing and unrelated: the same test
fails identically on the **unmodified base source** on the same host
(`python -m pytest -q tests/test_shell_init.py` in a clean extraction of
`11160f6` → exit `1`, 1 failed, 7 passed, 14 skipped). A rerun with a
process-scoped `PSExecutionPolicyPreference=Bypass` did **not** clear it on this
host and still exited `1`; the host's security policy was deliberately not
modified. Gate 14 therefore isolates the remainder of the suite as green.

**Gate 15 is honestly red** and pre-existing. All 48 errors are Windows
platform-stub errors in `builds/_windows.py`, `builds/cache_posix.py`,
`builds/source.py`, `config.py`, `locking.py`, and `transactions.py`. The base
source on the same host reports 50 errors in 7 files (also exit `1`); the
2-error / 1-file difference is `src/csk/__init__.py` failing on the
setuptools-scm-generated `_version.py`, which is absent from the extracted base
tree and present in the worktree — not a change in behavior. **Neither run
reports a single error in `shims.py`, `global_bins.py`, or `installer.py`**, and
strict mypy on the authoritative Linux/macOS configuration is green (gate 4).

## Acceptance-criteria coverage

| Criterion | Evidence |
|---|---|
| Reuse only when all required active script paths are regular, contained, present | `test_runtime_reuse_keeps_complete_commit_keyed_state`, `..._replaces_state_with_a_missing_required_command[unix\|windows]`, `..._when_a_required_path_is_a_directory`, `..._that_escapes_through_a_link`, `..._reached_through_a_linked_directory`, `..._a_runtime_path_that_is_not_a_directory`, `..._with_a_missing_declared_root`, `test_runtime_root_command_path_rejects_a_linked_command_file` |
| Missing or wrong-type paths trigger staged replacement | same set: each asserts the required path is restored as a real regular file and unmanaged leftovers are gone |
| Unix project and global activation targets the immutable cache artifact | `test_unix_project_activation_targets_the_immutable_cache_artifact`, `test_unix_global_activation_targets_the_immutable_cache_artifact` |
| Unix activation forwards argv and exit status without profile setup | `test_unix_activation_needs_no_shell_profile` (whole launcher asserted), `test_activated_unix_command_forwards_argv_and_nonzero_exit_status[False\|True]` (exit 9 + argv) |
| Windows `.cmd` quotes the executable, forwards `%*`, preserves ERRORLEVEL | `test_windows_activation_quotes_the_executable_and_forwards_argv[project\|global]`, `test_windows_launcher_forwards_arguments_and_preserves_errorlevel`, `test_activated_windows_command_forwards_argv_and_nonzero_exit_status[False\|True]` (native Windows, poisoned ERRORLEVEL, exit 9) |
| Windows rejects command-name or path injection | `test_launcher_rejects_command_name_injection` (16 cases), `test_launcher_rejects_target_path_injection` (6 cases), `test_activation_rejects_an_injectable_artifact_path`, `test_activation_rejects_a_command_name_that_is_not_a_portable_identifier`, `test_launcher_rejects_a_path_entry_carrying_the_platform_separator`, `test_launcher_rejects_a_relative_target`, `test_windows_activation_escapes_percent_in_the_artifact_path` |
| Mixed script and build commands share collision and stale-shim rules | `test_mixed_script_and_build_commands_share_one_launcher_namespace[unix\|windows]`, `test_a_build_command_replaces_a_script_launcher_of_the_same_name[unix\|windows]`, `test_stale_shim_removal_keeps_every_expected_command_name[unix\|windows]`, `test_windows_stale_shim_removal_maps_a_bare_name_and_a_dot_cmd_name`, `test_user_bin_publication_treats_build_and_script_commands_alike`, `test_user_bin_publication_reports_an_unmanaged_build_command_conflict` |
| Install-time tests prove no built artifact is launched | `test_activation_never_launches_the_built_artifact[project\|global]` (POSIX) and `test_windows_activation_never_launches_the_built_artifact[project\|global]` (native Windows): a runnable artifact writes a sentinel, absent after activation, present only after an explicit run |
| Explicit post-install tests prove argv and nonzero exit codes propagate | the two `test_activated_*_command_forwards_argv_and_nonzero_exit_status` families, exit 9 on both platforms |
| POSIX and Windows-focused pytest plus strict mypy pass | gates 2, 3, 4, 11, 12, 14 |

Negative selection matrix also covered: non-HIT statuses (miss, corrupt,
untrusted-provenance, unsupported), missing receipt/artifact state on a claimed
hit, marker/receipt disagreement on receipt hash, cache key, artifact hash and
artifact path, a receipt built for another command, a receipt target that is not
the activation platform, an artifact whose on-disk size left the receipt, a
missing artifact, an artifact reached through a link, an artifact that is not
owner-executable, an artifact outside the manager home, an artifact inside the
script runtime namespace, an artifact path that is not the derived tail, a script
command, an unsupported driver, a hand-built relative artifact path, a malformed
identity, and an untyped activation.

## Scope boundaries respected

- Optional shell-profile behavior is untouched: `shell_init.py` and
  `env_files.py` were not modified.
- `closure.active_commands` still exports script commands only. Selecting build
  commands into the activation plan belongs to the planner and integration tasks
  (`TASK-260720-2x6mjn`, `TASK-260720-3t8nr3`, `TASK-260720-g7kgox`); this task
  provides the typed activation API those tasks call and proves the launcher,
  collision, and stale-shim rules are already kind-agnostic.
- Receipt and cache-key semantics, and the protected-cache backends, were not
  changed.
