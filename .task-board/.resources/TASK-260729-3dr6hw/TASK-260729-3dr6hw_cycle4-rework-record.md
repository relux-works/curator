# TASK-260729-3dr6hw — cycle 4 rework record

Date: 2026-07-29
Role: researcher
Routing in: `analysis` (cycle-3 verdict = CHANGES REQUESTED to analysis)
Routing out: `to-review`
Deliverable: `.research/260729_install-race-timeouts.md`,
SHA-256 `46316048a5eec0f45c48d34d4dc2751c19a2b7790ddbba1dda85dd0474cf45cb`
(cycle-3 SHA-256 was `d2dc2cbc7e332c4590cf49b8b0d37e9753f497b1c177d371554262c73f940b6b`)
Board resource: `TASK-260729-3dr6hw_install-race-timeout-diagnosis.md` — updated in place, not duplicated.

**No Go test, build, vet, or tooling command was executed in this cycle**, consistent with the task
boundary ("diagnose without running Go tests and without edits"). No candidate file was modified.
The only writes were to `.research/`, `.temp/TASK-260729-3dr6hw/`, and the board.

The cycle-3 verdict scoped this as "one bounded research/evidence repair" of §3.5. That is what was
done. §4 (the patch), §5 (assertion matrix), §7 (validation commands), §8 (integrity gates), every
allowlist, and every timing band are byte-unchanged.

---

## 1. Cycle-3 required corrections — disposition

### Correction 1 — the §3.5 selection rule was not honest about what it excluded

**Reviewer claim independently verified: TRUE.** `internal/runtimestore` completed on both sides at
18.630s non-race and 18.244s race — **×0.98** — and cycle 3 omitted it while writing that the table
listed every completed package large enough for a meaningful ratio.

**Disposition.** §3.5 now states a mechanical selection rule up front: *completed (`ok`) on both
verifier-3 gates, and `max(non-race, race) ≥ 10s`*. That rule was applied programmatically to all 38
packages that completed on both sides and yields **exactly five**, all of which are now listed:

| package | non-race | race | factor |
| --- | ---: | ---: | ---: |
| `cmd/curator` | 384.270s | 557.779s | ×1.45 |
| `internal/godriver` | 55.306s | 76.752s | ×1.39 |
| `internal/transaction` | 27.218s | 34.106s | ×1.25 |
| `internal/runtimestore` | 18.630s | 18.244s | ×0.98 |
| `internal/managerlock` | 5.366s | 15.542s | ×2.90 |

The word "every" is withdrawn from the ×1.25–1.45 claim, and the claim itself is stated to have been
**false**. A second omission the reviewer did not name was found and disclosed while applying the
rule: `internal/managerlock` is under 10s non-race but 15.542s under race, so it enters under
`max(...)` and reads **×2.90** — the largest completed-package factor over 10s, and a direct
counterexample to any homogeneity claim. The corrected section states the completed range as
×0.98–×2.90 under the rule and ×0.78–×8.47 across all 38.

### Correction 2 — the subprocess mechanism attribution

**Reviewer claim independently verified: TRUE.** `internal/transaction` has 50 test functions and
exactly **2** `exec.Command` sites in `*_test.go`, both at `subprocess_test.go:56` and `:459`, both
re-executing the test binary (`exec.Command(os.Args[0], "-test.run=^TestTransactionCrashHelper$")`).
`grep '"git"' internal/transaction/*_test.go` returns nothing, and there is no `go build`
invocation. It is an overwhelmingly in-process package, and it sits at ×1.25.

**Disposition.** The subprocess mechanism claim is **withdrawn entirely**, for all three packages,
not softened for one. §3.5 now:

- states plainly that it corroborates no mechanism;
- carries a static-inventory table (tests / `exec.Command` sites / `"git"` literals) for all five
  selected packages, with the explicit caveat that call-site counts are textual occurrences and not
  wall-time shares;
- notes that `internal/managerlock` is *also* in-process, has zero `time.Sleep` in its tests, and
  reads ×2.90 — so the same-run data does not separate in-process from subprocess work in either
  direction;
