## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(13))

## Blocked By
- TASK-260728-1g0z69
- TASK-260728-2spy93

## Blocks
- TASK-260728-251p01

## Checklist
- [x] Normative schemas and generated cases encode valid and invalid toolchain requirements for every compiled driver
- [x] Specification fixes fail-fast ordering, stable errors, trusted resolution, fingerprint identity and manager-owned guidance semantics
- [x] Legacy schemas and accepted Go behavior remain compatible and all deterministic gates pass
- [x] Executable vectors prove the two-stage ordering and no-mutation boundary for local and external sources
- [x] Code written per task description and AC
- [x] Relevant tests written for new or changed behavior and passing
- [x] Lint clean
- [x] Relevant build/validation commands run after changes and build not broken
- [x] New outcome artifact attached on the board with a task-scoped name when the work produces notes, logs, screenshots, or other deliverables
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [x] Implementation matches AC
- [x] Solution fits project architecture
- [x] Tests green
- [x] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
EXECUTION DIRECTIVE 2026-07-28: The research/design prerequisite TASK-260728-1g0z69 is now independently accepted at review cycle 8. Land that closed contract normatively in the next curator-spec schema/conformance candidate, together with accepted TASK-260728-2spy93 version/artifact boundaries. Use the accepted task artifacts and candidate worktree as inputs, preserve frozen released rc.5 bytes, keep skill-build.json neutral and manager-independent, cover local vendored plus external-repository builds for Go/Rust/Swift/Kotlin, and include manager-owned toolchain preflight/guidance with no auto-install. Implement schemas, generation, positive/negative vectors, protocol/profile/security text, compatibility guards and release evidence required by the existing AC. Do not stage, commit, publish, advance pins or claim unmeasured platform validation; hand off only with deterministic full-gate evidence.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-88284a, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-88284a)
Execution base: /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2jaw7h/curator-spec-worktree, detached at 57c1f56, materialized as accepted TASK-260728-2kp3tv candidate + TASK-260728-1g0z69 delta (CHANGELOG.md, decisions/0007, docs/compiled-build-toolchain-requirements.md) + TASK-260728-2spy93 delta (decisions/0008, tools/validate.py, tools/test_validate.py). Base fidelity verified by diff -rq against both predecessor worktrees before any edit; predecessors left read-only. Plan: mint the three reserved slots decision 0007 names (agent-skill-v8, csk-skill-v8, skill-build-v2) plus toolchainRequirementV1 and the registry/guidance/diagnostic schemas, land protocol/profile/security text, generate the section 8 vector inventory, add the wire-surface + guidance-coverage release gates and the section 4.2.1.2 boundary probe as a maintained check. Reserved receipt/marker/claim slots stay unallocated; no reserved driver identifier leaves decisions/.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260728-88284a, pid=77299, exit=0)
REVIEW CYCLE 1 DIRECTIVE: Independently review the normative landing against both accepted prerequisites and the full task AC; do not edit candidate artifacts. Treat release preservation as load-bearing: the producer reports release/1.0.0-rc.5.json suite_digest moved even though the accepted research, execution directive and frozen-release policy require rc.5 bytes/pins preserved and rc.6 minting is deferred. Determine the authoritative contract and reject any silent rewrite of a frozen release; a clean internal regeneration is not sufficient if predecessor bytes moved. Verify exact three wire placements and shared ref, schemas 1-7 plus skill-build-v1 compatibility, Stage A/B fail-before-acquisition/cache/mutation/compiler ordering, complete 12-code payload/guidance mapping, manager-only paths/URLs/channels/install/env/trust, resolved identity binding, dry-run/status semantics, and exhaustive positive/negative inventories. Challenge the two design completions (source_ref.surface=registry and classifier matches=absence/value), cross-driver id schema versus primary-driver runtime rejection, prerelease-track unclassifiable semantics, and any reserved Rust metadata admitted before TASK-260728-12pnm1 acceptance. Run both Go 1.25.1 and 1.25.5 boundary probes plus all expected-red controls, full validate/unit/Go/vet/gofmt/compile/diff, deterministic regeneration and clean release gates, and compare every frozen artifact byte-for-byte to the accepted predecessor. Publish TASK-260728-2jaw7h_review-verdict-cycle-1.md with decisive ACCEPTED or CHANGES REQUESTED and route accordingly.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260728-b80704, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260728-b80704)
REVIEW CYCLE 1 VERDICT: CHANGES REQUESTED. Release-blocking evidence: candidate rewrites frozen release/1.0.0-rc.5.json (SHA-256 b32ee9d... -> 32114a32...) and its manifest pins (9ba9b8ec... -> 3675a28c...) although rc.5 bytes/pins must remain frozen and rc.6 minting is deferred to TASK-260728-251p01. A clean internal rc.5 release gate is insufficient because it validates the rewritten set against itself. Secondary defect: docs/compiled-build-toolchain-requirements.md header announces source_ref.surface=registry and matches=absence/value, but the body still states the old three-surface and classifier contract. Full evidence and exact rework are in TASK-260728-2jaw7h_review-verdict-cycle-1.md. Independent tests, both Go 1.25.1/1.25.5 boundary probes, deterministic regeneration, lint, compile, and clean scratch release gate are green. Routed to-dev for ordinary rework; not blocked.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260728-b80704, pid=11526, exit=0)
REWORK CYCLE 2 EXECUTION DIRECTIVE (review cycle 1): Fix only the two accepted-review defects, preserving all otherwise-green protocol/toolchain work. (1) Frozen rc.5 boundary: restore byte-for-byte the accepted predecessor identities from /Users/iv/Developer/ReluxWorks/.temp/TASK-260728-2kp3tv/curator-spec-worktree for release/1.0.0-rc.5.json, conformance/v1/manifest.json, and conformance/v1/schema-cases/index.json, including predecessor hashes b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441, 9ba9b8ec..., and 2faa2baa.... New schemas/vectors/generators must remain on an unreleased candidate surface and must not silently rewrite rc.5; rc.6 minting remains exclusively TASK-260728-251p01. Add an executable compatibility guard against the canonical accepted rc.5 bytes/pins so self-consistent regeneration cannot launder a frozen-release rewrite. (2) Repair docs/compiled-build-toolchain-requirements.md body so source_ref.surface includes registry and classifier matches explicitly distinguishes absence versus value exactly as protocol/core.md, common.schema.json, vectors, and diagnostics do. Add a drift guard/test for both reference rules. Do not broaden scope or weaken existing controls. Re-run validate (48+ schemas/593+ vectors), all Python tests, Go tests/vet/gofmt, compileall outside tree, git diff --check, deterministic regeneration in the correct candidate surface, release gate with explicit frozen-byte comparison, both Go 1.25.1 and 1.25.5 boundary probes, and all five expected-red controls. Preserve dirty user state; no commit, staging, publish, or release mint. Hand off only with exact byte hashes and green logs.
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [implementer] developer (claude) (run=RUN-260728-8b3fae, max_parallel=20)
spawn run started: [implementer] developer (claude) (run=RUN-260728-8b3fae)
REWORK CYCLE 2 COMPLETE — both accepted review-cycle-1 defects fixed, nothing else broadened. (1) Frozen rc.5 boundary restored byte-for-byte: release/1.0.0-rc.5.json b32ee9d35fc2fce0539d0b1c4b15f9c5239115c22fe54a72401f5da6d6646441, conformance/v1/manifest.json 9ba9b8ecf6f06cafda1425aed3539dee8d12af43dabd36f803f4f420dbb1dacf, conformance/v1/schema-cases/index.json 2faa2baaadca30b3eebe3b9248260efcfac30e92cef4fa209bf37f3f23efd4f0; whole conformance/v1 tree diff -rq exit 0 against the accepted TASK-260728-2kp3tv predecessor. The unreleased corpus (5 vector files, 6 schema-case directories, content unchanged) moved to a new candidate suite root conformance/next whose manifest carries no protocol_version and names TASK-260728-251p01 as pin owner. Executable compatibility guard added: release/frozen.json is authored, never generated, and is enforced by tools/generate-vectors (refuses to leave a rewritten release behind), tools/validate.py (every run) and tools/release_gate.py (before a tag), each also cross-checking the release documents own suite_root and both pins. Proven against the exact cycle-1 defect: in a clean scratch repo a manifest rewrite plus a matching repin — self-consistent, exactly what regeneration produces — fails the release gate with frozen release 1.0.0-rc.5 was rewritten; the generator fails the same rewrite at source with exit 1. (2) docs/compiled-build-toolchain-requirements.md body repaired: source_ref.surface now lists manifest/descriptor/registry/source_metadata matching protocol/core.md and common.schema.json, with the registry vs registry_baseline distinction stated correctly; section 3.1.1 gained the matches absence|value legend and the value-class-scoped precedence rule, and both classifier tables gained a matches column with the shipped class names. Drift guard validate_reference_document ties the surface token list in BOTH documents to the schema enum, the legend to toolchain-registry-v1 valueClass.matches, and every classifier row triple in order to the shipped registry entry. LATENT DEFECT FOUND AND FIXED: two agent-skill-v8/csk-skill-v8 case files were shipping in cycle 1 unindexed — the generator never prunes, so a renamed case left an orphan the manifest walk hashed into a pinned digest while nothing validated it; validate_schemas now requires index and filesystem to agree in both directions. GATES (each standalone, real exit codes): validate.py 0 (48 schemas, 592 vector files: 422 released + 170 candidate), unittest 0 (169 tests, was 141), go test 0, go vet 0, gofmt -l 0, compileall 0, git diff --check 0, generate 0; two deterministic regeneration passes 0; clean scratch repo make validate/regenerate-check x2/release-check VERSION=1.0.0-rc.5 all 0; boundary probe 0 on Go 1.25.1 and 1.25.5 (29 cases, 0 failures) and all five expected-red controls failed as required. EXPECTED RED, reported truthfully: in-worktree git diff --exit-code exits 1 (make regenerate-check exits 2) because the whole candidate is uncommitted against 57c1f56 — determinism established by snapshot comparison and the clean git-init probe instead. Nothing staged, committed, published, pinned or advanced; protocol/core.md, profiles/manager.md, SECURITY.md, every schema and tools/toolchain_gate.py untouched this cycle. Full evidence: TASK-260728-2jaw7h_rework-cycle-2.md and TASK-260728-2jaw7h_cycle2-gate-logs.tar.gz.
agent completed: [implementer] developer (claude) (exit=0)
spawn run completed: claude (run=RUN-260728-8b3fae, pid=20777, exit=0)
REVIEW CYCLE 2 DIRECTIVE: Independently verify the rework without editing candidate artifacts. Confirm byte-for-byte predecessor identity for frozen rc.5 release, conformance/v1 manifest, schema-case index and the full v1 tree; prove conformance/next is genuinely unreleased, carries no protocol mint or pin advancement, and cannot be confused with a released suite. Exercise release/frozen.json, generator, validator and release gate against a self-consistent rewritten release+repin so the guard cannot be laundered by updating both sides. Verify the reference-document drift guard for source_ref.surface=registry and classifier matches=absence/value, and the orphan case-file bidirectional index guard. Recheck id closure {go,kotlin,rust,swift}, prerelease-track unclassifiable behavior, reserved Rust expected_metadata_sources, Stage A/B ordering, both Go probes and all expected-red controls. Run full deterministic gates and publish a decisive cycle-2 verdict; ACCEPTED only if no frozen release, protocol version or implementation pin was advanced.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260728-73d506, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260728-73d506)
REVIEW CYCLE 2 VERDICT: ACCEPTED. Frozen rc.5 release document, manifest, schema-case index and full conformance/v1 tree match the accepted predecessor byte-for-byte; conformance/next is versionless, unreleased and pin-owner-scoped. Independent gates: 48 schemas/592 vectors, 169 Python tests, Go tests/vet/gofmt, compileall and diff-check green; two deterministic regeneration passes and clean rc.5 release gate green; anti-laundering, candidate-boundary, orphan-index and reference-drift regressions green; Go 1.25.1/1.25.5 boundary probe green with 29 cases, 331 closure checks and all five expected-red controls. Legacy check: 41 schema files and 57 prior common definitions unchanged. Evidence: TASK-260728-2jaw7h_review-verdict-cycle-2.md and TASK-260728-2jaw7h_review-cycle-2-logs.tar.gz.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260728-73d506, pid=54244, exit=0)

