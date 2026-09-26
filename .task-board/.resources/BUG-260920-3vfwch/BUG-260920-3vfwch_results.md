# BUG-260920-3vfwch results: hosted-macos git permission-denied flake

Role: developer. Worktree: STORY-260915-3w11un (uncommitted, no trunk integration
by this role). Product code untouched (test-only `*_test.go` diff); CHANGELOG not
required for test-only changes.

## 1. Cause classification per run (artifact evidence)

All three failures are the same errno and binary (`fork/exec
/opt/homebrew/bin/git: permission denied`, exactly one occurrence per
`go-test.json`), but NOT the same spawn shape. Concurrency below counts
tests actually executing (pause/cont aware), not merely started.

### Run 35340496757 (head 208953b, 2026-09-18) -- PRODUCT spawn

- Victim: `internal/install :: TestDryRunEffectBindingsSeeWhatARealOperationWrites`
  (SEQUENTIAL -- verified no `t.Parallel` at that head via `git show 208953b`).
- Failure line: `dryrun_conformance_test.go:857: the real global install failed:
  {... Errors:[fork/exec /opt/homebrew/bin/git: permission denied] ...}`.
- Timing: test body elapsed 4.81 s; failed 2026-09-18T11:48:14Z; package span
  11:40:37Z-11:49:25Z (528 s).
- Concurrency: 1 test executing in `internal/install` (the victim itself),
  38 paused; 3 packages overlapping (`install`, `swiftpmsource`, `transaction`).
- Shape: bare spawn error inside a product `Result` from `Global(Fetch:true)`
  over a manifest declaring a git-sourced skill. The chain is
  `Global -> closure -> gitops.Extract -> writeBlobs` (`git -C <repo> cat-file
  --batch`), whose `cmd.Start()` error returns raw with `cmd.Dir` UNSET. Every
  other product git spawn either wraps its error (`gitops.run`, `listTree`,
  envprofile) or swallows it (`gitignore.Missing`), so `writeBlobs` is the only
  bare-error git spawn on that path.

### Run 35481906193 (head fc19164, 2026-09-20) -- HELPER spawn

- Victim: `internal/install :: TestEndToEndInstall` (parallel).
- Failure line: `install_test.go:142: git [tag v1]: fork/exec
  /opt/homebrew/bin/git: permission denied` (fixture `e.skill -> e.git`).
- Timing: cont -> fail in 0.09 s at 2026-09-20T01:48:08Z.
- Concurrency: 3 executing in `internal/install`, 98 paused; 3 packages
  overlapping (`install`, `swiftpminterop`, `swiftpmsource`).
- Note: this run also failed `Test (windows-latest)`, but on platform-case
  gate pnpm assertions (`TestRealPinnedPNPM*`), unrelated to this flake and
  since fixed at HEAD by BUG-260916-2f3xbf. No `permission denied` there.

### Run 35510984798 (head 5f99188, 2026-09-20) -- HELPER spawn

- Victim: `internal/install :: TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState`
  (parallel).
- Failure line: `stage_test.go:723: git [add .]: fork/exec
  /opt/homebrew/bin/git: permission denied` (fixture `e.buildSkill -> e.git`).
- Timing: body elapsed 0.04 s at 2026-09-20T12:48:35Z; package span
  12:40:03Z-12:49:24Z (560 s).
- Concurrency: 3 executing in `internal/install`, 90 paused; 3 packages
  overlapping (`install`, `swiftpminterop`, `swiftpmsource`).

Correction to the brief: the "~100 t.Parallel tests in flight" were
run-event-started but PAUSED tests; the executing counts are 1 / 3 / 3.

## 2. Proven vs bounded

PROVEN (evidence-backed):

- P1. The error text cannot name its cause. Verified in Go source
  (`os/exec_posix.go`: parent `Stat(cmd.Dir)` failure -> `Op: "chdir"`;
  ANY `StartProcess` failure, including in-child chdir/dup/exec, ->
  `PathError{Op: "fork/exec", Path: <binary>}`). All three occurrences name
  the binary, so the parent `Stat(cmd.Dir)` succeeded in each; anything
  downstream (exec, in-child chdir, dup) remains possible from the text alone.
