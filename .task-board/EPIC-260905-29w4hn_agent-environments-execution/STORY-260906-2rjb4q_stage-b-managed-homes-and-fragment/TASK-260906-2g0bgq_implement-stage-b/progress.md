## Status
integrating

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(21))

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Managed homes provisioned with surfaces, forms and marker records incl. copy reasons; seeds and passthrough strategies per adapter with the isolation matrix and liveness rows
- [x] Read-only env resolve: lock-free verification, environment_home_stale with reasons and no fragment, --repair under the mutation lock with a distinct lock diagnostic
- [x] Closed launch-env-fragment-v1 (lock_sha256, precedence, env, system_prompt, mcp with env_names union and channel descriptor, path_prepend) in json/env/shell with the §10.3 boundary enforced
- [x] MCP channel files per adapter incl. the codex fixed layer path and the package allowlist; curator run umbrella dispatch; env status matrix rows
- [x] referenced-*, system-prompt-composed and mcp-* sets pass byte for byte; their stage-deferred skips removed; every skip in a registered truthful class; gates and platform-case gate green; signed commits; report attached
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
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] Gate, refusal, validation, authorization, and attestation behavior attacked, not read — positive-path-only evidence is not accepted
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: muse via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=muse; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (muse) (run=RUN-260906-24258e, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260906-24258e)
Stage (b) implemented on feat/agent-environments-stage-b (13 signed commits). Findings, mutant table, gate outputs, and deferred items live in attached TASK-260906-2g0bgq_drafting-report.md (no logbook CLI in this environment; the board resource is the durable record). Test-gate: go test exit=0, platform-case gate exit=0 (darwin lane; linux/windows rows ledger-checked, CI-owned).
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-260906-24258e, pid=69729, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260906-ab4d71, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260906-ab4d71)
Reviewer cycle 1 (RUN-260906-ab4d71) on feat/agent-environments-stage-b @ 73174cc3: CHANGES REQUESTED, repeat-of: none. 3 blocking, 3 major, 5 minor. Blocking: (F1) claude_code on macOS below the pinned 2.1.261 refuses every isolation value including the unconfigured default, so no managed home can be provisioned - internal/envregistry/envregistry.go:348-360, production call site internal/envprofile/managed.go:272; (F2) byte drift of a linked manager-authored surface is undetectable and env resolve still emits a fragment - the root context for codex_cli/opencode/pi, the system-prompt file and the MCP channel file link to <manager>/profiles/<p>/rendered/... which is NOT an immutable store entry, so the SS10.1 link-identity exemption does not apply and SS8.4 "a target whose bytes fail the recorded hash" is unimplemented; (F3) golangci-lint v2.12.2 (the ci.yml pin) exits 1 on internal/envfragment/fragment_schema_test.go:132 while the drafting report attests exit 0. Major: (F4) a NARROWING mutant of the SS10.3 root-containment gate (root -> its parent) survives envfragment+envprofile+cmd/curator; (F5) the SS8.1 always-copied claude_code root context has no test under the referenced form - disabling that branch survives the suite; (F6) checkClaudeProject tests project-entry presence, not hasClaudeMdExternalIncludesApproved, so a referenced home whose key the tool dropped resolves as current and every @-include is silently discarded. Empty repository_delta is CORRECT by brief design (implementation lives in the curator repo). Verified green independently: build/vet/gofmt, gate-selftest 81/0, test-gate go test exit=0 + platform-case gate exit=0 with 19 skips all in registered classes and zero stage-deferred, ledger-consistency 126 rows across 3 GOOS, -race on all touched packages, conformance 19/19 environments materialization cases byte-exact with zero skips at authority f39f4a9, 12 signed commits by the human identity. Detail: TASK-260906-2g0bgq_review-findings-stage-b-1.md; probes: TASK-260906-2g0bgq_review-probes-stage-b-1.tar.gz.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260906-ab4d71, pid=95337, exit=0)
spawn agent resolution: Agent selection: muse via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=muse; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (muse) (run=RUN-260906-76e22d, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260906-76e22d)
Rework 1 consumes the only outstanding verdict (cycle 1 CHANGES REQUESTED, F1-F6 + F7-F10): 5 signed commits 70e872b5, 198a45b0, 0933ea9f, 0ee00d19, 7bca4d4b on feat/agent-environments-stage-b over rebased origin/main 7320bc2a; evidence in TASK-260906-2g0bgq_rework-report-1.md. No newer verdict exists to route.
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-260906-76e22d, pid=1719, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260906-5ed60e, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260906-5ed60e)
Reviewer cycle 2 (RUN-260906-5ed60e) at 7bca4d4b: CHANGES REQUESTED -> to-dev. repeat-of: review-findings-stage-b-1 F3 (a gate line attested green that the command does not print), one level out.

Rework itself is good: F1-F6 and F8-F11 all verified by driving production entry points, each with a narrowing mutant I applied and watched kill a named test (10 mutants, 0 survivors). Conformance 19/19 byte-exact with zero skips, gate-selftest 81/81, lint 0 issues, 17 signed commits by the repository human identity, scope confined to stage (b).

