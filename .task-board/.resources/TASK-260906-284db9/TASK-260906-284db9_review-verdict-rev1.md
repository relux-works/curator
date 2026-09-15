# TASK-260906-284db9 — review cycle 1: ACCEPT

Reviewer run `RUN-260907-086d22`. Subject: branch `chore/promote-spec-pin` at
`04550e28`, two signed commits past `origin/main`, PR
https://github.com/relux-works/curator/pull/66. Six files, +243/−85.

Every number below was re-derived in this run, in throwaway detached worktrees under
`.temp/STORY-260907-2bddfc/worktree/.temp/review/`. Nothing was taken from the PR body or the
drafting report. The producer's worktree and the control root were never written to.

---

## 0. Verdict

**ACCEPT.** The premise is true, the pin is a qualified released revision, the thirteen removed
tolerances are exactly the rows that lost their deferral, the removal is load-bearing on the
production entry point rather than cosmetic, the PR #63 fail-closed guard survives intact, and all
five hosted default lanes **served** the environments and schema-2 config families and observed every
promoted case **passing** — not deferred, not tolerated.

Three non-blocking findings are recorded in §7. None of them is rework.

### On the empty repository delta

`CR-TASK-260906-284db9-1` rev 1 records `repository_delta=empty` with a zero-path patch. That is a
**workspace-provisioning artifact, not a producer failure**, and accepting it is right here:

