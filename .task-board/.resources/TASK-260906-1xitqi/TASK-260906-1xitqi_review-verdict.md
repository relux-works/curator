# TASK-260906-1xitqi reviewer verdict

## Verdict: changes requested

The dependency and release-source gate are technically sound, but the candidate does not yet carry meaningful production-wiring tests and Change Request revision 1 is already `changes_requested` after the configured validation suite exited 1. Do not publish or tag this revision.

## F1 — release production-wiring assertion admits bypasses

Severity: blocking for this release-readiness task.

`.github/ci/gate-selftest.sh` locates the release workflow's fresh-main fetch and release gate and asserts only `fetch_line < gate_line`. It does not locate the GoReleaser publication step or prove `gate_line < publish_line`, despite naming the assertion as proof that the gate runs "before GoReleaser". It also does not assert that normal CI invokes `.github/ci/release-source-gate.sh`.

Independent isolated mutants both left the complete self-test green at 91/91:

1. Inserted an additional `goreleaser/goreleaser-action@v6` publication step before the fetch and gate. The claimed release-order assertion still printed `ok`, and the suite exited 0.
2. Removed the `Verify versioned Go install compatibility` step from `.github/workflows/ci.yml`. The suite again reported 91 passed / 0 failed and exited 0.

Required rework: make the workflow test derive and assert the complete protected path. At minimum, require exactly the intended release-source gate invocation in normal CI; require the fresh-main fetch before the ancestry gate; and require that gate before every GoReleaser publication action. Add negative/mutant cases that fail when an earlier publication bypass is inserted and when either production invocation is removed. Avoid a hand-picked assertion that can be satisfied by one of multiple publisher steps.

## Candidate checks and bounded residual

- `bash .github/ci/release-source-gate.sh`: exit 0.
- `bash .github/ci/gate-selftest.sh`: exit 0, 91/91 on the unmutated candidate.
- `go mod verify`: exit 0.
- `go mod download -json github.com/relux-works/skill-go-testing-tools/tuitestkit@v0.1.1`: exit 0 with the exact `go.sum` module and go.mod sums.
- `go test -count=1 ./internal/ui`: exit 0 against published `tuitestkit v0.1.1`.
- Hosted baseline run 34015435043 is success at exact main SHA `7320bc2adbd15aa5ade4e78ef7ef9008274e7478`, including macOS/Linux/Windows Test and gate jobs. This is baseline evidence only, not candidate-SHA CI.
- The configured Change Request validation ran build and vet successfully, then `go test -count=1 -timeout 30m ./...` exited 1 on local capture-store rename/cleanup permission failures. The board therefore recorded CR revision 1 as `changes_requested`; there is no reviewable accepted revision to pass to `accept_cr`.
- I independently archived pristine `7320bc2` and ran `go test -count=1 ./internal/closureexec -run '^TestImmutableAdmittedTreeReplayAndTimeOfUseRechecks$'`; it failed with the same read-only capture-tree rename and cleanup `permission denied` shape. This establishes that focused failure as pre-existing on this host and not caused by the packaging delta. It does not make the full suite green and is not a waiver.

The pre-existing local host failure does not, by itself, show a packaging regression. Nevertheless, the mandatory Change Request gate remains red and must not be represented as passing. Republish only after F1 is fixed and the configured validation can produce a reviewable revision without weakening or silently changing the suite; the parent still requires hosted CI on the exact integrated SHA before publication.

The capture-store failure is confirmed as the current mechanical blocker to Change Request construction on this host. Route it as focused, separately evidenced rework: preserve the immutable/custody boundary while allowing atomic tree publication and cleanup on this macOS permission model, and require the named pristine-main failure plus the affected capture-store package tests to turn green. Do not relax the read-only invariant, skip the tests, or alter the validation configuration. Once that fix is integrated into the Story base, rebuild this packaging candidate and rerun the configured validation rather than rerunning the unchanged red suite.

## Other review conclusions

- Removing the root-module `replace` and selecting published `tuitestkit v0.1.1` resolves the concrete source-module blocker pre-publication; actual `go install ...@latest` remains a required post-publication check.
- The release source gate correctly rejects `replace`, `exclude`, missing module input, incomplete ancestry inputs, unresolved refs, and candidate commits outside main ancestry.
- GoReleaser config covers six archive targets, four Linux packages, six archive SBOMs, checksums/signature/certificate, Homebrew cask, and Scoop; publication remains parent-owned.
- The public-channel inventory, rc.2/rc.3 no-release finding, stable `v0.14.0` recommendation, and verification plan are adequately documented in the producer evidence.
