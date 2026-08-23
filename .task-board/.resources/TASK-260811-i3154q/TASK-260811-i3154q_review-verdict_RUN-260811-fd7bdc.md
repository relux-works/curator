# Reviewer verdict for TASK-260811-i3154q

Verdict: **changes requested -> to-dev**

## Goal and reviewed scope

- Reviewer run: `RUN-260811-fd7bdc`
- Authoritative goal immediately before verdict: `GOAL-260811-254c8e` revision 1
- Resolved scope: `TASK-260811-i3154q`
- Review policy: `required`
- Directive checkpoint: no directives recorded
- Producer evidence SHA-256: `6b0cb05d3e568d9ba216c8013a900df6aa99ffb951f56bea1585cd6998e89fd7`
- Reviewed `internal/closuregraph` sorted per-file SHA-256 manifest digest: `2746cd5eb178b1127ea26a41af572d474d817f377c508b3c562ebe0212893070`

## Findings requiring implementation rework

1. **C0 and the checkpoint chain do not enforce the accepted cross-record trust predicate.** `C0ProfilePayload.validate` in `internal/closuregraph/checkpoint.go:400` checks only ID syntax and membership in an ID list. It cannot reject a wrong-kind platform record or prove that an evidence toolchain ID resolves to a complete `toolchain_component`. `validateToolchainAuthority` in `internal/closuregraph/validation.go:412` accepts an arbitrary valid evidence ID plus a C0/C4 enum and never reconciles it with the C0 payload or a C4 selector. `ValidateCheckpointChain` in `internal/closuregraph/checkpoint.go:335` checks names, predecessors, and C0/C1/C4 selection identity only. It therefore accepts, among other invalid states, different C2 and C3 intake sets and stale or unrelated C4 graph, C5 plan, C6 execution, or C7 publication references. This contradicts the accepted C0 wrong-kind/toolchain gate and C1-C7 causal aggregation rules. Add typed cross-record validation that resolves C0 platform/tool records and reconciles every checkpoint reference with the supplied capture, admission, graph, plan, observation, execution, and publication records. Add negative tests proving no later checkpoint is accepted for each mismatch.

2. **A valid graph and plan can still contain hidden action references.** `ActionPayload.validate` in `internal/closuregraph/node.go:545` treats argv templates as arbitrary text. `validateActionSlots` in `internal/closuregraph/validation.go:467` counts declared edge slots but never parses `$TOOL(...)`, `$READ(...)`, or `$WRITE(...)`, and never verifies that `uses_tool.executable_relative_path` matches the selected toolchain or local-tool output. It also does not close produces path/class consistency before C5; `ProducedArtifactObservation.ValidateAgainst` in `internal/closuregraph/receipts.go:114` ignores a nonempty `ProducesPayload.WriteClass`. An action using `$TOOL(undeclared)` with no declared tool slot, or a bound tool edge naming a path different from its toolchain node, can pass graph validation and planning. Implement a closed placeholder grammar, exact template-to-slot equality, endpoint path/class reconciliation, and fail-closed negative tests before C5.

3. **Condition projection is incomplete and its failure order is not canonical.** `TargetUnitPayload` accepts `ConditionExpressions` at `internal/closuregraph/node.go:148`, but only `RequiresPayload.condition` returns a condition (`internal/closuregraph/edge.go:257`), so target-unit conditions are hashed but never evaluated or represented by selected/pruned activation evidence. In addition, `ProjectActive` evaluates the `resolved.captureEdges` map at `internal/closuregraph/projection.go:75` and sorts IDs only after evaluation; with multiple invalid/effectful evaluators, the first failure and call order depend on Go map iteration. Either project every accepted condition placement into exact activation evidence or reject unsupported placements, and sort conditional IDs before any evaluation. Cover target/feature branches and failing permutations.

4. **Ordering classification is broader than the accepted non-ordering boundary.** `edgeIsNonOrdering` in `internal/closuregraph/projection.go:235` categorically treats development, optional, and workspace requirements as non-ordering, while target ordering in `internal/closuregraph/plan.go:333` is emitted only for build/tool/toolchain/package-artifact scopes. The accepted contract says runtime/peer relations may be non-ordering; any selected relation consuming prior build/materialization output must order provider before consumer. A selected optional/workspace target dependency can therefore enter the same wave as its provider. Also, cycle evidence uses undirected graph reachability at `internal/closuregraph/plan.go:536`, so shared platform nodes can mark unrelated products/targets as affected. Derive ordering from actual consumption semantics, test selected optional/workspace provider dependencies, and derive affected cycle scope without platform-hub widening.

## Passing evidence retained

- Exact embedded corpus is byte-identical to the 53 accepted CGP05/CGP10 records: SHA-256 `fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb`.
- Accepted Ruby oracle passed: `canonical_goldens=pass labeled_records=53` and `canonical_references=pass`.
- `go test -count=1 -race ./internal/closuregraph`: exit 0.
- `go test -count=1 -cover ./internal/closuregraph`: exit 0, 80.5 percent statements.
- `go vet ./internal/closuregraph`, `go build ./internal/closuregraph`, task-scoped `golangci-lint`, `gofmt -d`, and `git diff --check`: exit 0 / no findings.
- `go test -count=1 ./...`: exit 0 across every package; `cmd/curator` 367.578s.
- `task-board validate`: exit 0.

Green tests establish compatibility but do not exercise the accepted-invalid states above. Re-review requires regression vectors for every finding plus the exact corpus/oracle, focused race/coverage/lint/vet/build, and repository test gates. No product code was modified by this reviewer.