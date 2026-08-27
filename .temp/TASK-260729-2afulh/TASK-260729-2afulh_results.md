# TASK-260729-2afulh — prototype: atomicity fixture trim (14-file scope)

Prototype of diagnosis §4.3 — the sanctioned next lever after TASK-260729-rfrdfo's Patch B missed the
480s bar. It removes the unasserted `references/info.md` fixture subtree from the atomicity suite's
skill builder, declares `internal/install/atomicity/fixture_test.go` as a **14th** allowlisted path,
and measures what that actually buys.

Everything below is measured in a task-owned copy. No main candidate, no accepted worktree, and no
production file was touched.

---

## 1. Trees, and what each one is for

| tree | contents | role |
| --- | --- | --- |
| `.temp/TASK-260729-2afulh/worktree` | accepted rfrdfo state + the one 14th-file edit | **the deliverable**; all gates run here |
| `.temp/TASK-260729-2afulh/measure/before` | accepted rfrdfo state + measurement-only test file | A/B arm "before" |
| `.temp/TASK-260729-2afulh/measure/after` | same + the 14th-file edit | A/B arm "after" |

`measure/*` exists only to count staging work. Its extra file
(`internal/install/atomicity/zz_measure_test.go`) is a **test** file, it is byte-identical in both
arms (`sha256 a16d178b6913bf5585ad02bd3417bab22a115124d2660ab0e22a99d4bc37f9d1`), and it is **not**
part of the 14-file patch. The deliverable worktree does not contain it.

---

## 2. Baseline provenance

`bin/manifest.sh` is the inherited path-sorted SHA-256 manifest tool (`find -type f`, `LC_ALL=C sort
-z`, `shasum -a 256`, `.git/` subtree excluded).

| check | command result | evidence |
| --- | --- | --- |
| copy is byte-identical to the rfrdfo prototype worktree | `diff` **exit 0**, 0 lines, 391 files each side | `evidence/manifest-rfrdfo-source.txt`, `evidence/manifest-baseline.txt`, `evidence/manifest-baseline-diff.txt` |
| copy equals rfrdfo's own recorded post-patch manifest | `diff evidence/manifest-final.txt(rfrdfo) …` **exit 0** | inherited `TASK-260729-rfrdfo/evidence/manifest-final.txt` |
| copy is the accepted 13-file state vs the source candidate | `modified=13 added=0 deleted=0 unexpected=0 forbidden_touched=0 not_modified_from_allowlist=0` → `INTEGRITY_OK` | `evidence/integrity-baseline-vs-candidate.txt` |
| the sibling task's baseline is the same source tree | source-only manifests (`.git`, `.task-board/` excluded) 305 files each, `diff` **exit 0** | `evidence/manifest-365r5r-baseline-src.txt` |

That last row matters: it makes `TASK-260729-365r5r`'s own baseline timing a valid *before* number
for this task, because it was produced on a byte-identical source tree.

---

## 3. The 14-file allowlist, literally

The inherited 13:

```
./internal/install/cache_conformance_test.go
./internal/install/commit_test.go
./internal/install/diagnostics_test.go
./internal/install/dryrun_conformance_test.go
./internal/install/generation_test.go
./internal/install/install_test.go
./internal/install/maintenance_test.go
./internal/install/private_test.go
./internal/install/registry_e2e_test.go
./internal/install/revalidation_test.go
./internal/install/stage_test.go
./internal/install/atomicity/activation_test.go
./internal/install/atomicity/commit_atomicity_test.go
```

plus the one path this task is scoped to add:

```
./internal/install/atomicity/fixture_test.go
```

`./internal/install/aba_test.go` and `./internal/install/atomicity/doc.go` stay forbidden.

**Only `fixture_test.go` is newly changed.** Manifest delta from this task's own verified baseline to
its final state (`evidence/delta-baseline-to-final.txt`, real exit **0**):

```
M ./internal/install/atomicity/fixture_test.go
modified=1 added=0 deleted=0
ONLY_FIXTURE_CHANGED
```