- The Story workspace for `STORY-260907-2bddfc` is a **curator-spec** checkout (bound to
  `relux-works/curator-spec`, base `87a0d006`). This leaf changes **curator**:
  `.github/workflows/ci.yml`, `.github/ci/{platform-cases,root-artifacts}.tsv`,
  `.github/ci/gate-selftest.sh`, `internal/interop/environments/{contract,suite}_test.go`. None of
  those paths exists in curator-spec, so the change could not be made in the managed workspace at
  all. curator-spec shares curator's board (`task-board.config.json → local.board_dir:
  ../curator/.task-board`), which is how a curator task drew a curator-spec workspace.
- The real delta exists, is signed, is pushed and is fully reviewable — I reviewed **that**. The task
  **Scope** names this branch and worktree by path and base, so the producer worked where the task
  told it to.
- The review brief pre-declares the condition: *"the delta will read empty … That is structural in
  this epic and is not the finding."*

**Consequence the orchestrator must not miss:** integrating *this revision* lands nothing, because it
carries zero paths. `integrating` here does not mean the work reached trunk. The delivery is PR #66 at
`04550e28`, and it still needs its own landing. This acceptance is of the change on that branch.

---

## 1. The premise, re-derived (not taken from the PR body)

Both roots materialised as **plain checkouts** (`git worktree add --detach`), never `git archive`,
and verified file-by-file against their own `manifest.json` **in both directions**:

| root | revision | declared files | missing | hash mismatch | undeclared on disk | verdict |
| --- | --- | ---: | ---: | ---: | ---: | --- |
| old pin | `0ed5c691` | 691 | 0 | 0 | 0 | ok |
| new pin | `87a0d006` | 1047 | 0 | 0 | 0 | ok |

`export-subst` trap checked explicitly: `conformance/v1/fixtures/byte-exact/subst.txt` in the
materialised rc.11 root is **40 bytes** and carries `$Format:%H$` / `$Format:%h$` unsubstituted. An
archived root would carry 65.

| claim | verified |
| --- | --- |
| old pin manifest | `sha256:803918bf8672f76cf990985e51db213b826674cd5bb54fbf47731b8404b44403` ✓ |
| new pin manifest | `sha256:0e195ecd26af2fcb5e5afb5c3fef0ebe60981888b057e8efc60c0b48cf584529` ✓ |
| `release/1.0.0-rc.9.json` → `downstream_consumption.required_manifest_sha256` | `sha256:0e195ecd…529` — **the new pin** ✓ |
| `committed_release_pin_advanced` | `false` ✓ |
| `v1.0.0-rc.11` | annotated tag, **Good** SSH signature, `oparin@me.com`, key `SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM` ✓ |
| tag peels to | `87a0d0060bad64ab883d007dcdf35df7485368bf` = the value now in `SPEC_PIN` ✓ |
| both manifests' `protocol_version` | `1.0.0-rc.9` ✓ |
| `release/` at the tag | nothing later than `1.0.0-rc.9.json` — the record the comment cites is the current one ✓ |

So curator was testing against a root the release record no longer names as required, and the pin now
carries the revision it does. The policy holds: a signed released tag's 40-hex revision, not a branch
and not a candidate.

## 2. The rewritten comment — every clause checked

Tag / protocol version / digest / `required_manifest_sha256` provenance: all four verified above.
"Serves the WHOLE module, defers nothing": verified in §3. "A candidate suite never appears here":
`SPEC_PIN` is referenced at ci.yml:90, 186, 275, 369 — four `actions/checkout` refs, and the candidate
root enters only through `steps.candidate.outputs.root` at line 392/422.

Residual references to the old world, repo-wide (`0ed5c691`, `803918bf`, `6001dc33`), excluding
`.git` and `.temp`: **only `LOGBOOK.md` history and one closed board progress note.** No live file
still describes the old pin.

The retained clause "promoting it is owned by TASK-260720-38l1sy, after TASK-260720-25d05o qualifies
the release" is **not stale**: both tasks read `backlog` on the board, so the file does not contradict
board state.

The two neighbouring comments the producer also corrected (`test` job header, `candidate-conformance`
load rationale) are now true, and the measured Windows-contention justification for the 60m budget is
preserved verbatim — the number still has its reason.

## 3. The partition, before and after — compared by name

`suite-plan.sh` driven against each root:

| root | GOOS | served | deferred | excluded | deferred packages |
| --- | --- | ---: | ---: | ---: | --- |
| old `0ed5c691` | darwin | 69 | **4** | 0 | `internal/config`, `internal/envfragment`, `internal/envmarker`, `internal/interop/environments` |
| new `87a0d006` | darwin | 73 | **0** | 0 | — |
| new `87a0d006` | linux | 72 | **0** | 1 | — |
| new `87a0d006` | windows | 73 | **0** | 0 | — |

`diff` of the two served sets by name: **exactly those four packages moved, nothing else.** So no row
lost a tolerance it still needs — the inverse direction the brief asked for.

## 4. The removed tolerances — counted from the tree, not from the report

Field-level accounting of `platform-cases.tsv` at `origin/main` vs `HEAD` (235 rows both sides):

**13 rows lost `skip_allowed_on` + `class`**, all in exactly the four newly-served packages:

| package | rows |
| --- | ---: |
| `internal/interop/environments` | 7 |
| `internal/config` | 3 |
| `internal/envfragment` | 2 |
| `internal/envmarker` | 1 |

Every one keeps `must_run_on=linux,darwin,windows`. **No other column changed anywhere in the file.**
One row renamed (`TestTheLedgerTolerationForThisPackageIsTheDeferredRootClass` →
`TestNoLedgerRowForThisPackageToleratesASkip`) with its behaviour text.

The three schema drivers the acceptance criteria name — `internal/config` (3),
`internal/envfragment` (2), `internal/envmarker` (1) — are among them and are now required on every
lane. Confirmed observed passing on all five hosted lanes in §6.

> **Discrepancy, recorded so nobody hunts for missing rows later:** the review brief says "nineteen
> removed tolerances". The tree says **13**, and the producer's report says 13. The producer is right;
> the brief's number is not.

## 5. The removal is load-bearing — attacked, not read

### 5.1 The counterfactual, on one real stream

`SPEC_PIN` reverted to `0ed5c691` and the **default** lane (no `CI_REQUIRE_FULL_ROOT`) run end to end
against the old root through `test-gate.sh`:

| ledger | gate exit | verdict for the 13 promoted cases | required-row outcome |
| --- | ---: | --- | --- |
| **HEAD** | **1** — 26 named FAILs | all 13 `FATAL-not-tolerated` | `FAIL required case skipped on darwin` ×13 |
| `origin/main` (same `go-test.json`, same deferred set) | 1 | all 13 **`tolerated-by-ledger`** | 13 `tol` lines; the only FAIL is the rename artifact of mixing base ledger with head code |

`go test` exited **0** in both. Only the gate caught it. Before this change the same state was green
by tolerance; after it, a pin regression fails closed by name on the default lane. That is the whole
point of the leaf, and it reproduces.

### 5.2 Mutants I wrote and ran — all killed

Every one **narrows** the gate; none is a deletion. Restored by file copy between runs, never
`git checkout`.

| # | mutation | killed by | exit |
| --- | --- | --- | ---: |
| **M-A** | `TestConformanceEnvironmentsHeader` regains `linux,darwin,windows` / `root-unset` — deliberately **not** the row the shipped behavioural check targets | `TestNoLedgerRowForThisPackageToleratesASkip`; `gate-selftest.sh` (139/1) | 1 / 1 |
| **M-B** | **token-preserving**, on `internal/envfragment` — a package the Go test does not cover. `root-unset` occurrences in the mutated row: **0** | `gate-selftest.sh` static half by name (139/1). Go test correctly exits 0 — it is scoped and says so | 1 |
| **M-E** | **vacuity**: the `internal/envmarker` row deleted outright | `FAIL internal/envmarker has ledger rows to check` — the guard fires instead of passing silently | 1 |
| **M-F** | **token-preserving + behavioural**: `TestConformanceContextVersions` → `linux,darwin,windows` / `-`. Whole-file `root-unset` count **11 before, 11 after** — a grep sees a clean ledger | `platform-case-gate.sh` on the **real** `go test -json` stream from 5.1 flips that case `FATAL-not-tolerated` → **`tolerated-by-ledger`**; `gate-selftest.sh` kills it with 7 assertions, including all three **behavioural** ones that run the production gate as its own process (133/7) | 1 |

M-F is the DoD's token-preserving requirement discharged independently: the searched-for token is
preserved, behaviour changes, and the harness executes the behavioural suite, not only the static
checker.

I also read `platform-case-gate.sh` to confirm the semantics the new comments assert. They are
accurate: `skip_allowed_on` alone decides whether the toleration branch is entered
(`listed("-", goos)` is 0, so a `-` row refuses on every runner for every reason), and with
`class=-` the wrong-class test is bypassed so the row would tolerate a skip for **any** reason —
including an otherwise-unclassified one. Checking both columns is therefore necessary, not belt-and-braces.

### 5.3 The PR #63 guard, attacked per artefact

`CI_REQUIRE_FULL_ROOT=1` — the candidate lane's own setting — against the rc.11 root with exactly one
declared environments artefact removed at a time:

| root | exit | package named | artefact named |
| --- | ---: | :---: | :---: |
| intact rc.11 | 0 | — | — |
| minus `vectors/environments.json` | 1 | yes | yes |
| minus `vectors/context-versions.json` | 1 | yes | yes |
| minus `vectors/context-detectors.json` | 1 | yes | yes |
| minus `vectors/snapshot-acquisition.json` | 1 | yes | yes |
| minus `expected/environments` | 1 | yes | yes |
| minus `expected/byte-exact-snapshot_sha256.txt` | 1 | yes | yes |
| minus `fixtures/byte-exact` | 1 | yes | yes |

All seven are individually load-bearing. The guard is not weakened by the pin move.

The producer's flagged consequence reproduces and is correct behaviour:
`candidate-suite.sh verify-ref 87a0d006…` now refuses with *"candidate revision equals the committed
released pin … a candidate run must not impersonate the qualified pin"*, while a different revision
is still accepted. Re-qualifying that exact suite now goes through `candidate_root`.

## 6. The hosted measurement — served, not merely green

Run **34163836478** on head `04550e28`. **11 pass, 1 skipping** (the candidate lane, correctly not
dispatched). Evidence downloaded and read from all five uploaded artifacts:

| lane | plan | `plan-deferred` | deferred `go test` stage | promoted cases observed **PASSING** | `root-unset` skips | tolerated/excluded rows in those pkgs | gate |
| --- | --- | ---: | ---: | ---: | ---: | ---: | --- |
| Test ubuntu | served=72 deferred=0 excluded=1 | 0 | none | **14 / 14** | 0 | 0 / 0 | ok |
| Test macos | served=73 deferred=0 excluded=0 | 0 | none | **14 / 14** | 0 | 0 / 0 | ok |
| Test windows | served=73 deferred=0 excluded=0 | 0 | none | **14 / 14** | 0 | 0 / 0 | ok |
| Race ubuntu | served=72 deferred=0 excluded=1 | 0 | none | **14 / 14** | 0 | 0 / 0 | ok |
| Race macos | served=73 deferred=0 excluded=0 | 0 | none | **14 / 14** | 0 | 0 / 0 | ok |

`ok` in `platform-cases.txt` means observed **passing**; a deferred lane would have produced `tol`
lines or a `go-test-deferred.json` stage, and there are none of either. This is the distinction the
brief asked for: a green lane that still deferred would look identical from the outside, and this one
does not.

Race macos-latest took 19m56s and Test windows-latest 26m56s — both inside budget, both green.

## 7. Findings — all non-blocking

**F1 — a stale comment the sweep did not reach, pre-existing.**
`.github/ci/platform-exclusions.tsv:11` reads *"The committed released pin is such a root: it
publishes no qualification vector at all"*, justifying `default_excluded_on` as a fallback. Both the
old and the new pin publish `vectors/conformance-claim-v3-qualification.json` — present in both
materialised roots, and `suite-plan.sh` reports reading it from the root on every run. The rationale
as written is false. **Not caused by this change** (the file is untouched by the branch; last modified
in `cfffd7cd`) and behaviourally inert, since the vector wins whenever it is present — which it always
is. Same class the leaf swept elsewhere; worth a follow-up, not rework here.

**F2 — a summary line that overstates under failure.** `contract_test.go:514` logs
*"N ledger row(s) … checked, none tolerating a skip"* unconditionally, so under M-A the failing run
prints "none tolerating a skip" beside the `Errorf` that names the offending row. The test fails
correctly and the failure output is precise; only the trailing summary is wrong. Trivial severity,
but it is the "note claiming more than its scan enforced" shape this epic keeps finding.

**F3 — the brief's row count.** See §4: 13, not 19.

**F4 — the seven remaining `root-unset` rows: the scoping call is right, and the file explains it.**
`internal/skillspec` (1), `internal/marker` (1), `internal/moduleroots` (1), `internal/scriptpolicy`
(4) keep their column. Against this pin they are unreachable — but the old-root partition in §3 shows
those four packages were **served under the previous pin too**, so they were already unreachable
before this change and lost no deferral in it. Removing them belongs to its own task, and enumerating
them in the sweep would have made the change describe something it did not do. The bound is stated
where a reader meets it: `gate-selftest.sh` says they "were already unreachable under the previous
pin", and the schema-8 section comment says the class policy refuses the skip whatever the row says.
A next reader does not have to re-derive it. The producer's recommended follow-up — remove the seven,
then let the gate derive the promoted set instead of enumerating it — is the right shape.

## 8. Gates I ran myself

All sequential, each its own process with its observed exit code, on darwin/arm64. The two
`test-gate` lanes were never run concurrently and never alongside a `-race` suite.

| Gate | Exit | Result |
| --- | ---: | --- |
| `go build ./...` | 0 | — |
| `go vet ./...` | 0 | — |
| `gofmt -l cmd internal` | 0 | 0 files |
| `golangci-lint run ./...` | 0 | 0 issues |
| `bash .github/ci/ledger-consistency.sh` | 0 | 235 rows across linux darwin windows |
| `bash .github/ci/gate-selftest.sh` | 0 | **140 passed, 0 failed**; the 10 new assertions all present and passing |
| `bash .github/ci/no-broad-suppression.sh` | 0 | ok |
| `test-gate.sh` vs the rc.11 root, default lane | 0 | served=73 deferred=0; **one** `go test` invocation, no deferred stage; 20 skips, **0** `root-unset` |
| `go test ./internal/interop/environments/ -v` | 0 | 12 top-level `=== RUN`, 12 PASS, 0 SKIP, 0 FAIL; 7 `TestConformance*` |
| ledger test, anchored `-run '^TestNoLedgerRowForThisPackageToleratesASkip$'` | 0 | reports `12 ledger row(s) … 7 of 7 TestConformance case(s) matched` — a measured ratio, with `rows==0` and `cases==0` vacuity guards, both attacked in M-E |

`-run` anchoring: single top level only, no `Parent/child` filter anywhere. `=== RUN` lines counted,
not inferred from the `ok` line.

## 9. Delivery hygiene

| Commit | Signature | Author | Subject |
| --- | :---: | --- | --- |
| `494198f9` | **G** | Ivan Oparin \<oparin@me.com\> | Pin the released revision the protocol record actually requires |
| `04550e28` | **G** | Ivan Oparin \<oparin@me.com\> | Retire the tolerances the deferral they existed for has ended |

The report's per-commit bisect evidence names pre-rebase SHAs (`a1290fca` / `25fc913a`), so I
re-verified at the **current** objects. At `494198f9`: `go build` 0, `go vet` 0,
`ledger-consistency.sh` 0, `gate-selftest.sh` 0 (**130 passed, 0 failed** — 10 fewer than at HEAD,
exactly the assertions the second commit adds). **The branch bisects.**

No `LOGBOOK.md` entry, per the producer brief. `.github/ci/platform-cases.tsv` and
`.github/workflows/ci.yml` in the producer's worktree were never touched by this review; all mutation
happened in my own detached copies, restored by file copy and confirmed clean afterwards.

## 10. Definition of Done

| Row | Status |
| --- | --- |
| SPEC_PIN names the rc.11 released revision; comment describes that pin, no stale clause | ✅ §1, §2 |
| The three schema drivers lose their root-unset tolerance, required on every lane | ✅ §4, §6 |
| Every other ledger row that loses a deferral loses its tolerance too, enumerated | ✅ §4 — 13 rows, exactly the four newly-served packages |
| `suite-plan` reports deferred=0 for the four packages against the new root | ✅ §3, all three GOOS |
| `test-gate` green against a materialised rc.11 root, sequential | ✅ §8 |
| The candidate lane still fails closed on a root missing an environments family | ✅ §5.3, all seven artefacts |
| `gate-selftest.sh` and `ledger-consistency.sh` green | ✅ §8 |
| Roots as plain checkouts verified against `manifest.json`, never `git archive` | ✅ §1 |
| Negative tests that fail when the gate admits what it must reject; production call site named | ✅ §5.1, §5.2 — `platform-case-gate.sh`, `gate-selftest.sh`, `test-gate.sh` |
| At least one NARROWING mutant per gate; no delete-only evidence | ✅ §5.2 — four narrowing mutants, zero deletions |
| Token-preserving mutant against the text-inspecting gate, behavioural suite executed | ✅ §5.2 M-F |
| AC coverage as a measured ratio | ✅ 8 of 8 AC rows driven, re-derived here; §10 rows 1–8 |
| Lint clean; build not broken | ✅ §8 |
| Implementation matches AC; fits project architecture | ✅ |
| Tests green | ✅ §6 hosted, §8 local |
