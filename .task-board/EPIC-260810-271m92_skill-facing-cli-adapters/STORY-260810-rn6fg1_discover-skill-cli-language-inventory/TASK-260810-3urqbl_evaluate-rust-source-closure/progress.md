## Status
done

## Review
required

## Task Class
research

## Estimate
estimated(fibonacci(5))

## Blocked By
- TASK-260810-1veyfw
- TASK-260810-29vk09

## Blocks
- TASK-260810-1uu9lk

## Checklist
- [x] Verify Cargo recursive source closure and offline behavior with authoritative evidence
- [x] Analyze git dependencies, build scripts, proc macros, features, and target-specific inputs
- [x] Recommend the Rust profile, checkpoints, diagnostics, and conformance fixtures
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
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [analyst] researcher (codex) (run=RUN-260811-2af959, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-260811-2af959)
Researcher logbook 2026-08-11, RUN-260811-2af959: Recommended fail-closed rust-source-v1. Cargo.lock is the conservative all-feature/all-target superset; exact active units require locked offline metadata for the declared target/features. cargo vendor captures registry and Git sources but root/path snapshots and independent immutable origin manifests remain required. Fresh-home frozen recursive rebuild succeeded with the original Git repository unavailable. Active build.rs and proc macros read ambient inputs despite frozen/offline Cargo, so v1 rejects them until an OS-enforced hermetic-hook capability exists. A source tamper failed with exit 101, but forging the vendored per-file .cargo-checksum.json made the build exit 0; Cargo directory checksums are therefore not a security trust root. Common compiled-artifact prohibition applies to all captured dependency payloads before Cargo runs. Outcome: TASK-260810-3urqbl_rust-cargo-source-closure.md, sha256 300f294974ae9d1ea29d3e63253517015889b04f4024c91be77c9175fc945202. Document gates: non-empty, trailing-whitespace, fence-balance, scope coverage, local-link, and SHA-256 commands all exit 0.
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-260811-2af959, pid=0, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260811-87b027, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260811-87b027)
Reviewer verdict RUN-260811-87b027 under GOAL-260811-33d6ad rev 1: changes requested. R1: the normative procedure runs cargo vendor before raw archive/Git admission, violating the accepted pre-extraction boundary. R2: vendor-to-origin leaf equality is undefined and contradicted by the retained itoa fixture, where .gitignore is raw-only and .cargo-checksum.json is vendor-only. Relevant Go baseline tests passed; full go test ./... was inconclusive under concurrent suite contention. See TASK-260810-3urqbl_reviewer-verdict_RUN-260811-87b027.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260811-87b027, pid=0, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [analyst] researcher (codex) (run=RUN-260811-a8f7f2, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-260811-a8f7f2)
Rework evidence 2026-08-11 (RUN-260811-a8f7f2): R1 is resolved by an immutable-acquisition -> pre-vendor shared admission -> private admitted CARGO_HOME/offline cargo vendor -> post-vendor verification boundary; Cargo metadata/build cannot run before both admission stages pass. R2 is resolved by cargo-vendor-transform-v1 pinned to Cargo 0.92.0 commit ea2d978, with explicit registry/Git per-leaf dispositions, normalized-manifest/checksum rules, both manifests bound into the checkpoint, and named PV/RV/GV/VF conformance cases. Probe anomaly: the retained old itoa vendor checksum had a trailing LF, so the initial independent mapping verifier truthfully exited 1; fresh empty-destination regeneration produced canonical checksum sha256 35abe1dc1588d7a386b232f534a6d6bab406859c89809454735e9c84d1c7c81d and the corrected verifier exited 0. Expected-red wrong vendor path exited 101; old-vs-fresh comparison exited 1. Fresh frozen build and binary each exited 0; document gates, resource parity, and go test ./internal/buildmeta ./internal/buildcache ./internal/godriver each exited 0. Updated outcome sha256 c85fa21cabdc53ffcd3dbae4149b537eb575616367fa711e4451f516039e7d39.
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-260811-a8f7f2, pid=0, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260811-e96528, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260811-e96528)
Reviewer verdict RUN-260811-e96528 under GOAL-260811-38ee21 rev 1: changes requested. R1 is resolved and outcome parity is sha256 c85fa21cabdc53ffcd3dbae4149b537eb575616367fa711e4451f516039e7d39. R2 remains incomplete: Cargo 0.92 registry unpack omits every basename .cargo-ok, while the proposed transform permits omission only at package root; Cargo PathSource re-adds committed root target files for Git packages without include, while the proposed transform excludes root target. Retained verifier, sandboxed fresh vendor/build, document gates, and focused Go tests pass. Full cached go test ./... reproduced the known 17 GB test-input scan anomaly and was terminated. Evidence: TASK-260810-3urqbl_reviewer-verdict_RUN-260811-e96528.md. Route to analysis for research rework; this is not blocked.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260811-e96528, pid=0, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [analyst] researcher (codex) (run=RUN-260811-c30696, max_parallel=20)
spawn run started: [analyst] researcher (codex) (run=RUN-260811-c30696)
Progress RUN-260811-c30696: directive request_progress observed under GOAL-260811-001e97 rev 1. Focus remains strictly R2a/R2b. Pinned Cargo 0.92 source confirms registry unpack drops every basename .cargo-ok and PathSource uses distinct no-include Git-index vs include filesystem-walk branches. New retained edge fixture/independent verifier exits 0; physical vendor probes exit 0; three fresh frozen rebuilds exit 0 with Git origins unavailable. Earlier build attempts truthfully exited 101 for ENOSPC, recovered by deleting only regenerable task-owned outputs. Research revision and focused gates are next.
Rework evidence 2026-08-11, RUN-260811-c30696 under GOAL-260811-001e97 rev 1: R2a now models registry .cargo-ok as basename-wide omit_registry_cargo_ok while Git retains its exact-root vendor filter. R2b now pins git_index_no_include tracked root-target re-addition and filesystem_include hard root-target exclusion, separately rejects dirty untracked/ignored intake, binds Cargo readiness-marker and normalized-manifest inputs, and adds GV03d compiled tracked-target pre-vendor zero-spawn coverage. Retained edge archive e1d506e60e9cd73c82d5aa97455fdb24982bd18aa327db6a6ecc8e6203dc1c6f generated checksum ae5e670fffa04e8cee82764ac110f92f474924b62dd95c02b19df2d6c406ac80; Git checksum outputs 3937bbcb49e3a82c7b844818926953ec23e44b4c8514d09de40c23bcff762f75 and d472136289d4f41b5fbc0f65ede9e5425feeaea82d8179e57945bfc27d38dda6. Document gate, both mapping verifiers, resource parity, and focused Go tests exit 0. Initial three edge rebuilds truthfully exited 101 for ENOSPC; after removing only regenerable task-owned build outputs, all three fresh frozen rebuilds and binaries exited 0 with Git origins unavailable. Revised outcome is 90084 bytes, sha256 620c789545273a1c4fc9c9baf25b9db8e4220c79f3fb1ad299ac1bdcd7e51423.
Directive nudge 60508a observed and satisfied: RV02/RV03 now assert root plus nested registry .cargo-ok dispositions, copied nested .gitignore, exact vendor leaf set and generated checksum bytes; GV03a-GV03e assert no-include tracked target, untracked and ignored dirty-intake rejection, explicit-include root-target omission, compiled tracked-target pre-vendor zero Cargo spawns, and monorepo/submodule coverage. Closure checkpoint and empirical tables bind the projection modes, marker/index inputs, hashes, commands, and exit codes. Focused document and mapping gates remain exit 0.
agent completed: [analyst] researcher (codex) (exit=0)
spawn run completed: codex (run=RUN-260811-c30696, pid=0, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260811-05742e, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260811-05742e)
Reviewer verdict RUN-260811-05742e under GOAL-260811-d4fba8 revision 1: accepted. Reviewed outcome/resource parity is SHA-256 620c789545273a1c4fc9c9baf25b9db8e4220c79f3fb1ad299ac1bdcd7e51423. R1/R2/R2a/R2b are resolved: pre-vendor admission is zero-spawn fail-closed; cargo-vendor-transform-v1 is pinned; registry .cargo-ok omission is basename-wide and Git-specific rules remain exact-root; both Git root-target PathSource modes and compiled tracked-target denial are covered. Independent edge/original mapping and document gates exited 0; three fresh-home frozen builds passed under OS network and Git-origin read denial and printed tracked-target/include-walk/17; focused buildmeta/buildcache/godriver tests passed. Evidence: TASK-260810-3urqbl_reviewer-verdict_RUN-260811-05742e.md.
Final directive checkpoint RUN-260811-05742e: request_progress:cce41a was observed under GOAL-260811-d4fba8 revision 1 and is satisfied. The accepted review remained focused on R2a/R2b, RV02/RV03, GV03a-GV03e, outcome parity, and focused gates; the updated verdict artifact records this directive evidence.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260811-05742e, pid=0, exit=0)

