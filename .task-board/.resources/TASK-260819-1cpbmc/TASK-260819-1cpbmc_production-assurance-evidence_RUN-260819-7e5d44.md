# TASK-260819-1cpbmc production assurance integration evidence

Run: `RUN-260819-7e5d44`

## Implementation outcome

- The CLI-selected `BuildAuthority` now remains the single operation authority
  across local and repository cache lookup/adoption, toolchain probing and final
  verification, compiler dispatch, receipt validation, and publication.
- Portable mode adapts the Curator-owned manager/worker build session and emits
  `portable-build-session-receipt-v1` evidence limited to the controls Curator
  actually establishes. It does not claim lossless host process, read, write, or
  network observation.
- Verified mode preserves the platform-neutral provider seam and revalidates the
  exact provider identity and canonical capability evidence at every authority
  boundary. Missing, unhealthy, identity-drifted, or capability-drifted providers
  fail before cache adoption or process start, without portable fallback.
- Local and repository protected caches use
  `curator-assured-build-cache-input-v1`, binding the logical build input to the
  full assurance binding. Portable, legacy, cross-mode, cross-provider, and
  capability-drifted entries cannot alias. Repository entries also persist and
  validate the exact build-session receipt.
- Marker currentness derives the same typed protected cache identity while
  continuing to validate the logical go-v1 build receipt.
- Security-negative coverage includes malformed/unknown configuration, missing
  provider, health/identity/capability drift, claim inflation, portable evidence
  presented as verified, legacy-address reuse, cross-mode/provider/capability
  cache reuse, and the repository pre-lookup TOCTOU boundary.

## Validation evidence

Commands were run directly and their real exit status was observed.

- Focused exact-tree packages:
  `go test -count=1 ./internal/closureexec ./internal/config ./internal/buildcache ./internal/buildrepo ./internal/install ./internal/marker` — exit 0.
- CLI plus focused integration:
  `go test -count=1 ./internal/closureexec ./internal/config ./internal/buildcache ./internal/buildrepo ./internal/install ./internal/marker ./cmd/curator` — exit 0; the real CLI toolchain path took 315.688 seconds.
- Mandatory repository TOCTOU negative:
  `go test -count=1 ./internal/buildrepo -run '^TestExternalPipelineAssuranceDriftPreventsCacheLookupAndCompile$'` — exit 0. The first authority check succeeds, the immediate pre-lookup check fails, and cache/compiler/artifact-store calls remain zero.
- Race: `go test -race -count=1 ./internal/closureexec` — exit 0.
- Repeated real portable runner:
  `go test -count=20 -run 'TestManagerProcessRunner' ./internal/closureexec` — exit 0.
- Compatibility:
  `go test -count=1 ./internal/closureexec ./internal/closuregraph ./internal/artifactpolicy ./internal/buildcache ./internal/godriver ./internal/buildsource` — exit 0.
- Windows compatibility compile:
  `GOOS=windows GOARCH=amd64 go test -c -o /tmp/TASK-260819-1cpbmc-closureexec-windows.test ./internal/closureexec` — exit 0.
- Linux compatibility compile:
  `GOOS=linux GOARCH=amd64 go test -c -o /tmp/TASK-260819-1cpbmc-closureexec-linux.test ./internal/closureexec` — exit 0.
- Authoritative exact-tree full Go rerun:
  `go test -timeout 30m ./...` — exit 0.
- Vet: `go vet ./...` — exit 0.
- Build: `go build ./...` — exit 0.
- Formatting: `test -z "$(gofmt -l cmd internal)"` — exit 0.
- Initial pinned lint after integration:
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` — exit 1 for staticcheck QF1003 in a new test. The test was rewritten as a tagged switch and its package test passed.
- Pinned lint on the handoff tree:
  `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run` — exit 0, `0 issues.`
- Canonical verifier:
  `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md` — exit 0; 53 labeled canonical records and references pass.
- Binary-deny focused gate over the six compiled detector/tree-denial tests — exit 0.
- The first Kotlin assertion targeted a non-authoritative research file and
  exited 1 because that file does not contain the scope sentence; no Kotlin
  implementation file was found. After materializing the authoritative epic
  scope through `task-board resource get`, the Kotlin exclusion gate (no
  `.kt`, `.kts`, `.gradle`, `.jar`, or `.class` files under `cmd`/`internal`,
  plus `Kotlin is explicitly deferred` in the authoritative scope) exited 0.
- `git diff --check` — exit 0.
- `task-board validate` — exit 0, `Board is valid. No issues found.`

No files were staged or committed. Existing accepted and unrelated worktree
changes were preserved.
