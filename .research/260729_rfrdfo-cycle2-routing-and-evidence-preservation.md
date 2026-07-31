# TASK-260729-rfrdfo — cycle-2 findings: evidence preservation and next-lever routing

Date: 2026-07-29
Role: researcher
Cycle: 2 (response to `TASK-260729-rfrdfo_review-verdict-cycle-1.md`)

> **CORRECTION, added 2026-07-29 by TASK-260729-3dr6hw cycle 7 (external, non-destructive).**
> §2.2 and §2.3 below reproduce the `TASK-260729-3dr6hw` diagnosis §11.2 arithmetic — and that
> arithmetic was **rejected as invalid** by `TASK-260729-3dr6hw_review-verdict-cycle-6.md` and
> **withdrawn** in cycle 7. Two independent errors: `saveJournal` fires three times per **staging
> entry** (`internal/transaction/staging.go:18-56`), not three times per *target*, so the `3·20`
> denominator term is not the source formula; and a fixture context stages **three** files
> (`SKILL.md`, `references/info.md`, `.csk-install.json`), not four — `csk-skill.json` is not in
> `whitelist.IncludeRoots` and `scripts/` is excluded — so the `4C` floor and therefore the whole
> `5C / (60 + 20C)` self-limiting bound do not hold. The trim also removes the `references/`
> **directory** entry, not only its file. The corrected numerator is **8 saves per affected context
> target**, at least ×1.6 the value used here.
>
> **§4.3 was then measured**, by `TASK-260729-2afulh` (`prototype-atomicity-fixture-trim`, `done`,
> review **accepted**): staging-path saves `100→84`, `140→116`, `110→94`, `107→91` — a
> **14.55–17.14 %** reduction, *inside* the 14.4–18.8 % band §2.2 called required, not the 7.1–12.5 %
> claimed here — and one valid `-count=1 -race` gate at real exit **0**, **492.231 s**: 12.231 s over
> the 480 s bar and 107.769 s under the 600 s alarm, on a single repetition (repetition 2 produced no
> exit file, repetition 3 was never run). So the disposition "not promoted" survives, but on wall
> clock rather than on the work-reduction ceiling asserted here, and the §3 conclusion "the test-only
> boundary is exhausted" is **withdrawn**: the best measured test-only configuration sits 2.5 % above
> the bar, inside the package's own 5.4 % spread. See `.research/260729_install-race-timeouts.md`
> §11.2–§11.5. Everything else in this document — the five AC integrity
> checks, the §2.4 mechanism analysis, the `journal.go:344` citation correction, the Patch A / Patch B
> dispositions, and the `TASK-260729-365r5r` routing — is unaffected and stands.

> **RESOLUTION, added 2026-07-29 by TASK-260729-rfrdfo cycle 3 (non-destructive).**
> The correction above is confirmed and §4.3 is **no longer undecided**. Independent convergence on
> the numerator: this cycle's source reading, review finding **R2-1**, and 3dr6hw cycle 7 all land on
> **8 saves per affected context target** — and TASK-260729-2afulh's measured staging manifest
> confirms it as an *equality*, `21 → 13` staging saves per context target (before: 2 directory +
> 3 file entries = `3×5 + 2×3 = 21`; after: `3×3 + 2×2 = 13`). Cycle-2 §2.1's "4 files per skill"
> is correct about what the fixture *writes*, but only **3** of them are *staged* — `csk-skill.json`
> and `scripts/` never enter a target. Both statements are true about different things.
> **§4.3 was then measured and rejected:** TASK-260729-2afulh built the 14-file trim and its focused
> count-one race gate exited **0 at 493 s** (`ok … 492.231s`) against the strict `<=480 s` bar — a
> **13 s miss with negative margin**. The lever is rejected on measurement, not arithmetic; the
> production route **TASK-260729-365r5r** stands. Full analysis:
> `.research/260729_install-race-timeout-corrected-routing.md`.

## 0. Headline

The cycle-1 verdict requested exactly two things: **preserve** the accepted evidence, and
**route** the next lever without widening this task's 13-file scope. Both are delivered.

1. **All five AC integrity clauses re-verified live this cycle**, against the current trees
   rather than by re-reading cycle-1 files. Real exit code 0 on every check.
2. **The next lever is fact-checked against source and rejected on arithmetic**, not on
   assertion. §4.3 (`references/info.md` fixture trim) cannot reach the 480s bar in any cell
   of its own best case. I reproduce the diagnosis's numbers independently and add a stronger
   structural bound it did not have.
