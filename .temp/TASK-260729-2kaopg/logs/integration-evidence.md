# TASK-260729-2kaopg integration evidence — RUN-260729-524af6

Date: 2026-07-29
Role: developer (integration cycle)
Tree: `.temp/TASK-260729-2kaopg/worktree` (base `origin/main` 17804ce, nothing staged or committed)

## 1. Accepted fingerprint artifact integrity

| Check | Result |
| --- | --- |
| `shasum -a 256 TASK-260729-1zex8r_fingerprint-cycle3.patch` | `a7e0906612ce6f007bfdb3776de632dd9c7a673e9b501443be5fb3eced8f1beb` — matches the directive and the reviewer verdict |
| Reviewer verdict `TASK-260729-1zex8r_reviewer-verdict-cycle3.md` | accepted |
| Patch surface | `internal/godriver/fingerprint.go` (modified) + `internal/godriver/fingerprint_equivalence_test.go` (new) — nothing else |

## 2. Base-identity precondition: FAILED as stated, then repaired

The directive asked to prove the target starts from the accepted currentness
production base plus only the global-status delta. **It did not.**

`.temp/TASK-260729-2kaopg/worktree` mirrored an older `TASK-260720-1nlmvv`
snapshot. Accepted-tree files carried timestamps up to 2026-07-29 09:13 while the
target's corresponding files were dated 2026-07-28 16:29 – 2026-07-29 04:25.
Measured divergence: **2396 diff lines across 15 files** plus one file missing
entirely.

### 2a. Accepted 1nlmvv work absent from the target (would have been silently reverted)

- `cmd/curator/builds.go` — `buildFacts.Expectation`, `cacheEvidence`,
  `plannedEvidence`, `observedEvidence`, `recheckBuildCache`, and the
  `retained` argument of `repairNotices`
- `cmd/curator/main.go` — `changedDuringCheck` and the protected-cache recheck
  wiring inside `statusReport`
- `cmd/curator/status_test.go` — 437 lines, including
  `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck` and
  `TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails`
- `internal/buildcache/` — `cache.go`, `cache_test.go`, `publish.go`, and
  `compensation_test.go` (absent from the target altogether)
- `internal/install/` — `commit.go`, `commit_test.go`, `global.go`, `install.go`,
  `plan.go`, `stage_test.go`, `targets.go`
- `README.md` — the cache-reversal and verdict-bracketing sections, and the
  richer `build-state-changed` vocabulary row

### 2b. Global-status delta owned by this task (preserved verbatim)

- `cmd/curator/main.go` — `statusScope`, `projectStatusScope`,
  `globalStatusScope`, the `statusReport`/`installedSkillDir` rescoping,
  `cmdGlobalStatus`, `globalStatusPlan`, and the usage line
- `cmd/curator/global_status_test.go` — new, unchanged from the accepted cycle
- `cmd/curator/builds_test.go`, `cmd/curator/status_test.go` — call-site rewires
  only
- `README.md` — the global-status contract section

### 2c. Repair taken

Rather than integrating on a stale base — which would have reverted accepted
`TASK-260720-1nlmvv` work at merge time — the owned delta was rebased onto the
accepted tree:

- 13 files taken from the accepted tree wholesale (this task owns nothing in
  them);
- `cmd/curator/main.go`, `cmd/curator/builds_test.go`,
  `cmd/curator/status_test.go`, and `README.md` hand-merged: accepted content as
  the base, owned delta re-applied on top;
- `cmd/curator/global_status_test.go` kept as is.

No assertion was altered. The `status_test.go` and `builds_test.go` deltas are
exclusively `statusStores(cfg, project)` → `scope.stores` and
`statusReport(cfg, project, …)` → `statusReport(cfg, scope, …)` rewires required
by the owned API change.

The accepted `changedDuringCheck` message wording is preserved; the older
message this task's tree carried was discarded in favour of the accepted one.

### 2d. Base identity after the repair

`diff -rq` between the accepted tree and the integrated tree now reports exactly
the owned delta plus the accepted fingerprint patch and nothing else:

