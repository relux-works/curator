# TASK-260729-1y7okj — independent cmd/curator runtime optimization audit (rev. 2)

Date: 2026-07-29
Role: solution-architect (independent audit)
Revision: cycle-1 rework, answering `TASK-260729-1y7okj_review-verdict-cycle-1.md`
Disposition: analysis only, handed off to review. **No source file, test, or spec was modified in any
worktree. No Go test, build, vet, race, coverage, or Windows command was executed by this audit.**
Every timing figure is *derived* from preserved verifier and producer evidence, never measured here;
the timing agent owns measurement.

## 0. What changed in this revision

| ref | reviewer requirement | resolution |
|---|---|---|
| R1 of the verdict | all **three** protected-cache mutation cases need a live post-tamper plan | accepted and verified independently against production source; §5 R3 rewritten with a 3/11 partition |
| R2 of the verdict | name the clean-phase consolidation in the ranked steps; reconcile 19→8 (or 19→9), matrix, savings | accepted; the consolidation is now the explicit step **R3.1**, and §5.5 carries a line-by-line plan-invocation ledger that lands on **8** |
| rework instruction | record immutable reviewed-snapshot provenance, distinguish the moving primary candidate | §1 and §9 |
| rework instruction | preserve the literal narrow producer allowlist and the read-only boundary | §10, unchanged commands, plus a labelled re-pointing rule for the moved test names |

## 1. Provenance — immutable reviewed snapshot vs. the moving candidate

This audit is written against a **frozen snapshot**, not against a live directory. The primary
candidate has changed twice since the artifact under review was produced, and it is owned by another
in-flight producer.

### 1.1 The reviewed snapshot (authoritative baseline for every claim below)

