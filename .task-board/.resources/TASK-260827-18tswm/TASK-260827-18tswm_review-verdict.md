# TASK-260827-18tswm review verdict — RUN-260828-c137a2

## Verdict

Accepted Change Request `CR-TASK-260827-18tswm-2` revision 2.

No correctness, architecture, assertion-integrity, or evidence-binding findings remain.

## Candidate identity

- Base OID: `b84b12e4b3051e69622dbe9c0bdd2192d5b040df`
- Candidate tree OID: `134adead0f9f482b1cca6c254e77293bb7406f32`
- Patch resource SHA-256: `f6a964c4fd12b7b987d7aba0fbd953f90b99956065f31ebb121afa21130e5524`
- Recomputed `git diff --binary` SHA-256 matched the patch resource.
- `git diff --check` passed.

## CI and candidate binding

GitHub Actions run https://github.com/relux-works/curator/actions/runs/33130874599 completed successfully on PR head `c2215f9b929e11a32d75bff1205d296c135ddd7f`.

The successful jobs cover:

- Test on Ubuntu, macOS, and Windows;
- Race on Ubuntu and macOS;
- Lint;
- Naming gate;
- Interop conformance gate;
- Gate self-test on Ubuntu, macOS, and Windows.

The landed workflow deliberately excludes Windows from the race matrix and records the C-toolchain rationale in `.github/workflows/ci.yml`.

All five downloaded test/race evidence artifacts were inspected. Their platform-case reports end in `platform-case gate: ok`; no `UNCLASSIFIED` or `FATAL` rows were present. Each artifact records exactly 3 pinned-pnpm skips, 1 Yarn Classic skip, and 4 Yarn Modern opt-in skips. Ubuntu records the three exact Swift linker host-capability skips. Linux records both verified-provider tests under the existing native-control-inventory platform class.

Independent tree comparison proved that 64 task-delivered paths in revision 2 are blob-identical to CI head `c2215f9b`. The only four delivered paths changed after that run are `LOGBOOK.md`, the two Swift integration call sites, and `internal/testtoolchain/lock.go`; revision 2 adds only `internal/testtoolchain/swift.go` and `internal/testtoolchain/swift_test.go` beyond that delivered set. The new predicate requires the exact two-diagnostic conjunction, and its negative regression rejects either diagnostic alone and an unrelated Swift permit failure.

## Acceptance-criteria review

- Rust absence is classified only for an unapproved native descriptor or an absent pinned root/executable. Present malformed or mismatched toolchains remain fatal. No descriptor or SHA was added.
- Crossconformance permits only the explicit closed six-obligation Rust gap set. Extra Rust and non-Rust gaps have negative regressions and fail the completeness gate.
- The two `cmd/curator` tests use `requireNativeControlInventoryPlatform`; Linux evidence shows the named platform-control skips rather than build-control failures.
- npm, pnpm, Yarn Classic, SwiftPM, Windows path/DACL/link, and host-GOROOT changes are covered by the successful CI run and affected-package tests. The skip reasons name the observed host limitation.
- Release-pin promotion files were untouched. `no-broad-suppression` and gate self-tests pass. Fatal assertions remain after the narrow skip predicates.

## Reviewer validation

Commands run directly by this reviewer:

- `go test -count=1 ./internal/testtoolchain ./internal/rustsource ./internal/crossconformance` — pass.
- `go test -count=1 -run '^(TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt|TestCLIVerifiedCapabilityDriftStartsNothingAndAdoptsNoCache)$' ./cmd/curator` — pass.
- `go test -count=1 ./internal/npmsource ./internal/pnpmsource ./internal/yarnclassicsource ./internal/yarnmodernsource ./internal/swiftpmsource ./internal/swiftpmbuild ./internal/privatedir ./internal/closureexec` — pass.
- `bash .github/ci/gate-selftest.sh` — 81 passed, 0 failed.
- `git diff --check <base> <candidate>` — pass.

Acceptance is recorded without `commit_ack`. The commit-owning orchestrator owns checkpoint/integration and the final `done` transition.
