# TASK-260728-1g0z69 — review verdict cycle 8

## Verdict

ACCEPTED. No blocking or non-blocking findings. Route to `done`.

The cycle-7 rejection-direction laundering defect is closed. The decision is canonical and implementation-ready, the reference and probe agree, and the full task acceptance criteria remain satisfied.

## Goal and scope

- Run `RUN-260728-f07699` reports no active goal; no reviewer-goal branch was required.
- Read-only `diff -qr -x .git` against accepted predecessor `TASK-260728-2kp3tv` found exactly three task files: `CHANGELOG.md`, `decisions/0007-compiled-build-toolchain-preflight.md`, and `docs/compiled-build-toolchain-requirements.md`.
- Attached decision/reference SHA-256 values exactly match the candidate worktree: decision `7eb41f159743378c10b099b92bb62767ab9b3fa158015e84fddd135f6ba8d057`; reference `8807c4961f4165fd615abd2d1b0c5eeac1780c2baecf7bb4ac1ec9c9ece6f3d9`.
- Nothing is staged. Candidate tracked artifacts were not edited during review.

## Contract review

The contract keeps the toolchain identifier set closed (`go`, `rust`, `swift`, `kotlin`, companion-only `jdk`) and resolves roots only from manager-bundled or owner-protected operator configuration. Package data can only narrow a canonical version interval; it cannot add executable paths, roots, URLs, mirrors, channels, install commands, environment overrides, credentials, checksums, or trust roots. Compatibility is a separate manager-owned exact family gate.

Canonical versions are bounded `major.minor.patch` triples, constraints are exact/at-least/half-open-range intervals, intersection is associative and commutative, and prereleases are neither expressible nor satisfiable. Stage A runs after manifest validation and before acquisition, cache lookup, or mutation; Stage B runs only after local validation or exact external acquisition plus audit and before cache reads or compiler work. Metadata is assertion-only and cannot select or re-resolve a toolchain. Resolved toolchain identities bind cache, receipts, markers, and currentness; requirements, compatibility sets, and guidance catalog revisions remain gates/presentation rather than artifact identity. v1 never auto-installs.

All twelve typed diagnostics have deterministic firing sites and payloads. Guidance IDs are manager-owned, revisioned, immutable, total across supported toolchains/reasons/platforms, primary-source-only, and carry text plus URL rather than executable instructions. Reserved Rust/Swift/Kotlin/JDK entries have explicit qualified-host obligations and cannot reopen selection or auto-install.

## Cycle-8 classifier evidence

- Attached probe source SHA-256 matches the runnable probe source: `ae9cf37dcb4e6dd994d94510ab47886ada49436129bb3f5468345f6cf3d114c7`.
- Production classification uses exact equality for each trimmed diagnostic line against forms predicted from the tested value and fixed run context. No prefix, substring, or open-ended matcher is consulted unless the explicit `-red open-classifier` control is enabled. Unknown is a distinct non-verdict and fails the run. Conflicting recognised states also become unknown.
- Independent inspection of both Go 1.25.1 and 1.25.5 `toolchain.Select` sources confirmed the reachable `local+path` value-bearing outcomes: exact cannot-find, invalid-toolchain-in-go.mod, and invalid-GOTOOLCHAIN forms. Colon-bearing initial-setting diagnostics quote the fixed environment setting and are unreachable for the tested value.
- Independent green probe on `/usr/local/go/bin/go` (1.25.1) and `/opt/homebrew/bin/go` (1.25.5): exit 0; 16 go-directive plus 13 toolchain-directive cases per toolchain, 331 closure checks per toolchain, zero failures, zero direction-A fabrications, zero direction-B fabrications. Probe `gofmt` and `go vet` passed.
- Expected-red controls independently rerun on both toolchains: `open-classifier` exit 1 with 9 direction-A and 15 direction-B laundered fabrications per toolchain; `unrelated-command-failure` exit 1 with 4 unknown-outcome failures; `tidy-exit` exit 1 with 12 failures; `patch-prerelease-compared` exit 1 with 12 failures; `c-equals-upstream` exit 1 with 2 failures. Each failed for its named reason.

## Repository and release gates

Independent candidate checks passed: `tools/validate.py` validated 42 schemas and 422 vectors; 29 Python unit tests passed; `go test ./...`, `go vet ./...`, `gofmt -l tools`, and `git diff --check` passed.

The clean cycle-7 snapshot matched the candidate byte-for-byte outside `.git` and runtime cache. `make regenerate-check` passed twice. The first ambient-Python release invocation stopped before the gate because ambient Python lacks `jsonschema`; rerunning with the task validation virtual environment passed validation, unit tests, Go tests, regeneration, and `make release-check VERSION=1.0.0-rc.5` at snapshot `e6bb0bd7442140f44d645705faecc73239b8e373`. The clean snapshot remained git-clean. `conformance/v1` and `release/1.0.0-rc.5.json` are byte-identical to the accepted predecessor.
