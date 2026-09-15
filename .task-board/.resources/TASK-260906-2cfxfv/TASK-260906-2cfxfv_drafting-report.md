# TASK-260906-2cfxfv — closing the silent-skip hole on the environments vector families

Repository `curator`, branch `feat/interop-root-artifacts`, worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage`, base
`fb916acd60dd1881f050c5d838356235455b5b89`. Conformance authority: curator-spec `main` `87a0d006`.

## 1. The hole, restated in the mechanism's own terms

`internal/interop` read five environments vector families behind a per-family guard:

```go
payload, err := os.ReadFile(vectorPath)
if errors.Is(err, os.ErrNotExist) {
    t.Skipf("conformance root %s publishes no vectors/... (pre-environments suite; root-content)", root)
}
```

`.github/ci/skip-classes.tsv:48` gives `root-content` policy **`allow`** — in every lane, the
candidate lane included. `internal/interop` declared **none** of those families in
`.github/ci/root-artifacts.tsv`, so `suite-plan.sh` never deferred the package and
`CI_REQUIRE_FULL_ROOT=1` never fired for it. A candidate root that stopped publishing
`vectors/environments.json` would have passed the candidate lane green with the environments cases
skipped.

## 2. The shape chosen, and why not the others

**Chosen: option 1 — split the package.** The five environments conformance files moved to
`internal/interop/environments`; that package declares its four vector families and the two
byte-exact trees they cross-reference in `root-artifacts.tsv`; `internal/interop` keeps its ten
pre-environments cases and stays absent from the table.

Why the split and not option 2 (a per-family assertion inside the served lane):

- The split reuses the mechanism the repository already trusts. `suite-plan.sh` +
  `root-artifacts.tsv` + `CI_REQUIRE_FULL_ROOT` is the *same* gate that cycle 3 verified fails
  closed for `internal/envfragment` and `internal/envmarker`. Option 2 would have added a *second*,
  package-local notion of "the root must be complete", which then needs its own gate to prove it is
  wired — the exact F9 shape (a gate installed but never run) this task exists to remove.
- The split fails **before `go test` starts**. On a root missing a family the lane never compiles
  the suite: `suite-plan.sh` names the package and the artefact and exits 1. Option 2 fails
  *during* the suite, which is later, noisier, and per-case rather than per-root.
- The split makes the skip class honest. The only skip left in the new package is
  `t.Skip("CURATOR_CONFORMANCE_ROOT is not set")` — class `root-unset`, policy
  **`deferred-only`**, legitimate solely for a package `suite-plan.sh` actually deferred. Option 2
  would have left `root-content` rows in the ledger, and `root-content` is `allow` everywhere.
- Wholesale registration of `internal/interop` (the one-line "fix") was rejected outright: it defers
  the *whole* package on the `SPEC_PIN` root, dropping ten pre-environments conformance cases from
  every default lane. Section 6 measures that this did not happen.

The cost of the split is one new package directory and a duplicated 30-line test helper
(`suiteRoot`, `requireFamily`, `readRootFile`). That is the whole price.

`requireFamily` is the second half of the shape: inside the new package a missing declared family
is a **`t.Fatalf`, never a skip**, because a root that `suite-plan.sh` said serves this package and
then does not publish a declared family has contradicted the plan. Same for a family published at
the wrong revision (`context_materialization_test.go`: the old `t.Skipf("is a pre-revision root")`
is now a `t.Fatalf`).

## 3. Roots — materialized as plain checkouts, verified against `manifest.json`

`git archive` was **not** used: the repository's `.gitattributes` filters rewrite
`conformance/v1/fixtures/byte-exact/subst.txt` from 40 to 65 bytes and corrupt the root silently.
Both roots are plain `git checkout` trees, verified file-by-file with
`.temp/TASK-260906-2cfxfv/verify-root.py` (sha256 per manifest entry, plus a walk for undeclared
extras).

| root | checkout | protocol_version | declared | verified | missing | mismatched | extra | exit |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: | ---: |
| `pin` (SPEC_PIN `0ed5c691e9…`) | `0ed5c69` | 1.0.0-rc.9 | 691 | 691 | 0 | 0 | 0 | 0 |
| `candidate` (authority) | `87a0d00` | 1.0.0-rc.9 | 1047 | 1047 | 0 | 0 | 0 | 0 |

## 4. The negative runs — the lane fails **by name**, it does not skip

Each root below is the verified `candidate` tree with exactly **one** declared artefact removed —
the narrowest possible weakening of a root. Lane:
`CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh`.

| removed artefact | lane exit | what the lane printed |
| --- | ---: | --- |
| `vectors/environments.json` | **1** | `FAIL internal/interop/environments was deferred … missing: vectors/environments.json` |
| `vectors/context-versions.json` | **1** | `missing: vectors/context-versions.json` |
| `vectors/context-detectors.json` | **1** | `missing: vectors/context-detectors.json` |
| `vectors/snapshot-acquisition.json` | **1** | `missing: vectors/snapshot-acquisition.json` |
| `expected/environments` | **1** | `missing: expected/environments` |
| `fixtures/byte-exact` | **1** | `missing: fixtures/byte-exact` |

The six rows were produced by run `RUN-260907-a1a8cf`
(`.temp/TASK-260906-2cfxfv/logs/verify-final.txt`). **This run re-executed the
`vectors/environments.json` case end to end on the tree as it stands** and observed exit **1** with
the full text:

```
defer internal/interop/environments
      the supplied root publishes none of: vectors/environments.json
      it runs with CURATOR_CONFORMANCE_ROOT unset, taking the path its own tests implement
