# Reviewer verdict for TASK-260810-1uu9lk

Verdict: **changes requested -> analysis**

## Goal and scope evidence

- Reviewer run: `RUN-260811-ef80a0`
- Authoritative goal immediately before verdict: `GOAL-260811-5f7d87` revision 1
- Resolved scope: `TASK-260810-1uu9lk`
- Review policy: `required`
- Observed directive: `request_progress:f08712`, requesting concrete findings and convergence on a persisted verdict
- Reviewed outcome: `TASK-260810-1uu9lk_cross-language-closure-graph-and-checkpoints.md`
- Reviewed SHA-256: `dce817c7a08742d7cc4b7d0aa74eddcdab4d1d124953d9b9aaf2f1587fc35ea6`

The goal requires exactly one evidence-backed reviewer branch. This artifact records only the `changes_requested` branch; it records neither acceptance nor a Stop-The-Line boundary.

## What passed

1. All five accepted prerequisite outcomes are cited and materially consumed: the revision-aware inventory, shared artifact taxonomy, Node/TypeScript and independent Python split, final Rust Cargo transform decision, and SwiftPM/C-family decision.
2. The design covers closed node and edge kinds, explicit interop modes, selected and pruned conditions, cycle handling, deterministic topological waves, canonical IDs, C0-C7 checkpoints, protected output receipts, stable diagnostics, and reusable positive and negative vectors.
3. The live implementation board has exactly 14 active code leaves. Every active leaf has required review, a Fibonacci estimate, task-specific scope and AC, three checklist gates, a concrete Spec trace, and at least one dependency. The related plan is acyclic with eight waves and the documented critical path.
4. Kotlin, Dart, .NET, verified-binary admission, non-SwiftPM native graphs, Node native addons, active Rust hooks or proc macros, and active Swift plugins or macros remain explicitly excluded or fail closed. No compiled-dependency policy weakening was found.
5. Artifact structure passed at 720 lines and 17 tables with valid UTF-8, no trailing whitespace, balanced fences, all required headings and prerequisite IDs, and exact SHA-256 above. The board resource and research copy are byte-identical.
6. `task-board validate` passed. `go test -count=1 ./...` passed across every package, including `cmd/curator` in 345.691s and `internal/godriver` in 66.183s. A default cached run entered the previously documented cache-input traversal anomaly and only that reviewer-launched process was terminated; the reliable uncached gate is green.

## R1 - Pre-C5 execution is ordered before admission and authoritative toolchain binding

The common checkpoint table executes SwiftPM manifests in C1, Cargo vendoring in C3, and SwiftPM mirror replay plus Cargo or Node metadata in C4 (outcome lines 377-380). Yet the first-binding table assigns toolchain, runtime, command, environment, sandbox, and build-order authority to C5 (line 394), and the protected-boundary section says only C5 actions may start (lines 412-414).

This is internally contradictory and violates accepted prerequisite ordering:

- The accepted SwiftPM decision freezes and recursively scans the root before manifest evaluation, and scans and freezes every fetched package tree before evaluating its manifest (accepted SwiftPM outcome lines 376-392).
- That decision also requires a toolchain time-of-use recheck before manifest evaluation (lines 348-367).
- The common C0 contains only policy selectors; C2 carries only initial toolchain fingerprints after C1 manifest execution; C5 is therefore too late to establish the exact runtime that already affected resolution.
- The action node includes materialization and other executable manager steps, so treating the pre-C5 invocations as unmodeled exceptions would reintroduce hidden executable edges.

Required rework:

1. Define an explicit pre-build evidence-derivation execution model for manifest evaluation, vendoring, mirror replay, and metadata, with exact executable/toolchain identity, command, environment, read/write/process/network policy, time-of-use recheck, and causal checkpoint binding before each invocation.
2. Preserve the accepted Swift ordering by admitting the complete root tree before its manifest and each complete fetched package tree before its manifest. If C1-C3 must iterate, specify the iteration/subreceipt chain rather than claiming one unsafe linear pass.
3. Clarify that the C5-only rule applies only to the derived build DAG, or move all executable steps into a plan that exists before they run.
4. Add negative vectors proving that a rejected root/dependency payload causes zero Swift manifest evaluations and that toolchain drift before manifest/vendor/metadata use fails before the affected process.

## R2 - Node identity payloads are recursive and output identity is mutable

The node table puts graph relationships inside node identity payloads: target units contain source-set refs, actions contain declared inputs and outputs, generated artifacts contain producer action and consumers, interop boundaries contain provider and consumer target refs, and output artifacts contain producer and consumers (outcome lines 118-124). The identity table then hashes the complete kind-specific payload while claiming that excluding edges prevents recursive hashes (lines 343-348).

That claim does not hold. An action that names an output and an output that names its producer creates a hash recursion even when edge tables are excluded. The same relationships also duplicate the explicit reads, produces, provides_interop, and consumes_interop edge kinds, creating two authorities for one edge. In addition, output_artifact says its post-execution identity adds size, hash, and receipt, which would mutate the C4 node and invalidate edge, graph, checkpoint, and closure IDs after C6.

Required rework:

1. Make every node ID depend only on intrinsic immutable fields, or define graph-local logical references whose encoding is explicitly independent of other node hashes. Keep dependency, producer, consumer, and interop relations in the typed edge table as the single authority.
2. Keep the expected output node immutable after C4. Bind observed class, path, size, hash, and receipt in a separate C6/C7 observation or produced-artifact record keyed to that expected node.
3. Specify canonical validation for dangling or duplicate logical references and prove that non-ordering runtime cycles, action/output edges, and FFI boundaries cannot create recursive IDs.
4. Add goldens showing that C4, edge IDs, and closure ID remain unchanged when C6 attaches observed output bytes, while execution and publication receipt IDs change appropriately.

## R3 - Captured graph identity conflicts with the two-selection conformance vector

The design says package_instance identity includes feature, marker, condition, and target selection (line 116), and captured_graph_id includes target context (line 346). The graph envelope is also defined for target T and carries target_platform_id (lines 75-90). However `CGP05` requires the same captured graph for two feature, target, or marker selections when captured bytes are unchanged, with only selected graph and downstream identities changing (line 484).

Both outcomes cannot be implemented without a missing rule that separates selection-neutral capture identity from selection-specific active identity.

Required rework:

1. Define exactly which context belongs in `G_capture` and `captured_graph_id`, and keep requested target/feature/marker selection out of that identity if `CGP05` is normative.
2. Distinguish a selection-neutral captured package instance from peer/target/feature activation context, or explicitly explain how all contexts appear in one unchanged superset.
3. Add exact two-selection canonical bytes and digest goldens that prove the chosen rule.

## Routing decision

These are implementation-blocking architecture defects, but they are ordinary research and decision rework. They require no human-only product or platform decision and do not meet the Stop-The-Line boundary. Route `TASK-260810-1uu9lk` to `analysis`, revise the same task-scoped architecture outcome, rerun the artifact, DAG, board, and repository gates, and send it through a new reviewer cycle.