- P2. No concurrent PATH/env rewrite in the `internal/install` test binary.
  Audit: every `Setenv` sits in a sequential test with no parallel subtest
  (Go panics on `Setenv` under a parallel ancestor and denies a later
  `Parallel`; verified in `testing.go`); no `os.Setenv`/`Chdir`/`Umask`
  anywhere in the binary (`cmd/curator` `os.Setenv(PATH)` sites are a
  different process). The "concurrent test rewriting PATH" theory is
  eliminated for this binary.
- P3. No in-repo chmod path targets fixture-dir search bits or shared
  ancestors. Test `Chmod`s touch owned files/dirs only; `privatedir`,
  `buildcache/protection_unix`, `transaction` chmod owned paths only.
  Candidate (b) via an in-repo actor is eliminated.
- P4. Run 1 involved no chdir at all (`writeBlobs` leaves `cmd.Dir` unset),
  so candidate (b) is eliminated for that occurrence: it was a pure exec-time
  refusal at low concurrency (1 executing test).
- P5. No known-Go-flake prior art: web search for a darwin transient-EACCES
  `os/exec` issue returned no relevant tracker entry.

BOUNDED (not provable from these artifacts):

- B1. Candidates (a) transient exec refusal vs (c) image-level Homebrew git
  momentarily non-executable: UNRESOLVED between them. Load theory (a) is
  weakened (1/3/3 executing, 3 packages each); (c) is neither proven nor
  eliminated. The new in-process diagnostic discriminates at the next
  occurrence (binary lstat/stat + symlink eval, dir stat, cwd, PATH,
  NOFILE/NPROC rlimits).
- B2. External-actor variant of (b) (something outside the test binary
  flipping `cmd.Dir` search bits between parent `Stat` and child chdir):
  no evidence either way; the diagnostic's dir-stat row covers it.
- B3. Product-shaped occurrences (run 1 shape) are NOT mitigated by this
  change: R1 forbids product edits, and retrying whole installs would change
  test semantics. Only helper-shaped spawns are bounded. Recommended
  orchestrator follow-up: product-spawn observability for `gitops.writeBlobs`
  (separate task; out of scope here).

NOT DONE (correctly, per R2): no `/usr/bin/git` pin, no lane git binding, no
`.github/ci/test-gate.sh` change -- all gated on proving (c), which is unproven.

## 3. Change (test-only)

Shared fixture-git path with in-process diagnostic + bounded retry:

- `internal/install/gitfixture_test.go` (new): `gitFixture` /
  `gitFixtureWithBinary` (fatal layer, preserves the `git <args>: ...`
  message shape) over `runGitFixture` (assertable core: attempts, fatal
  text, log/sleep seams). Retry rule: exactly one retry, only on spawn
  EACCES/EAGAIN (never on `*exec.ExitError`, never on other spawn errnos),
  after 200 ms, with the retry + first-attempt diagnostic logged via `t.Log`.
  Spawn failures (all errnos, including unresolvable binary) carry the
  `gitfixture diagnostic` block: argv0 resolution, binary lstat/stat + symlink
  target, dir stat, cwd, PATH, rlimits.
- `gitfixture_rlimit_{linux,darwin,other}_test.go` (new): stdlib `syscall`
  exposes no `RLIMIT_NPROC`, and Darwin has none at all, so Linux reports
  NOFILE+NPROC via `x/sys/unix` (already a direct dependency), Darwin reports
  NOFILE + `kern.maxproc` sysctl, other GOOS report unavailable. No new
  dependency.
- Wired call sites (same argv/env/dir/output as before):
  `install_test.go` `env.git`, `draftsources_test.go` `testGit`,
  `drafttransport_test.go` `draftBareFixture.run` (keeps its pre-resolved
  absolute git via `gitFixtureWithBinary`), `atomicity/fixture_test.go`
  `env.git`. The `draftGitTool` `--exec-path`/`--version` probes keep their
  fail-fast shape (not fixture spawns; deliberately out of scope).
