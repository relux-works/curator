# Reviewer verdict for TASK-260811-i3154q

Verdict: **changes requested -> to-dev**

## Goal, scope, and reviewed snapshot

- Reviewer run: `RUN-260811-cc0030`
- Authoritative reviewer goal at the last pre-verdict checkpoint:
  `GOAL-260811-b8347b` revision 1
- Resolved scope: `TASK-260811-i3154q`
- Review policy: `required`
- Reviewed implementation evidence:
  `TASK-260811-i3154q_implementation-evidence.md`, SHA-256
  `0636d0fe77d7718eead14e48577990cdb7d2e1ca840b1e8fe20526c312c01e06`
- Reviewed `internal/closuregraph` sorted per-file SHA-256 manifest:
  `f2ec25aa458d2a1993b136d09d9ab7cc84fc587c54fbe10d661cc7dc99abab29`
- Normative graph/checkpoint decision SHA-256:
  `874c3a40b9ff9fcf130f37f22e9f5aa2bdd6d9f246403e059c98e84736e27bbc`
- Exact CGP05/CGP10 corpus SHA-256:
  `fed9657b33297650e64178d605c7ee7f47dde640fefd12a99dec3651bb9cadcb`

This artifact records only the changes-requested branch. The findings are
package-local implementation defects with autonomous code/test remedies, not a
human-only or external blocker, so `blocked` is not appropriate.

## Acceptance-blocking findings

### 1. A target/package/product requirement on a produced output loses build order

The accepted edge contract permits `requires` from a product, package, target,
action, or boundary to an output and requires ordering for selected scopes that
consume prior build output. `deriveOrderingEdges` handles a direct action
requirement through the output producer table, but its owner-level pass looks
up providers only in `actionsByOwner[edge.ToNodeID]`
(`internal/closuregraph/plan.go:399-418`). An `output_artifact` is not an action
owner, so a target-level `requires(scope=build)` to a produced output emits no
provider-before-consumer arc.

Independent overlay probe
`TestReviewerProbeTargetRequiresProducedOutputOrdersConsumer` built a valid
selected target with a declared consumer action and a build requirement on an
exactly produced output. `DeriveBuildPlan` returned an empty ordering-edge
table. This can schedule the consumer in the same or an earlier wave than the
producer and violates the deterministic build projection.

Required rework: resolve owner-level requirement providers through exact
produced-output lineage as well as declared action ownership, retain canonical
edge evidence, and add target/package/product/boundary plus permutation and
cycle regressions.

### 2. Interop toolchain closure accepts a capture target as a fake toolchain binding

`SelectionBinding` is the sole authority for selection-specific external
toolchain records and toolchain-scoped `requires` edges. However,
`validateInteropBoundaries` adds every selected `requires` endpoint from an
interop boundary to `boundaryToolchains`, regardless of the owning table,
scope, or endpoint kind (`internal/closuregraph/validation.go:834-856`). It then
checks only that at least one endpoint exists and shares a platform
(`internal/closuregraph/validation.go:902-911`).

Independent overlay probe
`TestReviewerProbeInteropRejectsNonToolchainCaptureRequirement` removed the
valid binding edge and substituted a capture
`requires(scope=build)` from the `c_abi` boundary to a target on the same
platform. `ProjectActive` accepted it. Thus a wrong-kind capture record can
satisfy the explicit external-toolchain authority requirement.

Required rework: count only binding-owned
`requires(scope=toolchain)` edges whose endpoint is an authorized
`toolchain_component`, while retaining exact platform compatibility. Add
wrong-table, wrong-scope, wrong-kind, missing, and duplicate binding negatives.

### 3. C5 accepts ambiguous duplicate output/write paths

Active validation enforces one producer per output node, but it never enforces
path uniqueness across selected expected outputs
(`internal/closuregraph/validation.go:788-813`). `DeriveBuildPlan` merely sorts
the output node IDs and issues the plan (`internal/closuregraph/plan.go:254-300`).
Duplicate selected write paths are detected only while validating C6 execution
evidence (`internal/closuregraph/checkpoint_evidence.go:270-275,317-335`), after
C5 already authorized the action DAG.

