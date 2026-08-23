# TASK-260729-3dr6hw — cycle 5 rework record

Date: 2026-07-29
Role: researcher
Routing in: `analysis` (cycle-4 verdict = CHANGES REQUESTED → analysis)
Routing out: `to-review`

Deliverable: `.research/260729_install-race-timeouts.md`
SHA-256 `5c2e89baef81d1b3e33bbe23702ea060d8ecae3c01617c618459687c116d438d`
(cycle-4 SHA-256 was `46316048a5eec0f45c48d34d4dc2751c19a2b7790ddbba1dda85dd0474cf45cb`)

Board resource `TASK-260729-3dr6hw_install-race-timeout-diagnosis.md` updated in place via
`task-board resource update` (exit **0**); the board copy and the `.research/` copy hash identically.

**No Go test, build, vet, or tooling command was executed in this cycle**, consistent with the task
boundary ("diagnose without running Go tests and without edits"). No candidate file was modified. A
live `pgrep` at 18:33 found three Go processes belonging to a sibling task, which is a second,
independent reason nothing was run.

---

## 1. Cycle-4 verdict R4-1 — disposition

The verdict was a single blocking finding with four required items. Each is addressed literally.

### Required item 1 — replace "carries no information" with a censored lower bound

**Done.** The reviewer's analysis is correct and is adopted without argument. Both timed-out packages'
same-run ratios are now presented as the **same kind of object** — a right-censored one-sided lower
bound on the completed factor — which is what removes the internal inconsistency the verdict named
(one row called a "genuine floor", the other called information-free).

| package | non-race (completed) | race (censored) | corrected statement |
| --- | ---: | ---: | --- |
| `internal/install` | 341.415s | FAIL 603.306s | completed factor **> ×1.7670**; censoring **moderate** (72 of 107 finished) — usable as a floor |
| `internal/install/atomicity` | 441.122s | FAIL 603.701s | completed factor **> ×1.3685**; censoring **severe** (binding sweep unfinished *plus* 3 tests never started) — **too weak** for cross-package comparison or for estimating the completed factor |

The phrase "carries no information" is withdrawn everywhere it appeared.

### Required item 2 — remove or qualify "purely due to truncation"

**Done.** "Purely" is withdrawn. The corrected wording states that the *observed* value is low
**because the numerator is truncated**, and that how far above ×1.3685 the completed factor actually
sits is **not derivable from this row**. No claim is made about atomicity's position relative to
`cmd/curator`. §3.4's phase-level solve (×3.33 sweep, ×4.14 activation prefix) is named as the source
of the completed in-gate factor, and is itself unchanged.

### Required item 3 — apply the same correction to the rework record and all current-summary passages

**Done.** A programmatic sweep found exactly **three** occurrences, all now corrected:

1. `.research/260729_install-race-timeouts.md` §3.5, the two timed-out-package bullets (rewritten).
2. The same file's "What changed in cycle 4" table row (corrected, and cross-referenced forward).
3. `TASK-260729-3dr6hw_cycle4-rework-record.md` §2 item 1 (corrected in place, with an explicit
   block-quoted cycle-5 amendment so the historical record stays readable rather than being silently
   rewritten).

§3.5's closing paragraph ("What §3.5 is for") also restated ×1.77 as a floor and is corrected for
consistency; it is inside §3.5, the section the verdict scoped for repair.

Verification that no fourth occurrence survives:

```
grep -n "no information\|carries no\|purely\|1\.37\|1\.77" .research/260729_install-race-timeouts.md
```

Real exit **0**, **8 hits**, every one accounted for and none an surviving instance of the defect:

| line | what it is |
| ---: | --- |
| 38 | cycle-5 intro, "a new, **purely** additive §10" — unrelated sense of the word |
| 46, 47, 49 | the cycle-5 change table, quoting the withdrawn claims in order to withdraw them |
| 114 | §1.1, "the gate is red **purely** on duration" — a correct statement about the timeout, not about any ratio |
| 526 | §3.5 bullet, "That is the correct statement — **not that the row carries no information**" — the explicit repudiation |
| 532, 534 | §3.5 rounding-convention paragraph, where ×1.37 / ×1.77 appear as roundings and are named as *not* valid bounds |

### Required item 4 — preserve §1, §2, §4 through §9, both allowlists, the tradeoff, §7 and §8

**Proven, not asserted.** Byte-level checks, each a standalone command with its real exit code:

| # | Check | Real exit | Result |
| ---: | --- | ---: | --- |
| 1 | `diff -u <cycle-4 board copy> <cycle-5 file>` | **1** | 6 hunks; exit 1 is `diff`'s "files differ" code and is the expected outcome, reported as-is |
| 2 | SHA-256 of old lines `56..451` vs new `80..475` (Provenance + §1 + §2 + §3.1–§3.4) | 0 | `56ad0a0d…eab4e` **both sides — identical** |
| 3 | SHA-256 of old lines `527..1337` vs new `576..1386` (§4 through §9, 811 lines) | 0 | `90741bc2…e0fc` **both sides — identical** |
| 4 | `shasum -a 256` on the board copy and the `.research/` copy after `resource update` | 0 | `5c2e89ba…d438d` both — in sync |

