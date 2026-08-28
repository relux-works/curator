# Reviewer verdict for TASK-260811-i3154q

Verdict: **changes requested -> to-dev**

## Goal and reviewed scope

- Reviewer run: `RUN-260811-273b9c`
- Authoritative reviewer goal at the verdict checkpoint: `GOAL-260811-618c86` revision 1
- Resolved scope: `TASK-260811-i3154q`
- Review policy: `required`
- Reviewed implementation: `internal/closuregraph`
- Reviewed package-tree manifest digest: `sha256:e54221d110a44dd5f0dda788c469ff7d2cdd34ba62105fa4093b40f266a6cfe2`
- Producer evidence: `TASK-260811-i3154q_implementation-evidence.md`, `sha256:f4544a76f8e494307201cf58e4cebb5c4f4c097c109d9029918bd71d542d0dab`
- Normative graph/checkpoint contract: `TASK-260810-1uu9lk_accepted-cross-language-closure-graph-and-checkpoints.md`, `sha256:874c3a40b9ff9fcf130f37f22e9f5aa2bdd6d9f246403e059c98e84736e27bbc`

This artifact records only the changes-requested branch. The implementation is substantial and its positive and exact-golden suites are green, but the following independently reproduced fail-closed gaps violate the accepted graph/checkpoint contract and task acceptance criteria.

## Acceptance-blocking findings

### 1. C5 accepts a self-consistent plan that was not derived from the active graph

`ValidateCheckpointChain` validates the supplied plan structurally and checks that C5 names that plan and that the plan names the supplied active-graph ID (`checkpoint_evidence.go:197-216`). It does not rederive the plan from `CheckpointChainEvidence.Graph`, nor resolve every plan action/output/ordering record to the selected active graph.

The probe appended `sha256:ffff...ffff` to `ActionNodeIDs`, rebuilt the waves, C5, closure, expected-cache input, execution receipt, publication receipt, and C6/C7 checkpoint IDs consistently, while leaving the supplied graph unchanged. `ValidateCheckpointChain` returned nil. An action with no graph declaration, slots, target, tool, input, or output authority can therefore enter the trusted C5-C7 chain.

Required rework: derive or equivalently validate the entire C5 plan from the unchanged graph and execution policy, including exact action/output sets, ordering edges, and waves; reject every plan reference absent from or inconsistent with the active graph. Add a regression that mutates the plan and all downstream identities together while preserving C4.

### 2. Selected expected outputs can enter a plan without producer lineage

`DeriveBuildPlan` includes every selected `output_artifact` (`plan.go:250-272`). `deriveOrderingEdges` records selected `produces` edges and rejects only multiple producers (`plan.go:284-301`); it never rejects zero producers. Endpoint validation checks path/class consistency but no selected-producer cardinality (`validation.go:666-700`).

The probe selected a command product and expected output joined only by `publishes`. `GraphBundle.Validate` and `DeriveBuildPlan` accepted the graph and emitted the output in the plan with zero actions and no `produces` edge. This contradicts the causal output model and the accepted `CGN04` producer-lineage failure.

Required rework: require exact producer lineage for each selected local generated/output artifact that is consumed, used, or published, with canonical zero/multiple-producer diagnostics before C5. Add positive and negative lineage tests.

### 3. A target-level generated read does not create provider-before-consumer order

The closed edge schema deliberately allows `target_unit -> reads -> generated_artifact` (`validation.go:823-825`), and the accepted contract says a generated/local-output producer precedes the reader. The ordering projection skips every `reads` edge whose origin is not itself an action (`plan.go:313-329`), despite already computing target-to-action ownership in `actionsByOwner`.

The probe created provider action -> `produces` -> generated artifact and target -> `reads` -> that artifact, with a selected action declared by the target. The resulting plan had no ordering edge, allowing both actions into the same wave.

Required rework: project target-origin generated/local-output reads to the selected action or actions that execute that target, or reject an ambiguous target-level shape; the dependency must not disappear. Preserve canonical evidence-edge IDs and deterministic cycle/wave behavior in permutation tests.

### 4. An explicit empty platform-role array suppresses mandatory action targeting

`ActionPayload.declaredPlatformRoles` returns any non-nil `PlatformRoleNames` verbatim (`node.go:593-601`), so `[]` overrides the target/host default implied by `ExecutionDomain`. `validatePlatformBindings` then has no role to enforce (`validation.go:754-776`). Target units and toolchains use the same override pattern (`node.go:535-543`, `717-725`).

