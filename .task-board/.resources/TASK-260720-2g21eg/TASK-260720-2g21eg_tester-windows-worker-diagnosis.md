# TASK-260720-2g21eg — tester diagnosis: Windows worker launch

## Scope and provenance

- Role directive: diagnose the native Windows fixture failure without editing
  product source or racing the active producer.
- Preserved CocoaSkills worktree:
  `/Users/iv/Developer/Wildberries/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Task branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- Recorded worktree/base commit:
  `15860e3f309888845b9271a257fb95f7c2825b56`
- Clean canonical `main` and `origin/main` both resolved to that commit after
  `git pull --ff-only` exited 0 (`Already up to date`).
- Dependency handoffs `TASK-260720-3c0ss2` and `TASK-260720-3j8pp5` were both
  read from the board as `done` with outcome evidence present.
- Native Windows host: `ssh win`, hostname `DESKTOP-3PBO632`.
- No product source or test file was edited by this tester run. The only new
  local files are task-scoped diagnostic/evidence artifacts under
  `.temp/TASK-260720-2g21eg/`.

## Tool readiness

Every command below ran directly as a standalone process.

| Command | Exit | Result |
|---|---:|---|
| `task-board --version` | 0 | `task-board 0.23.0` |
| `git --version` | 0 | `git 2.50.1` |
| `rg --version` | 0 | `ripgrep 15.2.0` |
| `python3 --version` | 0 | `Python 3.14.4` |
| `python3 -m pytest --version` | 1 | Expected readiness red: system Python has no `pytest` |
| `.venv/bin/python -m pytest --version` | 0 | Project environment has `pytest 9.1.1` |
| `.venv/bin/python -m mypy --version` | 0 | Project environment has `mypy 2.3.0` |
| `ssh -V` | 0 | `OpenSSH_10.2p1` |
| `ssh ... win hostname` | 0 | Native host reachable as `DESKTOP-3PBO632` |

## Failure evidence

The existing native JUnit artifact
`TASK-260720-2g21eg_cycle5-windows-native.xml` records:

- 10 collected scenarios;
- 3 failures and 5 skips in 360.188 seconds;
- the plain accepted scenario failed earlier in toolchain fingerprinting with
  `toolchain_timeout`;
- `accepted-with-private-runtime` and
  `accepted-with-bound-startup-hook` reached the manager/worker boundary but
  failed with `BrokenPipeError` wrapped as
  `build_execution_worker_protocol_invalid: cannot write the worker session
  channel`.

That JUnit file was inspected, not rerun or represented as a green gate by this
tester.

### Native hidden-worker observation

The first diagnostic launched the installed wheel's hidden worker without
writing a request. It exited 2 after about five seconds with:

```text
csk: error: argument command: invalid choice: '__csk-go-worker-v1'
```

The diagnostic command itself exited 0 because it intentionally reported the
child result. This proves the child did not enter `run_worker`; the
`_write_message` exception in the manager is a downstream symptom.

### Inherited capability and process ancestry

The capability probe launched the same bound Windows venv interpreter with the
same `-S -s -B -P` isolation and explicit inherited-handle list. The SSH
command exited 3, the probe's expected rejection status. Exact observations:

```text
parent_pid=13192 capability_handle=528
child_pid=2136 parent_pid=2780
pipes=[..., (528, 1, 88, b'csk-go-launch-v1', 6)]
consume_error=GoV1Error:
  build_execution_worker_identity_invalid:
  hidden worker is not a direct child of its launch manager
