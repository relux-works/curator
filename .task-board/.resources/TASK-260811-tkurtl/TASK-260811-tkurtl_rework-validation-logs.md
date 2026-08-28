# TASK-260811-tkurtl rework validation evidence

All commands were run directly as standalone processes from the worktree root
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-agent-skill` and their real
exit codes are reported. No command was piped through `tee`.

## Gates

| Gate | Command | Exit | Result |
| --- | --- | ---: | --- |
| Format | `gofmt -l ./cmd ./internal` | 0 | no output |
| Build | `go build ./...` | 0 | clean |
| Vet | `go vet ./...` | 0 | clean |
| Lint (pinned v2.12.2) | `golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/... ./internal/closuregraph/...` | 0 | `0 issues.` |
| Focused suite + coverage | `go test -count=1 -cover ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/ ./internal/artifactpolicy/` | 0 | interop **86.0%**, swiftpmsource 80.0%, closuregraph 80.7%, closureexec 58.0%, artifactpolicy 75.6% |
| Race (focused) | `go test -race -count=1 ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/ ./internal/artifactpolicy/` | 0 | 5 ok, no data race |
| Repository suite minus `cmd/curator` | `go test -timeout 9m -count=1 $(go list ./... \| grep -v cmd/curator)` | 0 | **51 ok**, 0 FAIL |
| Bounded `cmd/curator` subset | `go test -count=1 -timeout 9m -run 'TestGlobalStatusFailsCheckWhenTheClosureCannotBeProven\|TestCLIExecutionAssuranceSelectionIsPortableDefaultAndVerifiedFailClosed' ./cmd/curator/` | 0 | ok, 1.312s |
| Canonical golden verifier | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | `canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2` / `canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true` |
| Whitespace | `git diff --check` | 0 | clean |
| Board | `task-board --no-update-check validate` | 0 | `Board is valid. No issues found.` |

`golangci-lint has version 2.12.2 built with go1.25.5`.

## Test matrix — internal/swiftpminterop

| Metric | Before rework | After rework |
| --- | ---: | ---: |
| Top-level tests | 58 | **62** |
| Tests including subtests | 96 | **114** |
| Statement coverage | 86.1% | **86.0%** |

The 0.1-point coverage move is the new fail-closed branches in
`publicHeaderRoot`, `confineModuleMapLayout`, and `literalIncludeOperand`, whose
error arms are exercised while a few defensive arms (Windows-absolute and
walk-error paths) are not reachable from a portable fixture.

## Negative controls (proof the new vectors are not vacuous)

Each fix was reverted in the working tree, the corresponding vector was run, and
the file was restored from a backup before continuing.

| Reverted fix | Vector | Exit | Observed |
| --- | --- | ---: | --- |
| include-grammar rejection replaced by the old `continue` | `go test -run TestH09 ./internal/swiftpminterop/` | 1 | 5 subtests FAIL, all `err=<nil>` — the computed include closed successfully |
| `publicHeaderRoot(..., "")` restored and layout guard removed | `go test -run 'TestH10\|TestH11' ./internal/swiftpminterop/` | 1 | 8 subtests FAIL, including both silently-passing escaping layouts |
| condition filter restored in `directTargetDependencies`, selected-only classification | `go test -run 'TestCGP05Conditional\|TestPrunedConditional' ./internal/swiftpminterop/` | 1 | capture digests diverged `sha256:b1c1e85c…` vs `sha256:d907f1c2…`; pruned destination dropped the boundary declaration |

The working tree was verified byte-identical to the fixed state afterwards; the
full focused suite re-ran green (exit 0) following each restore.

## Expected-red gates

None. Every gate listed above is expected green and exited 0. The negative
controls above are deliberately-reverted probes, not gates; their non-zero exits
are the evidence they were run for, and the reverts were undone immediately.

## Not run

`go test ./...` including the full `cmd/curator` package was **not** run: one
`cmd/curator` pass is roughly ten minutes and exceeds this session's bounded
call limit, so the monolithic suite remains the Orchestrator's gate.
`cmd/curator` imports only `closuregraph.ID` (a string type) from the changed
packages and no SwiftPM package at all, so this delta cannot reach it;
`go vet ./...` type-checked its test files and the bounded closure/assurance
subset above ran green.

## Log artifacts

Raw logs are under `.temp/TASK-260811-tkurtl/` in the worktree:
`gofmt-01.log`, `vet-01.log`, `vet-02.log`, `lint-01.log`, `lint-02.log`,
`go-test-nocmd-01.log`, `go-test-nocmd-02.log`, `go-test-race-01.log`,
`go-test-race-02.log`, `go-test-cover-02.log`, `cmd-curator-subset-01.log`,
`golden-verifier-01.log`, `golden-verifier-02.log`.
