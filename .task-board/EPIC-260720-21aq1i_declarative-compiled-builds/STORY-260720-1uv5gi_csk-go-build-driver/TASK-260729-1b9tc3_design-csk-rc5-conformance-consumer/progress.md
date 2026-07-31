## Status
done

## Review
required

## Task Class
code

## Estimate
notEstimated

## Blocked By
- (none)

## Blocks
- (none)

## Checklist
- [x] Immutable rc.5 root and all relevant vector clusters are inventoried with digests
- [x] Python fixture loader and independent consumer boundaries are designed without copied Go logic
- [x] Positive, rejection, identity, lifecycle, platform, and non-alias cases map to exact tests
- [x] Literal first-wave pytest implementation gates are attached in a task-scoped outcome
- [x] Board size is proportional to the spec and is the smallest decomposition that maps every requirement
- [x] Every story and task traces to a concrete spec requirement; justified-gap elements also carry a self-verified gap record
- [x] Beyond-literal-spec elements include a written justification naming the gap and the spec and out-of-scope checks performed before creation
- [x] Research tasks cite an exact question the spec genuinely leaves open
- [x] Dependencies linked
- [x] Tasks are atomic — one clear deliverable each
- [x] Completeness verified — nothing forgotten
- [x] Any planning artifacts actually produced are linked as new task-scoped outcome resources; diagrams are strictly optional, never a standing deliverable
- [x] Important findings, decisions, anomalies, or regressions recorded in logbook when relevant
- [ ] Implementation matches AC
- [ ] Solution fits project architecture
- [ ] Tests green
- [ ] If review does not accept the work — verdict evidence added and status routed by the explicit verdict branches

## Notes
spawn agent resolution: Agent selection: claude via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=claude; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [analyst] solution-architect (claude) (run=RUN-260729-608f9c, max_parallel=20)
spawn run started: [analyst] solution-architect (claude) (run=RUN-260729-608f9c)
Analysis evidence 2026-07-29. Immutable rc.5 root /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1: manifest.json sha256 b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c, protocol_version 1.0.0-rc.5, 447 manifest files + manifest.json = 448 on disk, aggregate sorted-file-digest e6a132157806bf747f0ad24a61bb5a9a4c8b915dac743d9465c889c0a1ad2fae. Matches the accepted TASK-260729-3nx97g candidate pin exactly, so the parity-map section 3.3 build-driver golden gap is closed: vectors/build-drivers.json (f412c107) and the 11-file expected/build-driver/ tree are present. Previous default pin measured: .github/workflows/ci.yml checks out relux-works/curator-spec at cbe912d064e06275b0a1aa6762b7c31f687051c5, which is protocol 1.0.0-rc.2, manifest sha256 728f772950414b9c3ddf38a8f1e9f2c7d2953bdca1d8c135c7e1a9abf40fff06, 81 files, 12 vectors. Baseline runs: tests/test_protocol_conformance.py is 98 passed against the pinned rc.2 root and 1 failed / 97 passed against the rc.5 root. Root cause is exact and not a protocol semantics change: the shared fixture manifest was renamed csk-skill.json -> agent-skill.json with byte-identical semantics, and csk.skillspec.load_skill_spec only reads csk-skill.json, so it silently returns an empty SkillSpec and copy_context includes scripts/golden-tool. Feasibility proved with zero product edits: csk.audit_registry.canonical_bytes reproduces the 869-byte expected/build-driver/build-input.ccj.json and the 1120-byte receipt.ccj.json byte-for-byte, re-derives all three cache_identity keys (portable 529370.., legacy_rc4 3fcd714a.., reserved_hardened 13736230..), aliases is false and the keys are distinct, and expected/build-driver/marker.json equals portable_identity.marker. Framing of build-source and toolchain preimages is fully recoverable in pure Python (tag byte D/F/L/V + uint64 big-endian path and payload lengths), so no Go logic needs copying. Structural constraint found: schemas/v1 lives outside CURATOR_CONFORMANCE_ROOT and is not covered by conformance/v1/manifest.json, so schema-file-dependent assertions need separate resolution and provenance.
Design handoff 2026-07-29. Attached TASK-260729-1b9tc3_rc5-conformance-consumer-design.md (SHA-256 aa2513c5a1bc138024eb59a858db67aa19ef3e350e41a6c7d22d71272e0b55ec), byte-identical to .planning/260729_csk-rc5-conformance-consumer-design.md, plus baseline-pin-rc2.log and candidate-rc5.log. Architecture: a tests/conformance/{root,clusters,goldens,platform}.py support package plus nine test modules; the harness may only read bytes from the root, call a public csk entry point, and compare, so no protocol arithmetic, canonicalization, framing or ordering lives in tests and no Go logic is copied. CURATOR_CONFORMANCE_ROOT keeps its exact current semantics including skip-when-unset, one optional CURATOR_CONFORMANCE_MANIFEST_SHA256 assertion is added, and the committed default pin stays on relux-works/curator-spec cbe912d0 (protocol 1.0.0-rc.2). Key design decision: cluster gating is fail-on-missing, never skip-on-missing, because parity-map section 3.3 proved a silent SKIP hid the absence of the rc.5 build-driver suite for the whole candidate line while every gate reported exit 0. Coverage mapped case by case: 8 positives, 77 rejections across 10 boundaries and 58 distinct error codes, 10 build-source and 12 toolchain cases, 5 argv forms, 28 fixed-environment keys, the 3-key non-alias cache identity with aliases false, all 11 expected/build-driver byte goldens, 102 in-scope schema cases across agent-skill-v6, csk-skill-v6, build-receipt-v1, install-marker-v2 and conformance-claim v1/v2/v3, the full go-host-execution-policy surface (18 mandatory controls, 5x2 native-control inventory, 14 identity/protocol, 8 package-influence, 11 capability-evidence, 8 consistency rules, 6 paired deferred guarantees, 3 failure boundaries, 13 session states, 4 process-graph nodes, 6 policy semantics), manager-lifecycle bootstrap/dry-run/launcher/upgrade, and the claim-v3 qualification rules with no claim emitted. Platform policy is three tiers: tier A pure-data runs everywhere including Linux, tier B host behavior on macOS/Windows, tier C real Go build on macOS/Windows with a resolvable trusted GOROOT; Linux is asserted fail-closed with build_execution_control_unavailable rather than skipped, which matches the AC TASK-260729-v5hqnv gave TASK-260720-3pemm6 during this run. Wave 1 needs no product change and was executed here: canonical_bytes reproduces both goldens byte-for-byte, all three cache keys re-derive and stay distinct, marker.json equals portable_identity.marker, and hashing.content_sha256(fixtures/skill) reproduces expected/snapshot_sha256.txt. Blocking prerequisite found and evidenced: the shared fixture manifest rename csk-skill.json -> agent-skill.json makes load_skill_spec return an empty spec, so the suite is 1 failed / 97 passed against rc.5 versus 98 passed against the pin, and fixture-driven build-root assertions would be vacuous rather than red. TASK-260720-z9j4c9 owns the fix in its description but not in its AC, and it is not one of the seven briefs v5hqnv retargeted; the exact one-clause AC amendment is in section 10.2 and is deliberately not applied because this task is read-only design. No new board element is proposed; section 10.3 records the six gap candidates checked and why each was rejected. Only new files are the artifact, the two logs, and a read-only export of the pinned rc.2 suite under .temp/TASK-260729-1b9tc3/; no CocoaSkills or Curator source, pin, release or publication was touched.
agent completed: [analyst] solution-architect (claude) (exit=0)
spawn run completed: claude (run=RUN-260729-608f9c, pid=21340, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: degraded_contract_unavailable; contract=agents-infra.child-launch-composition; provider=codex; schema=1; diagnostic=composition_contract_unavailable; bare child launch retained
spawn queued: [reviewer] reviewer (codex) (run=RUN-260729-4aea12, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260729-4aea12)
Reviewer verdict ACCEPTED 2026-07-29. Independent evidence is attached as TASK-260729-1b9tc3_review-verdict.md. Root/manifest digests, all cluster counts, CCJ/golden identities, three-key non-alias behavior, current CocoaSkills boundaries, fixture rename drift, platform policy, and literal implementation gates were verified read-only. No code, design, pin, publication, or claim was changed; no broad tests were run.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260729-4aea12, pid=38952, exit=0)

