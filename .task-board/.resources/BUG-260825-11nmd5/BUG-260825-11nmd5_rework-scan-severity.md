# BUG-260825-11nmd5 — rework: the directive scan resolves by severity, not by first hit

Rework for the blocking finding in `BUG-260825-11nmd5_review-verdict.md` against
PR #40 (`c9fe49c`, merged as `680f6a6`).

- **PR**: https://github.com/relux-works/curator/pull/42
- **Branch**: `fix/BUG-260825-11nmd5-directive-scan-shortcircuit`
- **Commit**: `438d557`
- **Base**: `680f6a6` (`origin/main`, i.e. PR #40 already landed)
- **Files**: `internal/godriver/graph.go`, `internal/godriver/graph_test.go`,
  `internal/godriver/moduleroots_test.go`, `LOGBOOK.md`

## 1. The defect, reproduced before it was fixed

`scanSourceDirectives` streamed an active Go file in 64 KiB windows and returned
on the **first window matching either needle**:

```go
err := scanFileWindows(path, len(forbiddenCompilerDirective)-1, func(window []byte) bool {
    for index, needle := range needles {   // {cgo_import_dynamic, //go:generate}
        if bytes.Contains(window, needle) {
            matched = index + 1
            return true                    // <- stops the whole scan
        }
    }
    return false
})
```

Within one window the cgo directive won, because it was checked first. **Across
windows it did not.** That was harmless while `matched == 2` rejected
unconditionally. PR #40 made `//go:generate` exempt inside a materialized vendor
tree, so a generate hit in window 1 now terminated the scan *and* the package was
then admitted — with a `//go:cgo_import_dynamic` in window 2 never read.

Reproduced on unfixed `origin/main` with the reviewer's probe, through the real
`Build()` entry point, on an audited non-replaced vendored module:

```
--- PASS: .../cgo_directive_alone_(control)
        diagnostic code = "go_forbidden_compiler_directive"
--- FAIL: .../generate_directive_in_an_earlier_window_than_the_cgo_directive
        diagnostic code = "" (err <nil>)          <- BYPASS: the build SUCCEEDED
--- PASS: .../both_directives_in_the_same_window_(control)
        diagnostic code = "go_forbidden_compiler_directive"
FAIL	github.com/relux-works/curator/internal/godriver	2.052s
```

The two passing controls prove the case is reachable and that the *only*
distinguishing feature is which window each directive lands in.

`profiles/manager.md` §2.3 is a **containment** predicate — the file is rejected
if it *contains* the directive, wherever it sits — so this was a conformance
break, not only a hardening gap. The Go compiler is not a backstop:
`cmd/compile/internal/noder` permits `//go:cgo_import_dynamic` for general use
(its comment names Solaris code in `golang.org/x/sys/unix`), and
`/usr/lib/libSystem.B.dylib` satisfies its argument check.

## 2. The fix

The scan resolves by **severity, not by first hit**. Only the cgo directive,
which nothing weaker can override, ends the read early; a generate hit is
recorded and the file is still read to EOF.

```go
func scanSourceDirectives(path string) (int, error) {
	matched := directiveNone
	err := scanFileWindows(path, len(forbiddenCompilerDirective)-1, func(window []byte) bool {
		if bytes.Contains(window, forbiddenCompilerDirective) {
			matched = directiveCgoImportDynamic
			return true
		}
		if matched == directiveNone && bytes.Contains(window, generatorDirective) {
			matched = directiveGenerate
		}
		return false
	})
	...
}
```

The three verdicts are named constants (`directiveNone`,
`directiveCgoImportDynamic`, `directiveGenerate`) so the two call-site branches
read as a verdict rather than as `matched == 1` / `matched == 2`.

Unchanged: the vendored-generator carve-out condition itself, the
`golang.org/x/sys` cgo allowlist, and the `SFiles` vendored-assembly exception.

Cost: a file carrying `//go:generate` is read whole instead of to the first hit.
Bounded by the frozen build source, and only for files that already carry it.

**Severity encoding is lossy but sound.** A file with both directives reports
only the cgo verdict. That never hides a rejection: the generate branch is
weaker in every reachable combination. For a vendored non-replaced package the
generate directive is exempt anyway; for a first-party or replaced one the cgo
branch already rejects. The one asymmetric case — a vendored `golang.org/x/sys`
package where cgo is allowlisted — is below the vendor tree and non-replaced, so
its generate directive is exempt too.

## 3. Tests

- `TestDirectiveScanReportsTheStrongestDirectiveAcrossWindows`
  (`graph_test.go`) — scanner level: generate-before-cgo, cgo-before-generate,
  generate-only, generate in a later window, neither.
- `TestVendoredGeneratorCarveOutDoesNotHideACgoImportDynamic`
  (`moduleroots_test.go`) — drives the real `Build()` production entry point over
  the exact class the carve-out serves (audited, non-replaced, below the build
  root vendor tree), with the same-window and cgo-only controls retained.

The pre-existing `TestDirectiveScanFindsExactTokenAcrossReadBoundary` covers one
needle across one boundary and structurally cannot catch this.

### Mutants applied and reverted — all four killed

| # | Mutant | Killed by |
| --- | --- | --- |
| 1 | Reintroduce `return true` on the generate branch (delete-only) | `.../generate_before_cgo`; e2e middle case |
| 2 | Check generate first, so it wins within a window too | both of the above **plus** the same-window control |
| 3 | **Narrowing:** keep scanning, but gate the cgo check on `matched == directiveNone` so a recorded generate suppresses a later-window cgo hit | `.../generate_before_cgo`; e2e middle case |
| 4 | Remove the vendored-generator carve-out entirely | PR #40's `//go:generate.../allowed_for_an_audited_third-party_module` |

Mutant 3 is the one that proves the bound. It keeps the "keep reading" change and
still reddens, so the tests pin *severity across the whole file*, not merely
"read one more window". Mutant 4 proves the hardening did not quietly undo the
relaxation it is protecting.

## 4. Gates — real exit codes, macOS host

| Gate | Exit | Detail |
| --- | ---: | --- |
| `go build ./...` | 0 | |
| `go test ./internal/godriver -count=1` | 0 | ok, 43.282s |
| `go test -race ./internal/godriver -count=1` | 0 | ok, 79.675s |
| `go vet ./...` | 0 | |
| `gofmt -l .` | 0 | no output |
| `golangci-lint run` | 0 | 0 issues |
| `bash .github/ci/gate-selftest.sh` | 0 | 81 passed, 0 failed |
| `bash .github/ci/no-broad-suppression.sh` | 0 | |

**Not run locally, stated explicitly:** `make ci-test` and `make race` require
`CURATOR_CONFORMANCE_ROOT` (a materialized `<curator-spec>/conformance/v1` at
`SPEC_PIN`), and the linux/windows lanes have no local runner on this host. CI
covers all of them — see the lane table below.

### CI lanes on PR #42 (run 32800314357, commit `438d557`)

`gh pr checks 42` exit 0; rollup **11 SUCCESS, 1 SKIPPED**; PR `MERGEABLE`.

| Lane | Result | Duration |
| --- | --- | ---: |
| Test (ubuntu-latest) | pass | 1m46s |
| Test (macos-latest) | pass | 10m23s |
| Test (windows-latest) | pass | 25m57s |
| Race (ubuntu-latest) | pass | 3m10s |
| Race (macos-latest) | pass | 13m18s |
| **Interop conformance gate** | **pass** | 21s |
| Lint | pass | 28s |
| Gate self-test (ubuntu-latest) | pass | 6s |
| Gate self-test (macos-latest) | pass | 9s |
| Gate self-test (windows-latest) | pass | 21s |
| Naming gate | pass | 8s |
| Candidate suite | skipping | — |

The Interop conformance gate is the lane that discharges the AC's "conformance
suite against `SPEC_PIN` stays green"; it runs with the pinned root that the
local host cannot materialize.

**PR #42 is left OPEN and unmerged** — lanes are green pre-merge, and merging is
the reviewer/orchestrator's step, not the developer role's.

## 5. Non-blocking observations from the verdict

1. The `firstParty` guard on `//go:generate` is stricter than §2.3 requires.
   Fail-closed, deliberately chosen by the AC, left as-is.
2. The pinned conformance suite has the same early-exit hole — a `curator-spec`
   follow-up, out of scope here.
3. `classifyDeclaredInput` (`moduleroots.go`) never scans for `//go:generate`
   despite its doc comment. Pre-existing, untouched.

Generalization worth carrying: **an early-exit byte scan silently weakens a
containment predicate to "contains, in the prefix before some other token".** It
is only safe while every needle carries the same verdict. Worth re-reading every
other early-exit scan in the tree with that sentence in hand.

## 6. Worktree hazard for the orchestrator

The `STORY-260822-2lvw0e` story worktree (HEAD `903af23`, **pre-PR-40**) still
carries uncommitted work from the timed-out run `RUN-260825-6e4450`: a divergent
second implementation of the same carve-out using a `Module.Main` guard instead
of the `firstParty` guard that actually landed, plus its own LOGBOOK entry. It is
dead work against a stale base and must not be committed or integrated. A patch
is preserved at `.temp/BUG-260825-11nmd5/evidence/story-worktree-superseded.patch`.
PR #42 is based on `origin/main` and is the live delta.
