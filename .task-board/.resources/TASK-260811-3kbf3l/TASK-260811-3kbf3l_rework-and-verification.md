# TASK-260811-3kbf3l rework and verification

## Outcome

The Rust offline build adapter now routes fresh frozen Cargo metadata and build
through manager-owned `closureexec` permits and issued receipts. Portable mode
remains the functional default and records only its honest portable evidence
(`network=not-observed`, no claimed lossless process/read/write set). Verified
mode uses the injected lossless provider and rejects any process, read, write,
network, evidence, or protected-input mismatch before publication.

The compact C6 publication projection is created only after both issued Cargo
receipts validate. The selected executable bytes must match the build receipt's
typed output digest before they enter the protected store.

## Reviewer-required corrections

- Removed the direct `osCargoBuildRunner` / `exec.CommandContext` build path.
- Added committed build-scoped permits for both Cargo invocations, immediate
  full native-toolchain rechecks, admitted replay/work-copy roots, typed Cargo
  JSON and executable evidence, and executor-issued receipt verification.
- Extended the portable runner to escrow only declared evidence before cleaning
  mutable work copies. Undeclared files in the evidence root still reject.
- Preserved portable-default operation without inflating its assurance claims;
  verified mode remains separately provider-backed and fail-closed.
- Added observed start instrumentation at the sole portable process-start seam.
  Manager construction/tool registration records zero starts before C0, the
  portable build records exactly two permitted starts, cache hits record none,
  and missing permits record none.
- Added verified-provider negative regressions for attempted network, an extra
  child, undeclared read/write, evidence mismatch, protected-input mutation,
  missing permit, and verified configuration without a provider. Every case
  proves no protected publication receipt exists.
- Replaced Cargo release inference from a user-writable toolchain directory
  label with an exact operator-approved Cargo executable SHA-256 binding. The
  approved `aarch64-apple-darwin` Cargo 1.91.0 descriptor binds
  `sha256:0da859e1130e00a81dac84fa1e86a3dbdd968ddfccef627a8d37255fcbb39e78`
  to implementation commit `ea2d97820c16195b0ca3fadb4319fe512c199a43`;
  other executable bytes or unregistered native targets fail closed.

## Verification evidence

- `go test ./internal/closureexec ./internal/rustsource -count=1`: exit 0.
- `go test -race ./internal/rustsource ./internal/closureexec -count=1`: exit 0.
- `make lint` with the repository CI-pinned golangci-lint v2.12.2: exit 0,
  zero issues. An earlier pre-correction run exited 2 with eight local findings;
  those findings were corrected before the green run.
- `go vet ./...`: exit 0.
- `go build ./...`: exit 0.
- Final `go test -count=1 ./...`: exit 0. `internal/rustsource` passed in
  93.585s and `internal/closureexec` in 36.499s.
- A preceding full-suite run exited 1 only because
  `TestStrictRegistryPolicyFailsUnknown` considered its fixture timestamp too
  far in the future. The exact isolated test rerun exited 0, and the subsequent
  authoritative full-suite rerun exited 0.
- `git diff --check`: exit 0.
- `task-board validate`: exit 0 (`Board is valid. No issues found.`).

## Important implementation note

The portable detailed derivation receipt deliberately does not claim OS-level
network or complete process/read/write observation. The verified negative
suite exercises those lossless guarantees through the provider contract. This
keeps cache and publication identities honest while retaining the project-wide
portable default required by the active architecture.
