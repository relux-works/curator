# Review findings, cycle 2 — interop environments coverage

**Verdict: ACCEPT.** PR #63 (curator `e3a97d7b`) is safe to land.

`repeat-of:` cycle-1 F1 (a guard whose ledger note claims a stronger property
than the scan enforces) — recurs as **F4 below, minor**, on a *different* guard
that did not exist at cycle 1 (`TestNoRootReadHereEscapesTheDeclaredArtefactGuard`).
F1, F1b, F2 and F3 themselves are fixed and verified fixed. Nothing else recurs.

- Subject: branch `feat/interop-root-artifacts` @ `e3a97d7b98688d42acae7e04fa064441d222a5cc`,
  4 commits past `origin/main` (`fb916acd`), all four `git log %G? = G`, author
  Ivan Oparin <oparin@me.com>. `origin/main` **is** an ancestor of the head — a
  fast-forward landing needs no rebase and no re-review.
- Everything below was driven, not read. Scratch lived in a throwaway
  `--shared` clone at `curator-spec/.temp/STORY-260905-1n0iy8/review2/`; the
  producer's worktree and the control root were never written to (verified: the
  real repo's `git worktree list` carries no scratch entry, and the throwaway
  checkout ended every mutant at `git status --porcelain` = 0 lines).

---

## 1. The acceptance criterion, reproduced from scratch

Both roots materialized as **plain checkouts** (`git clone` + `git checkout`,
never `git archive`) and verified file-by-file against `manifest.json` with my
own verifier:

| root | commit | manifest entries | verified | missing | mismatched | extra |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| candidate | curator-spec `87a0d006` | 1047 | 1047 | 0 | 0 | 0 |
| SPEC_PIN | curator-spec `0ed5c691` | 691 | 691 | 0 | 0 | 0 |

**Seven negative runs** — the candidate root with exactly one declared artefact
removed, `CI_REQUIRE_FULL_ROOT=1 suite-plan.sh`:

| artefact removed | exit | package served? | plan line |
| --- | ---: | --- | --- |
| `vectors/environments.json` | 1 | no | `missing: vectors/environments.json` |
| `vectors/context-versions.json` | 1 | no | `missing: vectors/context-versions.json` |
| `vectors/context-detectors.json` | 1 | no | `missing: vectors/context-detectors.json` |
| `vectors/snapshot-acquisition.json` | 1 | no | `missing: vectors/snapshot-acquisition.json` |
| `expected/environments` | 1 | no | `missing: expected/environments` |
| `expected/byte-exact-snapshot_sha256.txt` | 1 | no | `missing: expected/byte-exact-snapshot_sha256.txt` |
| `fixtures/byte-exact` | 1 | no | `missing: fixtures/byte-exact` |

**Positive run**: unmodified candidate root, `CI_REQUIRE_FULL_ROOT=1` →
`suite-plan` exit 0, both `internal/interop` and `internal/interop/environments`
served, `go test` both packages `ok`.

**Two-root partition, truthful on both:**

| root | served | deferred | deferred packages | `CI_REQUIRE_FULL_ROOT=1` |
| --- | ---: | ---: | --- | --- |
| SPEC_PIN | 69 | 4 | `internal/config internal/envfragment internal/envmarker internal/interop/environments` | exit 1, fails by name |
| candidate | 73 | 0 | — | exit 0 |

`internal/interop` itself is served on **both** roots — it is never dragged into
the deferral.

## 2. The split's blast radius — nothing lost, measured by name

Served case lists compared before (`fb916acd`) and after (`e3a97d7b`) against
the **candidate** root:

| | base `./internal/interop` | head `./internal/interop` | head `./internal/interop/environments` |
| --- | ---: | ---: | ---: |
| outcomes | 74 | 25 | 55 |

49 cases moved; **0 lost** (`comm` of base against the union of both head
packages is empty). The 6 extra in the new package are its contract cases.

On the **SPEC_PIN** root — the default lane — base `./internal/interop` was
`25 PASS + 6 SKIP`; head is `25 PASS + 0 SKIP`. The six that disappeared are
exactly the six that were *skipping*: the silent-skip guards this leaf exists to
delete. **No passing case was lost from the default lane.**

`internal/interop/golden_test.go`'s change is comment-only (a package doc block);
it cannot weaken what the pinned Go manager consumes. See F5 for the cross-repo
companion.

## 3. F1 — every skip receiver shape

Each shape injected into the real `TestConformanceContextDetectors`, compiled,
and asked two independent questions: does the case *really* skip on a fully
serving root, and does the scan kill it. `=== RUN` counted at the anchored level.

