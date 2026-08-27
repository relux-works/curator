# TASK-260720-2g21eg review verdict — cycle 4

## Verdict

Changes requested. Route to `to-dev`.

Cycle 4 closes the prior ordinary standard-library-file and plain hidden-mode
CLI findings, and all requested local gates are green. The worker identity is
still incomplete, however: it binds the Python launcher but not the mutable
Python runtime image that the operating system loads to execute the worker.
The Windows path also derives the runtime from a standard virtual environment
incorrectly. These are implementation defects with concrete rework paths, not
an external or human-only stop-the-line boundary.

## Candidate and provenance

- CocoaSkills worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- Base and `HEAD`: `495ad021847529ce5a544dba415ca2fe19949539`
- Clean local `main`, local `origin/main`, and live
  `refs/heads/main`: `495ad021847529ce5a544dba415ca2fe19949539`
- Dependencies `TASK-260720-3c0ss2` and `TASK-260720-3j8pp5`:
  accepted `done`
- Candidate scope remained exactly:
  `src/csk/builds/go_v1.py`, `src/csk/cli.py`,
  `tests/test_builds_go_v1.py`, and
  `tests/test_builds_go_v1_fixture.py`
- Candidate `go_v1.py`:
  `sha256:4d5d6b33ed7b5597531271d37372d926cc961519abc59346792893e2f0d71283`
- Candidate `cli.py`:
  `sha256:ad25c28b8918c3e03d945f4eef2c52e214e9161d25d477ec92def6a22f1aa943`
- Reviewed wheel:
  `sha256:ddf1ba4f7e779df818a1217bf5a0754fd627132838e39e6cc0c8a594679e4b36`;
  its installed `go_v1.py` and `cli.py` matched the candidate hashes.

No CocoaSkills product or test file was modified during review.

## Blocking finding 1 — the executing Python runtime image is outside identity

`_InterpreterIdentity` contains only the invocation/link chain and one final
executable file (`go_v1.py:1416-1441`). `_resolve_interpreter_identity`
therefore hashes the small interpreter launcher only
(`go_v1.py:2361-2392`). `_manager_runtime_roots` adds the standard-library
tree and, on Windows, `DLLs`, but it does not discover the interpreter's
actual loaded runtime image (`go_v1.py:2596-2638`). The worker proof reports
`sys.executable`, import paths, and `sys.modules`; it does not report loaded
Mach-O images or Windows runtime DLLs (`go_v1.py:5109-5266`).

The exact installed wheel used by the passing native fixture resolved:

```text
interpreter_launcher=/opt/homebrew/Cellar/python@3.14/3.14.4_1/Frameworks/Python.framework/Versions/3.14/bin/python3.14
loaded_python_runtime=/opt/homebrew/Cellar/python@3.14/3.14.4_1/Frameworks/Python.framework/Versions/3.14/Python
runtime_exists=True
runtime_user_writable=True
runtime_bound_as_file=False
runtime_directly_watched=False
runtime_size=5438672
```

`otool -L` identifies that 5.4 MB framework image as the code loaded by the
52 KB launcher. It is owner-writable on this host, absent from every launcher,
package-tree, runtime-tree, archive, and hook identity entry, and absent from
the direct mutation-watch set. The identity happens to watch the containing
`python_home` directory, but a separate kqueue reproduction changed an
existing child's bytes in place and reported:

```text
directory_guard_detected_child_content_write=false
```

Consequently this runtime image can change in place without changing the
manager identity, worker runtime proof, retained-identity result, or post-exec
identity re-resolution. Mutable code executes before the ready proof while
`identity-verified-manager-owned-worker`,
`pre-launch-worker-identity-verification`, and
`post-exec-identity-reverification` remain claimed.

Required rework:

- Discover and bind the actual non-system interpreter runtime image/load
  components for the supported installed-manager layout, not only the launcher.
- Carry those identities through the request/proof, pre-launch verification,
  retained mutation guard, worker-side proof, and post-exec verification.
- Add a wheel-installed native negative using a task-private runtime copy:
  mutate or replace the loaded runtime image after launch and require
  `build_execution_worker_identity_invalid`, no compiler start, no result, and
  complete worker-domain teardown.

## Blocking finding 2 — a normal Windows venv resolves the wrong runtime root

On Windows, `_manager_interpreter_path` selects
`<venv>/Scripts/python.exe` (`go_v1.py:2485-2498`), while
`_manager_stdlib_root` derives `<venv>/Lib` from that executable and
`_manager_python_home` returns the venv itself
(`go_v1.py:2596-2629`). `_indispensable_worker_environment` then exports that
incorrect venv as `PYTHONHOME` (`go_v1.py:5007-5018`).

CPython's standard Windows `venv` path defaults to copies and installs a
`venvlauncher.exe` as `Scripts/python.exe`; it creates the venv purelib
directory, not a private copy of the base standard library. The actual Python
DLL and `Lib` remain in the base installation selected through `pyvenv.cfg`.
A synthetic standard Windows layout produced:

```text
selected_stdlib=<venv>/Lib
selected_python_home=<venv>
selected_runtime_non_site_files=0
selected_venv_lib=True
error_code=build_execution_worker_identity_invalid
error_detail=installed manager Python runtime does not contain a bounded regular-file tree
```

Thus a normal wheel installed in a Windows venv fails before worker launch.
If incidental files make the venv `Lib` scan nonempty, the proof still binds
the wrong tree and omits the base Python DLL/runtime that actually executes.
The required Windows implementation is therefore not satisfied by the
platform-neutral mocks.

Required rework:

- Resolve Windows venv/base-runtime semantics from trusted physical
  installation data, including `pyvenv.cfg` and the real Python runtime DLL,
  without executing an unbound selector.
- Bind the base interpreter image, standard library, DLL roots, and relevant
  link/launcher chain across all identity phases.
- Add a native Windows wheel-in-venv fixture and a runtime-image replacement
  negative. The fixture must reach the accepted real Go build, while the
  negative must fail before compiler start and prove teardown.

## Independent green gates

| Gate | Result |
| --- | --- |
| Focused source/toolchain/metadata/go-v1/fixture pytest | 295 passed, 4 skipped; exit 0 |
| Full repository pytest | 968 passed, 6 skipped; exit 0 |
| Standalone wheel-installed native macOS Go fixture | 8 passed; exit 0 |
| Project-configured strict `python -m mypy` | no issues in 62 source files; exit 0 |
| `git diff --check` and task-file tabnanny | clean; exit 0 |
| Task-file trailing whitespace / conflict markers | none |

JUnit SHA-256:

- focused:
  `9f169f18d0dccc853c11c812b6cb89353fb43c2465d5c29bccc24a861da548b6`
- full:
  `83a1c9020ab5eb83aefeac19b17a1e0ef5f0d8c2c2cbfe4992c71177b6d11324`
- native fixture:
  `cbf2d225389ec9453236f4b527eff1875a196eec396c2d0ad4089543fbd41b99`

The green tests prove the fixed argv/environment, graph rejection, protocol,
native-control evidence, output verification, plain hidden-mode parser
rejection, and macOS real-Go path. They do not enumerate the interpreter load
image or exercise a standard Windows venv.

## Routing

This is ordinary implementation rework. Preserve this verdict, route the task
to `to-dev`, and require a fresh independent reviewer cycle after the runtime
TCB and Windows venv paths are closed. Do not use `blocked`; no external input
or human-only architecture decision is needed.