Manifest delta from the **source candidate** to the final state (`bin/integrity14.py`,
`evidence/integrity-14-vs-candidate.txt`, real exit **0**) lists exactly the 14 paths above and ends:

```
modified=14 added=0 deleted=0 unexpected=0 forbidden_touched=0 not_modified_from_allowlist=0
ALLOWLIST_SIZE=14 MODIFIED=14
INTEGRITY_OK
```

The checker also asserts `len(modified) == 14`, so a 13- or 15-file state fails it rather than
passing quietly.

---

## 4. The edit

One deleted line plus an in-source rationale (`evidence/fixture-diff.txt`):

```diff
 // skill creates a tagged skill repository with one exported script command.
+//
+// The context tree is deliberately the minimum this suite asserts on. It used
+// to carry a whitelisted references/info.md as well, which nothing here ever
+// read: …
 func (e *env) skill(name string) {
 	…
 	e.write(dir, "SKILL.md", "---\nname: "+name+"\ndescription: d\n---\n# "+name+"\n")
-	e.write(dir, "references/info.md", "ref")
 	e.write(dir, "scripts/"+name+"-tool", "#!/bin/sh\necho "+name+"\n")
```

`internal/install/install_test.go:73` keeps its own `references/info.md` write. That file belongs to
Patch A's package, it is byte-identical to the accepted state, and this task does not re-time
`internal/install`.

---

## 5. Assertion neutrality

`bin/neutrality.sh` (`evidence/assertion-neutrality.txt`):

```
BASELINE  t.Fatal(=17 t.Fatalf(=44 t.Error(=0 t.Errorf(=3 t.Skip=1 t.Parallel()=10 t.TempDir()=5 t.Setenv(=1 funcTest=8 t.Run(=4 timeout_tokens=0 e.write(=8
FINAL     t.Fatal(=17 t.Fatalf(=44 t.Error(=0 t.Errorf(=3 t.Skip=1 t.Parallel()=10 t.TempDir()=5 t.Setenv(=1 funcTest=8 t.Run(=4 timeout_tokens=0 e.write(=7
```

Every assertion, test, subtest, parallel marker, temp dir and `Setenv` count is **numerically
identical**. The only movement is `e.write(` 8 → 7: the deleted fixture write itself. `t.Skip=1` is
pre-existing (the Windows guard in `activation_test.go`), not added here. No `-timeout` token exists
in either state.

`commit_atomicity_test.go`, `activation_test.go` and `doc.go` are **byte-identical** between the
baseline and the final state (`cmp -s`), so the whole scenario-isolation surface is untouched:

* the seven project and five global `sweepScenario` entries, and the `scenarios` list that builds
  them, are unchanged;
* `injectClasses` partitioning, `scenario.classes` coverage lists, and the post-sweep full-class
  success assertion are unchanged;
* the five distinct `globalUserHomes` and the `PATH` prepend that keeps them from overlapping are
  unchanged, so the `sweepScenario` invariant *"two scenarios must never share it"* still holds;
* every `newEnv` still gets its own `t.TempDir()` skills root, manager home and project checkout.

Why removing the subtree cannot weaken an assertion:

1. Nothing in the package ever named it. The only occurrence in
   `internal/install/atomicity` was the write itself; after the trim the string survives only inside
   the explanatory comment (`grep` in `evidence/assertion-neutrality.txt`: *none outside comments*).
2. Every rollback check is a **whole-state digest comparison**. `snapshotState` enumerates *paths*,
   and `entryDigest` summarizes whatever a path actually holds by walking `os.ReadDir` recursively.
   Both arms therefore digest whatever the fixture wrote; neither depends on a specific file
   existing. `before.diff(after)` compares two digests taken in the same run, so a smaller tree
   changes both sides identically.
3. `assertReverseRollback`, `committedClasses`, `assertNoJournalRemains` and `adapterEntryState`
   operate on transaction events, journal directory entries, and link destinations. None counts
   context files.
