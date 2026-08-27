# BUG-260825-11nmd5 — review verdict: CHANGES REQUESTED (-> to-dev)

Scope reviewed: the PR 40 delta (commit `c9fe49c`, merged as `680f6a6`) —
`internal/godriver/graph.go` + `internal/godriver/moduleroots_test.go`.
Spec pin: `relux-works/curator-spec@0ed5c691e9208eea52f21db2fc05e226ce3516fd`
(read from a local checkout already at that revision).

Review worktree: `.temp/BUG-260825-11nmd5/review-wt` (detached at `origin/main`
= `680f6a6`), removed after the run. The story worktree
`.temp/STORY-260822-2lvw0e/worktree` sits at `903af23`, which predates the fix;
nothing was changed there.

## Verdict

The bug is real, the fix direction is right, and the regression tests are
honest. **But the delta introduces a bypass of the `//go:cgo_import_dynamic`
gate**, proven end to end through the real `Build` entry point. That is a
security-relevant regression in the same scan the AC says must stay intact, so
this goes back to `to-dev` for one follow-up commit rather than being accepted.

## What is correct

- **The carve-out shape matches decision 0005 and §2.3.** `profiles/manager.md`
  §2.3 at the pin: "`//go:generate` in `GoFiles` is inert — managers MUST NOT
  run generators and `go build -mod=vendor` does not execute them; its presence
  in vendored `GoFiles` (vendor already materialized) does not fail preflight."
  Decision 0005 §"Broad `if false` ... rejected — expands trust boundary beyond
  the vendored, audited cases" is exactly what the bounded condition avoids.
  The guard mirrors the pure-Go-assembly carve-out one block above it, which is
  the right internal precedent.
- **The motivating case is real.** The Go compiler treats `//go:generate` as an
  ordinary comment; nothing in `go list -mod=vendor` / `go build -mod=vendor`
  executes it. `clipperhouse/displaywidth` `gen.go` carries a bare one and no
  released version drops it.
- **All three halves of the guard are pinned by a test that reddens when the
  half is removed.** I ran the mutations myself in the review worktree:

  | Mutation applied to `graph.go:275` | Test that turns red |
  | --- | --- |
  | `if matched == 2 {` (carve-out reverted) | `TestAuditedVendorAllowancesAreWithheldFromAReplacedModule///go:generate_in_a_vendored_package/allowed_for_an_audited_third-party_module` |
  | `if matched == 2 && !strictlyBelow(...)` (`firstParty` half dropped) | `.../withheld_from_a_replaced_module` |
  | `if matched == 2 && firstParty` (path half dropped) | `TestBuildRejectsDirectivesPGOAndMultipleRootsBeforeBuild/generator` |

  This is narrowing-based proof, not delete-only, and the implementer's note
  about the fixture putting the audited and the replaced package in the same
  vendor tree is accurate — the `firstParty` guard is what does the work there,
  not the path.
- **Unmutated package is green**: `go test ./internal/godriver/ -count=1` → `ok
  github.com/relux-works/curator/internal/godriver 41.859s`.
- **PR 40 lanes were green pre-merge**: Test (ubuntu/macos/windows), Race
  (ubuntu/macos), Lint, Gate self-test (ubuntu/macos/windows), Interop
  conformance gate, Naming gate — all SUCCESS.

## Finding (blocking): the carve-out makes `go_forbidden_compiler_directive` bypassable

`internal/godriver/graph.go:300` `scanSourceDirectives` streams the file in
64 KiB windows and **returns on the first window that matches any needle**:

```go
err := scanFileWindows(path, len(forbiddenCompilerDirective)-1, func(window []byte) bool {
    for index, needle := range needles {          // needles = {cgo_import_dynamic, //go:generate}
        if bytes.Contains(window, needle) {
            matched = index + 1
            return true                            // <- stops the whole scan
        }
    }
    return false
})
```

Within one window `//go:cgo_import_dynamic` wins, because it is checked first.
**Across windows it does not.** A `//go:generate` in window 1 sets `matched = 2`
and terminates the scan; a `//go:cgo_import_dynamic` in window 2 is never read.

Before `c9fe49c` this was harmless: `matched == 2` rejected unconditionally, so
the file was refused either way. After `c9fe49c`, `matched == 2` on a vendored
non-replaced package is *accepted* — and the cgo directive rides in unnoticed.

### Failure scenario

A vendored third-party module (module has no `Replace`, package directory below
`<build root>/vendor` — i.e. the exact audited class the carve-out serves) ships
one `.go` file laid out as:

```
package board

//go:generate go run ./gen          <- window 1
<~64 KiB of anything>
//go:cgo_import_dynamic libc_x x "/usr/lib/libSystem.B.dylib"   <- window 2
```

go-v1 preflight passes and the build succeeds. The Go compiler does **not**
catch this: `cmd/compile/internal/noder/noder.go` (go1.25.5, line ~311) reads
*"This is permitted for general use because Solaris code relies on it in
golang.org/x/sys/unix and others"* — `//go:cgo_import_dynamic` is legal in any
package, and the library name only has to satisfy `safeArg`, which
`/usr/lib/libSystem.B.dylib` does. curator's gate is the only thing standing
between an audited-vendor dependency and dynamically importing arbitrary host
library symbols.

