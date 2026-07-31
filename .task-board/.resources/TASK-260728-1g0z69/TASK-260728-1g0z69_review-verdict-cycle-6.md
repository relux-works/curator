# Review verdict — cycle 6

## Verdict

CHANGES REQUESTED. Route to `analysis` for research/decision rework. This is ordinary recoverable rework, not a stop-the-line blocker.

## Blocking finding

The cycle-5 rework correctly isolates Go semantic representability from the running host TooNew gate for the `go` directive, but its `toolchain`-directive command corroboration is not fail-closed. In `.temp/TASK-260728-1g0z69/probe/main.go:489-500`, the state starts as accepted and the default branch classifies every unrecognized non-zero `go build` result as accepted. The code explicitly recognizes `invalid toolchain`, `invalid GOTOOLCHAIN`, and `cannot find`, but an unrelated failure or a future upstream rejection layer also lands in the accepted default.

For a shape-valid name that the isolated `FromToolchain` measurement considers representable, such an unknown command failure therefore agrees with the isolated result and the probe remains green. That contradicts `docs/compiled-build-toolchain-requirements.md:790-804` and `decisions/0007-compiled-build-toolchain-preflight.md:678-683`, which require the command outcome to corroborate composition and claim that a disagreement catches drift in either measurement. It also leaves the reviewer directive question about an unmeasured later acceptance layer unresolved.

## Required rework

1. Make the `toolchain` command-outcome classifier closed and fail-closed. Treat exit 0 as accepted; accept a non-zero search outcome only when it structurally identifies the exact tested toolchain name in the expected `cannot find ... in PATH` form; recognize the closed invalid-name outcomes as rejected; treat every other non-zero outcome as unknown and fail the probe rather than mapping it to accepted.
2. Add a runnable expected-red control that injects an unrelated non-zero command failure for a shape-valid, isolated-representable toolchain name and proves the probe fails for `unknown command outcome`.
3. Update the decision/reference probe obligation and vector/control inventory consistently, while preserving the cycle-5 representability/TooNew separation and the three existing expected-red controls.

## Independent evidence

- Exact candidate delta versus the accepted predecessor: only `CHANGELOG.md`, `decisions/0007-compiled-build-toolchain-preflight.md`, and `docs/compiled-build-toolchain-requirements.md` (excluding worktree metadata).
- `uv run --with jsonschema==4.25.1 python tools/validate.py`: exit 0, 42 schemas / 422 vector files.
- Python unittest discovery: exit 0, 29 tests.
- `go test ./...`, `go vet ./...`, `gofmt -l tools`, and `git diff --check`: exit 0 / no output where expected.
- Isolated boundary probe on Go 1.25.1 and 1.25.5: exit 0, 58 cases, zero failures. Future release/prerelease values are representable and command-classified too-new; patch-prerelease and invalid toolchain values remain rejected.
- Existing three red controls each exit 1.
- Clean-probe `make regenerate-check` and `make release-check VERSION=1.0.0-rc.5`: exit 0; release gate passed at `9c752ead345c716c44417350fef9d7a7ea700c02`.
- No candidate artifact was edited, staged, committed, published, pinned, or used for a new platform claim.