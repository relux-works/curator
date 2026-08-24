# Round-4 validation logs — TASK-260811-tkurtl

Every command was run directly as a standalone process; no `tee`, no pipe
chain around a gate. Exit codes are the real ones.

| # | Command | Exit | Log (SHA-256) |
| ---: | --- | ---: | --- |
| 1 | `gofmt -l ./internal ./cmd` | 0 (empty output) | — |
| 2 | `go vet ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/` | 0 | — |
| 3 | `go test -count=1 -cover ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/` | 0 | `round4-focused.log` `9daa4c70b50b41425cf19c071aa28dec849a08b4ce08789c00c2f0b358fb8715` |
| 4 | `go test -race -count=1 ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/` | 0 | `round4-race.log` `b2c1ce19164ee01acf77d6aaf40b614b5c57360dc23a252e76db6905c2afa7f7` |
| 5 | `golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/... ./internal/closuregraph/...` (v2.12.2) | 0 (`0 issues.`) | `round4-lint.log` `6f0e3023397c11bb8062fe00ed31743612574318beadb17a8cf2a9f438ae6139` |
| 6 | `ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb` | 0 | `round4-golden.log` `ea023e2f446eae478b14fc2ca91d7c7c20241f3ba9dbd4bdb9dda0fe6345e354` |
| 7 | `go test -count=1 <53 packages>` (suite minus `cmd/curator`) | 0 — 51 ok, 2 without test files | `round4-nocmd.log` `b86d8900c6bc921535197ad274b3a90202c003289697048cbd8eb2d7718933a6` |
| 8 | `git diff --check` | 0 | — |
| 9 | `task-board validate` | 0 — `Board is valid. No issues found.` | — |

Coverage from gate 3: `swiftpminterop` **86.4%**, `swiftpmsource` 80.0%,
`closuregraph` 80.7%, `closureexec` 58.0%.

Test matrix for `internal/swiftpminterop`: 69 top-level, 188 including
subtests.

## Not run, and why

- `cmd/curator`: `go list -deps ./cmd/curator` returns no `swiftpminterop`
  package, so this round's delta cannot reach it. The monolithic full suite is
  the Orchestrator's gate.

## Compiler probes (bounded foreground calls, Apple clang 21.0.0, arm64-apple-darwin25.5.0)

Probe sources under `.temp/TASK-260811-tkurtl/probe/`. Reads were confirmed by
an exact match on the `-H` output line, or by making the target header an
`#error SECRET_WAS_READ` marker — an earlier loose grep matched the echoed
source line and produced a false "READS" for the trigraph table, which is why
the verdict's default/gnu17 trigraph rows are corrected in the outcome.

- `clang [-std=…] -fsyntax-only -H` for the nine trigraph modes and the `??/`
  splice;
- `clang -fsyntax-only -H` for the BOM (leading and mid-file), NUL, U+0085,
  U+00A0, U+1680, U+2000, U+200A, U+2028, U+2029, U+202F, U+205F, U+3000,
  U+200B, a non-ASCII identifier byte, `#`+U+00A0+`include`, comment+U+00A0,
  and BOM/NBSP combined with a trigraph under `-std=c17`;
- `clang++ -x objective-c++` and `-std=c++14/17/20` for the same trigraph and
  BOM forms;
- `clang -x objective-c -fmodules -fimplicit-module-maps -fmodule-map-file=…`
  against a module whose only header is an `#error` marker, for every
  `@import` separator spelling and the C++14 digit-separator form;
- `clang -fsyntax-only -fmodules -fmodule-map-file=… -I include use.c` for the
  out-of-root module-map header and its two-hop transitive chain.
