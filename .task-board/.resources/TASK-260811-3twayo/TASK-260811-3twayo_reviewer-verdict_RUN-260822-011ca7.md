# Reviewer verdict for TASK-260811-3twayo

Verdict: **changes requested -> to-dev**

Reviewer run: `RUN-260822-011ca7`
Goal check: `task-board spawn goal` reported that this run is not goal-bound.
Reviewed outcome: `TASK-260811-3twayo_node-runtime-build-contract.md`
Rework baseline: `TASK-260811-3twayo_reviewer-verdict_RUN-260822-615b9e.md`

## Findings

1. **High — A metadata permit can name a C0-listed tool node while running a different executable identity.** `ExecuteMetadataDerivation` checks only that `Permit.ToolchainNodeID` occurs in the C0 ID list and that the checkpoint/subtype match (`internal/nodesource/nodesource.go:669-678`). It never resolves that node record or compares the permit's `ToolchainFingerprint`, `ExecutableSHA256`, `Executable`, execution domain, or recheck rule with the C0-bound `ToolchainComponentPayload`. The shared executor only compares the immediate recheck with the permit itself. Consequently, a permit can retain the manager node ID but substitute another fingerprint/path/digest, pass a matching recheck, and cross the process-start seam. This leaves reviewer finding 1 open and violates the exact C0/pre-C5 derivation contract. Require the exact tool record (or validated record tables) at this seam and reject every permit-field mismatch before `Commit`; add zero-start negatives for fingerprint, executable digest, path, domain, host/target, and recheck-rule substitution.

2. **High — Missing executable content evidence is silently accepted.** `toolNode` replaces an invalid or absent `ExecutableSHA256` with the broader tree fingerprint (`internal/nodesource/nodesource.go:479-482`). The task explicitly requires exact Node/manager executable evidence and fail-closed handling of missing/drifted bindings. A missing field must reject; it cannot be synthesized from a different identity field. Add missing-evidence tests for Node, manager, and compiler records.

3. **High — Final generated-output grammar is not bound into the closed graph.** For an intermediate, `GeneratedArtifactPayload` records `Grammar`; for a published output, `BuildGeneratedAction` drops `GeneratedOutput.Grammar` and creates an `OutputArtifactPayload` containing only path/class/role and an opaque caller-provided `DeclarationDigest` (`internal/nodesource/nodesource.go:560-579`). Empty grammar can therefore enter capture/C4/C5, and `ValidateOutputObservations` later compares the observation with a separate caller-supplied `[]GeneratedOutput`, not with graph-bound grammar evidence (`internal/nodesource/nodesource.go:703-729`). The same graph/plan can be validated under changed grammar if the caller reuses the opaque digest. Derive and verify the declaration digest from the complete path/class/grammar/role contract (or persist grammar in a closed graph record), reject incomplete declarations before C4, and validate observations against that graph-bound declaration.

4. **Medium — Multiple declared generators cannot form a real action DAG.** `AddGeneratedActions` initializes `known` from the original capture, rejects every input absent from that set, and after each action adds only the action ID—not its generated/output node IDs (`internal/nodesource/nodesource.go:592-630`). A second declared action therefore cannot read the first action's generated artifact, even though the shared graph and build planner support generated-artifact read edges and deterministic ordering. Use a two-pass canonical declaration/indexing scheme, validate references after all declared outputs exist, and add permutation, chaining, and cycle tests.

5. **Medium — The mandatory named conformance remains incomplete.** The Node-named CGP05 test asserts only relative identity changes for custom records, not the accepted exact CGP05 target records/digests. N10 exercises only a Linux-selected condition; it does not prove the corresponding pruned branch, feature selection, or distinct peer-context projection. Runtime and manager identities are changed together with the platform, so independent runtime-only and manager-only binding/plan changes are not demonstrated. The Python file now emits 13 labels, but they remain hand-authored semantic hash payloads rather than a protocol corpus that decodes/validates the shared schemas or compares canonical graph/diagnostic outcomes. Add the exact accepted-record assertions and the missing selected/pruned, feature, peer, independent binding-drift, and protocol-schema compatibility cases required by the rework brief.

## Verification

Independent reviewer gates:

- `go test -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — passed.
- `go test -race -count=1 ./internal/nodesource` — passed.
- `go test -cover ./internal/nodesource` — passed, 81.9% statement coverage.
- `go vet ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — passed.
- `golangci-lint run ./internal/nodesource/...` — passed with 0 issues.
- `go test -run '^$' ./...` and `go build ./...` — passed repository-wide compilation/build.
- Accepted Ruby canonical verifier — passed: 53 records, two CGP05 target branches, two CGP10 observation branches, all references resolved.
- `git diff --check` — passed.
- `task-board validate` — passed.

The green suite confirms that the landed tests pass; it does not cover the fail-open substitutions and missing graph-bound evidence above. No product code was modified by this reviewer.
