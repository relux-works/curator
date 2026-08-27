# TASK-260729-3dr6hw — cycle 3 rework record

Date: 2026-07-29
Role: researcher
Routing in: `analysis` (cycle-2 verdict = CHANGES REQUESTED to analysis)
Routing out: `to-review`
Deliverable: `.research/260729_install-race-timeouts.md`,
SHA-256 `d2dc2cbc7e332c4590cf49b8b0d37e9753f497b1c177d371554262c73f940b6b`
Board resource: `TASK-260729-3dr6hw_install-race-timeout-diagnosis.md` (updated in place, not duplicated)

**No Go test, build, vet, or tooling command was executed in this cycle**, consistent with the task
boundary ("diagnose without running Go tests and without edits"). No candidate file was modified.
The only writes were to `.research/`, `.temp/TASK-260729-3dr6hw/`, and the board.

---

## 1. Cycle-2 required corrections — disposition

| # | Reviewer requirement | Independently verified? | Disposition |
| ---: | --- | --- | --- |
| 1 | §8.3 rsync expectation is impossible; correct to 11 new modified entries / 34 lines; name the SHA-256 manifest as the proof for the two candidate-only files | **Yes** — `candidate-source-delta-post.txt` is 23 lines and contains `*deleting internal/install/dryrun_conformance_test.go` and `*deleting internal/install/cache_conformance_test.go` | §8.3 rewritten. Expectation is now a 4-row table ending at 34 lines. A new paragraph states that §8.3 **cannot** prove anything about the two candidate-only files and that §8.2's manifest gate is their only proof. A further paragraph splits the 23 recorded digests into 21 that must be unchanged and 2 that **must** change because they are patch targets. |
| 2 | `saveJournal` call sites are 23 (16 + 7), not 24 | **Yes** — `grep -c 'saveJournal'` ⇒ `engine.go` 16 (exit 0), `staging.go` 7 (exit 0) | §2.2 now carries a per-file table totalling 23 and names cycle 2's error precisely (the arithmetic was wrong, the seven anchors were right). §6 corrected to 23. The two copy-loop sites `staging.go:141`/`:161` are retained in both places. |
| 3 | Patch B1 needs a literal injection selector | **Yes** — `scenario.classes` is read at `commit_atomicity_test.go:163` (injection loop) **and** `:206` (coverage assertion) | §4.2 replaced with B1.a–B1.e: new field `injectClasses []string`; `classes` untouched; `sharedUserHome bool` → `userHome string`; both constructors take `injectClass string`; the scenario table is **generated** by ranging over `projectSweepClasses`/`globalSweepClasses`, so full class coverage holds structurally. New §4.2.1 is the function-level identifier allowlist. |
| 4 | Resolve "no assertion removed" vs the cross-class-chain reduction | **Yes** — the chained baseline is described at `commit_atomicity_test.go:151-153`; the direct check is `before.diff(snapshotState(t, e))` at `:178` | The blanket claim is gone. New **§5.1** treats the cross-class residue chain as intentionally retired defence in depth: 31 ordered class pairs today → 0, what still proves the property directly, what is genuinely lost, the cycle-2 sanction, and a two-class fallback that restores 9 of 31 pairs at a cost the 480s bar probably cannot absorb. §4 preamble and §9 risk 5 restated to match. |

## 2. Findings originated in cycle 3 (not requested by the reviewer)

1. **Estimate C is weaker than cycle 2 presented it.** Its non-race denominator, `internal/install
   228.344s`, comes from `.temp/TASK-260720-2284br/gates-rework1/gate-gotest.log` — a 25-line
   truncated log whose line 21 is `FAIL github.com/relux-works/curator/internal/godriver 99.383s`
   (two subtests failing with `go-v1 process_timeout: Go probe exceeded its deadline`, a host-load
   symptom). Additionally `.temp/TASK-260720-2284br/worktree` has since been reset — its
   `internal/install` now holds 4 test files / 58 tests and it has **no `atomicity` package** — so
   the tree behind ×2.67 cannot be re-inspected. §3.2 now carries this caveat, §3.1 downgrades the
   row, and §9 risk 2 names it the weakest link. The patch is sized against Estimate B (1000s).
2. **New same-run corroboration of the mechanism (§3.5).** Inside the failing gate itself, the three
   large packages that completed on both sides read ×1.45 (`cmd/curator`), ×1.39
   (`internal/godriver`), ×1.25 (`internal/transaction`). `internal/install` cannot be under ×1.77
   and is really well above it. This imports no cross-tree assumption and corroborates §2.2's
   mechanism claim — subprocess-dominated packages barely move under `-race`, in-process
   CPU-dominated ones do. Sub-10s packages are excluded as noise (`internal/audit` reads ×4.66 on a
   1.4s baseline).
3. **No accepted gate filter breaks (§4.2.2).** Patch B renames sweep subtest paths from
   `…/project-hybrid-auto/50-env-file` to `…/project-hybrid-auto-50-env-file/50-env-file`. Verified:
   all four accepted filters in `.temp/TASK-260720-2284br/gates-cycle5/run-gates.sh:35-41` (`$R5`,
   `$REVALIDATION`, `$CONCURRENCY`, `$ACTIVATION`) contain **zero** `/` characters, so none names a
   subtest path and none is affected.
