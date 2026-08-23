# TASK-260720-2qqq0w — independent rework test evidence

Date: 2026-07-29
Role: tester
Candidate: `.temp/TASK-260720-2qqq0w/worktree`

## Verdict

The reviewer-requested rework is ready for review.

The guide now documents the reachable `go_embed_input_escape` failure and its
authoring prerequisite. The documentation completeness test now discovers
driver boundary codes from Go syntax trees, including codes routed through
local variables, and fails closed when a diagnostic argument cannot be reduced
to a literal or explicitly identified relay.

## Delta audit

- The task candidate differs from the accepted integrated
  `.temp/TASK-260729-2kaopg/worktree` only in documentation ownership:
  `README.md`, new `docs/compiled-builds.md`, and new
  `cmd/curator/docs_test.go`. The recursive diagnostic comparison returned
  exit 1 because these expected entries differ and also reported the two
  pre-existing `skill-go-testing-tools` symlink loops.
- `diff -q internal/godriver/graph.go
  .temp/TASK-260729-2kaopg/worktree/internal/godriver/graph.go` returned exit 0.
  Product code for the reachable failure was not changed by the rework.
- The new table row links `go_embed_input_escape` to the existing Go source
  prerequisites section, whose authoring rule requires embedded inputs to be
  canonical regular files contained by the build root.
- No candidate product or documentation file was edited during this tester run.
  The regression mutation was performed only in the isolated scratch copy
  `.temp/TASK-260720-2qqq0w/mutation-worktree`.

## Standalone validation commands

| Command | Exit | Evidence |
|---|---:|---|
| `go test ./cmd/curator/ -run '^TestDocumentedBoundaryCodesAreComplete$' -count=1` | 0 | AST-derived emitted-code set and failure table match |
| `go test ./internal/godriver/ -run 'TestPackageGraphValidatesStandardMainModuleAndVendoredTransitiveInputs/transitive_embed_escape$' -count=1 -v` | 0 | reachable transitive embed escape emits the documented refusal |
| `go test ./cmd/curator/ -run '^(TestDocumentedLinksResolve\|TestDocumentedCompiledBuildAuthoringContract)$' -count=1` | 0 | local links resolve and the embed/security/operations authoring contract is present |
| `go build ./...` | 0 | all packages compile |
| `go vet ./...` | 0 | static analysis passes |
| `gofmt -l cmd internal` | 0 | empty output |
| `git diff --check` | 0 | no whitespace errors |

## Expected-red mutation gate

In the isolated scratch copy, the `Embedded inputs` table row was removed while
the implementation and prose prerequisite were left intact.

`go test ./cmd/curator/ -run '^TestDocumentedBoundaryCodesAreComplete$' -count=1`
failed with exit 1, as expected:

`docs/compiled-builds.md omits go-v1 boundary codes: go_embed_input_escape`

This is the direct regression proof for the prior false-green scanner. The
non-zero result is an expected failure, not a passing gate.

## Scope discipline

Per the rework test directive, this run did not start the full repository test
suite, the full `cmd/curator` package test suite, lint, cache clearing, a host
install, staging, committing, publishing, or protocol pinning. The earlier
rework outcome records its own lint and broader focused-package evidence; this
artifact reports only commands run directly by this tester.
