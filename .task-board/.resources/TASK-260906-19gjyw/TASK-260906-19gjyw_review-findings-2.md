# TASK-260906-19gjyw — review findings, cycle 2 (landing decision)

Verdict: **ACCEPT.**
repeat-of: none — every class cycle 1 raised is closed by measurement, and no
new blocking or major class recurs.

Subject: `relux-works/curator`, branch `feat/consume-overlay-rule`, head
`e4ddca19` (four signed commits past merge-base `7c74a492`). PR
https://github.com/relux-works/curator/pull/62. Authority: curator-spec
`87a0d0060bad`. Change Request `CR-TASK-260906-19gjyw-2` revision 2.

Everything below was measured in a throwaway copy of the producer's worktree
(`/tmp/rev2/{clean,curator}`, `rsync` excluding `.git`/`.temp`/`.task-board`).
Nothing was written into the producer's worktree, into PR #62, or into the
control root. No two suites ran concurrently and no `-race` suite ran alongside
anything here. No conformance root was materialized in this cycle, so no
`git archive` was involved.

---

## 0. The empty repository delta — why accepting it is right here

`repository_delta=empty` is **structural, not a producer omission.** The story
workspace this Change Request snapshots is a checkout of **curator-spec**, and
its `HEAD` is `87a0d0060bad…` — the authority revision itself:

```
HEAD           = 87a0d0060bad64ab883d007dcdf35df7485368bf
HEAD^{tree}    = 2e6ca472ae8542a95e446ed4afb11cb75820536d
base tree      = 2e6ca472ae8542a95e446ed4afb11cb75820536d
candidate tree = 2e6ca472ae8542a95e446ed4afb11cb75820536d
git diff <base> <candidate>  ->  zero paths
git status --short           ->  clean
```

This leaf's scope is the **curator** repository. Its deliverable is
`feat/consume-overlay-rule@e4ddca19` on PR #62, which lies outside the CR's
snapshot scope by construction. The producer could not have written anything
into this tree without writing into the wrong repository, and the rework brief
for this run explicitly scoped the work to the curator worktree. An empty tree
is therefore the truthful record of it, and every claim below is measured
against the curator branch rather than against this snapshot.

**For the orchestrator, unchanged from cycle 1:** integrating this CR moves no
curator code either way. Landing PR #62 is the act that delivers the leaf.

---

## 1. F1 — the property loop now checks its own subject. Confirmed by mutant.

**M10 applied, then measured**, each test as its own process with `=== RUN`
lines counted so a filter matching nothing cannot read as a survivor. The
mutant deletes the widening, leaving `return identity.ClassifySource(trimmed)`:

| test | `=== RUN` | exit | verdict |
|---|---:|---:|---|
| `TestInstallNeverDemotesANetworkIdentityToAPath` | 1 | **1** | **killed** |
| `TestOperandShadowedByDirectoryResolvesAsGit` | 1 | 1 | killed |
| `TestInstallOperandKindIsSyntactic` | 1 | 1 | killed |

The failing output names all three members of the class the widening exists for:

```
operand "github.com/evil-org/pkg"     carries network identity ... but classified path
operand "github.com/relux-works/pkg"  carries network identity ... but classified path
operand "example.com/org/pkg"         carries network identity ... but classified path
checked 8 network-identity operands, 3 of them bare canonical
```

At `38702164` this same test logged `checked 5` and exited 0. The repair is real.

### 1.1 The `bare` guard fires, and it cannot be satisfied vacuously

**Fires.** I removed the three bare-canonical rows from `installOperandCases`
and re-ran the test alone: exit **1**, `=== RUN` 1,
`the matrix carries no bare canonical identity: the widening's own class is unchecked`.
The guard is not decoration.

**Not vacuous.** The guard counts an operand only when `identity.Parse` yields
nothing *and* `identity.ValidCanonical` is true. I drove 90 spellings through
both predicates and the discriminator: **every spelling `ValidCanonical`
accepts is classified `path` by `ClassifySource`** — 15 such spellings in the
corpus, zero disagreements. `ValidCanonical`'s host grammar
(`^[A-Za-z0-9][A-Za-z0-9.-]*$`, no colon) makes this structural, not incidental:
no `://` prefix can exist, and no colon can precede the first `/`, so
`sourceGitSchemeRE`, `sourceSCPRE` and `sourceColonRE` all fail and the arm is
always `path`. Therefore every operand the counter counts is one the widening
**decides**, in the direction the property asserts. The count is a claim that
now holds.

