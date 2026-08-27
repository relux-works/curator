# TASK-260720-3j8pp5 review verdict

## Verdict

**Changes requested → `to-dev`.**

The existing focused, typing, full-suite, framing-vector, and real-Go controls
are green, but the telemetry containment check has a reproducible false-accept
path that violates the task acceptance criterion. This is ordinary bounded
implementation rework, not a stop-the-line blocker.

## Finding: relocated config roots are trusted outside the private boundary

Severity: high; security-boundary false acceptance.

`_establish_toolchain` passes only `layout.config` to `_validate_probe`
(`src/csk/builds/toolchain.py:622-629`). `_validate_probe` then resolves the
reported `GOTELEMETRYDIR` and the current config path and checks only that the
former is below the latter (`src/csk/builds/toolchain.py:779-790`). It does not
prove that:

- the physical config directory is still the real directory created in the
  operation-private root;
- the physical config directory remains inside `operation_root`; or
- the expected config path was not replaced by a symlink/reparse point during
  `go telemetry off`.

An isolated runner replaced Darwin's private
`HOME/Library/Application Support` directory with a symlink to an external
directory during the telemetry probe, then returned the exact fixed probe
shapes with `GOTELEMETRY=off` and the external telemetry path. The candidate
accepted the session:

```text
RESULT=FALSE_ACCEPT
telemetry_inside_operation_root=False
operation_root_exists_after_close=False
outside_telemetry_exists_after_close=True
```

The nominal operation root was removed, but the accepted telemetry state was
outside that root and remained after `close()`. This contradicts the explicit
requirements to verify `GOTELEMETRYDIR` inside the operation-private root and
delete the private telemetry state on every exit.

The existing test at `tests/test_builds_toolchain.py:367-380` covers a directly
external reported directory while leaving the expected config root intact. It
does not cover relocation of the config root itself, so both resolved paths can
agree outside the private boundary.

## Required rework

1. Bind probe-layout validation to the canonical operation root and to the
   original real platform config directory across all three probes.
2. Reject symlink/reparse or object replacement of that config directory and
   require both its physical location and the reported telemetry directory to
   remain below the canonical operation root at the platform-expected config
   location.
3. Add an outward config-repoint regression that proves establishment fails
   closed and verifies nominal operation cleanup; keep imports and runtime
   branches Windows-safe.
4. Rerun focused pytest, strict `python -m mypy`, the full suite, the exact
   shared digest vector, and the real Go smoke/cleanup check.

The implementation must not recursively clean an external target as a
workaround; containment must fail safely.

## Independent controls and provenance

- Product worktree:
  `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-3j8pp5/worktree`
- `HEAD = main = origin/main =
  dd76b570f88339fd1d659c02950e68b17f6ba834`
- Accepted dependency `TASK-260720-z9j4c9` is `done` with handoff artifacts.
- Product scope is exactly two untracked task-owned paths:
  `src/csk/builds/toolchain.py` and `tests/test_builds_toolchain.py`.
- File SHA-256 values match producer evidence:
  `a2284025742cfbfe53b914924c3c7289172a52b8c190ef6c0b577e7975851077`
  and
  `1c8c0faa5fb826d7b34c87bb6093c54cf57b8e18b1aabb10cc0b851ec2ea4d8f`.
- Focused pytest: exit 0, `59 passed in 0.21s`.
- Strict mypy: exit 0, `Success: no issues found in 57 source files`.
- Full pytest: exit 0, `633 passed, 19 skipped in 93.88s`.
- Task-file 119-column/trailing-whitespace gate: exit 0, no findings.
- Real Go smoke: exit 0; Go 1.25.5 Darwin/arm64, `GOARM64=v8.0`,
  `sha256:69f6b3484a10b288561c7fc66be60945e48b7628978c7baafbaa2ca5c823da0b`,
  nominal operation root removed, private base empty.
- Static scope search found no `go list`, `go build`, manager build-policy
  field, or out-of-scope product change.
- No product code or test file was modified during review.