| # | shape | compiles | case really | scan |
| --- | --- | --- | --- | --- |
| M1 | `t.Skipf(…)` | yes | SKIP (run=1) | killed |
| M2 | `guard := t; guard.Skipf(…)` | yes | SKIP (run=1) | killed |
| M3 | `h := struct{t *testing.T}{t: t}; h.t.Skipf(…)` | yes | SKIP (run=1) | killed |
| M4 | `skip := t.Skipf; skip(…)` (method value) | yes | SKIP (run=1) | killed |
| M5 | `var tb testing.TB = t; tb.Skipf(…)` (interface) | yes | SKIP (run=1) | killed |
| M6 | `(*testing.T).Skip(t, …)` (method expression) | yes | SKIP (run=1) | killed |
| M7 | skip inside a closure | yes | SKIP (run=1) | killed |
| M8 | helper in another **file** of the same package | yes | SKIP (run=1) | killed |
| M9 | helper in another **package**, not named `Skip*` | yes | SKIP (run=1) | **SURVIVED — documented bound** |

All four of my predecessor's shapes are killed, and four shapes nobody had tried
are killed too. M9 is the bound `skipCall`'s successor `skipSelector` names in
its own doc, and the note is honest about it — so I drove the second layer the
note points at, with four different skip wordings, through **both** gate versions:

| M9 skip reason | class | new gate | old gate |
| --- | --- | --- | --- |
| `CURATOR_CONFORMANCE_ROOT is not set` | `root-unset` | **FATAL-served-package** | `tolerated-by-ledger` |
| `…publishes no vectors/context-detectors.json` | `root-content` | FATAL-wrong-class | FATAL-wrong-class |
| `symlink unavailable: nope` | `host-capability` | FATAL-wrong-class | FATAL-wrong-class |
| `because I felt like it` | UNCLASSIFIED | FATAL-wrong-class | FATAL-wrong-class |

Row 1 is the finding of this review that most justifies the change: the one
wording that evades the static scan *and* would have been tolerated by the old
gate is closed **precisely by the F1b fix**. F1 and F1b genuinely compose; the
orchestrator's decision to fix F1b here was right, and it is load-bearing rather
than defence-in-depth.

## 4. F1b — the gate change and its blast radius