- The atomicity copy is byte-identical to the install copy modulo the package
  clause + header (Go test files cannot be imported across packages; R1 bars
  the only shared non-test seam, `internal/testcli`). Verified with diff in
  this session; keep the two in sync.

## 4. Tests and mutants (all killed)

Committed rows in `gitfixture_retry_test.go` (both packages, 9 rows each):

- Positive (driven through `runGitFixture`, the exact core the wired helpers
  call): injected spawn EACCES -> 2 attempts + backoff + retry/success log
  lines; injected EAGAIN -> same; REAL kernel EACCES via absolute
  non-executable binary + real runner (POSIX-only, `runtime.GOOS` skip with
  POSIX-only sibling precedent) -> 2 attempts, `permission denied`.
- Negative: real `git --bad-flag` exit via real runner -> 1 attempt, no
  sleep/log/diagnostic; wrapped real `*exec.ExitError` -> same;
  8-row `isRetryableSpawnError` predicate table (nil/exit/wrapped-exit/
  EACCES/EAGAIN/ENOENT/wrapped-EACCES/lookup-failure).
- Diagnostic (driven): persistent EACCES -> `gitfixture diagnostic` with
  resolution + binary + dir + cwd + PATH + rlimits + first-attempt text;
  ENOENT spawn -> fails fast WITH diagnostic; emptied-PATH lookup failure ->
  `unresolved:` diagnostic; outer `gitFixture` entry runs real `git --version`.

Mutants (install package, restored byte-identical after):

- A (retry removed, unsatisfiable `&&`): the 4 retry-path rows fail, negatives pass.
- B-inner (predicate widened to any error): predicate table + 2 fail-fast rows fail.
- B-outer (exit fast-path dropped): both NeverRetries rows fail (diagnostic shape).
- B-full (retry on any error): all 5 negative/predicate rows fail on attempts.

## 5. Validation (real exit codes, zsh/bash, `set -o pipefail` where piped)

- `go test ./internal/install/ ./internal/install/atomicity/ -run TestGitFixture
  -count=1`: ok both packages (9 rows each).
- Wired-path regressions: `TestEndToEndInstall`, `TestGlobalDryRunPlans...`,
  `TestDraftGitRuntimeMaterializesUnderSourceV1Key` (testGit),
  `TestDraftTransportAuth` + `TestAcquireDraftNetworkSelection`
  (draftBareFixture run) pass (36 s); atomicity
  `TestStableHybridActivationCommitsWithoutRestarting` passes (14 s).
- `go vet` on darwin + `GOOS=linux` + `GOOS=windows` for both packages: clean.
- `golangci-lint run ./internal/install/...` (v2.12.2, CI pin): 0 issues
  (8 self-introduced findings fixed in-session).
- `gofmt -l`: clean. `git status`: 4 modified `*_test.go`, 10 new `*_test.go`,
  zero product files.
- NOT run by this role (by rule): the full landing suite runs exactly once at
  handoff/CR publication. Linux/Windows lane behaviour is unchanged by
  construction (success path byte-identical; retry/diagnostic only fire on
  paths that previously failed the test outright).

## 6. Acceptance and observation window (for the orchestrator)

- AC disjunct 1 (10 consecutive green macos gates) is an OBSERVATION WINDOW
  this role cannot wait for: track the next 10 `Test (macos-latest)` runs
  after landing. If this flake recurs in a HELPER spawn, the go-test stream
  will now carry the `gitfixture diagnostic` + `retrying once` lines -- attach
  them to this bug instead of re-investigating blind.
- AC disjunct 2 (diagnostic proves image-level cause + stable-path binding) is
  armed but not triggered: cause (c) is unproven, so no binding was made.
- Known residual risk: a recurrence in a PRODUCT spawn (run 1 shape) will NOT
  be retried or diagnosed by this change (see B3); that outcome should open the
  recommended follow-up rather than re-open this fix.

---

# Revision 2 (rework-1: verdict F1, F2, N1–N3)

