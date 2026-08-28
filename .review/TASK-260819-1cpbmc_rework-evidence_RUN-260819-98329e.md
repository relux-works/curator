# TASK-260819-1cpbmc rework evidence

Run: `RUN-260819-98329e`

## Reviewer findings addressed

- R1: `cmd/curator/assurance.go` is the single production machine-config to
  assurance boundary. Project/global install and status paths preflight it
  before build-cache planning or compiler work. Portable is the omitted-mode
  default; unknown configuration fails parsing; verified selection has no
  shipped provider and fails with `verified_provider_missing` before process
  dispatch. Command tests cover all three cases.
- R2: `ProtectedStore.Publish` and `ProtectedStore.Inspect` now require
  `AssuredCacheInput`. The protected filename and persisted entry bind mode,
  policy, execution policy, exact capability evidence, provider contract,
  provider identity/binary/trust evidence, fresh capability receipt, and the
  independent closure cache input. Portable, cross-provider, and fresh-receipt
  drift lookups miss.
- R3: `Executor.Preflight` creates `AssuredOperation`; its exact verified
  capability receipt derives both cache input and committed permit. Unhealthy
  negotiation returns `verified_provider_unavailable`, with zero cache
  adoption and zero process starts.
- R4: `ManagerProcessRunner` now prepares immutable replay copies, checks the
  executable digest, requires an exact permit-bound output root, derives a
  deadline and combined-output budget from resource limits, verifies replay
  identity after execution, and terminates the portable process group on Unix.
  Executor rechecks original admitted inputs after execution and verifies the
  complete declared output set and output byte limit. Receipts still claim no
  lossless process/read/write/network observation. Real-process tests cover
  replay mutation, output binding, undeclared output, output overflow, timeout,
  descendant cleanup, nonzero exit, and honest evidence.
- R5: every present execution/provider field rejects null and wrong types in
  both modes. Provider id, version, digest, and trust evidence have closed
  validation; malformed portable fields cannot be erased into absence.

## Validation evidence

Every command below was run directly as a standalone gate unless explicitly
identified as a development-red run.

- `go test -count=1 ./internal/closureexec ./internal/config ./cmd/curator` —
  exit 0. The command package's host-toolchain fingerprint took 317 seconds.
- `go test -race -count=1 ./internal/closureexec` — exit 0.
- `go test -count=20 -run 'TestManagerProcessRunner' ./internal/closureexec` —
  exit 0.
- `go test -count=1 ./internal/closureexec ./internal/closuregraph ./internal/artifactpolicy ./internal/buildcache ./internal/godriver ./internal/buildsource` — exit 0.
- `GOOS=windows GOARCH=amd64 go test -c -o /tmp/TASK-260819-1cpbmc-closureexec-windows.test ./internal/closureexec` — exit 0.
- `GOOS=linux GOARCH=amd64 go test -c -o /tmp/TASK-260819-1cpbmc-closureexec-linux.test ./internal/closureexec` — exit 0.
- First `go test -timeout 30m ./...` development run — exit 1. Under full-suite
  load, two new real-process tests exceeded short wall-clock assumptions.
  These were genuine test failures and were fixed with a startup handshake and
  cancellation-driven timeout test.
- Authoritative `go test -timeout 30m ./...` rerun — exit 0.
- `go vet ./...` — exit 0.
- `go build ./...` — exit 0.
- `test -z "$(gofmt -l cmd/curator internal/closureexec internal/config)"` — exit 0.
- Initial post-rework pinned lint run — exit 1 for a `revive`
  context-parameter-order finding in a test helper. The helper was fixed.
- Final `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` — exit 0, `0 issues.`
- `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md` — exit 0; 53 canonical goldens and references pass.
- Focused compiled-artifact detector/tree denial gate over the six named
  `internal/artifactpolicy` tests — exit 0.
- Kotlin exclusion gate (no Kotlin/Gradle/JAR/class implementation files and
  authoritative scope contains `Kotlin is explicitly deferred`) — exit 0.
- `git diff --check` — exit 0.
- `task-board validate` — exit 0.

No files were staged or committed. Existing accepted and unrelated dirty
worktree state was preserved.
