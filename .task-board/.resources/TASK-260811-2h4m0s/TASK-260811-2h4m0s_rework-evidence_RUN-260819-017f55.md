# TASK-260811-2h4m0s rework evidence

Run: `RUN-260819-017f55`

Status target: developer handoff to review.

## Reviewer findings addressed

1. The production `cmd/curator` composition root now imports `internal/rustsource` and explicitly dispatches the exact `__curator_rust_git_oracle_v1` mode before ordinary CLI parsing. `go list -deps ./cmd/curator` resolves `github.com/relux-works/curator/internal/rustsource`.
2. Rust source package initialization no longer performs Cargo or rustup discovery and no longer dispatches the oracle through `init`. Cargo 1.91.0 selection happens inside `NewManager` while establishing manager-owned C0 state, uses the closed native standard rustup toolchain path derived from the OS user record and native GOOS/GOARCH, and performs no PATH, environment, shell, or child-process lookup.
3. The selected Cargo registration is stored in the sealed `managerState`. Vendor and metadata staging, RUSTC binding, and immediate toolchain rechecks consume that exact registration rather than rediscovering authority.
4. A built-production-binary regression successfully derives `rust-git-projection-v1` with a hostile fake `rustup` as the only PATH entry and proves its start marker remains absent. Hidden Rust and Go worker modes remain absent from the user-visible CLI command surface.
5. A test-only explicit composition root dispatches the same Rust worker for manager tests; production code exposes no request-controlled runner, executor, tool path, projection, or receipt seam.

## Validation evidence

- `go test ./cmd/curator -run 'TestProductionBinaryDispatchesRustOracleBeforeAmbientCargoDiscovery|TestHiddenWorkerModeIsNotAUserVisibleCommand' -count=1` — exit 0.
- `go test -count=1 ./internal/rustsource ./internal/closureexec` — exit 0 (`rustsource` 19.010s, `closureexec` 3.138s).
- `go vet ./internal/rustsource ./internal/closureexec ./cmd/curator` — exit 0.
- `go test -race -count=1 ./internal/rustsource` — exit 0 (`rustsource` 31.160s).
- `go build ./cmd/curator` — exit 0. The generated root binary was removed after validation.
- `go list -deps ./cmd/curator | rg '^github.com/relux-works/curator/internal/rustsource$'` — exit 0 and exact dependency present.
- `git diff --check` — exit 0.
- `go test -count=1 -timeout 30m ./...` — exit 0. Notable packages: `cmd/curator` 400.506s, `internal/artifactpolicy` 131.808s, `internal/rustsource` 36.800s, `internal/closureexec` 27.526s, `internal/install` 110.401s, `internal/install/atomicity` 114.345s.
- `golangci-lint run ./internal/rustsource/... ./internal/closureexec/... ./cmd/curator/...` — exit 127 because `golangci-lint` is not installed. This is not reported as green; `go vet` and the repository test/build gates above passed.

## Scope note

The worktree contains extensive pre-existing changes from the broader source-closure delivery. This rework preserved those changes and touched only the Rust manager/toolchain/worker wiring plus focused `cmd/curator` and Rust test coverage.