Role: developer, same Story worktree, revision-1 tree as the base (no
checkout/clean/stash). Rulings R1–R4 and bounds B1–B3 unchanged. Tree is now
4 modified + 8 new paths, all `*_test.go` (12 total: the 4 split
`gitfixture_rlimit_{linux,darwin}_test.go` files are deleted, 2 unified
`gitfixture_rlimit_unix_test.go` files added). Product code untouched;
CHANGELOG not required for test-only changes.

## F1 — wired helpers log the retry (fixed)

- `gitFixtureWithBinary` in both copies now builds
  `gitFixtureConfig{binary: binary, logf: t.Logf}`
  (`internal/install/gitfixture_test.go`,
  `internal/install/atomicity/gitfixture_test.go`).
- `gitFixture` / `gitFixtureWithBinary` take `testing.TB` instead of
  `*testing.T` (all wired callers pass `*testing.T`, which implements it;
  only `Helper`/`Fatalf`/`Logf` are used), so a committed row can pass a
  recording fake TB (`gitFixtureRecordingTB`: records `Logf`, delegates
  everything else to the real test).
- New row `TestGitFixtureOuterEntryLogsRetryOnTransientEACCES` (both
  packages, 11 `TestGitFixture*` rows each) drives the OUTER entry through
  a real transient kernel EACCES and asserts both lines. Deviation from the
  reviewer's probe shape, with cause: the probe's 0755-script + 0644-binary-
  interpreter + chmod-+x-at-50ms cannot run here — (a) a script interpreter
  fails on darwin with `exec format error` on retry (kernel rejects nested
  shebang; observed in-session), (b) this host hangs fresh tmp executables
  ~88 s then SIGKILLs them (`/bin/sh: line 1: 81585 Killed: 9 /tmp/shtest-g`,
  `real 1m28.359s`, exit 137; the same window killed two `go test` runs at
  ~75–83 s with no output, green on retry). The committed row keeps the
  probe's essence — real `fork/exec …: permission denied` on attempt 1,
  flip at 50 ms, real exec success on retry — via an absolute-path symlink
  that points at a 0644 blocker, then flips onto the real git (no freshly
  materialised executable is ever run; no PATH rewrite, so the row is
  `t.Parallel`; Windows skip follows the sibling-row precedent).
- Exact logged lines from the row (`go test -v`, install package):
  `gitfixture: git [--version] spawn failed (fork/exec …/001/git: permission denied); retrying once after 200ms`
  (+ the `gitfixture diagnostic` block) and
  `gitfixture: git [--version] succeeded on retry`.
  Row passes in 0.23–0.27 s per package; `-race -count=2` clean both packages.

## F2 — darwin rlimits render (fixed)

- Deleted `gitfixture_rlimit_darwin_test.go` (both packages); the linux file
  is replaced by `gitfixture_rlimit_unix_test.go` (`//go:build linux ||
  darwin`) using `unix.Getrlimit` for NOFILE and NPROC — no sysctl, no
  darwin special case. The `other` GOOS file is unchanged (reports
  unavailable). Correction to rev-1 results §3: Darwin does have
  RLIMIT_NPROC (`x/sys/unix` 0x7); `Getrlimit` returns the per-uid ceiling.
- `TestGitFixturePersistentSpawnFailureCarriesDiagnostic` pins the content on
  unix with `NOFILE cur=\d+ max=\d+; NPROC cur=\d+ max=\d+` (other GOOS keep
  the key-only check). Rendered on this darwin host:
  `rlimits: NOFILE cur=122880 max=9223372036854775807; NPROC cur=5568 max=8352`.

## N1–N3 (done)

- N1: predicate table gains `{"EPERM", spawnErr(syscall.EPERM), false}`.
- N2: backoff pinned against the literal `200*time.Millisecond`, not the
  constant.
- N3: row count corrected — rev 1 said 9, actually 10; now 11 per package
  with the outer-entry row.

## Mutants re-run (install package; atomicity copy is byte-identical modulo the package clause)

- A (retry branch unreachable): 5 FAIL — `RetriesOnceOnSpawnEACCES`,
  `RetriesOnceOnSpawnEAGAIN`, `PersistentSpawnFailureCarriesDiagnostic`,
  `RealSpawnEACCESRetriesOnce`, `OuterEntryLogsRetryOnTransientEACCES`;
  6 pass → killed.
