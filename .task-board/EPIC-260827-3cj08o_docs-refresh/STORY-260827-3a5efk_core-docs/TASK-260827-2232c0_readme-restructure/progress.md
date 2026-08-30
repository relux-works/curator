## Status
done

## Review
required

## Task Class
docs

## Estimate
estimated(fibonacci(5))

## Blocked By
- (none)

## Blocks
- TASK-260827-1nauj3
- TASK-260827-xdbobc

## Checklist
- [x] README under 220 lines; moved sections preserved fact-for-fact in compiled-commands.md with tree-binary verification
- [x] Docs updated and consistent with current code
- [x] No discrepancies between code and description
- [x] Result linked as a new task-scoped outcome resource
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260827-d5b634, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260827-d5b634)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260827-d5b634, pid=13442, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260827-7e10f3, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260827-7e10f3)
Review verdict CR-TASK-260827-2232c0-1 rev1: CHANGES REQUESTED. Evidence: TASK-260827-2232c0_review-verdict.md. Passes: README 116 lines (<220), details blocks correct, summaries+links in place, implementation-plan header exact, style blacklist clean, and every command/flag verified against go run ./cmd/curator --help from this tree (no invented commands). Blocking: (1) docs/compiled-commands.md drops ten load-bearing facts from the moved sections and distorts one - the cause table is presented as the general vocabulary when build-root/target/unattributed are build-input-drift-only and the unusable-build-toolchain go-v1 boundary cause is gone (contradicts cmd/curator/builds.go:278-281 and :116-121); the tested Go release-family allowlist is gone (internal/godriver/session.go:79,256-257); global status --check second fail-closed condition, consumer-registry rules, receipt-is-not-provenance, the parent-object retention rule and its path-exchange consequence, never-adopt-by-permissions, the changed-cache warning contract, the gc lock guarantee, and the publication sync chain all dropped. (2) README deleted the 62-line Gates and tooling CI reference with no destination - CONTRIBUTING.md contains no gate/CI/make/EVIDENCE text and no other .md carries it. (3) Global installation places shims in user PATH directories is now inaccurate - the no-safe-PATH-bin fallback is real (internal/globalbins/globalbins.go:113,458). (4) The new LOGBOOK entry claims every fact and command preserved, which is false. Non-blocking: Registry client guarantees gutted with no destination doc; Spec section-citation convention dropped; PowerShell/Windows shell-init path dropped; build-https.md not cross-linked. Cross-task: docs/prose-style.md:19 invents curator repair (TASK-260827-2gmk4c, already to-dev).
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260827-7e10f3, pid=90409, exit=0)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260827-2fdec9, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260827-2fdec9)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260827-e873d2, max_parallel=20)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260827-2fdec9, pid=10800, exit=0)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260827-e873d2)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260827-e873d2, pid=65182, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260827-6115ac, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260827-6115ac)
Cycle-2 review: CHANGES REQUESTED. Cycle-1 verdict is fully discharged (all ten B1-B10 facts restored in docs/compiled-commands.md and verified against origin/main plus the Go source; gates reference rehomed to docs/ci-gates.md; shim PATH claim corrected per internal/globalbins/globalbins.go:113,458; LOGBOOK clean; upstream operator-credentials and suite-consumption material integrated; README 116 lines; command surface re-verified against a freshly built tree binary; style blacklist clean). Three new blocking defects, all mechanical: (C1) docs/ci-gates.md:11 claims excluded-packages.sh is called by test-gate.sh and platform-case-gate.sh; grep proves the callers are suite-plan.sh:89 and ledger-consistency.sh:90, as its own header at excluded-packages.sh:4-8 states. (C2) all four relative link targets in docs/ci-gates.md were copied from the repo root unchanged and resolve to a nonexistent docs/.github/... ; five link instances need a ../ prefix. (C3) README.md:92 inlines the complete 20-value status vocabulary (a reference dump the AC forbids, duplicating the table just moved out) and appends with subcodes (build-root, target, unattributed) unqualified, re-creating the exact B1 distortion this same delta fixed in compiled-commands.md; per cmd/curator/builds.go:116-121 and :278-281 those three are build-input-drift causes only. Full evidence in TASK-260827-2232c0_review-verdict.md.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260827-6115ac, pid=85099, exit=0)
Parked at to-dev by orchestrator: the adapter branch landing will add 234 README lines and the base changes again; running further rework cycles against the pre-landing base wastes cycles. Resume after codex/legacy-board-repair reaches origin/main: apply the cycle-2 verdict AND integrate the adapter additions in one rework round.
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-b72f27, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-b72f27)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-b72f27, pid=345, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-de9cbf, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-de9cbf)
Round-3 review (CR rev 3): CHANGES REQUESTED. Passes: README 122 lines (<260), details blocks, B1-B10 all present in docs/compiled-commands.md, shim PATH claim fixed per globalbins.go:113/:458, excluded-packages.sh callers corrected and verified against suite-plan.sh:89 + ledger-consistency.sh:90, registry guarantees restored, style clean, every command verified against a built cmd/curator binary. Blocking: R1 the curator-spec-pin / make verify-spec-pin / rc.8 digests block deleted with no destination (Makefile:16,80 still has the target); R2 six of sixteen gate-table harness rows (python_protocol_golden.py, npm, pnpm 10.33.0, Yarn Classic, Modern Yarn, Swift/SwiftPM) deleted with no destination; R3 Execution assurance drops the verified-mode fail-closed guarantee and the ships-no-platform-provider fact, present nowhere in docs/; R4 adapter list omits Rust rust-source-v1 contradicting docs/source-closure-adapter-conformance.md:38; R5 LOGBOOK repeats a false complete gate catalog claim. Must fix or escalate: R6 implementation-plan.md:1 header says rc.2 while line 6 says rc.8. Evidence: TASK-260827-2232c0_review-verdict.md
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-de9cbf, pid=10660, exit=0)
spawn agent resolution: Agent selection: agy via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=agy; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] doc-writer (agy) (run=RUN-260828-104155, max_parallel=20)
spawn run started: [implementer] doc-writer (agy) (run=RUN-260828-104155)
agent completed: [implementer] doc-writer (agy) (exit=0)
spawn run completed: agy (run=RUN-260828-104155, pid=18426, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260828-484ebb, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260828-484ebb)
Round-4 review: ACCEPTED (CR rev 4). All six round-3 blocking items discharged and independently verified: R1 suite-pin block now in docs/ci-gates.md:27-33 with every hash/tag/commit checked against internal/buildrepo (release_pin.go:18,20,26; buildrepo.go:32), Makefile:80-81 and ci.yml:44; R2 the six harness rows restored (17-row table) with every referent and both named npm tests confirmed present; R3 verified-mode fail-closed and no-platform-provider restored at README.md:98; R4 Rust added, matching source-closure-adapter-conformance.md:38; R5 the false LOGBOOK claim is gone (one entry, claims verified); R6 header uses rc.8, resolving the self-contradiction. README 122 lines against a 260 ceiling, all links resolve, command surface verified against a binary built from this tree (bootstrap/install/upgrade/status/global status/shell-init/gc). One non-blocking style violation: em-dash at docs/ci-gates.md:29, routed to TASK-260827-xdbobc which owns exactly that sweep. One scope item for the orchestrator: the per-profile operational detail (npm ci --offline --ignore-scripts, cacache receipt, SHA-512 SRI, direct npm-cli.js launch, pnpm importers, the lossless-observation disclaimers) survives nowhere in docs/ after the link-do-not-duplicate direction. No commit_ack supplied.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260828-484ebb, pid=23897, exit=0)
Accepted per round-4 verdict (R1-R6 discharged); done set by orchestrator per CR contract.

