# TASK-260827-18tswm — macOS CI residual outcome

## Failure A — approved Cargo descriptor, absent pinned toolchain

- Root cause: the hosted arm64 macOS runner has an approved
  `aarch64-apple-darwin` descriptor, but the expected
  `~/.rustup/toolchains/1.91.0-aarch64-apple-darwin` root/executable is absent.
- Disposition: classified host-capability skip.
- Exact reason: `pinned Cargo toolchain root or executable unavailable for native target aarch64-apple-darwin`.
- Class: `host-capability`.
- Scope bound: the precheck returns that reason only for `fs.ErrNotExist` on
  the expected root or executable. Present malformed roots, read failures,
  canonicalization failures, descriptor mismatches, and executable byte
  mismatches remain fatal through `registerCargoAtC0`; no descriptor, digest,
  or toolchain identity was added.
- Crossconformance: uses the same Rust-unavailable path and its existing closed
  six manager-obligation gaps. `artifact.shared_admission/rust` remains
  mandatory.
- Local verification: the full native arm64 `internal/rustsource` suite ran
  rather than skipping (95.165s, exit 0); native `internal/crossconformance`
  passed (exit 0). An amd64 Rosetta build skipped the manager path for the
  existing no-approved-descriptor reason and kept shared admission plus the
  closed completeness gate green (exit 0).
- CI inference: absence of the pinned root comes from the supplied run
  33037369236 error, which names the missing exact path.

## Failure B — rc.9 dry-run reported `toolchain-unavailable`

- Root cause disposition: CI-observed transient trusted-Go probe failure; not
  classified and no published expectation or assertion changed.
- Evidence against a persistent host limitation: the same captured macOS
  stream records `TestFingerprintTraversalMatchesLegacyOnRealToolchain`
  successfully fingerprinting `/Users/runner/hostedtoolcache/go/1.25.5/arm64`
  (16,092 records) immediately before `internal/install`. Therefore the runner
  did have a resolvable real Go toolchain.
- Merge comparison: `origin/main` has the same rc.9 test and the same
  `toolchainInventory` selection. Delivery assurance changes cache-key and
  authority threading, but after a successful `Toolchain.Probe` they do not
  select `toolchain-unavailable`; that outcome is produced only by a probe
  error.
- Local verification: with the exact committed rc.9 spec pin
  `0ed5c691e9208eea52f21db2fc05e226ce3516fd`, the exact
  `compiled-cache-miss-is-read-only` case passed once and then three consecutive
  times (4/4 total, all exit 0). No skip was added because the supplied evidence
  does not prove the existing `no trusted Go toolchain is resolvable here`
  condition.
- Remote boundary: a fresh matrix rerun and passing run URL are owned by the
  landing Orchestrator and were not claimed here.

## Validation executed by this run

| Command | Exit | Result |
| --- | ---: | --- |
| `go test ./internal/rustsource -run '^TestCargoHostCapabilityReasonClassifiesOnlyAbsence$' -count=1 -v` | 0 | Negative absence/mismatch bounds pass |
| `go test ./internal/rustsource -count=1` | 0 | Full native arm64 package pass; no new skip |
| Native arm64 `TestRustConformanceR03R05R06R07PathWorkspaceBuild` | 0 | Explicit `PASS`, not `SKIP` |
| `go test ./internal/crossconformance -count=1` | 0 | Full native package pass |
| rc.9 install case, `-count=1` | 0 | Exact case pass |
| rc.9 install case, `-count=3` | 0 | Three consecutive exact-case passes |
| `GOARCH=amd64 go test ./internal/rustsource -count=1` | 0 | Rosetta package pass |
| `GOARCH=amd64 go test ./internal/crossconformance -run '^TestCrossAdapterConformance$' -count=1 -v` | 0 | Existing descriptor-unavailable degradation remains closed |
| platform-case replay, first captured macOS stream | 0 | 18 classified skips, zero `UNCLASSIFIED`/`FATAL` |
| platform-case replay, run 33037369236 stream | 0 | 20 classified skips, zero `UNCLASSIFIED`/`FATAL` |
| platform-case focused replay, new Rust reason | 0 | One `host-capability` skip, zero `UNCLASSIFIED`/`FATAL` |
| `go build ./...` | 0 | Build passes |
| `go vet ./...` | 0 | Vet passes |
| `gofmt -l cmd internal` | 0 | No output |
| `golangci-lint run` (v2.12.2) | 0 | 0 issues |
| `git diff --check` | 0 | Clean |

No files were staged, committed, pushed, reset, or cleaned.

## Lifecycle handoff boundary

The task was moved to `to-review`. `task-board handoff ... --role developer`
then refused because checklist items 6 (fresh full remote CI matrix plus URL and
artifacts) and 13 (overall AC) remain unchecked. This run was explicitly
instructed not to check the remote-CI item; the landing Orchestrator must push,
obtain that evidence, check those items, and retry the handoff. Marking either
item green here would violate the evidence-honesty contract.
