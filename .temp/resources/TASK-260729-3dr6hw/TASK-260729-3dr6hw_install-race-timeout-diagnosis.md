# TASK-260729-3dr6hw — diagnosis of the `internal/install` and `internal/install/atomicity` race timeouts

Date: 2026-07-29
Role: researcher
Revision: **cycle 5** (addresses reviewer verdict cycle 4)
Verdict: **test-only rework is viable**; the smallest patch is specified below.

Method: read-only static inventory, the verifier-3 evidence, and prior first-hand gate logs already
present in `.temp/`. **No Go test, build, vet, or tooling command was executed for this diagnosis**
(task boundary: "diagnose without running Go tests and without edits"). A `go test -count=1 ./...`
belonging to another agent was observed active during this analysis (`ps` PID 1686), which is a
second, independent reason nothing was run. Every number below is either quoted from an evidence
file or derived arithmetically from quoted numbers; each derivation states its inputs and its
assumptions.

## What changed in cycle 4

Cycle 4 is a bounded evidence repair of **§3.5** plus the two places elsewhere in the document that
carried the same unmeasured causal claim. The plan (§4), the assertion matrix (§5), the validation
commands (§7), the integrity gates (§8), the allowlists, and every timing band are **unchanged**.

| Reviewer item (cycle 3) | Resolution |
| --- | --- |
| 1 — §3.5 claims to list "every" large completed package but omits `internal/runtimestore` | Fixed. §3.5 now states a **mechanical** selection rule — completed on both sides, `max(non-race, race) ≥ 10s` — which yields exactly **five** packages, and all five are listed: `cmd/curator` ×1.45, `internal/godriver` ×1.39, `internal/transaction` ×1.25, `internal/runtimestore` **×0.98**, `internal/managerlock` **×2.90**. The word "every" is withdrawn from the ×1.25–1.45 claim, which was **false**: the completed-package factors under that rule span ×0.98–×2.90. |
| 2 — §3.5 attributes `internal/transaction`'s low factor to `git` / `go build` subprocesses; static source contradicts this | Fixed. The subprocess mechanism claim is **withdrawn entirely**. Verified against source: `internal/transaction` has 50 tests, ~33 `Prepare` call sites, exactly **2** `exec.Command` sites — both re-executing the test binary (`os.Args[0] -test.run=^TestTransactionCrashHelper$`, `subprocess_test.go:56,459`) — and **zero** `"git"` or `go build` invocations. It is an in-process package at ×1.25, which directly refutes cycle 3's "in-process ⇒ large inflation" framing. §3.5 is now explicitly **descriptive**; the workload-shape discussion the reviewer permitted (journal target count `P`) is labelled a hypothesis consistent with the captured stacks and static source, **not** a measured attribution. |
| — (consistency, cycle 4) | The same withdrawn claim appeared in two other places and is repaired there too: §2.2's closing clause and §3.3's closing clause both said atomicity inflates more because `internal/install` spends its time in `git` subprocesses "which `-race` does not slow". Both now state the target-count difference without the unmeasured subprocess attribution. §3.1's status table row for the mechanism is downgraded from **measured** to **hypothesis**. |
| — (new, found in cycle 4) | `internal/install/atomicity`'s raw same-run ratio is `603.701 / 441.122 =` **×1.3686**, which is *lower* than `cmd/curator`'s ×1.45. That ratio is **censored** by the alarm; §3.5 now says so instead of leaving a reader to infer that atomicity inflates less than `cmd/curator`. Atomicity's in-gate factor comes from §3.4's phase-level solve, not from this table. **Cycle 5 corrects how cycle 4 characterised it** — see the cycle-5 table below. |

## What changed in cycle 5

Cycle 5 does two things, and keeps them strictly separate.

1. **The bounded evidence repair the verdict ordered** — the characterisation of the censored
   atomicity ratio, in the three places it appears. **Nothing else is modified**: §1, §2, §4 through
   §9, both allowlists, the assertion/invariant tradeoff, the focused validation commands, the
   candidate-integrity checks, and every timing band are byte-unchanged apart from the two §3.5
   bullets and the §3.5 closing paragraph named below.
2. **A new, purely additive §10** recording evidence that did not exist when cycles 1–4 were written:
   a sibling producer task has since **applied this plan's Patch A + Patch B and measured it**, and
   §9 risk 4 has fired. §10 retracts nothing and modifies nothing; it is separable and can be deleted
   without touching the ordered repair. It is included because withholding a measured result that
   contradicts §4.2's projected band would fail this task's fact-checking acceptance criterion.

| Reviewer item (cycle 4) | Resolution |
| --- | --- |
| R4-1 — the censored atomicity ratio is mischaracterised as carrying **no information**, while the analogous censored `internal/install` ratio is treated as a genuine floor; the two readings are internally inconsistent | Fixed. Both censored ratios are now stated as the **same kind of object**: a one-sided (right-censored) lower bound on the completed factor. Atomicity is `> ×1.3685`; `internal/install` is `> ×1.7670`. The phrase "carries no information" is withdrawn everywhere. What is now said instead: the atomicity bound is **too weak to support cross-package comparison or to estimate the completed factor**, because the censoring is severe — the binding sweep chain was unfinished and three post-sweep tests had not started. |
| R4-1, required item 2 — remove or qualify "its position below `cmd/curator` is *purely* due to truncation" | Fixed. "Purely" is withdrawn. The corrected statement: the *observed* value is low **because the numerator is truncated**, and the *completed* factor is not derivable from this row at all — it comes from the separate phase-level solve in §3.4 (×3.33 sweep, ×4.14 activation prefix). No claim is made about where the completed factor would sit relative to `cmd/curator`. |
| R4-1, required item 3 — apply the same correction to the cycle-4 rework record and all current-summary passages | Done. The cycle-4 rework record (`TASK-260729-3dr6hw_cycle4-rework-record.md`) §2 item 1 is corrected in place with an explicit cycle-5 amendment note, and the cycle-4 change-table row above is corrected. Those are the only three occurrences; see the cycle-5 verification table in the rework record. |
| — (new, found in cycle 5 while applying the fix) | **Both rounded bounds were stated in the direction that makes them invalid.** `603.701 / 441.122 = 1.368557…` and `603.306 / 341.415 = 1.767075…`. Rounded to two decimals these are ×1.37 and ×1.77 — but both roundings go **up**, so neither "> ×1.37" nor "> ×1.77" is a strictly true floor. The document now carries the truncated values **> ×1.3685** and **> ×1.7670**, which are strictly true, and states the 2-dp figures as roundings rather than as bounds. The verdict's own shorthand ">×1.37" is honoured in substance and tightened in form. |
| — (new, additive, cycle 5) | **§10 added.** TASK-260729-rfrdfo applied Patch A + Patch B to a prototype worktree and ran the §7 gates. Atomicity under race: **exit 0** at **591.280s** and **560.828s** — clearing the unchanged 600s alarm by 8.72s / 39.17s, but **missing §7's 480s pass bar** by 111.3s / 80.8s and landing 22–29% above §4.2's projected 340–460s band. The `internal/install` race gates have not run. §9 risk 4's ladder (§4.3, then §6) is the sanctioned response. Also yields the first **uncensored** atomicity race factor: ×2.07 / ×1.97, well below the ×3.33–4.27 §3.4 assumed. |
| — (new, found in cycle 5) | §3.5's `internal/install` bullet said the package "was cut off with **35** of 107 tests unstarted". The alarm shows test **73** in flight, so 72 finished, 1 in flight, **34** never started. Corrected in the bullet being rewritten. §3.2's `(107/72) × 600` is **unaffected** — it partitions 107 into 72 *finished* and 35 *unfinished*, which is correct under its own definition and is left byte-unchanged. |

## What changed in cycle 3

| Reviewer item (cycle 2) | Resolution |
| --- | --- |
| 1 — §8.3 rsync expectation is impossible | Fixed. `cache_conformance_test.go` and `dryrun_conformance_test.go` are already `*deleting` in the 23-line delta; modifying them adds nothing. Corrected to **11** new `>fcsT....` entries ⇒ **34** lines. The pre/post SHA-256 manifest (§8.2) is now stated as the *only* integrity proof for those two candidate-only files. |
| 2 — `saveJournal` call sites miscounted | Fixed and re-verified from source: **16** in `engine.go` + **7** in `staging.go` = **23**, not 24. Corrected in §2.2 and §6. The two copy-loop sites (`staging.go:141`, `:161`) are retained. |
| 3 — Patch B1 is inconsistent without an injection selector | Fixed. §4.2 now specifies the literal struct/signature change: new field `injectClasses []string`, `classes` untouched, `sharedUserHome bool` → `userHome string`, and the constructors take an `injectClass string`. New identifiers are listed in a function-level allowlist (§4.2.1). |
| 4 — "no assertion removed" contradicts the cross-class-chain reduction | Fixed. The cross-class sequencing check is now described throughout as **intentionally retired defence in depth**, sanctioned by the cycle-2 reviewer. The blanket "no assertion removed" claim is gone. |
| — (new, found in cycle 3) | Estimate C's honesty caveat: its non-race denominator comes from a log that also recorded a real `internal/godriver` FAIL, and the tree it measured can no longer be inspected. Confidence downgraded; see §3.2. |
| — (new, found in cycle 3) | Same-run corroboration of the §2.2 *mechanism* added from the failing gate's own completed packages. See §3.5. |
| — (new, found in cycle 3) | The four accepted `-run` filters are all top-level test names, so Patch B's subtest renaming breaks none of them. Verified; see §4.2.2. |
| — (new, found in cycle 3) | Line-number fixes: `env.userHome` is `fixture_test.go:30` (not :32); `globalbins.Select` is `globalbins.go:114` (not :113), preferred loop `:148-155`. |
| — (new, found in cycle 3) | The 88-name allowlist was re-verified programmatically against the candidate: it is exactly `{all 107} \ {the 19}`, symmetric difference empty in both directions. See §4.1. |

## What changed in cycle 2

| Reviewer item | Resolution |
| --- | --- |
| 1 — HEAD-relative integrity gate cannot work on a dirty delivery worktree | Replaced entirely. §8 now specifies a pre/post path-sorted SHA-256 manifest of the candidate worktree plus a regenerated accepted-worktree delta. |
| 2 — literal baseline commands, including added/deleted-path detection | §8.1–§8.3 give runnable commands with real exit codes and an explicit added / deleted / modified split. |
| 3 — tighten the required file allowlist | §4.0 now lists **13** required files. `internal/install/aba_test.go` and `internal/install/atomicity/fixture_test.go` are explicitly **not** patch targets. |
| 4 — the two `internal/install` projections are algebraically identical, and contention was ignored | §3 rewritten. The uniform-cost extrapolation is presented as **one** assumption-based estimate; genuinely independent corroboration now comes from previously unused first-hand race gate logs; cross-package contention is modelled explicitly. |
| — (new, found in cycle 2) | The alarm ordinal was wrong: `TestStrictRegistryPolicyFailsUnknown` is test **73** of 107, not test 71. Corrected in §1 and §3. |
| — (new, found in cycle 2) | Cycle 1's §9 risk 4 ("class-independent injection cost is an inference, not a measurement") is now **measured** — see §2.4. |
| — (new, found in cycle 2) | Cycle 1's claim that both packages share one race factor is **wrong**; measured factors differ materially. See §3.3. |
| — (new, found in cycle 2) | `internal/install/atomicity` is **not** in the candidate delta. Its race overrun is pre-existing debt, not a candidate regression. See §1.4. |

## Provenance of inputs

