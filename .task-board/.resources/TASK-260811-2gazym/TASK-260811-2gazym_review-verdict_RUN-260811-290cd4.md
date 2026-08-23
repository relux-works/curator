# Reviewer verdict for TASK-260811-2gazym

Verdict: **changes requested -> to-dev**

This artifact records exactly one reviewer branch. The findings are ordinary,
autonomously repairable implementation work; they are not a Stop-The-Line
boundary.

## Goal and reviewed scope

- Reviewer run: `RUN-260811-290cd4`
- Authoritative reviewer goal immediately before verdict:
  `GOAL-260811-44ada4` revision 1
- Authoritative objective: review board scope `TASK-260811-2gazym` until
  exactly one evidence-backed verdict is persisted
- Resolved scope: `TASK-260811-2gazym`
- Reviewed producer run: `RUN-260811-04961c`
- Reviewed producer evidence:
  `TASK-260811-2gazym_rework-evidence_RUN-260811-04961c.md`
- Reviewed evidence SHA-256:
  `6a9dddb119777fa593bfbe72720f63e3b155fdfefff2f0bf8695297123f2b587`
- Prior verdict:
  `TASK-260811-2gazym_review-verdict_RUN-260811-0f5cd8.md`
- Prior verdict SHA-256:
  `fbfb103cb1505693b9aa39bba253db2552137eb225f24e197424d3e81ddc17c0`
- Reviewed Git revision:
  `8ff4a238f7725bada3cfb8aa7c9c135698483caa`
- Artifactpolicy source fingerprint:
  `d9a811e9f5ffadfcd4fda90b6a13b69ad3ba5a7a068e70fb1d12cd46a0bf7e1d`
- Repository 362-file Go/module fingerprint:
  `b6bb88c574f46754fba91e588a8c2215be68fb14f465acd5c37233b4833c204e`

## Acceptance blockers

### R1 - A07 has no production-usable manager-owned authority

The dangerous caller-configured `ManagerVerifier` path is correctly gone.
The resulting public contract, however, cannot produce the positive A07
manager-selected toolchain admission required by this task.

- `ToolchainAuthorization` is sealed by a package-private method and record
  type, so an external manager package cannot implement it
  (`types.go:325-333`).
- The only concrete authorization and seal are package-private
  (`types.go:615-642`). `NewService` explicitly cannot manufacture either
  trust role and no production issuer exists (`types.go:676-682`).
- The reusable A07 branch succeeds only through
  `validToolchainAuthorization` in a `_test.go` file
  (`conformance_test.go:114-118`,
  `test_helpers_test.go:67-95`). That helper inserts a synthetic root,
  fingerprint, selector, and the package-private seal; it does not select or
  fingerprint a real toolchain.
- The external-package regression proves the safe negative branch only: an
  arbitrary caller root cannot start an adapter. It does not provide a usable
  positive manager integration.

This is not the same as A08. The downstream protected-execution task
`TASK-260811-27xisf` owns causal output receipts, so production positive
local-output authority should remain unavailable until that evidence exists.
But this task explicitly requires A07, trust-role enforcement, and
pre-execution admission APIs. A package-internal fabricated record does not
prove the current positive criterion, and the downstream task cannot consume
the present sealed interface without reopening this package.

Required rework: add a narrow manager-owned, non-adapter-configurable toolchain
selection/fingerprint/recheck boundary that returns the opaque authorization.
It must derive evidence from a closed central selector and the actual captured
dependency boundary, never from caller-populated roots or boolean assertions.
Add an external-package positive integration for a real selected/fingerprinted
toolchain while retaining the arbitrary-root zero-start negative.

### R2 - metadata path limits still occur after payload allocation

The leaf and emitted-byte repair is real, but the directive required applicable
declared path bounds before metadata payload read/allocation as well.

- PAX local/global and GNU long-name/link records run only
  `preflightLeaf` and `preflightEmitted`, then allocate and read the entire
  declared payload with `readExactAt` (`containers.go:548-575`). PAX path
  records and GNU long-name values are validated only after that read.
