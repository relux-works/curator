# TASK-260720-2g21eg cycle-6 CI portability rework

Date: 2026-07-30

## Candidate

- Worktree:
  `/Users/iv/Developer/intranet/cocoaskills/.temp/TASK-260720-2g21eg/worktree`
- Branch: `task/TASK-260720-2g21eg-go-v1-compile-driver`
- HEAD before this uncommitted rework:
  `673a38dc2fac499cbcbfa3ff6e9be84d9bae3ee8`
- Pinned rc.5 conformance commit:
  `f5d7673039226ab81de2f4f87e2155ae995c4df3`
- Tracked delta remains exactly:
  `src/csk/builds/go_v1.py` and `tests/test_builds_go_v1.py`.

## Rework

- Made the Darwin kqueue boundary explicit to cross-platform mypy through
  runtime-preserving `Any` casts. No macOS control behavior changed.
- Added a test-local synthetic interpreter-runtime seam only on unsupported
  hosts. The captured production resolver is tested separately and still
  rejects Linux with `build_execution_control_unavailable`.
- Made the synthetic manager fixture represent the native Windows CPython
  layout: one `python.exe`, one versioned runtime DLL, `Lib`, and the installed
  manager package tree.
- Corrected Windows-only test assumptions: coherent host selection, no POSIX
  shebang/process-group assertions, portable fake termination status, Windows
  chmod semantics, runtime-root/archive expectations, and path-placeholder
  separators.
- Product support remains closed over exactly macOS and Windows. No Linux
  native-control path, host-label bypass, or weakened fail-closed behavior was
  added.

## Final green gates

| Gate | Host | Exit | Result |
| --- | --- | ---: | --- |
| Exact overlay/archive SHA-256 comparison | Windows amd64 over `ssh win` | 0 | Remote source, test, baseline archive, and pinned protocol archive matched local bytes |
| Focused `tests/test_builds_go_v1.py` | Windows, Python 3.14.4 | 0 | 131 passed, 10 platform skips |
| Full `python -m pytest -q` | Windows, Python 3.14.4 | 0 | 1,064 passed, 116 skips in 307.02 s |
| Focused accepted-root go-v1 unit suite | macOS arm64, Python 3.11.14 | 0 | 141 passed |
| Focused unsupported-Linux host selection | macOS process with `sys.platform=linux` | 0 | 129 passed, 12 platform skips |
| Full `python -m pytest -q` | macOS arm64, Python 3.11.14 | 0 | 1,159 passed, 21 skips in 81.46 s |
| Wheel-installed native Go fixture | macOS arm64, Python 3.14.4, Go 1.25.5 | 0 | 10 passed in 65.83 s |
| `python -m mypy` | macOS | 0 | No issues in 65 source files |
| `python -m mypy --platform linux` | macOS | 0 | No issues in 65 source files |
| `python -m build` | macOS | 0 | Final sdist and wheel built |
| `python -m twine check <final-dist>/*` | macOS | 0 | Both artifacts passed |
| `python -m compileall -q src tests` | macOS | 0 | Passed |
| `python -m tabnanny src tests` | macOS | 0 | Passed |
| `git diff --check` | macOS | 0 | Passed |

The final wheel-installed `csk/builds/go_v1.py` and source file both hash to
`f8433dc1e8eaecca5c4f04e574720e83f5fbf6a403a474576711297ff2cc0203`.

## Native Windows exact-byte method

The push-free Windows run used:

- a `git archive` of exact HEAD, SHA-256
  `86d3b32dd99b7f57a6f4f339f700f7cc3dbfe9d2ab5b1ea04b5c7b26211e9399`;
- an overlay of final `go_v1.py`, SHA-256
  `f8433dc1e8eaecca5c4f04e574720e83f5fbf6a403a474576711297ff2cc0203`;