| Input | Path |
| --- | --- |
| Candidate (read-only) | `.temp/TASK-260720-jrrgw9/worktree` (HEAD `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`) |
| Accepted comparison worktree | `.temp/TASK-260729-2kaopg/worktree` |
| Immutable rc.5 conformance root | `.temp/TASK-260729-3nx97g/worktree/conformance/v1` |
| Race log (the failing gate) | `.temp/TASK-260720-jrrgw9/verifier3/go-test-race-all.log` (320 lines) |
| Full non-race log | `.temp/TASK-260720-jrrgw9/verifier3/go-test-all.log` (40 lines) |
| Verifier-3 report | `.temp/TASK-260720-jrrgw9/verifier3/TASK-260720-jrrgw9_final-verifier3-results.md` |
| Candidate delta artifacts | `.temp/TASK-260720-jrrgw9/verifier3/candidate-source-delta-post.txt`, `candidate-delta-digests-post.txt` (23 lines each) |
| **Prior first-hand race gate (raised timeout)** | `.temp/TASK-260720-2284br/gates-rework1/gate-race.log` + `gate-race-exit.txt` |
| **Prior first-hand non-race gate, same cycle** | `.temp/TASK-260720-2284br/gates-rework1/gate-gotest.log` |
| **Prior verbose per-subtest run** | `.temp/TASK-260720-2284br/gates-rework1/gate-acceptance-verbose.log` |
| **Verbose per-subtest run, current tree shape** | `.temp/logwork/TASK-260729-2kaopg/-tester--tester--codex-_RUN-260729-051449.log:4697-4765` |
| Prior focused gate driver (exact `-run` filters) | `.temp/TASK-260720-2284br/gates-cycle5/run-gates.sh` |
| Host | `sysctl hw.ncpu` = **16 logical**, of which `hw.perflevel0.logicalcpu` = **12 performance** and `hw.perflevel1.logicalcpu` = **4 efficiency**. Go 1.25.5 darwin/arm64. |

---

## 1. Exact failing tests and timings from verifier-3 evidence

### 1.1 Package results (quoted verbatim)

| Package | non-race (`go test -count=1 ./...`) | race (`go test -count=1 -race ./...`) |
| --- | ---: | ---: |
| `cmd/curator` | `ok … 384.270s` | `ok … 557.779s` |
| `internal/godriver` | `ok … 55.306s` | `ok … 76.752s` |
| `internal/transaction` | `ok … 27.218s` | `ok … 34.106s` |
| **`internal/install`** | `ok … 341.415s` | **`FAIL … 603.306s`** |
| **`internal/install/atomicity`** | `ok … 441.122s` | **`FAIL … 603.701s`** |

Race gate real exit: **1**. Wall time 610s. No `-timeout` flag was passed; both failures are the
default 10-minute per-package alarm. No `WARNING: DATA RACE` / `DATA RACE` marker appears anywhere
in the 320-line race log — the gate is red purely on duration.

### 1.2 `internal/install` alarm (race log lines 19–160)

```
panic: test timed out after 10m0s
	running tests:
		TestStrictRegistryPolicyFailsUnknown (3s)
```

The dump's own frame confirms the package size:
`testing.runTests(0xc00000e180, {0x105903ba0, 0x6b, 0x6b}, …)` — `0x6b` = **107** top-level tests,
matching the static count exactly.

**Corrected alarm position.** `TestStrictRegistryPolicyFailsUnknown` is `registry_e2e_test.go:110`,
the **third** function declared in that file (`:88` `TestRegistryRevocationDeniesInstall`, `:97`
`TestRegistryAttestationLandsInMarker`, `:110` `TestStrictRegistryPolicyFailsUnknown`). Go runs a
package's tests in file order (alphabetical) then source order, so it is test **73 of 107**, and
**72 tests had completed** when the alarm fired 3s into it. Cycle 1 reported "test 1 of 3" and
"70 of 107"; that was wrong and is corrected here. It had been active **3 seconds** — it is not the
stuck test, it is the test that happened to start near the deadline.

| # | file | tests | cumulative | state at alarm |
| ---: | --- | ---: | ---: | --- |
| 1 | `aba_test.go` | 5 | 5 | done |
| 2 | `cache_conformance_test.go` | 2 | 7 | done |
| 3 | `commit_test.go` | 19 | 26 | done |
| 4 | `diagnostics_test.go` | 6 | 32 | done |
| 5 | `dryrun_conformance_test.go` | 1 | 33 | done |
| 6 | `generation_test.go` | 5 | 38 | done |
| 7 | `install_test.go` | 21 | 59 | done |
| 8 | `maintenance_test.go` | 3 | 62 | done |
| 9 | `private_test.go` | 8 | 70 | done |
| 10 | `registry_e2e_test.go` | 3 | 73 | **tests 71–72 done; test 73 was 3s in** |
| 11 | `revalidation_test.go` | 6 | 79 | never started |
| 12 | `stage_test.go` | 28 | 107 | never started |
|  | **total** | **107** |  | **72 of 107 completed in ~600s** |

### 1.3 `internal/install/atomicity` alarm (race log lines 161–299)

```
panic: test timed out after 10m0s
	running tests:
		TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder/global-auto (8m28s)
		TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder/global-auto/80-removal (44s)
		TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder/project-hybrid-auto (8m28s)
		TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder/project-hybrid-auto/50-env-file (52s)
```

`testing.runTests(0xc00000e180, {0x1034b1da0, 0x8, 0x8}, …)` — `0x8` = **8** top-level tests, again
matching the static count (4 in `activation_test.go`, 4 in `commit_atomicity_test.go`).

The two scenarios are `t.Parallel()` subtests (`commit_atomicity_test.go:148-149`), so both ran
concurrently for **508s**; the package spent the preceding **~95s** on the tests that sort before
them. Those are exactly the four `activation_test.go` tests —
`TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` is the first function in
`commit_atomicity_test.go`, so the other three tests in that file had **not** run yet and are
*additional* time the package still owed. Cycle 1 folded all seven non-sweep tests into the 95s
prefix; that was wrong and is corrected in §3.4.

Class positions at the alarm:

- `projectSweepClasses` (`commit_atomicity_test.go:77`) = `10-context, 20-runtime, 30-shim-canonical,
  50-env-file, 60-adapter-ledger, 80-removal, 90-consumer` (7). At the alarm: 3 injections finished,
  the 4th (`50-env-file`) was 52s in.
- `globalSweepClasses` (`commit_atomicity_test.go:90`) = `10-context, 40-shim-forwarding,
  60-adapter-ledger, 70-mirror-ledger, 80-removal` (5). At the alarm: 4 finished, the 5th
  (`80-removal`) was 44s in.

### 1.4 The atomicity overrun is pre-existing debt, not a candidate regression

`candidate-source-delta-post.txt` (23 lines) lists the candidate's entire divergence from the
accepted worktree: 20 candidate-only conformance tests plus 3 modified tests
(`cmd/curator/status_test.go`, `internal/buildcache/conformance_test.go`,
`internal/closure/conformance_test.go`). **No `internal/install/atomicity` path appears.** The
package is byte-identical to the accepted tree. The candidate's only additions inside
`internal/install` are `cache_conformance_test.go` and `dryrun_conformance_test.go` (2 of the 107
tests).

Corroborating this: `.temp/TASK-260720-2284br/gates-rework1/gate-race.log` records
`internal/install/atomicity` at **1422.407s** under race on **2026-07-28**, i.e. more than double the
default alarm, on a tree that predates this candidate. That gate passed only because it was run with
an explicit `-timeout 45m` (documented at
`TASK-260720-2284br_implementation-notes-rework-1.md:199,207`). The exact repo gate
`go test -count=1 -race ./...` with the default alarm has therefore been failing on atomicity for at
least a day before this candidate existed.

**Consequence for routing:** this task's rework is *not* undoing damage the candidate did. It is
paying down a pre-existing race-time budget overrun that the raised-timeout focused gates had been
masking. Worth stating on the board so the candidate is not judged as having caused it.

---

## 2. Root cause: where the seconds actually go

### 2.1 The captured hot path

Three goroutines in the two dumps are in state `[runnable]` (burning CPU, not blocked). **All three
are in the identical stack**, one in `internal/install` (goroutine 1115, race log line 73) and two in
`atomicity` (goroutines 290 and 261, lines 216 and 258):

```
strings.FieldsFunc                                strings/strings.go:441
transaction.namespaceComponents                   internal/transaction/namespace.go:221
transaction.namespaceContains                     internal/transaction/namespace.go:205
transaction.namespacePathsOverlap                 internal/transaction/namespace.go:178
transaction.validateIndependentTargetNamespaces   internal/transaction/namespace.go:104
transaction.validateJournal                       internal/transaction/journal.go:344   (install, atomicity/project)
transaction.(*Engine).validateJournal             internal/transaction/journal.go:354   (atomicity/global)
transaction.(*Engine).saveJournal                 internal/transaction/journal.go:72
transaction.(*Engine).stageTarget / copyStagingFile   internal/transaction/staging.go:37 / :161
transaction.(*Engine).Prepare                     internal/transaction/engine.go:78
install.runCommit                                 internal/install/commit.go:661
```

Target-slice lengths are readable directly off the frames:
`validateIndependentTargetNamespaces({…, 0x8, …})` = **8 targets** for `internal/install`, and
`0x13` / `0x14` = **19 / 20 targets** for the two atomicity scenarios.

**Honesty caveat (unchanged from cycle 1):** three stack samples are not a profile. This is strong
evidence — three independent goroutines across two separate test binaries, all CPU-runnable in the
same leaf — but it is not a measured attribution. Nothing in the recommended patch depends on this
being the dominant cost; it is reported because it explains the shape and because it points at a
much cheaper production-side fix (§6).

### 2.2 Why that path is expensive (verified against source)

`Engine.saveJournal` (`journal.go:71`) calls `Engine.validateJournal` (`journal.go:350`). That method
runs the namespace validation **twice** per save:

- `journal.go:351` → `validateJournal(journal)` → `validateIndependentTargetNamespaces(journal.Targets)` at `journal.go:344`
- `journal.go:354` → `validateIndependentTargetNamespaces(journal.Targets, targetNamespacePath{…engine.journalRoot})`

Each run (`namespace.go:26`):

1. Rebuilds up to `len(targets)*7 + len(reserved)` canonical paths (`namespace.go:27`). Per path it
   calls `canonicalNamespaceTargetPath` (→ `filepath.EvalSymlinks`, one syscall per component),
   `namespaceCaseInsensitive` (`namespace_case_darwin.go:13` → existing-ancestor walk +
   `unix.Pathconf`) and `namespaceNormalizationInsensitive` (`namespace_case_darwin.go:25` →
   ancestor walk + `unix.Statfs`).
2. Then runs an **O(P²)** pairwise loop (`namespace.go:100-111`) over those paths. Each pair calls
   `namespacePathsOverlap` → `namespaceContains` in both directions, and each of those splits both
   paths into components (`namespaceComponents`, `namespace.go:217`). On APFS/HFS `normInsensitive`
   is **true**, so `namespaceComponentEqual` (`namespace.go:230`) NFD-normalises every component of
   both sides of every comparison. Non-containing pairs additionally do `os.Stat`/`os.Lstat` via
   `namespaceIdentity`.

**Call frequency (re-verified in cycle 3).** `saveJournal` (declared `journal.go:71`) has **23 call
sites**:

| file | count | anchors |
| --- | ---: | --- |
| `internal/transaction/engine.go` | 16 | `:64, :105, :259, :325, :332, :339, :359, :401, :413, :443, :538, :547, :588, :658, :772, :823` |
| `internal/transaction/staging.go` | 7 | `:26, :33, :56, :141, :161, :218, :244` |
| **total** | **23** | |

`staging.go:141` and `:161` are inside the 32 KiB copy loop, so they fire per chunk. Cycle 1 counted
only the staging sites; cycle 2 wrote "24 … 8 in `staging.go`" while listing only seven staging
anchors — the arithmetic, not the anchor list, was wrong. Verified by
`grep -c 'saveJournal' internal/transaction/{engine.go,staging.go}` ⇒ 16 and 7.

With `P ≈ 7·T`, the atomicity scenarios pay roughly `(7·20)²/2 ≈ 9,800` pairwise comparisons —
twice — for *every* journal save, against `(7·8)²/2 ≈ 1,570` for `internal/install`. That is a ~6×
quadratic penalty for atomicity per save, on top of more saves.

This quadratic gap is the **hypothesis** for why atomicity's measured race factor is materially
higher than `internal/install`'s (§3.3): the pairwise loop is instrumented in-process Go and its cost
grows quadratically in `P`, so a package that builds larger journals pays more of it. It is
consistent with the
captured stacks (all three runnable alarm goroutines sit inside this loop, at `P` = 8, 19 and 20) and
with static source, but **no profile or package-level timing decomposition exists**, so no share of
wall time is attributed to it here. §3.5 states the limits of the same-run evidence explicitly. None
of §3.2's or §3.4's numbers depend on this attribution being correct — they rest on the measured
cross-run factors and their stated caveats.

### 2.3 Why injection cost is class-independent

