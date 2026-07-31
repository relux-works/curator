## Status
to-dev

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(3))

## Blocked By
- BUG-260731-33v6zz

## Blocks
- (none)

## Checklist
- [x] Reproduce the rc.6 multi-project dry-run binding failure against the authoritative conformance root.
- [x] Implement a real executable multi-project binding without weakening or skipping the authoritative assertion.
- [ ] Add focused regression coverage and prove internal/install against the rc.6 root on macOS, Linux, and Windows CI.
- [x] Publish a signed Curator PR targeting main with outcome evidence and hand off to independent Opus review.
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [x] Solution fits project architecture
- [ ] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-b497db, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-b497db)
READY FOR REVIEW — code + tests complete and verified on macOS; two DoD items externally blocked (no PR, no Linux/Windows CI). See BUG-260731-3a5q1p_implementation-and-evidence.md.

AC MET. CURATOR_CONFORMANCE_ROOT=<rc.6 b07ef1d5> go test -count=1 ./internal/install/... exits 0. The rc.3 pin still exits 0 on both dry-run tests. Neither default: branch of TestAuthoritativeDryRunCasesMutateNothingPersistent was weakened and the new case is not skipped — mutations 5 and 6 prove an unbound scope and an unbound effect are still fatal.

SCOPE WAS BIGGER THAN THE TITLE. rc.6 adds scope multi-project AND 19 new forbidden persistent effects (28 vs rc.3s 9), of which 18 had no binding, so assertNoEffects own default: was the second wall after the scope switch. It also adds five per-case fields (allowed/forbidden_go_commands, artifact_executed, logical_cache_key, operation_private_state_after, reported_build_outcomes) that nothing bound to a run. All are bound now.

BINDING. Curators multi-project operation is install --all / upgrade --all (cmd/curator/main.go:520,:635): every target planned in one run, one manager home, one skills root, one shared FetchedRepos set, in canonical project-identity byte order per profiles/manager.md 2.5. A dry run takes no lock (managerlock.ErrDryRun), so order is the only remaining 2.5 obligation. Both projects declare the same compiled command, so both derive ONE shared logical cache key and both miss the same protected entry. Real goToolchain (its probe runs exactly the three allowed commands) and real buildcache.Store; only the source-aware builder is replaced by one that records and refuses.

NOT VACUOUS. Effect bindings are now probes that report where a surface is visible, and the new TestDryRunEffectBindingsSeeWhatARealOperationWrites runs them backwards: a real project install + real global install + real non-dry managerlock operation produce 21 of 27 surfaces in production locations and every binding must see them; completeness is asserted so a new effect cannot get a binding without a witness. Six mutations each made the assertions fail and were reverted.

DELIBERATE: t.Parallel() dropped from the suite. isolateTempDir uses t.Setenv, which forbids a parallel ancestor, and that isolation is what binds operation_private_state_after: absent for real. Scheduling change, not an assertion change; ~0.2s at rc.3.

EXPECTED-RED, NOT MINE: the full rc.6 gate on this branch exits 1 solely on internal/interop TestManagerLifecycleVectors (golden_test.go:488 hardcodes len(DryRunCases) != 2) — pre-existing on main, BUG-260731-3gm8kc/PR 9 scope, fixed there by fee35c8. On main + PR 9 head bd6ba08 + this patch the same gate exits 0 with CI_REQUIRE_FULL_ROOT=1, served=40 deferred=0. So SPEC_PIN to rc.6 needs BOTH bugs; neither alone clears it.

GREEN: install rc.6 exit 0, install rc.3 exit 0, -race exit 0, golangci-lint v2.12.2 exit 0, gofmt exit 0, go vet ./... exit 0, ledger-consistency exit 0, no-broad-suppression exit 0.

BLOCKED (human input, not a decision): ~/.ssh/ivanopcode is passphrase-protected and the passphrase is not reachable from this spawn — ssh-add --apple-load-keychain loads nothing, ssh -T git@github.com is Permission denied (publickey). No signed commit, no push, hence no PR and no Linux/Windows CI. Change saved as BUG-260731-3a5q1p_multiproject-dryrun-binding.patch and also uncommitted on branch task/BUG-260731-3a5q1p-multiproject-dryrun in .temp/BUG-260731-3a5q1p/worktree (verified byte-identical). Exact commit/push/PR commands and the gh workflow run dispatch for the three-platform rc.6 candidate lane are in section 5 of the evidence artifact; the drafted commit message is in section 6. Checklist items 3 and 4 left unchecked because those commands did not run.

