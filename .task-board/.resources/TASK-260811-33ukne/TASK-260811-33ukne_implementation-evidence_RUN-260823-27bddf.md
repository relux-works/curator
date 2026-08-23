# TASK-260811-33ukne implementation evidence

Run: `RUN-260823-27bddf`

## Delivered boundary

- Added canonical shared `source-acquisition-permit-v1` and
  `source-acquisition-receipt-v1` records with single-use causal execution,
  exact C0 tool rechecks, honest portable capability evidence, and explicit
  verified-provider failure without fallback.
- Replaced SwiftPM's adapter-owned Git process authority with the shared
  acquisition executor. Production SwiftPM source contains no direct
  `exec.Command*` call or custom Git permit/receipt type.
- Separated the trusted read-only executable root from writable task execution
  and quarantine roots in both shared process runners. Production Git
  acquisition and admitted-mirror verification now launch the actual C0-bound
  Git executable, not a hashed shell wrapper. Additional Git process-family
  paths and byte digests are bound into the C0 aggregate fingerprint, declared
  in permits, and rechecked immediately before every start.
- Enforced symlink-aware containment of the C0 Git executable below its
  declared root before a process can start.
- Preserved executable bits in admitted trees and synthesized a deterministic
  minimal shallow bare Git repository from an admitted source tree plus an
  admitted exact commit object. The offline derivation permit has
  `network=none`, declares the retained local output, and preserves the pinned
  commit and Git tree.
- Added a narrow artifact-policy issuer for `source-control-mirror-v1`. It
  verifies the issued acquisition receipt, admitted source receipt, issued
  derivation permit/receipt, typed local output, transform evidence, same source
  kind, exact revision/tree, and mirror inventory. It cannot authorize an
  arbitrary Git store or verified binary.
- Re-admitted the authorized mirror into protected storage and verified its
  object closure through an ordinary admitted-input, network-none derivation
  using the exact C0 Git tool. Offline replay rechecks the opaque authorization
  before this process; tampering starts zero Git processes.
- Added acquisition, commit-evidence intake, mirror derivation, mirror intake,
  and verification permit/receipt identities to C1/C2/C3 evidence. Generated
  and supplied lock paths retain their complete acquisition/intake/manifest/
  mirror chain.
- Updated the Rust cross-adapter process-boundary guard to recognize exactly
  the two shared manager-owned process seams (`portable_runner.go` and
  `acquisition.go`) while retaining its ban on adapter-level process launches.

## Tests added or strengthened

- Canonical acquisition permit/receipt round trip and portable evidence.
- Stale permit and C0 tool drift reject before process start.
- Explicit verified acquisition has no portable fallback.
- Real SwiftPM/Git acquisition-to-mirror chain, executable-bit preservation,
  admitted offline verification, C1/C3 receipt reachability, and tampered
  authorization zero-start behavior.
- Relative and symlink C0 Git escapes reject before recheck or process start.
- Git process-family byte drift rejects before any primary or child process;
  the real integration fixture observes and asserts the exact absolute C0 Git
  launch path for acquisition and offline verification.
- Existing R01-R13, P01-P08, CGP05, CGP11, CGN zero-start, binary-deny, and
  offline replay vectors remain green.

## Standalone validation evidence

Every terminal gate below was run as a standalone process. No gate was piped
through `tee`.

- Focused post-fix test:
  `go test -count=1 ./internal/closureexec ./internal/artifactpolicy ./internal/swiftpmsource ./internal/rustsource` — exit 0.
- Race:
  `go test -race -count=1 ./internal/closureexec ./internal/swiftpmsource` — exit 0.
- Full repository first run — exit 1 because the Rust cross-adapter source scan
  did not yet recognize the new shared acquisition seam. The guard was fixed;
  this is retained in `full-go-01.log` as a genuine development failure.
- Full repository second run — exit 1 because
  `TestGCRetainsAndReportsReferencedCompiledState` hit a nondeterministic
  cross-test helper-process failure (`unknown command
  __curator_rust_git_oracle_v1`). The exact isolated test then exited 0 without
  a code change. Both results are retained in `full-go-02.log` and
  `cmd-flake-check-01.log`.
- Authoritative full repository:
  `go test -timeout 30m -count=1 ./...` — exit 0 (`full-go-03.log`), including
  `cmd/curator` 631.153s, `artifactpolicy` 229.493s, `rustsource` 212.603s, and
  `swiftpmsource` 72.458s.
- `go vet ./...` — exit 0.
- `go build ./...` — exit 0.
- Initial pinned lint — exit 1 with seven local findings; all were fixed and
  retained in `lint-01.log`.
- `go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2 run`
  — exit 0, zero issues (`lint-02.log`).
- Canonical golden verifier — exit 0; 53 labeled records and every reference
  pass (`canonical-01.log`).
- Six focused compiled-artifact denial tests — exit 0
  (`binary-deny-01.log`).
- Kotlin implementation exclusion plus authoritative deferred-scope check —
  exit 0 (`kotlin-exclusion-01.log`).
- Gofmt cleanliness — exit 0.
- `git diff --check` — exit 0.
- `task-board validate` — exit 0; board valid.

The orchestrator's final nudge identified that the first macOS integration
fixture staged a shell wrapper which executed ambient Git. That evidence was
not used for handoff. The runner/tool-root and process-family changes above
removed the wrapper from the Git authority path. After that substantive fix:

- exact-launch plus family-drift focused vectors — exit 0;
- focused closureexec/SwiftPM/Rust compatibility — exit 0;
- race — exit 0;
- authoritative `go test -timeout 30m -count=1 ./...` rerun
  (`full-go-04.log`) — exit 0;
- fresh vet, build, pinned lint, canonical verifier, binary-deny, Kotlin
  exclusion, gofmt, diff, and board validation — exit 0.

The final comment/nil-check lint fixes were followed by fresh focused, race,
vet, and build gates, all exit 0. No files were staged or committed. Existing
unrelated dirty-worktree changes were preserved.