- GNU ar `//` string tables are likewise allocated in full at
  `containers.go:1450-1459`; individual names are resolved and path-validated
  only later. Unreferenced unsafe table names are not validated at all.
- BSD `#1/<len>` checks only `nameLength > MaxPathBytes`
  (`containers.go:1395-1413`). The policy limit applies to the full canonical
  `container!/member` path, so a name that fits 4096 bytes but overflows once
  the container prefix is added is read first at `containers.go:1415` and
  `containers.go:1720), then rejected later.
- The focused regressions cover custom leaf/emitted limits and sparse 257 MiB
  leaf refusal. They do not cover PAX/GNU/ar name values that are below the leaf
  limit but exceed the remaining canonical path budget.

Required rework: parse PAX and native-name metadata with bounded readers,
preflight record/name lengths against the remaining full virtual-path budget
before materializing their values, validate every GNU string-table name
deterministically, and add local/global PAX, GNU long-name/link, GNU ar string
table, and BSD extended-name path-precedence cases.

### R3 - successful BSD extended-name reads are not charged

The BSD branch preflights `nameLength`, then `resolveARName` reads those
bytes and subtracts them from `dataSize` (`containers.go:1715-1725`).
Afterward only the remaining member content reaches `checkLeaf` and
`addEmitted` (`containers.go:1423-1495`). The successfully read extended
name is therefore absent from actual emitted-byte accounting, contrary to the
rule that declared sizes are only early refusal and actual reads are
authoritative.

Required rework: charge successful extended-name reads exactly once, preserve
the metadata/member evidence needed to rederive the accounting, and add a
successful bounded BSD-name regression plus forged-manifest accounting
negatives.

## Verified repairs and preserved behavior

- No exported caller-configured trust issuer, record-to-capability factory, or
  staging verifier remains.
- The authorization interfaces are opaque to external packages; caller-created
  roots and copied pre-existing objects reject before execution/publication.
- T04 now uses the public receiptless `AdmitLocalOutput` fail-closed path.
- No production positive local-output receipt is fabricated.
- Compiled classes remain globally denied for `dependency_input`, and
  `verified-binary-v1` remains unavailable.
- Kotlin is absent from implementation scope.
- The exact corpus source digest is
  `87a5cb6afb1c120cf75979cccd57fe2702c9a7dd74bee22dfa80418e1f26750e`.
  All 182 unique cases passed with
  `A=14, C=91, F=61, T=15, V=1`.
- The producer full test/vet/build/pinned-lint evidence is source-stable at the
  independently reproduced 362-file fingerprint above.

## Independent validation

| Gate | Result |
| --- | --- |
| Focused R1/R2/exact-corpus command | pass, 0.942s |
| Exact reusable corpus with independent enumeration | pass, 182 unique cases; A14/C91/F61/T15/V1 |
| `go test -count=1 ./internal/artifactpolicy/...` | pass, 30.168s |
| Targeted `go test -race` over R1/R2 and exact corpus | pass, 5.885s |
| `go vet ./internal/artifactpolicy/...` | pass |
| `go build ./internal/artifactpolicy/...` | pass |
| pinned golangci-lint v2.12.2 | pass, `0 issues.` |
| `gofmt -l internal/artifactpolicy` | pass, no files listed |
| `git diff --check` | pass |
| `task-board validate` | pass |

A repository-wide rerun was intentionally not duplicated: the operator
directive identified the producer full lane as source-stable, and both its
artifactpolicy and 362-file fingerprints were independently reproduced.
Green mechanical gates do not cover the three semantic gaps above.

## Route

Route `TASK-260811-2gazym` to `to-dev`. After the A07 boundary and bounded
metadata path/accounting repairs, refresh the task-scoped evidence and exact
digests, rerun focused/exact/race/package and source-stable repository gates,
then hand off to a new independent reviewer.

No product code was modified by this reviewer. As a reviewer-archetype run,
this verdict supplies no `commit_ack`.