The six diff hunks are exactly: the revision line; the cycle-4 table row plus the new cycle-5 section;
the §3.5 heading; the two §3.5 bullets plus the rounding paragraph; the §3.5 closing paragraph; and
the appended §10. Nothing else moved.

---

## 2. Findings originated in cycle 5

### 2.1 Both rounded bounds were stated in the invalid direction

`603.306 / 341.415 = 1.767075…` and `603.701 / 441.122 = 1.368557…`. Rounded to two decimals these
are ×1.77 and ×1.37 — and **both roundings go up**. So neither "> ×1.77" nor "> ×1.37" is a strictly
true floor; each asserts slightly more than the arithmetic supports. This applies equally to the
verdict's own shorthand ">×1.37", which is honoured in substance and tightened in form.

The document now asserts bounds only as the **truncated** values **> ×1.7670** and **> ×1.3685**, and
carries an explicit rounding-convention note stating that ×1.77 and ×1.37 appear as roundings of the
quotient, never as the bound. Verified:

```
python3 -c "print(603.306/341.415 > 1.77, 603.701/441.122 > 1.37)"   →  False False
python3 -c "print(603.306/341.415 > 1.7670, 603.701/441.122 > 1.3685)" →  True True
```

### 2.2 §3.5 mis-stated the unstarted-test count for `internal/install`

The bullet read "cut off with **35** of 107 tests unstarted". The alarm shows test **73** in flight,
so **72 finished, 1 in flight, 34 never started**. Corrected inside the bullet being rewritten.

§3.2's `(107/72) × 600` is **unaffected and byte-unchanged**: it partitions 107 into 72 *finished* and
35 *unfinished*, which is correct under its own definition. Only the word "unstarted" was wrong.

Independently re-verified from candidate source (not taken from the prior cycle):

```
for f in $(ls internal/install/*_test.go | sort); do grep -hE '^func Test' "$f"; done | nl | grep TestStrictRegistryPolicyFailsUnknown
```
→ `73  TestStrictRegistryPolicyFailsUnknown`, out of **107** total. Real exit 0.

And for atomicity — 8 top-level tests, the sweep **5th**, therefore exactly **3** post-sweep tests,
which is what §1.3/§3.4 claim:

```
1 TestStableHybridActivationCommitsWithoutRestarting
2 TestHybridDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleContext
3 TestHybridDeclarationRetargetedBeforeHomeLockRestartsAndCommitsNoStaleContext
4 TestHybridDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure
5 TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder   ← the sweep
6 TestAdapterMirrorLinksAreJournaledAndRestoredExactly
7 TestStaleAdapterEntryIsRemovedBeforeTheConsumerLedger
8 TestStaleAdapterRemovalRollsBackToTheExactPriorEntry
```

### 2.3 The plan has been executed and measured — §9 risk 4 has fired (new §10)

This is the substantive finding of cycle 5 and the reason §10 exists.

**TASK-260729-rfrdfo** applied this diagnosis's Patch A + Patch B to a prototype worktree and is
running the §7 gate list now. Its driver implements §7 faithfully (exact `CURATOR_CONFORMANCE_ROOT`,
task-owned `GOTMPDIR`, no `-timeout`, no `./...`, one gate at a time behind a two-scan barrier, exit
code written last). The patch touches **exactly the 13 files** of the §4.0 allowlist, with
`aba_test.go` and `atomicity/fixture_test.go` untouched.

| Gate | Real exit | Package line |
| --- | ---: | --- |
| `gate-gofmt` | 0 | empty |
| `gate-vet` | 0 | empty |
| `gate-atomicity-structure` (non-race, `-v`) | 0 | `ok … atomicity 285.434s` |
| `gate-race-atomicity-1` | **0** | `ok … atomicity` **591.280s** |
| `gate-race-atomicity-2` | **0** | `ok … atomicity` **560.828s** |
| `gate-race-atomicity-3` | — | **in flight** — no `.exit` file; under that driver's protocol a missing `.exit` means "killed or still running", never "passed" |
| `gate-race-install-1..3` | — | **not started**; Patch A is unmeasured |

**The honest reading.** Patch B clears the unchanged 600s alarm — by **8.72s** and **39.17s** — but
**misses §7's own 480s pass condition** by **111.3s** and **80.8s**, and lands **21.9–28.5% above the
top** of §4.2's projected 340–460s band. These are *focused, single-package* numbers on an otherwise
idle machine; §3.2, §7 and §9 risk 3 all state that focused numbers read optimistically relative to
`./...`, where `cmd/curator` (557.779s under race) overlaps for nearly the whole run. An 8.72s margin
does not survive that.

