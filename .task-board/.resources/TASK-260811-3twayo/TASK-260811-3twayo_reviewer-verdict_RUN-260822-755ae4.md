# Reviewer verdict for TASK-260811-3twayo

Verdict: **changes requested -> to-dev**

- Reviewer run: `RUN-260822-755ae4`
- Goal check: `task-board spawn goal RUN-260822-755ae4` reported that the run is not goal-bound.
- Reviewed outcome: `TASK-260811-3twayo_third-rework-evidence.md`
- Rework baseline: `TASK-260811-3twayo_reviewer-verdict_RUN-260822-011ca7.md`

## Findings

1. **High — Multi-root capture and multi-target binding identities depend on caller order.** `BuildCapture` sorts package instances but iterates `CaptureInput.RootKeys` in caller order and assigns `node.product.requires.%04d` edge keys from that index (`internal/nodesource/nodesource.go:197-220`). `Bind` likewise iterates `RuntimeBinding.TargetNodeIDs` without canonicalization and assigns `node.targets.%04d` keys from caller order (`internal/nodesource/nodesource.go:386-417`). A read-only probe using the same two workspace roots produced capture IDs `sha256:c2c9327f...` and `sha256:0ace5a2a...` when only `RootKeys` order changed. With one unchanged capture/selection/C0 and the same two target node IDs, reversing only `TargetNodeIDs` produced binding IDs `sha256:ded63166...` and `sha256:aa698296...`. This violates the accepted rule that discovery order is excluded from portable identity and breaks canonical two-root/two-target manager-independent records. Canonicalize and uniquely validate both collections before stable edge-key assignment; add all-manager two-root and multi-target permutation tests proving identical capture, binding, active, and plan identities for the same semantic sets.

2. **High — Output reconciliation validates every captured output instead of the selected C4/C5 output set.** `ValidateOutputObservations` collects every `output_artifact` in `RecordTables.CaptureNodes` and requires that count to equal observed outputs (`internal/nodesource/nodesource.go:808-845`). The shared projector explicitly prunes unreachable nodes per `SelectionContext`, and `DeriveBuildPlan` includes only selected action/output IDs. A capture containing generated outputs for two products therefore cannot publish a valid one-product selection: the Node validator requires the inactive product's output too, or rejects the exact active observation set before the shared publication contract. Bind output validation to the exact `ActiveGraph`/`BuildPlan.DeclaredOutputNodeIDs` (or an equivalent independently validated active-output authority), and add one-of-two/two-of-two product plus condition-pruned output tests. Inactive output observations must reject; inactive output absence must succeed.

3. **Medium — The Python P01-P13 artifact is still mostly hand-authored outcome hashing, not independent protocol outcome validation.** The Python oracle now validates four accepted CGP05 graph records, which is useful, but for P01-P13 it only checks that each `expected` object is nonempty, optionally checks a diagnostic-code prefix, then hashes that same hand-authored object (`internal/nodesource/testdata/python_protocol_golden.py:26-35`; `python_protocol_shared_records.json`). It does not decode a shared diagnostic schema, validate checkpoint/graph references for those cases, or compare canonical graph/diagnostic outcomes derived from fixture inputs. This does not fully close the mandatory rework requirement for an independent Python protocol oracle that validates shared schemas or real canonical outcomes rather than ad hoc hashes. Supply protocol fixture inputs and exact canonical expected graph/diagnostic records, then have both implementations independently decode/validate/compare them; keep Python code and mutable state separate.

## Closed prior findings

The third rework correctly closed the prior exact-C0 tool-record substitution seam, removed executable-digest fallback, bound final output grammar through a derived declaration identity, implemented two-pass generator indexing with chain/cycle coverage, and added selected/pruned, feature, peer, runtime-only, and manager-only identity tests. Those changes are not requested again.

## Independent verification

- `go test -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — passed.
- `go test -race -count=1 ./internal/nodesource` — passed.
- `go vet ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — passed.
- `golangci-lint run ./internal/nodesource/...` — passed with `0 issues.`
- `go test -run '^$' ./...` — passed repository-wide compile gate.
- `go build ./...` — passed.
- Accepted Ruby canonical verifier — passed all 53 records and references.
- Independent Python command reproduced the checked-in 13-line corpus exactly.
- `git diff --check` — passed.
- `task-board validate` — passed.
- Reproduction logs: `.temp/TASK-260811-3twayo-review-order-probe-02.log`; validation logs use the `.temp/TASK-260811-3twayo-review-*-01.log` task prefix.

The full uncached repository suite was not repeated after the deterministic acceptance failures above; the producer's attached third-rework evidence records a passing full uncached run. No product code was modified by this reviewer.