BLOCKING B1: the three schema drivers this stage adds (fragment_schema_test.go:312,327; marker_env_schema_test.go:78) t.Fatal instead of recording the registered root-content skip when the conformance root does not publish the family. ci.yml pins SPEC_PIN=0ed5c691, which predates schema-cases/launch-env-fragment-v1 and agent-environment-marker-v1. Reproduced locally against the pinned root; all five hosted Test/Race lanes are red. internal/interop already solves this in five places; copy that pattern and give the three platform-cases rows a root-content tolerated class.

BLOCKING B1b: TestCheckBoundary feeds POSIX-rooted literals to filepath.IsAbs; volumeNameLen is 0 on Windows so the conforming case is rejected as not absolute. Predicted from the stdlib source, then confirmed word for word by the Windows lane. Same class as the stage (a) Windows-only fixture defect. Production is unaffected -- fixture only.

MAJOR M1: F7 shipped TargetsFor/TargetFor with zero production call sites; switch.go:171 still calls the global TargetByID, so profile use --env pi --target xcode-coding-assistant is still admitted. The AC-named shape: the check present but uncalled from production.

MINORS: env status --json emits Go field names; the F2 write-through-link loop can go vacuous without saying so (measured non-vacuous today: 2-4 rendered surfaces per case); the rework report gate table was run against curator-spec main only and asserted the CI lanes without reading them.

Findings: TASK-260906-2g0bgq_review-findings-stage-b-2.md. Probes and mutant logs: TASK-260906-2g0bgq_review-probes-stage-b-2.tar.gz. Empty repository delta is correct by design (implementation lives in the curator repo per the producer brief); stated in the findings.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260906-5ed60e, pid=2169, exit=0)
spawn agent resolution: Agent selection: muse via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=muse; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (muse) (run=RUN-260906-53f212, max_parallel=20)
spawn run started: [implementer] developer (muse) (run=RUN-260906-53f212)
agent completed: [implementer] developer (muse) (exit=0)
spawn run completed: muse (run=RUN-260906-53f212, pid=24425, exit=0)
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=claude; schema=1; producer=v1.6.1-128-gab60e0d; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (claude) (run=RUN-260906-0644ed, max_parallel=20)
spawn run started: [reviewer] reviewer (claude) (run=RUN-260906-0644ed)
Cycle-3 review (RUN-260906-0644ed): ACCEPT at head 1a936e77 via accept_cr rev 3; element routed to integrating. 0 blocking, 0 major, 3 minors. repeat-of: none — the cycle-2 meta-finding (a gate line attested that the command does not print) does not recur. Rework 2 B1/B1b/M1/m2/m3/m4 all driven: 6 of 6 rows, 6 mutants applied, 5 killed, 1 survived and chased to a redundant clause. B1 override tested not accepted: candidate-lane suite-plan fails closed (exit 1) on a root missing either family, a served-but-empty family stays fatal, and the rejected root-content alternative is policy allow in every lane. Every hosted lane green on the exact accepted head (Test/Race x3 os, Lint, Gate self-test x3, Interop, Naming); the Windows lane ran and passed TestCheckBoundary and recorded the three drivers as tolerated root-unset skips. Two minors need orchestrator routing: m5 — internal/interop vector families are still unregistered in root-artifacts.tsv, so a candidate root dropping vectors/environments.json would pass green with all 25 environments cases silently skipped (inherited from stage (a) 4b5cd059, covers five further families, needs its own CI leaf); m6 — the three schema drivers assert nothing on any automatic hosted lane until SPEC_PIN is bumped past f39f4a9. m7 needs no action. Evidence: TASK-260906-2g0bgq_review-findings-stage-b-3.md, _review-verdict-rev3.md, _review-probes-stage-b-3.tar.gz.
agent completed: [reviewer] reviewer (claude) (exit=0)
spawn run completed: claude (run=RUN-260906-0644ed, pid=78134, exit=0)

