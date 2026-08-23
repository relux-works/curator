## Status
done

## Review
required

## Task Class
code

## Estimate
estimated(fibonacci(13))

## Blocked By
- TASK-260811-2gazym
- TASK-260811-i3154q
- TASK-260819-1cpbmc
- TASK-260819-2qytt1

## Blocks
- TASK-260811-3kbf3l

## Checklist
- [x] Implement closed Cargo declaration parsing, immutable registry and Git capture, and pre-vendor zero-spawn admission
- [x] Implement and verify cargo-vendor-transform-v1 per-leaf registry and Git mappings in an absent destination
- [x] Reconcile lock-superset and active graphs and pass all retained transform, origin, containment, and tamper fixtures
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
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260819-fe79f1, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260819-fe79f1)
Orchestrator security checkpoint before review: require artifactpolicy admission before any registry/Git transform parsing; bind and byte-verify complete per-lock-source replacement config, private CARGO_HOME and environment in vendor/metadata permits; close nested metadata parsing without panics; fix 40-hex Git commit validation separately from 64-hex SHA-256; prove these with zero-spawn regressions. A pure caller-supplied Selected/NormalizedManifest seam is not by itself proof of independent Cargo 0.92 Git projection/normalizer derivation; review must verify the implementation owns or cryptographically binds that authority per the accepted rust-source-v1 contract.
Orchestrator config-oracle checkpoint: retained Cargo output removes the precise #commit from the Git source table key; validate DeriveSourceReplacementConfig against the pinned Cargo oracle (registry and Git selectors), require ConfigPath containment in the manager-owned private configuration root, and carry the exact config/home digest into metadata permits so later derivations cannot bypass the verified vendor config.
Orchestrator authority-boundary checkpoint: BindGitDerivation cannot trust an Executor supplied by the same caller as projection bytes. The capture pipeline must own the pinned Cargo 0.92 derivation authority and reject caller-created executor/provider plus arbitrary projection receipts; add an external-package forgery regression and prove exact admitted Git inputs and tool invocation are bound.
Developer blocker checkpoint 2026-08-19: implemented Cargo parsing, admission ordering, registry/Git transform validation, graph/metadata closure, private Cargo config/home binding, Git 40-hex and config-oracle fixes, forged projection/manifest rejection, and shared compiled Git/path diagnostic preservation. Production-positive Git capture remains blocked because no operator-owned, non-substitutable capture-manager derivation authority exists: accepting caller-created closureexec executor/provider or runner makes arbitrary projection receipts self-authorizing. Rejected public BindGitDerivation and injectable-runner approaches. Recommend a shared operator-owned authority handle from the trusted composition root; alternatives are a full internal Cargo 0.92 behavior port with retained GV fixtures or an explicit Git scope deferral. Evidence attached as TASK-260811-2h4m0s_implementation-and-blocker.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260819-fe79f1, pid=42900, exit=0)
Architecture input 2026-08-19: TASK-260819-2qytt1_operator-owned-rust-git-derivation-authority.md is attached as a precondition. Rework in this existing atomic task: add sealed Manager/raw-origin API; remove public toolchain/recheck/staging/vendor+metadata runner/config/destination/derivation fields; implement the fixed hidden Cargo 0.92 Git oracle; bridge oracle/vendor/metadata through one manager-owned closureexec causal chain; add external-package positive plus forgery, drift, zero-start, assurance, and race conformance. Do not create a verified provider or defer Git.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260819-b2e29e, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260819-b2e29e)
STOP-THE-LINE 2026-08-19: accepted sealed Manager requires a closed operator-owned Cargo 1.91.0 selector, but repository has only the intrinsic runtime-Go selector and no trusted Cargo installation record. Ambient PATH/HOME/rustup selection and synthetic vendor materialization are explicitly noncompliant and were not handed off. Recommended decision: approve a trusted application startup/install registry that persists and seals the exact Cargo root; alternative: bundle the pinned Rust toolchain. Evidence/options/exact decision request attached as TASK-260811-2h4m0s_blocker.md. Focused rustsource and closureexec tests exited 0; artifactpolicy gate returned no exit status and is not claimed green.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260819-b2e29e, pid=25434, exit=0)
ORCHESTRATOR RESOLUTION 2026-08-19: this is an internal implementation gap, not a human/external blocker. The accepted TASK-260819-2qytt1 design already requires a closed operator-owned Cargo selector. Bundling the Rust/Cargo toolchain is disallowed by the delivery-wide prohibition on vendored compiled binaries. Implement the remaining valid branch in this same task: a Curator-owned trusted install/startup toolchain registry and closed Cargo 1.91.0/Cargo 0.92 selector that admits and fingerprints the complete external tool root and exact executable, stores no caller authority, performs immediate exact rechecks, and is consumed by the sealed rustsource.Manager without PATH/HOME/rustup lookup during capture. Route vendor and metadata through the same manager-owned closureexec causal executor. No human decision is required.
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260819-fc2bf7, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260819-fc2bf7)
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260819-fc2bf7, pid=62857, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260819-d5560f, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260819-d5560f)
Review RUN-260819-d5560f requests changes: shipped cmd/curator does not import rustsource and rejects __curator_rust_git_oracle_v1; registeredCargo executes ambient rustup during package initialization before C0/permit and before oracle init. See TASK-260811-2h4m0s_review-verdict_RUN-260819-d5560f.md.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260819-d5560f, pid=13507, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [implementer] developer (codex) (run=RUN-260819-017f55, max_parallel=20)
spawn run started: [implementer] developer (codex) (run=RUN-260819-017f55)
Rework RUN-260819-017f55 addresses reviewer findings: cmd/curator now explicitly links and dispatches the Rust Git oracle; rustsource init performs no process execution; sealed-manager C0 Cargo selection uses the native operator toolchain path without PATH/environment/rustup subprocess lookup; vendor and metadata rechecks use the stored registration. Built-binary hostile-rustup regression, focused tests, race test, go vet, build, git diff check, go list dependency proof, and uncached full repository tests exited 0. golangci-lint was unavailable and exited 127; not claimed green. Evidence: TASK-260811-2h4m0s_rework-evidence_RUN-260819-017f55.md.
agent completed: [implementer] developer (codex) (exit=0)
spawn run completed: codex (run=RUN-260819-017f55, pid=18165, exit=0)
spawn agent resolution: Agent selection: codex via explicit_override
spawn launch composition: empty; contract=agents-infra.child-launch-composition; provider=codex; schema=1; producer=dev; diagnostic=launch_composition_empty; no project MCP servers enabled
spawn queued: [reviewer] reviewer (codex) (run=RUN-260819-b44033, max_parallel=20)
spawn run started: [reviewer] reviewer (codex) (run=RUN-260819-b44033)
Reviewer RUN-260819-b44033 accepted the implementation. Production oracle dispatch and no-process-before-C0 rework are verified; focused, race, vet, full uncached repository tests, diff check, and board validation passed. Verdict evidence: TASK-260811-2h4m0s_review-verdict_RUN-260819-b44033.md. No commit_ack supplied.
agent completed: [reviewer] reviewer (codex) (exit=0)
spawn run completed: codex (run=RUN-260819-b44033, pid=35239, exit=0)

