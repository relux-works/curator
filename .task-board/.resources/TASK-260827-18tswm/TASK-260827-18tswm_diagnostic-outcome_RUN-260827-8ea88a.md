# TASK-260827-18tswm — macOS dry-run diagnostic outcome

## Disposition

The last macOS CI failure cannot yet be assigned a behavioral fix without
inventing its hidden cause. Runs 33037369236 and 33039156879 both reported only
`toolchain-unavailable`; the strict conformance assertion discarded the planned
row's `Reason()` and `DiagnosticCode()`. This run changes both sibling outcome
assertions to print those fields while preserving the published admitted set
and the exact `would-preflight-and-build` requirement.

## Proven failure boundary

`toolchain-unavailable` is emitted only when `Toolchain.Probe` returns an error.
The delivery wrapper has two ordered sources: `authority.revalidate`, then the
inner toolchain probe. A new negative regression proves they are disjoint:

- an invalid authority refuses before the inner probe (zero inner calls);
- `NewPortableBuildAuthority()` reaches the inner probe exactly once and
  preserves its error.

The failing published case always obtains that exact portable authority from
`BuildDeps.resolve`. Its revalidation is a deterministic validation of a closed
in-memory binding and reads no host state. Therefore the CI failure came from
the inner real `godriver.Probe`, not authority revalidation. The old artifact
does not retain enough information to say whether the inner cause was process
start, environment, timeout, unreadable state, mutation, or another stable
driver diagnostic.

## Local evidence

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 -run 'TestAssuredToolchainProbeKeepsAuthorityAndInnerFailuresDisjoint' ./internal/install` | 0 | `.temp/TASK-260827-18tswm/go-test-probe-routing-01.log` |
| rc.9 `go test -count=1 -run 'TestAuthoritativeDryRunCasesMutateNothingPersistent' ./internal/install` | 0 | `.temp/TASK-260827-18tswm/go-test-dryrun-03.log` |
| install dry-run + real whole-GOROOT fingerprint, package-parallel, `-count=5` | 0 | `.temp/TASK-260827-18tswm/go-test-concurrent-probes-01.log` |
| `go build ./...` | 0 | `.temp/TASK-260827-18tswm/go-build-all-01.log` |
| `go vet ./internal/install/...` | 0 | `.temp/TASK-260827-18tswm/go-vet-install-01.log` |
| `gofmt -l internal/install` | 0, empty | `.temp/TASK-260827-18tswm/gofmt-check-01.log` |
| `golangci-lint run` | 0 | `.temp/TASK-260827-18tswm/golangci-lint-01.log` |
| `git diff --check` after all edits | 0 | `.temp/TASK-260827-18tswm/git-diff-check-02.log` |

One earlier command used the nonexistent `$PWD/protocol-spec` root and failed
with exit 1 before executing a case; this setup error is preserved in
`.temp/TASK-260827-18tswm/go-test-dryrun-01.log` and was rerun with the existing
exact rc.9 checkout at `.temp/TASK-260827-18tswm/spec-pin/conformance/v1`.

## External blocker and next action

The landing Orchestrator must push this diagnostic patch and rerun the macOS
Test lane. The resulting failure line will contain the exact reason and stable
diagnostic needed to implement a non-speculative fix. If it proves the existing
host limitation, use only the existing `no trusted Go toolchain is resolvable
here` class; otherwise fix that named inner-probe defect. Do not retry blindly,
relax the published outcomes, skip the case, or invent a toolchain identity.

No files were staged, committed, pushed, reset, or cleaned.
