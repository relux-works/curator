## Status
done

## Assigned To
[reviewer] reviewer (codex)

## Created
2026-07-20T02:11:49Z

## Last Update
2026-07-29T21:07:34Z

## Blocked By
- TASK-260720-2qqq0w
- TASK-260720-jrrgw9
- BUG-260729-r0fe02
- BUG-260729-1o0m8f

## Blocks
- TASK-260720-38l1sy
- TASK-260720-3pvihp
- TASK-260720-z2z795
- TASK-260720-z9j4c9

## Checklist
- [ ] CI pins the reviewed immutable rc.4 protocol commit
- [x] Linux, macOS, and Windows exercise their platform-specific behavior
- [x] Race, vet, formatting, lint, and acceptance evidence are required
- [x] Candidate suite input is explicit and never advances or impersonates the qualified released pin
- [x] Keep every default committed protocol pin on the previous release; supply the candidate suite only through a non-default immutable input
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Composite proven: 356 product manifest entries byte-identical to the accepted TASK-260720-jrrgw9 candidate (cmp exit 0); overlay is CI/quality files only
- [x] Default pin lane and rc.5 candidate lane both exit 0 through .github/ci/test-gate.sh on the composite
- [x] Gate self-test 70/70 and ledger-consistency 49 rows across linux/darwin/windows both exit 0
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches
- [x] Ambiguous candidate_ref plus candidate_root dispatch fails before checkout or evidence stamping
- [x] Focused regression keeps ref-only and root-only candidate paths valid

## Notes
Cross-story release boundary from STORY-260720-21bsr2: candidate rc.4 tests may use an explicitly supplied CURATOR_CONFORMANCE_ROOT, but the committed curator-spec release pin must not move until TASK-260720-25d05o qualifies the actual published protocol release. TASK-260720-38l1sy audits the resulting pin and no-skip gate. Do not substitute a merely landed or reviewed but unreleased commit.
Cross-story checklist clarification 2026-07-20: inherited checklist item 1 uses stale pin language. It is superseded by the current description, scope, AC, and checklist items 4-5: candidate qualification uses an explicit immutable non-default input, while the committed suite pin remains on the previous release until TASK-260720-25d05o and TASK-260720-38l1sy.
NEXT-STEP DIRECTIVE 2026-07-29: Start only after TASK-260720-2qqq0w and TASK-260720-jrrgw9 are independently accepted. Build a fresh task-owned composite from their exact accepted deltas on top of TASK-260729-2kaopg. Update stale rc.4 execution wording to the accepted rc.5 schema-6 manager-worker-v1 candidate from TASK-260729-3nx97g while keeping every default committed release pin unchanged. CI must use Go 1.25.5 and an explicit immutable full candidate revision or CURATOR_CONFORMANCE_ROOT, run ubuntu/macos/windows platform cases plus a supported Linux race gate, preserve vet/gofmt/golangci-lint, and record exact candidate digest without a release claim. Supplemental native validation may use ssh relux and ssh win; ssh lev is available for several hours but currently has no Go binary, so use it only after an operator-approved absolute Go 1.25.5 GOROOT/trusted identity is provided. Never auto-install/download on remote hosts, accept ambient PATH, weaken platform assertions, stage, commit, publish or advance a pin.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-9ead43, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-9ead43)
ORCHESTRATOR REWORK DIRECTIVE 2026-07-29: Current producer outcome is not reviewable because it validates origin/main only and explicitly treats accepted compiled-build packages/controls as absent. The task requires a fresh composite on TASK-260729-2kaopg containing the exact accepted TASK-260720-2qqq0w and TASK-260720-jrrgw9 deltas. Rebuild and rerun CI evidence against that composite; do not downgrade DACL, resource-policy, readonly-source, godriver, buildcache/buildsource, transaction/cache/install/conformance, or platform requirements. Preserve released pin and candidate-only semantics. Supersede the first outcome before reviewer routing.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-703e35, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-703e35)
REWORK RUN-260729-703e35 in progress (developer). Fresh composite at .temp/TASK-260720-1pvfj5/rework/composite: git worktree at 17804ce + accepted TASK-260720-jrrgw9 candidate bytes (manifest 83b5df8b..., 359 entries, cmp exit 0 vs verifier4 post-run-manifest). Product bytes preserved byte-for-byte: 356 product manifest entries identical, cmp exit 0. Overlay is exactly 13 new .github/ci/ files plus modified .github/workflows/ci.yml, Makefile, README.md; zero product-code changes; .golangci.yml untouched; SPEC_PIN unchanged at 00b1688a (no promotion). Measured on the composite: default pin lane test-gate exit 0; candidate lane (rc.5 root, CI_REQUIRE_FULL_ROOT=1) exit 0; gofmt/go vet/make build/ledger-check/no-broad-suppression exit 0; gate-selftest 70 passed 0 failed. OPEN: golangci-lint v2.12.2 exits 1 with 4 findings in ACCEPTED product code (2x G115 ccj.go:211-212 and G602 journal.go:398 - both analysed as false positives on guarded branches/short-circuit index; 1 ineffassign in godriver test file, also present at v2.4.0 so pre-existing and not introduced by the version pin). Not fixable inside this task: the rework instruction forbids product-code mutation and the AC forbids broad suppression. Routed to review with evidence.
REWORK RUN-260729-703e35 COMPLETE (developer) — superseding outcome attached as TASK-260720-1pvfj5_rework-results.md; the prior TASK-260720-1pvfj5_results.md is now headed SUPERSEDED.

