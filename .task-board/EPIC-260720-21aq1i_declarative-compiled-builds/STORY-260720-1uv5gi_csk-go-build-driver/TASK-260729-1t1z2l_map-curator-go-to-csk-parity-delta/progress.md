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
- [x] Map every existing CocoaSkills Go task to concrete Curator source, tests and accepted protocol surfaces
- [x] Identify adaptation boundaries, critical-path order, macOS/Windows gates and exact prerequisites
- [x] Attach a task-scoped parity-map artifact with repository and revision provenance
- [x] Make no product-code, staging, commit or publication changes
- [x] Findings written to file
- [x] Key aspects highlighted
- [x] Fact-checking performed — claims verified, sources cited
- [x] Findings linked on the board as a new task-scoped outcome resource
- [x] All questions from task description answered
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [analyst] researcher (codex) (run=RUN-260728-a668f6, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-260728-a668f6)
Research handoff 2026-07-29: attached TASK-260729-1t1z2l_curator-go-to-csk-parity-delta.md (SHA-256 742b767073578159b1eb4d6f0ba7849a004378a66c8fbd94bfb5d39334b14850), byte-identical to .research/260729_curator-go-to-csk-parity-delta.md. It maps all 17 CocoaSkills tasks to Curator source/tests/protocol surfaces, critical path, adaptation boundaries, macOS/Windows/Linux gates, and exact prerequisites. Material anomaly: CocoaSkills briefs still say rc.4 while accepted Curator executes the rc.5 manager-worker-v1 amendment; owner must explicitly retarget (recommended) or authorize literal rc.4 publication. Validation: exact 17-ID and 17-row checks exit 0; cited resource existence and whitespace checks exit 0; three remote-head checks exit 0. Full Curator go test ./... is honestly red/terminated exit 143 after internal/install hit no space left on device. Curator product scope and CocoaSkills remained unchanged/clean; no staging, commit, pin or publication. Logbook entry 2026-07-29 0401 records the findings.
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-260728-a668f6, pid=78262, exit=0)
REVIEW CYCLE 1 DIRECTIVE: Independently verify the Curator-to-CocoaSkills Go parity map without editing product or producer artifacts. Re-resolve repository revisions and check that all 17 existing CocoaSkills tasks appear exactly once and map to concrete accepted Curator source, tests, protocol surfaces and adaptation boundaries. Validate the proposed critical-path ordering, exact prerequisites, macOS-primary and Windows-SSH gates, later Linux qualification, local/external source parity, toolchain preflight and release/pin boundaries. Challenge the material rc.4-vs-accepted-rc.5 manager-worker-v1 mismatch against accepted TASK-260729-1kq1rd: state precisely which CocoaSkills briefs/dependencies must be retargeted after owner approval and reject any fabricated or premature pin. Reproduce focused cited checks and assess the reported Curator full-test exit 143/no-space event against current disk evidence; do not accept it as a product failure without a controlled replay, and do not conceal it if unreproducible. Publish exact ACCEPTED or CHANGES REQUESTED evidence and route accordingly. No product edits, staging, commit, publish, pin or host install.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-b41f06, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-b41f06)
REVIEW CYCLE 1 CHANGES REQUESTED 2026-07-29. The 17-row source/test map and DAG are verified, but the artifact incorrectly treats rc4 canonical input/key/receipt identity as frozen into rc5, omits exact retargets for metadata and manager-worker-v1 execution, and requires Linux in the current rc5 gate although the accepted inventory supports macOS/Windows only. Windows SSH is reachable but Go is not on PATH. Detailed evidence and required rework are in TASK-260729-1t1z2l_review-verdict-cycle-1.md. Route to analysis for artifact correction and another reviewer cycle; no product changes, staging, commits, publication, pins, or host installs.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-b41f06, pid=9062, exit=0)
REWORK CYCLE 1 DIRECTIVE (queued): Preserve the independently verified 17-task coverage, Curator path/test mappings, live CocoaSkills DAG, repository provenance and no-publication/no-pin boundary. Correct the parity artifact and logbook only. Distinguish top-level schema-6 declaration bytes from rc.5 canonical behavior: require execution_policy=manager-worker-v1, key sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b, receipt sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd, and the rc.4 input/key/receipt as negative non-alias evidence. Enumerate exact retargets for TASK-260720-2dnqw2, 2g21eg, 12r55p, akf5kh, 3pemm6, 3s27te and th0jdi as specified in the cycle-1 verdict; do not claim all 17 briefs contain literal rc.4. Make macOS primary and ssh win the current rc.5 gates, remove Linux success from the current 17-task chain, name the later Linux ownership/gap, and record that win is reachable but has no Go on PATH so Windows qualification requires an approved supported Go install or operator-trusted absolute root plus identity evidence. Keep Go-specific TASK-260720-3j8pp5 separate from later generic TASK-260728-1j72zq. Preserve truthful focused/full-test and disk/signing attribution. Attach revised task-scoped artifact and exact checks, then hand off for fresh review. No product edits, dependency mutations, host install, stage, commit, publish or pin. Start only when an Opus slot is free after current language and Curator main-path producers.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [analyst] researcher (claude) (run=RUN-260729-c6022a, max_parallel=20)
spawn run started: [analyst] researcher (claude) (run=RUN-260729-c6022a)
Linux host ssh lev is temporarily available. Keep current parity-map rework scoped; record Linux as an available later qualification surface, not a reason to expand Rust/Swift/Kotlin implementations or delay current Curator/CocoaSkills Go critical path.
REWORK CYCLE 1 COMPLETE 2026-07-29. Revision 2 of the parity artifact attached (SHA-256 be77d15bc99e539ea4f1f66f027b7f5e99ca3262467e9ceb36873c9012012af4), byte-identical to .research/260729_curator-go-to-csk-parity-delta.md. All four cycle-1 corrections applied and two further defects found. (1) rc.5 canonical identity separated from frozen schema-6 declaration bytes: required execution_policy=manager-worker-v1, key sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b, receipt sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd, with rc.4 3fcd714a and reserved-hardened 13736230 as negative non-alias cases; all three re-derived independently in Python and matched, aliases=false. (2) Exact retargets enumerated for 2dnqw2, 2g21eg, 12r55p, akf5kh, 3pemm6, 3s27te, th0jdi; literal rc.4 wording confirmed in exactly two briefs (12r55p, 3s27te), not all 17. (3) macOS primary and ssh win are the current rc.5 gates; Linux success removed from the 17-task chain and ownership named -- rc.5 conformance-claim-v3-qualification.json declares linux excluded until_task TASK-260728-1skseh (parent STORY-260728-1eye8p), with TASK-260728-1e6811 covering deferred toolchain preflight; flagged that no task yet owns a full local go-v1 Linux lifecycle gate. (4) win reachable at Windows NT 10.0.19045.0 (exit 0) but ssh win go version exits 1, no Go on PATH; Windows qualification needs an approved supported Go install or operator-trusted absolute root plus identity evidence. TASK-260720-3j8pp5 kept separate from downstream generic TASK-260728-1j72zq. NEW FINDING A: cycle-1 cited non-canonical worktrees; canonical snapshots per accepted TASK-260729-1kq1rd are q5oy3o (rc.4, manifest 70cd2741) and 2kp3tv (rc.5, manifest 9ba9b8ec). NEW FINDING B: the accepted rc.5 publication target publishes NO vectors/build-drivers.json and NO expected/build-driver/ (0 of 422 manifest entries); Curator TestCandidateBuildMetadataArtifacts SKIPs against rc.5 and both candidate tests SKIP against rc.4, so no root exercises byte-exact build-driver goldens under rc.5 today -- a hard prerequisite for 12r55p/2dnqw2/3pemm6 and a Curator coverage hole. Verification: exact-17-ID, section-4 row-count=17, no-duplicate, cited-path existence and no-trailing-whitespace checks all exit 0; focused Curator batch 1 (buildmeta buildsource skillspec skillcheck marker whitelist godriver buildcache) exit 0; batch 2 exit 1 on a stale go-build cache object left by my own 120s tool timeout, transaction retry exit 0; both conformance-root probes exit 0. No repository-wide go test ./... green run is claimed. Curator product surfaces and CocoaSkills remain clean and unstaged; no staging, commit, pin, publication or host install. Logbook entry 2026-07-29 0455 records the findings.
ARTIFACT DIGEST CORRECTION 2026-07-29: the final revision-2 artifact SHA-256 is 04ff3e1aa1e5b995f6157d5cdf09aa885007b6c5903ca4a50de436c335b19e8e (supersedes the be77d15b digest recorded a moment earlier). The board resource TASK-260729-1t1z2l_curator-go-to-csk-parity-delta.md is byte-identical to .research/260729_curator-go-to-csk-parity-delta.md at that digest. The only delta is section 7.2, which now records the temporarily reachable Linux host ssh lev as an available later qualification surface per the operator directive on this run -- explicitly not a reason to expand the current 17-task chain or delay the Go critical path, and still owned by TASK-260728-1skseh / TASK-260728-1e6811. Post-edit re-verification: exact-17-ID exit 0, section-4 row count 17 exit 0, trailing-whitespace count 0. Five rework evidence logs attached: rework-conformance-rc5-root.log, rework-conformance-rc4-root.log, rework-focused-go-01.log, rework-focused-go-02.log, rework-focused-go-03-transaction-retry.log.
agent completed: [analyst] researcher (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-c6022a, pid=50898, exit=0)
REVIEW CYCLE 2 DIRECTIVE: Independently review revision 2 without editing producer artifacts or product code. Verify the preserved exact 17-task map and repository provenance; recompute rc.5 manager-worker-v1 execution identity key sha256:529370122ae11e2e961d5265b1a020e046bcd43165b2eb96b05e73a51187ac9b and receipt sha256:919fbbad8e6ce95532219fd952c2309d0d7026f85209650508fd6834af4020cd, and require the rc.4 input/key/receipt as a negative non-alias. Verify exact retargets TASK-260720-2dnqw2, TASK-260720-2g21eg, TASK-260720-12r55p, TASK-260720-akf5kh, TASK-260720-3pemm6, TASK-260720-3s27te and TASK-260720-th0jdi. Confirm macOS plus Windows are current gates, Linux via ssh lev is available but later/non-gating, and Windows currently has no Go on PATH. Confirm Go-specific TASK-260720-3j8pp5 remains distinct from generic TASK-260728-1j72zq. Challenge the newly reported missing rc.5 byte-exact build-driver golden coverage and deferred Linux ownership against canonical accepted snapshots, but do not mutate dependencies or expand implementation scope. Validate truthful test/disk/signing evidence and reject fabricated pins, publication or unsupported green claims. Publish exact ACCEPTED or CHANGES REQUESTED evidence and route accordingly. No product edits, staging, commit, publish, pin or host install.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-847f2e, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-847f2e)
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-847f2e, pid=73453, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260729-1t1z2l_spawn-log_-analyst--researcher--codex-_RUN-260728-a668f6.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_spawn-log_-analyst--researcher--codex-_RUN-260728-a668f6.log) — System spawn log captured by task-board
- [TASK-260729-1t1z2l_curator-go-to-csk-parity-delta.md](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_curator-go-to-csk-parity-delta.md) — Revision 2: rc.5 execution-policy identity vs frozen schema-6 bytes, exact 7-task retargets, macOS/Windows current gates with owned deferred Linux (incl. ssh lev as a later surface), rc.5 build-driver golden gap, 17-task parity map
- [TASK-260729-1t1z2l_spawn-log_-reviewer--reviewer--codex-_RUN-260729-b41f06.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_spawn-log_-reviewer--reviewer--codex-_RUN-260729-b41f06.log) — System spawn log captured by task-board
- [TASK-260729-1t1z2l_review-go-test-01.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-go-test-01.log) — Controlled full Curator replay: host Git SSH-signing contamination, no storage error
- [TASK-260729-1t1z2l_review-go-test-02.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-go-test-02.log) — Partial isolated Curator replay stopped when an unrelated concurrent gate consumed disk headroom
- [TASK-260729-1t1z2l_review-go-focused-01.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-go-focused-01.log) — Green focused Curator parity package tests
- [TASK-260729-1t1z2l_review-csk-pytest-01.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-csk-pytest-01.log) — CocoaSkills baseline full pytest log; one TMPDIR-inside-Git environmental failure
- [TASK-260729-1t1z2l_review-csk-pytest-focused-02.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-csk-pytest-focused-02.log) — Focused CocoaSkills rerun outside a Git worktree proving the sole full-run failure environmental
- [TASK-260729-1t1z2l_review-csk-mypy-02.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-csk-mypy-02.log) — Green official strict CocoaSkills mypy gate
- [TASK-260729-1t1z2l_review-verdict-cycle-1.md](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-verdict-cycle-1.md) — Reviewer changes requested: rc5 identity bytes, exact brief retargets, macOS/Windows current gates, deferred Linux, platform/toolchain readiness
- [TASK-260729-1t1z2l_spawn-log_-analyst--researcher--claude-_RUN-260729-c6022a.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_spawn-log_-analyst--researcher--claude-_RUN-260729-c6022a.log) — System spawn log captured by task-board
- [TASK-260729-1t1z2l_rework-conformance-rc5-root.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_rework-conformance-rc5-root.log) — Curator candidate conformance vs canonical rc.5 root: build-driver artifacts SKIP, receipt schema case PASS (exit 0)
- [TASK-260729-1t1z2l_rework-conformance-rc4-root.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_rework-conformance-rc4-root.log) — Curator candidate conformance vs canonical rc.4 root: both candidate tests SKIP, pre-revision root (exit 0)
- [TASK-260729-1t1z2l_rework-focused-go-01.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_rework-focused-go-01.log) — Green focused Curator parity packages batch 1 (exit 0)
- [TASK-260729-1t1z2l_rework-focused-go-02.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_rework-focused-go-02.log) — Focused batch 2: transaction build failed on a stale go-build cache object after an interrupted run (exit 1); seven other packages ok
- [TASK-260729-1t1z2l_rework-focused-go-03-transaction-retry.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_rework-focused-go-03-transaction-retry.log) — internal/transaction retry green (exit 0), confirming the batch-2 failure was a transient build-cache artifact
- [TASK-260729-1t1z2l_spawn-log_-reviewer--reviewer--codex-_RUN-260729-847f2e.log](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_spawn-log_-reviewer--reviewer--codex-_RUN-260729-847f2e.log) — System spawn log captured by task-board
- [TASK-260729-1t1z2l_review-verdict-cycle-2.md](file://TASK-260729-1t1z2l/TASK-260729-1t1z2l_review-verdict-cycle-2.md) — Reviewer cycle 2 accepted verdict with independent parity, protocol, platform, test, and no-change evidence

## Created
2026-07-28T23:30:33Z

## Last Update
2026-07-29T01:12:35Z

## Assigned To
[reviewer] reviewer (codex)