4. Whitelist behaviour for `references/` is asserted where it belongs and still is:
   `internal/whitelist/whitelist_test.go` (six `references/…` cases including nested, dotfile and
   excluded-name variants) and `internal/interop/golden_test.go` (`references/notes.md` in the
   golden file list). Neither package is in this patch; both keep full coverage of the whitelist rule
   this fixture merely happened to exercise incidentally.

---

## 6. Measured staging cost — before/after

Method: a measurement-only observation hook reads the durable journal back off disk during
`Prepare`, which is the **only** window in which `TargetRecord.StagingEntries` is populated
(`internal/transaction/engine.go:92-103` clears every target's entries immediately before
`PhasePrepared`). Chunk counts are the engine's own `PointAfterStagingChunkSync` emissions. No
production file was edited to obtain either number. Raw output: `evidence/measurement-raw.txt`,
`measure/out/measure-before.log`, `measure/out/measure-after.log` (both real exit **0**).

The workload is exactly the two installation shapes the sweep drives: `projectSweepScenario`'s
baseline + upgraded second install, and `globalSweepScenario`'s. No fault is injected, which does not
change the staging cost — staging completes inside `Prepare`, before `PointAfterBackup`, where every
sweep fault fires.

### 6.1 The staged context manifest

```
before: directory:            directory:references   file:.csk-install.json  file:SKILL.md  file:references/info.md
after:  directory:            file:.csk-install.json file:SKILL.md
```

Five entries → three. Three non-empty files → two.

### 6.2 Totals per installation

| scenario | phase | staging entries | non-empty chunks | derived staging `saveJournal` |
| --- | --- | --- | --- | --- |
| project | baseline | 24 → **20** (−4, −16.7%) | 14 → **12** (−2, −14.3%) | 100 → **84** (−16, −16.0%) |
| project | upgrade | 34 → **28** (−6, −17.6%) | 19 → **16** (−3, −15.8%) | 140 → **116** (−24, −17.1%) |
| global | baseline | 26 → **22** (−4, −15.4%) | 16 → **14** (−2, −12.5%) | 110 → **94** (−16, −14.5%) |
| global | upgrade | 25 → **21** (−4, −16.0%) | 16 → **14** (−2, −12.5%) | 107 → **91** (−16, −15.0%) |

Entries and chunks are **measured**. The `saveJournal` column is **derived** — see §6.3 — and is
labelled as derived everywhere it appears.

### 6.3 How the `saveJournal` column is obtained, and why it is not a measurement

`Engine.saveJournal` is unexported and has no observation hook, so no test-only instrumentation can
count its invocations directly. The count below is arithmetic over two measured quantities using
coefficients read straight off `internal/transaction/staging.go`:

* `stageTarget` (`staging.go:16-61`) saves **3×** per staging entry — before creation (`:26`), after
  creation (`:33`), after the index advances (`:56`);
* `copyStagingFile` (`staging.go:118-176`) saves **2×** per copied chunk — before the write (`:141`)
  and after the durable sync (`:161`).

so `staging saves = 3·entries + 2·chunks`. Per staged **context** target that is `3·5 + 2·3 = 21`
before and `3·3 + 2·2 = 13` after: **−8 saves per context target**, and every measured row above is
exactly `−8 ×` its number of trimmed context targets. This is the whole mechanism: the entries and
the chunks are what the O(P²) namespace revalidation inside `saveJournal` gets re-run over.

It is the *staging* share of `saveJournal`, not the whole-transaction total; commit, rollback and
cleanup add more that this trim does not touch.

---

## 7. Focused gate results

Fresh tester run from the 14-file deliverable worktree:

| gate | real exit | wall s | result |
| --- | ---: | ---: | --- |
| `gofmt -l internal/install internal/install/atomicity` | 0 | 0 | empty output |
| `go build ./...` | 0 | 1 | empty output |
| `go vet ./internal/install/...` | 0 | 0 | empty output |
| `go test -count=1 -v ./internal/install/atomicity` | 0 | 273 | PASS |
| `go test -count=1 -race ./internal/install/atomicity` repetition 1 | 0 | 493 | `ok … 492.231s`; no `DATA RACE` |
| race repetition 2 | **NOT EVIDENCE** | — | driver received SIGTERM before `.exit`; partial log ignored |
| race repetition 3 | **NOT RUN** | — | stopped by orchestrator directive after the strict bar was conclusively missed |