COMPOSITE: .temp/TASK-260720-1pvfj5/rework/composite = git worktree at 17804ce + accepted TASK-260720-jrrgw9 bytes. Base manifest 83b5df8b... 359 entries, cmp exit 0 vs verifier4 post-run-manifest. After overlay: 372 entries; product-only subset 356 entries, cmp exit 0 -> every accepted product byte preserved. Overlay is 13 new .github/ci files + ci.yml + Makefile + README.md. No Go file, no .golangci.yml, no fixture, no timeout, no pin touched. SPEC_PIN stays 00b1688a (rc.3), now declared once in env: and read by every job.

DESIGN: the run plans itself from the supplied conformance root. root-artifacts.tsv + suite-plan.sh split the module into served/deferred/excluded; pin root -> 33 served/7 deferred, rc.5 root -> 40 served/0 deferred. Candidate lane sets CI_REQUIRE_FULL_ROOT=1 so any deferral is fatal. Linux godriver exclusion is read from the root own conformance-claim-v3-qualification.json (until_task TASK-260728-1skseh), with default_excluded_on only for a pre-vector root, and is ASSERTED on the excluded runner by TestProbeRejectsAnUncoveredPlatformBeforeTheWorker (exit 0). Skip enforcement is two-tier: 49 ledger rows (37 darwin / 32 windows / 31 linux required) + 35 reason-regex classes; an unrecognised skip reason is fatal and every skip is recorded by name in skips-observed.tsv. ledger-consistency.sh proves each row against the real per-GOOS builds via go list, so Linux/Windows claims are checkable without a runner.

MEASURED (real exits): default pin lane 0; candidate lane 0; make race attempt1 2 / attempt2 0; gate-selftest 0 (70/70); ledger-check 0; no-broad-suppression 0; gofmt 0; go vet 0; make build 0; exclusion assert 0; workflow YAML shape 0; naming gate 0.

TWO OPEN FINDINGS IN ACCEPTED PRODUCT CODE, deliberately NOT fixed here (rework instruction forbids product-code mutation; AC forbids broad suppression):
1) golangci-lint v2.12.2 exit 1 — G115 x2 at internal/protocoljson/ccj.go:211,212 (guarded by character < 0x20) and G602 at internal/transaction/journal.go:398 (guarded by short-circuit index > 0) are false positives; ineffassign at internal/godriver/builddriver_positive_conformance_test.go:178 is real dead code and also fires at v2.4.0, so it predates the version pin. Checklist item 8 (Lint clean) is therefore UNCHECKED. Recommended fix for the file owner: line-scoped //nolint:gosec with named rule + reason, and delete the dead variable.
2) internal/godriver TestFingerprintCancellationStaysFailClosed/cancelled_between_the_walk_and_the_digest (fingerprint_equivalence_test.go:519) is an intermittent full-race-suite flake: it admits nil or toolchain_timeout but the racy go cancel() can also yield toolchain_mutated. 1 failure in 2 full -race runs; 0 failures in 185 targeted -race executions incl. -cpu 1,2,8. jrrgw9 verifier4 full race on byte-identical code was exit 0.

Checklist item 1 stays UNCHECKED: it asserts CI pins an rc.4 commit. No rc.4 pin exists and CI deliberately does not pin one — superseded by items 4-5 per the 2026-07-20 board clarification.