## 2. The three artifacts that carried the false claim

| artifact | now says | checked how |
|---|---|---|
| the test's own comment | that a Parse-only filter excludes the class, and that the `bare` guard exists so the count cannot go back to covering nothing | true by the §1 measurement |
| drafting report §3 + §4 row M10 + total | M10 kills 3 of 3 **after** the repair, and that the third survived as first written; total restated as 19 of 19 with M18 and M19 named | read against the measured result |
| `.github/ci/platform-cases.tsv:312` | "over both spellings of that class — the canonicalizable URL or SCP remote and the bare already-canonical `host/path` the install widening exists for" | matches the repaired loop exactly: `Parse` non-empty **or** `ValidCanonical` |

Row 311 was widened in the same commit and is likewise truthful: the matrix now
carries `packages/team:context → SourcePath`, which is what "a colon in a later
segment is a path operand" asserts.

Both edited rows were **observed passing by name in the full 225-row ledger gate
on all three hosted runners at this exact head** (run 34076889889): `Test
(ubuntu-latest)`, `Test (macos-latest)`, `Test (windows-latest)` and both `Race`
jobs all print `ok internal/envprofile :: TestInstallOperandKindIsSyntactic` and
`ok internal/envprofile :: TestInstallNeverDemotesANetworkIdentityToAPath`, and
`Ledger consistency` prints `[must=linux,darwin,windows skip=-]` for each. That
is stronger than the producer's own claim, which conservatively rested on a
two-row filtered local run.

## 3. F2 — the added class, and whether the enumeration is now complete

I re-derived the enumeration rather than reading the table, and more completely
than either predecessor: stage (c)'s `isPathOperand` **and** stage (c)'s
`identity.Parse` (with the `len(host) == 1` carve-out restored) reconstructed
beside the head's helpers, then **82 operands** driven through both *complete*
compositions — `isPathOperand → canonicalGit` versus `installOperandKind →
canonicalGit` — so a spelling refused downstream is not miscounted as a kind
change. Result: **82 operands, 61 unchanged, 21 changed.**

`packages/team:context` and `a/b:c` reproduce as **refused → path**, exactly as
F2 found. The row is in the report's table, and `packages/team:context` is
pinned in `installOperandCases`.

**The pin is load-bearing.** M19 narrows the install classifier to refuse a
colon reached after a `/` — stage (c)'s verdict for that class and nothing else:

| matrix | `=== RUN` | exit | verdict |
|---|---:|---:|---|
| reviewed matrix, operand removed | 1 | **0** | **SURVIVED** — the F2 gap, reproduced |
| with `packages/team:context` | 1 | **1** | **killed** — `installOperandKind("packages/team:context") = invalid, want path` |

**The security-critical direction is clean across all 82 operands.** I printed
each operand's *old* clone destination and filtered mechanically: of the 23
operands that reached `https://<canonical>` before this change, 22 stay `git`
and one (`github.com:/org/pkg`) becomes an explicit refusal. **Zero become a
local path.** That is the F14 invariant, measured rather than argued.

### 3.1 Two spellings outside the table's wording — observation, not a finding

My wider corpus surfaces two spellings the table's prose does not name. Both are
inside classes the tests already pin, and neither is a defect; I record them so
they are not rediscovered as novel:

* **`a b:c` and `example.com/a:b` — refused → path.** Same discriminator arm as
  `a/b:c`: the first colon is not reached without crossing a `/` **or
  whitespace**. The table and the ledger row both spell this arm as "a colon in
  a *later segment*", which describes the `/` case only. The arm itself is
  pinned by `packages/team:context` and M19 kills a narrowing of it, so this is
  a wording that is narrower than the class it names, not a coverage gap. No
  network identity is involved; the operand reaches `LoadManifest` and fails.
* **`""` and `" "` — git → refused.** Reachable from `run()`
  (`cmd/curator/profile.go:100` accepts any single positional, including an
  empty one). Before, an empty operand entered the git arm with an empty
  canonical; now `installOperandKind` trims and `ClassifySource("")` refuses it
  with `profile_source_invalid`. `""` is pinned in `installOperandCases`; `" "`
  is not. This is accept → refuse, the safe direction, and an improvement.

