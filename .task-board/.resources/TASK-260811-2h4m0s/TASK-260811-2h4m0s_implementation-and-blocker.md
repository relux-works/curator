# TASK-260811-2h4m0s implementation evidence and blocking constraint

Date: 2026-08-19

## Implemented work

- Added `internal/rustsource` with closed Cargo manifest and Cargo.lock v3/v4
  parsing, selection-neutral lock capture, immutable registry/Git/path origin
  validation, target/feature selection binding, metadata reconciliation, and
  stable Rust diagnostics.
- Added admission-before-transform ordering through the shared artifact policy.
  Registry archives/index records and complete Git/path trees must all admit
  before registry/Git transform parsing or any Cargo permit is committed.
- Added the pinned `cargo-vendor-transform-v1` registry and Git per-leaf
  mappings, canonical checksum bytes, unique lock-to-directory verification,
  containment checks, post-vendor admission, and tamper rejection.
- Added C0 Cargo toolchain validation and immediate rechecks, absent vendor and
  private-home requirements, complete source-replacement config derivation,
  exact config/home/environment permit binding, and metadata permit parity.
- Closed nested Cargo metadata parsing and added malformed-input no-panic
  coverage, full lowercase 40-hex Git commit validation, ambient/profile Cargo
  config rejection, and exact Git source-key normalization without the precise
  commit fragment.
- Added `closureexec.Executor.VerifyIssuedDerivationReceipt` so a receipt can
  be checked against the executor causal chain that actually issued it.
- Removed all public Git projection/normalizer fields and the public binding
  API. External callers cannot inject selected paths, normalized manifest
  bytes, a receipt, or an executor through `CaptureRequest`. Forged selection
  and normalized-manifest mutations are rejected by an internal canonical seal.
- Fixed directory admission error handling: the first shared classifier error
  is no longer swallowed. Compiled Git/path members retain
  `artifact_compiled_dependency_forbidden`, cause zero Cargo activity, and
  leave the vendor destination absent.

## Blocking authority constraint

The secure negative boundary is implemented, but a production-reachable
positive Git capture path cannot be completed without an authority owner that
does not currently exist in the repository API.

The accepted contract requires the pinned Cargo 0.92 Git package projection
and manifest normalization to be derived by a trusted capture manager after
admission, under an exact permit binding commit, tree, leaves, include rules,
package path, executable/toolchain, argv, and outputs. `closureexec` currently
exposes executor/provider construction to its caller. If `rustsource` accepts
that executor/provider (or a projection runner) from the same requestor whose
bytes are being checked, that requestor can issue a self-consistent receipt for
arbitrary selected paths and normalized bytes. The receipt is causal but does
not prove independent manager authority.

The current tree therefore fails closed: only package-private sealed
derivation state can reach the transform, and an external-package regression
proves an ordinary caller gets `rust_git_identity_invalid`, zero permit
commits/spawns, and no destination. A same-package test exercises the intended
positive transform, but this is deliberately not claimed as a production
entry point.

### Rejected attempts

1. Public projection/normalizer fields plus a string receipt ID: caller
   forgeable.
2. Public `BindGitDerivation(executor, receipt, bytes)`: a caller-created
   executor/provider can issue the matching receipt.
3. A caller-injected projection runner or test-only private map as the
   production entry point: merely moves the same authority substitution and
   leaves a dead or forgeable positive path.

### Viable options

1. **Recommended:** extend the shared protected-execution layer with an
   operator-owned capture-manager capability/handle that request callers
   cannot construct or substitute. Rust can then accept that handle from the
   trusted application composition root, commit the fixed Cargo derivation,
   verify the exact receipt/output, and seal the private transform input.
2. Implement Cargo 0.92 Git projection and normalized-manifest generation
   entirely inside `rustsource` from admitted Git/index/workspace bytes. This
   avoids an execution authority but is a substantial Cargo-behavior port and
   needs the retained GV01-GV03e oracle corpus as executable fixtures.
3. Explicitly narrow `rust-source-v1` to registry/path sources for this cycle
   and move Git support to a separately approved capability. This conflicts
   with the current task acceptance criteria and requires a product decision.

Required decision: choose the shared operator-owned authority seam (option 1),
approve the full internal Cargo behavior port and fixture budget (option 2),
or approve a scope change excluding Git (option 3).

## Validation evidence

Commands were run directly as standalone processes unless noted.

- `go test -count=1 ./internal/rustsource ./internal/closureexec` — exit 0.
- `go test -race -count=1 ./internal/rustsource ./internal/closureexec` — exit 0
  before the final classifier-error patch; the final patch has focused
  `internal/rustsource` coverage below.
- `go test -count=1 ./internal/rustsource` — exit 0 after the final
  classifier-error patch.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run ./internal/rustsource ./internal/closureexec`
  — exit 0, 0 issues, after the final patch.
- `go vet ./...` — exit 0 after the final patch.
- `go build ./...` — exit 0 after the final patch.
- `git diff --check -- internal/rustsource internal/closureexec/executor.go internal/closureexec/closureexec_test.go`
  — exit 0 after the final patch.
- `go test -count=1 -timeout 30m ./...` — exit 0 before the final
  classifier-error patch; notable packages: `cmd/curator` 349.880s,
  `internal/artifactpolicy` 126.539s, `internal/install` 103.780s,
  `internal/install/atomicity` 110.583s, `internal/rustsource` 3.140s.
- `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md`
  — exit 0; 53 labeled records and all references passed.
- Direct `golangci-lint` binary invocation — exit 127 because the binary is
  not installed; the pinned `go run ...@v2.12.2` invocation above is green.

The repository-wide test and race commands were not rerun after the final
classifier-error patch because the task is blocked on the authority ownership
decision; focused tests, lint, vet, build, and diff validation cover that patch.