```text
README.md                                       (global-status contract section)
cmd/curator/main.go                             (owned global-status delta)
cmd/curator/builds_test.go                      (owned call-site rewire)
cmd/curator/status_test.go                      (owned call-site rewire)
cmd/curator/global_status_test.go               (owned, new)
internal/godriver/fingerprint.go                (accepted fingerprint patch)
internal/godriver/fingerprint_equivalence_test.go (accepted fingerprint patch, new)
```

## 3. Fingerprint patch application

| Step | Exit | Detail |
| --- | ---: | --- |
| `git apply --check --whitespace=error-all` | 0 | mechanically applicable; no conflicts |
| `git apply --whitespace=error-all` | 0 | applied without redesign |

Post-apply hashes:

- `internal/godriver/fingerprint.go` — `560d0c98c665a5a83c3a6989a7b0cdcc9f26c4fb513c7688d9b1bd6e42552d1d`
- `internal/godriver/fingerprint_equivalence_test.go` — `6390e75c9848f575f2f4b50217ebd1d53481a58d349073fb0e819491b5fed484`

`internal/godriver` was byte-identical between the accepted tree and the target
before the patch (`fingerprint.go` sha256
`0fc34a359d4c8b207759b133b6f007ffa6b81af0e05156ce0852e778a97e327f`), which is why
the patch applied with no conflict despite the base divergence above.

## 4. Focused and static gates

Every gate ran as a standalone process, no `tee`, no pipe. Exit codes are real
process exit codes.

| Gate | Exit | Detail |
| --- | ---: | --- |
| `go build ./...` (post-rebase, pre-patch) | 0 | |
| `go vet ./...` (post-rebase, pre-patch) | 0 | |
| `go build ./...` (post-patch) | 0 | |
| `go vet ./...` (post-patch) | 0 | |
| `gofmt -l .` | 0 | no output |
| `git diff --check` | 0 | no output |
| fingerprint mutation/equivalence matrix — `go test -count=1 -run 'TestFingerprint\|TestToolchainWalk' -v ./internal/godriver` | 0 | 3.005s; 59 PASS, 0 FAIL, 1 SKIP (`TestFingerprintRejectsInvalidUnicodePath`, platform-dependent, pre-existing) |
| `go test -count=1 ./internal/godriver` | 0 | 27.288s |
| pinned `CURATOR_CONFORMANCE_ROOT` `go test -count=1 ./internal/godriver` | 0 | 28.531s |
| go-v1 vector gate — `TestToolchainFramingMatchesRC4Vector`, `TestFingerprintImplementationMatchesRC4ToolchainVector` | 0 | both PASS |
| focused global-status — `go test -count=1 -coverprofile -run '^TestGlobalStatus' ./cmd/curator` | 0 | 110.541s (was 147.322s pre-patch) |
| `go tool cover -func` | 0 | see below |
| rewired accepted regressions — `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck`, `TestStatusReportMarksCompiledStateThatMovedDuringTheCheck`, `TestStatusReportReportsCompiledCommandsOfAnUninstalledSkill`, `TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails` | 0 | 132.893s, all PASS |

Conformance root used: `.temp/TASK-260729-2kaopg/protocol-spec/conformance/v1`
(the pinned root the accepted cycle-3 evidence names).

### Coverage of the owned surface

```text
cmdGlobalStatus      96.9%
globalStatusPlan    100.0%
globalStatusScope   100.0%
statusReport         86.7%
installedSkillDir    80.0%
checkFailed         100.0%
factsOf             100.0%
plannedRows         100.0%
markerMoved         100.0%
Describe            100.0%
classifySkillBuilds  81.8%
```

`projectStatusScope` reports 0.0% in this profile because the run is scoped to
`-run '^TestGlobalStatus'`, which exercises the machine-wide scope only; the
project scope is covered by the project-status tests in the full suite.

## 5. Host barrier and decisive full-suite gates

Barrier taken immediately before run 1:

- foreign Go/test processes: none (`ps` filtered on `go test`, `go build`,
  `compile -o`, `link -o`, `*.test` — empty)
- disk free: 8.3 GiB on `/System/Volumes/Data`
- load average: 16.20 (host is a shared workstation with 63 user sessions; this
  is reported as measured and no run below is presented as taken on an idle host)

Shared Go caches were not cleared. The 1.41 GiB of stale `TMPDIR/go-build*`
directories left by other processes were measured but deliberately not deleted.

<!-- RUNS-PLACEHOLDER -->