Independent overlay probe
`TestReviewerProbeDuplicateExpectedOutputPathsRejectBeforeC5` supplied two
distinct immutable output nodes, each with an exact producer and publication,
but both at `bin/shared`. `DeriveBuildPlan` accepted both IDs. C5 therefore does
not bind an unambiguous exact declared output/write set.

Required rework: reject duplicate and conflicting selected output/write paths
before returning a `BuildPlan` or issuing C5, with canonical evidence and
permutation tests proving no later checkpoint is produced.

### 4. `resolves_to.artifact_manifest_id` is not resolved against capture authority

The canonical `resolves_to` edge binds package-to-source lock, origin,
checksum, and artifact-manifest evidence. `CaptureGraph` carries the closed
artifact-manifest authority. Yet `validateArtifactManifestRefs` checks only
`PackageInstancePayload` and `SourceSetPayload` references
(`internal/closuregraph/validation.go:586-600`); it ignores
`ResolvesToPayload.ArtifactManifestID`, even though that field is part of the
edge codec (`internal/closuregraph/edge.go:159-165`).

Independent overlay probe
`TestReviewerProbeResolvesToManifestMustResolveAndMatchSource` used valid
package/source node manifests but placed a third, absent manifest ID on their
`resolves_to` edge. Capture construction and active projection accepted the
graph. The graph can therefore claim immutable resolution evidence that is not
in its captured/admitted authority.

Required rework: resolve every edge-level artifact-manifest reference against
`CaptureGraph.artifact_manifest_ids` and enforce the profile-appropriate
package/source mapping without assuming equality where an explicitly captured
transform legitimately uses separate manifests. Add absent/mismatched and
permutation regressions.

### 5. Intrinsic validation is nondeterministic for the same invalid record

Many node, edge, and checkpoint payload validators iterate Go map literals and
return the first field error. For example,
`CommandProductPayload.validate` ranges over three fields at
`internal/closuregraph/node.go:383-391`; the package contains multiple similar
loops. Go map iteration order is unspecified.

Independent overlay probe
`TestReviewerProbeIntrinsicValidationFailureIsDeterministic` validated the
same node 1,000 times with invalid control characters in `skill_key`,
`command_key`, and `entry_point_contract`. It observed three distinct primary
errors:

- `skill_key must not contain control characters`
- `command_key must not contain control characters`
- `entry_point_contract must not contain control characters`

This violates deterministic canonical rejection and diagnostic ordering.

Required rework: validate fields in an explicit stable order (or collect and
canonically sort structured findings) throughout all codecs, then add repeated
and permutation tests for multiply-invalid records.

## Independent verification

The following current-snapshot gates passed:

- `go test -count=1 ./internal/closuregraph`
- exact CGP05/CGP10 and Go compatibility selectors
- focused interop/condition/platform regression selectors
- `go test -race -count=1 ./internal/closuregraph`
- `go test -count=1 -cover ./internal/closuregraph` — 80.6%
- `go test -shuffle=on -count=10 ./internal/closuregraph`
- `go vet ./internal/closuregraph`
- pinned `golangci-lint` v2.12.2 on `./internal/closuregraph/...` — 0 issues
- `gofmt -l internal/closuregraph` — no output
- accepted Ruby canonical oracle — 53 labeled records, both reference checks pass
- accepted Markdown corpus extraction is byte-identical to the embedded corpus
- `git diff --check`
- `task-board validate`

The overlay-only adversarial suite exited 1 with all five expected failures;
its source SHA-256 is
`7a7c45faf6dd57241b299c5803135872f4f73328a19879060db3276b4cbcc3b8`.
No product code was modified by this reviewer.

A fresh full-repository suite was deliberately not launched. Orchestrator
directives `nudge:3b6e50` and `nudge:7b08d3` reserve the exclusive full-suite
slot for sibling artifact-admission review and report that sibling package back
in active rework. The latter directive explicitly authorizes a
changes-requested verdict when closuregraph itself has blockers. Scoped tests
are sufficient to reproduce every finding above; a full suite would not turn
these accepted-invalid graphs into conformant behavior.

## Verdict route

Route `TASK-260811-i3154q` to `to-dev`. The next producer should close all five
findings, update task-scoped implementation evidence, run the focused and
adversarial regressions, and coordinate a stable full-repository gate before
another independent reviewer cycle.