REVIEWERS: section 7 of the artifact names the three judgement calls worth arguing with — two sequential dry runs as the meaning of multi-project, several effects sharing one root because of Curators layout, and 6 of 27 witnesses produced rather than reached.
ORCHESTRATOR CHECKPOINT 2026-07-31: predecessor RUN-260731-b497db was cancelled after becoming idle post-gates. Preserve and continue the existing worktree /Users/iv/Developer/ReluxWorks/curator/.temp/BUG-260731-3a5q1p/worktree on branch task/BUG-260731-3a5q1p-multiproject-dryrun. Only two unstaged files remain: .github/ci/root-artifacts.tsv and internal/install/dryrun_conformance_test.go; temporary debug file is absent. Evidence: CI_REQUIRE_FULL_ROOT=1 rc.6 combined test-gate exit=0 and platform-case gate ok at .temp/BUG-260731-3a5q1p/logs/test-gate-rc6-combined.log; focused internal/install rc.6, golangci-lint (0 issues), gofmt and vet completed. Successor must inspect the 700+ line test delta for necessity/maintainability, verify exact exit logs, create a signed commit and PR, attach outcome evidence, and hand off to-review. Do not discard the preserved worktree or broaden scope.
SIGNING UNBLOCKED: use SSH_AUTH_SOCK=/Users/iv/.ssh/agent/s.bCdCK3Af2e.agent.FkHJrmRP1K. ssh-add -l on that socket exposes signing key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM plus GitHub SSH keys. Continue autonomously: signed commit, push branch, open PR, handoff.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-c1854b, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-c1854b)
agent completed: [implementer] developer (claude) (exit=1)
spawn run completed: claude (run=RUN-260731-c1854b, pid=44641, exit=1)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260731-63da06, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260731-63da06)
RUN (current): re-verified on current main 3a047d5 (post PR 9/10/11 merges) in .temp/BUG-260731-3a5q1p/rebased.

Verified so far (real exit codes, no tee, no pipes):
- repro on clean main 3a047d5 + rc.6 root: exit 1 (expected-red) — scope "multi-project" has no executable binding
- go test ./internal/install/... + rc.6 root, with fix: exit 0 (install 102.2s, atomicity 92.6s) = the acceptance criterion
- gofmt -l cmd internal: exit 0, no output
- go vet ./...: exit 0
- no-broad-suppression.sh: exit 0
- golangci-lint v2.12.2 (CI-pinned): exit 0, 0 issues

In flight: full .github/ci/test-gate.sh with CI_REQUIRE_FULL_ROOT=1 against rc.6.
Remaining: rc.3 SPEC_PIN regression check, signed commit via GitHub createCommitOnBranch (SSH signing key is passphrase-locked and no agent holds it; the GraphQL route is the only signing path from this session), PR to main, 3-platform candidate CI dispatch.
Migration checkpoint 2026-07-31: implementation is signed and pushed as d345420, PR 14. Local rc6 combined gate and focused install/race/lint/gofmt/vet evidence passed. Spawn an independent Claude Opus 5 reviewer on the new host.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: disabled; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=launch_composition_disabled; bare child launch retained
spawn queued: [reviewer] reviewer (claude) (run=RUN-260731-cefd98, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260731-cefd98)
REVIEW VERDICT RUN-260731-cefd98: CHANGES REQUESTED -> to-dev. Full evidence in BUG-260731-3a5q1p_review-verdict_RUN-260731-cefd98.md.

INDEPENDENTLY REPRODUCED (darwin/arm64, Go 1.25.5, fresh git archive checkouts of origin/main and PR14 head d345420, rc.6 root b07ef1d5 = origin/release/v1.0.0-rc.6): (1) main + rc.6 fails with scope multi-project has no executable binding; (2) PR head + rc.6, both dry-run tests PASS incl. compiled-cache-miss-is-read-only; (3) go test ./internal/install/... + rc.6 exits 0 (install 406.7s, atomicity 365.5s) = AC clause 1 MET; (4) PR head + rc.3 pin still PASSES = no regression; (5) both default: branches (scope line 292, effect line 692) still fatal, new case not skipped, no assertion special-cased; (6) PR14 CI vs main CI: only Test (windows-latest) red on both, with a byte-identical failing-case set (7x cmd/curator, buildsource TestFrozenTokenRejectsRootReplacement, install TestEndToEndInstall, 3x atomicity) owned by BUG-260731-27h1yc/BUG-260731-fs3dht. This PR regresses nothing.

BLOCKING: AC clause 2 (advancing SPEC_PIN to rc.6 does not turn Curator CI red on this test) is NOT demonstrated and is currently FALSE on windows-latest. DoD item 3 was never run: no candidate-conformance workflow_dispatch exists for this branch, though rc.6 is an immutable published 40-char ref, the manifest sha256 is already recorded in the evidence artifact, signing was unblocked before push, and sibling tasks dispatched that same lane the same day. That unrun lane hides a real failure: planEveryProject drives the REAL goToolchain, and per BUG-260731-fs3dht godriver.Probe rejects the actions/setup-go Windows GOROOT; a failed probe makes planBuilds emit toolchain-unavailable, which rc.6 reported_build_outcomes does not admit. I reproduced the exact mechanism by running the compiled test binary against the rc.6 root with an unresolvable GOROOT: dryrun_conformance_test.go:291 project 0 reported build outcome toolchain-unavailable, which the published case does not admit. internal/install is never platform-excluded (only internal/godriver is), so under rc.6 it is served on all three runners. Linux is separately unverified; the probe path does not go through probeNativeControls so it should be fine, but that is unproven.