## Precondition Resources
- [TASK-260810-3urqbl_skill-facing-cli-source-closure.md](file://TASK-260810-3urqbl/TASK-260810-3urqbl_skill-facing-cli-source-closure.md) — Current delivery scope and source-closure security constraints
- [TASK-260810-29vk09_accepted-compiled-artifact-taxonomy.md](file://TASK-260810-3urqbl/TASK-260810-29vk09_accepted-compiled-artifact-taxonomy.md) — Accepted shared artifact taxonomy, recursive detection, diagnostics, audit evidence, verified-binary seam, and conformance vectors
- [TASK-260810-1veyfw_accepted-inventory-language-and-reference-surfaces.md](file://TASK-260810-3urqbl/TASK-260810-1veyfw_accepted-inventory-language-and-reference-surfaces.md) — Accepted revision-pinned skill-facing CLI and protocol inventory, including deterministic surface authority and mixed-language evidence
- [TASK-260810-3urqbl_rework-context_RUN-260811-87b027.md](file://TASK-260810-3urqbl/TASK-260810-3urqbl_rework-context_RUN-260811-87b027.md) — Reviewer changes-requested verdict preserved as immutable rework context for the next Rust research producer.
- [TASK-260810-3urqbl_rework-context_RUN-260811-e96528.md](file://TASK-260810-3urqbl/TASK-260810-3urqbl_rework-context_RUN-260811-e96528.md) — Second reviewer changes-requested verdict preserved as immutable rework context for Cargo transform edge cases.

## Outcome Resources
- [TASK-260810-3urqbl_spawn-log_-analyst--researcher--codex-_RUN-260811-2af959.log](file://TASK-260810-3urqbl/TASK-260810-3urqbl_spawn-log_-analyst--researcher--codex-_RUN-260811-2af959.log) — System spawn log captured by task-board
- [TASK-260810-3urqbl_rust-cargo-source-closure.md](file://TASK-260810-3urqbl/TASK-260810-3urqbl_rust-cargo-source-closure.md) — Rust/Cargo recursive source closure revised for registry .cargo-ok and Git root-target PathSource edge cases
- [TASK-260810-3urqbl_spawn-log_-reviewer--reviewer--codex-_RUN-260811-87b027.log](file://TASK-260810-3urqbl/TASK-260810-3urqbl_spawn-log_-reviewer--reviewer--codex-_RUN-260811-87b027.log) — System spawn log captured by task-board
- [TASK-260810-3urqbl_reviewer-verdict_RUN-260811-87b027.md](file://TASK-260810-3urqbl/TASK-260810-3urqbl_reviewer-verdict_RUN-260811-87b027.md) — Reviewer changes-requested verdict: pre-vendor admission ordering and deterministic origin-to-vendor mapping
- [TASK-260810-3urqbl_spawn-log_-analyst--researcher--codex-_RUN-260811-a8f7f2.log](file://TASK-260810-3urqbl/TASK-260810-3urqbl_spawn-log_-analyst--researcher--codex-_RUN-260811-a8f7f2.log) — System spawn log captured by task-board
- [TASK-260810-3urqbl_spawn-log_-reviewer--reviewer--codex-_RUN-260811-e96528.log](file://TASK-260810-3urqbl/TASK-260810-3urqbl_spawn-log_-reviewer--reviewer--codex-_RUN-260811-e96528.log) — System spawn log captured by task-board
- [TASK-260810-3urqbl_reviewer-verdict_RUN-260811-e96528.md](file://TASK-260810-3urqbl/TASK-260810-3urqbl_reviewer-verdict_RUN-260811-e96528.md) — Reviewer changes-requested verdict: remaining exact Cargo vendor transform gaps for nested registry .cargo-ok and tracked Git target files
- [TASK-260810-3urqbl_spawn-log_-analyst--researcher--codex-_RUN-260811-c30696.log](file://TASK-260810-3urqbl/TASK-260810-3urqbl_spawn-log_-analyst--researcher--codex-_RUN-260811-c30696.log) — System spawn log captured by task-board
- [TASK-260810-3urqbl_spawn-log_-reviewer--reviewer--codex-_RUN-260811-05742e.log](file://TASK-260810-3urqbl/TASK-260810-3urqbl_spawn-log_-reviewer--reviewer--codex-_RUN-260811-05742e.log) — System spawn log captured by task-board
- [TASK-260810-3urqbl_reviewer-verdict_RUN-260811-05742e.md](file://TASK-260810-3urqbl/TASK-260810-3urqbl_reviewer-verdict_RUN-260811-05742e.md) — Accepted reviewer verdict with final directive evidence, pinned Cargo source checks, retained mapping gates, sandboxed offline rebuilds, and scoped architecture tests

## Created
2026-08-10T18:58:20Z

## Last Update
2026-08-11T04:56:59Z

## Assigned To
[reviewer] reviewer (codex)
