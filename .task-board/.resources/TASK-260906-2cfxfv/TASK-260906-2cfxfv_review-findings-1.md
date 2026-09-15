# TASK-260906-2cfxfv — review findings, cycle 1

**Verdict: CHANGES REQUESTED.** Two findings; one major, one minor, one nit.
`repeat-of: none` (first review cycle for this leaf).

Subject: branch `feat/interop-root-artifacts` at `31106aa8`, worktree
`/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage`, two signed commits
(`G`, `Ivan Oparin <oparin@me.com>`) past `origin/main` `fb916acd`, working tree clean,
PR https://github.com/relux-works/curator/pull/63. 11 files, +560/−64.

Everything below was executed by this reviewer in a throwaway copy under `/tmp/rev-2cfxfv/`.
Nothing was written into the producer's worktree or into either control repository
(`git status --short` clean in the producer worktree at the end of this review).

---

## 0. The Change Request's empty repository delta

`CR-TASK-260906-2cfxfv-1` reports `repository_delta=empty`, and the candidate tree OID
`c3e47989…` equals its base `f39f4a93…`. **That is a cross-repository artefact of the CR
snapshot, not a producer failure.**

The CR measures the story worktree `curator-spec/.temp/STORY-260905-1n0iy8/worktree`, which is a
checkout of the **curator-spec** repository and has no `internal/` tree at all. The leaf's Scope
field, both producer briefs and the review brief all name the **curator** repository, branch
`feat/interop-root-artifacts` in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-interop-coverage`.
The work exists there as two signed commits and is what this review examined. The producer said so
explicitly in §14 of its report.

This has an orchestration consequence: **a rework run spawned into the curator-spec story worktree
will find nothing to fix.** The next producer run must be pointed at the curator worktree.

---

## 1. What was verified, first-hand

### 1.1 The hole was real (measured at base, not asserted)

Same doctored root, both trees:

| tree | root | `suite-plan.sh` + `CI_REQUIRE_FULL_ROOT=1` | environments cases |
| --- | --- | ---: | --- |
| base `fb916acd` | candidate less `vectors/environments.json` | **exit 0**, `served=72 deferred=0`, `ok` | `--- SKIP` ×2, reason `publishes no vectors/environments.json (… root-content)` |
| head `31106aa8` | same root | **exit 1**, `FAIL internal/interop/environments was deferred`, `missing: vectors/environments.json` | never reached — the lane refuses to run |

### 1.2 Six negative candidate-lane runs, all reproduced from scratch

`CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh` against the verified candidate root with
exactly one declared artefact removed. Not read from the producer's logs — re-run here.

| removed artefact | `test-gate.sh` exit | report line |
| --- | ---: | --- |
| `vectors/environments.json` | **1** | `missing: vectors/environments.json` |
| `vectors/context-versions.json` | **1** | `missing: vectors/context-versions.json` |
| `vectors/context-detectors.json` | **1** | `missing: vectors/context-detectors.json` |
| `vectors/snapshot-acquisition.json` | **1** | `missing: vectors/snapshot-acquisition.json` |
| `expected/environments` | **1** | `missing: expected/environments` |
| `fixtures/byte-exact` | **1** | `missing: fixtures/byte-exact` |

Every one also printed `FAIL  internal/interop/environments was deferred, but this lane requires a
root that serves the whole module` and `suite-plan: FAILED`. The AC's "fails by name rather than
skipping" is met, six times over, not twice.

### 1.3 Positive, same lane, unmodified candidate root

`suite-plan.sh` + `CI_REQUIRE_FULL_ROOT=1` → **exit 0**, `served=73 deferred=0 excluded=0`.
`go test ./internal/interop/environments/` on that root → **exit 0**, 9 top-level cases,
`pass=9 skip=0 fail=0`; `platform-case-gate.sh` on the real `-json` stream reports all nine `ok`
and `skips-observed.tsv` has **0 rows**.

### 1.4 Roots materialized as plain checkouts, verified against `manifest.json`

Independent verifier (`/tmp/rev-2cfxfv/verify.py`, sha256 per manifest entry plus an undeclared-extras
walk), not the producer's script:

| root | revision | protocol_version | declared | verified | missing | mismatched | extra |
| --- | --- | --- | ---: | ---: | ---: | ---: | ---: |
| candidate | curator-spec `87a0d006` | 1.0.0-rc.9 | 1047 | 1047 | 0 | 0 | 0 |
| SPEC_PIN | curator-spec `0ed5c691` | 1.0.0-rc.9 | 691 | 691 | 0 | 0 | 0 |

`git archive` was not used. Canary confirmed: `conformance/v1/fixtures/byte-exact/subst.txt` is
40 bytes in the checkout, and `git check-attr` shows `export-subst: set` on it.

### 1.5 The split's blast radius — nothing lost

Top-level `Test*` functions, base `internal/interop` vs. head `internal/interop` ∪
`internal/interop/environments`: 16 → 19. **Removed: none.** Added: the three contract cases.

Two-root partition:

| root | tree | served | deferred | which |
| --- | --- | ---: | ---: | --- |
| SPEC_PIN | base | 69 | 3 | `config`, `envfragment`, `envmarker` |
| SPEC_PIN | head | **69** | 4 | the same three **+ `internal/interop/environments`** |
| candidate | head | 73 | 0 | — |

The served set on the default lane is unchanged at 69; the only new deferral is the new package.
Behaviourally on the pin: `internal/interop` **served**, `exit 0`, **10 PASS / 0 SKIP / 0 FAIL**
(`TestCanonicalJSONVectors`, `TestGoldenContextCopy`, `TestGoldenFederationSemantics`,
`TestGoldenMarkerObject`, `TestGoldenRegistryObjects`, `TestGoldenSnapshotHash`,
`TestIdentifierVectors`, `TestLocaleSelectorVectors`, `TestManagerConfigVectors`,
`TestSourceIdentityVectors`). `internal/interop/golden_test.go`'s only change is a doc comment —
zero behavioural risk to the pinned Go manager that curator-spec's Implementations lane runs
against that package.

### 1.6 Mutants — this reviewer's own, all narrowing, all restored

| mutant | narrows the gate to | check | exit | verdict |
| --- | --- | --- | ---: | --- |
| **A** drop `vectors/context-detectors.json` from the row | admits a root missing that one family | `TestEveryFamily…` | 1 | **killed** |
| A | | `gate-selftest.sh` | 1 | **killed** (`120 passed, 1 failed`) |
| A | | *plan*: root less `context-detectors` | 0 | SURVIVED — the one admitted member |
| A | | *plan*: root less `environments.json` | 1 | **killed** — the narrowing admits exactly one |
| A | | *behavioural*: that case on that root | 1 | **killed** — `--- FAIL`, `serves this package but does not publish` — **red, not skip** |
| **B** over-declare `expected/byte-exact-snapshot_sha256.txt` | a declared artefact nothing reads | `TestEveryFamily…` | 1 | **killed** (`declared … but nothing here reads it`) |
| **D** flip one ledger row `root-unset`→`root-content` | admits one `root-content` skip | `TestTheLedgerToleration…` | 1 | **killed** |
| D | | *behavioural*: deferred stage → `platform-case-gate.sh` | 1 | **killed**, verdict `FATAL-wrong-class` |
| **E** register `internal/interop` wholesale (the rejected one-liner) | defers the pre-environments cases | `gate-selftest.sh` | 1 | **killed** (`123 passed, 2 failed`; both new assertions fire) |
| **G** add `requireFamily(…, "vectors/manager-config-v2.json")` — an undeclared family | a case reading an undeclared artefact | `TestEveryFamily…` | 1 | **killed** |
| G | | `gate-selftest.sh` | 1 | **killed** — proves the derived required-set is genuinely non-self-referential |

Degenerate-root probe (the report's own stated bound): an artefact **present but empty** — a
zero-byte `vectors/environments.json`, and an emptied `expected/environments` directory — is
`served` by the plan (`-e` passes) and then goes **RED** in `go test` (exit 1) rather than green.
The stated bound is honest and the fallback is loud.

### 1.7 Gates, observed by this reviewer on a clean copy of `31106aa8`

| command | exit | note |
| --- | ---: | --- |
| `go build ./...` | 0 | |
| `go vet ./...` | 0 | |
| `gofmt -l cmd internal` | 0 | 0 files |
| `bash .github/ci/no-broad-suppression.sh` | 0 | |
| `bash .github/ci/ledger-consistency.sh` | 0 | 228 rows across linux darwin windows |
| `bash .github/ci/gate-selftest.sh` | 0 | **125 passed, 0 failed** (63s) |
| 3 contract cases, each `-run` anchored separately | 0 | `RUN=1 PASS=1` each — no filter matched nothing |

### 1.8 Hosted lanes (PR #63, head `31106aa8`)

**No red lane.** `Test (ubuntu)` pass, `Test (macos)` pass, `Race (ubuntu)` pass,
`Gate self-test` pass ×3, `Interop conformance gate` pass, `Lint` pass, `Naming gate` pass.
Still pending at review time: `Race (macos-latest)`, `Test (windows-latest)`.

`Candidate suite (${{ matrix.os }})` reports **`skipping`** — it is dispatch-only. **The lane this
leaf's entire gate lives in has not run hosted.** A candidate dispatch is needed to prove this leaf
on CI rather than only on this host. Remove **`vectors/environments.json`** from the dispatched
candidate root: it is the family the task description names, it is read by two ledger cases, and the
expected observation is `suite-plan` exit 1 with `missing: vectors/environments.json` before any
`go test` starts.

---

## 2. Findings

### F1 — MAJOR: `TestNoCaseHereSkipsForAnythingButTheDeferredRoot` can be satisfied while a silent skip exists

`internal/interop/environments/contract_test.go:63` (`skipCall`) matches a skip only when the call's
receiver is a bare `*ast.Ident`:

```go
ident, isIdent := sel.X.(*ast.Ident)
if !isIdent { return "", "", false }
```

Two receiver forms therefore evade the scan entirely. Measured, each injected into
`TestConformanceContextDetectors` and each compiling and really skipping:

| injected skip | AST gate | `t.Skip` visible to `grep` | case really skips on a fully serving root |
| --- | --- | ---: | --- |
| `t.Skipf(…)` — plain | exit 1 **KILLED** | 1 | yes |
| `guard := t; guard.Skipf(…)` — renamed receiver | exit 1 **KILLED** | 0 | yes |
| `h := holder{t: t}; h.t.Skipf(…)` — **selector chain** | exit 0 **SURVIVED** | 1 | **yes** |
| `skip := t.Skipf; skip(…)` — **method value** | exit 0 **SURVIVED** | 1 | **yes** |

The review brief's criterion for this dimension was: *"If it can be satisfied while a silent skip
exists, it is decoration."* It can.

**The claim is also committed, and it is wrong.** `platform-cases.tsv`, the note for this very row,
states: *"the syntax tree of this package contains exactly one Skip call, inside suiteRoot; any
other Skip — **under any receiver name** — would classify as root-content"*. The drafting report §8
repeats it (*"Any other skip, under any receiver name"*). The gate does not see two receiver forms.
Under the DoD row *"Ledger rows describe what their tests actually assert"*, this note does not.

The producer's mutant **M4** is cited as the source-text-gate attack, and it is a good mutant — but
it exercised only the `guard := t` rename, which the gate **does** catch. It is therefore evidence
for a narrower claim than the one it is cited for.

**F1b — the second layer does not fully cover the gap.** I drove the real
`.github/ci/platform-case-gate.sh` over a real `go test -json` stream with the package **served**
(`CI_DEFERRED_PKGS=''`), for each reason shape:

| skip reason shape | class | platform-case-gate verdict |
| --- | --- | --- |
| `…publishes no vectors/context-detectors.json` | `root-content` | **FATAL-wrong-class** — killed |
| `detectors temporarily disabled` (unrecognised) | `UNCLASSIFIED` | **FATAL-wrong-class** — killed |
| `CURATOR_CONFORMANCE_ROOT is not set` | `root-unset` | **`tolerated-by-ledger`** — **not fatal** |

Cause: in `platform-case-gate.sh` the `deferred-only` policy check sits in the `else` branch *after*
`if (tol != "")`, so for a case the ledger lists **with** a tolerated skip the policy is never
reached. A `root-unset`-worded skip in a package this lane **serves** is therefore tolerated.

To be fair to the change: `platform-case-gate.sh` is **not touched by this diff**, the two most
plausible accidental reason shapes are killed there, and no production path today reaches
`suiteRoot` with the variable empty while the package is served. This is a defence-depth gap, not a
live regression. But F1 + F1b compose into exactly the silent-skip class this leaf exists to close,
inside the package it created to close it, and the recurrence guard that is supposed to prevent it
reports green.

**Fix, and it is small:** in `skipCall`, stop requiring `sel.X` to be an `*ast.Ident` — accept any
expression as the receiver and report it via `types`/position instead of a name — and additionally
flag a `Skip`/`Skipf`/`SkipNow` **selector used as a value** (assigned, passed, returned) rather than
called. Then correct the `platform-cases.tsv` note and report §8 to say what the scan actually
covers, and add the two surviving shapes above as mutants that must fail. If the `root-unset`
tolerance in `platform-case-gate.sh` is judged out of scope for this leaf, say so as a stated bound
in the report rather than leaving it implied — but do not leave the ledger note claiming coverage
the code does not have.

### F2 — MINOR: the declaration omits one root path the package reads unguarded

`snapshot_acquisition_test.go:75` reads `tc.Expected` through `readRootFile` — an unguarded
`os.ReadFile` + `t.Fatal`. On the real candidate root every `tc.Expected` resolves to
**`expected/byte-exact-snapshot_sha256.txt`**, which is:

- **not** declared in the `internal/interop/environments` row of `root-artifacts.tsv`;
- **not** in `contract_test.go`'s `crossReferencedTrees`;
- **not** in `gate-selftest.sh`'s hand-appended `ENV_REQUIRED="$ENV_REQUIRED expected/environments fixtures/byte-exact"`.

I enumerated every root-relative path the four declared vectors name and that resolves in the root:
26 paths — 24 under `expected/environments`, 1 under `fixtures/byte-exact`, and this one, uncovered.

**The hole did not move**: a root dropping it is `served`, and the case then fails **RED** with the
open() error naming the file — the second of the two loud shapes the rewritten `root-artifacts.tsv`
header names. But it fails mid-case rather than by name at the plan, inconsistently with
`expected/environments` and `fixtures/byte-exact`, which are declared precisely so that they do
*not* fail that way — the header comment and `crossReferencedTrees`' own comment both give that as
the reason for declaring them.

Note the fix is **not** a one-line `.tsv` edit: mutant B measured that adding this path to the row
alone makes `TestEveryFamily…` fail with *"declared … but nothing here reads it"*. Both the row and
`crossReferencedTrees` (and `gate-selftest.sh`'s `ENV_REQUIRED`) must move together — or, better,
the guard should derive the cross-referenced set from the vectors' own path fields, or at minimum
also scan `readRootFile(t, root, …)` literals, so a future unguarded root read cannot drift
undeclared the same way.

### F3 — NIT: the `root-artifacts.tsv` row note overstates

The row note says *"every case in the package reads one of them without a per-family guard"*. Three
of the nine cases in the package — the contract cases — read **no** root artefact at all; the ledger
correctly gives them `skip=-  class=-` for exactly that reason.

---

## 3. What is not a finding

- The six conformance ledger notes accurately describe what their tests assert (`contextaudit.Detect`
  findings/severities/spans/waivers; `pkgversion` parse/order/range plus rejection of invalid
  spellings; `contextresolve.Resolve` + `contextlock` canonical bytes and lock hash;
  `contextmaterialize.EmittedOrder`/`Header` and `header_type_line`; byte-for-byte
  `expected/environments` with surface hash; `gitops.Extract` re-extraction against
  `expected_sha256`). Verified against the test bodies.
- The class move `root-content` → `root-unset` is right: `root-content` is policy `allow` in every
  lane; `root-unset` is `deferred-only` and is now legitimate because `suite-plan.sh` genuinely
  defers this package. Mutant D confirms the flip back is killed statically and behaviourally.
- `requireFamily` fails on **any** read error, not only `os.ErrNotExist` — a failed read is not
  treated as a legitimate absence.
- The chosen shape (split the package) is justified against the alternatives, and mutant E measures
  the cost of the rejected wholesale registration rather than asserting it: `internal/interop`
  becomes deferred on the pin and the two new `gate-selftest.sh` assertions fire.
- The report's §10 "one red run worth naming" (the `GO_TEST_TIMEOUT=8m` `internal/install` timeout)
  is correctly diagnosed as a budget artefact; `test-gate.sh`'s default and CI's unix budget are
  both 30m.

## 4. Bounds of this review

- Executed on `GOOS=darwin` only. `linux`/`windows` behaviour is `go list`-derived through
  `ledger-consistency.sh` (exit 0, 228 rows across all three), not executed here.
- I did **not** execute the full multi-package `test-gate.sh` lane end to end on the intact
  candidate root — it exceeds this run's per-command bound. I ran its decisive stage
  (`suite-plan.sh` with `CI_REQUIRE_FULL_ROOT=1`, exit 0, `served=73 deferred=0`) plus the
  environments package and `platform-case-gate.sh` on the real stream. The producer's full-lane
  exit 0 at the 30m budget is accepted from attached evidence, not re-observed. The six **negative**
  lanes were run through `test-gate.sh` itself and are first-hand.
- The candidate lane has not run on hosted CI (see §1.8); all local.