Because the fault is injected at `PointAfterBackup`, which is a **commit-phase** boundary, every
fault injection first runs a **complete `Prepare`** — all 19–20 targets staged, every entry
journalled — before the fault can fire.

### 2.4 §2.3 is now measured, not inferred

Cycle 1 listed this as an inference. `.temp/logwork/TASK-260729-2kaopg/-tester--tester--codex-_RUN-260729-051449.log:4697-4765`
is an isolated `-v` **non-race** run of the atomicity package on the current tree shape, total
`ok … 371.708s`:

| top-level test | subtest | seconds |
| --- | --- | ---: |
| `TestStableHybridActivationCommitsWithoutRestarting` | – | 8.32 |
| `TestHybridDeclarationRemovedBeforeHomeLock…` | – | 3.02 |
| `TestHybridDeclarationRetargetedBeforeHomeLock…` | – | 3.01 |
| `TestHybridDeclarationAddedBeforeHomeLock…` | – | 8.57 |
| `TestFailureAtEveryTargetClass…` | **global-auto** | **186.86** |
| | `10-context` | 23.78 |
| | `40-shim-forwarding` | 26.24 |
| | `60-adapter-ledger` | 28.94 |
| | `70-mirror-ledger` | 30.68 |
| | `80-removal` | 31.27 |
| | **project-hybrid-auto** | **301.67** |
| | `10-context` | 35.11 |
| | `20-runtime` | 36.94 |
| | `30-shim-canonical` | 38.76 |
| | `50-env-file` | 40.20 |
| | `60-adapter-ledger` | 38.98 |
| | `80-removal` | 34.99 |
| | `90-consumer` | 35.95 |
| `TestAdapterMirrorLinksAreJournaledAndRestoredExactly` | auto / symlink | 17.76 / 17.89 |
| `TestStaleAdapterEntryIsRemovedBeforeTheConsumerLedger` | symlink / auto / copy | 14.94 / 14.98 / 18.46 |
| `TestStaleAdapterRemovalRollsBackToTheExactPriorEntry` | – | 9.97 |

**Wall reconstruction** (serial parents, parallel subtests): `22.92 + 301.67 + 17.89 + 18.46 + 9.97 =
370.91s` against the reported `371.708s` — 0.80s unattributed. The model is complete, and the
**`project-hybrid-auto` chain alone is 81% of package wall time**.

Two facts fall straight out:

1. **Injection cost is class-independent, as predicted.** The 7 project injections span
   34.99–40.20s (mean **37.28s**, spread ±7%) and the 5 global injections 23.78–31.27s (mean
   **28.18s**, ±13%). A rollback of one committed target (`10-context`) costs the same as a rollback
   of six (`90-consumer`). §2.3 is confirmed by measurement.
2. **Per-scenario fixed cost is now separable.** Project baseline + final success =
   `301.67 − 260.93 = 40.74s`. Global baseline + final = `186.86 − 140.91 = 45.95s`. These are the
   numbers Patch B's arithmetic needs, and they replace cycle 1's single-sample 114s/93s
   per-install figures.

### 2.5 The consequence that drives the patch

The sweep's cost is `(number of installs in the chain) × (per-install cost)`, the chain is inherently
serial (each injection must start from the state the previous rollback restored), and per-install
cost is fixed. **The only test-only lever on wall time is chain length** — i.e. partitioning — plus
ordinary parallelism for everything that is not in a chain.

---

## 3. Runtime model and race-factor evidence

Cycle 1 presented two `internal/install` figures (`600 / ((70/107)×341.415) = 2.69×` and
`(107/70)×600 = 917s`) as "two independent derivations that agree". **They are the same equation
rearranged.** The reviewer is correct; this section is rebuilt.

### 3.1 What is measured versus what is assumed

| Quantity | Status |
| --- | --- |
| Package wall times in the failing gate | **measured** (§1.1) |
| Alarm ordinal / subtest positions | **measured** (§1.2, §1.3) |
| Atomicity per-subtest non-race costs | **measured** (§2.4) |
| Race factor for `internal/install` | **measured once**, on a different tree, a 5-tree invocation, and a partially-red loaded run — lower confidence, see the caveat in §3.2 |
| Race factor for atomicity | **measured twice**, plus a third value derived in-gate (§3.3) |
| Race-factor *mechanism* (what makes one package inflate more than another) | **hypothesis, not measured** — consistent with the captured alarm stacks and static source, but no profile or timing decomposition exists. Cycle 3 rated this "measured in the failing gate itself"; that was wrong and is withdrawn (§3.5). No projection in §3.2 or §3.4 depends on it. |
| Projection of the *unfinished* portion of `internal/install` | **assumed** — no per-test timing for this package exists anywhere in `.temp/` |

There is no verbose per-test log for `internal/install` in the repository (searched `.temp/` and
`.task-board/` for `PASS: TestEndToEndInstall`: no hits). Its projection is therefore an
extrapolation and is labelled as such throughout.

### 3.2 `internal/install`

**Estimate A — uniform-cost extrapolation (assumption-based, single derivation).**
72 of 107 tests consumed ~600s ⇒ `(107/72) × 600 =` **892s**. This assumes every test costs the
package average.

**Estimate B — work-weighted extrapolation.** The assumption in A is optimistic, because the 35
unstarted tests are the heavy tail. Static inventory of the whole package:

| | total call sites | in the 35 unstarted tests | unstarted share |
| --- | ---: | ---: | ---: |
| `e.install(…)` / `e.installGlobal(…)` | 86 | 30 | 35% |
| `e.skill(…)` + `e.buildSkill…(…)` fixture repos | ~108 | 48 | 44% |
| top-level tests | 107 | 35 | 33% |

`stage_test.go` alone holds 28 tests, 24 install invocations and 37 fixture repositories. Weighting
the completed fraction by the mean of the two work shares (0.651 and 0.556 ⇒ 0.60) gives
`600 / 0.60 =` **1000s**.

**Estimate C — independent measured corroboration.** `.temp/TASK-260720-2284br/gates-rework1/`
records a first-hand race run with a raised timeout
(`go test -race -timeout 45m ./internal/install/... ./internal/transaction/... ./internal/managerlock/... ./internal/staging/... ./internal/adapters/... -count=1`,
real exit **0**, zero data races):

- `gate-race.log`: `internal/install` **609.117s** (`gate-race-exit.txt`: `race exit=0`)
- `gate-gotest.log`, same cycle, same tree, non-race: `internal/install` **228.344s**
- ⇒ measured race factor **×2.67**

Scaling to the candidate's current non-race number: `341.415 × 2.67 =` **912s**. This *is* independent
of A and B — different tree, different invocation, measured on both sides.

**Honesty caveat on Estimate C (new in cycle 3).** Two weaknesses were found on re-inspection and are
reported rather than smoothed over:

1. The denominator's log is not a clean green run. `gate-gotest.log` is 25 lines, truncated at
   `internal/install`, and line 21 reads `FAIL github.com/relux-works/curator/internal/godriver
   99.383s` — two `TestEveryControlSeamFailureRejectsBeforeTheWorkerExecutes` subtests failed with
   `go-v1 process_timeout: Go probe exceeded its deadline` (`boundary_test.go:66`). That is a
   host-load symptom, so the `228.344s` was measured on a **loaded** machine. A separate
   `gate-gotest-final-exit.txt` records `install packages exit=0` for a later re-run, but that
   re-run's own package seconds are not in this log.
2. The tree that produced it can no longer be inspected. `.temp/TASK-260720-2284br/worktree` has
   since been reset: its `internal/install` now holds **4** test files / **58** tests and it has **no
   `atomicity` package at all**, which cannot be the tree that produced a 353.629s atomicity run. So
   the package composition behind ×2.67 is asserted from the batch, not verifiable today.

Estimate C therefore stays in the band but is **lower-confidence corroboration**, not a second
measurement of equal weight. Estimate B (1000s) remains the number the patch is sized against.
The atomicity ×4.02 pair is in better shape: both its sides come from the same `gates-rework1`
batch, and `gate-acceptance-verbose.log` lets its package composition be read directly (confirmed:
zero occurrences of `TestStableHybridActivationCommitsWithoutRestarting`, so that tree predates
`activation_test.go` — which is exactly why §3.4 adds the activation prefix separately).

**Band: 890–1000s**, i.e. **+48% to +67%** over the 600s alarm.

**Cross-package contention.** The ×2.67 was measured with only 5 package trees in flight. The failing
gate is `./...` over 41 packages with `go test` package parallelism defaulting to `GOMAXPROCS` = 16
on a 12-performance-core host, and `cmd/curator` (557.779s under race) overlaps `internal/install`
for nearly its whole run. Contention therefore pushes the true value toward the **upper** end of the
band, and the patch must be sized against ~1000s, not ~890s. Focused per-package numbers measured
after the patch will be *optimistic* relative to the `./...` gate for the same reason — §7 says so
explicitly in the pass condition.

### 3.3 Atomicity — and the correction to "both packages agree"

Cycle 1 concluded "the two packages agree on the race factor (2.69× / 2.75×)". The measured evidence
contradicts this:

| source | non-race | race | factor |
| --- | ---: | ---: | ---: |
| `gates-rework1` package pair (2026-07-28 tree, no `activation_test.go`) | 353.629s (`gate-acceptance-verbose.log`) | 1422.407s (`gate-race.log`) | **×4.02** |
| `gates-cycle5` focused `ACTIVATION` filter, 4 tests | 22.92s (§2.4) | 97.813s (`gate-race-activation.log`) | **×4.27** |
| verifier-3 in-gate, derived from §2.4 per-phase costs | — | — | **×3.33** (sweep), **×4.14** (activation prefix) |

The in-gate ×3.33 is solved from the alarm: the project chain had elapsed `603.701 − 95 = 508s`, of
which 52s was the in-flight 4th injection, so `(baseline ≈ 25s + 3 × 37.28s) × f = 456s ⇒ f = 3.33`.
The activation prefix gives `95 / 22.92 = 4.14`.

`internal/install` measures ×2.67 and atomicity ×3.3–4.3. They do **not** agree. §2.2 gives the
candidate explanation: atomicity's journals carry 19–20 targets against `internal/install`'s 8, a
~6× larger O(P²) pairwise namespace validation per journal save, and that loop is instrumented
in-process Go. That remains a **hypothesis** — it is consistent with the alarm stacks and with
source, but nothing here measures how much of either package's wall time it accounts for. The
disagreement itself is measured; only its cause is inferred.

### 3.4 Atomicity projection (corrected)

Built from the measured §2.4 phases and the in-gate factors:

| segment | non-race | factor | race |
| --- | ---: | ---: | ---: |
| 4 activation tests (serial, before the sweep) | 22.92s | 4.14 (observed) | **95s** |
| sweep, `project-hybrid-auto` chain (the binding one) | 301.67s | 3.33 (observed) | **1004s** |
| 3 post-sweep tests (serial, never reached in the failing gate) | 46.32s | ~4.0 | **185s** |
| **package total** | **370.91s** | | **≈ 1284s** |

**+114% over the 600s alarm.** Cycle 1 projected 1121s; that figure omitted the three post-sweep
tests entirely (they sort *after* `TestFailureAtEveryTargetClass…` in `commit_atomicity_test.go`, so
the failing gate never reached them and their cost is invisible in the alarm). The upper bound from
the ×4.02 package pair applied to the isolated non-race total is `371.708 × 4.02 =` **1494s**.

**Band: 1284–1494s.** This matters: cycle 1 sized Patch B against 1121s. The corrected figure is
15–33% higher, which is the reason §4 keeps the partition at **one class per scenario** rather than a
gentler two-class split.

### 3.5 Same-run race-factor observation (descriptive; corrected in cycle 4, censoring re-stated in cycle 5)

Every projection above imports a race factor from *another* run. The failing gate itself supplies a
set of same-run ratios that import nothing. **Cycle 3 over-read them.** This section is rebuilt to
say only what they support.

**Selection rule (mechanical, stated so the table can be checked).** Every package that completed —
printed `ok` — on **both** the non-race and race verifier-3 gates, and whose larger of the two wall
times is **≥ 10s**. That rule is applied to all 38 packages that completed on both sides and yields
exactly **five**. Nothing is filtered by outcome, and the table below is complete under it.