- B-full (retry on any error: predicate `err != nil`, exit fast path
  dropped; kept compiling by retaining the `syscall` reference): 5 FAIL —
  `NeverRetriesNonZeroExit`, `NeverRetriesWrappedExitError`,
  `NonRetryableSpawnErrorFailsFastWithDiagnostic`, `RetryPredicate`,
  `UnresolvableGitFailsFastWithDiagnostic`; 6 pass → killed.
- C (backoff 5 ms): 2 FAIL — `RetriesOnceOnSpawnEACCES` (literal pin) and
  `OuterEntryLogsRetryOnTransientEACCES` (retry at 5 ms lands before the
  50 ms flip) → killed.
- E (predicate + EPERM): 1 FAIL — `RetryPredicate` → killed.
- Every mutant restored byte-identical (diff-verified); restored baseline
  11/11 green.

## Validation (real exit codes, bash, `set -o pipefail` where piped)

- `go test ./internal/install/ ./internal/install/atomicity/ -run
  TestGitFixture -count=1`: ok both (11 rows each).
- Wired-path regressions: `TestEndToEndInstall` +
  `TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState` ok (4.2 s);
  `TestDraftGitRuntimeMaterializesUnderSourceV1Key` + `TestDraftTransportAuth`
  + `TestAcquireDraftNetworkSelection` ok (15.9 s); atomicity
  `TestStableHybridActivationCommitsWithoutRestarting` ok (8.8 s).
- `go vet` darwin + `GOOS=linux` + `GOOS=windows` for both packages: clean.
- `golangci-lint run ./internal/install/...` (v2.12.2, CI pin): 0 issues.
- `gofmt -l`: clean. Both package copies diffed: only package clause (+ the
  standing header comment) differ.
- NOT run by this role (by rule): the full landing suite runs exactly once
  at handoff/CR publication. Linux/Windows lane behaviour unchanged by
  construction (success path byte-identical; retry/diagnostic only fire on
  paths that previously failed the test outright).

## Acceptance and observation window (for the orchestrator, unchanged)

- AC disjunct 1 remains an OBSERVATION WINDOW: track the next 10
  `Test (macos-latest)` runs after landing. A helper-shaped recurrence now
  leaves `retrying once after 200ms` / `succeeded on retry` lines in the
  go-test stream (visible in `go-test.json` output events even for passing
  tests), so "flake gone" vs "flake masked by a silent retry" is
  distinguishable — the F1 gap is closed.
- AC disjunct 2 armed but not triggered (cause (c) still unproven; no git
  binding, no `.github/ci/test-gate.sh` change).
- Residual B3 stands: a run-1-shaped PRODUCT spawn recurrence is neither
  retried nor diagnosed by this change; open the recommended follow-up leaf
  rather than re-opening this fix.

---

# Revision 3 (rework-2: verdict F3)

Role: developer, same Story worktree, revision-2 tree as the base (no
checkout/clean/stash). F1/F2/N1-N3 untouched (verified resolved — not
touched per the rework note). Rulings R1-R4 and bounds B1-B3 unchanged.
Diff rev2→rev3 touches exactly 2 paths:
`internal/install/gitfixture_retry_test.go` and
`internal/install/atomicity/gitfixture_retry_test.go` (the
`gitFixtureRecordingTB` + `TestGitFixtureOuterEntryLogsRetryOnTransientEACCES`
tail in each). Product code untouched; CHANGELOG not required for
test-only changes.

## F3 — outer-entry row is now event-driven, not timer-driven (fixed)

- Cause of the 4/80 stress failures: the row flipped the fixture symlink
  from a goroutine on a 50 ms timer and required attempt 1 to land before
  the flip. Under fork contention the first spawn lands after the flip,
  succeeds on the real git with no retry, and the row fails with
  `logs = []`.
