# TASK-260720-jrrgw9 — shared compiled fixture rework results

Date: 2026-07-29
Role: developer (producer)
Scope: one file, `cmd/curator/status_test.go`
Verdict: both allowed narrow commands passed; ready for review

Implements the "Smallest robust patch" of
`TASK-260720-jrrgw9_second-timing-diagnosis-results.md` in
`/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260720-jrrgw9/worktree`.

## 1. Before / after structure

Exactly one file changed. Proof that nothing else was touched:

```text
$ find . -path ./.git -prune -o -type f -mmin -90 -print
./cmd/curator/status_test.go
```

| | before | after |
| --- | ---: | ---: |
| `cmd/curator/status_test.go` lines | 1484 | 1555 |
| sha256 | `4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe` | `487b12bdf531e4714983eab83b804de7b4604513e435256e550f60391ee0d32e` |
| top-level `func Test...` in the file | 14 | 10 |
| real installed compiled fixtures built by the five surfaces | 7 | 1 |
| `t.Cleanup(func() { reinstall(t) })` in the cache-movement cases | 4 | 0 |

Five top-level tests became fixture-accepting helpers under one sequential
parent. No test was deleted, renamed away, or made parallel.

| removed top-level test | new helper | subtest under the parent |
| --- | --- | --- |
| `TestStatusReportsCompiledCurrentnessAndFailsCheck` | `assertCompiledCurrentnessAndFailedCheck` | `status reports compiled currentness and fails check` |
| `TestInstallAndUpgradeRepairCorruptCompiledState` | `assertCorruptCompiledStateIsRepaired` | `install and upgrade repair corrupt compiled state` |
| `TestInstallAndUpgradeRestoreTheCacheWhenTheCommitFails` | `assertTheCacheIsRestoredWhenTheCommitFails` | `install and upgrade restore the cache when the commit fails` |
| `TestStatusReportsProtectedCacheStateThatMovedDuringTheCheck` | `assertProtectedCacheStateThatMovedDuringTheCheck` | `status reports protected cache state that moved during the check` |
| `TestInstallRepairsUntrustedCompiledStateAndPreservesTheOldInstall` | `assertUntrustedCompiledStateIsRepaired` | `install repairs untrusted compiled state and preserves the old install` |

Added: `type compiledProjectFixture {project, home, installed}`,
`newInstalledCompiledProject(t)`, and
`TestCompiledProjectStatusRepairRollbackRecovery`.

Unchanged top-level tests in the file (still independent, still own their
fixtures): `TestMain`, `TestStatusReportsATransitivelyResolvedCompiledCommand`,
`TestStatusReportsAnUnusableToolchainPerCompiledCommand`,
`TestStatusKeepsLegacyBehaviourWhenAPlanFailsWithoutCompiledCommands`,
`TestGCRetainsAndReportsReferencedCompiledState`,
`TestDryRunNeverClaimsACompletedCompilerCheck`,
`TestStatusExplainsAnUnusableGoToolchain`,
`TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands`,
`TestStatusAcceptsAnUnchangedLegacyMarkerSchema`.

Nothing outside this file references the five removed test names (checked
repo-wide, exit 1 = no hits), so no Makefile, CI, doc, or board allowlist
selector broke. The second producer allowlist command still selects five
existing top-level tests.

## 2. Exact diff

`cmd/curator/status_test.go` is **untracked** in this worktree, so
`git diff` has no baseline for it. The pre-patch file was reconstructed by
reversing the ten applied edit pairs and checked against the recorded
pre-patch digest before the diff was rendered:

```text
$ python3 reverse_patch.py status_test.after.go status_test.before.go
reconstructed=4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe
expected     =4c2339dd36854cd51baf28c92c480b4341fa398c10118f41d44b8d29e845afbe
OK: reconstruction is byte-identical to the pre-patch file
```

```text
$ git diff --no-index --stat status_test.before.go status_test.after.go
 1 file changed, 124 insertions(+), 53 deletions(-)
```

