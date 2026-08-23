# TASK-260811-3kbf3l rework evidence after RUN-260819-e60581

Date: 2026-08-19

## Implemented reviewer requirements

### R07 real `include!` closure input

- `TestRustConformanceR03R05R06R07PathWorkspaceBuild` now builds a real
  `app/src/fragment.rs` through `include!("fragment.rs")`, alongside the
  existing `include_str!` and `include_bytes!` leaves.
- Before build, the test reads `fragment.rs` from the immutable protected
  workspace snapshot and validates the workspace intake receipt identity.
- Both fresh Cargo metadata and build process permits are observed, and each
  must name that exact protected workspace receipt in
  `AdmittedInputReceiptIDs`.
- The fresh frozen offline build publishes the executable only after those
  checks.

### RH10 end-to-end executor drift

- `TestRustConformanceRH10BuildRejectsToolchainDriftBeforeProcessOrPublication`
  replaces the former direct `rustBuildToolchain.recheck` probe.
- The test creates an admitted path-only closure, derives metadata, constructs
  the accepted C4 binding/publication evidence, preflights verified execution,
  and invokes `Manager.Build`.
- The injected verified provider mutates the manager-staged physical Cargo
  executable at the executor's post-permit time-of-use negotiation. The
  verified boundary hashes the executable against the committed permit before
  any child process start.
- The build returns stable `artifact_toolchain_identity_changed`; observed
  child-process starts remain zero; no execution receipt, publication receipt,
  artifact path, protected receipt, or protected blob is produced.
- `mapRustExecutionError` now preserves that shared executor diagnostic as the
  Rust adapter's typed `CodeToolchainIdentityChanged`, so callers receive the
  same stable machine code at every drift boundary.

## Validation evidence

Every command below was run directly and its real exit status was observed.

| Command | Exit | Result |
| --- | ---: | --- |
| `go test -count=1 ./internal/rustsource -run 'TestRustConformance(R03R05R06R07PathWorkspaceBuild|RH10BuildRejectsToolchainDriftBeforeProcessOrPublication)$'` | 0 | Focused R07/RH10 tests passed. |
| `go test -count=1 ./internal/rustsource ./internal/closureexec` | 0 | Rust 98.393s; closureexec 4.332s. |
| `go test -race -count=1 ./internal/rustsource ./internal/closureexec` | 0 | Rust 156.311s; closureexec 6.095s. |
| `make lint` | 0 | CI-pinned golangci-lint reported `0 issues.` |
| `go vet ./...` | 0 | No diagnostics. |
| `go build ./...` | 0 | Repository compiled. |
| `go test -count=1 ./...` | 0 | Full fresh suite passed; rustsource 150.941s, closureexec 22.042s, cmd/curator 406.710s. |
| `git diff --check` | 0 | No whitespace errors in tracked changes. |
| `task-board validate` | 0 | `Board is valid. No issues found.` |

## Scope and worktree note

Only the bounded R07/RH10 rework was changed in this run:

- `internal/rustsource/build_conformance_test.go`
- `internal/rustsource/build_test.go`
- `internal/rustsource/build_execution.go`

The broader `internal/rustsource` implementation and other repository changes
were already present in the shared dirty worktree and remain untracked or
unstaged. No commit, stage, reset, clean, or unrelated overwrite was performed.
