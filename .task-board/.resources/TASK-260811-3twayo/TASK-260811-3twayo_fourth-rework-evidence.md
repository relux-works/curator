# TASK-260811-3twayo fourth rework evidence

## Scope

Closed the three remaining findings from reviewer run `RUN-260822-755ae4`.

1. `CaptureInput.RootKeys` and `RuntimeBinding.TargetNodeIDs` are sorted and
   uniquely validated before index-derived edge keys are assigned. Generator
   tool bindings are emitted only for exact selected owners, so one-product
   selection does not bind inactive actions.
2. Node output reconciliation now requires a validated `GraphBundle` and its
   matching `BuildPlan`, and accepts only `BuildPlan.DeclaredOutputNodeIDs`.
   Missing inactive outputs pass; observed inactive or condition-pruned outputs
   reject.
3. The independent Python P01-P13 corpus now contains typed fixture inputs and
   exact canonical outcome records. Python and Go separately validate schemas,
   derive package/target graph IDs, apply diagnostic precedence, derive the
   outcome record, and compare exact IDs. No Python product code or mutable
   runtime state is shared with Go.

## Tests added or updated

- All four manager profiles: two-root and multi-target permutation invariance
  for capture, binding, active graph, and build plan identities.
- Duplicate root and target rejection.
- One-of-two and two-of-two selected product output reconciliation.
- Condition-pruned output absence and inactive observation rejection.
- Independent schema-aware P01-P13 Python/Go canonical outcome comparison.

## Standalone validation evidence

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/nodesource` | 0 | focused test output observed directly |
| `go test -race -count=1 ./internal/nodesource` | 0 | `.temp/TASK-260811-3twayo/race-01.log` |
| `go vet ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` | 0 | `.temp/TASK-260811-3twayo/vet-01.log` |
| `golangci-lint run ./internal/nodesource/...` | 0 | `.temp/TASK-260811-3twayo/lint-01.log` |
| `go test -run '^$' ./...` | 0 | `.temp/TASK-260811-3twayo/compile-01.log` |
| `go build ./...` | 0 | `.temp/TASK-260811-3twayo/build-01.log` |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md` | 0 | `.temp/TASK-260811-3twayo/canonical-01.log` |
| `python3 internal/nodesource/testdata/python_protocol_golden.py` | 0 | `.temp/TASK-260811-3twayo/python-oracle-01.log` |
| `git diff --check` | 0 | `.temp/TASK-260811-3twayo/diff-check-01.log` |
| `task-board validate` | 0 | `.temp/TASK-260811-3twayo/board-validate-01.log` |
| `go test -count=1 ./...` | 1 | `.temp/TASK-260811-3twayo/full-test-01.log`; expected harness failure rationale: `cmd/curator` exceeded Go's default 10-minute timeout, while every other package passed |
| `go test -count=1 -timeout=20m ./...` | 0 | `.temp/TASK-260811-3twayo/full-test-02.log`; authoritative uncached repository retry with adequate harness timeout |

No files were staged or committed.