Neither is blocking, major, or minor enough to spend a cycle on: the product
behaviour is what the landed schema decides, the arm is pinned, and the
direction the AC guards is measured clean.

## 4. F3, and its neighbourhood

The doc comment is moved: `sourceKindRefusal` now sits above
`func sourceKindRefusal`, and `installOperandKind` carries its own block —
including the widening's justification, which is the comment a reader most
needs.

I checked the neighbourhood by construction rather than by eye: a `go/ast` pass
over the six production files this change touches (`identity/sourcekind.go`,
`identity/identity.go`, `envprofile/envprofile.go`, `envprofile/overlays.go`,
`config/environments.go`, `cmd/curator/compose.go`) reporting every doc block
whose first word is not the symbol it is attached to. **Two hits, both
pre-existing at merge-base `7c74a492` and on `origin/main`, neither introduced
here:**

* `identity.go:23` `scpRE` — the block starts "scp-style remote…", lowercase
  prose describing `scpRE` itself. Benign.
* `envprofile.go:1160` — a `strictAuditMember` doc block sits above
  `var canaryPasses`, leaving `func strictAuditMember` undocumented. This is
  the same shape as F3 and worth a separate cleanup someday, but it is **not in
  this change's blast radius** and is out of scope for this leaf.

I also checked the load-bearing comment claims rather than trusting them:
`ClassifySource`'s "never touches the filesystem" holds — `sourcekind.go`
imports `regexp` alone, and the whole `identity` package's non-test sources
contain no `os.`, `filepath.`, `ioutil.` or `Stat(` reference. The "one
discriminator" claim holds: four production call sites, one helper —
`cmd/curator/compose.go:57`, `internal/config/environments.go:446`,
`internal/envprofile/overlays.go:53`, `internal/envprofile/envprofile.go:1327`.

## 5. Regression surface — proved, not assumed

`git diff 38702164..HEAD` is three files, +43/−20: one TSV, one test file, one
non-test Go file. For the non-test file I proved the stronger statement by
construction: I parsed `envprofile.go` at both revisions **without**
`parser.ParseComments`, printed every declaration through `go/printer`, sorted
by name to absorb the reordering, and diffed.

```
AST IDENTICAL (comments stripped, declarations sorted)
```

**No product behaviour moved.** The discriminator's classification and the
install kind decisions at `e4ddca19` are byte-for-byte the decisions cycle 1
verified at `38702164`. That is also why the substance cycle 1 established —
`canonicalGit` returning a bare `host/path` verbatim, `ensureRepo` cloning it
over https, the planted-directory attack reproducing under a widening-free
mutant, the 121-spelling three-engine schema differential — carries forward
without re-derivation, as the brief directs.

## 6. Gates re-run in my own hands, at `e4ddca19`

Each a standalone process, run sequentially, on a clean rsync copy with no
probe files:

| gate | command | result |
|---|---|---:|
| build | `go build ./...` | exit 0 |
| vet | `go vet ./...` | exit 0 |
| gofmt | `gofmt -l cmd internal` | 0 lines |
| suppression | `.github/ci/no-broad-suppression.sh` | ok, exit 0 |
| ledger | `.github/ci/ledger-consistency.sh <dir>` | 225 rows across linux darwin windows, ok |
| touched packages | `go test -count=1 ./internal/envprofile ./internal/identity ./internal/config` | ok — 67.2s / 0.4s / 1.4s |
| CLI suite | `go test -count=1 -timeout 25m ./cmd/curator` | ok — 298.0s |

Hosted, read from the runs' own records rather than from a report:

* **PR #62 head is `e4ddca19`** — the exact head under review — `MERGEABLE`,
  `CLEAN`, **11 checks SUCCESS**, `Candidate suite` `SKIPPED` as designed for a
  `pull_request` event (the lane is `workflow_dispatch`-only by its `if:`).
* **Candidate lane, run 34071813375 at `38702164`:** all three `Candidate suite`
  jobs `success` — ubuntu 3m38s, macos 8m7s, windows 32m17s. Each job's own log
  shows `CANDIDATE_REF: 87a0d0060bad64ab883d007dcdf35df7485368bf`,
  `candidate-suite: revision accepted (immutable, full 40-hex)`,
  `CI_REQUIRE_FULL_ROOT: 1`, `suite-plan: served=71 deferred=0 excluded=1`,
  `test-gate: stage served exit=0`, and `ok internal/config ::
  TestManagerConfigV2SchemaCases` — the test carrying the forty-one overlay
  cases and the thirty subcases this leaf exists to close.