NATIVE COVERAGE: no Linux or Windows runner was reachable (ssh relux = Darwin x86_64 without Go; ssh win / ssh lev exit 255 timeout; chip/reluxts do not resolve). All Linux/Windows claims are derived from real per-GOOS builds and the protocol vector; native execution is deferred to the hosted runners, with six named producer confirmations in results section 7.

HANDOFF TO TASK-260720-38l1sy: pin under audit 00b1688a (protocol 1.0.0-rc.3, untagged, v1.0.0-rc.2-1-g00b1688), unmoved. Candidate evidence format and the measured rc.5 identity (manifest b6f56aac..., tree e6a13215..., 448 files) are in results section 8, together with a promotion pre-flight: a proposed pin must additionally make suite-plan report deferred=0.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260729-3aa19a, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260729-3aa19a)
Final integration provider routing 2026-07-29: Claude Opus RUN-260729-3aa19a emitted no first token for more than 90 seconds during the continuing provider outage and was cancelled without source or board artifact changes. Final integration is rerouted to a tracked Codex developer under the same exact precondition; independent review remains a separate reviewer run.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (codex) (run=RUN-260729-e0c95c, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260729-e0c95c)
FINAL INTEGRATION 2026-07-30: Started from the accepted 372-entry rework composite (prepatch manifest cmp exit 0). Applied accepted lint patch sha256 8a07c0b2...68062 and cancellation patch 462f1ff0...09c28 verbatim; both reverse-check exit 0 and all seven files cmp exit 0 against accepted blocker worktrees. Final manifest proof: accepted product 356 entries -> integrated 358 entries solely through 5 modified + 2 added patch-owned paths; zero removed; all 16 CI/quality overlay paths unchanged; post-run manifest cmp exit 0. Direct gates: candidate identity 0; gofmt 0; no-broad-suppression 0; pinned golangci-lint v2.12.2 0/0 issues; go vet ./... 0; go build ./... 0; deterministic godriver cancellation 0; default pin test-gate 0 (33 served/7 explicit deferred); rc.5 CI_REQUIRE_FULL_ROOT candidate gate 0 (40/0/0); exactly one full race test-gate 0 (33/7/0). Race diagnostic rg exit 1 expected no matches; rc.4 wording rg exit 1 expected no matches. Reused unchanged-script selftest 70/70 and ledger 49 rows. SPEC_PIN remains exactly once at 00b1688a...e967. Candidate manifest b6f56aac...04c, tree e6a13215...2fae, 448 files, candidate-only/no release claim. New outcomes: final-integration-results.md, final-integration-evidence.tgz, both accepted patch copies, final candidate identity. Checklist item 1 remains unchecked as stale/superseded per board notes; lint item is now checked. No stage, commit, publish, pin, vector, timeout, fixture, suppression, or unrelated product/CI mutation.
HANDOFF MECHANICS 2026-07-30: canonical task-board handoff exited 1 because it requires every checklist item and the only unchecked item is stale item 1 (rc.4 pin), which the mandatory final-integration instruction and existing board clarification explicitly require to remain unchecked. It was not falsely checked. All genuine DoD items are green and evidence is attached. The developer end status is therefore applied through the ordinary set_status(to-review) transition, preserving the stale item and its supersession evidence; the canonical handoff command is then retried as the required final board command and is expected to retain its truthful evidence-missing exit for that stale item only.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-e0c95c, pid=38293, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-380607, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-380607)
FINAL REVIEW 2026-07-30 — CHANGES REQUESTED. The exact final composite, accepted patch provenance, unchanged released pin, candidate-only rc.5 identity, and attached green gate evidence are verified. One P1 CI provenance defect remains: workflow_dispatch accepts candidate_ref and candidate_root together, verifies/checks out the ref, prefers/tests the root, then candidate-suite.sh stamps the unrelated ref as candidate_revision alongside the root digests. Required: fail the ambiguous combination or prove root-to-ref identity, plus a focused regression. Evidence and exact lines are in TASK-260720-1pvfj5_review-verdict.md. Route to to-dev; no heavy suite rerun is requested solely for this narrow validation fix.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-380607, pid=68680, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (codex) (run=RUN-260729-7239ee, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260729-7239ee)
CANDIDATE INPUT REWORK 2026-07-30: In .temp/TASK-260720-1pvfj5/rework/composite, added fail-closed exact-one-source validation before candidate checkout/evidence. Exactly three CI paths changed; accepted 374-entry manifest comparison exit 0 with zero product or unrelated changes. Focused both-input gate exits 1 expected-red; ref-only/root-only each exit 0; gate-selftest 74/74 exit 0; actionlint, bash -n, focused shellcheck, git diff check, patch reverse-check, and SPEC_PIN assertion exit 0. Default/candidate/full/race suites were not rerun per mandatory narrow-rework instruction; accepted heavy evidence is reused. Outcomes attached: candidate-input-rework results, patch, and 374-entry manifest. Default shellcheck across both scripts exits 1 only for two pre-existing info-level findings in untouched regions; no suppression added. Checklist item 1 remains stale/superseded and intentionally unchecked; SPEC_PIN remains 00b1688a9b2457ca397a0bb550acf47cad8ee967.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-7239ee, pid=72915, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-b056b2, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-b056b2)
FINAL FOCUSED REVIEW 2026-07-30 — ACCEPTED. The ambiguous candidate_ref plus candidate_root path now fails before candidate checkout and identity recording; ref-only and root-only remain valid. Independent replay: both-input exit 1 expected, single-input exits 0, gate-selftest 74/74, actionlint/bash-n/focused-shellcheck/patch-reverse/live-manifest checks all exit 0. Exact accepted-to-rework delta is only candidate-suite.sh, gate-selftest.sh, and ci.yml; all seven accepted product paths, SPEC_PIN 00b1688a...e967, candidate-only wording, and unrelated bytes are unchanged. Accepted heavy default/candidate/race/lint/vet/build/ledger evidence reused by byte identity. The unchecked rc.4-pin checklist line remains stale and superseded by the existing board note/current scope; acceptance must not advance that pin. Verdict artifact: TASK-260720-1pvfj5_candidate-input-final-review-verdict.md. Route: done; reviewer supplied no commit_ack.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-b056b2, pid=87751, exit=0)

