## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] The mechanism is established from evidence, or its absence is stated plainly
- [x] The fix does not widen clockSkew, remove the assertion, or teach the test to ignore a future timestamp
- [x] A narrowing mutant admitting one future-dated snapshot fails a named test
- [x] go test -race on internal/install and internal/registry is green
- [x] Gates run as standalone processes with observed exit codes
- [x] No Windows result is claimed that was not measured on a hosted runner
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Every command, message, state, or refusal named in the AC is driven through the production entry point by a named committed test, or is declared a stated bound. Report coverage as a ratio — `n of m AC rows driven` — and name the production call site for each. Prose in place of the ratio is not evidence.
- [x] Gating, refusing, validating, authorizing, or attesting behavior covered by negative tests that fail when the gate admits what it must reject, with the production call site named
- [x] Every gate ships at least one NARROWING mutant — the gate stays present and is weakened to admit exactly one member of the class it must reject, and a named test must fail. A delete-only mutant proves only that the gate exists and is not accepted as evidence.
- [x] A gate that inspects source text is additionally attacked by a mutant that PRESERVES the searched-for token and changes behavior, and the mutant harness executes the behavioral suite, not only the static checker.
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Item 12 left unchecked: N/A — this gate compares two time.Time values and reads no source text, so no token-preserving mutant exists. Item 16 left unchecked: the producer brief forbids writing LOGBOOK.md for this task; findings are in the board notes and BUG-260906-1bdotx_drafting-report.md instead.
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Gate, refusal, validation, authorization, and attestation behavior attacked, not read — positive-path-only evidence is not accepted
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (claude) (run=RUN-260907-1b62ea, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260907-1b62ea)
Mechanism established from evidence, deterministically, on the production install path.

TWO facts compose; neither alone fires the gate:
1. internal/install newEnv builds config.Config as a struct literal, so Audit is the ZERO value. cfg.Audit.SnapshotClockSkewSeconds is 0, NOT the 300s default -- parseAudit (config.go:709) is the only place the defaults are set and the fixture bypasses it. snapshot.go documents a zero skew as literal. Measured: SnapshotClockSkewSeconds=0 SnapshotMaxAgeSeconds=0.
2. resolveRegistries samples now=time.Now() at install.go:1192 BEFORE the fetch. The fixture minted created_at inside the HTTP handler, strictly later. RFC3339 truncates down to the whole second, so when the fetch crosses a second boundary the served timestamp lands AFTER now. Under a literal zero tolerance, any positive difference is tampering.

Reproduced end to end with a handler forced past one second boundary; exact reported text:
  test: registry: registry test-reg snapshot timestamp is too far in the future
  Errors: [every trusted audit registry served a tampered snapshot]
Control (no crossing) passes.

Window measured inside the real install path on darwin/arm64: mean 667us, worst 3.18ms (856us under -race) => P(flake) per snapshot fetch ~0.067%. Windows magnitude NOT measured; no hosted-runner result claimed.

Fix is test-only and does NOT touch the gate: the fixtures mint created_at once at server construction (registry_e2e_test.go fakeRegistry, dryrun_conformance_test.go loopbackRegistry), like a published snapshot. clockSkew is unchanged and NOT widened; the assertion is untouched.

Mutants, all killed, no survivors:
- M1 After(now.Add(clockSkew + time.Second)) -> TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/{0s,30s,5m0s}/one_second_past_the_bound, TestSnapshotZeroClockSkewIsLiteral, TestSnapshotRequiresCompleteShapeAndRejectsEquivocation
- M2 After(now.Add(max(clockSkew, time.Second))) -> TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/0s/one_second_past_the_bound, TestSnapshotZeroClockSkewIsLiteral
- M3 delete -> the above plus TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries
- M0 revert the fixture fix -> TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch fails with the exact Windows text
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260907-1b62ea, pid=13289, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260907-dc106b, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260907-dc106b)
spawn run RUN-260907-dc106b failed because its runner heartbeat expired; operator action required; failure: spawn runner heartbeat expired
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260907-e84359, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260907-e84359)
REVIEW cycle 1 verdict: ACCEPT at 879b884b (branch fix/snapshot-timestamp-flake, base 919e2e9c).

repository_delta=empty explained: the CR snapshotted the curator-spec story worktree, a DIFFERENT repository with no internal/install or internal/registry. The work is 3 signed commits (G, Ivan Oparin) in curator at 879b884b, 4 files +200/-5. Not an absence of work; a workspace-routing artifact. INTEGRATION MUST TAKE 879b884b FROM THE curator BRANCH, NOT FROM THE CR CANDIDATE TREE.

Verified independently in a throwaway copy; producer worktree and story worktree confirmed clean afterwards.

MECHANISM re-derived by construction, not read. newEnv/newEnvIn print SnapshotClockSkewSeconds=0 (vs 300 production default); parseAudit config.go:709 is the only production writer; clockSkew has no zero-fallback in checkSnapshotsWithPolicy while maxAge does. now is an argument expression at install.go:1191, so Go evaluates it before the callee runs the fetch. REPRODUCED DETERMINISTICALLY ON DARWIN: reverting only the fixture (gate untouched) + forced boundary crossing => exit 1 with both exact strings. The defect is NOT Windows-specific; Windows was only the lane slow enough to hit the window naturally.

GATE INTACT AND ATTACKED. snapshot.go 0-line diff; no production file changed at all. No bypass path: all three exported entry points funnel through checkSnapshotsWithPolicy and the future gate sits BEFORE the persist branch, so both install branches carry it.

