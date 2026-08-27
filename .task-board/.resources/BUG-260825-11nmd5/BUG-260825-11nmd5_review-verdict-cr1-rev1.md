# BUG-260825-11nmd5 — review verdict: CHANGES REQUESTED → `to-dev`

Reviewer run over Change Request `CR-BUG-260825-11nmd5-1` revision 1.
Base `903af23a`, candidate tree `93b77020`, branch `task-board/story/STORY-260822-2lvw0e`.
Working tree verified byte-identical to the candidate tree (`git diff --stat 93b7702 -- .` empty)
before and after every probe.

## Verdict

**Changes requested.** Two blocking findings. The second is fatal on its own; the first
means the correct rework is almost certainly "drop this delta", not "patch it".

---

## Finding 1 (blocking) — this delta is superseded work against a stale base

The fix for this bug is **already merged on `origin/main`**, in two commits that both cite
this bug ID:

| Commit | Title | Merged as |
| --- | --- | --- |
| `c9fe49c` | Let a materialized vendor tree keep its generator directives | `680f6a6` (PR 40) |
| `438d557` | Resolve the directive scan by severity, not by first hit | `e027667` (PR 42) |

The CR base is not on that line:

```
$ git merge-base --is-ancestor 903af23 origin/main   → NO
$ git merge-base 903af23 origin/main                 → 1f55f1b
$ git log --oneline 903af23..origin/main | wc -l     → 20
$ git log --oneline origin/main..903af23             → 903af23 (board reconciliation only)
```

The board notes already name this exact delta as dead work. From the RUN-260825-3ef162
developer handoff:

> the STORY-260822-2lvw0e story worktree (HEAD 903af23, PRE-PR-40) still carries UNCOMMITTED
> work from the timed-out run RUN-260825-6e4450 — a divergent second implementation of the
> same carve-out using a Module.Main guard instead of the firstParty guard that actually
> landed in PR 40 [...] It is dead work against a stale base and must NOT be committed or
> integrated.

That is precisely the content of `CR-BUG-260825-11nmd5-1` revision 1: `graph.go` +
`graph_test.go` + `build_test.go` + a `LOGBOOK.md` entry, all built on the `Module.Main`
guard. Landed `main` uses a different guard shape:

```go
// origin/main:internal/godriver/graph.go:201
firstParty := item.Module != nil && item.Module.Replace != nil
// :285
if matched == directiveGenerate && (firstParty || !strictlyBelow(item.Dir, filepath.Join(validation.BuildRoot, "vendor"))) {
```

So even setting Finding 2 aside, accepting this revision would land a second, conflicting
implementation of a carve-out that already exists upstream.

---

## Finding 2 (blocking, proven) — the delta reintroduces a `go_forbidden_compiler_directive` bypass

`scanSourceDirectives` (`internal/godriver/graph.go:286` in the candidate) returns on the
**first read window that matches any needle**. Within one window `//go:cgo_import_dynamic`
wins by needle order; **across** windows it does not. A `//go:generate` in window 1 sets
`matched == 2` and terminates the scan, so a `//go:cgo_import_dynamic` past the 64 KiB
boundary in the same file is never read.

On the base commit that was harmless: `matched == 2` rejected unconditionally. The candidate
makes `matched == 2` an *acceptance* for vendored packages, so the file now rides in clean.

Reproduced end to end through `validatePackageGraph`, three trees, same probe, same fixture:

| Tree | `scanSourceDirectives` | `validatePackageGraph` |
| --- | --- | --- |
| base `903af23` | `2` (generate) | `go_generator_forbidden` — rejected |
| **candidate `93b7702`** | `2` (generate) | **`""` — NO ERROR, admitted** |
| `origin/main` `e027667` | `1` (cgo) | `go_forbidden_compiler_directive` — rejected |

Fixture: one vendored, non-replaced, non-main-module package under `<buildRoot>/vendor`;
`value.go` is 118 123 bytes — `//go:generate stringer -type=Value` on line 2, 2 000 lines of
padding, then `//go:cgo_import_dynamic libc_x x "/usr/lib/libSystem.B.dylib"`.

This is a conformance break, not only a hardening gap. `profiles/manager.md` §2.3 is a
containment predicate:

> Each active non-standard `GoFiles` file is scanned as exact bytes and rejected if it
> **contains** `//go:cgo_import_dynamic`, except for the audited allowlist `golang.org/x/sys`
> and `golang.org/x/sys/*`.