3. **The routing has already been actioned.** The §6 production-side fix is open as
   **TASK-260729-365r5r** (`prototype-savejournal-namespace-validation`), status
   `development`, and it already carries the ≤480s bar in its own acceptance criteria.
   Nothing is left dangling.

**Disposition for rfrdfo: keep Patch A and Patch B exactly as they are, claim neither more
nor less than they achieved, and do not widen to a 14th file.**

No Go, build, race, vet, format, lint, install, stage, commit, publish, or pin command was
run this cycle — see §6 for why, and §7 for what that means for the numbers quoted here.

---

## 1. Preserved evidence — re-verified live, cycle 2

Every command below was run directly as a standalone process this cycle. No `tee`, no pipe
chain masking a status. Real exit codes.

| # | Check | Command | Real exit |
| --- | --- | --- | ---: |
| A | Source candidate byte-identical to cycle-1 pre-baseline | `diff evidence/source-candidate-manifest-pre.txt evidence-c2/source-candidate-manifest-c2-pre.txt` | **0** |
| B | Source candidate byte-identical to cycle-1 final | `diff evidence/source-candidate-manifest-final.txt evidence-c2/…-c2-pre.txt` | **0** |
| C | Prototype unperturbed since cycle-1 | `diff evidence/manifest-final.txt evidence-c2/prototype-manifest-c2.txt` | **0** |
| D | Only the 13 allowlisted files differ | `python3 bin/integrity.py <live candidate> <live prototype>` | **0** |
| E | Patch applies to the *current* candidate | `patch -p1 -d <jrrgw9> --dry-run --forward < …patch` | **0** |
| F | The dry-run did not mutate the candidate | `diff …c2-pre.txt …c2-post.txt` | **0** |
| G | Conformance root immutable (448 files) | path-normalised digest `diff` | **0** |

Notes that matter:

- **Check D is stronger than cycle-1's.** Cycle-1 compared two *recorded* manifests. I
  regenerated both manifests from the live trees and compared those, so the result cannot be
  inherited from a stale file. Output: `modified=13 added=0 deleted=0 unexpected=0
  forbidden_touched=0 not_modified_from_allowlist=0 / INTEGRITY_OK`. Both manifests are 391
  files. `internal/install/aba_test.go` and `internal/install/atomicity/fixture_test.go`
  appear in no list.
- **Check E** produced 13 `patching file …` lines and **zero** hazard markers — a grep for
  `fuzz|offset|reject|failed|malformed|garbage|hunk` over the dry-run log exited 1 (no match).
- **Check G required path normalisation.** Cycle-1's recorded conformance manifest stores
  *absolute* paths; `bin/manifest.sh` stores relative ones. A raw `diff` therefore reports all
  448 lines as changed — a **format artifact, not a digest change**. After normalising both
  sides to `./`-relative and sorting, the diff is empty at exit 0, 448 lines each side. This
  is recorded because a future reader running the raw diff will otherwise see an alarming
  448-line delta and conclude the immutable root moved. It did not.

### 1.1 Semantic checks re-verified

| Check | Result |
| --- | --- |
| `t.Parallel()` in `internal/install/*_test.go` | **88** of **107** declared tests |
| `aba_test.go` contribution | **0** |
| Skip-string set, candidate vs prototype | **identical**, `diff` exit **0** |
| `-timeout` token anywhere in the patch | none, grep exit **1** |
| Assertion counts (`t.Fatal/Fatalf/Error/Errorf`) across all 13 files | **identical pre/post, 13/13 OK** |
| Top-level `func Test` counts across all 13 files | **identical pre/post, 13/13 OK** |
| `scenario.injectClasses` occurrences | **1** (the injection loop) |
| `scenario.classes` occurrences in body | **2** (coverage assertion + constructor) |
| `sharedUserHome` | **absent** |
| Atomicity sweep parent `TestFailureAtEveryTargetClass…ReverseOrder` | **stays sequential** — `t.Parallel()` sits at `:194` inside the per-partition `t.Run`, not on the parent at `:150`, so its `t.Setenv` cannot panic |

Concrete assertion parity, per file: `stage_test.go` 172/172, `commit_test.go` 113/113,
`install_test.go` 108/108, `commit_atomicity_test.go` 41/41, `private_test.go` 34/34,
`cache_conformance_test.go` 28/28, `generation_test.go` 25/25, `dryrun_conformance_test.go`
17/17, `diagnostics_test.go` 15/15, `revalidation_test.go` 14/14, `activation_test.go` 11/11,
`maintenance_test.go` 11/11, `registry_e2e_test.go` 10/10.

