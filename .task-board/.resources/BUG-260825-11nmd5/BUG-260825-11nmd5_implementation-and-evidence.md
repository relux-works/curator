# BUG-260825-11nmd5 — vendored `//go:generate` accepted, bounded to the vendor tree

## Problem

`9ba552f Restore Go generator rejection` dropped the vendor carve-out from
`internal/godriver/graph.go`, so `go-v1` rejected **any** active Go file
containing `//go:generate` with `go_generator_forbidden`, including audited
third-party vendored code.

This contradicts the normative profile. `curator-spec profiles/manager.md` §2.3:

> `//go:generate` in `GoFiles` is inert — managers MUST NOT run generators and
> `go build -mod=vendor` does not execute them; its presence in vendored
> `GoFiles` (vendor already materialized) does not fail preflight.

`decisions/0005-vendored-go-boundary-relaxation.md` names the motivating case:
`clipperhouse/displaywidth` and `skill-project-management`. Concretely,
`task-board-tui` cannot build because `bubbletea` → `charmbracelet/x/ansi` →
`clipperhouse/displaywidth` ships a bare `//go:generate` in `gen.go`, and no
released version drops it.

## Change

`internal/godriver/graph.go`

- The generator branch now reads
  `if matched == 2 && !vendoredDependency(item, validation.BuildRoot)`.
- New `vendoredDependency(item, buildRoot)` helper: true only when the result
  carries module metadata, that module is **not** the main module, and the
  package directory is strictly below `<buildRoot>/vendor`.

The main-module guard is what keeps the bound honest. A bare path-prefix test
would hand the exemption to any first-party package whose directory happens to
sit below `vendor/` (a package-declared `source_dir` can name such a path). A
main-module result is first-party by definition and stays rejected regardless of
where its directory sits.

Deliberately unchanged:

- `//go:cgo_import_dynamic` and its `golang.org/x/sys` allowlist — untouched;
  the carve-out is directive-specific and is covered by a negative test.
- The `SFiles` vendored-assembly carve-out — untouched.
- Every other rejection class in `validatePackageInputs`.

`CHANGELOG.md` was not touched: this repository records changelog entries at
release milestones, not per bugfix (the last twelve commits touching it are two
feature landings).

## Tests

`internal/godriver/graph_test.go` —
`TestPackageGraphExemptsGeneratorDirectiveOnlyBelowTheVendorTree`

| Case | Expectation |
| --- | --- |
| vendored dep `GoFiles` carries `//go:generate` | graph validates, no error |
| build-root (main) package carries it | `go_generator_forbidden` |
| first-party main-module package parked below `vendor/` | `go_generator_forbidden` |
| vendored dep carries `//go:cgo_import_dynamic` | `go_forbidden_compiler_directive` |

Each subtest rewrites both scanned files to a clean baseline before its own
setup, so no case can pass because an earlier one left a rejectable byte behind.

`internal/godriver/build_test.go` —
`TestBuildCompilesThroughAVendoredGeneratorDirective` drives the production
entry point `Build(...)`, not just the validator: with a real materialized
vendor tree carrying the directive, preflight passes and the worker is observed
issuing exactly one `list` and one `build` with the expected argv.

### Mutation evidence (the tests fail when the gate moves)

Every mutant was applied to `internal/godriver/graph.go`, run, then reverted.

| Mutant | Result |
| --- | --- |
| restore unconditional rejection (`matched == 2`) — the pre-fix behaviour | FAIL: `graph_test.go:123 vendored generator directive must not fail preflight: go-v1 go_generator_forbidden` |
| widen: drop the main-module guard, keep the path prefix | FAIL: `first-party_package_below_the_vendor_tree: error = <nil>, want go_generator_forbidden` |
| widen: delete the gate entirely (`if false`) | FAIL: `build_root_package` and `first-party_package_below_the_vendor_tree`, both `error = <nil>, want go_generator_forbidden` |
| restore unconditional rejection, against the `Build(...)` test | FAIL: `build_test.go:538 vendored generator directive stopped the build: go-v1 go_generator_forbidden` |

The narrowing mutant (row 2) is the load-bearing one: it proves the bound is a
first-party/third-party distinction, not a path-prefix string match.

## Gate results

Conformance root materialized from `relux-works/curator-spec` at the committed
`SPEC_PIN` `00b1688a9b2457ca397a0bb550acf47cad8ee967`
(`git archive 00b1688 conformance`), exported as `CURATOR_CONFORMANCE_ROOT`.
`suite-plan.sh` against that root: `served=34 deferred=7 excluded=0`, matching
what CI plans for this pin on darwin.

| Command | Exit | Log |
| --- | ---: | --- |
| `suite-plan.sh <pin-root> .temp/BUG-260825-11nmd5/plan` | 0 | `plan/` |
| `go test -count=1 internal/godriver` (root exported) | 0 | `godriver-served-02.log` |
| `go test -race -count=1 internal/godriver` (root exported) | 0 | `godriver-race-01.log` |
| `go test` served chunk 1 — `adapters`…`identity`, 15 pkgs | 0 | `served-rest-01.log` |
| `go test` served chunk 2 — `install/atomicity`…`registry`, 8 pkgs | 0 | `served-rest-02.log` |
| `go test` served chunk 3 — `runtimestore`…`version`, 9 pkgs | 0 | `served-rest-03.log` |
| `go test -count=1 cmd/curator` (root exported) | 0 | `cmd-curator-02.log` |
| `go test` deferred set, 7 pkgs, `CURATOR_CONFORMANCE_ROOT` **unset** | 0 | `deferred-01.log` |
| `golangci-lint run` | 0 | `lint-02.log` |
| `go vet ./...` | 0 | `vet-02.log` |
| `gofmt -l cmd internal` (empty) | 0 | — |
| `go build -o /dev/null ./cmd/curator` | 0 | — |
| `.github/ci/gate-selftest.sh` | 0 | `gate-selftest-01.log` — 75 passed, 0 failed |
| `.github/ci/ledger-consistency.sh` | 0 | `ledger-01.log` — 72 rows across linux darwin windows |
| `.github/ci/no-broad-suppression.sh` | 0 | `no-broad-suppression-01.log` |

The suite was run as bounded sequential chunks rather than through
`make ci-test`, because a single shell call in this run is time-bounded and the
whole planned set exceeds it. The chunks are exactly the plan's served and
deferred lists, with the deferred set run with the variable unset as
`test-gate.sh` does. Nothing in the plan was skipped.

One honest caveat, and it is not a passing gate wearing a green label: the first
`cmd/curator` run (`cmd-curator-01.log`) exited **1**. The test binary itself
printed `PASS`; the non-zero status came from Go's build cache —
`testing: can't write .../testlog.txt: file too large`. Rerunning the identical
package with `-count=1` (which makes the run non-cacheable, so no testlog is
written) exited 0 in `cmd-curator-02.log`. That is the run the table cites.

Not run here: the linux and windows lanes. This host is macOS and has no local
Linux runner; those lanes are CI's to run. `ledger-consistency.sh` did check the
per-platform claims for all three GOOS values and passed.

## Spec conformance

- `attempted-go-generate` in `conformance/v1/vectors/build-drivers.json` at the
  pin expects `go_generator_forbidden` for the condition *"package requires go
  generate output"*. The implementation vector
  (`builddriver_rejection_conformance_test.go`) drives it through the **root**
  package, which is first-party and still rejected. Green.
- The pin publishes no positive vector for the vendored case; the profile text
  and decision 0005 are the normative source, and the new tests pin them.