The probe selected an executable target-domain action with `platform_role_names: []` and no `targets` binding. `GraphBundle.Validate` returned nil. This bypasses exact target identity even though the accepted contract requires one target edge for each selected executable slot and treats missing target binding as a canonical failure.

Required rework: reject explicit-empty or incomplete platform-role declarations wherever kind/execution domain requires target or host identity, and test every affected node kind plus missing/duplicate/role-mismatched edges.

### 5. Selected interop boundaries need neither sides nor their mode-required evidence

`InteropBoundaryPayload.validate` checks the enum, sorted list shape, contract digest, and only the subprocess protocol-schema special case (`node.go:622-648`). Graph validation does not require a selected boundary to have provider and consumer edges or validate the accepted per-mode evidence and compatibility rules.

The probe selected a `c_abi` boundary with empty provider/consumer language lists, no ABI, interface, calling convention, or link/load evidence, and zero `provides_interop`/`consumes_interop` edges. `GraphBundle.Validate` returned nil. The record does not represent an interface shared by a provider and consumer and cannot support the required build order or target/toolchain compatibility proof.

Required rework: enforce the closed required contract for each interop mode, require selected provider and consumer sides with canonical cardinality/compatibility checks, and verify target/toolchain and ordering semantics. Add zero-side, duplicate-side, missing-mode-evidence, incompatible-side, and permutation regressions.

### 6. Canonical codecs silently normalize wrong-typed optional fields

Numerous node and edge decoders discard errors returned by `optionalString`, including `node.go:914-920`, `937-938`, `1009`, `1055-1060`, `1088-1091`, `1113`, `1139-1140` and `edge.go:607`, `616`, `627`, `644`, `683`.

The probe supplied canonical JSON for a package node with `"manager":1`. `DecodeNode` returned success, discarded the field, and re-encoded different canonical bytes. Thus a schema-invalid record is accepted as a different record instead of failing canonically. Edge payloads have the same error-discard pattern.

Required rework: propagate every decoder-helper error for required and optional fields and prove that a successful decode preserves the exact accepted canonical record. Add wrong-type mutations across all optional node/edge fields, not only one representative.

## Reproduction evidence

The review used an ignored Go overlay so no product or repository test file was modified:

```text
go test -count=1 -overlay=.temp/reviewer-probes/TASK-260811-i3154q_overlay.json ./internal/closuregraph -run '^TestReviewerProbe' -v
```

The command exited 1 because all six assertions expected the invalid form to be rejected but the implementation accepted it:

```text
DecodeNode accepted a wrong-typed optional field and silently normalized it
ValidateCheckpointChain accepted foreign action sha256:ffff... absent from the supplied active graph
DeriveBuildPlan accepted a selected expected output with no producer lineage
target-level generated read did not order provider before target action: []closuregraph.OrderingEdge{}
GraphBundle accepted a selected executable action with no target/host binding
GraphBundle accepted a selected c_abi boundary with no provider, consumer, languages, ABI/interface, or link evidence
```

## Passing evidence

The following gates passed independently and should remain green through rework:

- `go test -count=1 ./internal/closuregraph`
- `go test -count=1 -race ./internal/closuregraph`
- `go test -count=1 -cover ./internal/closuregraph` (`80.7%`)
- `go test -shuffle=on -count=10 ./internal/closuregraph`
- `golangci-lint run ./internal/closuregraph/...` (`0 issues`)
- `go vet ./internal/closuregraph`
- `go build ./internal/closuregraph`
- `gofmt -d internal/closuregraph` (no diff)
- exact accepted-record extraction is byte-identical to `internal/closuregraph/testdata/canonical-goldens.txt`, whose SHA-256 is `fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb`
- the accepted Ruby oracle reports `canonical_goldens=pass labeled_records=53` and `canonical_references=pass`
- `go test -count=1 ./...` exited 0 across every repository package; notable timings: `cmd/curator` 422.945s, `internal/artifactpolicy` 140.265s, `internal/install` 127.801s, `internal/install/atomicity` 125.197s
- `git diff --check` and `task-board validate` passed

## Next review gate

Rework all six findings and add their adversarial tests to the package. The next reviewer should rerun the existing exact corpus, race/shuffle/lint/build/full-suite gates and verify that each malformed/under-specified probe now rejects at C4 or C5 with the required stable diagnostic and no later checkpoint or action.

No product code, tracked tests, or configuration was modified by this reviewer. As a reviewer-archetype run, this verdict supplies no `commit_ack`.