The rendered patch is attached as
`TASK-260720-jrrgw9_shared-fixture-rework.diff`, and the reconstruction script
as `TASK-260720-jrrgw9_shared-fixture-reverse-patch.py`.

## 3. Assertion matrix

Every executable assertion listed in the diagnosis's preservation matrix is
still present and still ran. Only fixture construction was removed.

| surface | preserved executable assertions (all ran, all passed) | what was removed |
| --- | --- | --- |
| Compiled currentness | full identity row (skill/command/driver/build root/source dir/cache key/artifact path/target/build source), private-path exclusion for `home` and `.agents/bin`, clean `--check`, all 14 drift cases with state + cause + cache outcome + non-empty detail + skill demotion, production `Describe()` rendering per case, three `humanCLI` plain-CLI paths | its own `compiledProject` + install |
| Corrupt-cache repair | both `install` and `upgrade`; receipt-bytes and artifact-bytes corruption; pre-repair `corrupt-build-cache` status; unusable-toolchain refusal with byte-identical cache fingerprint, identical marker build record, surviving `SKILL.md`; real repair; "rebuilt corrupt build cache state into a new protected entry" notice; post-repair `--check`; stable logical key; exactly 1 live protected entry | two per-command `compiledProject` + install |
| Commit-failure rollback / recovery | both `install` and `upgrade`; real publication before the failure; withdrawn-entry count; staged-directory count 0; restored live artifact bytes; 1 live entry; unchanged marker build record; unchanged `installedFingerprint` across project `.agents`, both adapter mirrors, `home/runtime`, hybrid store, consumer ledger; still-`corrupt-build-cache` status; successful ordinary recovery; post-recovery `--check` | two per-command `compiledProject` + install |
| Cache moved during the check | all four movement cases (corrupted, removed, unprovable boundary, self-consistent replacement); fresh production dry-run plan per case; fresh marker fingerprint per case; marker bytes proven byte-identical before and after the move; `build-state-changed` row; `"build cache"` in the detail; skill demotion; `checkFailed` fail-closed | its own fixture, and the four post-assertion `reinstall` reconciliations |
| Untrusted cache | initial marker with compiled state; compiler-free dry-run vocabulary (`outcome=would-rebuild-untrusted-cache`, no "rebuilt untrusted" claim); real untrusted rebuild with its notice; post-repair `--check`; stable logical key; later toolchain failure preserving the marker build record and `SKILL.md` | its own fixture |

Two assertions were adapted, not weakened, because live state is now shared:

