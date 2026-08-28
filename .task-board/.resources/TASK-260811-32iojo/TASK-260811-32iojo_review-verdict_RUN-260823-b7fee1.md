# Reviewer verdict for TASK-260811-32iojo

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260823-b7fee1` (not goal-bound).

## Blocking findings

1. **The `yarn.lock` parser is not a closed typed, single-document grammar.** `parseLock` decodes one YAML document and never proves EOF (`internal/yarnmodernsource/lock.go:288-337`). Its `stringMap`, `stringSlice`, and `metaOptional` helpers silently coerce or erase malformed node types (`lock.go:917-950`). The executable probe showed that an external package entry with `dependencies: []` and a lock with a second YAML document are both accepted. Each produces the same canonical `LockDigest` as the valid baseline while its `RawLockSHA256` differs. This lets unsupported lock bytes alias an accepted canonical lock identity, contrary to the exact lock/profile binding and fail-closed parser requirements. The focused suite contains a second-document vector for `.yarnrc.yml`, but no equivalent lock/type-confusion regression.

2. **External-package peer metadata is not reconciled against the authoritative lock.** `reconcileEmbeddedMetadata` compares only `dependencies` and `optionalDependencies` (`internal/yarnmodernsource/capture.go:713-736`). Immediately afterward, `CaptureAndAdmit` replaces the lock-derived `PeerDependencies` and `PeerOptional` with artifact metadata and rebuilds graph edges (`capture.go:190-207`). An overlay probe passed when the admitted package declared `react` as a peer but the lock package declared no peer at all. Embedded metadata must agree with the lock graph; it cannot silently widen or replace it before `NodeCapture` construction.

## Prior findings rechecked

- Missing root workspace lock entry now returns `closure_lock_stale` with zero packages, edges, and configuration identity.
- Root workspace dependency metadata drift now returns `closure_lock_stale` with zero graph/config output.
- Malformed behavioral rc types and unknown nested `supportedArchitectures` keys now return `closure_lock_format_unsupported` with zero graph/config output.
- Source hashes match the latest rework evidence: `lock.go d53a79f6...` and `conformance_test.go 2edc362c...`.

## Verification evidence

- New lock probe: `go run ./.temp/TASK-260811-32iojo-review-4/probe.go`; exit 0. `reproductions.log` records accepted malformed dependency sequence and accepted second document, both canonical-digest aliases of the baseline.
- Peer reconciliation overlay probe: `go test -overlay .temp/TASK-260811-32iojo-review-4/overlay.json -run TestReviewerProbeLockPeerMetadataDrift -v ./internal/yarnmodernsource`; exit 0 and explicitly records the fail-open acceptance.
- Previous-review probe rerun: exit 0, with all three earlier unsafe inputs now rejected before graph/config emission.
- Focused modern Yarn suite with pinned Yarn 4.9.2 integration enabled: exit 0 (`2.133s`).
- Focused race suite: exit 0 (`5.621s`).
- `gofmt -l internal/yarnmodernsource`: empty.
- `git diff --check`: exit 0.

The green suite is not acceptance evidence for the two missing fail-closed cases above. Full repository/lint/vet/build gates from the latest rework remain green and source-identical, but rerunning expensive broad gates cannot change this verdict.

No product code was modified, staged, committed, reset, or cleaned by this reviewer. Scratch probes live only under `.temp/TASK-260811-32iojo-review-4/`.

## Required rework

- Require EOF after exactly one `yarn.lock` YAML document.
- Strictly type-check every supported lock field and nested metadata shape; reject maps/sequences/scalars of the wrong type rather than coercing or treating them as absent.
- Include every behavior-affecting admitted lock field in canonical identity and add regression vectors proving malformed type and trailing-document variants cannot alias a valid lock digest.
- Reconcile external package `peerDependencies` and `peerDependenciesMeta` against lock metadata before graph construction. Do not overwrite lock-derived graph authority with artifact metadata.
- Add negative tests proving zero later manager/build/publication starts for these variants.
