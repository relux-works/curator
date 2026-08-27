# TASK-260720-3pemm6 — candidate-suite and release-pin boundary

This planning artifact makes the compiled-build test and release sequence executable without guessing which commit may be written to csk CI.

## Sequence

1. `TASK-260720-3ag6pi` verifies the integrated protocol-v6 candidate suite and records its exact immutable SHA.
2. `TASK-260720-12r55p` consumes that candidate through caller-supplied `CURATOR_CONFORMANCE_ROOT`; it does not change the csk repository's committed curator-spec checkout ref.
3. `TASK-260720-3pemm6` adds the real Go fixture and Linux/macOS/Windows Go setup and executes candidate tests against that same exact suite through a non-default immutable revision or a pre-materialized root. CI/product changes must preserve the last qualified released suite pin.
4. `TASK-260720-3s27te` records integrated csk evidence against the exact candidate SHA. Candidate evidence is not represented as released-suite evidence.
5. `TASK-260720-25d05o` qualifies the published protocol release.
6. Only `TASK-260720-1utsx8` advances and audits every committed csk curator-spec checkout ref to the exact qualified release commit and enforces the no-skip cross-platform gate.

## Invariants

- No branch, mutable tag, guessed SHA, feature-worktree SHA, or unqualified candidate becomes the committed csk protocol pin.
- Candidate commands and reports always name the exact caller-supplied revision or `CURATOR_CONFORMANCE_ROOT` and its suite digest.
- Workflow edits in `TASK-260720-3pemm6` are limited to supported Go setup, real fixture execution, and non-default candidate-suite input; release-pin mutation is out of scope.
- Existing schema 1–5 and rc.3 coverage remains active until the qualified release pin is advanced by its owning task.
- The downstream interoperability consumer `TASK-260720-31zeo2` extends the external suite coverage after the csk implementation-vector handoff; it does not replace this platform E2E gate.
