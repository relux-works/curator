# BUG-260729-1o0m8f — resolve accepted-composite lint findings

Role: developer
Date: 2026-07-29
Status handed to: review

## What was wrong

The accepted Curator composite failed the CI-pinned `golangci-lint v2.12.2`
with four findings and no way to reach the gate without either suppressing
them or changing the source.

## Working tree and base

| Item | Value |
| --- | --- |
| Task worktree | `.temp/BUG-260729-1o0m8f/worktree` |
| Base | byte-exact mirror of `.temp/TASK-260720-1pvfj5/rework/composite` |
| Base commit under the mirror | `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` |
| Scoped sources at base | identical to the accepted `TASK-260720-jrrgw9` candidate |

The three scoped files are byte-identical in the accepted `jrrgw9` candidate
worktree and in the composite, so starting from the composite starts from the
exact accepted candidate for everything this bug touches:

```
a9080f8c4dd0f1f7a4c7375f468541a3a6965937bff507a6714521fcc6087243  internal/protocoljson/ccj.go
e11c9b17cfd3b9a1d859147131940cd492ad57afb520e291de45e822c49336e4  internal/transaction/journal.go
5fa833e8fd6501d428f2b90faf9047d8e6b8288c25db8c311143d852108cec39  internal/godriver/builddriver_positive_conformance_test.go
```

Lint binary: `.temp/TASK-260720-1pvfj5/bin/golangci-lint`, reporting
`golangci-lint has version 2.12.2 built with go1.25.5`, which is the version
pinned in `.github/workflows/ci.yml`. Go toolchain: `go1.25.5 darwin/arm64`.

### Baseline is genuinely red

`golangci-lint run` in the task worktree, exit **1**:

```
internal/protocoljson/ccj.go:211:38: G115: integer overflow conversion rune -> byte (gosec)
internal/protocoljson/ccj.go:212:38: G115: integer overflow conversion rune -> byte (gosec)
internal/transaction/journal.go:398:82: G602: slice index out of range (gosec)
internal/godriver/builddriver_positive_conformance_test.go:178:4: ineffectual assignment to environment (ineffassign)
4 issues:
* gosec: 3
* ineffassign: 1
```

Caching caveat worth knowing: golangci-lint's result cache is keyed on file
content, not on path. The first run in the fresh mirror replayed cached issues
and printed them against the *composite's* paths. `golangci-lint cache clean`
before each measured run fixes that; every run recorded here was preceded by a
cache clean.

## The three fixes

### 1. G115 — `internal/protocoljson/ccj.go`

The control-character branch narrowed a `rune` to a `byte` purely to index a
sixteen-entry hexadecimal table. The narrowing was already unreachable as a
defect — ranging a string never yields a negative rune, and the branch guard is
`character < 0x20` — but gosec cannot see that, and the conversion bought
nothing. Removed it: the rune indexes the table directly.

```go
buffer.WriteByte(hexadecimal[character>>4])
buffer.WriteByte(hexadecimal[character&0x0f])
```

For `0x00 <= character <= 0x1f`, `byte(character)>>4` and `character>>4` are the
same value, as are `byte(character)&0x0f` and `character&0x0f`. The emitted
bytes are unchanged for the whole C0 range; see the equivalence run below.

### 2. G602 — `internal/transaction/journal.go`

`validateRemovalEntries` compared each entry against `entries[index-1]` behind a
short-circuit `index > 0` guard. Safe, but only to a reader. The loop now
carries the preceding path forward in `previousPath`, so it never indexes behind
the cursor at all. The ordering predicate, the error text, and the fail-closed
behaviour are unchanged; the only structural change is where the previous path
comes from.

### 3. ineffassign — `internal/godriver/builddriver_positive_conformance_test.go`

The `fixed environment` subtest opened with:

```go
environment := indispensableEnvironment()
if environment == nil {
    environment = map[string]string{}
}
```

`environment` was never read after the nil check — the subtest asserts against
`values` and `vectors.FixedEnvironment` instead. `indispensableEnvironment()` is
a pure reader (`return nil` on unix, `os.LookupEnv` twice on windows), so the
whole block is dead. Removed all four lines. The function keeps three other
callers, so nothing became unused.

## Semantic evidence

### The refactors are behaviour-preserving

Both new test files were run against the **original accepted sources** with only
the tests swapped in. Exit **0**:

```
ok  github.com/relux-works/curator/internal/protocoljson  0.337s
ok  github.com/relux-works/curator/internal/transaction   0.487s
```

Since the same assertions pass on the pre-fix and post-fix encoder for every
rune in `0x00..0x1f`, the escape bytes are identical for all control
characters — that is a direct check, not an argument.