### Spec basis

`profiles/manager.md` §2.3 is a containment predicate, not a first-hit one:

> Each active non-standard `GoFiles` file is scanned as exact bytes and rejected
> if it **contains** `//go:cgo_import_dynamic`, except for the audited allowlist
> `golang.org/x/sys` and `golang.org/x/sys/*` ...

A file that contains it now passes preflight. That is a conformance break, not
only a hardening gap.

### Reproduction (end to end, real `Build` entry point)

Probe attached as `BUG-260825-11nmd5_cgo-bypass-probe_test.go`; drop it into
`internal/godriver/` and run it. Against `680f6a6`:

```
=== RUN   TestReviewProbeGeneratorCarveOutHidesCgoImportDynamic
    --- PASS: .../cgo_directive_alone_(control)
        diagnostic code = "go_forbidden_compiler_directive"
    --- FAIL: .../generate_directive_in_an_earlier_window_than_the_cgo_directive
        diagnostic code = "" (err <nil>)
        BYPASS: code = "", want go_forbidden_compiler_directive
    --- PASS: .../both_directives_in_the_same_window
        diagnostic code = "go_forbidden_compiler_directive"
```

The two controls prove the gate is reachable and that same-window ordering still
works, so the failure isolates the cross-window path. Mutation-checked against
the delta: reverting `graph.go:275` to `if matched == 2 {` turns the failing
subtest into `go_generator_forbidden` — i.e. the bypass is introduced by this
delta and did not exist at `9ba552f`.

### Recommended fix

Keep scanning after a `//go:generate` hit; only short-circuit on the strictly
worse `//go:cgo_import_dynamic`:

```go
func scanSourceDirectives(path string) (int, error) {
	matched := 0
	err := scanFileWindows(path, len(forbiddenCompilerDirective)-1, func(window []byte) bool {
		if bytes.Contains(window, forbiddenCompilerDirective) {
			matched = 1
			return true // nothing weaker can override this; stop
		}
		if matched == 0 && bytes.Contains(window, []byte("//go:generate")) {
			matched = 2 // record it, but keep reading: a later window may still carry the cgo directive
		}
		return false
	})
	if err != nil {
		return 0, err
	}
	return matched, nil
}
```

Cost: a file carrying `//go:generate` is read to EOF instead of to the first
hit. Bounded by the frozen build source, and only for files that already carry
the directive.

Pin it with the attached probe (or an equivalent case folded into
`TestAuditedVendorAllowancesAreWithheldFromAReplacedModule`). The existing
`TestDirectiveScanFindsExactTokenAcrossReadBoundary` in `graph_test.go` only
covers one needle across one boundary and cannot catch this — a second case
with both needles in different windows is the missing negative.

## Non-blocking observations (record, do not rework now)

1. **The `firstParty` guard on `//go:generate` is stricter than §2.3 requires.**
   §2.3 scopes "Both exceptions ... to results whose module carries no
   replacement" where *both* means the `SFiles` exception and the
   `cgo_import_dynamic` allowlist; the replaced-module sentence then enumerates
   only `SFiles MUST be empty` and `//go:cgo_import_dynamic MUST NOT appear`.
   The `//go:generate` sentence says only "its presence in vendored `GoFiles`".
   So a replaced module's vendored package carrying `//go:generate` is rejected
   by curator where the prose would let it pass. This is fail-closed, the AC
   explicitly chose it ("build-root and first-party packages are still rejected
   with `go_generator_forbidden`"), and the conformance gate is green on it —
   but if the spec ever grows a case for it, this is where it bites.

2. **The pinned conformance suite has the same hole.** The Interop conformance
   gate passed on `c9fe49c`, so nothing at `0ed5c69` exercises a file carrying
   both directives in different windows. Worth a follow-up on `curator-spec`;
   not a blocker for curator's own fix.

3. **`classifyDeclaredInput` never scans for `//go:generate` at all**
   (`internal/godriver/moduleroots.go:227`), although its doc comment claims the
   scan surface covers "the directive, cgo, and assembly" checks of §4.2 and
   `TestDeclaredModuleDirectoriesJoinTheScanSurface` has no generator case.
   Pre-existing, untouched by this delta, and consistent with observation 1 —
   noting it so it is not mistaken for a regression later.

## Checklist disposition

- [x] Vendored generator directives are inert at build time and the carve-out is
      bounded to materialized vendor trees per decision 0005 — yes, condition and
      spec basis both verified.
- [x] Regression test pins acceptance of a vendored `//go:generate` and continued
      rejection of a first-party one — yes, and mutation-verified in both
      directions plus the build-root path.
- [ ] Implementation matches AC — the carve-out half does; "the
      `cgo_import_dynamic` allowlist is unchanged" does not hold in effect. The
      allowlist *constant* is unchanged, but the gate it guards became
      bypassable.
- [x] Solution fits project architecture — mirrors the `SFiles` carve-out, uses
      the same `firstParty` / `strictlyBelow` vocabulary.
- [x] Tests green — `internal/godriver` `ok` in 41.859s; PR 40 lanes all SUCCESS.