4. **Stronger safety argument for the 5 distinct global user homes.** Cycle 2 argued only that
   `globalbins.Select` *prefers* each scenario's own bin. Re-reading the source found a second,
   independent guarantee: the fallback `pathDirs` scan is gated by `safeExistingUserBin`
   (`globalbins.go:195`), whose first condition is `underHome(path, userHome, platform)` (`:196`) —
   a sibling's bin is rejected outright even if the preferred probe were to miss.
5. **Line-number corrections.** `env.userHome` is `fixture_test.go:30` (cycle 2 said :32);
   `globalbins.Select` is `globalbins.go:114` (said :113); the preferred loop is `:148-155` (said
   `:146-152`); `entryDigest` is `fixture_test.go:226`.
6. **88-name allowlist re-verified programmatically** (see §3 below).

## 3. Verification commands run in this cycle, with real exit codes

Every command below is read-only. Each was run as a standalone process; no pipe chain hides an exit
code.

| # | Command | Real exit | Result |
| ---: | --- | ---: | --- |
| 1 | `grep -c 'saveJournal' …/internal/transaction/engine.go` | 0 | `16` |
| 2 | `grep -c 'saveJournal' …/internal/transaction/staging.go` | 0 | `7` — total 23, confirming reviewer item 2 |
| 3 | `wc -l < …/verifier3/candidate-source-delta-post.txt` | 0 | `23` |
| 4 | `grep -a "install" …/verifier3/candidate-delta-digests-post.txt` | 0 | both candidate-only install tests carry digests |
| 5 | `grep -h -c '^func Test' …/internal/install/*_test.go` (summed) | 0 | `107`, matching the alarm dump's `0x6b` |
| 6 | `grep -h -c '^func Test' …/internal/install/atomicity/*_test.go` (summed) | 0 | `8`, matching the alarm dump's `0x8` |
| 7 | `comm -23 excl-sorted.txt all-tests.txt` | 0 | empty — all 19 exclusions exist in the package |
| 8 | `comm -23 all-tests.txt excl-sorted.txt \| wc -l` | 0 | `88` |
| 9 | `comm -3 doc-88.txt parallel-88.txt` | 0 | **empty both directions** — the §4.1 allowlist is exactly `{all 107} \ {the 19}` |
| 10 | per-variable slash count over `run-gates.sh` filter lines | 0 | `R5=0 REVALIDATION=0 CONCURRENCY=0 ACTIVATION=0` |
| 11 | `sed -n '108,165p'`, `'193,215p'` on `internal/globalbins/globalbins.go` | 0 | `Select` at :114, preferred loop :148-155, `safeExistingUserBin`/`underHome` at :195-196 |
| 12 | `shasum -a 256 .research/260729_install-race-timeouts.md` | 0 | `d2dc2cbc7e332c4590cf49b8b0d37e9753f497b1c177d371554262c73f940b6b` |

**No expected-red gate was run in this cycle, so none is being presented as green.** No `go` command
of any kind was invoked.

Scratch artifacts backing gates 7–9 are retained at `.temp/TASK-260729-3dr6hw/`:
`all-tests.txt` (107), `exclusions.txt` / `excl-sorted.txt` (19), `parallel-88.txt` (88, derived from
source), `doc-88.txt` (88, extracted from the document).

## 4. Checklist disposition

Item 15 (**Tests green**) is left **unchecked**, deliberately. No Go test, build or vet command was
run in any cycle of this task, and the task scope explicitly forbids running the full or race suite.
There is no green command behind that item, so checking it would be a false green under the Evidence
Honesty Contract. Item 15 belongs to the downstream producer task that applies Patch A and Patch B
and runs the §7 focused commands.

Items 13 and 14 are checked as the producer-side self-assessment they are: every producer-owned
clause of the AC is satisfied (exact failing tests and timings recorded; every proposed optimisation
mapped to preserved assertions and isolation invariants, with the one retired invariant named;
smallest test-only patch with a literal file and function allowlist and quantified savings; focused
validation commands that do not overlap the active verifier; candidate-integrity checks). The AC's
final clause — *"reviewer independently checks the plan"* — is by construction the **next** cycle's
gate and is tracked by item 16, not self-certified here.

## 5. Unchanged from cycle 2 and preserved verbatim

The exact timeout inventory (§1), hot-path evidence (§2.1–2.5), the 88/19 parallelisation split with
its three hazard classes (§4.1), the 13-file required allowlist with `aba_test.go` and
`atomicity/fixture_test.go` explicitly excluded (§4.0), the atomicity partition arithmetic (§4.2), the
focused validation commands and 480s pass bar (§7), the pre/post SHA-256 manifest gate (§8.1–8.2),
the immutable-conformance-root invariant (§8.4), and the out-of-scope production-side finding routed
as a separate story (§6) are all carried forward unchanged.
