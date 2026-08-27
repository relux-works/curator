# TASK-260729-rfrdfo — prototype results: install/atomicity race-timeout test-only patch

**Status: prototype complete, all 13 focused gates real exit 0, ready for review.**
**Headline finding: Patch A clears §7's bar with a wide margin. Patch B passes but misses §7's 480s
pass condition on all three race runs. §9 risk 4 has fired. The next lever (§4.3) is outside this
task's 13-file allowlist, so it is routed, not taken.**

---

## 1. What was built

A byte-for-byte private copy of the TASK-260720-jrrgw9 candidate at
`.temp/TASK-260729-rfrdfo/worktree`, carrying TASK-260729-3dr6hw revision 3's Patch A + Patch B and
nothing else. The source candidate was never written to.

Deliverables in `.temp/TASK-260729-rfrdfo/`:

| Path | What it is |
| --- | --- |
| `TASK-260729-rfrdfo_install-race-timeout.patch` | 13-file unified diff, `-p1`, applies to jrrgw9 |
| `worktree/` | the patched prototype the gates ran against |
| `bin/` | `manifest.sh`, `integrity.py`, `accepted-delta.sh`, `hazards.py`, `patch_a.py`, `barrier.sh`, `run-gates.sh`, `report.sh` |
| `gates/` | per-gate `.log` / `.exit` / `.seconds` / `.barrier`, plus `DRIVER-START` / `DRIVER-DONE` |
| `evidence/` | manifests, integrity output, rsync deltas, conformance digests, reconciliation notes |

---

## 2. Focused gate results — every real exit code

Driver window `2026-07-29T17:58:17` → `2026-07-29T18:46:08`. Every gate is a standalone process
behind a two-scan `pgrep` barrier; the exit code is written last, no `tee`, no pipe chain. No gate
carries a `-timeout` token, no `./...`, no `-short`, no case-dropping `-run`, no `-count>1`.

| gate | real exit | wall s | package line |
| --- | ---: | ---: | --- |
| `gate-gofmt` | 0 | 0 | (empty) |
| `gate-vet` (`go vet ./internal/install/...`) | 0 | 0 | (empty) |
| `gate-atomicity-structure` (`-count=1 -v`, non-race) | 0 | 286 | `ok … internal/install/atomicity 285.434s` |
| `gate-race-atomicity-1` | 0 | 593 | `ok … internal/install/atomicity 591.280s` |
| `gate-race-atomicity-2` | 0 | 561 | `ok … internal/install/atomicity 560.828s` |
| `gate-race-atomicity-3` | 0 | 564 | `ok … internal/install/atomicity 564.022s` |
| `gate-race-install-1` | 0 | 234 | `ok … internal/install 232.088s` |
| `gate-race-install-2` | 0 | 235 | `ok … internal/install 235.124s` |
| `gate-race-install-3` | 0 | 227 | `ok … internal/install 226.191s` |
| `gate-race-r5` (`-run "$R5"`) | 0 | 46 | `ok … internal/install 45.432s` |
| `gate-race-revalidation` | 0 | 40 | `ok … internal/install 39.959s` |
| `gate-race-concurrency` | 0 | 19 | `ok … internal/install 19.248s` |
| `gate-race-activation` | 0 | 37 | `ok … internal/install/atomicity 36.499s` |

`grep -l 'DATA RACE' gates/*.log` → **none**. §9 risk 1 did not fire: 88 newly-concurrent tests
produced no race report in three repeats.

The four regression filters were byte-compared against the accepted driver
(`.temp/TASK-260720-2284br/gates-cycle5/run-gates.sh`): `R5`, `REVALIDATION`, `CONCURRENCY`,
`ACTIVATION` are all **IDENTICAL** — reused literally, not retyped.

### 2.1 Against §7's pass condition — the honest reading

§7 requires real exit 0 **and** package duration ≤ **480s**.

| package | run 1 | run 2 | run 3 | vs 480s bar | vs 600s alarm |
| --- | ---: | ---: | ---: | --- | --- |
| `internal/install` (Patch A) | 232.088s | 235.124s | 226.191s | **PASS**, margin 245–254s | margin 365–374s |
| `internal/install/atomicity` (Patch B) | 591.280s | 560.828s | 564.022s | **MISS** by 111.28 / 80.83 / 84.02s | margin 8.72 / 39.17 / 35.98s |

**Patch A works and is not the problem.** Its projected worst plausible case (§4.1's 2× row,
510–558s) was pessimistic by more than a factor of two; the measured 226–235s beats even the §4.1
12× row (188–196s territory) once the sequential tail is priced correctly.

