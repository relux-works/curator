# Round-3 validation logs — TASK-260811-tkurtl

Every command below was run as a standalone process; no pipe chain carried a
gate's status. The exit code recorded is the real `$?` of that process.

## gofmt / build / vet

```
$ gofmt -l internal/ cmd/
GOFMT_EXIT:0            (no output)
$ go build ./...
BUILD_EXIT:0
$ go vet ./...
VET_EXIT:0
```

## Focused packages with coverage

```
$ go test -count=1 -cover ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/ ./internal/closureexec/
FOCUSED_EXIT:0
ok  github.com/relux-works/curator/internal/swiftpminterop   3.639s  coverage: 86.6% of statements
ok  github.com/relux-works/curator/internal/swiftpmsource    8.470s  coverage: 80.0% of statements
ok  github.com/relux-works/curator/internal/closuregraph    10.211s  coverage: 80.7% of statements
ok  github.com/relux-works/curator/internal/closureexec      3.713s  coverage: 58.0% of statements
```

## Race

```
$ go test -race -count=1 ./internal/swiftpminterop/ ./internal/swiftpmsource/ ./internal/closuregraph/
RACE_EXIT:0
ok  github.com/relux-works/curator/internal/swiftpminterop   25.235s
ok  github.com/relux-works/curator/internal/swiftpmsource    13.642s
ok  github.com/relux-works/curator/internal/closuregraph    106.023s
```

## Lint (pinned)

```
$ golangci-lint --version
golangci-lint has version 2.12.2 built with go1.25.5
$ golangci-lint run ./internal/swiftpminterop/... ./internal/swiftpmsource/... ./internal/closuregraph/...
LINT_EXIT:0
0 issues.
```

## Canonical golden verifier

```
$ ruby .research/260811_cross-language-closure-canonical-golden-verifier.rb
GOLDEN_EXIT:0
canonical_goldens=pass labeled_records=53 cgp05_target_branches=2 cgp10_observation_branches=2
canonical_references=pass cgp05_capture_reused=true explicit_target_bindings=2 cgp10_all_refs_resolve=true
```

`internal/closuregraph/testdata/canonical-goldens.txt` is unmodified in this
round (`git status --short` reports it untouched).

## Suite minus cmd/curator, one bounded call

```
$ go test -timeout 9m -count=1 $(go list ./... | grep -v cmd/curator)
NOCMD_EXIT:0
ok lines: 51
FAIL lines: 0
```

## Interop test matrix

```
$ go test -count=1 -v ./internal/swiftpminterop/
VERBOSE_EXIT:0
top-level PASS: 66
all PASS incl subtests: 138
FAIL lines: 0
```

New this round, all PASS:

```
--- PASS: TestH12TransitiveIncludeClosureIsScanned                       (5 subtests)
--- PASS: TestH13DirectiveRecognitionMatchesTheCompilerTranslation      (15 subtests)
--- PASS: TestS05ConditionalCxxInteropOptInIsSelectionNeutral
--- PASS: TestS05ConditionalCxxInteropOptInStillRequiresAnAcceptedProfile
```

## Whitespace and board

```
$ git diff --check
DIFFCHECK_EXIT:0
$ task-board --no-update-check validate
Board is valid. No issues found.
```

## Compiler probes behind finding B

Run against the pinned Apple Clang on this host, each file including
`secret.h`; the count is `clang -std=c17 -fsyntax-only -H <file> | grep -c
secret.h`:

```
b1 (#inc\<newline>lude "secret.h"): 1
b2 (/* */ #include "secret.h"):     1
b3 (\f#include "secret.h"):         1
b4 (%:include "secret.h"):          1
```

Two further probes, exit 0 with the header reported in both cases:

```
p1  #include /*<newline>*/ "secret.h"     -> clang reads secret.h
p2  int a;<newline>/*<newline>*/ #include "secret.h" -> clang reads secret.h
```

## Explicitly not run

`cmd/curator`. The headless single-call cap is 10 minutes and that package
alone runs ~8 minutes; the monolithic full suite is the Orchestrator's gate.
`go list -deps ./cmd/curator | grep -c swiftpminterop` returns **0**, and this
round changed no file outside `internal/swiftpminterop/`, so no delta from this
round is reachable from that package.
