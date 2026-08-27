# BUG-260729-r0fe02 — stabilize godriver cancellation race contract

Baseline: the exact accepted `TASK-260720-jrrgw9` godriver bytes.
Only two files differ from that baseline: `internal/godriver/fingerprint.go` and
`internal/godriver/fingerprint_equivalence_test.go`.
Worktree: `.temp/BUG-260729-r0fe02/worktree`.

## Root cause

`digestToolchainRecords` re-opens each file record and copies it through
`copyWithContext`. That copy loop checks `ctx.Err()` at the top of every chunk
and returns `(bytesCopiedSoFar, ctx.Err())`. The caller then collapsed *every*
non-clean copy into one bucket:

```go
if copyErr != nil || closeErr != nil || written != opened.Size() {
    ... diagnosticErr("toolchain_mutated", ...)
}
```

So a cancellation landing inside a record's bytes was reported as
`toolchain_mutated`, while a cancellation landing anywhere else in the same
pipeline (walk descent, per-entry check, per-record check) was reported as
`toolchain_timeout`. Which one a racing cancellation produced was pure
scheduling, which is what the mandatory race gate intermittently observed:

```
fingerprint_equivalence_test.go:519: error = go-v1 toolchain_mutated:
    toolchain file "z" changed while reading, want nil or toolchain_timeout
```

## Precedence decision

**Only the abort this package itself raises inside its own copy loop counts as
cancellation. Everything else stays `toolchain_mutated`.**

`copyWithContext` now tags its own context check with a package-private sentinel
(`errCopyAbandoned`) wrapping the real cause, and `digestCopyDiagnostic`
classifies on that sentinel alone:

- copy carries `errCopyAbandoned` → `toolchain_timeout`
- read error, close error, or a length that moved on its own → `toolchain_mutated`
- a bare `context.Canceled` that did *not* come from that loop → `toolchain_mutated`

This is a reclassification between two refusals, not a relaxation. Both codes
fail closed: no fingerprint is produced and no toolchain is trusted in either
case. A concurrent write cannot be laundered into a deadline, because the
sentinel is only ever attached by this package's own check — that case is pinned
by `TestDigestCopyDiagnosticPrecedence/bare_context_error_is_not_this_loop's_abort`.
The precise cause stays reachable: the returned error still unwraps to
`context.Canceled` / `context.DeadlineExceeded`.

No new diagnostic value is introduced. `toolchain_timeout` was already produced
by the walk and per-record checks, so every existing consumer already handled
it. `cmd/curator/builds.go:768` (operator guidance) is the only place outside
godriver that branches on these codes; `toolchain_timeout` already fell into its
`default` branch, so a cancelled fingerprint now gets the same guidance as every
other cancelled fingerprint instead of the "could not be resolved or verified"
text. That is more consistent, not weaker.

## Test design — and a correction to the first approach

The first attempt at a gate repeated the racing subtest 200× per run. **That was
measured and it does not work.** With the fix reverted, a racing repetition was
run under `-race` and stayed green at:

- 10,000 racing attempts (`-count=50`) — exit 0
- 100,000 racing attempts (`-count=500`) — exit 0
- 2,000 racing attempts with a freshly-written 256 KiB fixture per attempt — exit 0

The original failure was found while this run's focused race **overlapped
TASK-260720-1pvfj5's final race gate**. Under that CPU contention, goroutine
scheduling latency widens the window enough to sample it; on an unloaded machine
`go cancel()` almost always lands before the first file is read. A racing
repetition is therefore a load-dependent smoke check and cannot be the proof.
The loop is kept (it is ~5 ms and costs nothing) but its comment now says
plainly what it does not prove.

The actual gate is deterministic. `countdownContext` returns nil for a fixed
number of `Err()` calls and cancellation for every call after that, which places
a cancellation on an exact check — including a check taken between two chunks of
one file. On top of the named boundary cases, the contract is proved by an
exhaustive sweep, `every cancellation point stays a deadline`: it walks the
budget from 0 upward until a budget completes the fingerprint cleanly, asserting
at every step that the outcome is `nil` or `toolchain_timeout` that unwraps to
`context.Canceled` — never `toolchain_mutated`. It covers both phases and is
self-extending: a cancellation check added anywhere later is covered the moment
it exists.

## Negative controls (fix reverted, exact bytes otherwise)

