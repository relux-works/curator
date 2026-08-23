# TASK-260819-1cpbmc implementation evidence

## Delivered behavior

- Machine configuration now defaults `execution.mode` to `portable`; verified
  selection is explicit and requires provider id, version, canonical binary
  SHA-256, and trust evidence. Unknown modes, portable/provider aliasing,
  incomplete verified policy, and fallback-shaped unknown fields are rejected.
- `closureexec` has closed portable and verified assurance identities. Permits,
  receipts, derived cache records, and checkpoint identities bind mode, policy,
  execution policy, exact capability evidence, and provider contract/identity
  where applicable.
- Portable execution uses a manager-owned process-runner seam, rechecks the
  toolchain and immutable admitted inputs immediately before start, records
  exit status, independently walks and hashes the complete declared output
  root, rejects missing/extra/link/special/drifting outputs, and emits no
  fabricated process/read/write/network observation claims.
- Verified execution retains the platform-neutral authoritative provider seam.
  Construction rejects missing, incompatible, untrusted, or non-lossless
  providers. Permit commit performs fresh nonce-bound health/capability
  negotiation, and execution rechecks exact provider identity and capability
  receipt before the provider can start.
- Receipt and cache policy validators reject cross-mode and cross-provider
  reuse. Portable evidence cannot validate for verified policy, and checkpoint
  identities are domain-separated across assurance modes.
- Existing immutable intake, single-use causal permit, C4/C5 publication,
  protected-cache reconciliation, canonical derivation evidence, compiled
  artifact denial, and Kotlin exclusion remain intact.

## Security-negative coverage

Tests cover unknown mode, missing verified provider, portable/provider alias,
provider identity drift, capability drift, portable claim inflation, portable
receipt and cache reuse under verified policy, cross-mode checkpoint identity,
undeclared portable output, and zero provider starts on verified preflight
failure. Existing closure execution security tests continue to pass.

## Standalone gate evidence

The following authoritative commands exited 0:

- `go test -count=1 ./internal/closureexec ./internal/config`
- `go test -race -count=1 -cover ./internal/closureexec` — 69.7% coverage
- `go test -count=1 ./internal/closureexec ./internal/closuregraph ./internal/artifactpolicy ./internal/buildcache ./internal/godriver ./internal/buildsource`
- `go test -timeout 30m ./...`
- `go vet ./...`
- `go build ./...`
- `test -z "$(gofmt -l internal/closureexec internal/config)"`
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` — `0 issues.`
- `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md` — 53 canonical goldens and references pass
- focused compiled-artifact conformance/detector/tree denial test gate
- Kotlin-exclusion assertion: no Kotlin/Gradle implementation files and the
  authoritative scope still states Kotlin is explicitly deferred
- `git diff --check`
- `task-board validate` — no issues

## Evidence anomalies

An earlier standalone `go test ./...` exited 1 because `cmd/curator` exceeded
Go's default 10-minute package timeout while re-fingerprinting the host Go
toolchain. The stack was in `godriver.digestToolchainRecords`, unrelated to
this change; all reported internal packages, including `closureexec` and
`config`, had passed. The authoritative full rerun used the repository's
documented 30-minute gate timeout and exited 0. The pinned lint gate also
truthfully exited 1 twice during development for missing exported comments;
those findings were fixed and the final pinned run exited 0 with zero issues.

No files were staged or committed. Existing unrelated and accepted dirty
worktree changes were preserved.
