# TASK-260906-2cfxfv — rework report 1

**Subject.** curator branch `feat/interop-root-artifacts` in
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage`, two new signed commits on top
of the reviewed head `31106aa8`:

| commit | signature | what |
| --- | --- | --- |
| `5b3d69fc` | `G` `Ivan Oparin <oparin@me.com>` | F1b — a ledger row must not repeal a skip class's policy |
| `e3a97d7b` | `G` `Ivan Oparin <oparin@me.com>` | F1 + F2 + F3 — see every skip shape; derive the artefact declaration from the vectors |

Working tree clean, nothing under `.temp/` committed, no push, PR #63 untouched, `LOGBOOK.md` not
written. Authority curator-spec `87a0d006`. Everything the cycle-1 reviewer verified first-hand at
`31106aa8` — the hole at base, the six negative runs, the two-root partition, the class move
`root-content` → `root-unset`, the six conformance ledger notes — is untouched by this rework.

---

## 1. F1 — the skip scan saw two of four receiver shapes

The reviewer injected four real, compiling skips into `TestConformanceContextDetectors` and measured
each. Two survived `TestNoCaseHereSkipsForAnythingButTheDeferredRoot` while the case really skipped
on a fully serving root, because `skipCall` required `sel.X` to be a bare `*ast.Ident`.

`skipCall` is replaced by `skipSelector`, which matches the **selector** and never the receiver's
shape:

```go
func skipSelector(node ast.Node) (*ast.SelectorExpr, bool) {
	sel, isSel := node.(*ast.SelectorExpr)
	if !isSel || !strings.HasPrefix(sel.Sel.Name, "Skip") {
		return nil, false
	}
	return sel, true
}
```

Matching the selector rather than the call also catches a `Skip*` **used as a value** — the method
value never appears as a `CallExpr.Fun` at all — and catches a method expression
(`(*testing.T).Skip`). The receiver is reported through `types.ExprString` and its position, never by
name. Widening from the exact three method names to the `Skip` prefix additionally covers helpers
such as `SkipUnless`.

Re-measured here, each shape run twice: once against the scan, once **behaviourally** against a fully
serving root to prove the injected skip is real and not merely visible.

| mutant | shape | `TestNoCaseHereSkips…` | on a SERVING root |
| --- | --- | ---: | --- |
| M1 | `t.Skipf(…)` | exit 1, RUN=1 — **KILLED** | the case skips |
| M2 | `guard := t; guard.Skipf(…)` | exit 1, RUN=1 — **KILLED** | the case skips |
| M3 | `h := holder{t: t}; h.t.Skipf(…)` — selector chain, **was a survivor** | exit 1, RUN=1 — **KILLED** | the case skips |
| M4 | `skip := t.Skipf; skip(…)` — method value, **was a survivor** | exit 1, RUN=1 — **KILLED** | the case skips |

M3 and M4 are also the DoD's *token-preserving* mutants: both leave the literal string `Skipf` in the
source, so a text scan still matches while behaviour changes. Every `-run` is anchored at one
top-level name and its `=== RUN` lines are counted, so a filter that matched nothing (`RUN=0`) is
reported as `NO-MATCH`, never as a kill — one mutant did read `NO-MATCH` on the first harness run
(a compile error from an incomplete rename) and was fixed and re-measured rather than credited.

**The bound, stated rather than implied.** The scan cannot see a skip reached through reflection, or
through a helper in another package whose own name does not begin with `Skip`. That residue is
covered behaviourally, not statically: `platform-case-gate.sh` classifies the reason of every skip
that actually happens, and an unrecognised reason is fatal there. The ledger note and the code
comment now say exactly this, replacing the false claim of coverage "under any receiver name".

---

## 2. F1b — the orchestrator's call: fixed here, and measured

`platform-case-gate.sh` put the `deferred-only` policy check in the `else` branch **after**
`if (tol != "")`, so for a case the ledger lists with a tolerated skip the policy was never reached.
A `root-unset`-worded skip in a package the lane **serves** recorded `tolerated-by-ledger`.

The check now also runs inside the tolerated branch. A row tolerates a skip; it does not repeal the
policy of its class.

**Does any existing tolerated row change verdict?** No — measured two ways, not argued.

*Static blast radius.* The shipped ledger carries 70 tolerated rows. Only `root-unset` is
`deferred-only`; the other 50 are policy `allow` and cannot reach the new branch.

| tolerated class | rows | policy | reachable by the new branch |
| --- | ---: | --- | --- |
| `host-capability` | 31 | allow | no |
| `platform-control` | 16 | allow | no |
| `root-content` | 2 | allow | no |
| `opt-in` | 1 | allow | no |
| `root-unset` | 20 | **deferred-only** | only while its package is served |

*Measured on real streams.* Both recorded lane streams replayed through the old gate and the new
gate, same ledger, same deferred set: `skips-observed.tsv` **byte-identical** in both lanes, reports
identical apart from the evidence-directory path. 52 real skips, zero verdict changes.

*Measured on a fresh lane.* The default `SPEC_PIN` lane, re-run end to end with the fix in place,
exits **0**: 33 skips recorded, all 13 `root-unset` ones inside packages `suite-plan.sh` genuinely
deferred (`internal/config`, `internal/envfragment`, `internal/envmarker`,
`internal/interop/environments`), all still `tolerated-by-ledger`.

*Negative evidence.* `gate-selftest.sh` gains three cases driving the real gate over a real
`go test -json` stream, using a ledger row that tolerates a `root-unset` skip and requires the case
nowhere:

| stream | lane state | expected | observed |
| --- | --- | ---: | ---: |
| ledger-listed case skips `CURATOR_CONFORMANCE_ROOT is not set` | package **served** | 1 | **1** |
| same stream | package **deferred** | 0 | **0** |
| an `allow`-policy (`platform-control`) tolerated skip beside a deferred package | — | 0 | **0** |

The third case is the guard against over-narrowing: the fix must bite exactly one class.

**Narrowing mutant M9.** The branch is kept and weakened by one conjunct —
`… && cls != "root-unset"` — so the gate still rejects every other `deferred-only` class and admits
exactly the one it must reject. `gate-selftest.sh` → exit **1**, `129 passed, 1 failed`. A
delete-only mutant was not used.

**One thing left alone, deliberately.** When a skip is fatal, the Tier-1 summary still prints
`tol   pkg :: case`, because that loop reports from the ledger row rather than from the verdict. The
run's exit code is correct (the failure is printed and `fail=1` is set); the display quirk predates
this change and applies equally to `FATAL-wrong-class`. Fixing it would touch Tier-1 output for rows
outside this finding, so it is recorded here as a stated bound rather than folded in.

---

## 3. F2 — one unguarded read was undeclared, and the fix is not a fourth list

`snapshot_acquisition_test.go:75` read `tc.Expected` through `readRootFile`, which on the real
candidate root resolves to `expected/byte-exact-snapshot_sha256.txt` — declared in neither the
`root-artifacts.tsv` row, nor `crossReferencedTrees`, nor `gate-selftest.sh`'s `ENV_REQUIRED`.

The reviewer's preference was the derivation, and that is what this does — plus a runtime guard,
because the derivation alone cannot see a read whose path comes from a struct field rather than a
literal, which is exactly how this one got in.

**(a) One accountable join.** `rootPath` is now the only place the package turns a root-relative path
into a path on disk, and it refuses a path no declared artefact covers — checked against the
committed table at the moment of the read, not against a list kept beside it. `requireFamily`,
`readRootFile` and every former `filepath.Join(root, …)` go through it.

**(b) The guard cannot be walked around.** `TestNoRootReadHereEscapesTheDeclaredArtefactGuard`: outside
`suite_test.go`, the identifier `root` may only be bound, returned, passed to an accountable helper,
or formatted into a failure message. 24 occurrences checked, and 8 bindings of the root.
`TestTheRootIsAlwaysBoundToThatOneName` keeps that scan from going blind to a rename.

**(c) The declaration is derived.** `TestConformanceEveryPathTheVectorsNameIsDeclared` reads the four
vectors, collects every string in them the root resolves as a path, and asserts both directions.
Observed on the candidate root — it reproduces the reviewer's own enumeration of 26 paths exactly:

```
declared expected/byte-exact-snapshot_sha256.txt covers 1 read path(s)
declared expected/environments                   covers 24 read path(s)
declared fixtures/byte-exact                     covers 1 read path(s)
declared vectors/context-detectors.json          covers 1 read path(s)
declared vectors/context-versions.json           covers 1 read path(s)
declared vectors/environments.json               covers 1 read path(s)
declared vectors/snapshot-acquisition.json       covers 1 read path(s)
```

`crossReferencedTrees` is deleted, and the hand-appended
`ENV_REQUIRED="$ENV_REQUIRED expected/environments fixtures/byte-exact"` is deleted from
`gate-selftest.sh`. The reviewer's trap — "adding the path to the row alone makes `TestEveryFamily…`
fail" — is gone with it: the static case now asserts only the direction it can prove without a root
(every `requireFamily` literal is declared), and the superset/cross-reference direction is derived.

**The F2 hole, before and after, on the same doctored root** (verified candidate minus
`expected/byte-exact-snapshot_sha256.txt`, nothing else):

| row | `suite-plan.sh` + `CI_REQUIRE_FULL_ROOT=1` | the case |
| --- | ---: | --- |
| as committed at `31106aa8` | **exit 0**, `served=73 deferred=0` | `--- FAIL` mid-case, `open … no such file or directory` |
| after this rework | **exit 1**, `FAIL internal/interop/environments was deferred`, `missing: expected/byte-exact-snapshot_sha256.txt` | never reached |

**Mutants.**

| mutant | narrows the gate to | checker | verdict |
| --- | --- | --- | --- |
| M5 | drop `expected/byte-exact-snapshot_sha256.txt` from the row — the F2 hole itself | `TestConformanceEveryPathTheVectorsNameIsDeclared` | exit 1, RUN=1 — **KILLED**; the plan then serves a root missing that file (exit 0) |
| M6 | over-declare `expected/build-driver/marker.json` — a real root path no vector here names | same | exit 1, RUN=1 — **KILLED** |
| M7 | read the expected file with a raw `filepath.Join(root, …)` — **preserves the ident `root`** | `TestNoRootReadHereEscapesTheDeclaredArtefactGuard` | exit 1, RUN=1 — **KILLED**; the case itself still passes (exit 0, fail=0), so only the scan sees it |
| M8 | bind the root to a second name (`suite := suiteRoot(t)`) | `TestTheRootIsAlwaysBoundToThatOneName` | exit 1, RUN=1 — **KILLED**; on the same tree the flow scan reports **green**, which is the blindness this case exists to catch |
| M10 | read a real root vector this package does not declare, **through the production read path** | `rootPath`, driven by `TestConformanceContextDetectors` | exit 1, RUN=1, fail=1 — **KILLED** |
| M11 | narrow `coveringArtefact` to wave through any `vectors/` path — the guard stays, one class is admitted | same | exit 0, RUN=1, fail=0 — **survives by design**: this is the bound M10 measures |

M11 is the deliberate survivor and it states a bound rather than hiding one: weaken the covering test
by one prefix class and the read M10 kills goes through. Reported as a survivor, not as a kill.

---

## 4. F3 — the row note

Before: *"every case in the package reads one of them without a per-family guard"*. Three of the nine
cases — now five of twelve — read no root artefact at all. After: *"every conformance case in the
package reads one of them with no per-family guard — the contract cases beside them read no root
artefact at all"*.

---

## 5. Ledger and declaration, before and after

`root-artifacts.tsv`, `internal/interop/environments`:

```
- vectors/environments.json,vectors/context-versions.json,vectors/context-detectors.json,
  vectors/snapshot-acquisition.json,expected/environments,fixtures/byte-exact