- takes the reviewer's permitted alternative and reframes the surviving reading as **workload
  shape**: the discriminator consistent with §2.2 is the journal target count `P` that `saveJournal`
  re-validates through an O(P²) pairwise loop. The alarm dumps record `P` = 8 (`internal/install`)
  and `P` = 19, 20 (atomicity); `internal/transaction`'s ~33 `Prepare` call sites use
  `Targets: []Target{…}` literals of one to a few entries, i.e. the opposite end of the same curve.
  This is labelled a hypothesis that predicts a direction and not a magnitude.

The corrected section keeps the observation the reviewer allowed — that `internal/install`'s ×1.77 is
a genuine same-run **floor** on the exact candidate — and separates it from the unprofiled causal
explanation.

## 2. Findings originated in cycle 4 (not requested by the reviewer)

1. **Atomicity's same-run ratio is censored and was silently misleading.**
   `603.701 / 441.122 =` **×1.3686**, which is *lower* than `cmd/curator`'s ×1.45, with the numerator
   truncated by the alarm while the sweep was still running and three post-sweep tests had not
   started. Cycle 3's closing line "Atomicity is worse still" was true but unsupported by the table it
   sat under. §3.5 now marks that row as censored and points at §3.4's phase-level solve for
   atomicity's real in-gate factor.

   > **Cycle-5 amendment (verdict R4-1).** Cycle 4's own wording of this finding was itself wrong in
   > two ways, and both are corrected here and in §3.5 of the main diagnosis:
   >
   > * **"No information" is withdrawn.** The ratio is a **right-censored one-sided lower bound** on
   >   the completed factor: the completed race duration is strictly greater than 603.701s, so the
   >   completed factor is strictly greater than `603.701 / 441.122 = 1.368557… ⇒` **> ×1.3685**. It is
   >   the *same kind of object* as `internal/install`'s **> ×1.7670**, which cycle 4 correctly called
   >   a floor — treating one as a floor and the other as information-free was internally
   >   inconsistent. The correct statement is that atomicity's bound is **too weak** to support
   >   cross-package comparison or to estimate the completed factor, because its censoring is severe
   >   (the binding sweep chain unfinished *plus* three whole tests never started), not that it is
   >   empty.
   > * **"Purely because the alarm truncated the numerator" is withdrawn.** The *observed* value is
   >   low because the numerator is truncated; how far above ×1.3685 the completed factor sits is
   >   **not derivable from this row**, and no claim is made about its position relative to
   >   `cmd/curator`. Atomicity's completed in-gate factor comes from the separate phase-level solve
   >   in §3.4 (×3.33 sweep, ×4.14 activation prefix), which is unchanged and still accepted.
   >
   > Additionally, both two-decimal figures round **up** from their quotients (`1.368557…` → 1.37,
   > `1.767075…` → 1.77), so "> ×1.37" and "> ×1.77" are not strictly true floors. Bounds are now
   > stated as **> ×1.3685** and **> ×1.7670**; the 2-dp values appear only as roundings.
   >
   > Everything else in this record — the §3.5 selection rule and its five packages, the withdrawn
   > subprocess attribution, the static-inventory table, the target-count hypothesis and its
   > counterexamples, §9 risk 6, and the checklist disposition — stands unchanged.
2. **The same defect existed in two places the reviewer did not name.** §2.2's closing clause and
   §3.3's closing clause both asserted that atomicity inflates more because `internal/install` spends
   its time in `git` subprocesses "which `-race` does not slow". Both are repaired to state the
   target-count difference as a hypothesis without the subprocess attribution. Leaving them would
   have made the §3.5 repair cosmetic.
3. **§3.1's status table over-rated the mechanism.** Its row read "Race-factor *mechanism* … measured
   in the failing gate itself, no cross-tree assumption". Downgraded to **hypothesis, not measured**,
   with the note that no projection in §3.2 or §3.4 depends on it.
4. **New §9 risk 6** records the mechanism as unmeasured, names the two in-process counterexamples,
   states that no band and no patch decision depends on it, and names the cheap confirmation (a
   focused `-race -cpuprofile` run of the atomicity sweep) as out of bounds for this no-execution
   task. Old risk 6 renumbered to 7; no cross-reference pointed at it.

## 3. Verification commands run in this cycle, with real exit codes

Every command below is read-only. Each was run as a standalone process; no pipe chain hides an exit
code. **No `go` command of any kind was invoked, and no expected-red gate was run**, so nothing here
is presented as a green gate that did not run.