## Precondition Resources
- [producer-brief-stage-b.md](file://TASK-260906-2g0bgq/producer-brief-stage-b.md) — Producer brief: стадия (b) — managed homes, seeds и passthrough, read-only resolve и фрагмент, MCP-каналы, umbrella dispatch, env status
- [review-brief-stage-b-1.md](file://TASK-260906-2g0bgq/review-brief-stage-b-1.md) — Reviewer brief cycle 1: стадия (b) at 73174cc3
- [producer-brief-stage-b-rework-1.md](file://TASK-260906-2g0bgq/producer-brief-stage-b-rework-1.md) — Rework 1 стадии (b): F1 isolation-gate ниже пина, F2 дрейф байтов линкованной поверхности, F3 красный линт, F4-F6 и минорки
- [review-brief-stage-b-2.md](file://TASK-260906-2g0bgq/review-brief-stage-b-2.md) — Reviewer brief cycle 2: rework F1-F10 стадии (b) at 7bca4d4b
- [producer-brief-stage-b-rework-2.md](file://TASK-260906-2g0bgq/producer-brief-stage-b-rework-2.md) — Stage (b) rework 2: CI root registration, Windows fixture portability, target resolution wiring
- [review-brief-stage-b-3.md](file://TASK-260906-2g0bgq/review-brief-stage-b-3.md) — Review brief cycle 3: rework 2 and the hosted lanes on the pushed head

## Outcome Resources
- [TASK-260906-2g0bgq_spawn-log_-implementer--developer--muse-_RUN-260906-24258e.log](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_spawn-log_-implementer--developer--muse-_RUN-260906-24258e.log) — System spawn log captured by task-board
- [TASK-260906-2g0bgq_drafting-report.md](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_drafting-report.md)
- [TASK-260906-2g0bgq_change-request_rev1.patch](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_change-request_rev1.patch) — Change Request CR-TASK-260906-2g0bgq-1 revision 1 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260906-2g0bgq_spawn-log_-reviewer--reviewer--claude-_RUN-260906-ab4d71.log](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_spawn-log_-reviewer--reviewer--claude-_RUN-260906-ab4d71.log) — System spawn log captured by task-board
- [TASK-260906-2g0bgq_review-findings-stage-b-1.md](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_review-findings-stage-b-1.md) — Reviewer cycle 1 findings for stage (b) at 73174cc3: 3 blocking, 3 major, 5 minor, plus independently rerun gate results
- [TASK-260906-2g0bgq_review-probes-stage-b-1.tar.gz](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_review-probes-stage-b-1.tar.gz) — Go probe tests driving the production entry points behind findings F1, F2, F6 and the composed-fragment check
- [TASK-260906-2g0bgq_spawn-log_-implementer--developer--muse-_RUN-260906-76e22d.log](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_spawn-log_-implementer--developer--muse-_RUN-260906-76e22d.log) — System spawn log captured by task-board
- [TASK-260906-2g0bgq_rework-report-1.md](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_rework-report-1.md) — Rework 1 report: F1-F6 plus F7-F10 fixes, narrowing mutants, gate outputs
- [TASK-260906-2g0bgq_change-request_rev2.patch](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_change-request_rev2.patch) — Change Request CR-TASK-260906-2g0bgq-2 revision 2 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260906-2g0bgq_spawn-log_-reviewer--reviewer--claude-_RUN-260906-5ed60e.log](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_spawn-log_-reviewer--reviewer--claude-_RUN-260906-5ed60e.log) — System spawn log captured by task-board
- [TASK-260906-2g0bgq_review-findings-stage-b-2.md](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_review-findings-stage-b-2.md) — Reviewer cycle 2 findings at 7bca4d4b: F1-F11 rework verified with narrowing mutants; 2 blocking (hosted lanes red on the pinned conformance root and on Windows), 1 major, 3 minors
- [TASK-260906-2g0bgq_review-probes-stage-b-2.tar.gz](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_review-probes-stage-b-2.tar.gz) — Cycle 2 reviewer probes, mutant harness and all ten mutant logs, local darwin test-gate evidence, and the hosted ubuntu/windows lane evidence
- [TASK-260906-2g0bgq_spawn-log_-implementer--developer--muse-_RUN-260906-53f212.log](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_spawn-log_-implementer--developer--muse-_RUN-260906-53f212.log) — System spawn log captured by task-board
- [TASK-260906-2g0bgq_rework-report-2.md](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_rework-report-2.md) — Rework 2 report: B1, B1b, M1, m2, m3 fixes, M1 narrowing mutant, two-root gate table, B1b sweep, honest bounds
- [TASK-260906-2g0bgq_change-request_rev3.patch](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_change-request_rev3.patch) — Change Request CR-TASK-260906-2g0bgq-3 revision 3 candidate patch (repository_delta=empty, 0 changed paths)
- [TASK-260906-2g0bgq_spawn-log_-reviewer--reviewer--claude-_RUN-260906-0644ed.log](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_spawn-log_-reviewer--reviewer--claude-_RUN-260906-0644ed.log) — System spawn log captured by task-board
- [TASK-260906-2g0bgq_review-findings-stage-b-3.md](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_review-findings-stage-b-3.md) — Review findings stage (b) cycle 3 at 1a936e77: ACCEPT, 0 blocking, 0 major, 3 minors (m5-m7); B1/B1b/M1/m2/m3/m4 driven with 6 mutants
- [TASK-260906-2g0bgq_review-probes-stage-b-3.tar.gz](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_review-probes-stage-b-3.tar.gz) — Cycle-3 reviewer probes: two run()-driven probe tests, mutant table and logs, two-root test-gate log, four mutilated-root suite-plan logs, both lanes' skips-observed, Windows CI evidence
- [TASK-260906-2g0bgq_review-verdict-rev3.md](file://TASK-260906-2g0bgq/TASK-260906-2g0bgq_review-verdict-rev3.md) — Cycle-3 review verdict for CR revision 3: ACCEPT; empty repository delta justified; every hosted lane green on head 1a936e77

## Created
2026-09-06T00:55:07Z

## Last Update
2026-09-06T08:39:05Z

## Assigned To
[reviewer] reviewer (claude)
