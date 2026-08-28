# TASK-260827-18tswm review verdict — RUN-260827-34e3b7

## Verdict

**Changes requested.** The host-capability fixes are narrow and the local gates
are green, but the cross-conformance Rust-unavailability exception is not a
closed six-gap allowlist. It derives the permitted set from every current and
future obligation except shared artifact admission. A future seventh Rust
manager obligation would therefore be silently permitted instead of failing
the completeness gate. That is the exact high-risk failure mode this review was
asked to exclude.

Route: `to-dev` for a focused correction, then another reviewer cycle.

## Required correction

In `internal/crossconformance/suite_test.go`, replace the dynamically derived
Rust-unavailable expected set with an explicit closed enumeration of the six
manager-obligation/path pairs. Add a negative regression seam/test that supplies
an extra Rust gap and proves rejection, as well as a different/non-Rust gap.
Keep `artifact.shared_admission/rust` mandatory. Update the LOGBOOK statement
only when the implementation really is explicitly enumerated.

## Disposition by review item

1. **Accepted — Rust native-descriptor skip.** `approvedCargoDescriptors` still
   has exactly the existing `aarch64-apple-darwin` descriptor and no new digest.
   `NativeCargoDescriptorAvailable` performs only target/registry lookup. An
   approved target continues into `registerCargoAtC0`, where missing/renamed
   toolchains and executable-byte mismatch fail; `recheck` also rejects changed
   bytes or a descriptor/executable mismatch. The existing negative test mutates
   the descriptor digest and requires `rust_vendor_transform_unsupported`.
   Native arm64 `rustsource` ran for 109.084s with zero skip events. An amd64
   Rosetta run produced exactly 13 descriptor-unavailable skips and passed.

2. **Changes requested — cross-conformance Rust unavailability.** The amd64 run
   visibly skipped only the Rust projection/manager path, still ran Rust shared
   artifact admission, and logged the unavailability reason. A deliberately
   omitted non-Rust normative path made completeness fail. However, lines
   constructing `expected` loop over `crossconformance.Obligations()` and allow
   every obligation except `ObligationSharedArtifactAdmission` for Rust. This is
   not the promised exact, closed six-item enumeration: adding a seventh manager
   obligation automatically expands the exception. No negative test covers that
   mutation.

3. **Accepted — Swift compound classifier.** Both call sites require the exact
   conjunction of `Invalid manifest` and clang's full `posix_spawn failed: No
   such file or directory` diagnostic. `Invalid manifest` alone, the spawn error
   alone, a drift diagnostic, or a permit/receipt failure remains fatal. Merely
   finding Swift does not skip. Both affected packages pass locally.

4. **Accepted — npm native host derivation.** Only the two real native-tool cases
   use `newNativeNPMFixture`; the base parser and cross-target fixture remains
   frozen at Darwin/arm64/none. Linux native selection legitimately includes
   `node_modules/opt`. Production `validateInstalledTree` still rejects any
   observed package outside the selected graph, and `S07 extra package` drives
   `Materialize` and requires `closure_graph_incomplete`.

5. **Accepted — pnpm fake runner.** The fake runner now omits the link when the
   selected target is intentionally absent instead of attempting
   `os.Symlink("", link)`. The `ambient cannot satisfy` case still drives the
   production `Materialize` validator and requires `closure_graph_incomplete`;
   the full package passes.

6. **Accepted — Yarn Classic layouts.** Discovery accepts only Homebrew
   `libexec/bin/yarn.js` or an exact resolved package-root `bin/yarn.js` layout.
   Unsupported layouts return an error and the real pinned case records the
   classified refusal; it never falls through to another ambient Yarn. Positive
   tests cover both layouts and a negative test rejects a missing tool root.

7. **Accepted — cmd/curator carve-out.** Both named verified-provider cases now
   call the established inventory-derived guard. It skips only when
   `godriver.InventoryPlatform(runtime.GOOS)` has no record, so it runs on
   covered macOS. The companion production-boundary test still proves uncovered
   hosts refuse with no publication. All three targeted cases pass locally.

8. **Accepted — skip classes.** Replaying both captured macOS and Ubuntu JSON
   streams against the current table exits 0 with zero `UNCLASSIFIED` rows.
   Observed counts are pnpm 3 and Yarn Modern 4 per lane, plus the Ubuntu Yarn
   Classic network-denial case. Each new regex was tested against every other
   distinct captured reason; none cross-matched an unrelated reason. Classes are
   appropriate (`host-capability` or explicit-path `opt-in`).

9. **Accepted for patch scope; delivery evidence still pending.** Workflow
   `SPEC_PIN` and release-pin files are untouched. Diff inspection found no
   removed assertion outside the intentional Rust-unavailable branches;
   production behavior is unchanged except for the read-only registry query.
   The remote Test/Race matrix, Naming, Interop, passing run URL, and fresh CI
   artifacts are not present and were not inferred. Per the review brief those
   remain the landing Orchestrator's responsibility, but the task cannot satisfy
   its final CI acceptance criterion until they exist.

## Reviewer validation

| Check | Result |
| --- | --- |
| `go test -count=1 -json ./internal/rustsource` | pass; 0 skips |
| `GOARCH=amd64 go test -count=1 -json ./internal/rustsource` | pass; exactly 13 native-descriptor skips |
| Seven remaining adapter packages, including crossconformance | pass |
| `GOARCH=amd64 go test -count=1 -v ./internal/crossconformance` | pass; Rust unavailable is visible; shared admission still runs |
| amd64 crossconformance with normative npm/pnpm paths omitted via `-skip` | expected failure; extra non-Rust gaps rejected |
| filtered native completeness run with proving suites omitted | expected failure |
| Three targeted `cmd/curator` inventory/verified-provider cases | pass |
| Captured macOS and Ubuntu Tier-2 platform-case replay | pass; zero UNCLASSIFIED |
| `go build ./...` | pass |
| `go vet ./...` | pass |
| `gofmt -l cmd internal` | clean |
| `golangci-lint` 2.12.2 | pass; 0 issues |
| `.github/ci/gate-selftest.sh` | pass; 81/81 |
| `git diff --check` | pass |

Reviewer-generated logs are under `.temp/TASK-260827-18tswm-review/`. The
reviewer did not modify source, stage, commit, push, reset, or clean anything.
