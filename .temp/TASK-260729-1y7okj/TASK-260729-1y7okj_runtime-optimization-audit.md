# TASK-260729-1y7okj — independent cmd/curator runtime optimization audit

Date: 2026-07-29
Role: solution-architect (independent audit)
Disposition: analysis only, handed off to review. **No source file was modified. No Go test, build,
vet, race, coverage, or Windows command was executed by this audit.** Every timing figure below is
*derived* from the preserved verifier and producer evidence, not measured here; the timing agent owns
measurement.

## 1. Provenance and immutable inputs

| item | value |
|---|---|
| Candidate worktree | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree` |
| Candidate HEAD | `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8` |
| Accepted integrated comparison | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-2kaopg/worktree` |
| Immutable conformance root | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1` |
| `cmd/curator/status_test.go` sha256 | `4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe` |
| `cmd/curator/global_status_test.go` sha256 | `25aebe8550f7f80f57287fe68e5aa39b219462efb34cc382d771bc2dbaf24f71` |
| `cmd/curator/lifecycle_conformance_test.go` sha256 | `5e83281171c62dcedaef3774686f0526849d556e8b9692e03d39f47bb18658a2` |

The `status_test.go` digest is byte-identical to the **post-patch** digest recorded in
`TASK-260720-jrrgw9_status-timing-patch-results.md`, so this audit reads exactly the candidate that
verifier 2 failed, with the one-file timing patch already applied.

Evidence read: `TASK-260720-jrrgw9_final-verifier2-results.md`,
`TASK-260720-jrrgw9_go-test-all-verifier2-failed.log`,
`TASK-260720-jrrgw9_cmd-timing-diagnosis-results.md`,
`TASK-260720-jrrgw9_status-timing-patch-results.md`, `TASK-260720-jrrgw9_second-timing-diagnosis.md`.

## 2. Failure re-confirmed as cumulative, not a hang

The goroutine dump at `go-test-all-verifier2-failed.log:569` shows
`TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands (1s)` as the only running test when the
alarm fired. That test uses `legacyProject` — a script-only skill, no compiled command, no toolchain
session — so it cannot itself consume minutes. It was merely active at t=600s. The stack tail below
it (`transaction.namespaceIdentity` under `install.runCommit`) belongs to a live `os.Lstat`, i.e.
forward progress, not a deadlock. **Diagnosis: cumulative package duration with insufficient margin.**

Full-run package table from the same log: `cmd/curator` FAIL 600.591s, `internal/install` 411.094s,
`internal/install/atomicity` 543.719s, `internal/godriver` 72.516s, everything else below 30s.

## 3. Root cost model (static, derived)

Two product mechanisms account for essentially all `cmd/curator` wall time. Neither is changed by
this plan.

**(a) Toolchain fingerprint — the cost of every plan.**
`internal/install/plan.go:363 planBuilds` resolves the trusted toolchain for *any* scope that
activates a compiled command: `Toolchain.Probe` on a dry run (`plan.go:382`), `Toolchain.Establish`
otherwise (`plan.go:389`). Both land in `internal/godriver/session.go:157 establish`, which runs
`go telemetry off`, `go version`, `go env -json`, and then
`fingerprintToolchain` (`session.go:266` → `internal/godriver/fingerprint.go:34`) — a full walk of
GOROOT that re-opens and SHA-256s **every** file. `godriver.Probe` (`session.go:141`) closes the
session immediately, and `Session.Close` (`session.go:332`) calls `VerifyToolchain`, which
fingerprints GOROOT **a second time** (`session.go:316`). `Probe` is documented as *not* memoizing
(`session.go:140`).

Measured host GOROOT: **233 MB across 14,502 files**. So one plan-only CLI invocation against a
compiled scope = 3 `go` subprocesses + 2 full 233 MB / 14.5k-file digests.

**(b) Cold private GOCACHE — the cost of every real build.**
`internal/godriver/session.go:404 bootstrapEnvironment` points `GOCACHE`, `GOMODCACHE`, `GOPATH`,
`GOTMPDIR` at `layout.*` inside the operation-private root created per operation
(`internal/install/builddeps.go:166`). Every real compiled install therefore compiles the whole
dependency set from an **empty** build cache, every time.

**Unit costs derived from the preserved measurements.** Let `P` = one plan-only compiled CLI
invocation, `B` = one real compiled install/upgrade that rebuilds.

- `TestStatusReportsCompiledCurrentnessAndFailsCheck` pre-patch = `44P + 3B = 167.062s`
- same test post-patch = `19P + 1B = 69.816s`
- verifier-2 focused status/lifecycle/repair group = `37P + 15B + 13.1s(lifecycle) ≈ 267.207s`

Solving gives **P ≈ 3.2 s** and **B ≈ 8–11 s** (a cache-*hit* install without a rebuild ≈ 4 s).
Applying the model to the whole package predicts ≈ 463 s, against the accepted comparable-load clean
derivation of **465.944 s** — the model reproduces the known baseline, so the savings below are
sound. Observed high-load inflation (`545.195 s` / `554.967 s` / `600.591 s` vs 465.944 s) is
**×1.17 to ×1.29**; savings scale by the same factor under `./...`.

## 4. Static call/fixture inventory — `cmd/curator`

`builds_test.go`, `gc_test.go`, `main_test.go` and `lifecycle_conformance_test.go` contain **no**
compiled command (`lifecycle_conformance_test.go:233` publishes `"commands": {}`; the other three
never call `capture` with an install of a schema-6 build skill). The entire compiled cost lives in
`status_test.go` and `global_status_test.go`.

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

### Duplicate expensive work identified

1. **`status_test.go:1216` — `t.Cleanup(func() { reinstall(t) })` inside a four-case loop.**
   All four moves mutate **only** `home/cache/build`: `corruptCacheArtifact` (`status_test.go:270`),
   `os.RemoveAll` over `cacheEntries` (`status_test.go:1177`), `os.Chmod(entry, 0o777)`
   (`status_test.go:1185`), and `replaceCacheEntry` (`status_test.go:890`). Nothing touches the
   installed tree — the test itself asserts the marker is byte-identical (`status_test.go:1220-1226`).
   Repairing a *reporting* fixture with four real compiled reconciliations is exactly the pattern the
   accepted `snapshotBuildCacheAfter` helper (`global_status_test.go:185`) already replaced elsewhere
   in this package. **3 rebuilds + 1 cache-hit install are pure fixture cost.**
2. **`status_test.go:1203` — the per-case `install.Project(..., DryRun: true)` re-plan.** Its own
   comment (`status_test.go:1200-1202`) states it exists *because* a preceding case reinstalled and
   rewrote the marker. Remove the reinstall and the reason disappears: the outer plan at
   `status_test.go:1156` and the outer `before` at `1162` become valid for every case. **4 plans are
   pure consequence of finding 1.**
3. **`status_test.go:377-620` — 14 live compiled-scope CLI acquisitions where the accepted global
   counterpart uses one.** `global_status_test.go:386 acquireGlobalPlan` + `:149 replay` prove the
   identical contract from **one** plan, with per-shape end-to-end CLI pins. The project-scope test
   still re-derives a fresh full plan (2 GOROOT fingerprints each) for every drift case. **11 plans
   are duplicated toolchain work.**
4. **`status_test.go:1019/1031` and `1042/1050` — split `--json` then `--check`.** Four invocations
   where two combined ones prove strictly more (one run's document *and* its exit code). The same
   consolidation is already applied at `status_test.go:566` and `global_status_test.go:353`.
   **2 plans duplicated.**
5. **`status_test.go:344/373` — the clean phase still splits `status --json` and `status --check`.**
   1 plan duplicated; folded into item 6 of the patch plan.

## 5. Ranked patch plan

All changes are **test-only, in `cmd/curator/status_test.go`**. No product file, no
`lifecycle_conformance_test.go`, no `global_status_test.go`, no schema, golden, registry, pin or
config change. This keeps the task delta at its current shape (20 added tests + 3 modified tests,
zero product delta).

### R1 — replace the four fixture reinstalls with a cache snapshot  *(largest, lowest risk)*

- **File/function:** `cmd/curator/status_test.go`, `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck`, line `1216`.
- **Change:** delete `t.Cleanup(func() { reinstall(t) })`; call `snapshotBuildCacheAfter(t, home)`
  immediately **before** `move(t, ...)` at line `1215`. Update the stale comment at `1200-1202`.
- **Why it is safe:** `snapshotBuildCacheAfter` (`global_status_test.go:185`) captures every byte and
  permission bit under `home/cache/build`, makes the protected tree writable, replaces it, and
  restores modes deepest-last-written-first. Every one of the four moves is confined to that tree.
- **Expected saving:** 3 rebuilds + 1 cache-hit install ≈ **31 s** clean, ≈ **36–40 s** at `./...` load.

### R2 — hoist the plan and marker fingerprint out of the same loop  *(only valid together with R1)*

- **File/function:** same test, lines `1203-1208`.
- **Change:** drop the per-case `install.Project(...)`, `facts := factsList(...)` and
  `before := markerDigests(...)`; reuse `facts`/`before` derived at `1160-1162` and take
  `move(t, planned.Builds[0].Expectation().Input)` from the outer `planned` at `1156`.
- **Why it is safe:** with R1 applied the marker is never rewritten and the cache is restored to
  identical bytes, so every case starts from the state the outer plan described. Subtest map order
  is randomized in Go; order-independence holds because each case restores its own fixture.
- **Expected saving:** 4 plans ≈ **13 s** clean, ≈ **15–17 s** at `./...` load.

### R3 — adopt the accepted acquire-once / replay pattern for the project drift matrix

- **File/function:** `cmd/curator/status_test.go`, `TestStatusReportsCompiledCurrentnessAndFailsCheck`, lines `337-621`.
- **Change:** after the clean phase, acquire one read-only plan the way
  `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck` already does with production
  functions only —
  `install.Project(cfg, project, "app", install.Options{DryRun: true})` → `factsList` →
  `projectStatusScope`. Per case: tamper, then `before := markerDigests(scope.stores...)`
  (**after** the tamper, exactly as `main.go:634` does relative to the tamper), then
  `drift, rows := statusReport(cfg, scope, facts, before)` (`main.go:752`) and
  `checkFailed(drift, rows)` (`builds.go:133`). Assert state/cause/cache-outcome/detail/
  skill-demotion/fail-closed off those production values and the human line off
  `rows[0].Describe()`.
  Keep live full-CLI runs for: the clean phase, the two cache-damage cases (their plan input really
  changes), and the three `humanCLI` representatives already marked at `status_test.go:407, 428, 539`.
- **Why the plan may be frozen:** `planOne` (`internal/install/plan.go:400`) takes the build-source
  token, target, toolchain and cache — **never** the install marker. The eleven marker/context cases
  cannot change what a plan derives; the two cache cases can, and keep their own live plans. This is
  the identical justification `global_status_test.go:329-343` already carries.
- **Expected saving:** 19 plans → 8 ≈ **36 s** clean, ≈ **42–46 s** at `./...` load.

### R4 — combine `--json` and `--check` in the transitive test

- **File/function:** `cmd/curator/status_test.go`, `TestStatusReportsATransitivelyResolvedCompiledCommand`, lines `1019/1031` and `1042/1050`.
- **Change:** one `capture(t, "status", "app", "--json", "--check")` per phase; assert `exitOK` for
  the clean phase and `exitFail` for the missing-entry phase from the same run that produced the
  document.
- **Expected saving:** 2 plans ≈ **6.5 s** clean, ≈ **8 s** at `./...` load.

### Total

| | clean / focused load | `./...` load (×1.17–1.29) |
|---|---:|---:|
| R1 | 31 s | 36–40 s |
| R2 | 13 s | 15–17 s |
| R3 | 36 s | 42–46 s |
| R4 | 6.5 s | 8 s |
| **total** | **≈ 86 s** | **≈ 101–111 s** |

Projected `cmd/curator`: **600.6 s → ≈ 490–500 s** at the load that produced the failure, i.e.
**≈ 100 s of restored margin**, above the 90 s requirement. Clean-load projection ≈ 380 s.

### Reserve — only if measurement shows the margin is still short

| id | change | saving | coverage cost |
|---|---|---:|---|
| R5 | fold `TestGCRetainsAndReportsReferencedCompiledState` (`status_test.go:1337`) into the clean phase of `TestStatusReportsCompiledCurrentnessAndFailsCheck`, reusing its installed fixture | ≈ 13 s | couples two top-level tests; loses independent-fixture isolation for the gc claim |
| R6 | drop the trailing "ordinary reconciliation still repairs it afterwards" block at `status_test.go:782-787` (both commands) | ≈ 26 s | **real loss** — that block is the only place repair-after-a-reversed-commit is proven. Do not take it without an explicit owner decision. |

R5 and R6 are listed for completeness and are **not** recommended as part of the primary patch.

## 6. Assertion-preservation matrix

| assertion class | today | after R1–R4 | verdict |
|---|---|---|---|
| 14 drift states, causes, cache outcomes, details | 14 live CLI runs | 11 replayed through `statusReport`, 2 live-plan, 1 live CLI | preserved (same production classifier) |
| fail-closed non-zero `--check` per case | 14 CLI exits | 11 via `checkFailed`, 3+ via real CLI exits | preserved semantically; **the CLI exit-code mapping is exercised for 4 representatives instead of 14** |
| JSON document assembly (`alias`/`path`/`skills`/`builds`) | 14 decoded documents | clean phase + 4 representatives | **reduced** — assembly is `main.go:662-668` and is scope-independent; the global test accepts the same reduction |
| skill demotion map | per case | per case (from `drift`) | preserved |
| human operator line | `Describe()` for all 14 + 3 CLI pins | unchanged | preserved |
| clean current row (all identity fields) | live CLI | live CLI | preserved |
| no manager-private path in JSON | live CLI | live CLI | preserved |
| marker untouched during the cache-move window | asserted per case | asserted per case | preserved and strengthened (no reinstall can rewrite it) |
| `build-state-changed` verdict for 4 moves | per case | per case | preserved |
| corrupt-receipt / corrupt-artifact repair by install **and** upgrade | dedicated test, unchanged | unchanged | preserved |
| commit-failure rollback + exact restoration | dedicated test, unchanged | unchanged | preserved |
| untrusted-cache repair + old install preserved | dedicated test, unchanged | unchanged | preserved |
| authoritative bootstrap/upgrade lifecycle vectors | unchanged file | unchanged file | preserved (byte-identical) |
| global-scope contract | unchanged file | unchanged file | preserved (byte-identical) |
| transitive compiled command reported and failing `--check` | 4 CLI runs | 2 combined CLI runs | preserved and strengthened (one run's document + exit) |

The only genuine reductions are the two marked above: per-case CLI **exit mapping** and per-case
**document assembly** drop from 14 representatives to 4. Both are already-accepted reductions in
`global_status_test.go`, which review has passed.

## 7. Risks

1. **Cleanup ordering (R1).** `snapshotBuildCacheAfter` must be registered *before* the move so it
   captures clean bytes, and `denyWrites`-style cleanups are not involved here. LIFO cleanup means a
   later-registered cleanup runs first; there is no marker restore in this test, so no interaction.
   *Mitigation:* register the snapshot as the first statement of the subtest body after the plan.
2. **Protected-mode restore (R1).** The cache tree is read-only by construction. `clearProtection`
   (`global_status_test.go:256`) is what makes the restore possible; if a future entry type is not a
   plain file or directory the restore would silently skip it. *Mitigation:* the existing helper is
   already exercised by two accepted global cases and by two cases in this same file
   (`status_test.go:559`).
3. **Frozen-plan faithfulness (R2, R3).** If a future change makes the install marker an input to
   `planOne`, replay would go stale. *Mitigation:* the three `humanCLI` representatives plus the
   clean phase run the real command end to end and would diverge loudly.
4. **`before` must be taken after the tamper (R3).** Taking `markerDigests` before the tamper turns
   every marker case into `build-state-changed` instead of its own code — the exact mechanism
   `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck` relies on. This is the single most
   likely way to get a green-but-wrong patch. *Mitigation:* the expected `want` state per case is
   asserted explicitly, so an inverted order fails immediately rather than silently.
5. **No `t.Parallel`.** `capture` (`status_test.go:39`) swaps process-global `os.Stdout`/`os.Stderr`,
   `t.Setenv` mutates process-global `CURATOR_CONFIG` and `CURATOR_GO`, and the drift cases share one
   marker/cache fixture. Parallelism would corrupt all three. Unchanged constraint.
6. **Second latent deadline — outside this task's file scope but on the same gate.**
   `internal/install/atomicity` measured **543.719 s** and `internal/install` **411.094 s** against
   the same unchanged 600 s package deadline. `atomicity` sits 56 s from its own timeout. Cutting
   `cmd/curator` also cuts host contention and should pull both down, but if the next full run fails
   it may fail in `atomicity` rather than `cmd/curator`. Flagged for the timing agent; **no change is
   proposed for those packages here.**
7. **Estimates are derived, not measured.** `P ≈ 3.2 s` and `B ≈ 8–11 s` come from solving three
   preserved measurements. The model reproduces the accepted 465.944 s baseline to within 1 %, but
   the producer must attach real focused before/after timings per the allowlist below.

## 8. Literal narrow producer command allowlist

Run each command standalone, sequentially, with no other Go command overlapping, and report its real
exit code. No `-timeout` token. No `./...`, no `-race`, no coverage, no whole-package `./cmd/curator`
without `-run`, no Windows or Linux execution, no cache clearing, no host install, no staging, no
commit, no publication, no pin change is authorized in this rework.

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

Baseline discipline: run command 4, 5 and 6 **before** the edit and again after, and attach both wall
times, so the saving is a measurement rather than a projection. Commands 7, 8 and 9 are
regression-only and must stay green and unchanged in shape.

The producer must remain in `development` after this rework. Only an independent verifier may run the
exact `CURATOR_CONFORMANCE_ROOT=<immutable rc.5 root> go test -count=1 ./...` gate at the unchanged
default timeout, then the race gate, then the native Windows launcher matrix. This audit makes **no**
claim that the full gate is fixed.

## 9. Board decomposition note

No new board element was created. The rework this plan feeds is already owned by
`TASK-260720-jrrgw9` (`verify-rc4-build-conformance`, status `development`), and
`TASK-260729-1zex8r` (`optimize-go-toolchain-fingerprint-walk`) already closed the product-side
fingerprint work. Creating a parallel rework task would duplicate an existing in-flight owner and
would not trace to any requirement the existing board does not already cover, so under the
proportional-board rule it was rejected rather than created. Spec sections checked before deciding:
this task's own Scope ("Produce a concrete ranked patch plan that can be merged with the primary
timing diagnosis") and Acceptance Criteria, plus the attached `TASK-260729-1y7okj_audit-scope.md`
instruction ("another timing agent owns measurements", "Do not edit either worktree"). Both place the
deliverable in a task-scoped outcome, not in new board scope.