**Patch B works and is not enough.** Three green runs, all above 480s, all within 40s of the alarm —
and these are focused single-package numbers on an idle host. §3.2, §7 and §9 risk 3 all state that
focused numbers read optimistically against the `./...` gate, where `cmd/curator` (557.779s under
race) overlaps for nearly the whole run. An 8.72s margin does not survive that.

### 2.2 Measured race factor for atomicity — first uncensored same-package pair

Same tree, same patch, non-race `285.434s`:

| run | race s | factor |
| --- | ---: | ---: |
| 1 | 591.280 | ×2.071 |
| 2 | 560.828 | ×1.965 |
| 3 | 564.022 | ×1.976 |

§3.4/§4.2 assumed ×3.33–4.27. The assumption was **too pessimistic**; the non-race baseline model
was **too optimistic**. The errors partly cancel, which is the only reason the result clears 600s at
all. (The −35.3% non-race comparison against a 441.122s pre-patch figure is quoted from diagnosis
§10.4; this task did not itself measure a pre-patch non-race atomicity run.)

---

## 3. Integrity evidence — all five AC clauses

All commands below were re-run at handoff time, after the gates finished, and their real exit codes
are quoted.

### 3.1 Only the 13 allowlisted files changed (§8.2)

`bin/integrity.py evidence/manifest-pre.txt evidence/manifest-final.txt` → **`integrity exit=0`**

```
modified=13 added=0 deleted=0 unexpected=0 forbidden_touched=0 not_modified_from_allowlist=0
INTEGRITY_OK
```

All 13 modified paths are the §4.0 allowlist. `internal/install/aba_test.go` and
`internal/install/atomicity/fixture_test.go` appear in **neither** the modified nor the added list.
Zero added, zero deleted. Manifest size 391 files, path-sorted SHA-256, `.git` excluded, `pipefail`
set so the recorded exit is the chain's.

The prototype's pre-patch manifest is byte-identical to the pristine candidate's
(`diff evidence/manifest-pre.txt evidence/source-candidate-manifest-pre.txt` → **exit 0**), so the
copy really was byte-for-byte before the first edit.

### 3.2 Prototype unperturbed by the gate runs

`diff evidence/manifest-post.txt evidence/manifest-final.txt` → **exit 0**. `GOTMPDIR` is task-owned
(`.temp/TASK-260729-rfrdfo/gotmp/{install,atomicity}`) and the build cache is outside the worktree,
so twelve Go invocations left no trace in the tree. §8.1's "if it does, that is itself a finding"
did not trigger.

### 3.3 Patch applies cleanly to the current jrrgw9 candidate, without mutating it

```
patch -p1 -d <jrrgw9-worktree> --dry-run --forward < TASK-260729-rfrdfo_install-race-timeout.patch
```
→ **`patch-dry-run exit=0`**, 13 `patching file …` lines, **no fuzz, no offset, no reject**
(`evidence/patch-dry-run-final.log`).

Immediately after, `bin/manifest.sh` over the candidate →
`diff source-candidate-manifest-pre.txt source-candidate-manifest-final2.txt` → **exit 0**.

Independently, `evidence/manifest-applycheck.txt` (patch applied to a third, throwaway copy) is
byte-identical to `evidence/manifest-post.txt` → **exit 0**: the patch reproduces the exact tree the
gates were run against, file for file.

### 3.4 Main candidate byte-identical

`diff evidence/source-candidate-manifest-pre.txt evidence/source-candidate-manifest-final.txt` →
**`source-candidate-unchanged exit=0`**, 391 files both sides. No commit, no stash, no checkout, no
index refresh — see reconciliation note 2 on why `git diff --check` was deliberately skipped.

### 3.5 Accepted-worktree delta (§8.3) — the corrected 34-line expectation holds

`bin/accepted-delta.sh` reproduces verifier-3's `candidate-source-delta-post.txt` **byte for byte**
against the pre-patch tree (**exit 0**). Flags had to be reconciled first — see reconciliation
note 1; the prescribed flags emit 119 lines, not 23.

| group | expected | measured |
| --- | ---: | ---: |
| candidate-only (`*deleting`) | 20 | **20** |
| already-modified (`>fcsT....`) | 3 | **3** — `cmd/curator/status_test.go`, `internal/buildcache/conformance_test.go`, `internal/closure/conformance_test.go` |
| newly modified (`>fcsT....`) | 11 | **11** |
| total lines | 34 | **34** (was 23) |

