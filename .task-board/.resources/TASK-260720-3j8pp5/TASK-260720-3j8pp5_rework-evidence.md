# TASK-260720-3j8pp5 — telemetry-containment rework evidence

## Provenance and scope

- Product repository: `/Users/iv/Developer/intranet/cocoaskills`.
- Existing task worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3j8pp5/worktree`.
- Task branch: `task/TASK-260720-3j8pp5-toolchain-identity`.
- Recorded clean, fast-forwarded base: `dd76b570f88339fd1d659c02950e68b17f6ba834`.
- Accepted dependency `TASK-260720-z9j4c9` remains `done`.
- Product scope remains exactly:
  - `src/csk/builds/toolchain.py`
  - `tests/test_builds_toolchain.py`
- Current file SHA-256 values:
  - `636da1d5a9ad97bb5224783fdbee66549f9bb466f31e8504918847c2714ad5f5`
    — `src/csk/builds/toolchain.py`
  - `c0bc352b85545afac5c2b8303ea727e2c6f06ca7a3fc8e6967d0b0b774eef00a`
    — `tests/test_builds_toolchain.py`

## Review finding closure

The prior reviewer proved that replacing the platform config directory with an
outward symlink during `go telemetry off` could make the old validation accept
both the relocated config path and an external `GOTELEMETRYDIR`.

The rework now:

- captures immutable anchors for the canonical operation root and every real
  directory component through the platform-expected config root;
- rejects symlinks, Windows reparse points, changed directory objects, and
  physical path relocation;
- verifies the anchors immediately before and after each of the three exact Go
  probes, again during environment validation, and after tree fingerprinting;
- requires the real config root to remain below the canonical operation root
  and the resolved telemetry directory to remain below both roots; and
- keeps cleanup scoped to the nominal operation root rather than traversing or
  deleting the external target.

The regression repoints the config directory during each exact probe form:
`go telemetry off`, `go version`, and the fixed `go env -json` vector. Each
case fails with `telemetry_directory_untrusted`, removes the nominal operation
root, leaves the private base empty, and proves the external target remains
untouched.

## Validation ledger

Every gate ran as a standalone process without `tee` or a status-masking pipe.

| Command | Exit | Result |
|---|---:|---|
| Focused regression before the implementation | 1 | Expected red: `DID NOT RAISE`; old code reproduced the false accept |
| Focused repoint regression after the implementation | 0 | `3 passed, 59 deselected` |
| `python -m pytest tests/test_builds_toolchain.py -q` | 0 | `62 passed in 0.24s` |
| `python -m mypy` | 0 | `Success: no issues found in 57 source files` |
| Exact shared vector/preimage pytest selection | 0 | `2 passed, 60 deselected` |
| `PYTHONPATH=src python -m compileall -q src/csk` | 0 | No output |
| 119-column/trailing-whitespace `awk` gate on both task files | 0 | No findings |
| `python -m pytest -q` | 0 | `636 passed, 19 skipped in 81.42s` |
| Real task-source `establish_toolchain` smoke | 0 | Go 1.25.5 Darwin/arm64; `GOARM64=v8.0`; identity `sha256:69f6b3484a10b288561c7fc66be60945e48b7628978c7baafbaa2ca5c823da0b`; operation root absent and private base empty after close |
| `python -m build --wheel` | 0 | Built `cocoaskills-0.12.6.dev1+gdd76b570f-py3-none-any.whl` with `csk/builds/toolchain.py` |
| `python -m twine check dist/*.whl` | 0 | Wheel `PASSED` |
| Static scope search for forbidden `go list`, `go build`, and manager build-policy additions | 1 | Expected absence: `rg` found no matches |

The authoritative Python executable was
`/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python` (Python 3.14.4,
pytest 9.0.3, mypy 2.1.0). Native Windows execution was unavailable on this
macOS host. The implementation adds no platform-specific import, checks reparse
attributes through guarded stdlib access, compiles on POSIX, and passes strict
mypy; the symlink regression is skipped on Windows because unprivileged
symlink creation is not portable there.
