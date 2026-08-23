# TASK-260811-3twayo sixth rework evidence

Date: 2026-08-23

Rework source: reviewer `RUN-260822-c1073c`.

## Implemented finding

The independent Python P10 protocol vector now emits two target-scoped
canonical outcomes, one for `cp313/linux/cp313` and one for
`cp314/darwin/cp314`. Both outcomes reference one selection-neutral
`curator-capture-graph-v1`, while their selection-context, selection-binding,
active-graph, and outcome identities are distinct. A separate cross-target
reuse-negative references the exact two binding identities and carries the
shared diagnostic wire shape with `closure_target_identity_changed`.

Go and the independent Python oracle separately construct, decode, validate,
canonicalize, and hash the accepted shared wire shapes:

- `curator-capture-graph-v1` / `closure-capture-graph-v1`;
- `curator-selection-context-v1` / `closure-selection-context-v1`;
- `curator-selection-binding-v1` / `closure-selection-binding-v1`;
- `curator-active-graph-v1` / `closure-active-graph-v1`; and
- the closed diagnostic shape (`code`, `fields`, `subject`).

Each target outcome includes the complete resolvable node/edge record table,
including the binding target node and `targets` edge. Both implementations
reject missing and unknown fields in every shared graph envelope and preserve
the existing strict nested `lock`, `artifact`, and `build` rejection cases.
The corpus schema is versioned as `node-python-protocol-golden-v3` because the
wire outcome semantics changed from one aggregate P10 result to two exact
target outcomes plus a separate reuse-negative.

No production Node behavior changed. The accepted exact C4-to-C5 plan
re-derivation, forged-plan negatives, exact C0 tool binding, mandatory
executable SHA, graph-bound output grammar, generator chaining/cycles,
lifecycle/native rejection, canonical root/target ordering, and manager
coverage remain intact. No forced-fit condition or external blocker was
encountered.

## Standalone validation

Every gate below ran as a direct standalone process. Exit codes are the real
process exit codes.

| Gate | Command | Exit | Evidence |
| --- | --- | ---: | --- |
| Focused protocol | `go test -count=1 ./internal/nodesource -run 'TestIndependentPythonProtocolGoldens|TestPythonProtocolCorpusRejectsMissingAndUnknownNestedFields|TestPythonP10SharedWireRejectsMissingAndUnknownFields'` | 0 | `.temp/TASK-260811-3twayo/protocol-final.log` |
| Focused packages | `go test -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` | 0 | `.temp/TASK-260811-3twayo/focused-final.log` |
| Race | `go test -race -count=1 ./internal/nodesource` | 0 | `.temp/TASK-260811-3twayo/race-final.log` |
| Vet | `go vet ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` | 0 | `.temp/TASK-260811-3twayo/vet-final.log` |
| Lint | `golangci-lint run ./internal/nodesource/...` | 0 | `.temp/TASK-260811-3twayo/lint-final.log` |
| Repository compile | `go test -run '^$' ./...` | 0 | `.temp/TASK-260811-3twayo/compile-final.log` |
| Repository build | `go build ./...` | 0 | `.temp/TASK-260811-3twayo/build-final.log` |
| Independent Python oracle | `python3 internal/nodesource/testdata/python_protocol_golden.py` | 0 | `.temp/TASK-260811-3twayo/python-oracle-final-standalone.log` (13 semantic vectors; P10 contains two target outcomes and one reuse-negative) |
| Accepted Ruby verifier | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md` | 0 | `.temp/TASK-260811-3twayo/canonical-verifier-final.log` |
| Diff check | `git diff --check` | 0 | `.temp/TASK-260811-3twayo/diff-check-final.log` |
| Board validation before evidence attachment | `task-board validate` | 0 | `.temp/TASK-260811-3twayo/board-validate-final-pre.log` |
| Full uncached repository suite | `go test -count=1 -timeout=20m ./...` | 0 | `.temp/TASK-260811-3twayo/full-suite-final.log` |

The accepted Ruby oracle reported 53 passing labeled records, one reused
CGP05 capture across two explicit target bindings, and all CGP10 references
resolved. The full suite included `internal/nodesource` in 5.165s and
`internal/rustsource` in 156.174s.

## Result fingerprints

| File | SHA-256 |
| --- | --- |
| `internal/nodesource/nodesource_test.go` | `2090c9858378472f5403ca3712e7542c94cc374ed5292e9bf20fa778971ff40d` |
| `internal/nodesource/python_protocol_shared_test.go` | `e32f76d1ab7e983579c6a26c206bd47c8ab83166e3b18cc876e4c12cfe4c350d` |
| `internal/nodesource/testdata/python_protocol_golden.py` | `ea7cdaef6c0c6f45ec89f41c4e87c8be8563005460653ab89579689d2c71a201` |
| `internal/nodesource/testdata/python_protocol_shared_records.json` | `5b16f4ec51fffc25ec671c53eea131e9c7c515e679c559590802cad8feeacf7f` |

The shared worktree contains broad pre-existing implementation and board
changes. This rework preserved them and did not stage or commit files.