1. **Per-case baseline re-read.** `before := marker.Read(installed)` moved into
   each repair/rollback subtest (and into each corruption case), with the same
   `before == nil || len(before.Builds) != 1` guard. Previously a single
   command-level read; now every case compares against the state it actually
   started from, which is what the diagnosis required ("do not share a stale
   baseline across cases").
2. **Withdrawn-entry count is a delta.** `quarantinedEntries(t, home) != 1`
   became `quarantinedEntries(t, home) - withdrawnBefore != 1`, with
   `withdrawnBefore` sampled immediately before the failing run. The claim is
   unchanged — "exactly the one this run published" — and is now stated against
   the shared cache root rather than an assumed-empty one. See §7.

One guard was added: the rollback case registers a cleanup that reinstalls
**only if the case skipped**. `denyWrites` skips on a privileged process that
can write through a `0o500` directory, which would otherwise hand a corrupted
cache to the following cases without failing anything. It is a no-op on every
executed path (the run below recorded zero skips).

## 4. Real exits and timings

No Go commands overlapped. Both are verbatim from the diagnosis producer
allowlist, `-count=1`, no `-timeout` token. Output was redirected to a file, not
piped, so `$?` is the real process status.

**Command 1 — exit 0, wall 216s, package `ok ... 215.657s`**

```text
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 -v ./cmd/curator -run '^TestCompiledProjectStatusRepairRollbackRecovery$'
```

```text
--- PASS: TestCompiledProjectStatusRepairRollbackRecovery (215.25s)
    --- PASS: .../status_reports_compiled_currentness_and_fails_check (49.50s)          [15 cases]
    --- PASS: .../install_and_upgrade_repair_corrupt_compiled_state (66.68s)            [2 commands x 2 corruptions]
    --- PASS: .../install_and_upgrade_restore_the_cache_when_the_commit_fails (57.24s)  [2 commands]
    --- PASS: .../status_reports_protected_cache_state_that_moved_during_the_check (12.89s) [4 cases]
    --- PASS: .../install_repairs_untrusted_compiled_state_and_preserves_the_old_install (16.21s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	215.657s
```

34 leaf cases, 0 failures, 0 skips.

**Command 2 — exit 0, wall 44s, package `ok ... 43.732s`**

```text
CURATOR_CONFORMANCE_ROOT=/Users/iv/Developer/ReluxWorks/curator/.temp/TASK-260729-3nx97g/worktree/conformance/v1 go test -count=1 -v ./cmd/curator -run '^(TestAuthoritativeBootstrapCasesAreExecutable|TestAuthoritativeUpgradeCasesAreExecutable|TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck|TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands|TestStatusAcceptsAnUnchangedLegacyMarkerSchema)$'
```

```text
--- PASS: TestGlobalStatusReportsCompiledCurrentnessAndFailsCheck (28.32s)   [7 cases]
--- PASS: TestAuthoritativeBootstrapCasesAreExecutable (0.00s)               [3 authoritative cases]
--- PASS: TestAuthoritativeUpgradeCasesAreExecutable (12.16s)                [3 authoritative cases]
--- PASS: TestStatusJSONKeepsTheLegacyShapeWithoutCompiledCommands (1.46s)
--- PASS: TestStatusAcceptsAnUnchangedLegacyMarkerSchema (1.44s)
PASS
ok  	github.com/relux-works/curator/cmd/curator	43.732s
```

0 failures, 0 skips. The authoritative rc.5 vectors were consumed from the
authoritative root `.../TASK-260729-3nx97g/worktree/conformance/v1`
(`manifest.json` sha256 `b6f56aacc0e37dcc6692f73f641bff761e89b645adfe20a47a06d81c6fda204c`);
this worktree publishes no private copy of it.

**Formatting and whitespace gates**

```text
$ gofmt -l cmd/curator/status_test.go            -> no output, exit 0
$ git diff --check -- cmd/curator/status_test.go -> exit 0   (VACUOUS: file is untracked, no baseline)
$ git diff --check --no-index -- /dev/null cmd/curator/status_test.go
                                                 -> no output, exit 1
```

Exit-code calibration for the non-vacuous form, measured on this host: a file
with trailing whitespace exits **3**, a clean file exits **1** (the
"files differ" code). Exit 1 with no output therefore means no whitespace
errors. The literal `--check` form is reported honestly as vacuous rather than
as a passing gate.

## 5. Process, disk, and cleanup evidence

- Before each Go command: `ps -Ao pid,command | grep -Ei '(^|/)(go|.*\.test)( |$)|go-build|cmd/curator' | grep -v grep` returned nothing (exit 1).
- The `pgrep -af '...'` form from the diagnosis returns transient PIDs that resolve to the probe's **own** shell, whose command line contains the pattern text; `ps -p <pid>` on them exits 1 (already terminal) or shows the probe itself. Both commands were started only after the broad `ps` scan came back empty.
- After the runs: same broad scan empty (exit 1). No orphaned `go`, `go build`, or `curator.test` process.
- `t.TempDir` cleanup: `ls -d $TMPDIR/TestCompiledProject*` and `/var/folders/*/*/T/TestCompiledProject*` both match nothing — every fixture tree was removed.
- Disk: `/System/Volumes/Data` 26Gi free before and after; no measurable growth from the runs.
- Worktree: only `cmd/curator/status_test.go` has an mtime inside the session window. No stage, commit, publication, cache clearing, host installation, or pin change was performed.

## 6. Expected full-suite saving

Measured, per surface. "Before" is the diagnosis's ranked inventory; "after" is
this run.

| surface | before | after | delta |
| --- | ---: | ---: | ---: |
| Compiled currentness | 69.816s | 49.50s | −20.3s |
| Corrupt-cache repair | 85.75s | 66.68s | −19.1s |
| Commit-failure rollback | 76.75s | 57.24s | −19.5s |
| Cache moved during check | 59.68s | 12.89s | −46.8s |
| Untrusted cache | not isolated | 16.21s | — |
| shared fixture install (new, paid once) | — | 12.73s | +12.7s |

The shared install cost is derived as `215.25s` parent minus `202.52s` of
subtest time = **12.73s**, which is also the per-fixture setup cost the six
removed fixtures each used to pay.

- Four measured surfaces: 291.996s before → 186.31s + 12.73s shared setup = **199.04s**, saving **92.96s**.
- Untrusted cache: its own separate fixture (~12.7s) is gone; its assertion body still costs 16.21s.
- **Total measured saving ≈ 105s**, against the diagnosis's predicted 108–112s.

Caveats stated plainly:

- The 69.816s currentness baseline was measured in an earlier focused run under
  different concurrent load, so the per-surface deltas mix measurements taken
  at different times. The 105s figure is an estimate built on those numbers,
  not a single controlled A/B.
- The verifier-2 full gate **timed out** at `cmd/curator 600.591s` on the
  unchanged 10-minute package alarm. It never completed, so the true uncached
  full-package duration was never established and **no claim is made here that
  the full suite now fits**. Removing ~105s from a run that was killed at 600s
  reduces pressure but does not prove a bound. An independent verifier with an
  amended allowlist still has to establish the exact uncached full-suite result
  before race, coverage, or native Windows gates are attempted.

## 7. ROUTED, not fixed: withdrawn cache entries are never collected

Discovered while making the rollback assertion correct under a shared cache
root. Not fixed here — this task is test-focused and the defect belongs to the
owning build-cache implementation task.

`internal/buildcache/publish.go` documents, on `Revert`:

> No byte is deleted here: the withdrawn entry is quarantined exactly like any
> other unusable one and is collected by the ordinary sweep.

The ordinary sweep does not collect it. `internal/buildcache/collect.go:133`
classifies cache-root members three ways:

- `entryNameRE` = `^[0-9a-f]{64}$` — live keys, swept when unreferenced;
- `sweepPrefix` = `".sweep-"` — retired wreckage, deletion finished;
- `default` — warned about as an *unrecognized cache root member* and retained.

`quarantinePrefix` is `".quarantine-"` (`publish.go:156`), which matches
neither of the first two. Every quarantined entry therefore accumulates in the
protected cache root forever and produces a
`retained unrecognized cache root member` warning on every `gc`.

Impact observed: the corrupt-repair and untrusted-repair paths each withdraw an
entry, so by the time the rollback case runs on a shared fixture the root
already holds withdrawn entries. That is why its count assertion is now a
delta. No test asserts the sweep's treatment of `.quarantine-*` today, so the
gap is currently unguarded.

Recommendation for the owning task: either teach `Sweep` to collect
`quarantinePrefix` members under the same grace rule, or correct the `Revert`
doc comment and give operators a documented way to reclaim them. A regression
test should pin whichever is chosen.

## 8. Not run

Per the task instructions, deliberately not executed and not claimed:

- broad `go test ./cmd/curator` without `-run`
- `go test ./...`
- `go test -race ./...`
- coverage measurement
- native Windows execution
- `go vet` / `go build` (outside the two-command allowlist; the package did
  compile, as both commands built and ran the test binary)

No `-timeout` flag was passed to either command; the package default alarm was
left unchanged.