| Control | Exit | Result |
| --- | --- | --- |
| `-run 'TestFingerprintCancellationStaysFailClosed/every_cancellation_point_stays_a_deadline' -count=1` | **1** (expected red) | fails at budget 17 in 0.47 s: `toolchain_mutated: toolchain file "a/b/c/d/leaf" changed while reading` |
| `-run 'TestFingerprintCancellationStaysFailClosed\|TestDigestCopyDiagnosticPrecedence' -count=1` | **1** (expected red) | 3 deterministic cancellation subtests + 3 precedence rows red |
| `-race -run '.../cancelled_between_the_walk_and_the_digest' -count=50` | 0 | **did not detect** — racing repetition is not a gate |
| `-race -run '.../cancelled_between_the_walk_and_the_digest' -count=500` | 0 | **did not detect** — 100k attempts, still green with the bug present |

`internal/godriver/fingerprint.go` was restored byte-identical after the
controls (sha256 `d53cb4f4212777496da7cbf80826366f5274406c0c4bb083b1c0800b9a822d23`
before and after).

## Gates on the final bytes

All run as standalone processes, no pipes; exit codes are the real ones.

| Gate | Command | Exit |
| --- | --- | --- |
| Format | `gofmt -l` on both changed files | 0 (no output) |
| Build/vet | `go vet ./internal/godriver/` | 0 |
| Focused non-race | `go test ./internal/godriver/ -run 'TestFingerprint\|TestToolchain\|TestDigestCopy\|TestGoVersion\|TestCandidateRC4' -count=1 -v` | 0 (3.06 s) |
| Deterministic cancellation | `go test ./internal/godriver/ -run 'TestFingerprintCancellationStaysFailClosed\|TestDigestCopyDiagnosticPrecedence' -count=1 -v` | 0 — 19 subtests pass |
| Bounded repeated race | `go test -race ./internal/godriver/ -run 'TestFingerprintCancellationStaysFailClosed\|TestDigestCopyDiagnosticPrecedence' -count=1000` | 0 (28.0 s) |
| Focused family under race | `go test -race ./internal/godriver/ -run 'TestFingerprint\|TestToolchain\|TestDigestCopy\|TestGoVersion\|TestCandidateRC4' -count=5` | 0 (18.1 s) |

The `-count=1000` race gate is 200,000 racing attempts plus 1,000 exhaustive
deterministic sweeps plus 1,000 runs of the precedence table, with no race
detector reports. The original repro shape (`-count=5000`, one racing attempt
per run) produced 3 failures; this gate carries 40× that racing volume and adds
the deterministic proof the racing volume cannot give.

`TestFingerprintRejectsInvalidUnicodePath` skips on this host — APFS rejects the
invalid-UTF-8 filename the fixture needs (`illegal byte sequence`). Pre-existing,
platform-driven, unrelated to this change.

## Lint — one pre-existing finding, NOT from this patch

`golangci-lint` was not installed; it was run at the version CI pins
(`.github/workflows/ci.yml` → `v2.12.2`) via
`go run github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2`, against
the repo's shipped `.golangci.yml`, scoped to the package:

```
golangci-lint run ./internal/godriver/...      exit 1
internal/godriver/builddriver_positive_conformance_test.go:178:4:
    ineffectual assignment to environment (ineffassign)
1 issues: * ineffassign: 1
```

**The checklist item "Lint clean" is left unchecked because that command exits 1.**

The two files this patch changes produce **zero** findings. The one finding is in
`builddriver_positive_conformance_test.go`, which is byte-identical to the
accepted `TASK-260720-jrrgw9` bytes. Running the same command against that
accepted baseline worktree produces byte-identical output and the same exit 1
(`lint-godriver-accepted-baseline.log`), which proves the finding is pre-existing.

**This needs an owner outside this bug.** `ineffassign` is on by default in
golangci-lint v2, so the composite's CI lint job is red on this today,
independently of this patch. It belongs to whoever owns
`builddriver_positive_conformance_test.go`, not to this task's scope
(godriver fingerprint cancellation only).

## Not run, and why

Per the resume instruction, `./...` and the Curator full/race suite were **not**
run — that gate belongs to TASK-260720-1pvfj5 and must not overlap another heavy
race run. `make lint` was not run as such, because it is repo-wide; the
package-scoped equivalent above was run instead, at the CI-pinned version.

## Handoff

`BUG-260729-r0fe02_patch.diff` applies cleanly to the composite:
`git -C .temp/TASK-260720-1pvfj5/rework/composite apply --check -p1 <patch>` → exit 0.
The composite's godriver baseline was verified byte-identical to the accepted
jrrgw9 bytes before that check. Applying it to the composite is
TASK-260720-1pvfj5's step, after this patch is reviewed.