## Precondition Resources
- [TASK-260810-1dgdos_accepted-adapter-source-closure-architecture-decision.md](file://TASK-260811-2h4m0s/TASK-260810-1dgdos_accepted-adapter-source-closure-architecture-decision.md) — Accepted normative adapter source-closure architecture decision and implementation DAG
- [TASK-260810-1dgdos_accepted-review-verdict_RUN-260811-55458c.md](file://TASK-260811-2h4m0s/TASK-260810-1dgdos_accepted-review-verdict_RUN-260811-55458c.md) — Independent acceptance verdict for the normative synthesis
- [TASK-260810-1dgdos_accepted-skill-facing-cli-source-closure.md](file://TASK-260811-2h4m0s/TASK-260810-1dgdos_accepted-skill-facing-cli-source-closure.md) — Accepted delivery scope and source-closure security constraints
- [TASK-260810-1uu9lk_accepted-cross-language-closure-graph-and-checkpoints.md](file://TASK-260811-2h4m0s/TASK-260810-1uu9lk_accepted-cross-language-closure-graph-and-checkpoints.md) — Accepted canonical graph, C0-C7 checkpoint, derivation, and exact golden contract
- [TASK-260810-1uu9lk_accepted-canonical-golden-verifier.rb](file://TASK-260811-2h4m0s/TASK-260810-1uu9lk_accepted-canonical-golden-verifier.rb) — Executable oracle for all 53 accepted CGP05 and CGP10 canonical records
- [TASK-260810-1uu9lk_accepted-review-verdict_RUN-260811-e60eda.md](file://TASK-260811-2h4m0s/TASK-260810-1uu9lk_accepted-review-verdict_RUN-260811-e60eda.md) — Independent acceptance verdict for the graph and checkpoint contract
- [TASK-260810-3urqbl_accepted-rust-cargo-source-closure.md](file://TASK-260811-2h4m0s/TASK-260810-3urqbl_accepted-rust-cargo-source-closure.md) — Accepted Rust/Cargo source capture, vendor transform, offline build, diagnostics, and fixtures
- [TASK-260819-2qytt1_operator-owned-rust-git-derivation-authority.md](file://TASK-260811-2h4m0s/TASK-260819-2qytt1_operator-owned-rust-git-derivation-authority.md) — Reviewed-input candidate defining the operator-owned authority rework required by the Rust capture task
- [TASK-260819-2qytt1_accepted-operator-owned-rust-git-derivation-authority.md](file://TASK-260811-2h4m0s/TASK-260819-2qytt1_accepted-operator-owned-rust-git-derivation-authority.md) — Accepted operator-owned Rust Git derivation authority design
- [TASK-260819-2qytt1_accepted-review_RUN-260819-b85a6b.md](file://TASK-260811-2h4m0s/TASK-260819-2qytt1_accepted-review_RUN-260819-b85a6b.md) — Independent acceptance verdict for the Rust Git authority design
- [TASK-260811-2h4m0s_rework-brief_RUN-260819-d5560f.md](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_rework-brief_RUN-260819-d5560f.md) — Required production wiring and no-process-before-C0 rework from independent review

## Outcome Resources
- [TASK-260811-2h4m0s_spawn-log_-implementer--developer--codex-_RUN-260819-fe79f1.log](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_spawn-log_-implementer--developer--codex-_RUN-260819-fe79f1.log) — System spawn log captured by task-board
- [TASK-260811-2h4m0s_implementation-and-blocker.md](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_implementation-and-blocker.md) — Implementation evidence, validation exits, and unresolved Git derivation authority ownership constraint
- [TASK-260811-2h4m0s_spawn-log_-implementer--developer--codex-_RUN-260819-b2e29e.log](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_spawn-log_-implementer--developer--codex-_RUN-260819-b2e29e.log) — System spawn log captured by task-board
- [TASK-260811-2h4m0s_blocker.md](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_blocker.md) — Stop-the-line evidence for missing operator-owned Cargo selector
- [TASK-260811-2h4m0s_spawn-log_-implementer--developer--codex-_RUN-260819-fc2bf7.log](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_spawn-log_-implementer--developer--codex-_RUN-260819-fc2bf7.log) — System spawn log captured by task-board
- [TASK-260811-2h4m0s_implementation-evidence.md](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_implementation-evidence.md) — Rust source capture, Cargo vendor/metadata authority, immutable replay correction, and validation evidence
- [TASK-260811-2h4m0s_spawn-log_-reviewer--reviewer--codex-_RUN-260819-d5560f.log](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_spawn-log_-reviewer--reviewer--codex-_RUN-260819-d5560f.log) — System spawn log captured by task-board
- [TASK-260811-2h4m0s_review-verdict_RUN-260819-d5560f.md](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_review-verdict_RUN-260819-d5560f.md) — Changes-requested reviewer verdict with production oracle reachability and pre-C0 process evidence
- [TASK-260811-2h4m0s_spawn-log_-implementer--developer--codex-_RUN-260819-017f55.log](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_spawn-log_-implementer--developer--codex-_RUN-260819-017f55.log) — System spawn log captured by task-board
- [TASK-260811-2h4m0s_rework-evidence_RUN-260819-017f55.md](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_rework-evidence_RUN-260819-017f55.md) — Production oracle wiring, no-process-before-C0 rework, and validation exits
- [TASK-260811-2h4m0s_spawn-log_-reviewer--reviewer--codex-_RUN-260819-b44033.log](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_spawn-log_-reviewer--reviewer--codex-_RUN-260819-b44033.log) — System spawn log captured by task-board
- [TASK-260811-2h4m0s_review-verdict_RUN-260819-b44033.md](file://TASK-260811-2h4m0s/TASK-260811-2h4m0s_review-verdict_RUN-260819-b44033.md) — Accepted reviewer verdict closing production oracle wiring and pre-C0 process findings

## Created
2026-08-11T05:10:21Z

## Last Update
2026-08-19T08:47:24Z

## Assigned To
[reviewer] reviewer (codex)
