# Reviewer verdict for TASK-260810-1uu9lk

Verdict: **changes requested -> analysis**

## Goal and review scope

- Reviewer run: `RUN-260811-e4013e`
- Authoritative run goal immediately before verdict: `GOAL-260811-80f1e4` revision 1
- Resolved scope: `TASK-260810-1uu9lk`
- Required predicate: exactly one evidence-backed reviewer branch for the scoped task
- Directive checkpoint: `request_progress:745288`, acknowledged; this verdict addresses R1-R3, the canonical goldens, scope exclusions, compiled-byte denial, DAG validity, and the uncached repository gate.
- Reviewed outcome: `TASK-260810-1uu9lk_cross-language-closure-graph-and-checkpoints.md`
- Reviewed SHA-256: `0edaf167097ef10216bb282c65be9851992c222b7752d8ecb7d21cb7912131bf`

This artifact records only the `changes_requested` branch. The remaining defects are implementation-blocking architecture/evidence gaps, but they are ordinary autonomous research rework; they require no human-only product, platform, architecture, approval, or external-input decision and do not satisfy Stop-The-Line.

## Prerequisite outcomes consumed

- `TASK-260810-1veyfw`: checked the revision-aware inventory, Go baseline, real Swift-to-C and Go-to-Swift subprocess surfaces, extensionless Mach-O case, independent Python boundary, and deferred ecosystems.
- `TASK-260810-29vk09`: checked shared trust roles, recursive deny-dominant artifact admission, compiled-payload prohibition, toolchain/output causal roles, audit evidence, diagnostics, and reusable byte vectors.
- `TASK-260810-2n3sbi`: checked selection/peer/target semantics, separate Node manager profiles, independent Python protocol boundary, lifecycle/generator controls, C0-C7 checkpoints, and offline vectors.
- `TASK-260810-3urqbl`: checked the lock-superset versus active Cargo graph, pre-vendor admission, pinned vendor transform, host/build-unit rejection, exact toolchain identity, and zero-Cargo-spawn vectors.
- `TASK-260810-zddzh7`: checked intake-before-Swift-manifest ordering, mirror replay, separate Swift/Clang targets, module/header/system boundaries, platform/toolchain identities, and active extension/binary rejection.

## What passed

1. Prior R1 is substantively closed. C0 now binds every pre-C5 evidence toolchain; per-tree intake admission precedes Swift manifest evaluation; Cargo uses C3a -> permitted vendor receipt -> C3b; mirror/metadata execution uses C4 permits; every invocation binds command, environment, process/read/write/network policy and an immediate toolchain recheck. `CGN16-CGN18` cover zero affected process starts and drift.
2. The intrinsic-identity core of prior R2 is closed. Node IDs are relation-free, typed edges are the single producer/consumer/interop authority, expected outputs remain immutable, C6 observations are separate, and duplicate/dangling/recursive-reference validation is specified.
3. The artifact is valid UTF-8 with no trailing whitespace, balanced fences, all required headings and all five prerequisite IDs; it is 959 lines, 84,562 bytes, and 19 Markdown tables. The board outcome and `.research` copy are byte-identical at the SHA-256 above.
4. All 11 records that actually publish both a domain label and exact CCJ payload independently recompute to their stated SHA-256 values.
5. The live board has exactly 14 active implementation leaves plus four explicitly closed superseded placeholders. Every active leaf has code class, required review, a Fibonacci estimate, concrete Spec trace/scope/AC, exactly three checklist gates, and blockers. `plan(TASK-260810-1uu9lk, mode=related, active=true)` is acyclic with the claimed eight waves and critical path.
6. Kotlin, Dart, .NET, new Python implementation, verified binaries, non-SwiftPM native graphs, Rust hooks/proc macros/native/cross support, Node native addons, and active Swift plugins/macros remain excluded or fail closed. The shared compiled-byte policy is not weakened.
7. `task-board validate` passed. `go test -count=1 ./...` passed across every package; notable timings: `cmd/curator` 343.989s, `internal/godriver` 69.303s, `internal/install` 102.249s, and `internal/install/atomicity` 105.925s.