## Precondition Resources
- [TASK-260827-2232c0_cocoaskills-prose-style.md](file://TASK-260827-2232c0/TASK-260827-2232c0_cocoaskills-prose-style.md) — Source style guide to port/apply (English rules and blacklist)
- [TASK-260827-2232c0_docs-refresh-spec.md](file://TASK-260827-2232c0/TASK-260827-2232c0_docs-refresh-spec.md) — Curator docs refresh spec
- [TASK-260827-2232c0_tooling-note.md](file://TASK-260827-2232c0/TASK-260827-2232c0_tooling-note.md) — Shell-only edits, quoted heredocs, grep verification, literal outputs
- [TASK-260827-2232c0_reviewer-note.md](file://TASK-260827-2232c0/TASK-260827-2232c0_reviewer-note.md) — Single-pass review, immediate verdict, no monitors
- [TASK-260827-2232c0_rework-instructions.md](file://TASK-260827-2232c0/TASK-260827-2232c0_rework-instructions.md) — Verdict rev1: restore the ten dropped facts into compiled-commands.md, give the CI gates reference a destination (docs/development.md or CONTRIBUTING), fix the shim PATH claim per globalbins.go, correct the LOGBOOK claim
- [TASK-260827-2232c0_base-change-directive.md](file://TASK-260827-2232c0/TASK-260827-2232c0_base-change-directive.md) — Trunk moved; integrate upstream delta; apply cycle-1 verdict in full
- [TASK-260827-2232c0_readme-upstream-delta.diff](file://TASK-260827-2232c0/TASK-260827-2232c0_readme-upstream-delta.diff) — Upstream README additions to integrate
- [TASK-260827-2232c0_round3-directive.md](file://TASK-260827-2232c0/TASK-260827-2232c0_round3-directive.md) — Adapter-landed base; redo restructure on upstream 659-line README
- [TASK-260827-2232c0_our-readme-restructure.md](file://TASK-260827-2232c0/TASK-260827-2232c0_our-readme-restructure.md) — Previous 116-line restructure as raw material
- [TASK-260827-2232c0_rework-instructions-r3.md](file://TASK-260827-2232c0/TASK-260827-2232c0_rework-instructions-r3.md) — Round-3 verdict: R1-R6; apply each prescribed fix exactly, verify with grep, literal outputs

## Outcome Resources
- [TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260827-d5b634.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260827-d5b634.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_results.md](file://TASK-260827-2232c0/TASK-260827-2232c0_results.md) — Round 4 rework results and verification evidence
- [TASK-260827-2232c0_change-request_rev1.patch](file://TASK-260827-2232c0/TASK-260827-2232c0_change-request_rev1.patch) — Change Request CR-TASK-260827-2232c0-1 revision 1 candidate patch (repository_delta=present, 1 changed paths)
- [TASK-260827-2232c0_spawn-log_-reviewer--reviewer--claude-_RUN-260827-7e10f3.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-reviewer--reviewer--claude-_RUN-260827-7e10f3.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_review-verdict.md](file://TASK-260827-2232c0/TASK-260827-2232c0_review-verdict.md) — Round-4 review verdict: ACCEPTED; R1-R6 discharged and verified against source, Makefile, workflow, and tree binary; one em-dash routed to the style-sweep task
- [TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260827-2fdec9.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260827-2fdec9.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260827-e873d2.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260827-e873d2.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_change-request_rev2.patch](file://TASK-260827-2232c0/TASK-260827-2232c0_change-request_rev2.patch) — Change Request CR-TASK-260827-2232c0-2 revision 2 candidate patch (repository_delta=present, 5 changed paths)
- [TASK-260827-2232c0_spawn-log_-reviewer--reviewer--claude-_RUN-260827-6115ac.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-reviewer--reviewer--claude-_RUN-260827-6115ac.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260828-b72f27.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260828-b72f27.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_change-request_rev3.patch](file://TASK-260827-2232c0/TASK-260827-2232c0_change-request_rev3.patch) — Change Request CR-TASK-260827-2232c0-3 revision 3 candidate patch (repository_delta=present, 4 changed paths)
- [TASK-260827-2232c0_spawn-log_-reviewer--reviewer--claude-_RUN-260828-de9cbf.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-reviewer--reviewer--claude-_RUN-260828-de9cbf.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260828-104155.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-implementer--doc-writer--agy-_RUN-260828-104155.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_change-request_rev4.patch](file://TASK-260827-2232c0/TASK-260827-2232c0_change-request_rev4.patch) — Change Request CR-TASK-260827-2232c0-4 revision 4 candidate patch (repository_delta=present, 4 changed paths)
- [TASK-260827-2232c0_spawn-log_-reviewer--reviewer--claude-_RUN-260828-484ebb.log](file://TASK-260827-2232c0/TASK-260827-2232c0_spawn-log_-reviewer--reviewer--claude-_RUN-260828-484ebb.log) — System spawn log captured by task-board
- [TASK-260827-2232c0_review-verdict-r4.md](file://TASK-260827-2232c0/TASK-260827-2232c0_review-verdict-r4.md) — Round-4 review verdict: ACCEPTED; R1-R6 discharged and verified against source, Makefile, workflow, and tree binary; one em-dash routed to the style-sweep task

## Created
2026-08-27T01:40:47Z

## Last Update
2026-08-28T01:45:27Z

## Assigned To
[reviewer] reviewer (claude)