## Precondition Resources
- [TASK-260720-1pvfj5_composite-rework.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_composite-rework.md) — Mandatory accepted-composite rework instructions
- [TASK-260720-1pvfj5_final-integration.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_final-integration.md) — Final Curator integration recipe after the two accepted product blockers
- [TASK-260720-1pvfj5_final-review.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_final-review.md) — Independent final-review boundaries; reuse heavy evidence and preserve release pin
- [TASK-260720-1pvfj5_candidate-input-rework.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-input-rework.md) — Focused final-review rework for candidate input provenance
- [TASK-260720-1pvfj5_candidate-input-final-review.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-input-final-review.md) — Independent focused review of candidate input provenance rework

## Outcome Resources
- [TASK-260720-1pvfj5_candidate-release-ci-gates.puml](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-release-ci-gates.puml) — PlantUML source separating candidate CI evidence from official released-suite pin promotion
- [TASK-260720-1pvfj5_candidate-release-ci-gates.svg](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-release-ci-gates.svg) — Rendered candidate CI and released-suite evidence gates
- [TASK-260720-1pvfj5_spawn-log_-implementer--developer--claude-_RUN-260729-9ead43.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_spawn-log_-implementer--developer--claude-_RUN-260729-9ead43.log) — System spawn log captured by task-board
- [TASK-260720-1pvfj5_results.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_results.md) — SUPERSEDED by TASK-260720-1pvfj5_rework-results.md — the RUN-260729-9ead43 outcome, produced against origin/main without the accepted compiled-build packages
- [TASK-260720-1pvfj5_final-verification.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_final-verification.log) — Verbatim gate transcript on branch c06aa1a: selftest, gofmt, vet, golangci-lint v2.12.2, suppression guard, naming gate, build, test-gate (real exit codes)
- [TASK-260720-1pvfj5_gate-origin-main.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_gate-origin-main.log) — Gate runs on origin/main 17804ce: pin lane exit 0, rc.5 candidate lane exit 0, race lane exit 0, platform-case gate green in all three
- [TASK-260720-1pvfj5_candidate-suite-identity.txt](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-suite-identity.txt) — Candidate suite identity from the shipped candidate-suite.sh against the rc.5 root; reproduces the accepted TASK-260729-3nx97g digests; stamped candidate-only, no release or conformance claim
- [TASK-260720-1pvfj5_spawn-log_-implementer--developer--claude-_RUN-260729-703e35.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_spawn-log_-implementer--developer--claude-_RUN-260729-703e35.log) — System spawn log captured by task-board
- [TASK-260720-1pvfj5_rework-results.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_rework-results.md) — Superseding rework outcome: composite provenance, root-derived suite plan, two-tier platform-case gate, measured lane exits, and the two open findings in accepted product code
- [TASK-260720-1pvfj5_ci-overlay.patch](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_ci-overlay.patch) — The complete CI/quality overlay against the accepted TASK-260720-jrrgw9 candidate: 13 new .github/ci files plus ci.yml, Makefile, README.md; zero product-code changes
- [TASK-260720-1pvfj5_verification.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_verification.log) — Consolidated verification transcript with the real exit code of every gate, including the two red ones
- [TASK-260720-1pvfj5_rework-evidence.tgz](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_rework-evidence.tgz) — Raw evidence packet: per-lane suite plans, platform-case reports, skips-observed ledgers, gate self-test, golangci-lint at both versions, both race attempts, composite manifest, candidate identity
- [TASK-260720-1pvfj5_spawn-log_-implementer--developer--claude-_RUN-260729-3aa19a.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_spawn-log_-implementer--developer--claude-_RUN-260729-3aa19a.log) — System spawn log captured by task-board
- [TASK-260720-1pvfj5_spawn-log_-implementer--developer--codex-_RUN-260729-e0c95c.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_spawn-log_-implementer--developer--codex-_RUN-260729-e0c95c.log) — System spawn log captured by task-board
- [TASK-260720-1pvfj5_final-integration-results.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_final-integration-results.md) — Final accepted-composite provenance, exact seven-path delta proof, gate ledger, candidate identity, and released-pin handoff
- [TASK-260720-1pvfj5_final-integration-evidence.tgz](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_final-integration-evidence.tgz) — Raw final integration evidence: direct gate logs, manifests, immutable patches, candidate identity, and reused script-identical selftest/ledger logs
- [TASK-260720-1pvfj5_accepted-lint-fix.patch](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_accepted-lint-fix.patch) — Verbatim independently accepted lint patch integrated into the composite; sha256 8a07c0b239548235aea7dfa05fdb1d1cb2926971d4444d3435a9e6f8da368062
- [TASK-260720-1pvfj5_accepted-cancellation-fix.patch](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_accepted-cancellation-fix.patch) — Verbatim independently accepted godriver cancellation patch integrated into the composite; sha256 462f1ff0326f74540eeb2815cc80542c55f47b35c6b1baef17b80b8815709c28
- [TASK-260720-1pvfj5_final-candidate-suite-identity.txt](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_final-candidate-suite-identity.txt) — Explicit rc.5 root identity stamped candidate-only with expected manifest and tree digests; no release or conformance claim
- [TASK-260720-1pvfj5_spawn-log_-reviewer--reviewer--codex-_RUN-260729-380607.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_spawn-log_-reviewer--reviewer--codex-_RUN-260729-380607.log) — System spawn log captured by task-board
- [TASK-260720-1pvfj5_review-verdict.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_review-verdict.md) — Independent final review: changes requested for ambiguous candidate identity inputs
- [TASK-260720-1pvfj5_spawn-log_-implementer--developer--codex-_RUN-260729-7239ee.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_spawn-log_-implementer--developer--codex-_RUN-260729-7239ee.log) — System spawn log captured by task-board
- [TASK-260720-1pvfj5_candidate-input-rework-results.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-input-rework-results.md) — Focused provenance rework outcome: exact three-path delta, 74/74 self-test, static checks, manifest proof, and reused heavy evidence
- [TASK-260720-1pvfj5_candidate-input-rework.patch](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-input-rework.patch) — Focused three-file patch rejecting ambiguous candidate inputs and adding regression coverage
- [TASK-260720-1pvfj5_candidate-input-rework-manifest.txt](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-input-rework-manifest.txt) — 374-entry post-rework composite manifest used for exact accepted-to-rework byte comparison
- [TASK-260720-1pvfj5_spawn-log_-reviewer--reviewer--codex-_RUN-260729-b056b2.log](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_spawn-log_-reviewer--reviewer--codex-_RUN-260729-b056b2.log) — System spawn log captured by task-board
- [TASK-260720-1pvfj5_candidate-input-final-review-verdict.md](file://TASK-260720-1pvfj5/TASK-260720-1pvfj5_candidate-input-final-review-verdict.md) — Independent focused final review: accepted candidate-input provenance rework with manifest, ordering, regression, and reused heavy-evidence audit