> **Correction to a cycle-1 count.** Cycle-1's results table records `commit_test.go` at
> 109/109. Measured with `t.Fatal|Fatalf|Error|Errorf` it is **113/113**. The count differs
> because of the matching pattern, not because anything changed — pre and post are equal
> either way, which is the property being asserted. Flagged so the two documents can be
> reconciled rather than read as a discrepancy.

### 1.2 Recorded gate evidence, re-read

All 13 `gates/gate-*.exit` files contain **0**; `nonzero exits: 0`. No gate log contains
`DATA RACE`. The driver's 7 `go test` invocations each carry `-count=1`, each names a single
package, and none carries `-timeout`.

> A naive grep reports one `-timeout` and one `./...` in `bin/run-gates.sh`. **Both are on
> comment line 14**, which documents the prohibition. There is no such token in any command.
> Recorded so a future reviewer's grep does not read as a violation.

---

## 2. The routing question, and the answer

The cycle-1 verdict, item 2, offered two branches. **Branch B is correct: the test-only
boundary is exhausted.** Below is the evidence, checked against source rather than inherited.

### 2.1 §4.3's premises — all confirmed against the tree

| Claim (diagnosis §4.3/§11.1) | Verified? | Evidence |
| --- | --- | --- |
| `references/info.md` written at `fixture_test.go:83` | **yes** | `e.write(dir, "references/info.md", "ref")` |
| Exactly one occurrence in the whole package | **yes** | recursive grep returns that single line |
| Nothing asserts on it → removal is assertion-neutral | **yes** | the write is the only reference |
| Each fixture skill contributes 4 files | **yes** | `SKILL.md`, `references/info.md`, `scripts/<name>-tool`, `csk-skill.json` (`fixture_test.go:75-100`) |
| `runtime_roots` is `["scripts"]`, so the runtime tree never carried it | **yes** | `fixture_test.go:88` |
| The project scenario installs **3** contexts | **yes** | post-upgrade: `skill-a`, `skill-b` (project) + `skill-h2` (hybrid). `skill-h` is a managed *removal*; `e.skill()` is called 4× but creates source repos, not installed contexts |
| Global scenario is *less* favourable | **yes, new** | post-upgrade it installs **2** contexts (`skill-a`, `skill-b`), so it sheds 2 files, not 3 |

### 2.2 The arithmetic, recomputed independently

Required reduction to bring each measured run to 480s:

| Run | Measured | Required reduction | Absolute gap |
| --- | ---: | ---: | ---: |
| run 2 (best) | 560.828s | **14.41%** | 80.83s |
| run 3 | 564.022s | **14.90%** | 84.02s |
| run 1 (worst) | 591.280s | **18.82%** | 111.28s |

Save model, read off source this cycle: `saves = 3N + 5F`, with `N` = targets (20) and `F` =
staged files. `staging.go` calls `saveJournal` at `:26, :33, :56` (3 per staging entry) and at
`:141, :161` (2 per 32 KiB chunk — confirmed by reading the copy loop; every fixture file is
one chunk). Seven call sites in `staging.go`, sixteen in `engine.go` = **23 production call
sites**, exactly as §6 states.

Applying §11.2's range at the measured C=3:

| | Reduction | run 2 | run 3 | run 1 |
| --- | ---: | ---: | ---: | ---: |
| Floor `F=12` (best case) | 12.50% | **490.72s** | 493.52s | 517.37s |
| Ceiling `F=30` (worst case) | 7.14% | 520.77s | 523.73s | 549.05s |

**All six cells miss.** The most favourable cell reachable under the most favourable
assumption on every free parameter is 490.72s — **10.7s above the bar**. This reproduces
§11.2's table to the decimal.

### 2.3 A stronger bound than the diagnosis had

§11.2's weakness is that `F` is *statically bounded, not measured*, so a reader can ask what
happens if the real inventory is more favourable than assumed. It cannot be, and here is why:

The lever removes one file per **installed context**, so it removes `C` files. But the
absolute floor of `F` is *also* set by `C` — at minimum `4C`, since each context stages four
files. So the best-case reduction is not a free parameter at all:

```
reduction_max(C) = 5C / (3·20 + 5·4C) = 5C / (60 + 20C)
```

