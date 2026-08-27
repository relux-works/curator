# TASK-260720-jrrgw9 — one-file cmd/curator status timing patch

Date: 2026-07-29
Role: developer (producer rework)
Disposition: handed off to review; **the producer does not claim the full suite is fixed**

## Scope

Exactly one file changed: `cmd/curator/status_test.go`, exclusively inside
`TestStatusReportsCompiledCurrentnessAndFailsCheck` and its own case table. No product code, no
`lifecycle_conformance_test.go`, no other test, schema, golden, fixture, config or pin was touched.

Candidate worktree: `/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`
HEAD `17804ceae42d3bb62abf6e8a9d9cbef6dd5370b8`, unchanged. Nothing staged, committed, published or
repinned (`git diff --cached` empty).

## File identity

| | value |
|---|---|
| pre-patch `cmd/curator/status_test.go` sha256 | `c62d07913f7a35b546a84ebfe9d87562d9da0e4074d01dfcd22b092a4717ae87` |
| pre-patch md5 | `0ce3992c794c6b2a357f1d5fc98313dc` |
| post-patch sha256 | `4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe` |
| post-patch md5 | `4f0e42698f0a1f337dce0bf9afaca42f` |
| unified diff sha256 | `03c637da916ea5526298ec6ac41d0f234ccda4bfce098acc63f9e3dc29984acd` |
| diff size | 154 lines |

The pre-patch bytes are the accepted integrated tree's own copy at
`.temp/TASK-260729-2kaopg/worktree/cmd/curator/status_test.go` (md5 matched before the edit), so the
attached diff is a true before/after of this task's single change.

`rsync -nrc` from the accepted integrated worktree into the candidate lists exactly three differing
source files: `cmd/curator/status_test.go` (this patch) plus the two pre-existing task-delta files
`internal/buildcache/conformance_test.go` and `internal/closure/conformance_test.go`. No product,
schema, golden or pin delta was introduced by this rework.

## What changed

1. **One acquisition per drift case.** Each of the 14 cases now runs `status app --json --check`
   once and asserts `exitFail`, then decodes the same document. Previously each case ran
   `status --json`, then `status --check`, then plain `status` — three fresh compiled plans against
   live state that a second classification could legitimately re-read differently.
2. **Human rendering through production code.** Every state and cause is asserted through
   `buildReport.Describe()` — the exact method `cmd/curator/main.go:679` prints — applied to the row
   the run published. Three representative cases additionally run the whole plain-text command path
   and assert the printed line contains `"app: " + described`, which is strictly stronger than the
   pre-patch `state=` / `cause=` substring checks:
   - no-cause marker drift: *artifact hash recorded by the marker no longer matches the entry*
   - cause-bearing input drift: *logical key recorded by the marker was derived from another build input*
   - cache-boundary drift: *protected cache boundary is no longer provable*
3. **Byte-and-mode fixture reset for the two cache-damage cases.** `snapshotBuildCacheAfter` replaces
   the two `restore: reinstall` callbacks. Recompiling the command to reset a *reporting* fixture
   proved nothing this test owns; install, repair and rollback keep their own dedicated real tests,
   which are unchanged and were re-run green (below).
4. **Cleanup ordering.** The marker restore is now registered *before* `snapshotBuildCacheAfter`, so
   under LIFO cleanup the cache snapshot restores first and the marker restores last.

Not done, deliberately: no `t.Parallel`, no skip, no weakened or static-only substitute, no shared
mutable fixture concurrency, no timeout change. The running test binary was observed with
`-test.timeout=10m0s`, the unchanged default.

## Assertion preservation

All 14 drift cases retained. All state, cause, cache-outcome, detail, row-identity, skill-demotion
and fail-closed assertions retained. Full matrix in the attached
`TASK-260720-jrrgw9_status-timing-assertion-matrix.md`. Summary of the only two reductions:

- the separate `--check` invocation is folded into the combined one, so the document and the exit
  code are now provably one run's verdict rather than two;
- the per-case plain-CLI invocation drops from 14 to 3, with `Describe()` covering all 14 and the 3
  survivors pinning `Describe()` to the real command output.

Compiled plans executed inside the case loop: **44 before, 17 after** (14 `--json` + 14 `--check` +
14 plain + 2 fixture reinstalls → 14 combined + 3 plain + 0 reinstalls).

## Executable validation — every command standalone, real exit code

Host: Darwin 25.5.0. `CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1`
on every `go test`. No command contained the token `-timeout`. Runs were sequential; no two Go
suites overlapped.

| # | command | exit | wall |
|---|---|---|---|
| 1 | `gofmt -l cmd/curator/status_test.go` (no output) | 0 | — |
| 2 | `git diff --check` | 0 | — |
| 3 | `go vet ./cmd/curator` | 0 | — |
| 4 | `go test -count=1 ./cmd/curator -run '^TestStatusReportsCompiledCurrentnessAndFailsCheck$'` **(pre-patch baseline)** | 0 | 167.062s |
| 5 | `go test -count=1 ./cmd/curator -run '^TestStatusReportsCompiledCurrentnessAndFailsCheck$'` **(patched)** | 0 | 69.816s |
| 6 | `go test -count=1 ./cmd/curator -run '^(TestAuthoritativeBootstrapCasesAreExecutable\|TestAuthoritativeUpgradeCasesAreExecutable)$'` | 0 | 13.125s |
| 7 | `go test -count=1 ./cmd/curator -run '^(TestInstallAndUpgradeRepairCorruptCompiledState\|TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall\|TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails\|TestDryRunNeverClaimsACompletedCompilerCheck)$'` | 0 | 218.923s |

The baseline (#4) was launched before the edit and measured the pre-patch binary that `go test` had
already compiled; the source edit landed after that compile and could not affect it.

### Timing

| | before | after | delta |
|---|---|---|---|
| `TestStatusReportsCompiledCurrentnessAndFailsCheck` | 167.062s | 69.816s | **−97.2s (−58.2%)** |

The saving exceeds the diagnosis estimate of 70–80s. The accepted high-load `cmd/curator` package
baselines were 545.195s and 554.967s against an unchanged 600s package deadline, and this task adds
the 12.8–15.2s authoritative lifecycle group. Removing 97s restores roughly 100s of margin.
**This is a projection, not a measurement:** no whole `cmd/curator` package, `./...`, race, Windows,
Linux, coverage or host-install command was authorized or run in this pass, so the producer makes no
claim that the full suite now fits the deadline. Only the independent Codex verifier can establish
that.

## Not run and not claimed

`go test ./...`, `go test -race ./...`, coverage measurement, whole-package `cmd/curator`, Windows
and Linux gates, cache clearing, host installs, staging, commits, publication and pin changes were
all outside the authorized allowlist for this rework and were **not executed**. Nothing above is
presented as covering them.

## Process and disk cleanup

| | value |
|---|---|
| Go test / build processes after the last gate | 0 |
| Disk free before gates | 22,871,292 KB |
| Disk free after gates | 22,921,004 KB (net +49,712 KB) |
| `${TMPDIR}go-build*` directories left by these runs | 0 |

47 `go-build*` directories totalling 1,659,452 KB predate this session (newest mtime 12:30, all gates
ran from 14:43). They belong to earlier runs, were not created by this pass, and were **not** removed —
no shared cache or foreign artifact was touched.

## Handoff

Ready for independent full verification. The next verifier should re-run the exact required
`CURATOR_CONFORMANCE_ROOT=<immutable rc.5 root> go test -count=1 ./...` at the unchanged default
timeout, then the full race and native Windows launcher gates.
