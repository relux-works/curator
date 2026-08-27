# TASK-260720-jrrgw9 — independent verifier 4 results

Date: 2026-07-29  
Role: tester  
Candidate: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`  
Immutable conformance root: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`

## Result

All required executable macOS gates passed against the exact integrated
candidate. The exact full race suite passed, and
`internal/install/atomicity` completed in 115.687 seconds, 364.313 seconds
inside the established 480-second acceptance bar.

Native Windows qualification was attempted only after both macOS gates were
green. `ssh win` could not connect to `100.120.84.42:22` and exited 255 after
the configured 15-second connection timeout. No remote directory or remote
state was created. Windows is therefore externally unqualified, not emulated
and not claimed as passing.

## Candidate and immutable-input provenance

- Candidate HEAD: `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`.
- Go toolchain: `go version go1.25.5 darwin/arm64`.
- The live pre-test manifest had 359 entries and was byte-identical to the
  accepted integration `manifest-post.txt`; both SHA-256 values are
  `83b5df8b81ec21f472a90ad082d1ab3464b807968418a7d3d82c8672ff6a2819`.
- All 15 integrated file hashes matched `target15-post.txt`.
- `internal/install/atomicity/fixture_test.go` remained unchanged at
  `e0732e2e3df9adee95321ba28723a878699722747cb231a5309902a56f1f6120`.
- The rejected cross-save state/cache symbols were absent:
  `namespaceGraphAccepted`, `acceptNamespaceGraph`,
  `forgetNamespaceGraph`, `namespaceChecked`, `namespaceGraph`, and
  `namespaceMu`.
- No file was staged.
- The immutable rc.5 digests matched before and after testing:
  - `vectors/build-drivers.json`:
    `f412c107091cf82f980523afe5361212a3b89a3425f5d885373191f8acb12aea`
  - `vectors/manager-lifecycle.json`:
    `2ddbd2665a63f44dc0e03e060f4cd34bfde219a56b3192511fe1ef81047feedf`
  - `vectors/external-repository-lifecycle.json`:
    `175d709f183775b22ed3db27bc923ca78d6394d93a13096b5d6890c509aab072`
  - `manifest.json`:
    `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`

## Required gate ledger

Each Go gate was run as a standalone process, sequentially, with `-count=1`,
the immutable conformance root, a distinct task-owned `GOTMPDIR`, no timeout
override, and no output pipe.

| Gate | Real exit | Wall time | Evidence |
| --- | ---: | ---: | --- |
| Focused preflight: two stable empty process scans and >=20 GiB free | 0 | 2 s | 22,573,444 KB free |
| Focused authoritative 12-package consumer barrier | 0 | 15.26 s | all 12 selected packages `ok` |
| Full preflight: two stable empty process scans and >=20 GiB free | 0 | 2 s | 22,558,280 KB free |
| `CURATOR_CONFORMANCE_ROOT=... go test -count=1 ./...` | 0 | 352.86 s | every package `ok` |
| Race preflight: two stable empty process scans and >=20 GiB free | 0 | 2 s | 22,455,408 KB free |
| `CURATOR_CONFORMANCE_ROOT=... go test -count=1 -race ./...` | 0 | 441.11 s | every package `ok`; no race diagnostic |
| Race-log failure/race diagnostic absence assertion | 0 | <1 s | no `WARNING: DATA RACE`, `DATA RACE`, or `FAIL` |
| Native Windows connection/inventory attempt | **255** | 14.89 s | externally unqualified: SSH connection timed out |
| Final two-scan process/disk check | 0 | 2 s | empty; 22,374,820 KB free |

The focused command used the accepted 12-package list:
`internal/runtimestore`, `internal/install`, `internal/scopes`, `cmd/curator`,
`internal/buildcache`, `internal/buildsource`, `internal/buildmeta`,
`internal/godriver`, `internal/skillcheck`, `internal/skillspec`,
`internal/whitelist`, and `internal/interop`, with the exact 26-test accepted
authoritative filter recorded in `focused-authoritative.log`.

## Package timing

| Package | Full | Race |
| --- | ---: | ---: |
| `cmd/curator` | 352.426 s | 439.622 s |
| `internal/install` | 120.107 s | 126.503 s |
| `internal/install/atomicity` | 109.614 s | **115.687 s** |
| `internal/godriver` | 78.690 s | 98.540 s |
| `internal/transaction` | 52.704 s | 40.881 s |

The race atomicity package is 364.313 seconds below the required 480-second
bar and 488.014 seconds faster than verifier 3's timed-out package result.

## Validation-command exits

All source/provenance validations exited 0:

- live manifest generation; accepted-manifest comparison;
- accepted 15-path hash verification;
- unchanged fixture hash assertion;
- rejected-symbol absence assertion;
- HEAD and staged-index checks;
- immutable vector digest assertions;
- post-run manifest generation and comparisons against both the accepted
  manifest and the pre-run manifest;
- post-run 15-path and fixture hash assertions;
- post-run staged-index and immutable-vector assertions.

The only non-zero validation/qualification command was the truthful Windows
SSH connection attempt (exit 255). It was an external availability failure,
not a candidate test failure.

## Post-run immutability and cleanup

- The post-run candidate manifest again had 359 entries and SHA-256
  `83b5df8b81ec21f472a90ad082d1ab3464b807968418a7d3d82c8672ff6a2819`.
- Pre-run, accepted post-integration, and post-run manifests are byte-identical.
- All three task-owned `GOTMPDIR` leaves measured 0 KB after test descendants
  stopped. The exact `focused`, `full`, and `race` leaves and their empty parent
  were removed with `rmdir`; absence verification exited 0.
- Shared Go caches were not cleared. No source/test/configuration file was
  edited, staged, committed, stashed, checked out, pinned, or published.
- No Windows remote cleanup was needed because connection failed before any
  remote state was created.

## Raw evidence

Evidence archive:
`TASK-260720-jrrgw9_verifier4-evidence.tgz`

Archive SHA-256:
`2e46e3f5b7fb717ca003653df344a9fc03e0bb35349abb02b3e7e8c01aed976c`

Principal raw-log SHA-256 values:

- focused authoritative:
  `9acca4a0857534c84a4902fba4e2ece635f8abf948bd559dec29955058a059d7`
- full macOS:
  `3359fa0e28d2ceb9756a033f15fd591d89a5ba69a846e2721c48331c20c04d2f`
- full race:
  `4fa9c7c787fa8b86885027cd3b63f8c8aa6278ae9b7e449128d6035a69a257b2`
- Windows connection failure:
  `9dd4a56cd9db796205e59dd984d1477bdd0351d781da4aebdf40e9b666c7e986`

The archive also contains all three preflight logs, the final process/disk
log, and the byte-identical pre/post candidate manifests.