Every executed gate had its own `BARRIER_OK` two-scan evidence. The partial
race-2 process left no `.exit` or `.seconds` file. At the next safe checkpoint,
directive `RUN-260729-b2a441:nudge:3afa0e` required the tester to stop because
one valid race result above 480 seconds makes acceptance impossible, release
the shared Go slot for the production fallback, and package a conclusive
rejection rather than spend another 16 minutes proving the same strict
predicate. `gates/DRIVER-STOP-REASON` records that decision.

The task's first interrupted driver evidence was preserved separately as
`gates-partial-RUN-260729-7ade05/`; none of it is used here.

---

## 8. Verdict against the 480s bar

**Reject this fixture-only trim as the timeout solution.**

It is a real improvement:

* atomicity non-race fell from the inherited 286s/305.113s baselines to 273s;
* the valid race run fell from the inherited 561–593s range to 493s;
* per-install staging work fell 12.5–17.6%, and the exact staging-path
  `saveJournal` count fell 14.5–17.1%.

But the acceptance predicate is strict: the count-one race gate must be
`<=480s` with defensible margin. The valid run took **493s**, missing the bar by
**13s**. Even a hypothetical 480s result would have zero margin; 493s is
unambiguously outside the accepted range. The fixture reduction should not be
integrated as the claimed atomicity timeout fix. Preserve the measurement as
evidence and proceed with the separately scoped production-side
`saveJournal`/namespace-validation optimization.

---

## 9. Evidence-honesty statement

Every accepted number here is either read from a
`gates/gate-*.{log,exit,seconds}` file written by `bin/run-gates.sh`, read from
a `measure/out/*.log` written by a standalone `go test` process, or is the real
exit code of a command run directly as a standalone process. No gate was piped
through `tee`. No gate carries a `-timeout` token. No `go test ./...`, no
coverage run, no host install, no stage, no commit, no publish, no pin, no
Windows execution was performed. The two-scan process barrier was green before
every executed gate; `gates/gate-*.barrier` records it per gate, and it is the
same barrier `TASK-260729-365r5r` uses, so no heavy Go run of the two tasks
overlapped.

`gates/DRIVER-DONE` is annotated `cooperative-stop-after-race1` and was written
manually after the orchestrator directive. It does **not** claim that
`run-gates.sh` naturally reached its last statement. Race repetition 2 has no
exit file and its partial log is not evidence; repetition 3 was not run.

The tester's first delta-check command exited **1** because the one-liner
accidentally keyed manifest rows by digest rather than path. The corrected
standalone command exited **0** and proved exactly one baseline-to-prototype
change: `internal/install/atomicity/fixture_test.go`, with no additions or
deletions. This checker-command error did not touch the candidate.

Invoking `bin/report.sh` directly exited **126** because the prepared helper
was not executable. Invoking it explicitly through `bash` exited **0**. This
affected report rendering only, not any gate.

The `saveJournal` column in §6.2 is **derived from measured entries/chunks, not
directly observed**, and is labelled as such at every occurrence. The
derivation is exact for the staging path: `stageTarget` calls `saveJournal`
three times for every entry, and `copyStagingFile` calls it twice for every
non-empty chunk. A direct observation hook would require a production edit or a
forced test-only seam outside the literal 14-file deliverable, so no such seam
was added.

`internal/install` (Patch A) was **not re-timed** here; its acceptance rests on
TASK-260729-rfrdfo's immutable gate evidence, and this task's
`internal/install` sources are byte-identical to that accepted state.

The *before* wall clocks are inherited, not re-run: they come from
`TASK-260729-rfrdfo/gates/` and `TASK-260729-365r5r/gates-baseline/`, both produced on a source tree
this task proved byte-identical to its own baseline (§2).
