# Rework required after reviewer RUN-260822-755ae4

Implement only the remaining findings from `TASK-260811-3twayo_reviewer-verdict_RUN-260822-755ae4.md`, preserving the prior closures.

1. Canonicalize and uniquely validate `CaptureInput.RootKeys` and `RuntimeBinding.TargetNodeIDs` before stable edge-key assignment. Add all-manager two-root and multi-target permutation tests proving identical capture, binding, active-graph, and build-plan identities for the same semantic sets.
2. Reconcile output observations against the exact active C4/C5 output set (`ActiveGraph` / `BuildPlan.DeclaredOutputNodeIDs`), not every output in the capture. Cover one-of-two and two-of-two product selections plus condition-pruned outputs. Missing inactive outputs must succeed; observing inactive outputs must fail.
3. Replace the hand-authored P01-P13 outcome-hash check with an independent schema-aware Python protocol oracle. Provide protocol fixture inputs and exact canonical graph/diagnostic outcomes, and have Go and Python independently decode, validate, derive, and compare those outcomes. Keep Python code and mutable state separate.

Do not regress the already accepted exact-C0 tool-record binding, mandatory executable SHA, graph-bound final output grammar, two-pass generator chaining/cycle handling, or selected/pruned/feature/peer/runtime/manager conformance coverage.

Run focused, race, vet, lint, repository compile/build, canonical verifier, Python oracle, diff, and board-validation gates. Run the full uncached repository suite unless a deterministic acceptance failure makes it irrelevant. Attach task-scoped evidence and hand off with the developer role.