| C | Floor F | Best-case reduction | run 2 → | run 1 → |
| ---: | ---: | ---: | ---: | ---: |
| 2 | 8 | 10.00% | 504.7s | 532.2s |
| **3 (measured)** | **12** | **12.50%** | **490.7s** | **517.4s** |
| 4 | 16 | 14.29% | 480.7s | 506.8s |
| 5 | 20 | 15.62% | 473.2s | 498.9s |
| 20 | 80 | 21.74% | 438.9s | 462.7s |
| → ∞ | | **25.00% asymptote** | 420.6s | 443.5s |

The ratio is **self-limiting**: adding contexts to make the lever bigger adds staged files to
the denominator at four times the rate. Even the unreachable asymptote — infinitely many
contexts and *zero* runtime, shim, env-file and ledger files — tops out at 25%, and the
worst measured run would still need 18.82% with no margin for the `./...` gap.

At the actual C=3, **no assumption inside the model reaches the bar.** And the predicted
7.1–12.5% gain is only 1.3–2.3× the measurement's own 5.4% run-to-run spread, so even a
single post-trim run landing under 480s would not be distinguishable from noise.

### 2.4 §6's mechanism — confirmed, with one citation correction

`saveJournal` (`internal/transaction/journal.go:71`) calls `engine.validateJournal`, which
validates the target-namespace independence graph **twice per journal write**:

- `journal.go:344` — inside the plain `validateJournal(journal)` path
- `journal.go:354` — again in `engine.validateJournal`, with the journal root as a reserved path

> **Citation correction:** the diagnosis §6 cites "`journal.go:351` and `:354`". The first
> validation is actually at **`:344`**; `:351` is the *call* into `validateJournal` that
> reaches it. The load-bearing claim — validated **twice on every write** — is correct.

`validateIndependentTargetNamespaces` (`namespace.go:26`) is genuinely quadratic: it builds
`len(targets)*7` paths (4 candidates + 3 cleanup tombs) and runs a full pairwise loop at
`namespace.go:100-101`. Per path it performs `canonicalNamespaceTargetPath` (EvalSymlinks),
`namespaceCaseInsensitive` (ancestor walk + `unix.Pathconf`) and
`namespaceNormalizationInsensitive` (ancestor walk + `unix.Statfs`) — and on Darwin
(`namespace_case_darwin.go`) **none of these is memoised**. They re-run per path, per
validation, per save.

At N=20 that is ~140 paths, ~9,800 pairwise comparisons per pass, **two passes per save**,
and 120–210 saves per install: on the order of **2.4M–4.1M pairwise comparisons and >100k
syscalls per installation**. This is the dominant cost, and it is untouched by any test-only
lever because `references/info.md` is a file *inside* a target, not a target — **N does not
move**.

---

## 3. Routing conclusion

**The test-only boundary is exhausted.** Branch B of the cycle-1 verdict is taken.

| Item | Disposition |
| --- | --- |
| Patch A (`internal/install`, 88 × `t.Parallel()`, 11 files) | **Keep.** Measured 226.191–235.124s across three runs against a 600s alarm — passes the 480s bar with 245–254s of margin. Nothing further needed. |
| Patch B (atomicity partition, 2 files) | **Keep**, and keep the honest claim: it replaces a hard timeout (`FAIL 603.701s`) with three green runs at 560.828–591.280s. It is strictly better than what it replaces and it **does not satisfy the 480s bar**. |
| §4.3 `references/info.md` trim | **Not promoted.** Rejected on the arithmetic in §2.2–§2.3, not deferred for lack of effort. Remains optional and opportunistic *after* §6 lands, with a `StagingEntries` measurement attached. Not worth a 14th file. |
| §6 `saveJournal` namespace revalidation | **Already routed and live** — see below. |

### 3.1 The successor exists

**TASK-260729-365r5r — `prototype-savejournal-namespace-validation`**, status
`development`, under the same story (`STORY-260720-3plyvy`). Its acceptance criteria already
require: a literal file/function allowlist with pre/post manifests; a static call-path proof
that every externally supplied or recovered journal target graph is validated before mutation;
every `saveJournal` call still revalidating current filesystem namespace facts but at
**O(P) filesystem reads per pass instead of O(P²)**; fail-closed behaviour on malformed,
overlapping and between-save symlink changes; and focused evidence that either demonstrates a
defensible atomicity margin **≤480s** or explicitly rejects the prototype.