### The new coverage has teeth

Three mutations of the fixed sources, each run against the new tests:

| Mutation | Exit | First failure |
| --- | --- | --- |
| hexadecimal table upper-cased | 1 | `U+000B encoded as {"s":"\u000B"}, CCJ-1 requires {"s":"\u000b"}` |
| guard narrowed to `< 0x1f` | 1 | `U+001F encoded as {"s":"<raw 0x1f byte>"}, CCJ-1 requires {"s":"\u001f"}` — the round-trip test also fails with `invalid character '\x1f' in string literal` |
| `previousPath` update dropped | 1 | four ordering subtests: `a manifest out of strict unsigned-byte order was accepted` |

### The godriver subtest still runs, and still passes

Run against the accepted candidate conformance root
`.temp/TASK-260729-3nx97g/worktree/conformance/v1`, whose `manifest.json` is
`b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c` — the exact
digest the `jrrgw9` review verdict accepted. Its `build-drivers.json` publishes
28 `fixed_environment` keys, so the subtest executes rather than skipping.

The subtest ledger before and after the removal is identical apart from wall
time: same six subtests, all PASS, exit **0** both ways. The committed
`SPEC_PIN` root publishes no `build-drivers` vector at all, so this test can
only be exercised through the candidate root.

## New tests

- `internal/protocoljson/ccj_test.go`
  - `TestMarshalCanonicalEscapesEveryControlCharacter` — exhaustive over
    `0x00..0x1f`, byte-exact expected output, plus the boundary cases (space,
    `"`, `\`, `0x7f`, a non-ASCII rune).
  - `TestRequireCanonicalAcceptsEveryControlCharacterEscape` — the encoder's
    own output for each control character must be accepted as canonical and
    round-trip.
- `internal/transaction/journal_order_test.go`
  - `TestValidateRemovalEntriesOrdersByUnsignedBytes` — ten cases: ascending,
    duplicate, descending, tail descending, prefix/extension both ways, root
    repeated, and a `~` / `é` pair where unsigned and signed byte order
    disagree.
  - `TestValidateRemovalEntriesStaysFailClosedWithoutARoot` — nil, empty, and
    rootless manifests are rejected, under a `recover` that fails the test if
    validation faults instead.
  - `TestValidateRemovalEntriesRejectsUnorderedTextDefects` — NUL and invalid
    UTF-8 in a path are still refused independently of ordering.

## Verification ledger

Every command was run as a standalone process in
`.temp/BUG-260729-1o0m8f/worktree`; exit codes are the real ones.

| Command | Exit | Log |
| --- | --- | --- |
| `golangci-lint run` (v2.12.2, baseline, expected red) | 1 | `logs/lint-baseline.log` |
| `golangci-lint run` (v2.12.2, after fixes) | 0 | `logs/lint-final.log` |
| `bash .github/ci/no-broad-suppression.sh` | 0 | `logs/no-broad-suppression-01.log` |
| `gofmt -l` on the three packages | 0, no output | `logs/gofmt-01.log` |
| `go vet` on the three packages | 0 | `logs/go-vet-scoped.log` |
| `go build ./...` | 0 | `logs/go-build-01.log` |
| `go test` protocoljson + transaction + godriver | 0 | `logs/focused-packages-01.log` |
| `go test -race` protocoljson + transaction + godriver | 0 | `logs/focused-packages-race-01.log` |
| new tests against the **original** sources | 0 | `logs/equivalence-original-01.log` |
| mutation runs (expected red x3) | 1, 1, 1 | `logs/mutation-*.log` |
| godriver conformance, before / after | 0 / 0 | `logs/godriver-before-01.log`, `logs/godriver-after-01.log` |
| `go test ./...` full module | see `logs/go-test-all-01.log` | `logs/go-test-all-01.log` |

## Scope proof

`diff -rq` between the accepted composite and the task worktree reports exactly
five in-scope entries and nothing else:

```
internal/godriver/builddriver_positive_conformance_test.go   differ
internal/protocoljson/ccj.go                                 differ
internal/protocoljson/ccj_test.go                            new
internal/transaction/journal.go                              differ
internal/transaction/journal_order_test.go                   new
```

`.golangci.yml` and `.github/workflows/ci.yml` compare byte-identical to the
composite. No `nolint` directive appears anywhere in the patch, no exclusion was
added, and no protocol vector, timeout, or unrelated behaviour was touched.

Patch: `BUG-260729-1o0m8f_lint-fix.patch`, sha256
`8a07c0b239548235aea7dfa05fdb1d1cb2926971d4444d3435a9e6f8da368062`.
`git apply --reverse --check` against the task worktree exits 0.
