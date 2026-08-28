# TASK-260811-3twayo Node runtime/build contract rework evidence

Rework target: reviewer RUN-260822-615b9e.

## Implemented

- Exact C0 checkpoint authority for the selected Node and manager executable records, including full toolchain and executable content identities.
- A single executable C1-C4 metadata seam backed by `closureexec` permit commit, immediate tool/input rechecks, execution, and executor-issued derivation receipts.
- Canonical manager-independent package, source-set, workspace, peer, conditional, and shipped-generated-text capture records.
- Canonical dependency declaration sorting and duplicate rejection before edge-key assignment.
- Declared TypeScript/generator actions and immutable expected outputs added to capture before C4; exact compiler tools and host/target edges remain binding-only; C5 derives the plan without graph growth.
- Exact output path/class/grammar/content reconciliation through `ProducedArtifactObservation` and `PublicationEvidence`, with duplicate declaration and invalid digest rejection.
- Lifecycle/native/compiled rejection, workspace containment, fresh manager-state regeneration, and derived-state receipt contracts.
- A checked-in independent Python P01-P13 JSONL corpus generated without importing Go code or state.

## Named conformance

The Node suite includes CGP05, CGP10-CGP11, CGN09, CGN15-CGN18, and N04-N10/N13 coverage. Negative permit, runtime drift, undeclared generator, native-build, and compiled-byte cases assert zero protected process starts where applicable.

## Validation evidence

All commands ran directly as standalone gates.

- `go test -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — exit 0.
- `go test -race -count=1 ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — exit 0; artifact policy completed in 253.909s.
- `go vet ./internal/nodesource ./internal/closuregraph ./internal/closureexec ./internal/artifactpolicy` — exit 0.
- First `golangci-lint run ./internal/nodesource/...` — exit 1 for six missing exported-symbol comments. Comments were added.
- Second focused `golangci-lint run ./internal/nodesource/...` — exit 0, `0 issues`.
- Repository `golangci-lint run` — exit 0, `0 issues`.
- `go build ./...` — exit 0.
- `go test -count=1 ./...` — exit 0; `cmd/curator` 436.718s, `internal/nodesource` 4.638s.
- Ruby canonical verifier — exit 0; 53 labeled records, two CGP05 target branches, two CGP10 observation branches, all references resolved.
- `git diff --check` — exit 0.

No unrelated worktree changes were modified.