## Precondition Resources
- [TASK-260729-1b9tc3_consumer-scope.md](file://TASK-260729-1b9tc3/TASK-260729-1b9tc3_consumer-scope.md) — Immutable rc.5 root and CocoaSkills consumer design scope
- [TASK-260729-1b9tc3_review-instructions.md](file://TASK-260729-1b9tc3/TASK-260729-1b9tc3_review-instructions.md) — Independent review scope for CocoaSkills rc.5 conformance consumer design

## Outcome Resources
- [TASK-260729-1b9tc3_spawn-log_-analyst--solution-architect--claude-_RUN-260729-608f9c.log](file://TASK-260729-1b9tc3/TASK-260729-1b9tc3_spawn-log_-analyst--solution-architect--claude-_RUN-260729-608f9c.log) — System spawn log captured by task-board
- [TASK-260729-1b9tc3_rc5-conformance-consumer-design.md](file://TASK-260729-1b9tc3/TASK-260729-1b9tc3_rc5-conformance-consumer-design.md) — rc.5 immutable-root digests, cluster-to-pytest-module map, fixture/digest provenance, non-alias and negative coverage, platform gating and skip policy, literal first-wave commands
- [TASK-260729-1b9tc3_baseline-pin-rc2.log](file://TASK-260729-1b9tc3/TASK-260729-1b9tc3_baseline-pin-rc2.log) — Existing csk conformance module against the committed default pin (protocol 1.0.0-rc.2): 98 passed, exit 0
- [TASK-260729-1b9tc3_candidate-rc5.log](file://TASK-260729-1b9tc3/TASK-260729-1b9tc3_candidate-rc5.log) — Same module against the immutable rc.5 candidate root: 1 failed / 97 passed, exit 1; failure is the agent-skill.json manifest-name regression
- [TASK-260729-1b9tc3_spawn-log_-reviewer--reviewer--codex-_RUN-260729-4aea12.log](file://TASK-260729-1b9tc3/TASK-260729-1b9tc3_spawn-log_-reviewer--reviewer--codex-_RUN-260729-4aea12.log) — System spawn log captured by task-board
- [TASK-260729-1b9tc3_review-verdict.md](file://TASK-260729-1b9tc3/TASK-260729-1b9tc3_review-verdict.md) — Independent accepted review of the rc.5 Python conformance consumer design

## Created
2026-07-29T11:45:14Z

## Last Update
2026-07-29T12:11:48Z

## Assigned To
[reviewer] reviewer (codex)