MUTANTS REPRODUCED BY THE REVIEWER, base vs head:
- M2 (max(clockSkew, time.Second)) SURVIVED AT BASE (exit 0) and DIES AT HEAD (exit 1) via TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew/0s/one_second_past_the_bound + TestSnapshotZeroClockSkewIsLiteral. This is the careless-fix shape; nothing caught it before. Real coverage gain.
- M3 (delete) LEFT internal/install FULLY GREEN AT BASE (exit 0). Zero production-path coverage existed. Now dies via TestRegistryFutureSnapshotDeniesInstallThroughResolveRegistries.
- M1 dies at head as reported; it was already killed at base inside internal/registry by the pre-existing equivocation test, which the report honestly lists.
- M0 (fix-revert) dies with the exact reported text.

WINDOWS MEASURED ON A HOSTED RUNNER at the exact accepted head: run 34130394522, headSha 879b884b, 12/12 checks pass, Test (windows-latest) pass 43m24s. All four ledger cases reported ok by the platform-case gate, which prints ok only on a pass action in the go test -json stream (platform-case-gate.sh:270, with a distinct never-ran branch at :285) -- execution evidence, not declaration.

REVIEWER GATES: go build 0, go vet 0, gofmt 0, gate-selftest 0 (130/0), go test -race ./internal/install/ ./internal/registry/ 0, 25x forced crossing under -race 25 RUN/25 PASS/0 FAIL, anchored registry set 14 RUN/14 PASS.

BLAST RADIUS: 8 tests consume the fixtures; none asserts on a timestamp; freezing created_at cannot trip equivocation (version/head/merkle/log-size constant) or staleness (7-day fallback).

MINOR NON-BLOCKING FINDINGS: F1 the ledger row and test doc say the bound is pinned at every skew a config can carry, but three skews {0,30s,300s} are covered and a config accepts any integer -- fair sample, overstated prose. F2 TestRegistrySnapshotAtTheSkewBoundIsAcceptedThroughInstall is not at the bound (-1min under skew 0); valid positive control, inaccurate name, and the only new test absent from the ledger. F3 report says not pushed/no PR, true when written; orchestrator has since opened PR 64 at the exact head.

Evidence: BUG-260906-1bdotx_review-verdict-rev1.md
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260907-e84359, pid=17231, exit=0)
Landed on curator main as eca87fe3 through PR #64: eleven hosted lanes green including Windows, reviewer ACCEPT on all four acceptance rows with the mechanism reproduced deterministically on darwin and mutant M2 surviving at base and dying at head. The Change Request records an empty delta because the story worktree is a curator-spec checkout while the fix lives in the curator repository; the integration is the fast-forward of the reviewed head, already performed.
close-landed (legacy, no Change Request record): closed as landed on refs/heads/main at 12f1287ee0fb: pull request 64 names the element and its merge commit eca87fe38957 is an ancestor; method=legacy_pr_attested attested=true landing_commit=eca87fe38957ed08f4819836cc6fe12a3efd18dd authority=12f1287ee0fb538f9ca004dd53b870e824e5baf2; reason: landed by curator PR #64 (reviewed: review-verdict-rev1 resource); record-less legacy closure

## Precondition Resources
- [producer-brief-snapshot-flake.md](file://BUG-260906-1bdotx/producer-brief-snapshot-flake.md) — Producer brief: establish the Windows snapshot-timestamp flake mechanism before changing the gate
- [review-brief-snapshot-flake-1.md](file://BUG-260906-1bdotx/review-brief-snapshot-flake-1.md) — Review brief cycle 1: verify the two facts, reproduce deterministically, keep the tampering gate intact

## Outcome Resources
- [BUG-260906-1bdotx_spawn-log_-implementer--developer--claude-_RUN-260907-1b62ea.log](file://BUG-260906-1bdotx/BUG-260906-1bdotx_spawn-log_-implementer--developer--claude-_RUN-260907-1b62ea.log) — System spawn log captured by task-board
- [BUG-260906-1bdotx_drafting-report.md](file://BUG-260906-1bdotx/BUG-260906-1bdotx_drafting-report.md) — Mechanism (literal zero skew + per-fetch created_at across a second boundary), the test-only fix, mutant table with no survivors, gate table with observed exit codes, and the curator-spec/curator workspace discrepancy
- [BUG-260906-1bdotx_change-request_rev1.patch](file://BUG-260906-1bdotx/BUG-260906-1bdotx_change-request_rev1.patch) — Change Request CR-BUG-260906-1bdotx-1 revision 1 candidate patch (repository_delta=empty, 0 changed paths)
- [BUG-260906-1bdotx_spawn-log_-reviewer--reviewer--claude-_RUN-260907-dc106b.log](file://BUG-260906-1bdotx/BUG-260906-1bdotx_spawn-log_-reviewer--reviewer--claude-_RUN-260907-dc106b.log) — System spawn log captured by task-board
- [BUG-260906-1bdotx_spawn-log_-reviewer--reviewer--claude-_RUN-260907-e84359.log](file://BUG-260906-1bdotx/BUG-260906-1bdotx_spawn-log_-reviewer--reviewer--claude-_RUN-260907-e84359.log) — System spawn log captured by task-board
- [BUG-260906-1bdotx_review-verdict-rev1.md](file://BUG-260906-1bdotx/BUG-260906-1bdotx_review-verdict-rev1.md) — Review verdict CR rev1: ACCEPT at 879b884b. Mechanism re-derived by construction, flake reproduced deterministically on darwin, M2 shown to survive at base and die at head, M3 shown to leave internal/install green at base, hosted Windows execution measured at the exact head, three minor non-blocking findings.

## Created
2026-09-06T09:21:16Z

## Last Update
2026-09-16T10:08:26Z

## Assigned To
[reviewer] reviewer (claude)
