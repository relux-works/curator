# TASK-260729-1y7okj — reviewed-snapshot provenance record (cycle-1 rework)

Date: 2026-07-29
Purpose: make the baseline of `TASK-260729-1y7okj_runtime-optimization-audit.md` (rev. 2)
independently reproducible, and separate it from the concurrently moving primary candidate owned by
`TASK-260720-jrrgw9`.

## 1. The audited baseline is a frozen snapshot, not a live directory

| item | value |
|---|---|
| logical identity | the `cmd/curator/status_test.go` that verifier 2 failed on |
| sha256 | `4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe` |
| materialized read-only at | `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-1y7okj/reviewed-snapshot/status_test.go` |

### Reconstruction (byte-exact, reproducible)

Base: `.temp/TASK-260729-2kaopg/worktree/cmd/curator/status_test.go`
= `c62d07913f7a35b546a84ebfe9d87562d9da0e4074d01dfcd22b092a4717ae87`
— the pre-patch digest recorded in `TASK-260720-jrrgw9_status-timing-patch-results.md:21`.

Patch: board resource `TASK-260720-jrrgw9_status-timing-patch.diff`
= `03c637da916ea5526298ec6ac41d0f234ccda4bfce098acc63f9e3dc29984acd`
— the unified-diff digest recorded at `…patch-results.md:25`.

Result: applying that patch to that base yields
`4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe`
— the post-patch digest recorded at `…patch-results.md:23`, and the digest the reviewer confirmed at
review start.

Both source files were read only; the reconstruction wrote exclusively into this task's own
`.temp/TASK-260729-1y7okj/` scratch path.

## 2. Files that did not move between the candidate and the accepted worktree

Compared `.temp/TASK-260720-jrrgw9/worktree` against `.temp/TASK-260729-2kaopg/worktree`:

| file | sha256 (prefix) | result |
|---|---|---|
| `cmd/curator/builds.go` | `6aab0a0fd753f74d…` | identical |
| `cmd/curator/main.go` | `a0798e7dcc9a3e92…` | identical |
| `internal/install/plan.go` | `b9684fdec4b05dde…` | identical |
| `internal/godriver/session.go` | `a4186cd2254c4911…` | identical |
| `internal/godriver/fingerprint.go` | `560d0c98c665a5a8…` | identical |
| `cmd/curator/global_status_test.go` | `25aebe8550f7f80f…` | identical |
| `cmd/curator/lifecycle_conformance_test.go` | `5e83281171c62dce…` | candidate only (absent from the accepted worktree) |

Every product-source line reference in the audit is against those digests.

**Trap for the next reader.** The repo-root main checkout
`/Users/iv/Developer/ReluxWorks/curator/cmd/curator/main.go` is a *different, pre-feature* file,
sha256 `1d2cc9c9310b51f8…`; its `cmdStatus` carries no compiled-command surface, no `statusReport`
call and no `checkFailed` exit. Reading product line numbers from the main checkout instead of the
worktree silently produces wrong references.

## 3. The moving primary candidate

`.temp/TASK-260720-jrrgw9/worktree/cmd/curator/status_test.go`:

| point in time | sha256 (prefix) | source |
|---|---|---|
| audit cycle 1 was written against | `4c2339dd…` | one-file timing patch under `TASK-260720-jrrgw9` |
| 2026-07-29 15:54:27 +0400 | `bb21c15c…` | primary producer, recorded in the cycle-1 review verdict |
| mtime 2026-07-29 15:54:47 +0400, read here 15:57 | `487b12bd…` | primary producer |

`global_status_test.go` and `lifecycle_conformance_test.go` did not move across those points.

### What the moving candidate is doing

A **shared-fixture** rework on a different cost axis: five formerly independent top-level tests
(`TestStatusReportsCompiledCurrentnessAndFailsCheck`,
`TestInstallAndUpgradeRepairCorruptCompiledState`,
`TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails`,
`TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck`,
`TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall`) become subtests of one umbrella
test `TestCompiledProjectStatusRepairRollbackRecovery` sharing a single `compiledProjectFixture`. It
removes duplicate real *installs*; the audit removes duplicate *plans*. The axes are additive.

Directly relevant hunks (read-only, at `487b12bd…`):

- the umbrella test and `newInstalledCompiledProject` replace five per-test
  `compiledProject(t)` + `capture(t, "install", "app")` setups;
- in `assertProtectedCacheStateThatMovedDuringTheCheck`, `t.Cleanup(func(){ reinstall(t) })` is gone
  and `snapshotBuildCacheAfter(t, home)` is now called before `move(...)` — **the audit's R1 is
  already implemented upstream**;
- the per-case `install.Project(..., DryRun: true)` is **retained**, with a new comment stating the
  plan is deliberately re-acquired per case rather than carried in — **the audit's R2 is open and
  contested by an explicit design choice**;
- lines `344-620` of the reviewed snapshot (the clean phase, the 14-case drift matrix, the three
  `humanCLI` representatives) are byte-identical in the candidate, and
  `TestStatusReportsATransitivelyResolvedCompiledCommand` is unchanged — **R3 and R4 apply
  verbatim**.

Consequence for the allowlist: four `-run` selectors in the audit's §10.1 name functions that are no
longer tests at `487b12bd…`. §10.2 of the audit gives the mechanical re-pointing rule, with the
caution that a Go `-run` pattern matching nothing exits 0.

This record makes no claim about the correctness or completeness of the moving candidate. It was read
read-only, at one instant, solely to state which recommendations still apply.

## 4. Boundary attestation

- No file in `.temp/TASK-260720-jrrgw9/worktree` or `.temp/TASK-260729-2kaopg/worktree` was written.
- No Curator source, test, spec, schema, golden, registry or pin file was modified anywhere.
- No Go test, build, vet, race, coverage or Windows command was executed.
- No board element of `TASK-260720-jrrgw9` was mutated.
- Files written: `.temp/TASK-260729-1y7okj/reviewed-snapshot/status_test.go`,
  `.temp/TASK-260729-1y7okj/runtime-optimization-audit.md`,
  `.temp/TASK-260729-1y7okj/reviewed-snapshot-provenance.md`, and this task's own board resources.
