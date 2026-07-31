## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- BUG-260729-1o0m8f

## Blocks
- TASK-260720-1pvfj5

## Checklist
- [x] Reproduce the exact race failure and record competing fail-closed outcomes
- [x] Implement the narrow deterministic contract without broad error acceptance
- [x] Run focused repeated race and non-race godriver gates with exact exits
- [x] Attach a task-scoped patch and outcome for independent review
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
Discovery evidence: TASK-260720-1pvfj5 rework final-race/race/go-test-served.json. Failing subtest got go-v1 toolchain_mutated for a file changed while reading; assertion expected nil or toolchain_timeout. No WARNING: DATA RACE diagnostic. Preserve accepted product candidate provenance.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-eb4053, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-eb4053)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-3afcd3, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-3afcd3)
RESOLVED - ready for review. Root cause: digestToolchainRecords collapsed every non-clean record copy into toolchain_mutated, so a cancellation landing inside a file read reported mutation while a cancellation anywhere else in the same pipeline reported toolchain_timeout. Precedence decision (narrow, not cancellation-wins): copyWithContext tags only its OWN ctx check with package-private errCopyAbandoned; digestCopyDiagnostic branches on that sentinel alone. Read error, close error, a length that moved on its own, and a bare context.Canceled that did not come from that loop all still fail closed as toolchain_mutated - pinned by TestDigestCopyDiagnosticPrecedence. Reclassification between two refusals, both fail closed; toolchain_timeout was already emitted by the walk so no consumer sees a new value; error still unwraps to context.Canceled.

CORRECTION TO THE FIRST APPROACH - MEASURED: repeating the racy subtest is NOT a gate. With the fix deliberately reverted it stayed green at 10,000 racing attempts (-count=50), 100,000 attempts (-count=500), and 2,000 attempts with a fresh 256KiB fixture each. The original flake was load-induced (the discovery run overlapped TASK-260720-1pvfj5 final-race). The real gate is deterministic: countdownContext places cancellation on an exact check including between two chunks of one file, and the new sweep subtest every_cancellation_point_stays_a_deadline walks the budget 0..N asserting nil or toolchain_timeout at every point - it fails the reverted code at budget 17 in 0.47s and is self-extending. The racy loop is kept as a ~5ms smoke check with a comment saying what it does not prove.

GATES on final bytes, standalone processes, real exits: gofmt -l 0 (no output); go vet ./internal/godriver/ 0; focused non-race go test -run TestFingerprint|TestToolchain|TestDigestCopy|TestGoVersion|TestCandidateRC4 -count=1 exit 0 (3.06s); deterministic cancellation -count=1 -v exit 0, 19 subtests pass; go test -race -run TestFingerprintCancellationStaysFailClosed|TestDigestCopyDiagnosticPrecedence -count=1000 exit 0 (28.0s = 200k racing attempts + 1000 deterministic sweeps); focused family under -race -count=5 exit 0 (18.1s).

LINT ITEM LEFT UNCHECKED, honestly: golangci-lint was not installed, so it was run at the CI-pinned v2.12.2 via go run against the shipped .golangci.yml, scoped: golangci-lint run ./internal/godriver/... EXIT 1 on ineffassign at builddriver_positive_conformance_test.go:178. The two files this patch changes give ZERO findings; the same command on the accepted TASK-260720-jrrgw9 baseline worktree gives byte-identical output and exit 1, proving it pre-existing. ineffassign is default-on in golangci-lint v2, so composite CI lint is red on this today independently of this patch - needs an owner outside this bug. make lint (repo-wide) was not run.

NOT RUN by instruction: ./... and any Curator full/race suite - that gate belongs to TASK-260720-1pvfj5 and must not overlap another heavy race run. TestFingerprintRejectsInvalidUnicodePath skips on this host (APFS rejects the invalid-UTF-8 fixture name) - pre-existing and unrelated.