| # | Command | Real exit | Result |
| ---: | --- | ---: | --- |
| 1 | `python3` reducer over `verifier3/go-test-{all,race-all}.log` (parse `^(ok\|FAIL)` lines, join by package) | 0 | 40 packages in both logs; **38** completed `ok/ok`; selection rule `max ≥ 10s` yields **exactly 5**; min ×0.78 `internal/capabilities`, max ×8.47 `internal/buildmeta` |
| 2 | same reducer, timed-out packages | 0 | `internal/install` ×1.77, `internal/install/atomicity` ×1.37, both with race side = `FAIL` (censored) |
| 3 | `grep -h -c '^func Test' internal/transaction/*_test.go` (summed) | 0 | `50` |
| 4 | `grep -n 'exec\.Command' internal/transaction/*_test.go *.go` | 0 | 2 sites, `subprocess_test.go:56,459`, both `os.Args[0] -test.run=^TestTransactionCrashHelper$` |
| 5 | `grep -n '"git"\|go build\|git init' internal/transaction/*_test.go` | **1** | **no match** — expected empty. Exit 1 is what `grep` returns for zero matches; reported as-is, not dressed up as a pass |
| 6 | `grep -c 'Prepare(' internal/transaction/*_test.go` (summed) | 0 | `33` |
| 7 | per-package `exec.Command` / `"git"` / `time.Sleep` static inventory over the 5 selected packages plus the 2 timed-out ones | 0 | populates the §3.5 static table |
| 7a | `grep -h -c 'exec\.Command' internal/managerlock/*_test.go` | 0 | `0` + `1` = **1** |
| 7b | `grep -h -c 'time\.Sleep' internal/managerlock/*_test.go` | **1** | `0` + `0` = **0 sleeps**; exit 1 is grep's zero-match code, reported as-is |
| 8 | `grep -h -c '^func Test' internal/runtimestore/*_test.go` (summed) | 0 | `18` |
| 9 | `grep -n 'subprocess\|does not slow\|does not instrument' .research/260729_install-race-timeouts.md` | 0 | only the cycle-4 changelog rows and the explicitly-withdrawn/hypothesis-labelled passages remain |
| 10 | `shasum -a 256 .research/260729_install-race-timeouts.md` | 0 | `46316048a5eec0f45c48d34d4dc2751c19a2b7790ddbba1dda85dd0474cf45cb` |

Commands 5 and 7b are the only non-zero exits in this cycle. Both are `grep` no-match results, which
is the expected and required outcome in each case; neither is a failing gate and neither is being
reinterpreted as a pass.

## 4. Checklist disposition

Unchanged from cycle 3, and restated so it is not silently re-litigated:

- Item 15 (**Tests green**) stays **unchecked**. No Go test, build or vet command has been run in any
  cycle of this task, and the task scope forbids running the full or race suite. There is no green
  command behind that item; checking it would be a false green under the Evidence Honesty Contract.
  It belongs to the downstream producer task that applies Patch A and Patch B and runs the §7
  focused commands.
- Items 13 and 14 remain checked as producer-side self-assessment. The AC clause "reviewer
  independently checks the plan" is by construction the next cycle's gate, tracked by item 16.

## 5. Unchanged from cycle 3 and preserved

Everything the cycle-3 verdict listed as independently accepted is carried forward without edit: the
exact timeout inventory (§1), the hot-path evidence (§2.1, §2.3–2.5), the 88/19 Patch A partition
(§4.1), Patch B's `injectClasses` selector, generated per-class scenarios and distinct global user
homes (§4.2–4.2.2), the 13-file required allowlist with `aba_test.go` and
`atomicity/fixture_test.go` excluded (§4.0), the retired-invariant disclosure (§5.1), the 220–340s /
340–460s projected bands and the unchanged 600s alarm, the focused producer commands and 480s pass
bar (§7), the pre/post SHA-256 candidate manifest and accepted-worktree delta gates (§8), the
immutable conformance-root invariant (§8.4), and the out-of-scope production-side finding routed as
a separate story (§6).

§2.2's call-frequency table (23 `saveJournal` sites: 16 + 7) and the two copy-loop anchors
`staging.go:141`/`:161` are unchanged; only the paragraph after the table was reworded.
