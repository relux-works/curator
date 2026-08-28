# TASK-260827-18tswm — macOS host-GOROOT isolation outcome

## Proven failure boundary

The new diagnostic from macOS run `33040614025` identifies the failing inner
boundary as `godriver` diagnostic `toolchain_mutated`: project 1's real probe
observed a content/size change in the selected host GOROOT while its whole-tree
fingerprint window was open.

The retained JSON timelines prove that the collision is with serial real
compiled-build activity in `cmd/curator`, not with the 36 tests changed to use
`t.Parallel`:

- Run `33040614025`: the dry-run compiled case ran from
  `04:52:43.931312Z` to its failure at `04:52:59.268006Z`.
  `TestAuthoritativeUpgradeCasesAreExecutable` ran real compiled upgrades from
  `04:52:45.344372Z` to `04:52:53.940172Z`, followed immediately by the real
  portable seed in `TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt`.
- The preceding macOS evidence has the same boundary: the dry-run compiled
  case began at `03:48:38.913279Z`; the all-project/global compiled upgrades
  and then `TestCLICompatibleVerifiedProviderOwnsBuildDispatchAndReceipt`
  overlapped it; it failed at `03:48:49.499237Z`.
- The parallel `cmd/curator` phase in run `33040614025` did not begin until
  `04:54:00.790568Z`, more than a minute after the dry-run failure. That
  falsifies the proposed direct `t.Parallel` explanation.

A repository-wide write/remove/rename search found no test that directly
installs into or mutates host GOROOT; all intentional mutation fixtures own a
`t.TempDir()`. The observed mutating activity is therefore the real Go-tool
execution against the writable hosted tool tree while another package process
holds a fingerprint over that same external fixture. The exact changed leaf is
not present in the old evidence stream and is not invented here.

## Fix

Added `internal/testtoolchain.LockHostGOROOT`, a Darwin-only cross-process test
lock backed by the existing `managerlock` implementation. It is acquired only
by tests that either:

1. hold a whole-host-GOROOT fingerprint across an operation; or
2. execute real Go tools from that same tree.

The lock is wired through the compiled project/global fixture helpers, the
authoritative compiled-upgrade case, the production-binary build case, and the
authoritative dry-run case. It does not serialize unrelated tests and does not
change Linux or Windows behavior.

Production `VerifyToolchain`, admitted outcomes, assertions, toolchain
identities, and release-pin promotion are untouched. The 36 parallelized tests
remain parallel.

## Reproduction and validation

All commands were run directly; no gate was piped through `tee`.

| Command | Exit | Evidence |
| --- | ---: | --- |
| Pre-fix two-package real-build overlap, `-count=20 -p=2 -parallel=16` | 0 | `.temp/TASK-260827-18tswm/go-test-ci-overlap-repro-01.log` |
| Post-fix same overlap shape, `-count=5 -p=2 -parallel=16` | 0 | `.temp/TASK-260827-18tswm/go-test-ci-overlap-after-lock-01.log` |
| `go test -count=1 -run 'TestAuthoritativeDryRunCasesMutateNothingPersistent' ./internal/install` | 0 | `.temp/TASK-260827-18tswm/go-test-required-dryrun-after-lock-01.log` |
| Focused `cmd/curator` compiled/toolchain mask | 0 | `.temp/TASK-260827-18tswm/go-test-cmd-host-toolchain-mask-01.log` |
| `go test -count=1 ./internal/testtoolchain` | 0 | `.temp/TASK-260827-18tswm/go-test-testtoolchain-01.log` |
| `go build ./...` | 0 | `.temp/TASK-260827-18tswm/go-build-all-after-lock-01.log` |
| `go vet ./...` | 0 | `.temp/TASK-260827-18tswm/go-vet-all-after-lock-01.log` |
| `golangci-lint v2.12.2 run` | 0 | `.temp/TASK-260827-18tswm/golangci-lint-after-lock-01.log` |
| `gofmt -l cmd internal` | 0, empty output | `.temp/TASK-260827-18tswm/gofmt-check-after-lock-01.log` |
| `git diff --check` | 0 | `.temp/TASK-260827-18tswm/git-diff-check-final-02.log` |

The pre-fix local pressure run remaining green is recorded deliberately: the
hosted mutation is timing/host-specific, so local success is not presented as
a reproduction of the failure. The post-fix run verifies the scoped ownership
mechanism and all relevant real entry points locally.

## Remaining remote evidence boundary

No files were staged or committed. A fresh full CI matrix, passing run URL, and
uploaded test-evidence artifacts require the landing Orchestrator to publish
the candidate branch. The full-matrix checklist item remains unchecked until
that exact remote evidence exists.

## Lifecycle blocker

`task-board --no-update-check handoff TASK-260827-18tswm --role developer`
was attempted after attaching this outcome and exited 1. The board refused
unchecked checklist items 6 and 13. Item 6 requires a fresh full CI run URL and
artifacts from a published candidate; checking it from local evidence would be
false. Item 13 cannot truthfully claim the complete acceptance criteria while
item 6 is absent.

- Constraint: this developer role is forbidden to stage or commit, while the
  Story landing Orchestrator owns candidate publication and remote CI.
- Failed approach: ordinary developer handoff; exact refusal was
  `handoff evidence missing` for checklist items `[6 13]`.
- Rejected tradeoff: mark the remote matrix green without a run. That would
  violate the evidence contract.
- Viable path: the landing Orchestrator publishes this candidate, runs the full
  matrix, attaches the passing URL and artifacts, checks items 6 and 13, then
  resumes the developer handoff. If the board owner intends remote CI to be a
  downstream post-review task instead, the checklist ownership must be changed
  explicitly; this worker will not silently reinterpret it.
- Recommendation/exact external input: publish the candidate and provide the
  passing full-matrix run plus test-evidence artifacts.
