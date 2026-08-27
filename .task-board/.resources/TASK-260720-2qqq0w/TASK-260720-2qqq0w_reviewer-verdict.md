# Reviewer verdict: changes requested

Route: `to-dev`.

## Material finding

`docs/compiled-builds.md` says its go-v1 boundary-code table is the complete set (lines 517-524), but the Package graph row omits the reachable `go_embed_input_escape` diagnostic. `internal/godriver/graph.go` lines 244-250 emits that code for an escaped or invalid non-standard-library embed input. Existing implementation tests and candidate conformance mapping also name it: `internal/godriver/build_test.go` line 129, `internal/godriver/graph_test.go` line 51, and `internal/godriver/build_conformance_test.go` line 511. This violates the acceptance criterion that an author can understand every failure class without reading source.

The new `TestDocumentedBoundaryCodesAreComplete` is false-green. Its regex in `cmd/curator/docs_test.go` line 35 discovers only a string literal passed directly to `diagnostic` or `diagnosticErr`, or an exported `Code...` constant. The omitted code is assigned to a local variable and then passed to `diagnosticErr`, so it is invisible to the scanner.

## Reproduction

- `rg -n go_embed_input_escape docs/compiled-builds.md cmd/curator/docs_test.go` returns no match.
- `go test ./cmd/curator/ -run ^TestDocumentedBoundaryCodesAreComplete$ -count=1` passes, demonstrating the false negative.
- `go test ./internal/godriver/ -run TestPackageGraphValidatesStandardMainModuleAndVendoredTransitiveInputs/transitive_embed_escape -count=1 -v` passes and exercises the reachable code.

## Required rework

Document `go_embed_input_escape` in the prerequisite/failure guidance and make the documentation contract test cover this emission path so the claimed complete-set check cannot pass with the code missing. Re-run the focused documentation/parser tests plus build, vet, formatting, and diff checks before another reviewer cycle.

## Other evidence

The mixed schema-6 example loads through the real parser; all documented JSON parses; local links and five external protocol links resolve; toolchain precedence, build-root exclusion, vendor-only operation, trust boundary, cache/currentness, dry-run, repair, locked GC, shims, unsupported features, verification commands, and schema 1-5 compatibility are otherwise explicit. `go test ./cmd/curator/ -run ^TestDocumented -count=1`, the focused schema-6 parser tests, and `git diff --check` passed during review.