| item | value |
|---|---|
| logical identity | the exact `cmd/curator/status_test.go` verifier 2 failed on |
| sha256 | `4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe` |
| materialized read-only at | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-1y7okj/reviewed-snapshot/status_test.go` |
| reconstructed from | `.temp/TASK-260729-2kaopg/worktree/cmd/curator/status_test.go` (`c62d0791…`, the recorded **pre**-patch file) |
| plus | board resource `TASK-260720-jrrgw9_status-timing-patch.diff`, sha256 `03c637da916ea5526298ec6ac41d0f234ccda4bfce098acc63f9e3dc29984acd` |
| verification | applying that diff to that base reproduces `4c2339dd…` **byte-exactly**; both digests recorded in `TASK-260720-jrrgw9_status-timing-patch-results.md` |

The snapshot lives under this task's own `.temp/` scratch path. Neither the candidate worktree nor the
accepted worktree was written to.

### 1.2 Files that did **not** move

| file | sha256 (prefix) | candidate | accepted `TASK-260729-2kaopg` |
|---|---|---|---|
| `cmd/curator/global_status_test.go` | `25aebe85…` | same | same |
| `cmd/curator/lifecycle_conformance_test.go` | `5e832811…` | same | not present |
| `cmd/curator/builds.go` | `6aab0a0f…` | same | same |
| `cmd/curator/main.go` | `a0798e7d…` | same | same |
| `internal/install/plan.go` | `b9684fde…` | same | same |
| `internal/godriver/session.go` | `a4186cd2…` | same | same |
| `internal/godriver/fingerprint.go` | `560d0c98…` | same | same |

Every product-source line reference in this audit is against those digests.

**Trap recorded for the next reader:** the repo-root main checkout
`/Users/iv/Developer/ReluxWorks/curator/cmd/curator/main.go` is a *different, pre-feature* file
(`1d2cc9c9…`); its `cmdStatus` has no compiled-command surface at all. Product line numbers here are
valid only against the worktree file `a0798e7d…`.

### 1.3 The concurrently moving primary candidate — explicitly **not** this audit's baseline

`.temp/TASK-260720-jrrgw9/worktree/cmd/curator/status_test.go` observed sequence:

| time | sha256 (prefix) | who |
|---|---|---|
| when the audited artifact was written | `4c2339dd…` | the one-file timing patch under `TASK-260720-jrrgw9` |
| 15:54:27 +0400 | `bb21c15c…` | primary producer, recorded by the reviewer |
| 15:54:47 +0400 (mtime), read here at 15:57 | `487b12bd…` | primary producer |

Nothing in this audit claims `4c2339dd…` is the current candidate. §9 states, per hunk, which
recommendations the moving candidate has already absorbed and which are still open. The R3 defect the
reviewer found is a defect in the submitted **plan**, and it is corrected here regardless of which
plan the primary producer ultimately merges.

Evidence read: `TASK-260720-jrrgw9_final-verifier2-results.md`,
`TASK-260720-jrrgw9_go-test-all-verifier2-failed.log`,
`TASK-260720-jrrgw9_cmd-timing-diagnosis-results.md`,
`TASK-260720-jrrgw9_status-timing-patch-results.md`,
`TASK-260720-jrrgw9_status-timing-patch.diff`, `TASK-260720-jrrgw9_second-timing-diagnosis.md`,
`TASK-260729-1y7okj_review-verdict-cycle-1.md`.

## 2. Failure re-confirmed as cumulative, not a hang

Unchanged from rev. 1, and unchallenged by review.

The goroutine dump at `go-test-all-verifier2-failed.log:569` shows
`TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands (1s)` as the only running test when the
alarm fired. That test uses `legacyProject` — script-only skill, no compiled command, no toolchain
session — so it cannot itself consume minutes; it was merely active at t=600 s. The stack tail below
it (`transaction.namespaceIdentity` under `install.runCommit`) is a live `os.Lstat`, i.e. forward
progress. **Diagnosis: cumulative package duration with insufficient margin.**

Full-run package table from the same log: `cmd/curator` FAIL 600.591 s, `internal/install`
411.094 s, `internal/install/atomicity` 543.719 s, `internal/godriver` 72.516 s, everything else
below 30 s.

## 3. Root cost model (static, derived)

Unchanged from rev. 1. Two product mechanisms account for essentially all `cmd/curator` wall time;
neither is changed by this plan.

**(a) Toolchain fingerprint — the cost of every plan.** `internal/install/plan.go:363 planBuilds`
resolves the trusted toolchain for any scope that activates a compiled command: `Toolchain.Probe` on
a dry run (`plan.go:382`), `Toolchain.Establish` otherwise (`plan.go:389`). Both land in
`internal/godriver/session.go:157 establish`, which runs `go telemetry off`, `go version`,
`go env -json`, then `fingerprintToolchain` (`session.go:266` → `internal/godriver/fingerprint.go:34`)
— a full walk of GOROOT that re-opens and SHA-256s **every** file. `godriver.Probe`
(`session.go:141`) closes the session immediately, and `Session.Close` (`session.go:332`) calls
`VerifyToolchain`, which fingerprints GOROOT **a second time** (`session.go:316`). `Probe` is
documented as *not* memoizing (`session.go:140`). Measured host GOROOT: **233 MB across 14,502
files**. One plan-only CLI invocation against a compiled scope = 3 `go` subprocesses + 2 full
233 MB / 14.5k-file digests.

**(b) Cold private GOCACHE — the cost of every real build.**
`internal/godriver/session.go:404 bootstrapEnvironment` points `GOCACHE`, `GOMODCACHE`, `GOPATH`,
`GOTMPDIR` at `layout.*` inside the operation-private root created per operation
(`internal/install/builddeps.go:166`). Every real compiled install compiles the whole dependency set
from an **empty** build cache.

**Unit costs.** Let `P` = one plan-only compiled invocation, `B` = one real compiled install that
rebuilds.

- `TestStatusReportsCompiledCurrentnessAndFailsCheck` pre-patch = `44P + 3B = 167.062 s`
- same test post-patch = `19P + 1B = 69.816 s`
- verifier-2 focused status/lifecycle/repair group = `37P + 15B + 13.1 s ≈ 267.207 s`

Solving the first two exactly gives **P ≈ 3.26 s**, **B ≈ 7.9 s** for a rebuild-free solve; the third
equation needs an average `B ≈ 8.9 s`, consistent with **B ≈ 8–11 s** for a rebuild and **≈ 4 s** for
a cache-hit install. Applying the model to the whole package predicts ≈ 463 s against the accepted
comparable-load clean derivation of **465.944 s** — the model reproduces the known baseline within
1 %. Observed high-load inflation (545.195 / 554.967 / 600.591 s vs 465.944 s) is **×1.17 to ×1.29**;
savings scale by the same factor under `./...`.

All savings below use **P = 3.26 s**. Every rounded figure in this document is reproducible from that
constant.

## 4. Static call/fixture inventory — `cmd/curator` (reviewed snapshot)

`builds_test.go`, `gc_test.go`, `main_test.go` and `lifecycle_conformance_test.go` contain **no**
compiled command (`lifecycle_conformance_test.go:233` publishes `"commands": {}`; the other three
never install a schema-6 build skill). The entire compiled cost lives in `status_test.go` and
`global_status_test.go`.

| test | file:line | real builds `B` | plan-only `P` | cheap (toolchain-refused) |
|---|---|---:|---:|---:|
| `TestStatusReportsCompiledCurrentnessAndFailsCheck` | `status_test.go:337` | 1 | **19** | 0 |
| `TestInstallAndUpgradeRepairCorruptCompiledState` | `status_test.go:628` | 6 | 8 | 4 |
| `TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails` | `status_test.go:712` | 6 | 6 | 0 |
| `TestStatusReportsATransitivelyResolvedCompiledCommand` | `status_test.go:993` | 1 | 4 | 0 |
| `TestStatusReportsAnUnusableToolchainPerCompiledCommand` | `status_test.go:1059` | 1 | 0 | 4 |
| `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck` | `status_test.go:1144` | **1 + 3 rebuild + 1 hit** | **5** | 0 |
| `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` | `status_test.go:1273` | 2 | 2 | 1 |
| `TestGCRetainsAndReportsReferencedCompiledState` | `status_test.go:1337` | 1 | 1 | 0 |
| `TestDryRunNeverClaimsACompletedCompilerCheck` | `status_test.go:1367` | 0 | 2 | 0 |
| `TestStatusExplainsAnUnusableGoToolchain` | `status_test.go:1401` | 0 | 0 | 3 |
| legacy-scope tests (`1248`, `1428`, `1454`) | `status_test.go` | 0 | 0 | 0 |
| `TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck` | `global_status_test.go:344` | 1 | 6 | 2 |
| `TestGlobalStatusReportsATransitivelyResolvedCompiledCommand` | `global_status_test.go:560` | 1 | 2 | 0 |
| other global tests | `global_status_test.go:623+` | 0 | 0 | 0 |
| **total** | | **≈ 24** | **≈ 55** | 14 |

The 19 `P` of `TestStatusReportsCompiledCurrentnessAndFailsCheck` decompose exactly as:
clean `status --json` (`:344`) 1 + clean `status --check` (`:373`) 1 + 14 per-case
`status --json --check` (`:566`) + 3 `humanCLI` `status app` (`:612`, cases at `:407`, `:428`,
`:539`) = **19**.

### Duplicate expensive work identified

1. **`status_test.go:1216` — `t.Cleanup(func() { reinstall(t) })` inside a four-case loop.**
   All four moves mutate **only** `home/cache/build/go-v1/<entry>`: `corruptCacheArtifact`
   (`:270`, artifact bytes), `os.RemoveAll` over `cacheEntries` (`:1177`), `os.Chmod(entry, 0o777)`
   (`:1185`), `replaceCacheEntry` (`:890`, artifact + receipt bytes). Nothing touches the installed
   tree — the test itself asserts the marker is byte-identical (`:1220-1226`). Repairing a
   *reporting* fixture with four real compiled reconciliations is exactly what the accepted
   `snapshotBuildCacheAfter` helper (`global_status_test.go:185`) replaced elsewhere in this package.
   **3 rebuilds + 1 cache-hit install are pure fixture cost.**
2. **`status_test.go:1203` — the per-case `install.Project(..., DryRun: true)` re-plan.** Its own
   comment (`:1200-1202`) states it exists *because* a preceding case reinstalled and rewrote the
   marker. Remove the reinstall and that reason disappears. **4 plans are a pure consequence of
   finding 1.**
3. **`status_test.go:377-620` — 14 live compiled-scope CLI acquisitions where the accepted global
   counterpart uses one plus per-case replay.** `global_status_test.go:386 acquireGlobalPlan` +
   `:149 replay` prove the identical contract from one plan for the cases a plan cannot observe.
   **11 plans are duplicated toolchain work** (corrected from rev. 1, which said the same number but
   drew the partition in the wrong place).
4. **`status_test.go:1019/1031` and `1042/1050` — split `--json` then `--check`.** Four invocations
   where two combined ones prove strictly more (one run's document *and* its exit code). The same
   consolidation is already applied at `status_test.go:566` and `global_status_test.go:353`.
   **2 plans duplicated.**
5. **`status_test.go:344` and `:373` — the clean phase still splits `status --json` and
   `status --check`.** **1 plan duplicated.** This is now an explicitly ranked step (**R3.1**), not a
   footnote.

## 5. Ranked patch plan

All changes are **test-only, in `cmd/curator/status_test.go`**. No product file, no
`lifecycle_conformance_test.go`, no `global_status_test.go`, no schema, golden, registry, pin or
config change. The task delta keeps its current shape (added/modified tests only, zero product
delta).

### R1 — replace the four fixture reinstalls with a cache snapshot  *(largest, lowest risk)*

- **File/function:** `cmd/curator/status_test.go`,
  `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck`, line `1216`.
- **Change:** delete `t.Cleanup(func() { reinstall(t) })`; call `snapshotBuildCacheAfter(t, home)`
  immediately **before** `move(t, ...)` at line `1215`. Update the stale comment at `1200-1202`.
- **Why it is safe:** `snapshotBuildCacheAfter` (`global_status_test.go:185`) captures every byte and
  permission bit under `home/cache/build`, makes the protected tree writable via `clearProtection`
  (`:256`), replaces it, and restores modes deepest-written-last-first (`:245-249`). All four moves
  are confined to that tree (finding 1 above).
- **Expected saving:** 3 rebuilds + 1 cache-hit install ≈ **31 s** clean, ≈ **36–40 s** at `./...`.
- **Status against the moving candidate:** already implemented upstream — see §9.

### R2 — hoist the plan and marker fingerprint out of the same loop  *(only valid together with R1)*

- **File/function:** same test, lines `1203-1208`.
- **Change:** drop the per-case `install.Project(...)`, `facts := factsList(...)` and
  `before := markerDigests(...)`; reuse the `facts`/`before` derived at `1160-1162`, and take
  `move(t, planned.Builds[0].Expectation().Input)` from the outer `planned` at `1156`.
- **Why it is safe:** with R1 applied the marker is never rewritten and the cache is restored to
  identical bytes and modes, so every case starts from exactly the state the outer plan described.
  Subtest map order is randomized in Go; order-independence holds because each case restores its own
  fixture.
- **Expected saving:** 4 plans ≈ **13 s** clean, ≈ **15–17 s** at `./...`.
- **Status against the moving candidate:** open, but the producer has written an explicit design
  comment preferring per-case re-acquisition — see §9. This is the one recommendation that is a
  contested judgement call rather than a defect.

### R3 — adopt the accepted acquire-once / replay pattern for the project drift matrix  *(corrected)*

- **File/function:** `cmd/curator/status_test.go`,
  `TestStatusReportsCompiledCurrentnessAndFailsCheck`, lines `337-621`.

The step has three named parts. The reviewer's R1 correction lands in **R3.3**; the reviewer's R2
naming requirement lands in **R3.1**.

#### R3.1 — consolidate the clean phase into one invocation  *(saves 1 P)*

Replace the split `capture(t, "status", "app", "--json")` (`:344`) and
`capture(t, "status", "app", "--check")` (`:373`) with a single
`capture(t, "status", "app", "--json", "--check")` asserted `exitOK`, keeping every existing
assertion on the decoded document (row identity fields at `:356-362`, skill map at `:363`, the
no-manager-private-path scan at `:368-372`).

Why it is safe and strictly stronger: the document and the zero exit become one run's verdict instead
of two classifications of independently re-planned state. `cmdStatus` prints the JSON payload at
`main.go:686-693` regardless of the exit code, and sets the exit only through
`if *check && checkFailed(drift, builds)` (`main.go:682`), so a clean scope still exits `exitOK` with
the full document. Identical precedent already accepted at `status_test.go:566` and
`global_status_test.go:351-356`.

#### R3.2 — one frozen plan for the eleven cases a plan cannot observe  *(saves 11 P net of R3.3)*

After the clean phase, acquire one read-only plan with production functions only, exactly the way
`TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck:1149-1162` already does:

```
cfg, code := loadConfig()                                                    // main.go
scope     := projectStatusScope(cfg, project, "app")                         // main.go:716
unchanged := install.Project(cfg, project, "app", install.Options{DryRun: true})
facts     := factsList(unchanged.Builds)                                     // builds.go:694
```

Per replayed case: tamper, then `before := markerDigests(scope.stores...)` **after** the tamper
(exactly the order `main.go:634` uses relative to the plan), then
`drift, rows := statusReport(cfg, scope, facts, before)` (`main.go:752`) and
`checkFailed(drift, rows)` (`builds.go:133`). Assert state / cause / cache-outcome / detail /
skill-demotion / fail-closed off those production values, and the human line off
`rows[0].Describe()`.

The eleven cases are the seven marker-content tampers (`:398`, `:409`, `:419`, `:430`, `:441`,
`:452`, `:461`), the three marker-refusal tampers (`:489`, `:497`, `:506`), and the installed-tree
context exposure (`:510`).

**Why the plan may be frozen for exactly these eleven.** `planOne`
(`internal/install/plan.go:469-530`) derives its `buildmeta.Input` from the build-source token, build
root, command name, source dir, target, toolchain and fixed policy (`:483-493`), then takes the cache
verdict from `cache.Inspect` (`:506`) — it **never** reads the install marker, and it never reads the
installed skill directory. So neither a rewritten/refused marker nor a directory materialized inside
the installed tree can change what a plan derives. The marker is still read live inside the replay,
by `classifySkillBuilds(installed, marker.Read(installed), facts)` at `main.go:791`, so every tamper
is honoured by the classifier. This is the same justification `global_status_test.go:329-343` already
carries, and its context case at `global_status_test.go:482-504` replays the identical shape.

#### R3.3 — three protected-cache cases must classify from a live post-tamper plan  *(the correction)*

The reviewer is right, and the mechanism is confirmed independently in production source:

- planning reads cache evidence — `planOne` → `cache.Inspect` (`plan.go:504-509`) — and embeds the
  observed outcome, and for a hit the receipt hash and artifact, into the planned facts
  (`plan.go:521-528` → `factsOf`, `builds.go:173-190`);
- `plannedEvidence` / `observedEvidence` / `recheckBuildCache` (`builds.go:208-262`) compare the
  planned tuple with a fresh inspection and name any command whose evidence moved;
- `statusReport` (`main.go:795-808`) overwrites the classified state with `buildStateChanged`
  whenever that happens;
- the candidate's own `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck`
  (`:1183-1188`, `:1228-1239`) proves that chmodding those same entries **after** planning must
  produce `buildStateChanged`;
- `global_status_test.go:338-343` and `:471-473` state and apply the governing rule.

Therefore all **three** cache mutations classify from a plan acquired after their own tamper:

| # | case | file:line | tamper | expected verdict |
|---|---|---|---|---|
| 1 | protected cache entry cannot be interpreted | `:473` | `corruptCacheArtifact` | `buildCorruptCache` / `corrupt` |
| 2 | protected cache holds no entry for the recorded key | `:478` | `RemoveAll` over `cacheEntries` | `buildMissingArtifact` / `would-preflight-and-build` |
| 3 | protected cache boundary is no longer provable | `:523` | `os.Chmod(entry, 0o777)` | `buildUntrustedCache` / `would-rebuild-untrusted-cache` |

Case 3 is the one rev. 1 misclassified into the frozen group. Its chmod flips the
`stat.Mode & 0o022` boundary test (`internal/buildcache/protection_unix.go:239,259`), which moves the
inspected outcome from `cache-hit` to untrusted provenance (`internal/buildcache/cache.go:167-168`).
With a frozen `cache-hit` plan the comparison in `recheckBuildCache` fails and the row is published
as `build-state-changed`; keeping the expected `untrusted-build-cache` assertion would be impossible,
and changing it would be a semantic coverage regression.

**Recommended realization — the live post-tamper plan is the real CLI run.** For these three cases the
plan itself is the cost, so an in-process `install.Project` + replay and a real
`capture(t, "status", "app", "--json", "--check")` cost exactly the same **1 P** — and the real run
additionally proves the published document and the fail-closed process exit. So these three cases
keep the code they have today, unchanged, and the requirement "classified from a plan acquired while
the tampered state is live" is satisfied by construction, because the command takes its own plan
after the tamper. (The in-process variant used by `global_status_test.go:466-480` is equally correct
and equally priced; it is only worse in coverage, so it is not recommended here.)

A live post-tamper plan is safe to take for all three: `blocking()` (`plan.go:56-60`) is true only for
`unsupported` and `toolchain-unavailable`, so `corrupt`, `would-preflight-and-build` and
`would-rebuild-untrusted-cache` all return `Status == "ok"` with the tampered evidence in the plan,
and the recheck then compares equal.

Fixture restoration for these three: cases 1 and 2 keep `snapshotCache: true` (`:475`, `:486`); case 3
keeps its `restore` chmod back to `0o700` (`:531-537`). `0o700` clears every `0o022` bit, so the
restored entry inspects as the same hit, with the same receipt hash and artifact, that the frozen plan
of R3.2 describes — the evidence tuple compared by `recheckBuildCache` is `{Outcome, ReceiptHash,
Artifact}` (`builds.go:201-226`) and carries no mode, so an exactly-permissible mode is enough.
Optional hardening: give case 3 `snapshotCache: true` as well and drop its `restore`, which makes all
three cache cases restore byte-and-mode identically for zero extra cost.

#### R3 saving

19 plans → 8 = **11 P ≈ 36 s** clean, ≈ **42–46 s** at `./...` load.

### R4 — combine `--json` and `--check` in the transitive test

- **File/function:** `cmd/curator/status_test.go`,
  `TestStatusReportsATransitivelyResolvedCompiledCommand`, lines `1019/1031` and `1042/1050`.
- **Change:** one `capture(t, "status", "app", "--json", "--check")` per phase; assert `exitOK` for
  the clean phase and `exitFail` for the missing-entry phase from the same run that produced the
  document.
- **Why it is safe:** the JSON payload is printed independently of the exit code
  (`main.go:686-693`), so the drifted phase still yields a decodable document with `exitFail`.
- **Expected saving:** 2 plans ≈ **6.5 s** clean, ≈ **8 s** at `./...`.

### 5.5 Plan-invocation ledger for `TestStatusReportsCompiledCurrentnessAndFailsCheck`

| | today | after R3 | step |
|---|---:|---:|---|
| clean phase `status --json` | 1 | — | R3.1 |
| clean phase `status --check` | 1 | 1 (combined) | R3.1 |
| frozen `install.Project(DryRun)` acquisition | 0 | 1 | R3.2 |
| 11 marker/context cases, replayed through `statusReport` | 11 | 0 | R3.2 |
| 3 protected-cache cases, live post-tamper plan (real CLI) | 3 | 3 | R3.3 |
| 3 `humanCLI` representatives, real `status app` | 3 | 3 | unchanged |
| **total P** | **19** | **8** | **−11** |

The reviewer's *minimal* variant — keep the clean phase split, take the third live cache plan — lands
at **9** and saves 10 P ≈ 33 s clean. It is not recommended: R3.1 is a strict coverage improvement at
zero risk, and the preferred variant is the one every figure below uses.

### Total (against the reviewed snapshot `4c2339dd…`, i.e. the build verifier 2 failed)

| | clean / focused load | `./...` load (×1.17–1.29) |
|---|---:|---:|
| R1 | 31 s | 36–40 s |
| R2 | 13 s | 15–17 s |
| R3 (R3.1 + R3.2 + R3.3) | 36 s | 42–46 s |
| R4 | 6.5 s | 8 s |
| **total** | **≈ 86 s** | **≈ 101–112 s** |

Projected `cmd/curator`: **600.6 s → ≈ 489–500 s** at the load that produced the failure, i.e.
**≈ 100 s of restored margin**. Clean-load projection ≈ 380 s.

Stated precisely, without rounding in this audit's favour: the clean-load saving is **86 s**, below
the task's 90 s bar; the saving **under the `./...` load that actually produced the 600.591 s
failure** is **101–112 s**, above it. The deadline is only ever hit under that load, so the plan
clears the requirement in the model that matters, and the producer must convert the projection into a
measurement using §10.

### Reserve — only if measurement shows the margin is still short

| id | change | saving | coverage cost |
|---|---|---:|---|
| R5 | fold `TestGCRetainsAndReportsReferencedCompiledState` (`:1337`) into an already-installed compiled fixture | ≈ 13 s | couples two top-level tests; largely subsumed by the producer's own shared-fixture axis (§9) — leave it to them |
| R6 | drop the trailing "ordinary reconciliation still repairs it afterwards" block at `:782-787` (both commands) | ≈ 26 s | **real loss** — the only place repair-after-a-reversed-commit is proven. Do not take it without an explicit owner decision |

Neither is recommended as part of the primary patch.

## 6. Assertion-preservation matrix (corrected partition)

| assertion class | today (reviewed snapshot) | after R1–R4 | verdict |
|---|---|---|---|
| 14 drift states, causes, cache outcomes, details | 14 live CLI runs | 11 replayed through `statusReport`, 3 live-CLI post-tamper | preserved — all 14 still classified by the same production classifier |
| skill demotion map per case | 14 | 11 from `drift`, 3 from the CLI document | preserved |
| operator detail non-empty per case | 14 | 14 | preserved |
| human operator line per case | `Describe()` × 14 + 3 CLI pins | `Describe()` × 14 + 3 CLI pins | preserved |
| fail-closed non-zero `--check` per case | 14 real process exits | 11 via `checkFailed` + **3 real process exits** (the cache cases) | semantically preserved; real process exit-code mapping drops from 14 to 3 drift representatives (+ the clean `exitOK`) |
| JSON document assembly (`alias`/`path`/`skills`/`builds`) | 14 decoded documents | clean phase + 3 cache cases = 4 decoded documents | **reduced** — assembly is scope-independent wiring at `main.go:661-668`; the accepted global test takes the same reduction |
| clean current row, all identity fields | live CLI `--json` | live CLI `--json --check`, same assertions | preserved and strengthened (document + exit from one run) |
| no manager-private path in JSON | live CLI | live CLI | preserved |
| marker untouched during the cache-move window | asserted per case | asserted per case | preserved and strengthened (no reinstall can rewrite it) |
| `build-state-changed` verdict for 4 moves | per case | per case | preserved |
| corrupt-receipt / corrupt-artifact repair by install **and** upgrade | dedicated test | unchanged | preserved |
| commit-failure rollback + exact restoration | dedicated test | unchanged | preserved |
| untrusted-cache repair + old install preserved | dedicated test | unchanged | preserved |
| transitive compiled command reported and failing `--check` | 4 CLI runs | 2 combined CLI runs | preserved and strengthened |
| authoritative bootstrap/upgrade lifecycle vectors | `lifecycle_conformance_test.go` | byte-identical | preserved |
| global-scope contract | `global_status_test.go` | byte-identical | preserved |

The only genuine reductions are the two marked: per-case **document assembly** and per-case **real
process exit mapping** drop from 14 representatives to 4. Both are reductions already accepted in
`global_status_test.go`. Note the correction against rev. 1: because the three cache cases stay
live CLI runs, the exit-mapping representatives are three *drift* cases plus the clean case, not
one — strictly better than rev. 1 claimed.

## 7. Risks

1. **Cleanup ordering (R1).** `snapshotBuildCacheAfter` must be registered *before* the move so it
   captures clean bytes. Go cleanups are LIFO; there is no marker restore in this test, so no
   interaction. *Mitigation:* register the snapshot as the first statement after the plan.
2. **Protected-mode restore (R1).** The cache tree is read-only by construction; `clearProtection`
   (`global_status_test.go:256`) is what makes the restore possible. A future entry type that is
   neither a plain file nor a directory would be skipped silently. *Mitigation:* the helper is
   already exercised by two accepted global cases and by two cases in this file (`:559`).
3. **Frozen-plan faithfulness (R2, R3.2).** If a future change ever makes the install marker or the
   installed tree an input to `planOne`, the replay goes stale. *Mitigation:* the three `humanCLI`
   representatives, the three live cache cases and the clean phase run the real command end to end and
   would diverge loudly.
4. **`before` must be taken after the tamper (R3.2).** Taking `markerDigests` before the tamper turns
   every marker case into `build-state-changed` instead of its own code — the exact mechanism
   `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck` relies on. This is the single most
   likely way to get a green-but-wrong patch. *Mitigation:* the expected `want` state per case is
   asserted explicitly, so an inverted order fails immediately.
5. **Cross-case cache fidelity under a frozen plan (new in rev. 2).** The frozen plan of R3.2 stays
   valid for the eleven replayed cases only while the three cache cases restore the entry to
   evidence-identical state. Verified: the compared tuple is `{Outcome, ReceiptHash, Artifact}`
   (`builds.go:201-226`), snapshot restore replays exact bytes and modes, and the chmod restore to
   `0o700` clears every `0o022` bit that the boundary test reads. Subtest map order is randomized, so
   this must hold for any order. *Mitigation:* the optional hardening in R3.3 (snapshot all three)
   removes the reasoning step entirely at zero cost.
6. **No `t.Parallel`.** `capture` (`status_test.go:39`) swaps process-global `os.Stdout`/`os.Stderr`,
   `t.Setenv` mutates process-global `CURATOR_CONFIG` and `CURATOR_GO`, and the drift cases share one
   marker/cache fixture. Parallelism would corrupt all three. Unchanged constraint.
7. **Allowlist name drift (new in rev. 2).** The moving candidate renames the tests these `-run`
   selectors target. §10.2 gives the mechanical re-pointing rule; a `-run` pattern that matches
   nothing exits 0 in Go and would silently prove nothing. *Mitigation:* the producer must confirm a
   non-zero test count per command before trusting any before/after timing.
8. **Second latent deadline — outside this task's file scope, on the same gate.**
   `internal/install/atomicity` measured **543.719 s** and `internal/install` **411.094 s** against
   the same unchanged 600 s package deadline; `atomicity` sits 56 s from its own timeout. Cutting
   `cmd/curator` also cuts host contention and should pull both down, but the next full run may fail
   in `atomicity` instead. Flagged for the timing agent; **no change is proposed for those packages
   here.**
9. **Estimates are derived, not measured.** `P ≈ 3.26 s`, `B ≈ 8–11 s` come from solving three
   preserved measurements; the model reproduces the accepted 465.944 s baseline to within 1 %. The
   producer must attach real focused before/after timings per §10.

## 8. Why no product-side seam is proposed

The accepted global replay is stronger than the project replay because `reportGlobalStatus`
(`main.go:1189`) takes an injectable acquisition function, so the global test replays the whole
command — document assembly, human rendering and exit code included. The project path has no such
seam: `cmdStatus` (`main.go:597-695`) parses flags, selects targets, plans, classifies, marshals and
exits in one function.

Extracting an equivalent seam would recover the two reduced assertion classes in §6 — but it is a
**product** change to `cmd/curator/main.go` inside an RC conformance candidate whose current delta is
deliberately test-only. Enlarging the product review surface to buy back thin wiring coverage is not a
trade this audit recommends, and it is not this task's decision to make. Recorded as an option with
its cost, not proposed.

## 9. Applicability to the concurrently moving candidate

Read-only comparison of the reviewed snapshot against
`.temp/TASK-260720-jrrgw9/worktree/cmd/curator/status_test.go` at `487b12bd…` (279 diff lines). The
primary producer is running a **shared-fixture** rework on a different axis from this audit: it folds
five formerly independent top-level tests into one umbrella test
`TestCompiledProjectStatusRepairRollbackRecovery` over a single `compiledProjectFixture`, removing
duplicate *installs* (`B` units). This audit removes duplicate *plans* (`P` units). The two axes are
additive, not competing.

| ref | state in the moving candidate | consequence |
|---|---|---|
| R1 | **already implemented** — `t.Cleanup(func(){ reinstall(t) })` is gone, `snapshotBuildCacheAfter(t, home)` is called before `move(...)` | its ≈ 31 s is presumably already banked; it is no longer an incremental saving there, but it remains part of the saving measured against the failing baseline |
| R2 | **open, and contested** — the per-case `install.Project(...)` is retained with a new comment stating the plan is deliberately "acquired again here rather than carried in" | a judgement call the primary producer has made the other way. R2 is technically valid on top of their R1 (the snapshot restores evidence-identical state), but it trades their evidence-independence preference for ≈ 13 s. Owner's call, not a defect |
| R3 | **open** — lines `344-620` of the reviewed snapshot are byte-identical in the candidate; only the enclosing function name changed | applies verbatim; ≈ 36 s |
| R4 | **open** — `TestStatusReportsATransitivelyResolvedCompiledCommand` is unchanged | applies verbatim; ≈ 6.5 s |
| R5 | largely subsumed by the producer's own shared-fixture axis | leave to them |

Incremental saving still available on the moving candidate, on top of what it already banks:
**R3 + R4 ≈ 42.5 s** clean (≈ 50–55 s under `./...`), or **R2 + R3 + R4 ≈ 55.5 s** clean
(≈ 65–72 s under `./...`) if the owner also accepts R2. That is *in addition to* whatever the
shared-fixture rework itself saves, which only the timing agent can measure.

This audit makes no claim about the correctness or completeness of the moving candidate; it was read
read-only, at one instant, for applicability only.

## 10. Literal narrow producer command allowlist

Run each command standalone, sequentially, with no other Go command overlapping, and report its real
exit code. No `-timeout` token. No `./...`, no `-race`, no coverage, no whole-package `./cmd/curator`
without `-run`, no Windows or Linux execution, no cache clearing, no host install, no staging, no
commit, no publication, no pin change is authorized in this rework.

### 10.1 Primary allowlist (unchanged from rev. 1 — valid against the reviewed snapshot)

```text
gofmt -l cmd/curator/status_test.go
git diff --check
go vet ./cmd/curator
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run '^TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck$'
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run '^TestStatusReportsCompiledCurrentnessAndFailsCheck$'
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run '^TestStatusReportsATransitivelyResolvedCompiledCommand$'
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run '^(TestInstallAndUpgradeRepairCorruptCompiledState|TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall|TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails|TestDryRunNeverClaimsACompletedCompilerCheck)$'
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run '^(TestAuthoritativeBootstrapCasesAreExecutable|TestAuthoritativeUpgradeCasesAreExecutable)$'
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 ./cmd/curator -run '^TestGCRetainsAndReportsReferencedCompiledState$'
```

Baseline discipline: run commands 4, 5 and 6 **before** the edit and again after, and attach both wall
times, so the saving is a measurement rather than a projection. Commands 7, 8 and 9 are
regression-only and must stay green and unchanged in shape.

### 10.2 Re-pointing rule if the patch is merged onto the shared-fixture candidate

At `487b12bd…` four of the selectors above name functions that are no longer tests. If and only if
the patch lands on that shape, the affected selectors become — same count, same narrowness, no new
scope:

```text
CURATOR_CONFORMANCE_ROOT=<root> go test -count=1 ./cmd/curator -run '^TestCompiledProjectStatusRepairRollbackRecovery$/^status_reports_compiled_currentness_and_fails_check$'
CURATOR_CONFORMANCE_ROOT=<root> go test -count=1 ./cmd/curator -run '^TestCompiledProjectStatusRepairRollbackRecovery$/^status_reports_protected_cache_state_that_moved_during_the_check$'
CURATOR_CONFORMANCE_ROOT=<root> go test -count=1 ./cmd/curator -run '^TestCompiledProjectStatusRepairRollbackRecovery$'
```

with commands 3, 6, 8 and 9 of §10.1 unchanged, and `<root>` still the immutable
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`.