| package | non-race | race | factor |
| --- | ---: | ---: | ---: |
| `cmd/curator` | 384.270s | 557.779s | ×1.45 |
| `internal/godriver` | 55.306s | 76.752s | ×1.39 |
| `internal/transaction` | 27.218s | 34.106s | ×1.25 |
| `internal/runtimestore` | 18.630s | 18.244s | **×0.98** |
| `internal/managerlock` | 5.366s | 15.542s | **×2.90** |

Cycle 3 listed only the first three and wrote "every large completed package lands at ×1.25–1.45".
**That claim is withdrawn — it was false.** `internal/runtimestore` (omitted by the reviewer's
finding) reads ×0.98 and `internal/managerlock` reads ×2.90. Under the stated rule the completed
range is **×0.98–×2.90**, and across all 38 completed packages it is **×0.78** (`internal/capabilities`)
to **×8.47** (`internal/buildmeta`, on a 0.507s baseline). Sub-10s ratios remain noise-dominated;
that is why the rule exists, not a reason to present the surviving five as homogeneous. **There is no
single race factor for this host and run.**

**What the two timed-out packages contribute here (corrected in cycle 5).**

Both timed-out rows are the **same kind of object** and cycle 4 wrongly treated them as two different
kinds. Each package's race side is **right-censored**: the alarm stopped the clock before the package
finished, so the recorded race second-count is a strict *under*-estimate of what a completed race run
would have cost, while the non-race side is a completed measurement. Each ratio is therefore a
**one-sided lower bound** on the completed factor — never a point estimate, and never information-free.
What separates them is only *how severe* the censoring is, and therefore how useful the bound is.

- `internal/install`: completed non-race **341.415s**; race **FAIL at 603.306s** with test 73 of 107
  in flight — 72 finished, 1 in flight, **34** never started. The completed race duration is strictly
  greater than 603.306s, so the completed factor is strictly greater than
  `603.306 / 341.415 = 1.767075… ⇒` **> ×1.7670**. Censoring here is **moderate**: about two thirds of
  the package's tests had finished, so the bound is already informative — it exceeds four of the five
  completed packages in the table above, on the exact candidate, with no imported factor. It remains a
  bound, not the factor.
- `internal/install/atomicity`: completed non-race **441.122s**; race **FAIL at 603.701s** with the
  binding sweep chain unfinished and **three** post-sweep tests never started (§1.3, §3.4). The
  completed race duration is strictly greater than 603.701s, so the completed factor is strictly
  greater than `603.701 / 441.122 = 1.368557… ⇒` **> ×1.3685**. Censoring here is **severe**: the
  numerator is missing the tail of the most expensive test in the package *plus* three whole tests, so
  the bound is **too weak to compare against `cmd/curator`'s completed ×1.45, against the other
  completed packages, or to estimate the completed factor**. That is the correct statement — not that
  the row carries no information. Its *observed* value is low because the numerator is truncated; how
  far above ×1.3685 the completed factor actually sits is **not derivable from this row**, and this
  document makes no claim either way from it. Atomicity's in-gate factor is solved independently from
  phase costs in §3.4 (**×3.33** sweep, **×4.14** activation prefix), and that solve — not this row —
  is what §3.4's band rests on.

**Rounding convention (cycle 5).** `1.767075…` and `1.368557…` round *up* to ×1.77 and ×1.37 at two
decimals, so those two-decimal figures are **not** themselves valid strict floors. Wherever a bound is
asserted, this document uses the truncated values **> ×1.7670** and **> ×1.3685**; ×1.77 and ×1.37
appear only as roundings of the quotient, never as the bound.

**This section corroborates no mechanism.** Cycle 3 attributed the low factors of `cmd/curator`,
`internal/godriver` and `internal/transaction` to time spent in `git` and `go build` subprocesses.
Static source refutes that for `internal/transaction`, and the claim is withdrawn for all three:

| package | tests | `exec.Command` sites in `*_test.go` | `"git"` literals | what those sites are |
| --- | ---: | ---: | ---: | --- |
| `internal/transaction` | 50 | 2 | **0** | both re-execute the test binary (`os.Args[0] -test.run=^TestTransactionCrashHelper$`, `subprocess_test.go:56,459`); no `git`, no `go build` |
| `internal/runtimestore` | 18 | 5 | 0 | launcher/shim execution |
| `cmd/curator` | 62 | 1 | 4 | — |
| `internal/godriver` | 103 | 2 | 0 | — |
| `internal/managerlock` | 17 | 1 | 0 | — |

`internal/transaction` is an overwhelmingly **in-process** package and it sits at ×1.25, which
directly contradicts cycle 3's "in-process ⇒ large inflation, subprocess ⇒ small inflation" story.
`internal/managerlock` is likewise in-process, has no `time.Sleep` in its tests, and sits at ×2.90.
Call-site counts are textual occurrences, not wall-time shares; no profile or per-package timing
decomposition exists in the evidence, so **no causal attribution is made here at all**.

**The workload-shape reading that survives, labelled as a hypothesis.** The discriminator consistent
with §2.2 is not in-process-versus-subprocess but the **journal target count `P`** that
`saveJournal` re-validates through an O(P²) pairwise loop on every save. The alarm dumps record
`P` = 8 for `internal/install` and `P` = 19 and 20 for atomicity. `internal/transaction`'s own tests
exercise that code at the opposite end of the curve: ~33 `Prepare` call sites across the package,
with `Targets: []Target{…}` literals of one to a few entries. A package that never builds a large
journal cannot pay the quadratic penalty, whether or not it is in-process — which is a coherent
reading of the ×1.25 row without asserting anything about subprocesses. It is still a hypothesis: it
predicts a direction, not a magnitude, and it is unmeasured.

**What §3.5 is for, after the correction.** Two observations on the exact candidate, in the exact
failing gate. First, `internal/install`'s censored lower bound of **> ×1.7670** is a real same-run
measurement — moderately censored, therefore usable as a floor. Second, the completed packages in the
same run span ×0.98–×2.90, so a single imported factor cannot be assumed. Atomicity's own row
(**> ×1.3685**) is a valid bound of the same kind but is too severely censored to add anything here,
which is why §3.4 solves atomicity's factor from phase costs instead. All of this is descriptive
context for §3.2 and §3.4 — it neither raises nor lowers their bands, and the patch is still sized
against Estimate B (1000s) and the §3.4 atomicity band (1284–1494s).

---

## 4. Recommended smallest test-only patch

Two changes: **no case skipped, no `-timeout` flag, no production file touched, no assertion deleted
from any test body**.

One thing *is* deliberately given up, and it is stated here rather than buried: partitioning the
atomicity sweep retires the **cross-class residue chain** — the property that a rollback from class X
leaves nothing that trips a later injection at class Y. That was defence in depth layered on top of
an explicit per-injection check, not the primary assertion; the cycle-2 reviewer has sanctioned
retiring it. §5 row 1 and §9 risk 5 describe exactly what is lost and what still proves the same
property directly.

### 4.0 Required file allowlist (13 files)

`internal/install/`
`cache_conformance_test.go`, `commit_test.go`, `diagnostics_test.go`, `dryrun_conformance_test.go`,
`generation_test.go`, `install_test.go`, `maintenance_test.go`, `private_test.go`,
`registry_e2e_test.go`, `revalidation_test.go`, `stage_test.go`

`internal/install/atomicity/`
`activation_test.go`, `commit_atomicity_test.go`

**Explicitly NOT patch targets** (reviewer item 3):

- `internal/install/aba_test.go` — all 5 of its tests are on the exclusion list, so it must be
  byte-identical after the patch.
- `internal/install/atomicity/fixture_test.go` — belongs only to the optional, unmeasured
  `references/info.md` lever (§4.3). `env.userHome` already exists at `fixture_test.go:30`, so
  Patch B's per-scenario user homes need no change here.
- `internal/install/atomicity/doc.go` — unchanged.
- Any non-test file, anywhere.

### 4.1 Patch A — `internal/install`: mark 88 of 107 tests `t.Parallel()`

One line (`t.Parallel()`) as the first statement of each allowlisted test. Nothing else changes.