REWORK (small): (a) dispatch candidate-conformance on this branch with candidate_ref=b07ef1d51dc6ebcc04cf59a20375d191bf82f6bb + the recorded candidate_manifest_sha256, attach the three per-platform evidence artifacts; (b) report Linux and Windows results honestly, and if Windows fails as predicted add an explicit blocked_by on BUG-260731-fs3dht and note on fs3dht that internal/install rc.6 now depends on its fix, so this does not merge as a silent landmine under the next SPEC_PIN bump; (c) correct the ordering rationale.

NON-BLOCKING (c): the §2.5 claim is inverted. rc.6 profiles/manager.md §2.5 governs PROJECT LOCK acquisition order (implemented by managerlock.AcquireProjects via CanonicalProjects), not planning-iteration order; production install --all iterates by ALIAS (cmd/curator/main.go:264 sort.Strings(aliases)), not canonical-path order; and a dry run takes no lock, so §2.5 ordering does not apply at all rather than being the only remaining obligation. Nothing asserts order so no test outcome depends on it, but fidelity-to-production is the whole justification for the binding, so the comment should state what is true: one manager home, one skills root, one shared FetchedRepos set, deterministic test-chosen order.

ALSO NOTED (accepted as-is): revocation-state binds all of <home>/state, subsuming project-lock, cache-build-lock, manager-home-lock and journal, so those four are witnessed trivially; permission-repair is identical to compiled-artifact-cache. Both are documented consequences of Curator layout and were flagged by the implementer. The anti-vacuity witness test, the t.Parallel removal (t.Setenv in isolateTempDir), the Fetch:true arming, and the root-artifacts.tsv line are all correct and verified.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260731-cefd98, pid=23831, exit=0)
Migration routing: reviewer requested candidate-conformance evidence after integrated Windows trusted-GOROOT behavior lands. The active dependency is BUG-260731-33v6zz / Curator PR #13; the closed BUG-260731-fs3dht branch is comparison-only and must not be merged. After PR #13 lands, rebase PR #14, correct the rc6 ordering rationale, dispatch the three-platform candidate lane, then route to a fresh independent Opus reviewer.

## Precondition Resources
(none)

## Outcome Resources
- [BUG-260731-3a5q1p_spawn-log_-implementer--developer--claude-_RUN-260731-b497db.log](file://BUG-260731-3a5q1p/BUG-260731-3a5q1p_spawn-log_-implementer--developer--claude-_RUN-260731-b497db.log) — System spawn log captured by task-board
- [BUG-260731-3a5q1p_implementation-and-evidence.md](file://BUG-260731-3a5q1p/BUG-260731-3a5q1p_implementation-and-evidence.md) — Reproduction, multi-project binding design, 21 effect bindings incl. 18 new, 12 verification commands with real exit codes, 6 mutation checks, tool readiness, and the blocked-publishing handoff
- [BUG-260731-3a5q1p_multiproject-dryrun-binding.patch](file://BUG-260731-3a5q1p/BUG-260731-3a5q1p_multiproject-dryrun-binding.patch) — The complete change against main cfffd7c: internal/install/dryrun_conformance_test.go +727/-46 and one root-artifacts.tsv line
- [BUG-260731-3a5q1p_verification-logs.tgz](file://BUG-260731-3a5q1p/BUG-260731-3a5q1p_verification-logs.tgz) — All verification logs: rc.6 and rc.3 package runs, race run, both CI test-gate runs with their evidence dirs (observed-cases.tsv, skips-observed.tsv), lint, vet, ledger
- [BUG-260731-3a5q1p_spawn-log_-implementer--developer--claude-_RUN-260731-c1854b.log](file://BUG-260731-3a5q1p/BUG-260731-3a5q1p_spawn-log_-implementer--developer--claude-_RUN-260731-c1854b.log) — System spawn log captured by task-board
- [BUG-260731-3a5q1p_spawn-log_-implementer--developer--claude-_RUN-260731-63da06.log](file://BUG-260731-3a5q1p/BUG-260731-3a5q1p_spawn-log_-implementer--developer--claude-_RUN-260731-63da06.log) — System spawn log captured by task-board
- [BUG-260731-3a5q1p_spawn-log_-reviewer--reviewer--claude-_RUN-260731-cefd98.log](file://BUG-260731-3a5q1p/BUG-260731-3a5q1p_spawn-log_-reviewer--reviewer--claude-_RUN-260731-cefd98.log) — System spawn log captured by task-board
- [BUG-260731-3a5q1p_review-verdict_RUN-260731-cefd98.md](file://BUG-260731-3a5q1p/BUG-260731-3a5q1p_review-verdict_RUN-260731-cefd98.md) — Independent reviewer verdict (changes requested): AC clause 1 reproduced green on rc.6; clause 2 not demonstrated and currently false on windows-latest; three-platform candidate CI lane never dispatched

## Created
2026-07-31T09:29:30Z

## Last Update
2026-07-31T13:41:18Z

## Assigned To
[reviewer] reviewer (claude)