Two cautions. First, the umbrella test shares one installed fixture across its subtests, so a single
subtest run is **not** a like-for-like baseline against the reviewed snapshot's standalone timings —
compare umbrella-to-umbrella. Second, a `-run` pattern that matches nothing exits 0 in Go: the
producer must confirm a non-zero executed-test count before recording any timing. The candidate is
moving, so these names must be re-verified at merge time rather than trusted from this document.

### 10.3 Gate boundary

The producer must remain in `development` after this rework. Only an independent verifier may run the
exact `CURATOR_CONFORMANCE_ROOT=<immutable rc.5 root> go test -count=1 ./...` gate at the unchanged
default timeout, then the race gate, then the native Windows launcher matrix. This audit makes **no**
claim that the full gate is fixed.

## 11. Board decomposition note

No new board element was created, in rev. 1 or rev. 2. The rework this plan feeds is already owned by
`TASK-260720-jrrgw9` (`verify-rc4-build-conformance`, status `development`), and
`TASK-260729-1zex8r` (`optimize-go-toolchain-fingerprint-walk`) already closed the product-side
fingerprint work. A parallel rework task would duplicate an in-flight owner and would not trace to any
requirement the existing board does not already cover, so under the proportional-board rule it was
rejected rather than created.

Sections checked before deciding: this task's own Scope ("Produce a concrete ranked patch plan that
can be merged with the primary timing diagnosis") and Acceptance Criteria; the attached
`TASK-260729-1y7okj_audit-scope.md` ("another timing agent owns measurements", "Do not edit either
worktree"); and `TASK-260729-1y7okj_rework-cycle-1.md` ("Do not edit Curator source/tests/specs, run
Go commands, or interfere with TASK-260720-jrrgw9", "Update the existing task-scoped audit outcome").
All three place the deliverable in a task-scoped outcome, not in new board scope. The product-seam
option in §8 is likewise recorded as an owner decision, not created as scope.

## 12. Boundary attestation for this revision

- No file in `.temp/TASK-260720-jrrgw9/worktree` or `.temp/TASK-260729-2kaopg/worktree` was written,
  and no board element of `TASK-260720-jrrgw9` was mutated.
- No Curator source, test, spec, schema, golden, registry or pin file was modified anywhere.
- No Go test, build, vet, race, coverage or Windows command was executed.
- The only files written are under this task's own scratch path
  `.temp/TASK-260729-1y7okj/` (the reconstructed reviewed snapshot and this artifact) and this task's
  own board resources.