SCOPE: internal/godriver/fingerprint.go + fingerprint_equivalence_test.go only; every other godriver file byte-identical to the accepted jrrgw9 candidate. Patch applies clean to .temp/TASK-260720-1pvfj5/rework/composite (git apply --check exit 0); the composite godriver baseline was verified byte-identical to jrrgw9 first. Applying it there is TASK-260720-1pvfj5 step, after review. Worktree .temp/BUG-260729-r0fe02/worktree.
Routing note: producer patch and focused gates are complete. Generic Lint clean remains unchecked solely because accepted builddriver_positive_conformance_test.go has the pre-existing ineffassign now owned by BUG-260729-1o0m8f. After that cleanup is accepted, replay lint on the combined two-patch candidate, complete the generic checklist, and route this patch to independent review.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-4752b2, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-4752b2)
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260729-4752b2, pid=32593, exit=1)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-1f4480, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-1f4480)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-f6ebe7, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-f6ebe7)
Provider routing note 2026-07-29: Opus replay RUN-260729-4752b2 failed before first token with Claude API 500; retries RUN-260729-1f4480 and RUN-260729-f6ebe7 produced no first token for more than two and one-and-a-half minutes and were cancelled without code or board artifact changes. To keep the critical path moving, the strictly mechanical combined-evidence replay is rerouted to Codex; independent review remains a separate Codex reviewer run.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (codex) (run=RUN-260729-1521c1, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260729-1521c1)
Combined lint replay 2026-07-29: Applied accepted BUG-260729-1o0m8f patch sha256 8a07c0b239548235aea7dfa05fdb1d1cb2926971d4444d3435a9e6f8da368062 exactly; reverse apply check and five-file byte comparison exit 0. Cancellation sources and task patch remain byte-identical at d53cb4f4..., fb2f286b..., and 462f1ff0.... CI-pinned golangci-lint v2.12.2 full run exits 0 with 0 issues. Focused deterministic cancellation plus mutation controls, accepted protocoljson/transaction tests, gofmt, and scoped vet all exit 0. Existing repeated race evidence reused; no ./..., full suite, or new race run per precondition. New outcome: BUG-260729-r0fe02_lint-compatibility-results.md. All checklist items now satisfied; ready for independent review.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-1521c1, pid=34142, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-25bb90, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-25bb90)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-25bb90, pid=36112, exit=0)

## Precondition Resources
- [BUG-260729-r0fe02_resume.md](file://BUG-260729-r0fe02/BUG-260729-r0fe02_resume.md) — Resume instructions after enforced heavy-test barrier
- [BUG-260729-r0fe02_combined-evidence.md](file://BUG-260729-r0fe02/BUG-260729-r0fe02_combined-evidence.md) — Precondition for lint-compatible cancellation evidence after BUG-260729-1o0m8f acceptance

## Outcome Resources
- [BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-eb4053.log](file://BUG-260729-r0fe02/BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-eb4053.log) — System spawn log captured by task-board
- [BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-3afcd3.log](file://BUG-260729-r0fe02/BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-3afcd3.log) — System spawn log captured by task-board
- [BUG-260729-r0fe02_results.md](file://BUG-260729-r0fe02/BUG-260729-r0fe02_results.md) — Root cause, precedence decision, measured negative controls, exact gate exits, pre-existing lint finding
- [BUG-260729-r0fe02_patch.diff](file://BUG-260729-r0fe02/BUG-260729-r0fe02_patch.diff) — Narrow patch vs accepted TASK-260720-jrrgw9 godriver bytes; applies cleanly to the 1pvfj5 composite
- [BUG-260729-r0fe02_gatelogs.tar.gz](file://BUG-260729-r0fe02/BUG-260729-r0fe02_gatelogs.tar.gz) — Raw gate logs: vet, focused non-race, deterministic cancellation, race count=1000, race family, lint, and all negative controls
- [BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-4752b2.log](file://BUG-260729-r0fe02/BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-4752b2.log) — System spawn log captured by task-board
- [BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-1f4480.log](file://BUG-260729-r0fe02/BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-1f4480.log) — System spawn log captured by task-board
- [BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-f6ebe7.log](file://BUG-260729-r0fe02/BUG-260729-r0fe02_spawn-log_-implementer--developer--claude-_RUN-260729-f6ebe7.log) — System spawn log captured by task-board
- [BUG-260729-r0fe02_spawn-log_-implementer--developer--codex-_RUN-260729-1521c1.log](file://BUG-260729-r0fe02/BUG-260729-r0fe02_spawn-log_-implementer--developer--codex-_RUN-260729-1521c1.log) — System spawn log captured by task-board
- [BUG-260729-r0fe02_lint-compatibility-results.md](file://BUG-260729-r0fe02/BUG-260729-r0fe02_lint-compatibility-results.md) — Pinned lint replay, cancellation hash proof, scoped non-race gates, and reused race evidence
- [BUG-260729-r0fe02_spawn-log_-reviewer--reviewer--codex-_RUN-260729-25bb90.log](file://BUG-260729-r0fe02/BUG-260729-r0fe02_spawn-log_-reviewer--reviewer--codex-_RUN-260729-25bb90.log) — System spawn log captured by task-board
- [BUG-260729-r0fe02_reviewer-verdict.md](file://BUG-260729-r0fe02/BUG-260729-r0fe02_reviewer-verdict.md) — Independent reviewer verdict and validation evidence

## Created
2026-07-29T19:11:18Z

## Last Update
2026-07-29T20:07:08Z

## Assigned To
[reviewer] reviewer (codex)