## Precondition Resources
(none)

## Outcome Resources
- [TASK-260728-2jaw7h_spawn-log_-implementer--developer--claude-_RUN-260728-88284a.log](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_spawn-log_-implementer--developer--claude-_RUN-260728-88284a.log) — System spawn log captured by task-board
- [TASK-260728-2jaw7h_results.md](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_results.md) — Developer handoff: toolchain requirement contract landed in curator-spec (schemas, vectors, normative text, gates, probe) with full gate evidence
- [TASK-260728-2jaw7h_protocol-core.md](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_protocol-core.md) — Landed protocol/core.md including the new section 4.2.3 toolchain requirement and two-stage preflight contract
- [TASK-260728-2jaw7h_boundary-probe-green.log](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_boundary-probe-green.log) — Boundary probe green run against Go 1.25.5: 16 go-directive and 13 toolchain-directive cases plus 331 closure checks, 0 failures
- [TASK-260728-2jaw7h_gate-logs.tar.gz](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_gate-logs.tar.gz) — All gate, determinism, clean-probe and boundary-probe logs with their real exit codes
- [TASK-260728-2jaw7h_toolchain-preflight-vectors.json](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_toolchain-preflight-vectors.json) — Generated preflight vector corpus: inventory cases 1-64 and 91-113, two-stage ordering and no-mutation assertions
- [TASK-260728-2jaw7h_spawn-log_-reviewer--reviewer--codex-_RUN-260728-b80704.log](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_spawn-log_-reviewer--reviewer--codex-_RUN-260728-b80704.log) — System spawn log captured by task-board
- [TASK-260728-2jaw7h_review-verdict-cycle-1.md](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_review-verdict-cycle-1.md) — Reviewer cycle 1 verdict: changes requested for frozen rc.5 rewrite and stale reference contract
- [TASK-260728-2jaw7h_spawn-log_-implementer--developer--claude-_RUN-260728-8b3fae.log](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_spawn-log_-implementer--developer--claude-_RUN-260728-8b3fae.log) — System spawn log captured by task-board
- [TASK-260728-2jaw7h_rework-cycle-2.md](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_rework-cycle-2.md) — Rework cycle 2: frozen rc.5 boundary restored via conformance/next plus release/frozen.json guard; reference contract repaired with a drift guard; full gate evidence
- [TASK-260728-2jaw7h_cycle2-gate-logs.tar.gz](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_cycle2-gate-logs.tar.gz) — Cycle 2 gate, determinism, negative-probe, clean-checkout and boundary-probe logs with their real exit codes
- [TASK-260728-2jaw7h_release-frozen.json](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_release-frozen.json) — Authored frozen-release identity record: the accepted rc.5 release document, suite manifest and schema-case index digests
- [TASK-260728-2jaw7h_candidate-suite-manifest.json](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_candidate-suite-manifest.json) — Generated conformance/next manifest: 170 candidate files, no protocol version, pin owner TASK-260728-251p01
- [TASK-260728-2jaw7h_spawn-log_-reviewer--reviewer--codex-_RUN-260728-73d506.log](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_spawn-log_-reviewer--reviewer--codex-_RUN-260728-73d506.log) — System spawn log captured by task-board
- [TASK-260728-2jaw7h_review-verdict-cycle-2.md](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_review-verdict-cycle-2.md) — Reviewer cycle 2 accepted verdict with frozen-release, candidate-boundary, compatibility, deterministic gate, anti-laundering and two-Go-version probe evidence
- [TASK-260728-2jaw7h_review-cycle-2-logs.tar.gz](file://TASK-260728-2jaw7h/TASK-260728-2jaw7h_review-cycle-2-logs.tar.gz) — Independent reviewer cycle 2 validation, unit, Go, lint, determinism, release, targeted regression and boundary-probe logs

## Created
2026-07-28T09:12:39Z

## Last Update
2026-07-28T23:13:58Z

## Assigned To
[reviewer] reviewer (codex)