- Fix (reviewer's `zz_fixshape_test.go` shape, adopted in both copies):
  `gitFixtureRecordingTB` gains a `flip func()` field; its `Logf` runs the
  flip when it sees the `retrying once` line, which `runGitFixture`
  emits synchronously after attempt 1 failed and before
  `sleep(gitFixtureBackoff)` (`gitfixture_test.go:127`). The goroutine
  and the 50 ms timer are gone; the flip's `os.Symlink`/`os.Rename`
  errors fail the test via `t.Fatal` (no `_ =`). `Logf` runs in the test
  goroutine (synchronous call chain from the row), so `t.Fatal` there is
  legal. Attempt 1 always sees the 0644 blocker (real kernel EACCES) and
  the retry always sees the real git — no timing anywhere.
- Both copies diffed: identical modulo the package clause (verified
  in-session; `diff` shows only `1c1`).

## Stress evidence (reviewer's harness, rerun)

- Ran the verdict's `zz_stress_test.go` verbatim (install copy as-is;
  atomicity copy with only the package clause adjusted) as a throwaway
  alongside the committed row — removed after the runs, final tree
  verified without it:
  `go test -run "TestZZStressSpawn|TestGitFixtureOuterEntryLogsRetryOnTransientEACCES" -count=10 -parallel 32`
- install plain: 3 rounds x 10 = 30/30 row iterations pass; install
  `-race`: 10/10 pass, no data-race report; atomicity plain: 10/10 pass.
  Total 50/50, 0 failures (rev2 baseline for the timer shape: 4/80
  plain). Each passing iteration takes ~0.3 s (two spawns + 200 ms
  backoff), i.e. the retry path is exercised every time, never skipped
  by an early flip.

## Mutants re-run (both packages, byte-restore + sha256 verified)

- C (backoff 5 ms): killed — `TestGitFixtureRetriesOnceOnSpawnEACCES`
  FAILs on the N2 literal pin in both packages. Note: the outer-entry
  row now PASSES under mutant C (previously it also failed) — expected,
  and itself evidence the timing window is gone: the flip keys off the
  logged line, not the clock.
- W (`logf: t.Logf` dropped from `gitFixtureWithBinary`): killed —
  `TestGitFixtureOuterEntryLogsRetryOnTransientEACCES` FAILs in both
  packages. Failure mode changed honestly with the design: without the
  log line the flip never fires, so the retry hits the blocker and the
  outer entry fails fatal with persistent `permission denied`
  (`(first attempt: ...)`), instead of `logs = []`.
- Restored byte-identical (sha256 `c9133e62…` install,
  `8b83e41f…` atomicity for `gitfixture_test.go`); restored baseline
  11/11 green in both packages.

## Validation (real exit codes, bash, `set -o pipefail` where piped)

- `go test ./internal/install/ ./internal/install/atomicity/ -run
  TestGitFixture -count=1`: ok both (11 rows each).
- `go vet` darwin + `GOOS=linux` + `GOOS=windows` for both packages: clean.
- `golangci-lint run ./internal/install/...`: 0 issues.
- `gofmt -l`: clean. No `time.Sleep(50` / goroutine / `_ = os.` remains
  in the F3 row (grep-verified). Final `git status`: 4 modified + 8 new
  paths, all `*_test.go`, zero product files.
- NOT run by this role (by rule): the full landing suite runs exactly
  once at handoff/CR publication. Linux/Windows lane behaviour unchanged
  by construction (this revision only reshapes one test row's trigger;
  the retry policy and diagnostic are untouched).

## Corrections and carry-overs (for the orchestrator / reviewer)

- Rev2 results §F1 said "a script interpreter fails on darwin with
  `exec format error` on retry (kernel rejects nested shebang)": per the
  rev2 verdict, that holds only for a *script* interpreter; a binary
  interpreter works. The committed symlink shape stands; only the
  generalisation was loose. (Rev2 text above left intact; this note is
  the correction.)
- AC disjunct 1 remains an OBSERVATION WINDOW: track the next 10
  `Test (macos-latest)` runs after landing. The new row can no longer
  false-red under spawn pressure, so a lane-level red in the window
  reads as signal, not test noise.
- AC disjunct 2 armed but not triggered (cause (c) still unproven).
- Residual B3 stands (product-shaped run-1 occurrence neither retried
  nor diagnosed; follow-up leaf, not this fix).
