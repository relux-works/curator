# TASK-260729-2kaopg integrated tester evidence

Date: 2026-07-29. Role: tester. Candidate: .temp/TASK-260729-2kaopg/worktree. No product code was edited, staged, committed, published, or pinned by this tester run.

## Audit and provenance

Every real curator global status invocation reaches cmdGlobalStatus and passes globalStatusPlan into reportGlobalStatus, so it acquires one fresh read-only plan. Test-only replays inject an already acquired immutable install.Result into that same production classifier, renderer, marker/cache bracketing, and fail-closed check path. The pre-rework versus rework audit retained all stable states, causes, cache outcomes, non-empty details, JSON rows, human lines, and check exits; representative unchanged, source drift, context drift, cache drift, unusable toolchain, and transitive paths remain end-to-end CLI runs.

Both exact patches apply cleanly to the accepted currentness snapshot: git apply --check --whitespace=error-all exited 0 for the owned rework patch and 0 for the accepted cycle-3 fingerprint patch. Owned patch SHA-256: 5becface29bdc22cb82f14efb95264a9590b09c4530a5665b1e663dc5ca028bd. Accepted fingerprint patch SHA-256: a7e0906612ce6f007bfdb3776de632dd9c7a673e9b501443be5fb3eced8f1beb. Integrated fingerprint.go SHA-256: 560d0c98c665a5a83c3a6989a7b0cdcc9f26c4fb513c7688d9b1bd6e42552d1d; equivalence test SHA-256: 6390e75c9848f575f2f4b50217ebd1d53481a58d349073fb0e819491b5fed484.

## Focused gates

- go test -count=1 -run ^TestGlobalStatus ./cmd/curator: exit 0, 44.638s.
- Same selection with coverprofile: exit 0, 43.082s, package coverage 28.9%. go tool cover -func: exit 0. Owned functions: cmdGlobalStatus 100%, reportGlobalStatus 100%, globalStatusPlan 100%, globalStatusScope 100%, statusReport 86.7%, classifySkillBuilds 81.8%, installedSkillDir 80.0%; checkFailed, factsOf, Describe, plannedRows, demoteSkill, and markerMoved 100%.
- Accepted cache-bracketing and rollback regressions: exit 0, 134.959s.
- Fifteen-test fingerprint mutation/equivalence matrix: exit 0, 3.496s; real GOROOT produced 16093 identical records and digest sha256:ea13c6bb11293e951ab9f189144a1f660cb2f398385109c0a3f7ad4875942191.
- go test -count=1 ./internal/godriver: exit 0, 28.901s.
- Same with the pinned CURATOR_CONFORMANCE_ROOT: exit 0, 31.012s.
- Explicit RC4 framing and implementation vectors: exit 0, 0.342s; both passed.
- go build ./...: exit 0. go vet ./...: exit 0. Changed-file gofmt -l: exit 0 with no output. git diff --check: exit 0 with no output.
- command -v golangci-lint: exit 1 because the executable is absent; it was not installed per directive. Preserved lint artifact lint-02.log reports 0 issues.

## Consecutive repository gates

Both commands were foreground, standalone, literal go test -count=1 ./... invocations with the standard default timeout, no exclusions, no timeout override, no tee, and no overlapping Go suite. Host barriers found no foreign Go/test process.

1. Gate 1: exit 0. cmd/curator 554.967s; internal/install 306.470s; internal/install/atomicity 441.457s; every package passed. Disk available before 4775308 KiB and after 7624416 KiB.
2. Gate 2: exit 0. cmd/curator 545.195s; internal/install 308.679s; internal/install/atomicity 436.404s; every package passed. Disk available before 7624416 KiB and after 7415448 KiB.

The integrated candidate satisfies the tester gates and is ready for independent review. Item 17 remains unchecked for the reviewer.