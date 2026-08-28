# Reviewer verdict for TASK-260811-3twayo

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260822-d8c978`
- Goal check: `task-board spawn goal RUN-260822-d8c978` reported that the run is not goal-bound.
- Reviewed outcome: `TASK-260811-3twayo_fourth-rework-evidence.md`
- Rework baseline: `TASK-260811-3twayo_reviewer-verdict_RUN-260822-755ae4.md`

## Findings

1. **High — output reconciliation still trusts a caller-supplied C5 output subset instead of proving the exact plan derived from C4.** `ValidateOutputObservations` validates the supplied plan and checks only that `plan.ActiveGraphID` equals the bundle's active ID (`internal/nodesource/nodesource.go:830-840`). It never calls `closuregraph.DeriveBuildPlan` or compares the supplied plan identity with the independently derived C4 projection before using `plan.DeclaredOutputNodeIDs` as authority (`internal/nodesource/nodesource.go:855-866`). A read-only probe built a valid one-output Node graph, removed that output from a copy of the plan, and called `ValidateOutputObservations(nil, bundle, forgedPlan)`. The result was `real_outputs=1 forged_outputs=0 verdict=<nil>`. This permits an active output to disappear from Node reconciliation while the plan still names the same ActiveGraph. The shared `PublicationEvidence.ValidateForPublication` correctly performs the missing derivation/identity check, so the Node boundary must either consume that validated authority or independently rederive and compare the exact plan before checking observations. Add zero-output/subset/action/order forged-plan negatives, not only genuine-plan inactive-output cases.

2. **Medium — the Python corpus still does not model P10's required distinct target closures and is not a closed shared-schema graph oracle.** The accepted P10 contract requires two separately bound target graphs and rejection of reuse across interpreter/platform/ABI identity. The corpus instead places `cp313-linux` and `cp314-darwin` together in one admitted case and emits one outcome containing both selected target IDs (`internal/nodesource/testdata/python_protocol_shared_records.json:39`). Both implementations then hash package and target ID lists rather than deriving a canonical graph record with edges/binding/active identity (`internal/nodesource/testdata/python_protocol_golden.py:34-85`, `internal/nodesource/nodesource_test.go:118-145`). Consequently Go and Python agree on the same weakened abstraction but never prove distinct graph identities or cross-target reuse rejection. The decoders are also not closed for nested `lock`, `artifact`, and `build` objects: Python uses optional `.get(...)` fields and Go uses permissive `json.Unmarshal` into structs (`internal/nodesource/testdata/python_protocol_golden.py:47-71`, `internal/nodesource/nodesource_test.go:25-27,94-115`). Split P10 into exact target branches with distinct canonical graph/outcome identities plus a reuse-negative, emit/validate real canonical graph and diagnostic records, and reject missing/unknown nested schema fields independently in Go and Python.

## Closed portions of the fourth rework

- `CaptureInput.RootKeys` and `RuntimeBinding.TargetNodeIDs` are sorted and duplicate-checked before stable edge-key assignment; all-manager permutation and duplicate tests pass.
- Genuine C4/C5 plans now reconcile selected outputs rather than all capture outputs. One-of-two, two-of-two, condition-pruned absence, and inactive-observation rejection tests pass. The remaining defect is substitution of a structurally valid but non-derived plan.
- The Python oracle now independently derives diagnostic precedence and canonical outcome hashes from typed fixture inputs; this is meaningful progress over hand-hashing expected objects, but P10 and closed graph/schema proof remain incomplete.
- Prior exact-C0 tool-record matching, mandatory executable SHA evidence, graph-bound output declaration identity, generator chaining/cycle handling, lifecycle/native rejection, and selected/pruned/feature/peer/runtime/manager coverage remain green in focused review.

## Independent verification

- Fail-open probe: `go run ./.temp/TASK-260811-3twayo-review/output_plan_probe.go` — reproduced acceptance of a forged zero-output plan; log `.temp/TASK-260811-3twayo-review/output-plan-probe-01.log`.
- `go test -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — passed.
- `go test -race -count=1 ./internal/nodesource` — passed.
- `go vet ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — passed.
- `golangci-lint run ./internal/nodesource/...` — passed with `0 issues.`
- `go test -run '^$' ./...` and `go build ./...` — passed.
- Independent Python oracle — exited 0 for P01-P13 as currently encoded.
- Accepted Ruby canonical verifier — passed all 53 records and references.
- `git diff --check` and `task-board validate` — passed.
- The full uncached repository suite was not repeated after the deterministic acceptance failure. The producer evidence records a passing `go test -count=1 -timeout=20m ./...` run.

No product code was modified by this reviewer. As a reviewer-archetype run, it supplies no `commit_ack`.
