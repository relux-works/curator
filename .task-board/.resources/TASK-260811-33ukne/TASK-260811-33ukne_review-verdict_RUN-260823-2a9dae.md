# Reviewer verdict for TASK-260811-33ukne

Verdict: **changes requested -> to-dev**

## Review context

- Reviewer run: `RUN-260823-2a9dae`
- Spawn goal: none; the run is not goal-bound
- Scope reviewed: `internal/swiftpmsource`, the task-specific README section,
  developer outcome, attached full-suite log, accepted SwiftPM research, and
  accepted graph/checkpoint contract
- No product code was modified by this reviewer

## Blocking findings

1. **The delivered adapter has no operational SwiftPM acquisition/evaluation/replay implementation.**
   `ManifestEvaluator`, `LockResolver`, `AcquisitionBroker`, and
   `OfflineReplayer` are interfaces in `internal/swiftpmsource/types.go`, but
   repository search finds implementations only in `swiftpmsource_test.go`.
   No non-test call site invokes `CaptureAndClose`. The real Swift smoke calls
   `exec.CommandContext` directly and does not exercise intake, permits,
   acquisition, mirrors, replay, or C4 closure. This does not implement the AC
   requiring controlled manifest evaluation, exact source-control acquisition,
   kind-preserving mirror replay, and no-network planning.

2. **The committed manifest permit cannot be executed by the pinned SwiftPM.**
   `manifestPermit` records `swift package dump-package ... --manifest-path
   <selected>`, but Apple SwiftPM 6.3.2 rejects `--manifest-path` with exit 64
   (`Unknown option '--manifest-path'`). The real smoke omits that option, so it
   validates a different argv from the permit. A faithful production evaluator
   would therefore fail before every manifest derivation.

3. **Local mirrors are neither captured nor content-rechecked.**
   `validateSnapshot` verifies only the mirror directory shape and broker-
   supplied strings. `ReplayOffline` later checks only that the directory still
   exists and has matching in-memory metadata; it never captures/admit-scans the
   mirror repository, proves that it contains the pinned revision/tree, or
   verifies its bytes immediately before replay. A mirror may change after C2
   while all current checks continue to pass, violating immutable one-pin/one-
   mirror closure and exact offline replay.

4. **The pre-C5 resolution derivation journal is incomplete.**
   The generated-lock path creates a `ResolutionPermit` and calls
   `Resolver.Resolve`, but creates no derivation receipt for the resulting lock.
   `evidenceIDs` journals only manifest permit/receipt IDs, and those are what
   C1/C3 receive. The accepted contract requires a committed permit and causal
   receipt for every pre-C5 executable derivation, including generated lock
   resolution.

5. **Dangling lock pins execute affected manifests before rejection.**
   `CaptureAndClose` loops over and evaluates every lock pin first; only after
   all evaluations does `reconcileManifestDependencies` reject a pin that no
   manifest discovered. This violates the zero-affected-process rule for
   rejected dependencies and the discovery-ordered intake/permit/receipt
   journal.

6. **The claimed R01-R13/P01-P08 conformance coverage is not the specified
   corpus.** Test names group IDs, but several required vectors are not
   represented by their actual conditions: for example R04 is tested as a
   mutable lock revision rather than a stale manifest/forced-lock mismatch;
   R08 is represented by a duplicate identity or direct helper mutation rather
   than an acquired mirror commit mismatch; R13 is partly represented by an
   unknown lock field rather than a broker/materialization zero-start vector;
   and P03/P05/P06/P07 lack real adapter-path process/fetch assertions. The only
   real Swift test bypasses the adapter interfaces. Green unit tests therefore
   do not establish the task's named acceptance vectors.

## Verification evidence

- `go test -count=1 ./internal/swiftpmsource`: pass
- `go test -count=1 -race ./internal/swiftpmsource`: pass
- `go test -count=1 -cover ./internal/swiftpmsource`: pass, 80.4%
- adjacent artifact/closure packages: pass
- `go vet ./internal/swiftpmsource`: pass
- `golangci-lint run ./internal/swiftpmsource`: pass, 0 issues
- `go build ./...`: pass
- `make no-broad-suppression`: pass
- canonical golden verifier: pass, 53 records and references
- producer full-suite log: SHA-256 verified as
  `ad8b0d56ff878d4e587969a2d869537215c7bd83f8deb1208fa7b93924404078`;
  recorded suite is green
- exact manifest-permit probe: fail as evidence, exit 64 with
  `Unknown option '--manifest-path'`

## Required rework

- Add real production implementations (or an already-integrated protected
  executor implementation) for manifest evaluation, generated-lock resolution,
  source-control acquisition, and offline replay, then exercise them through
  `CaptureAndClose` in real fixtures.
- Record valid exact SwiftPM commands, with version-specific manifest selection
  implemented in a way the pinned tool supports and tested against the same
  committed argv.
- Capture and admit mirror repositories as immutable evidence; verify their
  revision/tree/object graph and recheck their bytes before replay.
- Add a generated-resolution derivation receipt and include its permit/receipt
  in the canonical C1/C3 journal.
- Acquire/evaluate only packages reached by the discovery walk and reject
  dangling pins before their manifest process starts.
- Implement the literal R01-R13 and P01-P08 vectors, including exact zero-start,
  no-network, empty-output, and no-publication assertions.

This is ordinary implementation rework, not a Stop-The-Line boundary.