**Why it is safe here.** Every test builds its own fixture from three independent `t.TempDir()` roots
(`install_test.go:27-37`: `skillsRoot`, `home`, `project`), so there is no shared home, no shared
project, no shared manager lock and no shared consumer ledger between tests. Static sweep found **no**
`os.Chdir`, `os.Setenv`, `os.Stdout`/`os.Stderr` capture, `GOMAXPROCS`, `runtime.NumGoroutine`, fixed
port, `time.Sleep`, `os.Getwd`, or `TestMain` anywhere in the package's tests (verified: `grep -rn
"func TestMain" internal/install/` returns nothing).

The installer is already proven concurrency-safe on a *shared* manager home under `-race`:
`TestConcurrentProjectInstallsPreserveBothConsumers` and
`TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets` run two installs against one
home, and the `CONCURRENCY` gate passed with real exit 0
(`.temp/TASK-260720-2284br/gates-cycle5/gate-race-concurrency.log`, 58.885s).

**Exclusions — 19 tests must stay sequential.** Three hazard classes:

*Class 1 — `t.Setenv` (Go panics if called after `t.Parallel`; the variable is process-global):*

| Test | Anchor |
| --- | --- |
| `TestBuildRootExcludedBeforeLocaleRenderingWithoutCompilerExecution` | `install_test.go:225` |
| `TestRuntimeLauncherCapturesDeclaredSystemDependency` | `install_test.go:321` |
| `TestGlobalInstall` | `install_test.go:678` |
| `TestDefaultToolchainProbeRemovesItsProbeRootOnFailure` | `stage_test.go:1539-1540` |
| `TestDefaultToolchainProbeRemovesItsProbeRootOnSuccess` | `stage_test.go:1556` |
| `TestDefaultToolchainEstablishRemovesItsPrivateRootOnFailure` | `stage_test.go:1572-1573` |
| `TestProjectDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot` | via `isolateTempDir`, `private_test.go:24-26` |
| `TestGlobalDryRunKeepsEveryEphemeralPathInOneOperationPrivateRoot` | same helper |
| `TestDryRunRemovesItsOperationPrivateRootOnFailure` | same helper |
| `TestGlobalDryRunRemovesItsOperationPrivateRootOnFailure` | same helper |
| `TestRealRunKeepsEveryEphemeralPathInOneOperationPrivateRoot` | same helper |

*Class 2 — the `afterDocumentOpen` package-level hook.* `internal/install/generation.go:57` declares
`var afterDocumentOpen func(path string)`, read unsynchronised on every document read
(`generation.go:81-82`). Two test helpers assign it: `duringDocumentRead` (`aba_test.go:37`, cleared
at `:47`) and `onceAfterOpen` (`generation_test.go:44`, cleared at `:52`). **Running any of these in
parallel would produce a genuine `DATA RACE` on a production global**, so they stay sequential:

| Test | File |
| --- | --- |
| `TestProjectManifestABAAroundTheReadRestartsClosure` | `aba_test.go` |
| `TestGlobalManifestABAAroundTheReadRestartsClosure` | `aba_test.go` |
| `TestSubstitutionsABAAroundTheReadRestartsClosure` | `aba_test.go` |
| `TestHybridActivationABAAroundTheReadRestartsClosure` | `aba_test.go` |
| `TestByteIdenticalRewriteDuringTheReadDoesNotRestart` | `aba_test.go` |
| `TestReadDocumentBindsGenerationToBytesRewrittenInPlace` | `generation_test.go` |
| `TestReadDocumentBindsGenerationToBytesReplacedByRename` | `generation_test.go` |

All five `aba_test.go` tests are excluded, which is why that file is not a patch target.

*Class 3 — helper process:* `TestInstallHelperProcess` (`commit_test.go`) is the child half of
`TestConcurrentProjectInstallsPreserveBothConsumers`, re-executed via
`exec.Command(os.Args[0], "-test.run=^TestInstallHelperProcess$")` (`commit_test.go:726`). Leave it
sequential.

**Literal allowlist — add `t.Parallel()` to exactly these 88 functions.**

*Verification (re-run in cycle 3, on the candidate tree.)* The 107 declared test names were extracted
with `grep -h '^func Test' internal/install/*_test.go`, the 19 exclusions listed above were written
to a file, and the two sets were compared with `comm`:

- every one of the 19 exclusions exists in the package (`comm -23 excl all` ⇒ empty)
- `{all 107} \ {the 19}` has exactly **88** members
- the 88 names printed below are that set: symmetric difference against it is **empty in both
  directions** — no name that does not exist, none missing, none of the 19 smuggled in

- `cache_conformance_test.go` (2): `TestAuthoritativeCacheOutcomesDriveInstallation`,
  `TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted`
- `commit_test.go` (18): `TestCommitOrdersTargetClassesWithConsumerLast`,
  `TestConsumerLedgerIsAbsentAfterAFailedFirstInstallAndCommitsLastOnSuccess`,
  `TestCachePublicationFailureLeavesInstallationAndProtectedCacheUntouched`,
  `TestAFailedCommitRestoresTheBuildCacheItReplaced`,
  `TestARolledBackTargetCommitRestoresTheBuildCacheItReplaced`,
  `TestAnInFlightTransactionKeepsThePublishedCacheEntry`,
  `TestAPublicationThatChangedTheCacheIsReportedAsRetained`,
  `TestAReversalThatDidNotCompleteIsReportedAsRetained`,
  `TestBuildPublicationFailuresAreRedactedInTheResult`,
  `TestConcurrentProjectInstallsPreserveBothConsumers`,
  `TestRollbackCannotRestoreOverAnotherProjectsCommittedSharedTargets`,
  `TestRecoveryCompletesBeforeAnyNewMutation`,
  `TestStaleInstalledGenerationRestartsInsteadOfApplyingTheOldPlan`,
  `TestMaintenanceFailureAfterCommitIsAWarning`,
  `TestNoSharedTargetChangesBeforeTheManagerHomeLock`,
  `TestGlobalCommitCarriesNoConsumerLedger`,
  `TestStagedPlanRejectsTwoProducersClaimingOneLivePath`,
  `TestAdapterLedgerCommitsAfterTheMirrorsItClaims`
- `diagnostics_test.go` (6): `TestRedactDiagnosticReplacesEveryAbsoluteLocation`,
  `TestRedactDiagnosticCannotDriveATerminal`, `TestRedactDiagnosticIsBounded`,
  `TestRedactDiagnosticKeepsInvalidUTF8Out`, `TestPlanLinesRedactAnUntrustedReason`,
  `TestBuildFailuresAreRedactedInTheResult`
- `dryrun_conformance_test.go` (1): `TestAuthoritativeDryRunCasesMutateNothingPersistent`
- `generation_test.go` (3): `TestDeclarationGenerationIsContentAddressed`,
  `TestDocumentGenerationIsStableWhenUnreadable`,
  `TestObservationsRecheckDocumentsAgainstTheParsedBytes`
- `install_test.go` (18): `TestEndToEndInstall`,
  `TestRuntimeLauncherResolvesSkillDependencyWithoutShellHook`, `TestSecondInstallIsUpToDate`,
  `TestTamperTriggersReinstall`, `TestRemovedSkillCleanedUp`, `TestDryRunTouchesNothing`,
  `TestGitignoreGateSkips`, `TestMissingSystemCommandFails`, `TestMovedTagWarningAndStrict`,
  `TestRuntimeOnlyProviderGetsMarkerNoAdapter`, `TestRuntimeOnlyProviderStillRequiresSkillMd`,
  `TestHybridSkillActivatesWithoutTouchingProjectStore`, `TestHybridShadowedByProjectDeclaration`,
  `TestGlobalUpgradeDryRunLeavesPersistentStateUnchanged`,
  `TestGlobalInstallUsesManifestLocaleAndAuditGate`, `TestGlobalStrictTagsDetectMovedTag`,
  `TestMcpRequirementGatesInstall`, `TestAuditGateBlocksUndeclaredNetwork`
- `maintenance_test.go` (3): `TestPostCommitMaintenanceRunsUnderTheHeldHomeLock`,
  `TestPostCommitMaintenanceWarningsReachTheResult`,
  `TestPostCommitMaintenanceMarksInFlightJournals`
- `private_test.go` (3): `TestGlobalMcpFailureBlocksToolchainCacheAndBuild`,
  `TestGlobalRegistryFailureBlocksToolchainCacheAndBuild`,
  `TestGlobalMarkersCarryMcpAndAttestationEvidence`
- `registry_e2e_test.go` (3): `TestRegistryRevocationDeniesInstall`,
  `TestRegistryAttestationLandsInMarker`, `TestStrictRegistryPolicyFailsUnknown`
- `revalidation_test.go` (6): `TestStableDeclarationInputsCommitWithoutRestarting`,
  `TestProjectDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState`,
  `TestProjectDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure`,
  `TestGlobalDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleState`,
  `TestGlobalDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure`,
  `TestDevSubstitutionAppearingBeforeHomeLockRestartsClosureResolution`
- `stage_test.go` (25): `TestDryRunPlansBuildsWithoutToolchainSessionOrPersistentState`,
  `TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState`,
  `TestGlobalStagingFailureLeavesGlobalScopeUnchanged`, `TestDryRunReportsCacheHitWithoutBuilding`,
  `TestCacheHitPerformsNoSourceAwareGoCommand`, `TestStagingRunsProviderFirstAndCommandLexical`,
  `TestStagedOutputsStayPrivateAndAreReleased`,
  `TestSecondBuildFailurePreservesPriorInstallationAndLiveCache`,
  `TestToolchainDriftAfterTheFinalBuildBlocksHandoffAndPreservesLiveState`,
  `TestGlobalToolchainDriftAfterTheFinalBuildPreservesGlobalScope`,
  `TestCacheHitOnlyPlanStillFinalizesToolchainTrust`,
  `TestCacheInspectionIsBracketedByTheFrozenSource`,
  `TestSourceMutationDuringStagingBlocksHandoffOfACacheHit`,
  `TestToolchainFailureBlocksEveryPersistentMutation`,
  `TestCorruptCacheEntryIsRebuiltAndNeverReused`, `TestUnsupportedCacheProtectionFailsClosed`,
  `TestIncompletePlansNeverClaimACompleteDiagnostic`,
  `TestToolchainFailurePlansAnInventoryOfEveryActiveCommand`,
  `TestUntrustedCacheEntryIsRebuiltAndNeverReused`,
  `TestScriptOnlyInstallPerformsNoToolchainOrCacheWork`,
  `TestInjectedClockAndGenerationReaderDriveInstallMarkers`,
  `TestPlannedBuildAccessorsAreImmutable`,
  `TestReleaseTakesNoToolchainVerdictAfterLiveMutation`,
  `TestGlobalReleaseTakesNoToolchainVerdictAfterLiveMutation`,
  `TestSessionReleaseFailureWarnsWithoutFailingACommittedInstall`

**Ordering guarantee this relies on.** Go's `testing` releases parallel top-level tests only after
every sequential top-level test in the same parent has finished. The 19 sequential tests therefore
never overlap the 88 parallel ones, so their `t.Setenv` windows and their `afterDocumentOpen`
assignments cannot leak into a parallel test. This is why Patch A does not need `TestMain`.

**Expected savings.** The 19 sequential tests hold ~10 of 86 install invocations and ~16 of ~108
fixture repositories ⇒ ~13% of package work ⇒ **116–130s** of the 890–1000s projection. The
remaining **774–870s** of race-time CPU is spread over 88 independent tests, bounded by `-parallel`
(default `GOMAXPROCS` = 16) and by 12 performance cores:

| effective concurrency | parallel part | + sequential ~123s | total |
| ---: | ---: | ---: | ---: |
| 2× (worst plausible) | 387–435s | 123s | **510–558s** |
| 4× | 194–218s | 123s | **317–341s** |
| 8× (realistic once the package tail is alone) | 97–109s | 123s | **220–232s** |
| 12× | 65–73s | 123s | **188–196s** |

**Correction to cycle 1:** cycle 1 said "even the pessimistic 2× case clears the alarm". It clears the
**600s alarm** but not the **480s pass condition** cycle 1 itself set — 2× lands at 510–558s. The 2×
case is the failure boundary, not a comfortable floor. Realistic landing zone **220–340s**.

### 4.2 Patch B — `internal/install/atomicity`: partition the sweep, parallelise the rest

**B1 — one class per scenario.** Today `sweepScenario.classes` (`commit_atomicity_test.go:26`) is
overloaded: it drives **both** the fault-injection loop (`:163`) **and** the post-sweep full-coverage
assertion (`:206`). Narrowing it would silently narrow the coverage assertion too. Cycle 2 said "keep
the full list on the coverage assertion, narrow only the loop" without saying how — which is not
implementable against the current struct. Reviewer item 3 is correct. The literal change follows.

**B1.a — split the field.** Add one field; leave `classes` exactly as it is.

```go
type sweepScenario struct {
	name string
	mode string
	// classes are the classes this scenario must actually commit; a scenario
	// that stops producing one silently would otherwise weaken the sweep.
	classes []string
	// injectClasses are the classes this entry injects a fault at. It is a
	// subset of classes; the union across a scope's entries is the whole of it.
	injectClasses []string          // NEW
	baseline func(t *testing.T, e *env)
	upgrade  func(t *testing.T, e *env)
	install  func(e *env, opts install.Options) install.Result
	userHome string                 // NEW, replaces sharedUserHome bool
}
```

**B1.b — the two call sites.** Exactly one line changes in the body:

- `:163` `for _, class := range scenario.classes {` → `for _, class := range scenario.injectClasses {`
- `:206` `for _, want := range scenario.classes {` — **unchanged**, still the full 7 / 5 list.

**B1.c — the constructors take the injected class.**

```go
func projectSweepScenario(name, mode, injectClass string) sweepScenario {
	return sweepScenario{
		name: name, mode: mode,
		classes:       projectSweepClasses,        // unchanged: full 7
		injectClasses: []string{injectClass},      // NEW
		…                                          // baseline/upgrade/install verbatim
	}
}

func globalSweepScenario(name, mode, injectClass string) sweepScenario { … }   // same shape, full 5
```

**B1.d — build the table from the class lists, so coverage cannot drift.** Replace the literal
two-entry slice at `:144-147`:

```go
var sweep []sweepScenario
for _, class := range projectSweepClasses {
	sweep = append(sweep, projectSweepScenario("project-hybrid-auto-"+class, "auto", class))
}
for index, class := range globalSweepClasses {
	scenario := globalSweepScenario("global-auto-"+class, "auto", class)
	scenario.userHome = globalUserHomes[index]
	sweep = append(sweep, scenario)
}
for _, scenario := range sweep { … }   // body unchanged
```

Because the table is generated *from* `projectSweepClasses` and `globalSweepClasses`, the union of
injected classes is equal to the full class lists **by construction** — §8.5 check 5 becomes
structural rather than a grep. Adding a class to either list automatically adds its entry.

**B1.e — `sharedUserHome bool` → `userHome string`.** At `:158`:

```go
if scenario.userHome != "" {
	e.userHome = scenario.userHome
}
```

and the parent builds one home per global class instead of one shared home (`:135-142`):

```go
globalUserHomes := make([]string, len(globalSweepClasses))
var userBins []string
for index := range globalSweepClasses {
	globalUserHomes[index] = t.TempDir()
	if runtime.GOOS != "windows" {
		userBin := filepath.Join(globalUserHomes[index], ".local", "bin")
		if err := os.MkdirAll(userBin, 0o755); err != nil {
			t.Fatal(err)
		}
		userBins = append(userBins, userBin)
	}
}
if runtime.GOOS != "windows" {
	t.Setenv("PATH", strings.Join(append(userBins, os.Getenv("PATH")),
		string(os.PathListSeparator)))
}
```

No new imports: `runtime`, `os`, `path/filepath`, `strings` and `testing` are all already imported by
`commit_atomicity_test.go` (`:4-17`). `env.userHome` already exists at `fixture_test.go:30`, so
**`fixture_test.go` is still not a patch target**.

Each entry keeps its own `env` (`:154`), baseline (`:159`), `before` snapshot (`:160`), upgrade
(`:161`), one-iteration injection loop (`:163`), and its own post-sweep success + class-order
assertions (`:200-215`). Chain length drops from **9 → 3** installs (project) and **7 → 3** (global).
The post-sweep full-coverage assertion then runs **7× and 5×** instead of once — strictly more often
than today.

**Why 5 distinct user homes on one `PATH` is safe (strengthened in cycle 3).** `sweepScenario`'s
comment at `:38` carries the invariant *"two scenarios must never share it: their snapshots would
overlap."* Distinct homes satisfy it more strictly than the shared one does. Two independent
mechanisms in `globalbins.Select` (`internal/globalbins/globalbins.go:114`) guarantee no scenario can
reach a sibling's bin:

1. The preferred loop (`:148-155`) tries `filepath.Join(userHome, ".local", "bin")` **first** and
   returns as soon as it is on `PATH` and passes `safeExistingUserBin`. Each scenario's own bin
   always hits on this first probe, so the `pathDirs` fallback scan (`:156-160`) is never reached.
2. Even if it were, the fallback is gated by the same
   `safeExistingUserBin(candidate, userHome, home, platform)` (`:195`), whose first condition is
   `underHome(path, userHome, platform)` (`:196`). A sibling's bin is not under this scenario's
   `userHome`, so it is rejected outright. `t.TempDir()` hands out sibling directories, never nested
   ones, so no scenario's home is an ancestor of another's.

Project entries are unaffected: they never call `installGlobal`, their `e.userHome` stays `""`, and
`snapshotState` (`fixture_test.go:210-214`) only adds the `user/*` keys when `e.userHome != ""`. They
already run today with one foreign user bin on `PATH`; four more changes nothing on that side.

### 4.2.1 Function-level producer allowlist for Patch B

New or changed identifiers in `internal/install/atomicity/commit_atomicity_test.go` — nothing outside
this list may be introduced:

| Identifier | Change |
| --- | --- |
| `sweepScenario.injectClasses` | **added** field, `[]string` |
| `sweepScenario.userHome` | **added** field, `string` — replaces `sweepScenario.sharedUserHome bool` (**removed**) |
| `projectSweepScenario(name, mode, injectClass string)` | signature gains the third parameter |
| `globalSweepScenario(name, mode, injectClass string)` | signature gains the third parameter |
| `globalUserHomes` | **added** local (`[]string`) in the parent test |
| `userBins` | **added** local (`[]string`) in the parent test |
| `sweep` | **added** local (`[]sweepScenario`) in the parent test |

`projectSweepClasses`, `globalSweepClasses`, `commitProbe` and every helper in `fixture_test.go` are
unchanged. No new helper function is required; if the producer prefers one, it must be named
`sweepScenarios` and declared in the same file.

### 4.2.2 Subtest paths change; no existing gate filter breaks

Scenario names become `project-hybrid-auto-<class>` and `global-auto-<class>`, so a full subtest path
goes from `…/project-hybrid-auto/50-env-file` to
`…/project-hybrid-auto-50-env-file/50-env-file`. Verified: the four accepted regression filters in
`.temp/TASK-260720-2284br/gates-cycle5/run-gates.sh:35-41` (`$R5`, `$REVALIDATION`, `$CONCURRENCY`,
`$ACTIVATION`) contain **only top-level test names** — no `/`-qualified subtest path anywhere — so
none of them is affected. Any *future* filter that names a sweep subtest path must be updated, which
is why the rename is recorded here rather than left to be discovered.

**B2 — parallelise the seven non-sweep top-level tests.** Add `t.Parallel()` to:

| Test | File | Note |
| --- | --- | --- |
| `TestStableHybridActivationCommitsWithoutRestarting` | `activation_test.go` | own `newEnv` |
| `TestHybridDeclarationRemovedBeforeHomeLockRestartsAndCommitsNoStaleContext` | `activation_test.go` | own `newEnv` |
| `TestHybridDeclarationRetargetedBeforeHomeLockRestartsAndCommitsNoStaleContext` | `activation_test.go` | own `newEnv` |
| `TestHybridDeclarationAddedBeforeHomeLockRestartsAndCommitsTheNewClosure` | `activation_test.go` | own `newEnv` |
| `TestAdapterMirrorLinksAreJournaledAndRestoredExactly` | `commit_atomicity_test.go` | parent only; subtests already `t.Parallel()` |
| `TestStaleAdapterEntryIsRemovedBeforeTheConsumerLedger` | `commit_atomicity_test.go` | parent only; subtests already `t.Parallel()` |
| `TestStaleAdapterRemovalRollsBackToTheExactPriorEntry` | `commit_atomicity_test.go` | own `newEnv` |

`TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` **must stay sequential** — it calls
`t.Setenv("PATH", …)` at `:141`, which panics after `t.Parallel`. Because Go releases parallel
top-level tests only after the sequential ones finish, the seven above run *after* the sweep, not
alongside it; their contribution is `max` of their individual costs rather than the sum.

**Expected savings, from the measured §2.4 phases:**

- one-class project chain = `40.74s (baseline+final) + 37.28s (mean injection)` = **78.02s** non-race
- one-class global chain = `45.95s + 28.18s` = **74.13s** non-race
- longest chain under race at ×3.33–4.27 ⇒ **260–333s**
- the seven non-sweep tests, now parallel, collapse from 69.24s serial to `max(8.57, 17.89, 18.46,
  9.97) = 18.46s` non-race ⇒ **61–79s**
- **package wall (chain bound) ≈ 320–410s**

**CPU floor check (corrected core count).** Total non-race CPU = project `7 × 78.02 = 546.1s` +
global `5 × 74.13 = 370.7s` + non-sweep `69.2s` = **986.0s**; under race ×3.33–4.27 ⇒
**3283–4210s**. Twelve sweep chains plus seven other tests want 19 slots; `-parallel` caps at
`GOMAXPROCS` = 16 and the host has **12 performance cores** (cycle 1 divided by 16, counting the 4
efficiency cores as equals — corrected here). CPU-bound wall = `3283–4210 / 12` = **274–351s**.

The two bounds coincide at **≈320–410s**; neither dominates. Allowing for `./...` contention with
`cmd/curator` in flight: **realistic landing zone 340–460s, margin 140–260s** against the unchanged
600s alarm.

Total CPU rises from 16 installs to 36 (`7×3 + 5×3`), i.e. **+125%**. That is the price of the
partition and it is charged against the whole-suite wall time; §7 requires the producer to report the
focused package seconds *and* to hand the `./...` gate back to the verifier rather than declaring
victory on the focused numbers.

### 4.3 Optional lever, not counted in the margin, not a required patch target

`atomicity/fixture_test.go:83` writes `references/info.md` into every fixture skill. Grep over the
whole `atomicity` package finds **exactly one occurrence** — the write itself; nothing asserts on it.
Dropping it removes staging entries per skill context, and `saveJournal` (hence the quadratic
namespace validation) fires repeatedly per staging entry. This is a genuine linear multiplier on the
dominant cost, but its size cannot be derived statically. **Do not adopt it in the smallest patch.**
If the focused runs land above the pass condition, measure it first (count `StagingEntries` with and
without) and adopt it only with that measurement attached — at which point `fixture_test.go` becomes a
declared, justified 14th file.

---

## 5. Assertion- and invariant-preservation matrix

Read this section with §5.1 — one invariant is **intentionally retired**, and the table says which.

| Surface | Assertions preserved | What changes | Why it is still sound |
| --- | --- | --- | --- |
| Rollback matrix, all 12 class injections | `result.Status == "failed"` (`:175`); full-state `before.diff(snapshotState)` (`:178`); no class committed after the failing one (`:181`); `assertReverseRollback` (`:186`); `assertNoJournalRemains` (`:187`) — every one of these still runs once per class, 12 times in total | Each class runs against its own baseline env instead of a shared one | Residue is asserted **directly** by the whole-snapshot comparison at `:178` after *every* injection, not inferred from the next injection failing. `snapshotState` (`fixture_test.go:193`) digests all 12–15 state classes, and `entryDigest` (`fixture_test.go:226`) digests a symlink by its destination rather than dereferencing, so a re-pointed or replaced mirror shows as a change. **Retired:** the cross-class sequencing link — see §5.1. |
| Post-sweep health + class coverage | `install after the rollback sweep failed` (`:203`); every class in `scenario.classes` committed (`:206-210`); classes committed in non-decreasing order (`:211-215`) | Runs once per partition (7× project, 5× global) instead of once per scope | Each partition proves its own machine is healthy after its rollback and that the full class list was genuinely on the table. Coverage is unchanged in extent and 7×/5× in frequency. |
| Global forwarding shims / user-bin mirror ledger | baseline `.local/bin/skill-a-tool` exists (`:109-111`); `ClassForwardingShim`, `ClassMirrorLedger` swept | `sharedUserHome bool` → `userHome string`; 5 distinct user homes, all on `PATH` | The `sharedUserHome` invariant (`:38`) is *satisfied more strictly*: no two scenarios share a user home, so no two snapshots can overlap. `globalbins.Select` (`globalbins.go:114`) prefers each scenario's own `<userHome>/.local/bin` (`:148-155`), and its fallback scan is gated by `safeExistingUserBin` → `underHome` (`:195-196`), which rejects a sibling's bin outright. |
| Mirror re-pointing, stale-removal ordering, activation restarts | every existing assertion in `TestAdapterMirrorLinks…`, `TestStaleAdapterEntry…`, `TestStaleAdapterRemoval…`, and the 4 activation tests | `t.Parallel()` on the parent only | Each already builds its own `newEnv` from independent `t.TempDir()` roots; none touches process-global state. The existing `t.Skipf` at `commit_atomicity_test.go:246` (filesystem link-capability guard) survives unchanged. |
| All 88 `internal/install` tests | every existing assertion, byte for byte | `t.Parallel()` added as the first statement | Per-test `skillsRoot`/`home`/`project` `t.TempDir()` roots; no `Chdir`, no `os.Setenv`, no stream capture, no `TestMain`, no package-var mutation. The 19 tests that *do* touch process globals are excluded by name. |

**Nothing is skipped, no case is dropped, no timeout is changed, and no production file is touched.**
No assertion is deleted from any test body. One *emergent* invariant is retired — §5.1.

### 5.1 The one invariant that is intentionally retired

This is a real reduction, not a re-labelling, and it is described as such because cycle 2's blanket
"no assertion removed" was contradicted by its own §9 risk 5.

**What exists today.** `TestFailureAtEveryTargetClassRestoresPriorStateInReverseOrder` runs all seven
project injections against **one shared baseline** (`commit_atomicity_test.go:154-161`), and the
comment at `:151-153` is explicit about why: *"any residue one rollback leaves behind shows up as a
failure of the injection after it."* Over a 7-long chain that exercises 21 ordered (X-then-Y) class
pairs, plus 10 for the 5-long global chain.

**What the partition leaves.** Each entry has its own baseline, so it exercises **zero** ordered
pairs. 31 ordered pairs → 0.

**Why this is defence in depth and not the primary check.** The residue property is asserted
*directly*, on every injection, by `before.diff(snapshotState(t, e))` at `:178`: the full state map
must be byte-identical to the pre-upgrade snapshot. The chained baseline can only detect residue that
`snapshotState` already digests — it observes the same paths through the same `entryDigest`. If a
rollback leaves residue, `:178` fails on **that** injection, immediately and with a precise diff,
rather than as a confusing failure of a later unrelated class. The chain adds ordering coverage
(*sequences* of rollbacks), not state coverage.

**What is genuinely lost.** A defect that (a) leaves residue outside every path in `snapshotState`'s
map (`fixture_test.go:195-214`) **and** (b) only manifests when a later class's install trips over it.
That is a narrow but non-empty class of defect. It is not hypothetical-only: `snapshotState` is a
hand-maintained path list, so anything a future class writes outside those 12–15 roots is invisible
to `:178`.

**Sanctioned.** The cycle-2 reviewer accepted this trade explicitly, on the grounds that every
injection still checks whole-state digest equality, the committed-class cutoff, reverse rollback
order, journal cleanup, and a full successful install covering every class. Recorded here so the
decision is attributable rather than implied.

**Cheaper alternative if a future reviewer wants some chain back.** Two classes per entry instead of
one restores 9 of the 31 ordered pairs at a cost of `40.74 + 2 × 37.28 = 115.3s` non-race per project
chain ⇒ ~384s race before contention, against ~260s for one class. That clears the 600s alarm but
eats most of the margin under `./...` and would likely miss the 480s pass bar in §7. It is the
fallback, not the recommendation.

---

## 6. Out-of-scope finding to route separately

The cheapest real fix is production-side and is **explicitly outside this task's scope**. Recording
it so it is not lost:

`Engine.saveJournal` re-validates the entire target-namespace independence graph **twice on every
journal write** (`journal.go:351` and `:354`), and it has **23 call sites** (16 in `engine.go`, 7 in
`staging.go`), two of which are inside the 32 KiB staging copy loop (`staging.go:141`, `:161`) and
therefore fire per chunk. The validation is O(targets²) with per-path
`EvalSymlinks` + `Pathconf` + `Statfs` syscalls and, on APFS/HFS, NFD normalisation of every path
component of every comparison.

Two production-side options, either of which would very likely remove the need for **any** test
restructuring:

1. Hoist namespace validation out of `saveJournal` — the target set is fixed at `Prepare` time;
   validate once when the target list is established rather than on every progress checkpoint.
2. Memoise per journal: cache the canonical key, `caseInsensitive` and `normInsensitive` per path,
   and cache the pairwise-overlap result, invalidating only when the target list changes.

**Recommendation:** open this as a separate story under the same epic. It is a performance change to
production code with a real correctness surface (the independence check is a safety invariant), so it
needs its own design, review and conformance pass — it must not be smuggled into a test-only rework.

---

## 7. Focused validation commands for the producer

**Non-overlap requirement.** These must not run while any other Go process is active — a
`go test -count=1 ./...` belonging to another agent was running while this diagnosis was written.
Before each command, run the verifier's two-scan barrier and require an **empty** result twice:

```
pgrep -af '(^|/)(go|.*\.test)( |$)|go-build|cmd/curator'
```

Exit 1 (no match) is the passing state for that probe. Do not start a Go command until two
consecutive scans are empty.

**Evidence protocol.** Every command below is a standalone process. Do not pipe a gate through `tee`
or any pipe chain; capture the real exit code as the last action and report it, including when it is
non-zero.

**Non-Go checks (cheap, run first):**

```
gofmt -l internal/install internal/install/atomicity
git -C <candidate> diff --check -- internal/install
go vet ./internal/install/...
```

**Allowed Go commands, sequential, no `-timeout` token, no `./...`:**

```
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 \
GOTMPDIR=<task-owned-dir>/atomicity \
go test -count=1 -race ./internal/install/atomicity
```

```
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 \
GOTMPDIR=<task-owned-dir>/install \
go test -count=1 -race ./internal/install
```

**Regression filters, unchanged from the accepted driver**
(`.temp/TASK-260720-2284br/gates-cycle5/run-gates.sh:35-41` defines `$R5`, `$REVALIDATION`,
`$CONCURRENCY`, `$ACTIVATION` verbatim; reuse them literally):

```
go test -race ./internal/install -count=1 -run "$R5"
go test -race ./internal/install -count=1 -run "$REVALIDATION"
go test -race ./internal/install -count=1 -run "$CONCURRENCY"
go test -race ./internal/install/atomicity -count=1 -run "$ACTIVATION"
```

**Flake surface.** Parallelism is the point of the patch, so repeat the two full-package race runs
**three times each** and require exit 0 every time. A single green run is not evidence that 88
newly-concurrent tests are stable.

**Pass condition.** Real exit 0 **and** printed package duration **≤ 480s** for both packages — a
2-minute margin under the unchanged 600s alarm. These focused numbers are measured with far fewer
packages in flight than the `./...` gate, so they are **optimistic**; the 480s bar exists to absorb
that gap. Do **not** add `-timeout`, `-short`, `-run` filters that drop cases, or `-count>1` as a
substitute for repeats.

**Do not run `go test -count=1 -race ./...` yourself.** That is the verifier's gate; running it would
overlap the verifier and burn ~10 minutes of contended CPU. Hand back for a fresh verifier pass once
both focused packages are green with margin.

**Forbidden in the producer pass:** `go test ./...`, coverage, `-timeout` in any form, Windows
execution, overlapping Go commands, cache clearing, any production-file edit, any change to
`conformance/v1`, and any board acceptance claim.

---

## 8. Candidate-integrity checks (rebuilt — reviewer item 1 and 2)

**Why cycle 1's approach was wrong.** The candidate at
`.temp/TASK-260720-jrrgw9/worktree` is an intentionally dirty delivery worktree: HEAD is
`17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` and the working tree carries the whole candidate delivery
plus untracked files. Raw `git diff --stat` / `--name-only` from HEAD therefore reports the
pre-existing delivery, not the proposed patch, and cannot isolate one from the other. The
`authoritative-digests*.txt` 448-file comparison covers only the **immutable conformance root** and
says nothing about candidate source bytes. Both are replaced below.

### 8.1 Capture the baseline immediately before the first edit

```
CAND=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree
BASE=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3dr6hw/candidate-baseline
mkdir -p "$BASE"
git -C "$CAND" rev-parse HEAD > "$BASE/head-pre.txt"
git -C "$CAND" status --porcelain > "$BASE/porcelain-pre.txt"
```

```
set -o pipefail
cd "$CAND" && find . -type f -not -path './.git/*' -print0 \
  | LC_ALL=C sort -z \
  | xargs -0 shasum -a 256 > "$BASE/manifest-pre.txt"
echo "manifest-pre exit=$?"
```

`pipefail` is set so the recorded exit is the real one for the whole chain, not `shasum`'s alone.
Record `wc -l "$BASE/manifest-pre.txt"` as the baseline file count.

Run this **after** the process barrier and **before** any edit, with no Go command in between. Go's
build cache lives in `~/Library/Caches/go-build` and `GOTMPDIR` is task-owned outside the worktree,
so a focused test run must not perturb the manifest — if it does, that is itself a finding to report.

### 8.2 Capture and compare after the patch

```
set -o pipefail
cd "$CAND" && find . -type f -not -path './.git/*' -print0 \
  | LC_ALL=C sort -z \
  | xargs -0 shasum -a 256 > "$BASE/manifest-post.txt"
echo "manifest-post exit=$?"
git -C "$CAND" rev-parse HEAD > "$BASE/head-post.txt"
git -C "$CAND" status --porcelain > "$BASE/porcelain-post.txt"
diff "$BASE/head-pre.txt" "$BASE/head-post.txt"; echo "head-unchanged exit=$?"
```

`head-unchanged exit=0` is required: the producer must not commit, stash, checkout or rebase.

The gate itself is a standalone process with a real exit code and explicit added / deleted / modified
detection:

```
python3 - "$BASE/manifest-pre.txt" "$BASE/manifest-post.txt" <<'PY'
import sys

ALLOWED = {
    "./internal/install/cache_conformance_test.go",
    "./internal/install/commit_test.go",
    "./internal/install/diagnostics_test.go",
    "./internal/install/dryrun_conformance_test.go",
    "./internal/install/generation_test.go",
    "./internal/install/install_test.go",
    "./internal/install/maintenance_test.go",
    "./internal/install/private_test.go",
    "./internal/install/registry_e2e_test.go",
    "./internal/install/revalidation_test.go",
    "./internal/install/stage_test.go",
    "./internal/install/atomicity/activation_test.go",
    "./internal/install/atomicity/commit_atomicity_test.go",
}

def load(path):
    table = {}
    for line in open(path):
        digest, name = line.rstrip("\n").split("  ", 1)
        table[name] = digest
    return table

pre, post = load(sys.argv[1]), load(sys.argv[2])
modified = sorted(p for p in pre.keys() & post.keys() if pre[p] != post[p])
added    = sorted(post.keys() - pre.keys())
deleted  = sorted(pre.keys() - post.keys())

for label, paths in (("M", modified), ("A", added), ("D", deleted)):
    for p in paths:
        print(label, p)

unexpected = [p for p in modified if p not in ALLOWED] + added + deleted
print("modified=%d added=%d deleted=%d unexpected=%d"
      % (len(modified), len(added), len(deleted), len(unexpected)))
print("INTEGRITY_OK" if not unexpected else "INTEGRITY_FAIL")
sys.exit(0 if not unexpected else 1)
PY
echo "integrity exit=$?"
```

Pass condition: **`integrity exit=0`**, i.e. every modified path is in the 13-file allowlist, and
there are **zero** added and **zero** deleted paths. `./internal/install/aba_test.go` and
`./internal/install/atomicity/fixture_test.go` must appear in **neither** the modified nor the added
list.

### 8.3 Second, independent check — regenerate the accepted-worktree delta

Verifier-3's `candidate-source-delta-post.txt` (23 lines) and `candidate-delta-digests-post.txt`
(23 lines) are an rsync-itemised comparison of the candidate against the accepted comparison
worktree. The exact flags are not recorded in the evidence, but this reproduces the observed output
shape (`*deleting <path>` for candidate-only files, `>fcsT....` for differing files):

```
rsync -rin --delete --exclude='.git/' \
  /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree/ \
  /Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree/ \
  > "$BASE/accepted-delta-post.txt"
echo "rsync exit=$?"
```

**Before trusting it**, run the same command against the *pre-patch* candidate and confirm the output
is byte-identical to `candidate-source-delta-post.txt`. If it is not, the flags differ from
verifier-3's and the producer must reconcile them before using this check; report that rather than
silently proceeding.

**Post-patch pass condition (corrected in cycle 3 — reviewer item 1).** Cycle 2 expected the delta to
grow by 13 `>fcsT....` entries. That is impossible. Two of the 13 allowlisted files are
**candidate-only** and already appear in the 23-line delta as `*deleting` lines:

```
*deleting internal/install/dryrun_conformance_test.go
*deleting internal/install/cache_conformance_test.go
```

They do not exist in the accepted worktree, so rsync has nothing to compare them against. Editing
them changes their line **not at all** — it stays `*deleting`. Only the **11** allowlisted files that
exist in *both* trees can become `>fcsT....` entries.

Correct expectation:

| group | count | state after the patch |
| --- | ---: | --- |
| candidate-only conformance tests (`*deleting`) | 20 | unchanged, including the two allowlisted `internal/install` conformance tests |
| already-modified tests (`>fcsT....`) | 3 | unchanged |
| **newly modified allowlisted files (`>fcsT....`)** | **11** | new |
| **total lines** | **34** | was 23 |

The 11 are the 13-file allowlist minus `internal/install/cache_conformance_test.go` and
`internal/install/dryrun_conformance_test.go`.

**Therefore: this check cannot prove anything about those two files.** Their integrity is established
**only** by the pre/post SHA-256 manifest gate in §8.2, where they appear in the `modified` list and
are matched against `ALLOWED`. If §8.2 is skipped, the two candidate-only test files are unverified —
§8.3 is a second, partial check, not a substitute.

**Digest expectations, split by whether the file is a patch target.**
`candidate-delta-digests-post.txt` (23 lines) records a SHA-256 for every path in the delta,
including — verified — both candidate-only install tests:

```
cf1e825a9ad7a45ee27377f7255c63a77374634ecc4437c1011b50d45b971170  internal/install/cache_conformance_test.go
975c9504b91281e132ea5c52db7cff4ea16978b46ff17bab903a398d610b113a  internal/install/dryrun_conformance_test.go
```

- **21 of the 23 must be byte-identical after the patch** — the 18 non-`internal/install`
  candidate-only tests and the 3 already-modified tests. In particular `cmd/curator/status_test.go`
  must remain `487b12bdf531e4714983eab83b804de7b4604513e435256e550f60391ee0d32e`.
- **The 2 quoted above must change**, because they are patch targets. Their new digests are not
  predictable in advance, so this file cannot validate them; §8.2's `ALLOWED` set is what bounds the
  change. Do not report their digest change as a violation, and do not report it as a pass either.

### 8.4 Immutable conformance root — a separate invariant

Kept as its own check, **not** presented as candidate-source integrity: re-run verifier-3's 448-file
digest comparison over
`.temp/TASK-260729-3nx97g/worktree/conformance/v1` and require it byte-identical to
`authoritative-digests-post.txt`. This proves no conformance vector moved; it proves nothing about
the candidate's own source.

### 8.5 Semantic checks on the patch itself

1. `grep -c 't.Parallel()' internal/install/*_test.go` must sum to **88**, and the set of enclosing
   test names must equal the §4.1 allowlist exactly — no extra test, none of the 19 excluded.
2. `grep -rn 't.Skip\|t.Skipf' internal/install internal/install/atomicity` must show **no added**
   skip; the one existing skip at `commit_atomicity_test.go:246` must survive unchanged.
3. `grep -rn -- '-timeout' internal/install internal/install/atomicity` must be empty.
4. Assertion count per touched file must be unchanged:
   `grep -c 't.Fatal\|t.Fatalf\|t.Errorf' <file>` before vs after.
5. Class-coverage check. If §4.2 B1.d is followed, the scenario table is generated by ranging over
   `projectSweepClasses` and `globalSweepClasses`, so the union of injected classes equals the full
   lists **structurally** — confirm by reading the table construction, not by grepping names. What
   still must be grepped:
   - `scenario.injectClasses` appears exactly **once**, at the injection loop
     (`commit_atomicity_test.go:163`);
   - `scenario.classes` appears exactly **once** in the body, at the post-sweep coverage assertion
     (`:206`), plus once in each of the two constructors;
   - `injectClasses:` is assigned only from the constructors' `injectClass` parameter, never from
     `projectSweepClasses` / `globalSweepClasses` directly (that would re-merge the two roles);
   - `sharedUserHome` no longer appears anywhere.
6. Global user-home check: `len(globalUserHomes) == len(globalSweepClasses)`, every element comes
   from its own `t.TempDir()` call, no two entries share a value, and each `.local/bin` is present on
   the `PATH` set in the parent. `env.userHome` must be assigned only from `scenario.userHome`.
7. Identifier check: the set of added/removed identifiers in `commit_atomicity_test.go` must equal
   the §4.2.1 table exactly — no other new field, local or helper.

---

## 9. Residual risks

1. **`-race` may now surface a real data race.** Patch A runs 88 tests concurrently against
   production code previously only ever exercised serially. `afterDocumentOpen` (`generation.go:57`)
   proves unsynchronised package-level test hooks exist in this package; there may be others. If a
   focused race run reports `WARNING: DATA RACE`, that is a **genuine finding about production
   code**, not a reason to revert the patch — route it as a new defect. Mitigating evidence: the
   `CONCURRENCY` race gate already runs two concurrent installs against one shared manager home and
   passed with real exit 0.
2. **The `internal/install` projection is an extrapolation, not a measurement.** No per-test timing
   for this package exists in the repository. The 890–1000s band rests on a uniform-cost assumption
   (corrected upward by static work weighting) plus one measured race factor from a different tree —
   and that measurement was itself downgraded in cycle 3: its non-race denominator comes from a log
   that also recorded a genuine `internal/godriver` FAIL under load, and the tree behind it has since
   been reset and cannot be re-inspected (§3.2). Size the patch against Estimate B (1000s). Measure,
   do not assume. This is the weakest link in the whole document.
3. **Cross-package contention is modelled, not measured.** The measured ×2.67 and ×4.02 factors come
   from 5-tree invocations. Under `./...` with 41 packages and `cmd/curator` overlapping, focused
   post-patch numbers will read better than the real gate. This is why §7 sets the bar at 480s
   rather than 600s and why the `./...` gate must go back to the verifier.
4. **Atomicity's margin is the tighter of the two** (140–260s projected). If the focused run lands
   above 480s, the next lever is the `references/info.md` fixture reduction (§4.3) *with a
   measurement attached*, then the production-side fix in §6 — **not** a timeout override.
5. **One invariant is intentionally retired — sanctioned in cycle 2, restated for the record.**
   Partitioning the sweep retires the cross-class residue chain: 31 ordered (X-then-Y) class pairs
   today (21 project + 10 global), zero after. This is defence in depth on top of the per-injection
   `before.diff(snapshotState)` digest check, which still proves complete state restoration after
   *every* injection, and the post-sweep success + ordering assertion, which still proves the machine
   is usable afterwards. The residual exposure is a defect that both escapes `snapshotState`'s
   hand-maintained path list *and* only manifests through a later class — narrow, but non-empty.
   Full treatment in §5.1, including the two-class fallback that restores 9 of the 31 pairs at a cost
   the 480s pass bar probably cannot absorb. If a later reviewer withdraws the sanction, atomicity
   cannot reach the default alarm test-only, and the decision escalates to: (i) accept a documented
   `-timeout` on the repo gate, (ii) gate the sweep behind a build tag, or (iii) take the
   production-side fix in §6. That is a human decision about what the suite is for, not something to
   settle inside a test-only rework.
6. **The race-factor *mechanism* is unmeasured (added in cycle 4).** §2.2's O(P²) namespace-validation
   explanation is a hypothesis consistent with the alarm stacks (`P` = 8, 19, 20) and with static
   source; it is not backed by a profile or any per-package timing decomposition. Cycle 3 presented
   the same-run package ratios as corroboration of it, which the cycle-3 reviewer correctly rejected:
   `internal/transaction` is an in-process package at ×1.25 and `internal/managerlock` an in-process
   package at ×2.90, so the same-run data does not separate in-process from subprocess work at all
   (§3.5). This affects only the *explanation*. None of the projected bands depend on it, and the
   patch does not act on it — the production-side fix that would is routed separately in §6. If the
   producer wants the mechanism confirmed, the cheap check is a focused `-race -cpuprofile` run of
   the atomicity sweep, which is outside this diagnosis task's no-execution boundary.
7. **`go test` package-level parallelism is unchanged.** Both packages still run concurrently with
   `cmd/curator` under `./...`. Patch B raises atomicity's total CPU by ~125%, which slightly
   lengthens the whole-suite wall clock even as it shortens atomicity's own. The verifier's `./...`
   wall time should be expected to move, and should be reported rather than treated as noise.

---

## 10. Cycle-5 addendum — the plan is now partly *executed*, and risk 4 has fired

**This section is additive. It changes no number, command, allowlist, or recommendation in §1–§9.**
It exists because evidence that did not exist when cycles 1–4 were written now does, and a diagnosis
that omitted it would be dishonest. Everything the cycle-4 verdict required to be preserved is
preserved.

### 10.1 What happened

A sibling producer task, **TASK-260729-rfrdfo**, applied this diagnosis's Patch A + Patch B to a
prototype worktree (`.temp/TASK-260729-rfrdfo/worktree`) and is running the §7 gate list right now.
Its driver (`.temp/TASK-260729-rfrdfo/bin/run-gates.sh`) implements §7 faithfully: exact
`CURATOR_CONFORMANCE_ROOT`, task-owned `GOTMPDIR`, **no** `-timeout` token, **no** `./...`, one gate
at a time behind a two-scan process barrier, each gate a standalone process with its real exit code
written last.

The patch touches **exactly the 13 files** of the §4.0 allowlist — 11 in `internal/install`,
`atomicity/activation_test.go` and `atomicity/commit_atomicity_test.go` — with `aba_test.go` and
`atomicity/fixture_test.go` untouched, as §4.0 requires.

### 10.2 Results available at the time of writing (2026-07-29T18:33 local)

| Gate | Real exit | Log line | Wall |
| --- | ---: | --- | ---: |
| `gate-gofmt` | 0 | empty | 0s |
| `gate-vet` (`go vet ./internal/install/...`) | 0 | empty | 0s |
| `gate-atomicity-structure` (`-count=1 -v`, **non-race**) | 0 | `ok … internal/install/atomicity 285.434s` | 286s |
| `gate-race-atomicity-1` (`-count=1 -race`) | **0** | `ok … internal/install/atomicity` **591.280s** | 593s |
| `gate-race-atomicity-2` (`-count=1 -race`) | **0** | `ok … internal/install/atomicity` **560.828s** | 561s |
| `gate-race-atomicity-3` | — | **in flight** (`.log` empty, no `.exit`) | — |
| `gate-race-install-1..3` | — | **not started** | — |

No `internal/install` race gate has run yet, so the Patch A half of the plan is **unmeasured**.

### 10.3 The honest reading

**Patch B works, and it is not enough.**

- It clears the unchanged **600s alarm**: exit 0 twice, margins of **8.72s** and **39.17s**.
- It **misses §7's own 480s pass condition** by **111.3s** and **80.8s**.
- It lands **21.9%–28.5% above the top** of §4.2's projected 340–460s realistic landing zone.
- And these are **focused, single-package** numbers on an otherwise idle machine. §3.2, §7 and §9
  risk 3 all state that focused numbers read *optimistically* relative to the `./...` gate, where
  `cmd/curator` (557.779s under race) overlaps for nearly the whole run. An 8.72s margin does not
  survive that. **On this evidence the patched atomicity package should be expected to time out
  again under `go test -count=1 -race ./...`.**

This is precisely **§9 risk 4**, which reads: *"Atomicity's margin is the tighter of the two … If the
focused run lands above 480s, the next lever is the `references/info.md` fixture reduction (§4.3)
with a measurement attached, then the production-side fix in §6 — **not** a timeout override."*
That contingency is now live and should be followed as written.

### 10.4 Two projection errors this exposes, and which direction each went

The completed patched pair is the **first uncensored same-package race factor** ever measured for
this package, and it corrects §3.4 in the *favourable* direction while §4.2 was wrong in the
*unfavourable* one.

| Quantity | §3.4 / §4.2 assumed | rfrdfo measured | Direction |
| --- | --- | --- | --- |
| atomicity race factor | ×3.33–4.27 (imported / phase-solved) | `591.280 / 285.434 =` **×2.07**; `560.828 / 285.434 =` **×1.97** | assumption was **too pessimistic** |
| patched non-race package wall | modelled from a 78.02s longest chain | **285.434s** (from 441.122s, **−35.3%**) | model was **far too optimistic** |
| patched race package wall | 320–410s chain bound / 274–351s CPU floor ⇒ **340–460s** | **560.8–591.3s** | projection **missed low** |

**Why the wall model missed.** §4.2 priced one class chain at its *isolated* §2.4 cost (project
`40.74s + 37.28s = 78.02s`). Patch B gives each class its own `env`, so twelve chains plus seven
other tests want 19 slots on **12 performance cores**, and each chain pays a full baseline rebuild
under that self-contention. The verbose structure log shows it directly: the twelve per-class outer
subtests span **196.25s–247.78s** wall with inner injections of **68.50s–133.91s**, against the
modelled 37.28s injection and 40.74s baseline+final. That is roughly **×3 self-contention inflation
per chain**, and the package wall is bound by the longest chain (**247.78s**), not by the modelled
78.02s. Effective concurrency across the twelve outer subtests is `2732.07 / 285.434 ≈` **×9.6** —
good parallelism, but spent on **+125% more total CPU** (§4.2 predicted that CPU rise correctly; it
simply under-priced what the rise does to each individual chain).

The net: the two errors partly cancel (a gentler race factor against a heavier non-race baseline),
which is why the result clears 600s at all rather than by the projected 140–260s.

### 10.5 What this does *not* change

- **No section of §1–§9 is retracted.** §1 and §2 are evidence; §4's patch is the same patch that was
  in fact applied and does in fact run green; §5's assertion matrix, §7's commands, §8's integrity
  gates and both allowlists are all confirmed workable by the fact that a producer executed them
  end-to-end without hitting a contradiction.
- **The 480s bar in §7 stays.** It was set exactly to absorb the focused-versus-`./...` gap, and it
  is doing its job: it is refusing a result that would probably fail the real gate.
- **No `-timeout` override, skip, or assertion weakening becomes acceptable** because the margin is
  thin. §9 risk 4's ladder is the sanctioned response.

### 10.6 Recommended next step for the downstream producer

1. Let `gate-race-atomicity-3` and the three `gate-race-install-*` gates finish; report all six real
   exit codes and package seconds. Patch A is still unmeasured and may well clear 480s comfortably —
   its lever (88 tests going parallel) is much larger than Patch B's.
2. If atomicity's three race runs land in the 560–600s band as the first two did, apply **§4.3**
   (the `references/info.md` fixture reduction) *with a before/after measurement attached*, and
   re-run the three atomicity race gates. §4.3 is explicitly not counted in any margin above, so any
   saving it produces is additional.
3. If §4.3 is insufficient, escalate to §6 — the production-side `saveJournal` / O(P²) namespace
   revalidation fix — as a separate story. Do **not** hand back a test-only patch whose focused
   margin is 8.72s and call the race gate solved.

### 10.7 Evidence-honesty statement for this section

**No Go command, build, vet, or test was executed by this task in any cycle, including cycle 5.**
Every number in §10 is read verbatim from files that TASK-260729-rfrdfo wrote:
`gates/gate-*.log`, `gates/gate-*.exit`, `gates/gate-*.seconds`, and `bin/run-gates.sh`. The two
derived quantities (×2.07 / ×1.97, and the ×9.6 effective concurrency) are arithmetic on those
quoted values and state their inputs. `gate-race-atomicity-3` is reported as **in flight**, not as
passing or failing — it has no `.exit` file, and under that driver's protocol a missing `.exit`
means "killed or still running", never "passed". A live `pgrep` at 18:33 confirmed three Go
processes active, which is a second reason nothing was run from this task.