FAIL  internal/interop/environments was deferred, but this lane requires a root that serves the whole module
      missing: vectors/environments.json

suite-plan: served=72 deferred=1 excluded=0
suite-plan: FAILED
test-gate: suite-plan exit=1 -- refusing to run a suite whose shape is already wrong
```

Before this change the same doctored root produced a **green** lane with the environments cases
recorded as tolerated `root-content` skips.

The remaining five rows are additionally covered first-hand by `gate-selftest.sh`, which this run
executed at exit 0: it deletes **each** of the six declared artefacts in turn from a serving root
and asserts the plan fails, names the package, names the artefact, and still serves
`internal/interop` — 18 assertions, lines 78–108 of `run2/gate-selftest.txt`.

## 5. Two-root partition — truthful on both

| root | served | deferred | excluded | `internal/interop` | `internal/interop/environments` |
| --- | ---: | ---: | ---: | --- | --- |
| `pin` (SPEC_PIN) | 69 | 4 | 0 | **served** | **deferred** — publishes none of the six |
| `candidate` (intact) | 73 | 0 | 0 | **served** | **served** |
| `candidate` less `environments.json` | 72 | 1 | 0 | **served** | **deferred → lane FAILS** |

On the executed default lane the `pin` root produced exactly the intended outcome: all ten
pre-environments `internal/interop` cases **pass**, the six `internal/interop/environments`
conformance cases record `root-unset` / `tolerated-by-ledger` / `CURATOR_CONFORMANCE_ROOT is not
set`, and the three contract cases still **pass** (they read committed files, never the root).

On `pin` the four deferrals are `internal/config`, `internal/envfragment`, `internal/envmarker`,
`internal/interop/environments` — exactly the packages whose declared families rc.9 does not
publish, and nothing else.

## 6. The default lane keeps every case it ran

Top-level `Test*` functions declared by `internal/interop` at base `HEAD` vs. the union of
`internal/interop` + `internal/interop/environments` after the split:

```
--- diff HEAD vs worktree union ---
7a8
> TestEveryFamilyThisPackageReadsIsDeclaredInRootArtefacts
15a17
> TestNoCaseHereSkipsForAnythingButTheDeferredRoot
16a19
> TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass
```

Nothing removed; three cases added. In the observed candidate lane the ten non-environments cases
(`TestCanonicalJSONVectors`, `TestGoldenContextCopy`, `TestGoldenFederationSemantics`,
`TestGoldenMarkerObject`, `TestGoldenRegistryObjects`, `TestGoldenSnapshotHash`,
`TestIdentifierVectors`, `TestLocaleSelectorVectors`, `TestManagerConfigVectors`,
`TestSourceIdentityVectors`) all pass under `internal/interop`, and `gate-selftest.sh` asserts
`internal/interop is never deferred by a partial root`.

## 7. Ledger rows, before and after

Before — six rows, all class `root-content` (policy **allow in every lane**), all under
`internal/interop`:

| case | must-run | skip-ok | class |
| --- | --- | --- | --- |
| TestConformanceSnapshotAcquisition | linux,darwin,windows | linux,darwin,windows | `root-content` |
| TestConformanceContextVersions | linux,darwin,windows | linux,darwin,windows | `root-content` |
| TestConformanceContextResolution | linux,darwin,windows | linux,darwin,windows | `root-content` |
| TestConformanceEnvironmentsHeader | linux,darwin,windows | linux,darwin,windows | `root-content` |
| TestConformanceEnvironmentsMonolithic | linux,darwin,windows | linux,darwin,windows | `root-content` |
| TestConformanceContextDetectors | linux,darwin,windows | linux,darwin,windows | `root-content` |

After — the same six under `internal/interop/environments` at class `root-unset` (policy
**`deferred-only`**), plus three contract rows that tolerate no skip at all:

| case | must-run | skip-ok | class |
| --- | --- | --- | --- |
| TestConformanceSnapshotAcquisition | linux,darwin,windows | linux,darwin,windows | `root-unset` |
| TestConformanceContextVersions | linux,darwin,windows | linux,darwin,windows | `root-unset` |
| TestConformanceContextResolution | linux,darwin,windows | linux,darwin,windows | `root-unset` |
| TestConformanceEnvironmentsHeader | linux,darwin,windows | linux,darwin,windows | `root-unset` |
| TestConformanceEnvironmentsMonolithic | linux,darwin,windows | linux,darwin,windows | `root-unset` |
| TestConformanceContextDetectors | linux,darwin,windows | linux,darwin,windows | `root-unset` |
| TestEveryFamilyThisPackageReadsIsDeclaredInRootArtefacts | linux,darwin,windows | `-` | `-` |
| TestNoCaseHereSkipsForAnythingButTheDeferredRoot | linux,darwin,windows | `-` | `-` |
| TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass | linux,darwin,windows | `-` | `-` |

The rows' notes were rewritten to say what the tests actually assert (byte-exact re-extraction and
`expected_sha256` equality for snapshot-acquisition; `pkgversion` parse/order/range-match and
rejection of every invalid spelling for context-versions; and so on), replacing the old
"a root that publishes no such vector records a root-content skip" phrasing that described the hole
rather than the assertion.

## 8. The three self-defending cases (`contract_test.go`)

Registering artefacts in a `.tsv` is only worth something if the registration cannot drift away from
what the code reads. Three cases hold it in place, all driven from the package itself:

- `TestEveryFamilyThisPackageReadsIsDeclaredInRootArtefacts` — parses every `requireFamily(t, root,
  "…")` call out of the package's own source, adds the two cross-referenced trees, and requires the
  `root-artifacts.tsv` row to declare each. A case that starts reading an undeclared family cannot
  leave that family droppable.
- `TestNoCaseHereSkipsForAnythingButTheDeferredRoot` — walks the package's **AST** and requires
  exactly one `Skip*` call, inside `suiteRoot`. Any other skip, under any receiver name, would
  classify as `root-content`.
- `TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass` — requires every `TestConformance*`
  row for this package to declare class `root-unset` on all three platforms, and every
  `TestConformance*` function in the package to have a row.

`gate-selftest.sh` derives the required-artefact set from the package's `requireFamily` calls, not
from the row it is checking — a deliberately non-self-referential read, so shrinking the row shrinks
nothing on the checking side.

## 9. Mutant evidence

Harness: `.temp/TASK-260906-2cfxfv/mutants.py`. Each mutant leaves the gate **present** and weakens
it by the smallest edit that admits **one** member of the class it must reject. Every `-run` level
is anchored separately and `=== RUN` lines are counted, so a filter matching nothing is visible as
`absent` rather than reading as a pass.

**The whole harness was re-run in this session, first-hand, on the committed tree**
(`.temp/TASK-260906-2cfxfv/run2/mutant-run.txt`): harness exit 0, **22 checks, 14 killed,
6 SURVIVED**, and `git status --short` afterwards is `(clean)` — every mutant restored. One change
was needed first: the harness drove its lanes at `GO_TEST_TIMEOUT=8m`, the budget that makes
`internal/install` time out on a loaded host. That would have **forged** a kill for every mutant
expecting nonzero and an `UNEXPECTED` for M6, which expects zero. The lanes now run at CI's real
unix budget of `30m`.

| mutant | narrows the gate to | check | exit | verdict |
| --- | --- | --- | ---: | --- |
| M0 none (baseline) | — | `TestEveryFamily…` (RUN×1 → pass) | 0 | as expected |
| M0 none | — | candidate lane on root missing `environments.json` | 1 | killed |
| M1 delete the whole row | nothing declared at all | `TestEveryFamily…` (RUN×1 → fail) | 1 | **killed** |
| M1 | | `gate-selftest.sh` | 1 | **killed** |
| M1 | | *plan level*: `suite-plan.sh`, root missing `environments.json` | 0 | SURVIVED — this survival **is** the mutant's effect: the plan-level gate is gone |
| M1 | | *behavioural*: candidate lane, same root | 1 | **killed** — the plan now serves the package and `requireFamily` fails the cases red |
| **M2 drop one artefact from the row** | admits a root missing `vectors/context-detectors.json` | `TestEveryFamily…` (RUN×1 → fail) | 1 | **killed** |
| M2 | | `gate-selftest.sh` | 1 | **killed** |
| M2 | | *plan level*: root missing `context-detectors` | 0 | SURVIVED — the **one** member the narrowed gate now admits |
| M2 | | *plan level*: root missing `environments.json` | 1 | **killed** — the narrowing admits exactly one |
| M2 | | *behavioural*: candidate lane, root missing `context-detectors` | 1 | **killed** by `requireFamily` |
| M2 | | *behavioural*: candidate lane, root missing `environments.json` | 1 | **killed** at the plan |
| **M3 put one `t.Skipf` back** | admits a *served* root missing `context-detectors` | `TestNoCaseHereSkips…` (RUN×1 → fail) | 1 | **killed** |
| M3 | | *behavioural*: `go test` on that root (RUN×1 → **skip**) | 0 | as expected — the mutant makes it skip where the gate made it fail red |
| **M4 rename the receiver so the skip is invisible to text search** | admits a skip a `grep` gate cannot see | `TestNoCaseHereSkips…` (RUN×1 → fail) | 1 | **killed** — the AST gate sees the call |
| M4 | | *behavioural*: candidate lane on that root | 1 | **killed** by the platform-case gate, `wrong-class=True` |
| **M5 flip one ledger row to `root-content`** | admits a `root-content` skip of one case | `TestTheLedgerToleration…` (RUN×1 → fail) | 1 | **killed** |
| M5 | | *behavioural*: default lane on the SPEC_PIN root | 1 | **killed** — the deferred-root skip no longer matches the row's class |
| M6 neuter one assertion (`t.Fatalf`→`t.Logf`) | *probe for a stated bound* | `ledger-consistency.sh` | 0 | SURVIVED |
| M6 | | `go test` (contract cases, root unset) | 0 | SURVIVED |
| M6 | | `gate-selftest.sh` | 0 | SURVIVED |
| M6 | | candidate lane, intact root | 0 | SURVIVED |

M4 is the source-text-inspecting gate attacked with a mutant that **preserves** the searched-for
token and changes behaviour, and the harness ran the behavioural lane, not only the static checker.
The harness measures the naive gate's blindness rather than asserting it:

```
[a naive text gate grepping `t.Skip` in internal/interop/environments/context_detectors_test.go
 counts 0 -> it would SURVIVE this mutant]
