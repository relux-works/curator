# Reviewer verdict for TASK-260811-3twayo

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260822-c1073c`
- Goal check: `task-board spawn goal RUN-260822-c1073c` reported that the run is not goal-bound.
- Reviewed outcome: `TASK-260811-3twayo_fifth-rework-evidence.md`
- Rework baseline: `TASK-260811-3twayo_reviewer-verdict_RUN-260822-d8c978.md`

## Finding

1. **Medium — the independent Python P10 oracle still does not emit separately bound target outcomes through the shared canonical graph contract.** The fixture still represents P10 as one case containing two targets (`internal/nodesource/testdata/python_protocol_shared_records.json:527-568`), and each implementation derives one `python-protocol-outcome-v1` identity that contains two bindings and two active graphs (`internal/nodesource/nodesource_test.go:291-382`, `internal/nodesource/testdata/python_protocol_golden.py:67-150`). The Go assertion checks that the two nested binding and active IDs differ, but it never requires two distinct target-scoped outcome IDs (`internal/nodesource/nodesource_test.go:67-70`). This does not satisfy the rework requirement that P10 represent separately bound target closures with distinct canonical graph/outcome identities and reject cross-target reuse.

   The records are also a parallel reduced protocol rather than the shared canonical graph schemas: `python-protocol-capture-graph-v1` contains only `schema_id`, `node_ids`, and `edge_ids`; `python-protocol-selection-binding-v1` contains only `schema_id`, `captured_graph_id`, and one `target_node_id`; and `python-protocol-active-graph-v1` contains node/edge ID lists instead of the accepted selection context, binding node/edge sets, activation records, roles, and SCC contract (`internal/nodesource/nodesource_test.go:285-315`, `internal/nodesource/testdata/python_protocol_golden.py:64-84`). Thus Go and Python independently hash matching custom objects, but they do not independently decode and validate the real shared capture/binding/active schemas required at the protocol boundary.

   Rework: encode P10 as two target-scoped canonical outcomes (one per interpreter/platform/ABI binding) with distinct outcome identities over one selection-neutral capture, then a separate reuse-negative referencing those exact bindings. Both Go and Python must independently decode, validate, canonicalize, and hash the accepted shared capture, selection context, selection binding, active graph, and diagnostic wire shapes rather than adapter-local abbreviated substitutes. Preserve the now-correct exact nested `lock`, `artifact`, and `build` field rejection.

## Closed portion of the fifth rework

- `ValidateOutputObservations` now independently calls `closuregraph.DeriveBuildPlan` from the supplied C4 bundle and compares exact plan identity before trusting declared outputs (`internal/nodesource/nodesource.go:829-849`). The zero-output, subset-output, action-set, and ordering substitutions reject with `closure_generated_output_drift`; genuine selected and pruned output cases remain accepted.
- Missing and unknown nested fields in `lock`, `artifact`, and `build` are rejected independently by both Go and Python.
- Previously closed exact-C0 tool binding, mandatory executable SHA, graph-bound output grammar, generator chaining/cycles, canonical root/target ordering, lifecycle/native rejection, and manager-profile coverage were not reopened by this review.

## Independent verification

- `go test -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — passed; log `.temp/TASK-260811-3twayo-review-focused-02.log`.
- `go test -count=1 ./internal/nodesource -run 'TestOutputObservationsRejectForgedBuildPlans|TestOutputObservationsUseExactActivePlanSet|TestConditionPrunedOutputIsAbsentAndCannotBeObserved'` — passed; log `.temp/TASK-260811-3twayo-review-forged-plan-01.log`.
- `python3 internal/nodesource/testdata/python_protocol_golden.py` — exited 0 and exposed one P10 outcome ID containing two nested bindings/active graphs; log `.temp/TASK-260811-3twayo-review-python-oracle-01.log`.
- Accepted Ruby canonical verifier — passed all 53 records and references; log `.temp/TASK-260811-3twayo-review-canonical-01.log`.
- `git diff --check` — passed before review; no product code was modified by this reviewer.

The producer's full uncached repository suite, race, vet, lint, build, diff, and board gates are recorded as passing in the reviewed fifth-rework evidence. They were not repeated after the deterministic protocol-contract acceptance failure.

No product code was modified by this reviewer. As a reviewer-archetype run, it supplies no `commit_ack`.