- an overlay of final `test_builds_go_v1.py`, SHA-256
  `5fad24308d5e3ec0640e91e4ea553b1cbbeef101097bb1b173d7a7afc828f254`;
- the pinned protocol archive, SHA-256
  `5cf64fec9f2d54c5dbd81fbed5ec35014791c059fe46a62c4ce05880473405bb`.

The host reported Windows amd64, Python 3.14.4, Go 1.25.5, and Git 2.50.1.
The task-private remote root and all uploaded staging inputs were removed after
the JUnit files were downloaded. A final remote glob found no matching
task-owned path.

## Non-zero and superseded diagnostics

These results are reported as failures, not passes:

- Initial `python -m mypy --platform linux`: exit 1, 13 Darwin-symbol
  `attr-defined` errors. The final Linux-target mypy gate supersedes it.
- Initial forced-Linux focused suite: exit 1, 50 failed, 78 passed, 12 skipped.
  All failures came from synthetic fixtures reaching the intentionally
  unsupported production runtime resolver.
- One forced-Linux focused invocation used an incorrect absolute conformance
  path: exit 1, one `FileNotFoundError`; the corrected pinned-root invocation
  exited 0.
- A whole-suite `sys.platform=linux` emulation on macOS: exit 1, 101 failed,
  1,040 passed, 39 skipped. This is not a Linux-equivalent gate because
  unrelated cache/transaction tests attempted Linux kernel APIs against the
  Darwin C library.
- `python -m mypy --platform win32 src/csk/builds/go_v1.py`: exit 1 with 64
  platform-stub errors in existing POSIX/macOS branches and platform-specific
  `type: ignore` variance. CI strict mypy targets Linux; both configured and
  Linux-target strict gates are green.
- First native Windows focused run after the broad portability fix: exit 1,
  one failed, 130 passed, 10 skipped. The sole failure was a literal path
  separator in the Darwin environment-vector assertion; final focused and full
  Windows gates supersede it.
- Two preliminary macOS manager environments were rejected by the driver:
  the `uv` shell-wrapper launcher and a Python 3.11 installation with a
  symlinked startup `.pth` file. Both exits were 1 and were not claimed as
  fixture passes. The clean Python 3.14 installed-wheel fixture is the final
  native result.
- Native Linux `ssh lev hostname`: exit 255, connection timeout. Docker and
  Podman were unavailable locally, so no native Linux full-suite claim is
  made. The exact affected Linux-focused suite and Linux-target strict mypy are
  green.
- Conflict-marker search: exit 1 as expected for no matches.
- Remote preflight and post-cleanup glob checks: exit 1 as expected for no
  matching task path.

No GitHub rerun was triggered because this rework is intentionally uncommitted
and no commit/push authorization was provided.

## Outcome artifacts

| Artifact | SHA-256 |
| --- | --- |
| Final wheel | `63a45f87e295fffb13f1140e5e6f045448e690d3174b59c07d676a187093f602` |
| Final sdist | `8d1cb067f7d1f768c36f81dce95880960bc9299322ac6c1798ec1fdcec39ccf9` |
| macOS focused JUnit | `1ba917d6c1794bbcc9bf01046c682e079732744329b89daf2df3bb16f947ce47` |
| Linux-focused JUnit | `f190275512a5f9abc51fe40a7c235aa2b31ea1a9191bef834edaf71b4a102e36` |
| Windows focused JUnit | `2e0a7cf9c401dbb3f1ac2549c0b17ed1d49afa34f8d32fd28f94a29453edb253` |
| macOS full JUnit | `f506c9adb15c6181ba2c190c117dd640570715e5c9c6dc4ea95acf5609fd536e` |
| Windows full JUnit | `c019b9188723d794687ae327d88ac3dd2753ba63febb2900b040e9a14aba6ebc` |
| macOS native fixture JUnit | `157ffc3768a4cb39a5bb82da9e23374795dfb527aab22db2212eee6a4d664af7` |