+ …same four vectors…,expected/environments,expected/byte-exact-snapshot_sha256.txt,fixtures/byte-exact
```

`platform-cases.tsv` — one note corrected, one note narrowed, three rows added:

| row | change |
| --- | --- |
| `TestNoCaseHereSkipsForAnythingButTheDeferredRoot` | note no longer claims "under any receiver name"; states what the selector scan covers and where the residue is covered instead |
| `TestEveryFamilyThisPackageReadsIsDeclaredInRootArtefacts` | note narrowed to the direction it proves without a root, and points at the derived case for the rest |
| `TestNoRootReadHereEscapesTheDeclaredArtefactGuard` | **new**, `skip=- class=-` |
| `TestTheRootIsAlwaysBoundToThatOneName` | **new**, `skip=- class=-` |
| `TestConformanceEveryPathTheVectorsNameIsDeclared` | **new**, `skip=linux,darwin,windows class=root-unset` |
| section header | records that `platform-case-gate.sh` now applies the class policy to a ledger-tolerated skip too |

`ledger-consistency.sh`: 228 → **231 rows** across linux, darwin, windows.

---

## 6. The two-root partition, re-measured

| root | lane | `suite-plan` | served / deferred | `test-gate.sh` |
| --- | --- | ---: | --- | ---: |
| candidate `87a0d006` (verified) | `CI_REQUIRE_FULL_ROOT=1` | ok | 73 / 0 | **0** |
| candidate less `vectors/environments.json` | `CI_REQUIRE_FULL_ROOT=1` | **FAILED** — `missing: vectors/environments.json` | 72 / 1 | **1** |
| candidate less `expected/byte-exact-snapshot_sha256.txt` | `CI_REQUIRE_FULL_ROOT=1` | **FAILED** — `missing: expected/byte-exact-snapshot_sha256.txt` | 72 / 1 | **1** |
| `SPEC_PIN` `0ed5c691` | default | ok | 69 / 4 | **0** |

On the candidate lane all **12** cases of `internal/interop/environments` are observed `ok` by
`platform-case-gate.sh`, and `skips-observed.tsv` records **zero** `root-unset` rows. On the pin lane
the deferred set is unchanged from the reviewed head apart from the package itself, and the new
derived case correctly records the tolerated `root-unset` skip.

Both roots were materialized as **plain checkouts** (`git archive` not used) and verified file by
file against their own `manifest.json` before use:

| root | protocol_version | declared | verified | missing | mismatched | extra |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| candidate | 1.0.0-rc.9 | 1047 | 1047 | 0 | 0 | 0 |
| `SPEC_PIN` | 1.0.0-rc.9 | 691 | 691 | 0 | 0 | 0 |

Each doctored root is the verified candidate with exactly one artefact removed — confirmed by
`diff -rq` against the verified copy, and by re-running the manifest verifier on the doctored root,
which reports `missing 1` naming that artefact and nothing else.

---

## 7. Gate table — every command its own process, real exit code

Run on the delivered tree, after the mutant harness had restored it.

| command | exit | note |
| --- | ---: | --- |
| `go build ./...` | **0** | |
| `go vet ./...` | **0** | |
| `gofmt -l cmd internal` | **0** | 0 files listed |
| `golangci-lint run ./...` | **0** | 0 issues |
| `bash .github/ci/no-broad-suppression.sh` | **0** | |
| `bash .github/ci/ledger-consistency.sh` | **0** | 231 rows across linux darwin windows |
| `bash .github/ci/gate-selftest.sh` | **0** | **130 passed, 0 failed** (was 125 at `31106aa8`) |
| `go test ./internal/interop/environments/` (candidate root) | **0** | 12 cases, 0 skip, 0 fail |
| `test-gate.sh` — candidate lane, intact root | **0** | see §8 |
| `test-gate.sh` — candidate lane, less `vectors/environments.json` | **1** | fails by name at the plan |
| `test-gate.sh` — candidate lane, less `expected/byte-exact-snapshot_sha256.txt` | **1** | fails by name at the plan |
| `test-gate.sh` — default lane, `SPEC_PIN` root | **0** | 33 skips, all `root-unset` ones in deferred packages |
| `gate-selftest.sh` at the intermediate commit `5b3d69fc` | **0** | 128 passed, 0 failed — the history is green at every commit |

The `gate-selftest.sh` tally moved 125 → 130: −2 (the two hand-appended trees left `ENV_REQUIRED`),
+4 (the holing loop now covers 7 declared artefacts instead of 6, at 4 assertions each), +3 (the new
F1b cases).

---

## 8. One red run, named

The **first** execution of the candidate lane on the intact root exited **1**, on
`internal/install :: TestStrictRegistryPolicyFailsUnknown` — not on anything this change touches:

```
advisory must pass unknown: {… Messages:[… registry test-reg snapshot timestamp is too far
in the future] Errors:[every trusted audit registry served a tampered snapshot] …}
```

What is established:

- the diff touches **no** file in `internal/install`, `internal/registry` or `internal/config`
  (`git diff 31106aa8 -- internal/install cmd` is empty), and `internal/interop/environments`
  contains only `_test.go` files, so nothing in it is linked into another package's binary;
- the assertion is `parsed.CreatedAt.After(now.Add(clockSkew))` in `internal/registry/snapshot.go`,
  comparing a `time.Now()` captured in `resolveRegistries` against a `time.Now()` stamped later by
  the fixture's own HTTP handler — a sub-second timing relation;
- standalone it does not reproduce: 5/5 pass on this tree, 5/5 pass and 40/40 pass under `-count=40`
  on a pristine checkout of `31106aa8`, and 3/3 full-package runs pass on that pristine checkout;
- **re-running the same lane on the same intact root exits 0**, with that case recorded `pass`.

What is **not** established: I did not reproduce the failure on the pristine `31106aa8` checkout, so
I cannot call it "pre-existing" from measurement — only that it is load-dependent, that it did not
recur, and that it lies outside everything this diff can influence. Reported as unknown rather than
inferred from the proxy.

---

## 9. AC coverage — 3 of 3 rows driven, with the production call site named

| AC row | driven by | production call site |
| --- | --- | --- |
| a candidate root missing any environments vector family fails the candidate lane **by name** | `test-gate.sh` → `suite-plan.sh`, two negative runs, exit 1 each | `.github/ci/suite-plan.sh` root-coverage partition, `CI_REQUIRE_FULL_ROOT` branch |
| the default lane keeps every non-environments `internal/interop` case | `test-gate.sh` on the `SPEC_PIN` root, exit 0, `internal/interop` served | `.github/ci/suite-plan.sh` + `platform-case-gate.sh` Tier 1 |
| `gate-selftest.sh` green | run as its own process, exit 0, 130 passed / 0 failed | `.github/ci/gate-selftest.sh` |

Rework-specific coverage: **4 of 4** findings driven — F1 by M1–M4 plus the corrected note, F1b by
three new `gate-selftest` cases plus narrowing mutant M9, F2 by M5–M8/M10/M11 plus the before/after
plan measurement, F3 by the reworded row.

---

## 10. Bounds

- Executed on `GOOS=darwin` only. linux/windows behaviour is `go list`-derived through
  `ledger-consistency.sh` (exit 0, 231 rows across all three), not executed here.
- The static skip scan cannot see a skip reached through reflection or through a helper elsewhere
  whose name does not begin with `Skip`; that residue is covered behaviourally by
  `platform-case-gate.sh`, which is fatal on an unrecognised reason (§1).
- `TestConformanceEveryPathTheVectorsNameIsDeclared` needs a served root, so the
  over-declaration/cross-reference direction is asserted in the candidate lane and not on the pin
  lane, where the package is deferred. The static half runs on both.
- `rootPath` bounds the class it rejects to paths no declared artefact covers; M11 measures that
  bound rather than asserting it.
- The Tier-1 `tol` display line for a fatally-skipped case is unchanged (§2); the exit code is right,
  the display is not, and the quirk predates this change.
- The candidate lane has still not run on hosted CI — it is dispatch-only. A candidate dispatch
  remains the outstanding proof; removing `vectors/environments.json` from the dispatched root should
  produce `suite-plan` exit 1 with `missing: vectors/environments.json` before any `go test` starts.
- Scratch, roots, logs and the mutant harness are under the worktree's
  `.temp/TASK-260906-2cfxfv/rework1/`; nothing there is committed.