The candidate admits a file that contains those exact bytes and is not on the allowlist. It
also fails the CR's own AC clause "the cgo_import_dynamic allowlist is unchanged" in
substance: the literal allowlist is untouched, but its enforcement became evadable.

Probe source attached as `BUG-260825-11nmd5_cgo-shadow-probe_test.go.txt`; log as
`BUG-260825-11nmd5_cgo-shadow-probe.log`. The probe was added, run, and removed; the
candidate tree was re-verified intact afterwards.

### Why the new tests did not catch it

`graph_test.go` subtest `vendored cgo_import_dynamic` writes a **three-line** `value.go`. Both
directives — in fact only one — land in a single read window, which is exactly the control
case that passes on every version including the broken one. The failing class is
cross-window, and nothing in the delta exercises a file larger than one `Read` chunk. A
positive control proved the branch reachable; it did not bound the class.

`origin/main` already carries the two tests that do bound it, both green here:
`TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows` (5 subtests, incl.
`generate_before_cgo`) and `TestVendoredGeneratorCarveOutDoesNotHideACgoImportDynamic`
(3 subtests, drives `Build()`).

---

## What is right in the delta

Recording this so the rework does not throw away the good parts.

The three-part guard is genuinely pinned. I applied and reverted four mutants myself:

| # | Mutant | Result |
| --- | --- | --- |
| A | `if matched == 2 && !vendoredDependency(...)` → `if false` | **killed** — `build root package` and `first-party package below the vendor tree` both red |
| B | drop the `Module.Main` half, keep the path prefix | **killed** — `first-party package below the vendor tree`: `error = <nil>, want go_generator_forbidden` |
| C | drop the path half (`strictlyBelow(...)` → `true`) | **survived — equivalent mutant, not a coverage gap.** `validateModule` (`graph.go:320`) runs before the directive scan and already rejects any non-main module not strictly below `<buildRoot>/vendor` with `vendor_dependency_missing`, so the path half is unreachable defence-in-depth. Correct to keep; nothing can test it. |
| D | extend `vendoredDependency` to the `matched == 1` branch | **killed** — `vendored cgo_import_dynamic`: `error = <nil>, want go_forbidden_compiler_directive` |

Mutant B is the narrowing one that matters and it is dead, so the first-party/third-party
bound really is what the test asserts, not a path prefix. `TestBuildCompilesThroughAVendoredGeneratorDirective`
correctly drives the real `Build()` entry point rather than the validator alone.

Also note the `Module.Main` guard is **defensible in itself** — arguably stricter than
`profiles/manager.md` §2.3 needs, since the replaced-module sentence there enumerates only
`SFiles` and `cgo_import_dynamic`. It is not the reason for this verdict. It is simply not
the shape that landed.

## Gates I ran myself

All from `.temp/STORY-260822-2lvw0e/worktree` at candidate tree `93b7702`, darwin/arm64:

| Gate | Result |
| --- | --- |
| `go build ./...` | exit 0 |
| `go vet ./internal/godriver/` | exit 0 |
| `gofmt -l internal/godriver/` | exit 0, no output |
| `golangci-lint run ./internal/godriver/...` | exit 0, 0 issues |
| `go test ./internal/godriver/ -count=1` | ok, 29.816s |
| new tests, `-v` | 2 tests / 3 subtests, all PASS |
| 4 mutants | 3 killed, 1 equivalent (table above) |
| cgo-shadow probe, 3 trees | table above |

Not re-run by me, accepted from the implementer's attached evidence: the 34 served + 7
deferred conformance packages, `-race`, `gate-selftest`, `ledger-consistency`,
`no-broad-suppression`. Given Finding 1 those numbers describe a delta that should not land
regardless, so re-running them would not change the verdict. linux/windows lanes: not run
here, no local runner.

---

## Required rework

1. **Drop this delta.** Do not commit `93b7702`. The fix is on `origin/main` at `e027667`.
2. Bring the story branch `task-board/story/STORY-260822-2lvw0e` up to `origin/main` so it
   inherits `c9fe49c` + `438d557`, rather than re-deriving the carve-out from `903af23`.
3. If any part of the story genuinely needs its own change here after the rebase, it must
   keep `main`'s severity-resolving scanner: only `//go:cgo_import_dynamic` may terminate the
   scan; a `//go:generate` hit is recorded and the file is still read to EOF.
4. Any re-added regression fixture must include the **cross-window** case, not only the
   same-window control.

Integration/rebase sequencing is the orchestrator's call, not mine — I did not touch the
branch. `git status` at hand-off is the four CR files and nothing else.