```

The manager-created launch capability was therefore present, uniquely
discoverable, exactly 88 bytes, and had the expected magic. Transport loss and
pipe backpressure are ruled out. The Windows venv `Scripts\python.exe`
redirector introduced PID 2780 between manager PID 13192 and executing Python
worker PID 2136. `_parse_worker_launch_record` correctly rejected the
grandchild before hidden dispatch; `cli.main` then deliberately fell through
to the public parser, which closed the request pipe.

## Minimal defensible fix and proof

For Windows only, launch the already identity-bound
`identity.interpreter.runtime.base_executable.path` instead of the venv
redirector `identity.interpreter.invocation_path`. Keep:

- the installed manager `csk.exe` as the fixed script;
- the exact hidden mode and `-S -s -B -P` flags;
- the manager-built environment;
- the base executable, runtime DLL, `pyvenv.cfg`, stdlib, and package tree in
  the verified identity and retained mutation guard.

The same capability probe with
`C:\Program Files\Python314\python.exe` exited 0:

```text
parent_pid=3492 capability_handle=532
selected_interpreter=C:\Program Files\Python314\python.exe
popen_pid=12372
child_pid=12372 parent_pid=3492
pipes=[..., (532, 1, 88, b'csk-go-launch-v1', 6)]
context={'parent_pid': 3492, 'transport': 'inherited-anonymous-pipe'}
```

This proves the proposed selection restores the exact manager-parent/worker
relationship while preserving the authenticated anonymous-pipe launch
capability.

The active producer independently placed this direction in the shared
candidate before the tester reported it:

- `_resolve_windows_interpreter_runtime` binds the base executable as the
  actual process image;
- `worker_argv` selects the bound base executable for a Windows runtime;
- `validate_worker_runtime` permits and checks that bound executable;
- `test_windows_venv_runtime_layout_resolves_bound_base_installation` asserts
  the exact fixed argv selects the base interpreter.

## Follow-up native failure: console archive rewrites `sys.argv[0]`

The producer installed the base-interpreter candidate and reran the complete
native Windows fixture. The new JUnit had 10 tests, 3 failures, and 5 skips in
172.752 seconds. Hidden literal rejection and the runtime-image mutation
negative passed, so the direct-child launch change was active. All three
accepted build cases failed after worker authentication with:

```text
build_execution_worker_identity_invalid:
installed manager launcher is unavailable
```

The manager reached `_receive("ready")`; the worker returned this as a framed
failure. This is later than the original broken pipe and proves hidden dispatch
and the request channel now work.

A task-scoped overlay replaced only the console archive's imported
`csk.cli.main` with an argument reporter. Running the unchanged installed
`csk.exe` under the bound base interpreter, exact isolation flags, and hidden
mode exited 0 with:

```text
sys_argv=[
  'C:\\Users\\admin\\TASK-260720-2g21eg-cycle5-61f29f\\venv\\Scripts\\csk',
  '__csk-go-worker-v1'
]
sys_orig_argv=[
  'C:\\Program Files\\Python314\\python.exe',
  '-S', '-s', '-B', '-P',
  'C:\\Users\\admin\\TASK-260720-2g21eg-cycle5-61f29f\\venv\\Scripts\\csk.exe',
  '__csk-go-worker-v1'
]
```

The distlib Windows console archive strips `.exe` from `sys.argv[0]` before
calling `csk.cli.main`. `run_worker` therefore defaults `worker_executable` to
the nonexistent extensionless path, and `_resolve_worker_manager_identity`
correctly reports that the launcher is unavailable. The actual fixed launch
record remains intact in `sys.orig_argv`.

The producer implemented an equivalent closed recovery: on Windows only,
restore the one fixed `.exe` suffix when distlib supplies a suffixless
`sys.argv[0]`; do not consult `PATH`, the environment, or package input.
The authenticated request still binds and verifies the recovered launcher.
The focused unit case
`test_windows_distlib_argv0_restores_only_fixed_executable_suffix` covers both
the distlib form and the already-suffixed form.

## Follow-up native failure: unavoidable Python-home import root

The next producer wheel passed hidden-literal and runtime-image mutation
negatives, but all three accepted scenarios failed after authenticated worker
startup with:

```text
build_execution_worker_identity_invalid:
worker import path escapes the bound manager TCB:
C:\Program Files\Python314
```

The private-runtime accepted scenario reported its own fixed private
`base-python` root in the same position. The bound base interpreter therefore
necessarily contributes its Python installation prefix to `sys.path` under
the fixed `-S -s -B -P` launch. The previous runtime identity covered `Lib`
and `DLLs` but not root-level modules or namespace directories below that
prefix, so simply allowing the prefix would be a forced fit: code placed at
the root could execute before the worker sent its identity proof.

The minimal defensible correction is to make the complete Windows
`python_home` the bounded runtime tree:

- recursively bind every directory plus importable `.py`, `.pyc`, `.pyd`,
  `.dll`, and `.so` file;
- retain the explicit exclusion of `site-packages`/`dist-packages`, which the
  fixed `-S -s` launch disables;
- admit only that exact bound tree on the worker import path;
- preserve the fixed file-count and byte bounds and mutation verification.

This closes root-level module insertion instead of weakening the import-path
check. A synthetic insertion of `root_level_attack.py` must make
`runtime_tree.verify()` reject with
`build_execution_worker_identity_invalid`.

### Coupled identity invariants

The first implementation of that correction exposed two stale assumptions.
The task-scoped expected-red regression
`TASK-260720-2g21eg_test_windows_python_home_identity.py` ran standalone and
exited 1 with two failures:

1. `_startup_identity_from_mapping(startup.to_dict(), interpreter)` rejected
   the manager's own Windows identity because deserialization still required
   `runtime_trees[0].path == stdlib_root`; the exact expected roots must be
   derived with `_manager_runtime_roots(stdlib_root, interpreter)`.
2. An existing fixed `python314.zip` was not included in `startup.archives`.
   Archive discovery still inspected `runtime_root.parent`; with
   `runtime_root == python_home`, the fixed archive slot is inside the root.
   Existing archives must be resolved and fingerprinted from the declared
   fixed archive-slot parent, never admitted only because their path matches
   an unbound slot.

Both checks are protocol/identity requirements, not cosmetic test
expectations. A native rerun should start only after the round trip and
existing-archive cases pass.

The unchanged native Windows fixture independently confirmed the first
failure: 10 tests, 3 failures, 5 skips in 235.638 seconds; every accepted
variant returned
`build_execution_worker_protocol_invalid: manager runtime identity carries
inconsistent roots`. The raw JUnit is attached as
`TASK-260720-2g21eg_cycle5-windows-python-home-producer.xml`.

After the producer derived the exact runtime roots during deserialization and
resolved existing archives from the fixed declared slots, the same tester
regression ran green with 2 passed and exit code 0. This is the required local
gate before another wheel/native run; it does not by itself claim a native
build pass.

## Required regression boundary

The minimal test set is:

1. The synthetic Windows venv layout test must assert the runtime process image
   and `worker_argv[0]` are the bound base executable, not the redirector.
2. Hidden-dispatch tests must assert that the exact `sys.orig_argv` launcher is
   passed to `run_worker`, and reject malformed, extra, or reordered original
   argv even when `sys.argv` contains the hidden literal.
3. A native Windows wheel-installed capability test must prove the worker PID
   equals `Popen.pid`, its parent PID equals the manager PID, the one 88-byte
   inherited capability is consumed, and hidden dispatch does not fall through
   to argparse.
4. The native Windows fixture must pass all accepted build scenarios, prove one
   list and one build, and retain the never-run invariant.
5. Windows startup identity must survive an exact manager-to-worker mapping
   round trip after changing the runtime root to `python_home`.
6. If the fixed `pythonXY.zip` slot exists, it must be included in the bound
   archive identity before the worker launches.

The separate 120-second toolchain fingerprint timeout is not caused by this
worker ancestry defect and should retain its own performance/fixture
investigation if it reproduces after the worker fix.

## Final tester gate ledger

Every gate command ran directly as a standalone process; no gate was piped
through `tee`.

| Gate | Exit | Result |
|---|---:|---|
| Focused `tests/test_builds_go_v1.py` | 0 | 139 passed, 1 platform skip |
| Default full pytest | 0 | 991 passed, 70 platform/optional skips |
| Strict `python -m mypy` | 0 | No issues in 65 source files |
| `python -m compileall -q src tests` | 0 | No compile errors |
| `python -m tabnanny src tests` | 0 | No indentation errors |
| Tester Python-home regression before coupled fix | 1 | Expected red: 2 failures reproduced stale root deserialization and omitted fixed archive identity |
| Same tester regression after coupled fix | 0 | 2 passed |
| Exact candidate-8 native macOS fixture | 0 | 10 passed, no skips or failures in 63.985 seconds |
| Exact candidate-8 native Windows fixture | 0 | 5 passed, 5 host-specific skips, 0 failures/errors in 287.548 seconds |
| Independent tester native Windows plain accepted fixture on candidate 8 | 0 | 1 passed, 0 skipped in 62.09 seconds; JUnit case time 62.089 seconds |

The first attempt to invoke the tester Windows script exited 1 before pytest
because the command addressed it at `C:\Users\admin` rather than its actual
task-root location. This was a setup-path failure and is not reported as a
test result. The corrected invocation changed only that path and produced the
green independent gate above.

The candidate-8 full Windows fixture exercised all three accepted build
variants plus the hidden-literal and runtime-image mutation negatives. The
same exact wheel passed all 10 native macOS scenarios. The separate tester run
repeated the plain accepted Windows case against that installed wheel and fixed
Go tree with no skip. The source-aware process reached exactly the fixed
`go list` and `go build` forms; output verification passed and the fixture's
never-run marker remained absent.

The first cycle-5 changed-scope measurement was genuinely red at 64.99%
combined, so the coverage checklist was not checked. Narrow tests were then
added for the uncovered cycle-5 interpreter-layout, configuration, archive,
descriptor-capacity, and worker-runtime-proof behavior. Against the exact
reviewed cycle-4 source and final cycle-5 source, the standalone measurement
then exited 0 with:

- changed executable lines: 271/291 (93.13%);
- changed branches: 94/106 (88.68%);
- combined changed line/branch outcomes: 365/397 (91.94%).

That is the board item’s affected-code scope and exceeds the requested ~80%.
For context, coverage across the entire 3,093-statement driver remains 57%
branch-aware (1,858 covered lines; 507/1,050 covered branches); no 80%
whole-module claim is made.