That is the correct successor and it carries the bar. Cycle-1 verdict item 4 — "any successor
implementation must return through a new reviewer cycle with count-one focused evidence" — is
structurally satisfied by that task's own AC.

**No new task needs to be opened by this cycle.** The verdict's branch-B obligation is
discharged.

---

## 4. Cycle-1 verdict items — disposition

| Verdict item | Status |
| --- | --- |
| 1. Preserve Patch A, the 88/19 split, the 13-file evidence, the assertion map, the cross-class retirement record | **Done.** §1 and §1.1 re-verify all of it live at exit 0. Nothing was edited this cycle. |
| 2. Do not widen scope; route the next lever | **Done.** Zero files changed. Branch B taken with the arithmetic in §2, and the successor is already open (§3.1). |
| 3. No timeout change, no skip, no weakened assertion, no full-repo race, no candidate mutation | **Held.** Skip parity exit 0; assertion parity 13/13; no `-timeout` in the patch; no Go command run at all; candidate byte-identical (checks A/B/F). |
| 4. Successor returns through a new reviewer cycle with count-one focused evidence | **Structurally satisfied** by TASK-260729-365r5r's own AC. |

---

## 5. Key aspects

1. **Patch A is a clean win and should not be held hostage to Patch B.** 226–235s against a
   600s alarm, with no `DATA RACE` across three race repeats of 88 newly-concurrent tests.
2. **Patch B is a real improvement that must not be over-claimed.** Timeout → three green
   runs is progress. 8.72s of margin on the worst run, inside the measurement's own 5.4%
   spread, is not a solved race gate.
3. **The 480s bar is load-bearing, not bureaucratic.** It exists to absorb the
   focused-versus-`./...` gap, where `cmd/curator` (557.779s under race) overlaps for nearly
   the whole run. It is correctly refusing a result that would likely fail the real gate.
4. **The quadratic is the whole problem.** Every test-only lever moves the *save count*; none
   moves `N`. That asymmetry is why §2.3's bound is structural rather than circumstantial.
5. **`cmd/curator` is next in line to fail** at 557.779s of a 600s alarm, and the §6 fix
   protects it too — which is an argument for §6 independent of the atomicity gate.
6. **New this cycle:** the conformance-manifest path-format artifact (§1) and the
   comment-line `-timeout`/`./...` grep hits (§1.2) are both benign, and both look like
   violations to a fresh grep. Recorded so they are not re-litigated.

---

## 6. Why no Go command ran this cycle

At the start of this cycle the process barrier
(`pgrep -af '(^|/)(go|.*\.test)( |$)|go-build|cmd/curator'`) was **not empty**: PID 53895 was
`…/TASK-260729-365r5r/gotmp-baseline/atomicity/atomicity.test`, i.e. the successor task
measuring its own atomicity baseline concurrently.

Any timing measurement taken under that contention would be worthless, and re-running a ~560s
race gate would additionally corrupt *its* baseline. This cycle's obligations are preservation
and routing, neither of which needs a timing run, so no Go command was issued. The SHA-256
manifests, the patch dry-run and the grep-based semantic checks are all contention-insensitive
and were run normally.

---

## 7. Evidence-honesty statement

- Every exit code in §1 and §1.1 is the real status of a command run directly this cycle as a
  standalone process. No `tee`, no pipe chain hiding a status.
- **No Go command was run this cycle** — no build, test, race, vet, gofmt, lint, install,
  stage, commit, publish or pin. All timing numbers (226.191–235.124s, 560.828–591.280s,
  285.434s, 603.701s, 557.779s) are **quoted from recorded cycle-1 `gates/*.exit` and
  `*.seconds` files or from the diagnosis document**, re-read this cycle, not re-measured.
- The three atomicity race runs are reported as **exiting 0 and missing the 480s pass
  condition** — both facts together. Patch B is not presented as a clean pass.
- §2.3's model is arithmetic over a save model read from source. `F` remains **statically
  bounded, not measured**; the conclusion is taken from the *best* case of that range and from
  the asymptote, so the residual uncertainty cannot flip it.
- The `N=20` target count is inherited from the diagnosis's §1.3 alarm-frame reading and was
  **not independently re-derived** here.
- One cycle-1 figure is corrected (`commit_test.go` 109 → 113, a pattern difference, not a
  change) and one diagnosis citation is corrected (`journal.go:351` → `:344`). Neither
  affects any conclusion.
- The task-owned prototype and the source candidate were **not written to** at any point this
  cycle. Zero files changed.