The 11 are exactly the allowlist minus the two candidate-only conformance tests, as §8.3 predicts.

Digest check against `candidate-delta-digests-post.txt`: **21 identical, 2 changed, 0 missing,
`DIGEST_OK`, exit 0**. The 2 changed are precisely
`internal/install/cache_conformance_test.go` and `internal/install/dryrun_conformance_test.go` — the
two patch targets §8.3 says this check cannot validate. `cmd/curator/status_test.go` still hashes
`487b12bd…0d32e`.

### 3.6 Immutable conformance root (§8.4)

448-file digest sweep over `.temp/TASK-260729-3nx97g/worktree/conformance/v1`, diffed against
verifier-3's `authoritative-digests-post.txt` → **`conformance-immutable exit=0`**, 448 lines both
sides. No vector moved.

---

## 4. Semantic checks on the patch itself (§8.5)

| # | check | result |
| --- | --- | --- |
| 1 | `t.Parallel()` sums to 88 in `internal/install/*_test.go` | **88** — `aba_test.go` contributes 0 |
| 1b | marked set == §4.1 allowlist | derived `{107 declared} \ {88 marked}` = **exactly the 19 sanctioned exclusions**, no extra, none smuggled |
| 2 | no added skip | skip-string sets pre vs post **identical, exit 0**; `commit_atomicity_test.go:292` guard survives |
| 3 | no `-timeout` token | grep **empty**, exit 1 |
| 4 | assertion counts unchanged | all 13 files **OK**, pre == post (e.g. `stage_test.go` 172/172, `commit_test.go` 109/109, `commit_atomicity_test.go` 41/41) |
| 5 | selector split | `scenario.injectClasses` appears **once**, at the injection loop; `scenario.classes` appears **once** in the body, at the post-sweep coverage assertion, plus once per constructor; `sharedUserHome` appears **nowhere** |
| 5b | coverage cannot drift | table is generated by ranging over `projectSweepClasses` / `globalSweepClasses` — union of injected classes equals the full lists **structurally** |
| 6 | user homes | `len(globalUserHomes) == len(globalSweepClasses)`, each from its own `t.TempDir()`, all `.local/bin` on one `PATH` set in the sequential parent; `e.userHome` assigned only from `scenario.userHome` |
| 7 | identifier set | see §4.1 below — matches §4.2.1 with three recorded naming deviations |
| — | atomicity split | the 7 §4.2-B2 tests are parallel; `TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` **stays sequential** (its `t.Setenv("PATH", …)` would panic) |
| — | gofmt | `gofmt -l internal/install internal/install/atomicity` → **empty, exit 0** |
| — | vet | `go vet ./internal/install/...` → **empty, exit 0** |

### 4.1 Deviations from §4.2.1's literal identifier table

Three, all recorded rather than silent:

1. **`projectSweepScenario(name, mode string, injectClasses ...string)`** and
   **`globalSweepScenario(name, mode, userHome string, injectClasses ...string)`** — variadic slice,
   not the `injectClass string` scalar in §4.2.1. This is cycle-2 reviewer correction 3: the
   selector must stay a genuine slice so a future entry can inject more than one class without
   re-merging the two roles. `injectClasses:` is assigned only from that parameter, never from
   `projectSweepClasses` / `globalSweepClasses` directly.
2. **Local named `scenarios`, not `sweep`.** Same role, same construction.
3. **No `userBins` local.** The `PATH` string is accumulated directly, so the intermediate slice
   §4.2.1 anticipated is unnecessary. Nothing else new is introduced.

Everything else matches: `injectClasses []string` added, `userHome string` added replacing
`sharedUserHome bool` removed, `globalUserHomes []string` added. `projectSweepClasses`,
`globalSweepClasses`, `commitProbe` and every `fixture_test.go` helper are untouched.

### 4.2 Scenario naming — deviation from §4.2.2, and why it is the safer one

§4.2.2 predicted `project-hybrid-auto-<class>`. The implementation uses
`project-hybrid-auto/<class>`, so the full path is
`…/project-hybrid-auto/10-context/10-context` rather than `…/project-hybrid-auto-10-context/…`.

The doubled segment is the honest consequence of a length-one chain that keeps its inner per-class
`t.Run` (which carries the `passed`/`break` semantics). The slash form **preserves the existing
`…/project-hybrid-auto/` and `…/global-auto/` path prefixes**, so it is strictly less likely to
break a future subtest-path filter than the hyphen form would have been. The four accepted
regression filters contain only top-level test names and are unaffected either way — verified, and
`gate-race-activation` passing at exit 0 confirms it empirically.

