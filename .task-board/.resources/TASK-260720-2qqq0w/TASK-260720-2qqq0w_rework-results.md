# TASK-260720-2qqq0w — rework after the 2026-07-29 reviewer verdict

Scope of this cycle: the two defects named in
`TASK-260720-2qqq0w_reviewer-verdict.md`. No other documentation content was
changed, and no product code was touched.

## Defect 1 — the failure table was not the complete set

`docs/compiled-builds.md` claimed its `go-v1` boundary-code table was complete
but omitted `go_embed_input_escape`, which `internal/godriver/graph.go:246-250`
raises for an escaped or invalid embed input in a non-standard-library package.

Fixed in two places:

- **New failure-table row** (`## Failure classes`): a dedicated `Embedded
  inputs` group rather than an extra token in the `Package graph` row, because
  that row's remedy is about `source_dir` and does not apply to an embed
  violation.
- **New authoring prerequisite** (`## Go source prerequisites`, "Embedded inputs
  must be regular files inside the build root."): states the actual rule
  `validateRegularInput` enforces — regular file, already-canonical path,
  contained by the build root, no symlink or special file — and notes that the
  same violation inside a standard-library package surfaces as
  `go_standard_input_escape` instead, because there it means the toolchain is
  not intact.

## Defect 2 — `TestDocumentedBoundaryCodesAreComplete` was false-green

The old scanner was a regex (`codeRE`) that only matched a string literal passed
directly to `diagnostic`/`diagnosticErr`, or an exported `Code…` constant. The
omitted code is assigned to a local variable first, so the scanner never saw it
and the completeness assertion passed while the table was incomplete.

`cmd/curator/docs_test.go` now resolves codes from the driver package's AST
(`go/parser` + `go/ast`) instead of its text:

- Every non-test `.go` file is parsed regardless of build constraints, so
  Windows-only and Darwin-only codes are still in scope on any host.
- A diagnostic's first argument is resolved as a string literal, a package-level
  constant, or a local variable. Names are tracked as `map[string][]string`, so
  a variable reassigned on a branch (`packagePathError`, and the embed site
  itself) contributes *every* value it can hold, and a constant declared once
  per build-constrained file contributes each of its values.
- An argument that resolves to no literal is a **test failure**, not a silent
  skip. The one deliberate exception is registered explicitly in
  `relayedCodeArguments`: `message.Failure.Code` in `workerclient.go`, where the
  manager relays a code the worker raised from this same package.
- Per file, the number of diagnostics reached through function bodies is
  compared against the number present anywhere in the file, so a diagnostic
  raised outside a function body cannot escape the scan.

The scan now discovers 73 codes; the previous regex discovered 72.

## Mutation evidence

The reviewer's objection was that the gate could pass while the claim was false,
so the rework is verified by mutation rather than by a green run alone. Each
mutation was applied to the shipped file, the focused test run, then the file
restored and confirmed identical with `diff -q`.

| Mutation | Result |
|---|---|
| Delete the `Embedded inputs` row from the guide | **FAIL** — `docs/compiled-builds.md omits go-v1 boundary codes: go_embed_input_escape` |
| Add `go_totally_invented_code` to the table | **FAIL** — `documents unknown go-v1 boundary codes: go_totally_invented_code` |
| Route the embed code through `strings.ToLower(code)` in `graph.go` | **FAIL** — `graph.go:250:11: boundary code strings.ToLower(code) resolves to no literal` |
| Raise a diagnostic at package level in `graph.go` | **FAIL** — `1 diagnostic calls sit outside a function body and were not scanned` |

Mutation 1 is the direct regression test for the reviewer's finding: under the
old regex scanner that mutation passed. Note that mutation 1 leaves the prose
mention of `go_embed_input_escape` in place and the test still fails, which
confirms the check is scoped to the table that makes the completeness claim.

An earlier attempt at mutation 3 (`item.EmbedCode`) did not compile and
therefore proved nothing; it was replaced with the compilable
`strings.ToLower(code)` form above, which builds (`go build ./internal/godriver/`
exit 0) and exercises the resolver.

## Gate evidence

Every command run standalone from the candidate worktree
`.temp/TASK-260720-2qqq0w/worktree`; real exit codes, no pipes.

| Gate | Exit |
|---|---|
| `gofmt -l cmd internal` | 0 (empty output) |
| `go build ./...` | 0 |
| `go vet ./...` | 0 |
| `~/go/bin/golangci-lint run` | 0 (`0 issues.`) |
| `git diff --check` | 0 |
| `go test ./cmd/curator/ -run '^TestDocumented' -count=1` | 0 (8/8) |
| `go test ./internal/skillspec/ -count=1` | 0 |
| `go test ./internal/godriver/ -run 'TestPackageGraphValidatesStandardMainModuleAndVendoredTransitiveInputs' -count=1` | 0 |
| `go test ./cmd/curator/ -count=1` | 0 (502s) |

`golangci-lint` is not on `PATH`; it is invoked from `~/go/bin`.

The full repository suite was not re-run for this delta, per the standing
direction after the cancelled `RUN-260729-12a232`. The delta is documentation
plus one documentation test file, and the packages it reads
(`internal/godriver`, `internal/skillspec`, `internal/install`) were exercised
by the focused runs above.

## Delta containment

`diff -rq --exclude=.git --exclude=vendor .temp/TASK-260729-2kaopg/worktree
.temp/TASK-260720-2qqq0w/worktree` reports exactly three entries (excluding the
two `skill-go-testing-tools` symlink loops):

- `README.md` differs
- `cmd/curator/docs_test.go` is new
- `docs/compiled-builds.md` is new

`internal/godriver/graph.go` is byte-identical to the accepted worktree after
the mutation runs. No accepted worktree was mutated. Nothing was staged,
committed, published, or pinned, and no rc.4/rc.5 release is claimed.

## Carried forward unchanged

The pre-existing finding still stands and is not addressed here: the committed
CI pin `00b1688a9b2457ca397a0bb550acf47cad8ee967` (`1.0.0-rc.3`) carries no
`conformance/v1/vectors/build-drivers.json`, so the build-driver conformance
tests fail against it. Reconciling that pin is `TASK-260720-1pvfj5`.