That is exactly §9 risk 4, written before the fact: *"If the focused run lands above 480s, the next
lever is the `references/info.md` fixture reduction (§4.3) with a measurement attached, then the
production-side fix in §6 — not a timeout override."* §10.6 restates that ladder as the recommended
next step.

**Two projection errors, in opposite directions**, both now quantified in §10.4:

- The **race factor** assumption was too *pessimistic*: the completed patched pair gives the first
  **uncensored** atomicity factor, `591.280 / 285.434 =` **×2.07** and `560.828 / 285.434 =` **×1.97**,
  against §3.4's assumed ×3.33–4.27.
- The **non-race wall** model was far too *optimistic*: 441.122s → **285.434s** is only **−35.3%**,
  because Patch B gives each class its own `env` and twelve chains plus seven tests contend for
  12 performance cores. The verbose log shows per-class outer subtests at **196.25–247.78s** with
  inner injections of **68.50–133.91s**, against the modelled 37.28s injection and 40.74s
  baseline+final — roughly **×3 self-contention inflation per chain**. The package wall is bound by
  the longest chain (247.78s), not by the modelled 78.02s.

The two errors partly cancel, which is why the result clears 600s at all rather than by the projected
140–260s.

**§10 retracts nothing.** §4's patch is the same patch that was applied and does run green; §5, §7,
§8 and both allowlists are confirmed workable by a producer executing them end to end without hitting
a contradiction. The 480s bar stays — it was set to absorb exactly this gap and is doing its job by
refusing a result that would probably fail the real gate. §10 is separable and can be deleted without
disturbing the ordered repair.

---

## 3. Commands run in this cycle, with real exit codes

Every command is read-only and was run as a standalone process; no pipe chain hides an exit code.
**No `go` command of any kind was invoked.**

| # | Command | Real exit | Result |
| ---: | --- | ---: | --- |
| 1 | `task-board m 'set_status(TASK-260729-3dr6hw, status=analysis)'` | 0 | already `analysis` |
| 2 | `grep -n "no information\|carries no\|purely\|1\.37\|censor\|truncat" .research/260729_install-race-timeouts.md` | 0 | located exactly 3 occurrences to repair |
| 3 | `python3` quotient/bound check on both censored ratios | 0 | `>×1.37` and `>×1.77` both **False**; `>×1.3685` and `>×1.7670` both **True** |
| 4 | ordinal inventory of `internal/install/*_test.go` | 0 | `TestStrictRegistryPolicyFailsUnknown` = **73** of **107** |
| 5 | ordinal inventory of `internal/install/atomicity/*_test.go` | 0 | 8 tests, sweep 5th, 3 post-sweep |
| 6 | `grep -c "t.Parallel()" internal/install/*_test.go` | 0 | **0** in all 12 files — confirms Patch A's premise |
| 7 | `diff -u <cycle-4 board copy> <cycle-5 file>` | **1** | 6 hunks; exit 1 = "files differ", the expected outcome, reported as-is |
| 8 | SHA-256 range compare, Provenance + §1 + §2 + §3.1–3.4 | 0 | identical |
| 9 | SHA-256 range compare, §4 through §9 (811 lines) | 0 | identical |
| 10 | `cat` of `.temp/TASK-260729-rfrdfo/gates/*.exit`, `*.seconds`, `*.log`, `bin/run-gates.sh` | 0 | §10's table, quoted verbatim |
| 11 | `pgrep -af '(^\|/)(go\|.*\.test)( \|$)\|go-build\|cmd/curator'` | 0 | **three Go processes active** — the sibling producer's run; a second reason nothing was executed |
| 12 | `shasum -a 256` on `.research/` and board copies after `resource update` | 0 | `5c2e89ba…d438d` both |

Command 7 is the only non-zero exit. It is `diff`'s files-differ code, which is the required and
expected outcome of a rework cycle; it is not a failing gate and is not being reinterpreted as a pass.

---

## 4. Checklist disposition

Unchanged from cycles 3 and 4, restated so it is not silently re-litigated:

- **Item 15 (Tests green) stays unchecked.** No Go test, build or vet command has been run in any
  cycle of this task, and the task scope forbids running the full or race suite. Checking it would be
  a false green under the Evidence Honesty Contract. It belongs to the downstream producer task —
  and §10 now records that that producer's own atomicity race gates, while exiting 0, land
  **above** the §7 pass bar, so the item is not merely unverified here but actively **not yet
  satisfied** downstream.
- Items 13 and 14 remain checked as producer-side self-assessment. "Reviewer independently checks the
  plan" is by construction the next cycle's gate, tracked by item 16.
