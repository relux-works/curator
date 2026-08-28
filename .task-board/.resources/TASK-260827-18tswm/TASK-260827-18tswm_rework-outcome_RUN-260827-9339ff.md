# TASK-260827-18tswm focused rework outcome

## Scope

Corrected reviewer item 2 only. The Rust-unavailable completeness exception is now a closed literal enumeration of these six manager-obligation/path pairs:

- `binding.diverges_per_target/rust`
- `binding.target_authority/rust`
- `capture.selection_neutral/rust`
- `capture.stable_across_targets/rust`
- `evidence.causal_chain/rust`
- `records.deterministic/rust`

`artifact.shared_admission/rust` is not in the exception and remains mandatory.

Added two negative tests at the production completeness-gate seam:

- `TestRustUnavailableCoverageRejectsExtraRustGap` rejects a future seventh Rust manager gap.
- `TestRustUnavailableCoverageRejectsNonRustGap` rejects a gap in the npm path.

Updated the existing LOGBOOK claim to name the six explicitly enumerated classes. No other accepted review item was reworked.

## Validation

| Command | Exit code | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/crossconformance` | 0 | `go-test-crossconformance.log` |
| `go test -count=1 -run 'TestRustUnavailableCoverageRejects(ExtraRustGap|NonRustGap)$' -v ./internal/crossconformance` | 0 | `go-test-negative-coverage.log` |
| `GOARCH=amd64 go test -count=1 -v ./internal/crossconformance` | 0 | `go-test-crossconformance-amd64.log`; Rust manager projection skipped for unapproved `x86_64-apple-darwin`, while shared artifact admission passed |
| `go build ./...` | 0 | `go-build.log` |
| `go vet ./internal/crossconformance/...` | 0 | `go-vet-crossconformance.log` |
| `gofmt -l internal/crossconformance/suite_test.go` | 0, no output | `gofmt.log` |
| `git diff --check` | 0 | `git-diff-check.log` |

No files were staged, committed, pushed, reset, or cleaned.

## Lifecycle blocker

The required developer handoff command refused with exit code 1 because checklist items 6 (full remote CI matrix plus passing URL) and 13 (overall AC) remain unchecked. This run cannot honestly check them: the focused rework brief forbids commit and push, the corrected bytes are therefore not on the PR branch, and reviewer RUN-260827-34e3b7 explicitly assigns fresh remote matrix evidence to the landing Orchestrator.

- Constraint: `task-board handoff` has no scoped or partial mode and fails closed until every checklist item is checked.
- Failed attempt: `task-board --no-update-check handoff TASK-260827-18tswm --role developer` exited 1 and left the status unchanged.
- Rejected workaround: checking the remote-CI and overall-AC items without a passing run URL would violate the evidence honesty contract.
- Viable path: the landing Orchestrator publishes/integrates the reviewed candidate, obtains and attaches a green full-matrix run URL and artifacts, checks items 6 and 13, then reruns the developer handoff or otherwise routes the evidence-backed review lifecycle.
- Exact external input needed: a passing CI run URL and its evidence artifacts for these corrected bytes on the PR branch.