## F1 \u2014 Concrete target-platform identity still contradicts selection-neutral capture

The design says `CaptureGraph` owns `nodes[]` and `edges[]`, while `SelectionContext` carries `target_platform_id` (lines 84-115). It defines `target_platform` as a closed graph node for the exact OS/architecture/ABI/triple/SDK/language/tuning destination (line 163), makes `targets` a typed edge to that node (line 220), and hashes every capture node and edge into `captured_graph_id` while excluding requested target values (lines 404-410).

Those rules cannot all hold without a missing representation rule. If the exact requested platform node and `targets` edges are in `CaptureGraph`, changing the target changes the capture hash. If they are added only after selection, `ActiveGraph` has no schema for new nodes/edges\u2014it only activates capture records\u2014and the edge table is no longer the single authority. The published `CGP05` bytes expose the gap: its capture record contains only two package node IDs (lines 454-456), but both selection records reference placeholder `target_platform_id=sha256:4444...` (lines 458-464) with no published platform node, no `targets` edge, and no rule saying that the reference is external to graph validation. The golden varies only a feature, so it does not prove the document's stronger target-neutral claim.

Required rework:

1. Define one implementable authority for selection-neutral platform declarations and selection-specific concrete target bindings. Either move concrete platform records/target bindings into a canonical selection/active overlay that is allowed to add typed records, or define and validate a finite selection-neutral candidate superset and explain how exact SDK/triple identities enter without request contamination.
2. State whether every `SelectionContext.target_platform_id` must resolve in capture, in a separate immutable platform registry, or in another closed table; add duplicate/dangling/wrong-kind validation for that authority.
3. Publish exact two-target canonical bytes and digest goldens showing unchanged capture bytes/ID, distinct selection/active/plan/checkpoint IDs, and explicit non-hidden `targets` bindings.
4. Refine `TASK-260811-i3154q`, affected adapter tasks, and `TASK-260811-x611eq` so the executable contract tests that exact rule rather than only a feature toggle.

## F2 \u2014 CGP10's claimed exact checkpoint/receipt goldens are not reproducible

The outcome publishes exact labeled CCJ payloads only for the action node, output node, and `produces` edge (lines 483-498). It then lists `C4.close`, `C5.plan`, and `closure_id` hashes without their canonical payload bytes or domain labels (lines 500-509). For each `one`/`two` branch it publishes an observation JSON object and the observation/execution/publication hashes, but omits the observation label and the complete execution/publication receipt payloads and labels (lines 511-524).

Consequently the independent review can recompute the 11 labeled records but cannot derive the three enclosing IDs or six branch receipt IDs from the task-scoped outcome. The producer's `canonical_goldens=pass ... observation_branches=2` sentence is indirect evidence; no new task-scoped verifier/payload resource supplies the missing bytes. That is insufficient for a reusable exact conformance vector and leaves implementers free to invent incompatible receipt encodings.

Required rework:

1. Publish the exact CCJ bytes and domain labels for the CGP10 C4 checkpoint, C5 checkpoint/build plan, closure identity, produced observation, execution receipt, and publication receipt for both branches, including every referenced fixture record.
2. Make every stated hash independently derivable from those bytes and the single documented `ID(label,payload)` rule; attach or inline the validator evidence as a task-scoped resource.
3. Preserve the intended invariant: both branches must have identical action/output/edge/C4/C5/closure IDs and distinct observation/execution/publication IDs.
4. Refine `TASK-260811-i3154q`, `TASK-260811-27xisf`, and `TASK-260811-x611eq` to consume the published bytes rather than unspecified fixture state.

## Routing

Route `TASK-260810-1uu9lk` to `analysis`, revise the same task-scoped architecture outcome, rerun the labeled and receipt golden verifier, live-board/DAG checks, `task-board validate`, and `go test -count=1 ./...`, then send the revision through a new independent reviewer cycle. Do not add new delivery scope and do not use `blocked`.
