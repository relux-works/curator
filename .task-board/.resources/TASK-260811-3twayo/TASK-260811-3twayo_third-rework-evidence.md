# TASK-260811-3twayo third rework evidence

Developer run: `RUN-260822-37b4f9`

## Reviewer findings closed

1. `ExecuteMetadataDerivation` now resolves the exact C0-bound tool record and compares node identity, tree fingerprint, executable SHA-256, executable path, recheck rule, execution domain/platform role, host, and target before permit commit or process start. Zero-start substitution tests cover each field and immediate recheck drift.
2. Node, manager, and compiler tool records now require distinct valid tree and executable-content identities. The previous executable-SHA fallback was removed and missing-evidence tests fail closed.
3. Generated output declarations derive a domain-separated identity from complete path, class, grammar, role, and intermediate status before graph closure. Observation validation reconstructs that identity from the graph-bound output record and observed grammar, not from a caller-side declaration copy.
4. Generator declarations use two-pass canonical indexing. Declared generated artifacts can feed later actions independent of declaration order; the shared plan derives stable waves and rejects action cycles.
5. Node conformance now asserts the 17 accepted CGP05 branch record digests, selected and pruned conditions, feature and peer projections, independent runtime and manager binding/plan drift, and N04-N10/N13 behavior. The independent Python oracle decodes, canonicalizes, hashes, and reference-validates shared graph records before exporting P01-P13 outcomes.

## Changed task scope

- `internal/nodesource/nodesource.go`
- `internal/nodesource/nodesource_test.go`
- `internal/nodesource/conformance_test.go`
- `internal/nodesource/testdata/python_protocol_golden.py`
- `internal/nodesource/testdata/python_protocol_shared_records.json`

## Standalone validation evidence

Every command below was run directly as a standalone gate; no gate was piped through `tee`.

| Command | Exit | Evidence |
| --- | ---: | --- |
| `go test -count=1 ./internal/nodesource` | 0 | Focused Node and independent Python protocol tests passed. |
| `go test -race -count=1 ./internal/nodesource` | 0 | Race-enabled Node suite passed. |
| `go test -cover -count=1 ./internal/nodesource` | 0 | 83.3% statement coverage. |
| `go vet ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` | 0 | No findings. |
| `golangci-lint run ./internal/nodesource/...` | 0 | `0 issues.` |
| `go test -run '^$' ./...` | 0 | Repository-wide compile-only test gate passed. |
| `go build ./...` | 0 | Repository build passed. |
| `go test -count=1 ./...` | 0 | Full uncached repository suite passed; `cmd/curator` 455.572s, `internal/nodesource` 4.946s. |
| `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md` | 0 | 53 canonical records passed; both CGP05 target branches and CGP10 observation branches resolved. |
| `python3 internal/nodesource/testdata/python_protocol_golden.py` | 0 | Independent shared-schema oracle emitted all P01-P13 canonical outcomes. |
| `git diff --check` | 0 | No whitespace errors. |
| `task-board validate` | 0 | Board valid with no issues. |

No task-external blocker or forced-fit architecture conflict was encountered.