```

The mutant reintroduces the skip as `guard := t` / `guard.Skipf(...)`, so `grep -c 't\.Skip'`
returns **0** and a text gate reports the file clean. The AST gate kills it (exit 1), and so does
the behavioural candidate lane (exit 1, platform-case gate `FATAL-wrong-class`).

**M6 is reported as a survivor and is a stated bound.** Neutering one *value* assertion inside a
conformance case changes nothing about the run's **shape** — the case still runs, still passes, and
every gate here measures shape (did the case run, on which platforms, with which skip class). None
of these gates is a substitute for the conformance assertions themselves; that bound is stated, not
hidden.

**Note on M5.** The previous run's mutant harness was killed by the launcher timeout mid-M5, so its
`finally: restore()` never ran and the M5 mutation was left in the working tree
(`internal/interop/environments :: TestConformanceContextDetectors` at class `root-content`). This
run found it because `TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass` fails on it —
exit 1, naming the row and the class. That is an unplanned, first-hand re-observation of M5's kill.
The row is restored to `root-unset`. `root-artifacts.tsv` and `context_detectors_test.go` were
verified un-mutated (M1/M2/M3/M4/M6 patterns absent; the harness log timestamps confirm the kill
landed inside the M5 block). The re-run above then executed M5 to completion — contract case exit 1,
SPEC_PIN default lane exit 1 — and restored the row itself.

## 10. Gate table — observed exit codes, this run, each command a standalone unpiped process

| command | exit | note |
| --- | ---: | --- |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `gofmt -l cmd internal` | 0 | 0 files listed |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `bash .github/ci/no-broad-suppression.sh` | 0 | |
| `bash .github/ci/ledger-consistency.sh <ev>` | 0 | 228 rows across linux darwin windows |
| `bash .github/ci/gate-selftest.sh` | 0 | 125 passed, 0 failed |
| `go test ./internal/interop/environments/ -run '^Test(TheLedgerToleration\|EveryFamily\|NoCaseHere)' -v` | 0 | 3 `=== RUN`, 3 `--- PASS` |
| `bash .github/ci/suite-plan.sh <pin>` | 0 | served=69 deferred=4 |
| `test-gate.sh` CANDIDATE lane, intact root, `CI_REQUIRE_FULL_ROOT=1`, `GO_TEST_TIMEOUT=30m` | 0 | 9/9 environments cases pass, 0 interop skips |
| `test-gate.sh` CANDIDATE lane, root less `vectors/environments.json`, `CI_REQUIRE_FULL_ROOT=1` | **1** | **expected red** — fails by name at the plan |
| `test-gate.sh` DEFAULT lane, SPEC_PIN root, `GO_TEST_TIMEOUT=30m` | 0 | 10/10 pre-environments cases pass; 6 environments cases record `root-unset`, `tolerated-by-ledger` |
| `python3 .temp/TASK-260906-2cfxfv/mutants.py` (7 mutants, 22 checks, lanes at 30m) | 0 | 14 killed, 6 SURVIVED — all 6 accounted for in §9; tree restored `(clean)` |

The two `test-gate` lanes ran **sequentially**, never alongside a `-race` suite.

### One red run worth naming

The candidate lane was first run at `GO_TEST_TIMEOUT=8m` (the previous run's choice) and came back
**exit 1**: `internal/install` panicked with `test timed out after 8m0s` at 483.4s, and the
platform-case gate then reported `TestDryRunTouchesNothing` and `TestEndToEndInstall` as
`required case never ran on darwin`. That is a budget artefact, not a regression: `internal/install`
took 443.6s in the previous run at the same 8m budget, and CI's real per-package budget is **30m**
on unix (`.github/workflows/ci.yml:135,413`; `test-gate.sh` default `30m`). Re-run at the CI budget
the same lane is **exit 0**. Neither the change nor `internal/install` was touched between the two
runs; only `GO_TEST_TIMEOUT` differed.

## 11. Coverage ratio against the acceptance criteria

`n of m` AC rows driven through the production entry point by a named committed test or gate case:

| # | AC row | driven by | production call site |
| --- | --- | --- | --- |
| 1 | a candidate root missing an environments family fails the candidate lane by name | `gate-selftest.sh` cases `a candidate root missing <artefact> fails the lane` + `the failure names …`, ×6 artefacts | `.github/ci/suite-plan.sh` via `.github/ci/test-gate.sh` |
| 2 | …shown for at least two families | 6 families, negative lanes §4 | same |
| 3 | the same lane is green on the unmodified candidate root | candidate lane exit 0, §10 | same |
| 4 | the SPEC_PIN root defers exactly what it should | `suite-plan.sh <pin>` §5 + `gate-selftest.sh` `internal/interop is never deferred by a partial root` | same |
| 5 | the default lane keeps every non-environments `internal/interop` case | §6 diff + observed-cases; `ledger-consistency.sh` compiles-in check across 3 GOOS | `go list` per GOOS in `ledger-consistency.sh` |
| 6 | ledger rows describe what their tests assert | `TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass` (class + completeness); notes rewritten by hand | `.github/ci/platform-case-gate.sh` reads the same rows |
| 7 | `gate-selftest.sh` green | exit 0, 125/0 | — |
| 8 | roots are plain checkouts verified against `manifest.json` | `verify-root.py`, 691/691 and 1047/1047, exit 0 | — |

**8 of 8 driven.** No AC row is carried by prose alone.

**Stated bounds** (things these gates do **not** establish):

- The gates measure the **shape** of a run — which cases ran, on which platforms, with which skip
  class. They do not verify the conformance assertions inside those cases; M6 demonstrates that
  explicitly and survives on purpose.
- Only `GOOS=darwin` was executed. `linux` and `windows` coverage is `go list`-derived
  (compiled-in / not-compiled-in per GOOS via `ledger-consistency.sh`), not executed here.
- `gate-selftest.sh`'s per-artefact loop deletes one artefact at a time. A root that publishes a
  family with **wrong content** is out of scope for the plan-level gate by construction; that is
  the conformance cases' job, and `requireFamily`/the `header_type_line` `t.Fatalf` cover the two
  cases (absent, wrong revision) where it used to skip.

## 12. Files changed

| file | change |
| --- | --- |
| `.github/ci/root-artifacts.tsv` | new row for `internal/interop/environments` (6 artefacts); header rewritten to name the three shapes and say which one is forbidden and why |
| `.github/ci/platform-cases.tsv` | six rows moved to the new package at class `root-unset`, notes rewritten; three contract rows added; section header explains the class choice |
| `.github/ci/gate-selftest.sh` | +71 lines: the per-artefact hole-punching loop, the required-set derivation from `requireFamily` calls, and the `internal/interop is never deferred` assertion |
| `internal/interop/environments/{context_detectors,context_materialization,context_resolution,context_versions,snapshot_acquisition}_test.go` | moved from `internal/interop`; every per-family `t.Skipf` replaced by `requireFamily`; the pre-revision `t.Skipf` replaced by `t.Fatalf` |
| `internal/interop/environments/suite_test.go` | new — `suiteRoot` (the one legitimate skip), `requireFamily`, `readRootFile`, and the doc comment stating the contract |
| `internal/interop/environments/contract_test.go` | new — the three self-defending cases of §8 |
| `internal/interop/golden_test.go` | doc comment: why this package is deliberately absent from `root-artifacts.tsv` and where the environments cases went |

## 13. One thing fixed in the inherited tree

Besides the M5 restoration (§9), `contract_test.go`'s final `t.Logf` printed the literal string
`[root-unset]` for every row regardless of the class it had just read — a log line asserting
something it had not measured, in a test whose whole subject is that class. It now prints the
observed class (`declared` became `map[string]string`, `sortedKeys` became generic). Behaviour of
the assertions is unchanged; the diagnostic no longer lies.

## 14. Workspace note

The spawn prompt's "Story Workspace" block points at
`curator-spec/.temp/STORY-260905-1n0iy8/worktree`, which is the **curator-spec** repository and has
no `internal/` tree. The task Scope, both producer briefs, and the inherited work all name the
**curator** worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage` on
`feat/interop-root-artifacts`. Work was done and committed there, per the briefs. The curator-spec
story worktree is untouched.