**Replay, re-measured myself.** Same ledger and class table (HEAD's), old gate
(`origin/main`) vs new gate, over both recorded lane streams:

| lane | skips | `skips-observed.tsv` old vs new | verdict tally (new) |
| --- | ---: | --- | --- |
| `lane-pin` | 32 | **byte-identical** | 16 tolerated-by-ledger, 7 host-capability, 7 opt-in, 1 root-content, 1 helper-process |
| `lane-candidate` | 20 | **byte-identical** | 4 tolerated-by-ledger, 7 host-capability, 7 opt-in, 1 root-content, 1 helper-process |

The only textual difference in `platform-cases.txt` is the report's own output
path. Static check independently: of 231 ledger rows, 70 tolerate a skip
(31 host-capability, 20 root-unset, 16 platform-control, 2 root-content,
1 opt-in) and `root-unset` is the **only** class whose policy is `deferred-only`.
13 of the 20 root-unset rows predate this change.

**Attacked with real shipped ledger rows**, not synthetic ones:

| probe | row | package | old gate | new gate | right? |
| --- | --- | --- | --- | --- | --- |
| A | `internal/envfragment :: TestFragmentAuthoritativeSchemaCases` (root-unset) | **served** | `tolerated-by-ledger` | **`FATAL-served-package`** | yes — must fire |
| B | same row | deferred | `tolerated-by-ledger` | `tolerated-by-ledger` | yes — must not fire |
| C | `internal/godriver :: TestCandidateGoV1SourceAwareContract` (root-content, policy `allow`) | served | `tolerated-by-ledger` | `tolerated-by-ledger` | yes — must not fire |
| D | host-capability tolerated skip | served | `allowed-host-capability` | `allowed-host-capability` | yes — must not fire |

The change narrows exactly one class and nothing else.

## 5. The three new `gate-selftest.sh` cases, and the holing loop

Each mutant leaves the gate **present** and weakens it (or, G1, deletes it as a
control). Every one names a specific assertion that must go red.

| # | mutant | shape | assertion that must fail | result |
| --- | --- | --- | --- | --- |
| G1 | delete the new `else if` branch | DELETE (control) | `a LEDGER-TOLERATED root-unset skip in a SERVED package fails` | FAIL ✓ |
| G2 | branch fires only when `want == "-"` — admits every real root-unset row | **NARROW** | same | FAIL ✓ |
| G3 | branch drops the `!(pkg in isdeferred)` exemption | NARROW | `…passes once the package is deferred` | FAIL ✓ |
| G4 | policy check applied to every class, not just `deferred-only` | WIDEN | `an allow-policy tolerated skip survives beside a deferred package` | FAIL ✓ |
| G5 | `suite-plan.sh` stops checking `expected/*` artefacts | **NARROW** | `a candidate root missing expected/environments fails the lane` + `the failure names expected/environments` | FAIL ✓ |
| G6 | drop a **required** family from the `root-artifacts.tsv` row | **NARROW** | `root-artifacts.tsv declares vectors/environments.json for internal/interop/environments` | FAIL ✓ |
| D3 | drop `expected/byte-exact-snapshot_sha256.txt` from the row, root intact | **NARROW** | `TestConformanceEveryPathTheVectorsNameIsDeclared` **and** the real consumer `TestConformanceSnapshotAcquisition` | both FAIL ✓ |

G6 confirms `ENV_REQUIRED` is genuinely non-self-referential: it is derived from
the package's `requireFamily` calls, so shrinking the row cannot shrink the check
that guards the row. No new self-test case is decoration.

The holing loop runs **all seven** declared artefacts on every PR on all three
platforms — the AC's negative evidence is now a permanent gate, not a one-off log.

## 6. F2 — the derivation

Independently enumerated every string the four vectors carry that the candidate
root resolves as a path: **26**, matching my predecessor's count exactly —
24 under `expected/environments`, 1 `fixtures/byte-exact`, and
`expected/byte-exact-snapshot_sha256.txt`, the one that was forgotten. All 26 are
covered, and no declared artefact is a superset.

Attacked in both directions:

| probe | change | result |
| --- | --- | --- |
| D1 | a vector names a real root path no artefact covers | FAIL: *"vectors/environments.json names expected/marker.json … declares no artefact covering it"* |
| D2 | the root drops `expected/byte-exact-snapshot_sha256.txt` | FAIL: *"declares … but no case requires it and no vector names a path under it"* |
| D3 | the row is narrowed by one artefact | derivation **and** `rootPath`'s refusal both fire |

The "short set" question: the total `crossReferenced == 0` fatal is only a total
guard, but the **reverse** direction is per-artefact — a family that stops naming
paths leaves its artefact covering nothing and fails loudly (D2 is exactly that
shape). The row, `rootPath` and `ENV_REQUIRED` no longer drift independently:
D3 fires in two places at once and `ENV_REQUIRED`'s hand-kept cross-reference
list is gone.

## 7. Findings

### F4 (minor) — one ledger note still claims more than its scan enforces

`TestNoRootReadHereEscapesTheDeclaredArtefactGuard` tracks the conformance root
by the **identifier `root`**. A case that reaches the same root by a different
door is invisible to it. Injected into the real `TestConformanceContextDetectors`:

```go
sneaky := os.Getenv("CURATOR_CONFORMANCE_ROOT")
if b, err := os.ReadFile(filepath.Join(sneaky, "vectors/agent-context-v1.json")); err == nil { … }
```

compiles, reads an **undeclared** root path, and **all four** contract scans pass:
`TestNoRootReadHereEscapesTheDeclaredArtefactGuard`,
`TestTheRootIsAlwaysBoundToThatOneName`,
`TestConformanceEveryPathTheVectorsNameIsDeclared` and
`TestNoCaseHereSkipsForAnythingButTheDeferredRoot` all SURVIVED, and the case
itself PASSED.

The overstatement is the `so` clause in the ledger note — *"…so every root path
this package reads goes through rootPath"* — and the same claim in
`contract_test.go:228`, *"keeps `rootPath` the only way this package turns the
conformance root into a path on disk."* The scan enforces a syntactic proxy
(the identifier) and the note asserts the semantic property.

**Why this is minor and not a repeat at F1's severity.** F1's shapes were
innocent refactors (renaming `t`) and produced the failure class this leaf
exists to close — a *green* lane with the cases skipped. This gap requires
deliberately bypassing the package's only documented entry point, and its worst
outcome is an undeclared read that goes **red** with an `open()` error mid-case
rather than by name at the plan — the pre-F2 state, which cycle 1 itself scored
minor. It cannot produce a silent skip.

**What to do, next time this file is touched:** either state the bound the way
`skipSelector` states its own (name the identifier as the tracked proxy), or add
`os.Getenv("CURATOR_CONFORMANCE_ROOT")` outside `suite_test.go` to the scan —
one `ast.Inspect` clause. Do not leave the `so` clause standing.

### F5 (companion, cross-repo — not a defect in this change)

curator-spec's Implementations lane exports
`CURATOR_CONFORMANCE_ROOT: ${{ github.workspace }}/conformance/v1` and runs an
**explicit package list** that includes `./internal/interop` but not
`./internal/interop/environments`. Its coverage ledger
(`.github/ci/implementation-coverage.tsv`) declares **no** `internal/interop`
rows, so nothing there would fail by name.

Measured: the pinned Go commit `a3abcf34` carries only
`internal/interop/golden_test.go` — the environments cases do not exist at that
pin, so **nothing is lost today and this does not block landing**. But the moment
that pin advances past `e3a97d7b`, those 49 cases stop running in the
Implementations lane, silently. That is this leaf's own failure class, one
repository over.

**Action:** when the Go pin is next bumped, add `./internal/interop/environments`
to the invocation list in `.github/workflows/implementations.yml`, or give it
coverage-ledger rows so a dropped package fails by name.

### F6 (nit) — the gate reports the same case two ways

A case killed by the new `FATAL-served-package` branch is still printed as
`tol   pkg :: case (tolerated skip: root-unset)` by the Tier-1 loop further down
the report. Observed together in one report on probe A. The gate still exits 1
(`fail` is sticky), so this is cosmetic — but a reader scanning the Tier-1
section sees "tolerated" for a case the gate just declared fatal.

### F3 (cycle 1) — fixed

The row note now says *"the contract cases beside them read no root artefact at
all"*, and `TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass` enforces
`skip=- class=-` for exactly those. Accurate.

## 8. Gates, observed exit codes on `e3a97d7b`

| gate | exit | note |
| --- | ---: | --- |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `gofmt -l cmd internal` | 0 | no output |
| `golangci-lint run ./...` | 0 | `0 issues.` |
| `bash .github/ci/gate-selftest.sh` | 0 | **130 passed, 0 failed** (~90s) |
| `bash .github/ci/no-broad-suppression.sh` | 0 | `no-broad-suppression: ok` |
| `bash .github/ci/ledger-consistency.sh` | 0 | **231 rows** across linux/darwin/windows; all 12 rows for the new package `ok` |

Hosted, PR #63 @ `e3a97d7b`: Gate self-test **pass** on ubuntu/macos/windows,
Interop conformance gate **pass**, Lint **pass**, Naming gate **pass**, Test
**pass** on ubuntu + macos, Race **pass** on ubuntu + macos. `Test
(windows-latest)` was still **pending** at the time of writing; **no lane is
red**. `Candidate suite` is `skipping` (workflow-dispatch only).

## 9. Ledger rows, before and after

| | before (`origin/main`) | after |
| --- | ---: | ---: |
| `internal/interop` conformance rows carrying a `root-content` toleration | 6 | 0 |
| `internal/interop/environments` rows | 0 | 12 (6 conformance `root-unset`, 6 contract `skip=- class=-`) |
| rows tolerating a `root-unset` skip, repo-wide | 13 | 20 |
| total ledger rows | 225 | 231 |

The class move `root-content` → `root-unset` is what the whole leaf turns on:
`root-content` is policy `allow` in **every** lane, `root-unset` is
`deferred-only`. Verified independently against `skip-classes.tsv`:
`root-unset` is the only `deferred-only` class in the table.

## 10. Is a candidate dispatch needed?

**No.** The seven negatives reproduce from scratch on roots I materialized and
manifest-verified myself, and `gate-selftest.sh`'s holing loop reruns the same
seven on all three hosted platforms on every PR. A dispatch would add a real
end-to-end green candidate lane, which is nice to have but proves nothing the
above does not. If one is run anyway, remove `expected/byte-exact-snapshot_sha256.txt`
— it is the artefact the previous cycle found missing from the declaration, so it
exercises the F2 fix and the holing loop at once.

## 11. Landing

**PR #63 is safe to land, without hedging.** `origin/main` (`fb916acd`) is an
ancestor of the reviewed head, so the exact reviewed, signed objects
fast-forward; no rebase, no re-review. Four commits, all good signatures, human
identity, `LOGBOOK.md` untouched, nothing under `.temp/` committed.

F4 is a wording correction to carry on the next touch of `contract_test.go`.
F5 is a curator-spec companion for whenever the Go pin advances. Neither blocks.