* **The `SPEC_PIN` root still defers as designed** at `e4ddca19`:
  `platform-case gate: deferred packages (root unset): internal/config
  internal/envfragment internal/envmarker` on every default-lane job.

**Bound, stated rather than smoothed over.** The candidate lane has not been
dispatched at `e4ddca19`; its standing measurement is at `38702164`. I judged
the producer's portable argument rather than accepting it, and it holds at every
step: the non-test delta is AST-identical (§5); `grep -rl CURATOR_CONFORMANCE_ROOT
--include='*.go' internal/envprofile` returns **0 files**, so the one changed
test file's verdict cannot depend on which root is served; the ledger edit
touches the free-text description column only, leaving both test-name tokens and
both `must_run_on` values untouched; and the changed tests were themselves
observed `ok` **on all three runners at `e4ddca19`** in the default and race
lanes (§2). What remains unmeasured at this head is only the *conformance-root*
dimension, and nothing in the delta reads it. I report that as a reasoned bound,
not as a measurement.

## 7. Acceptance-criteria coverage as measured by this review

**7 of 7 AC rows driven**, two carrying stated bounds that are the authority's
shape rather than gaps in the work.

| # | AC row | production call site | verdict |
|---|---|---|---|
| 1 | path overlay reachable from all three surfaces, joins the closure with its weight, through `run()` | `config.Load` → `PolicyFromConfig` → `Install` → `resolveOverlay`; `cli.cmdComposeAdd` | driven — cycle 1 verified; product code AST-identical since (§5) |
| 2 | one exported discriminator matches the landed schema, never probes the filesystem | `identity.ClassifySource`, 4 production call sites | driven — cycle 1's 121-spelling three-engine differential, 0 disagreements; no-filesystem re-confirmed here (§4) |
| 3 | a path overlay carrying range/tag/branch/revision/directory stays `profile_source_invalid` | `parseOverlay`, `resolveOverlay` | driven for four of five; `branch` is an accepted stated bound — `$defs/overlay` has no such property and `additionalProperties: false` refuses it at the reader |
| 4 | `profile install` uses the same helper, **no source silently changing kind** | `installOperandKind` → `installLocked` | driven — 82-operand re-derivation, F14 direction measured clean, `packages/team:context` pinned, M19 kills a narrowing (§3) |
| 5 | candidate lane green on three runners with `CI_REQUIRE_FULL_ROOT=1`, closing the 30 subcases | `.github/ci/test-gate.sh` | driven — run 34071813375 read from its own logs; bound at §6 for the head offset |
| 6 | the stale bound and ledger rows 303–305 retired and **truthful** | `.github/ci/platform-cases.tsv` | driven — F1 closed; both edited rows observed passing by name in the full ledger gate on three runners at `e4ddca19`; no residual stale-bound text anywhere in the tracked tree |
| 7 | every new refusal driven through `run()` and proven by a narrowing mutant | as above | driven — 3 mutants re-applied here (M10, the `bare`-guard removal, M19), 3 killed; cycle 1 re-ran 16 of 17 |

## 8. The landing question

**PR #62 is safe to land, without hedging.**

Cycle 1 established the substance and found no product defect. This cycle's
subject was three corrections to evidence, and all three are closed by
measurement: the mutant that used to survive now kills, the guard that reports
the count now fails when the class empties, the omitted class is pinned by an
operand whose removal lets a narrowing mutant survive, and the doc comment
documents its own function. The product code did not move — proved by AST
identity, not asserted. The head is mergeable and clean, eleven hosted lanes are
green on it, and the candidate lane that this leaf exists to turn green is green
on all three runners against curator-spec `87a0d0060bad` with
`CI_REQUIRE_FULL_ROOT=1`.

Two follow-ups for the orchestrator, neither blocking and neither this
producer's to fix:

1. The `strictAuditMember` / `canaryPasses` doc block (§4) is the F3 shape
   living on `origin/main` since before this branch. Worth a one-line cleanup on
   some future leaf.
2. `origin/main` is one commit ahead (`a66eec88`, board and logbook records
   only, no code). Its logbook entry still describes the consumption gap as
   open; after landing, that entry wants updating. Runs are forbidden to write
   `LOGBOOK.md`, so it is the orchestrator's.