---

## 5. Assertion retention and the one retired invariant

**Retained, and still asserted once per class, 12 times total** — verbatim from the patched source:

- `result.Status == "failed"` per injection
- `before.diff(snapshotState(t, e))` — the whole-state digest comparison, after **every** injection
- the committed-class cutoff (no class committed after the failing one)
- `assertReverseRollback`
- `assertNoJournalRemains`
- the post-sweep success + full-`scenario.classes` coverage + non-decreasing class order — which now
  runs **7× (project) and 5× (global)** instead of once per scope, i.e. **strictly more often than
  before**

Assertion counts are numerically unchanged in all 13 files (§4 check 4). No assertion was deleted
from any test body. No case is skipped. No timeout is changed. No production file is touched.

**Retired, deliberately and with cycle-2 reviewer sanction: the cross-class residue chain.**
Previously all seven project injections ran against one shared baseline, exercising 21 ordered
(X-then-Y) class pairs, plus 10 for the 5-long global chain — **31 ordered pairs → 0**. The comment
that justified the chain (*"any residue one rollback leaves behind shows up as a failure of the
injection after it"*) has been replaced in-source by a comment naming the trade explicitly, so the
retirement is visible at the call site and not only in this document.

Why it is defence in depth and not the primary check: residue is asserted **directly**, on every
injection, by the whole-state digest at the per-injection comparison — which observes the same paths
through the same `entryDigest` the chain could. What is genuinely lost is a defect that both escapes
`snapshotState`'s hand-maintained path list **and** only manifests through a later class. Narrow,
non-empty, and stated as a real reduction rather than a re-labelling. Full treatment: diagnosis §5.1
and §9 risk 5.

Global-scope safety under 5 distinct user homes: the `sweepScenario` invariant *"two scenarios must
never share it: their snapshots would overlap"* is satisfied **more strictly** than before, since no
two entries share a home at all. `globalbins.Select` prefers each entry's own `<userHome>/.local/bin`
on the first probe, and its fallback scan is gated by `safeExistingUserBin` → `underHome`, which
rejects a sibling's bin outright.

---

## 6. What this prototype does NOT establish, and what must happen next

1. **The `./...` gate is unmeasured.** §7 forbids this task from running it. Patch B raises
   atomicity's total CPU by ~125%, so the whole-suite wall clock should be expected to move and must
   be reported by the verifier rather than treated as noise (§9 risk 7).
2. **§9 risk 4 has fired and its ladder is the sanctioned response.** All three atomicity race runs
   landed above 480s. The written next lever is **§4.3** — dropping the unasserted
   `references/info.md` write from `atomicity/fixture_test.go` — *with a before/after `StagingEntries`
   measurement attached*. That makes `fixture_test.go` a declared, justified **14th** file, which
   this task's scope explicitly forbids ("Do not touch … `atomicity/fixture_test.go`"; "modify only
   the 13-file required allowlist"). **It was therefore not attempted here.** It needs a separate
   task with a 14-file allowlist, not a quiet scope expansion inside this one.
3. **If §4.3 is insufficient, escalate to §6** — the production-side `saveJournal` / O(P²) namespace
   revalidation fix — as its own story. Per §10.5, no `-timeout` override, skip, or assertion
   weakening becomes acceptable merely because the margin is thin. That ladder is a human decision
   about what the suite is for.
4. **The 480s bar is doing its job.** It was set to absorb exactly the focused-versus-`./...` gap,
   and it is refusing a result that would probably fail the real gate. Recommending its removal is
   not this task's call.

---

## 7. Evidence-honesty statement

Every number in this document is either read from a `gates/gate-*.{log,exit,seconds}` file written
by `bin/run-gates.sh`, or is the real exit code of a command this task ran directly at handoff time
as a standalone process. No gate was piped through `tee`. No gate carries a `-timeout` token. No
`go test ./...`, no coverage run, no host install, no stage, no commit, no publish, no pin, no
Windows execution was performed. The process barrier
(`pgrep -af '(^|/)(go|.*\.test)( |$)|go-build|cmd/curator'`) was empty before each measurement;
`gates/gate-*.barrier` records it per gate.

Three atomicity race runs are reported as **passing exit 0 and missing the 480s pass condition** —
both facts, together. They are not presented as a clean pass, and the 8.72s margin on run 1 is not
presented as a solved race gate.

The single number in this document that this task did not measure is the 441.122s pre-patch non-race
atomicity baseline in §2.2, quoted from diagnosis §10.4 and labelled as such.
