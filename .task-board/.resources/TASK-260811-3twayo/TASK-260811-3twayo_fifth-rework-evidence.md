# TASK-260811-3twayo fifth rework evidence

Date: 2026-08-22

Rework source: reviewer `RUN-260822-d8c978`.

## Implemented findings

1. `nodesource.ValidateOutputObservations` now independently derives the exact
   `closuregraph.BuildPlan` from the supplied C4 `GraphBundle`, using the
   supplied execution-policy identity, and compares canonical plan identities
   before trusting any declared output. Structurally valid substitutions of the
   zero-output set, a subset output set, action set, or ordering/waves fail with
   `closure_generated_output_drift`. Genuine one-of-two, two-of-two, and
   condition-pruned selections remain accepted.
2. The independent Python protocol corpus is now
   `node-python-protocol-golden-v2`. Go and Python separately enforce exact
   nested `lock`, `artifact`, and `build` fields; derive package nodes,
   dependency edges, capture graph, target nodes, selection bindings, active
   graphs, and diagnostic record envelopes; canonicalize and hash each record;
   and compare the complete canonical outcome. P10 contains two separate
   interpreter/platform/ABI branches with distinct binding and active-graph
   identities plus an explicit cross-target reuse rejection using
   `closure_target_identity_changed`. Both implementations exercise missing and
   unknown nested-field negatives.
3. `README.md` documents the independent Python oracle command and output
   location in the project tooling table.

No forced-fit condition or external blocker was encountered. Previously closed
Node C0 tool binding, mandatory executable SHA, graph-bound output grammar,
generator chaining/cycle, lifecycle/native rejection, canonical root/target
ordering, and manager-profile coverage were preserved.

## Standalone validation

Every command below ran directly as a standalone process. All reported exit
codes are the real process exit codes.

| Gate | Command | Exit | Evidence |
| --- | --- | ---: | --- |
| Focused packages | `go test -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` | 0 | `.temp/TASK-260811-3twayo/focused-01.log` |
| Forged plan negatives | `go test -count=1 ./internal/nodesource -run 'TestOutputObservationsRejectForgedBuildPlans'` | 0 | `.temp/TASK-260811-3twayo/forged-plan-01.log` |
| Race | `go test -race -count=1 ./internal/nodesource` | 0 | `.temp/TASK-260811-3twayo/race-01.log` |
| Vet | `go vet ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` | 0 | `.temp/TASK-260811-3twayo/vet-01.log` |
| Lint | `golangci-lint run ./internal/nodesource/...` | 0 | `.temp/TASK-260811-3twayo/lint-01.log` |
| Repository compile | `go test -run '^$' ./...` | 0 | `.temp/TASK-260811-3twayo/compile-01.log` |
| Repository build | `go build ./...` | 0 | `.temp/TASK-260811-3twayo/build-01.log` |
| Accepted canonical verifier | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb .research/260811_cross-language-closure-graph-and-checkpoints.md` | 0 | `.temp/TASK-260811-3twayo/canonical-verifier-01.log` |
| Independent Python oracle | `python3 internal/nodesource/testdata/python_protocol_golden.py` | 0 | `.temp/TASK-260811-3twayo/python-oracle-03.log` (13 outcomes) |
| Diff check | `git diff --check` | 0 | `.temp/TASK-260811-3twayo/diff-check-01.log` |
| Board validation | `task-board validate` | 0 | `.temp/TASK-260811-3twayo/board-validate-01.log` |
| Full uncached repository suite | `go test -count=1 -timeout=20m ./...` | 0 | `.temp/TASK-260811-3twayo/full-suite-01.log` |

The accepted Ruby oracle reported 53 passing labeled records, capture reuse in
both CGP05 target branches, two explicit target bindings, and all CGP10
references resolved. The full suite included `internal/nodesource` in 5.103s;
the slowest observed package was `internal/rustsource` at 153.676s.

## Result fingerprints

| File | SHA-256 |
| --- | --- |
| `internal/nodesource/nodesource.go` | `6a792afc67bb9bff509eb936e662d15f10a1f00e0f70c595613e22c0b4674194` |
| `internal/nodesource/rework_test.go` | `fef8934d69c52eb31bc60dce3e8a0cb79488e36675a84322c2bf78651000ad30` |
| `internal/nodesource/nodesource_test.go` | `88d9657e5ebb427b386c749bebe08b17b3b9b35979c4d04e3443d497353f52bf` |
| `internal/nodesource/testdata/python_protocol_golden.py` | `2c34e188afd035127cf873a4d97aa5674b7db2e67a3a090404d5bfa172476cdc` |
| `internal/nodesource/testdata/python_protocol_shared_records.json` | `024e8f9191a47c3258d2111812e053f250cfcc158d5eedb51f447b71572a7eaa` |
| `README.md` | `1f85d12f4347af4d1ed9e7a04fd48fb8a5132912b4954adea5af78ead5276e90` |

The shared worktree contains broad pre-existing implementation changes and
board state. This rework preserved them and did not stage or commit anything.
