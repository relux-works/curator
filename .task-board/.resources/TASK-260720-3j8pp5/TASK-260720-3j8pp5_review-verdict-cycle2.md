# TASK-260720-3j8pp5 review verdict — cycle 2

## Verdict

**Accepted → `done`.**

No acceptance-blocking findings remain. The telemetry config-repoint false
accept from cycle 1 is closed, the implementation matches the task boundary,
and all independent validation gates are green.

## Provenance and read-only scope

- Reviewed worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3j8pp5/worktree`.
- Branch: `task/TASK-260720-3j8pp5-toolchain-identity`.
- Recorded base, worktree `HEAD`, `main`, and `origin/main`:
  `dd76b570f88339fd1d659c02950e68b17f6ba834`.
- Accepted dependency `TASK-260720-z9j4c9` is `done` and has handoff artifacts.
- Product scope remains exactly two untracked task-owned files:
  `src/csk/builds/toolchain.py` and `tests/test_builds_toolchain.py`.
- Reviewed SHA-256 values match the producer rework handoff:
  - `636da1d5a9ad97bb5224783fdbee66549f9bb466f31e8504918847c2714ad5f5`
    — `src/csk/builds/toolchain.py`
  - `c0bc352b85545afac5c2b8303ea727e2c6f06ca7a3fc8e6967d0b0b774eef00a`
    — `tests/test_builds_toolchain.py`
- The reviewer did not modify, stage, or commit product code.

## Prior finding closure

The rework anchors the canonical operation root and every real directory
component through the platform configuration root, rejecting symlink, reparse,
object-identity, and physical-location changes. Those anchors are checked
before and after each exact probe, during environment validation, and after
fingerprinting. `GOTELEMETRYDIR` must resolve strictly below both the anchored
configuration root and canonical operation root.

The regression independently passed for replacement during each permitted
probe:

1. direct `go telemetry off`;
2. direct `go version`;
3. direct `go env -json` with the fixed field vector.

Every case returned `telemetry_directory_untrusted`, removed the nominal
operation root, left the private base empty, and preserved the external target.
Cleanup stays scoped to the nominal private root and does not traverse the
outward symlink.

## Acceptance audit

- Operator `PATH` is captured immutably before project augmentation; relative
  or empty entries fail closed.
- Repository/project-managed candidates, wrappers, non-executable launchers,
  and executable/GOROOT disagreements fail before probes.
- The only bootstrap argv forms are the three required direct invocations,
  all from the operation-private empty working directory and a constructed
  bootstrap environment.
- The exact unique string-valued `go env -json` field set is enforced.
- Go 1.25 is the sole allowlisted release family; pre-1.23, unknown families,
  malformed releases, host/target mismatches, and malformed tuning fail closed.
- The native target freezes `GOOS`, `GOARCH`, and exactly one applicable tuning
  variable in immutable mappings.
- `curator-go-toolchain-v1` frames the complete normalized GOROOT tree,
  including directories, file bytes, internal relative link text, and
  normalized version output; shared preimage/digest vectors are byte-exact.
- Invalid/duplicate protocol paths, special files, absolute/escaping/dangling
  links, executable failures, and tree mutation fail closed.
- No `go list`, `go build`, package-directory probe, manager build-policy
  field, or out-of-scope product change is present.
- Imports and reparse handling are guarded for Windows; strict mypy targets
  Python 3.11. Native Windows execution was not available on this Darwin host.

## Independent validation

Authoritative interpreter:
`/Users/iv/Developer/intranet/cocoaskills/.venv/bin/python`
(Python 3.14.4, pytest 9.0.3, mypy 2.1.0).

| Gate | Exit | Result |
|---|---:|---|
| Config-repoint regression, verbose | 0 | `3 passed, 59 deselected` |
| Focused toolchain pytest | 0 | `62 passed in 0.25s` |
| Strict `python -m mypy` | 0 | `Success: no issues found in 57 source files` |
| Full pytest | 0 | `636 passed, 19 skipped in 83.79s` |
| Task-file 119-column/trailing-whitespace check | 0 | No findings |
| Forbidden `go list` / `go build` scope search | 0 | No forbidden invocation |
| Real Go probe and close | 0 | Go 1.25.5 Darwin/arm64, `GOARM64=v8.0`; expected identity; operation root absent and private base empty |

Real-Go identity:
`sha256:69f6b3484a10b288561c7fc66be60945e48b7628978c7baafbaa2ca5c823da0b`.
