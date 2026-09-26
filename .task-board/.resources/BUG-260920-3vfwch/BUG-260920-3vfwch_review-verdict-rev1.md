# BUG-260920-3vfwch — review verdict, revision 1 (CR-BUG-260920-3vfwch-1)

Reviewer: claude-opus-5 (RUN-260920-fa9cc7), independent exact-head review, 2026-09-20.
Shell: zsh on the host (darwin/amd64, go1.26.0), `set -o pipefail` where piped; all
throwaway tests and mutants ran in a disposable clone, never in the Story worktree.

## Verdict: CHANGES REQUESTED (→ `to-dev`)

The retry policy, the negative rows and the diagnostic block are as ruled and the mutants
A / B-full die, but two ruled properties are not delivered on the wired path:

- **F1 (must fix) — the retry is silent in the wired helpers.** `gitFixtureWithBinary`
  builds `gitFixtureConfig{binary: binary}` without `logf`
  (`internal/install/gitfixture_test.go:83`, `internal/install/atomicity/gitfixture_test.go:87`),
  so `runGitFixture` replaces the seam with a no-op. Every wired call site (`env.git`,
  `testGit`, `draftBareFixture.run`, atomicity `env.git`) therefore retries with **nothing
  in the test log**. Brief R2(2) requires "the retry recorded in the test log"; the review
  note pins "200 ms backoff, logged"; results.md §3/§6 claim `t.Log` lines ("the go-test
  stream will now carry the `gitfixture diagnostic` + `retrying once` lines") — that claim
  is false for the wired path. Consequence: the AC observation window (10 consecutive green
  macOS gates) cannot distinguish "flake gone" from "flake masked by a silent retry", so the
  evidence the orchestrator is asked to track does not exist. No committed row pins the
  wiring: `TestGitFixtureOuterEntryRunsRealGit` (`gitfixture_retry_test.go:257`) drives the
  outer entry on the success path only; the logging rows all inject `logf` themselves.
  Reproduction (driven through the wired `env.git`, probe source attached):
  `TestReviewProbeWiredEnvGitRetrySucceeds` — PATH-resolvable `git` script whose shebang
  interpreter is mode 0644 (kernel EACCES at exec, LookPath accepts the 0755 script — the
  gate's exact `fork/exec …/git: permission denied` shape), a goroutine chmods the interpreter
  +x after 50 ms; `e.git(e.project, "--version")` PASSES after 862 ms (≥200 ms proves the
  backoff+retry ran) and the `-v` stream contains **0** `retrying once` / `succeeded on retry`
  lines (`grep -c` = 0 on the attached `probe-stream.log`).
  Fix shape: `gitFixtureConfig{binary: binary, logf: t.Logf}` in both copies, plus a row that
  drives the OUTER entry through a retry and asserts the log line (either accept
  `testing.TB` and pass a recording fake TB, or re-exec `os.Args[0] -test.run=^X$ -test.v`
  like `commit_test.go:736` and grep the child output). `go test -json` captures `t.Log` of
  passing tests (the green run's `go-test.json` carries output events from passing tests), so
  once wired the retry will be visible in `test-evidence-macos-latest`.

- **F2 (must fix) — the darwin rlimits line is wrong on the only lane that flakes.**
  `internal/install/gitfixture_rlimit_darwin_test.go:18-19` (and the atomicity copy) says
  "Darwin has no RLIMIT_NPROC" and renders `unix.Sysctl("kern.maxproc")`, which returns the raw
  int32 bytes as a string. Driven output (attached `probe-stream.log`, `od -c`):
  `rlimits: NOFILE cur=122880 max=9223372036854775807; NPROC n/a on darwin (kern.maxproc=\xa0 \x00)`
  — unreadable bytes in the go-test stream (test2json turns them into U+FFFD, the value is
  unrecoverable; `grep` even classifies the mutant logs as binary). Both premises are wrong:
  `golang.org/x/sys@v0.36.0/unix/zerrors_darwin_{arm64,amd64}.go:1188` defines
  `RLIMIT_NPROC = 0x7`, and on this host `unix.Getrlimit(unix.RLIMIT_NPROC)` returns
  `cur=5568 max=8352` (= `kern.maxprocperuid` / `kern.maxproc`; `ulimit -u` = 5568) — the
  per-uid ceiling XNU's `fork1()` enforces with EAGAIN, i.e. exactly the fact brief R2(1)
  asked for (`Getrlimit(RLIMIT_NPROC/NOFILE)`). `unix.SysctlUint32("kern.maxproc")` = 8352 if
  the sysctl is kept. The committed row `TestGitFixturePersistentSpawnFailureCarriesDiagnostic`
  only checks the key `"rlimits:"` (`gitfixture_retry_test.go:179`), so garbage passed the
  green macOS gate. Fix shape: one `//go:build linux || darwin` file using `Getrlimit` for
  NOFILE and NPROC (drop the darwin special case or fix the sysctl call), and pin the content
  on unix with a regexp such as `NOFILE cur=\d+ max=\d+; NPROC cur=\d+ max=\d+` so a garbage
  or "n/a" rendering fails the row.

Non-blocking (fix if cheap, otherwise state as bounds in results.md):

- N1 `isRetryableSpawnError` widened to EPERM survives every row (mutant E below): the
  predicate table (`gitfixture_retry_test.go:150`) has no EPERM row. An
  `errors.Is(err, fs.ErrPermission)` implementation would retry EPERM too; add
  `{"EPERM", spawnErr(syscall.EPERM), false}`.
- N2 The 200 ms value is not pinned (mutant C: `gitFixtureBackoff = 5ms` survives; the row at
  `gitfixture_retry_test.go:60` compares against the constant itself). A one-line pin is enough.
- N3 results.md §4 says "9 rows each"; there are 10 `TestGitFixture*` rows per package.

## What I verified (review-note items 1–6)

1. **R1 / test-only.** `git diff --name-only c3f9eeea..20fb22ef` = 14 paths, all `*_test.go`;
   `.github/`, `cmd/`, `internal/**` non-test untouched. The attached patch is byte-identical to
   `git diff --binary base..candidate` (sha256 `72c1581f…` matches the CR record).
   Tree identity: temp-index `write-tree` of the Story worktree = `20fb22ef…` = candidate =
   `d57ca591^{tree}` (gate commit of run 35524579878, parent c3f9eeea); the disposable clone
   (`git clone --no-checkout` control root → `checkout --detach c3f9eeea` → `apply --index`
   patch → commit) has `HEAD^{tree}` = `20fb22ef…` and ended the review at that tree
   (`git status` clean after every mutant restore).
2. **Retry rule.** Code: exactly one retry (`runGitFixture`, no loop), only when
   `errors.Is(err, EACCES) || errors.Is(err, EAGAIN)` and not `*exec.ExitError`
   (`isRetryableSpawnError`, `gitfixture_test.go:175-180`); ENOENT / lookup failure / exit
   fail fast; backoff 200 ms via the `sleep` seam. Reran the 10 rows in both packages in the
   clone: `go test ./internal/install/ ./internal/install/atomicity/ -run TestGitFixture
   -count=1 -v` → `ok` both (1.198 s / 0.684 s), including the real-kernel EACCES row
   (`TestGitFixtureRealSpawnEACCESRetriesOnce`, absolute 0644 binary) and the predicate table.
   Mutants (my own, precompiled `go test -c` + `codesign -s -` per mutant because the host
   watchdog SIGKILLs fresh unsigned test binaries at ~75 s; byte-copy restore + sha check
   between mutants; logs attached):
   - baseline: 10/10 PASS
   - **A** retry branch unreachable (`|| true` on the non-retryable test): 4 FAIL
     (RetriesOnceOnSpawnEACCES, RetriesOnceOnSpawnEAGAIN, PersistentSpawnFailureCarriesDiagnostic,
     RealSpawnEACCESRetriesOnce), negatives PASS → killed
   - **B-full** retry on any error (predicate `err != nil`, exit fast path dropped): 5 FAIL
     (NeverRetriesNonZeroExit, NeverRetriesWrappedExitError,
     NonRetryableSpawnErrorFailsFastWithDiagnostic, RetryPredicate,
     UnresolvableGitFailsFastWithDiagnostic) → killed
   - **H** diagnostic drops the `binary:` line: PersistentSpawnFailureCarriesDiagnostic FAIL → killed
   - **L** `logf` seam disconnected inside `runGitFixture`: RetriesOnceOnSpawnEACCES +
     RealSpawnEACCESRetriesOnce FAIL → killed (but see F1: the wired entry never connects it)
   - **E** predicate + EPERM: 10/10 PASS → survives (N1)
   - **C** backoff 5 ms: 10/10 PASS → survives (N2)
3. **Diagnostic in the go-test stream.** Driven once through the wired `env.git` with a
   persistent kernel EACCES (`TestReviewProbeWiredEnvGitPersistentEACCES`): the failure line
   `zz_review_probe_test.go:43: git [--version]: fork/exec …/bin/git: permission denied (first
   attempt: …)` is followed by the block — `argv0 "git" resolved: …`, `binary: lstat: mode=-rwxr-xr-x;
   stat: mode=-rwxr-xr-x size=114`, `dir "<cmd.Dir>": lstat: mode=drwxr-xr-x; stat: …`, `cwd:`,
   `PATH:`, `rlimits:` — all present (F2 covers the rlimits content). Symlink target/eval and
   dir-stat rows give the discriminators the brief asked for ((b) would show `lstat: permission
   denied` on the dir; (c) shows the binary mode/target at failure time).
4. **Cause classification.** Re-derived from the three `test-evidence-macos-latest`
   artifacts (`gh run download`, my own `analyze.py` over `test/go-test.json`, output attached):
   - 35340496757 (head 208953b): `internal/install :: TestDryRunEffectBindingsSeeWhatARealOperationWrites`,
     lifecycle run→fail with no pause (sequential; `git show 208953b:internal/install/dryrun_conformance_test.go`
     lines 832-876 has no `t.Parallel`), elapsed 4.81 s, fail 11:48:14.468Z, package span
     11:40:37→11:49:25 (528 s), failure text is a product `Result` with
     `Errors:[fork/exec /opt/homebrew/bin/git: permission denied]` from `Global(… Fetch:true)`
     at `dryrun_conformance_test.go:857`; 38 paused parallel tests, packages in flight
     install/swiftpmsource/transaction. Source at 208953b: `internal/gitops/gitops.go:386-403`
     `writeBlobs` = `exec.Command("git", "-C", repo, "cat-file", "--batch")`, no `cmd.Dir`,
     `if err := cmd.Start(); err != nil { return err }` (bare); `run()` (:37-54) and `listTree`
     (:238-244) wrap with `git … failed:`; `envprofile/gitsource.go:197,342` wrap with
     `DiagSourceInvalid`; `envprofile/import.go:457 gitOutput` is bare but not on the install
     path. Run-1 claim holds.
   - 35481906193 (head fc19164): `TestEndToEndInstall`, cont 01:48:08.068Z → fail 01:48:08.158Z
     (0.09 s), `install_test.go:142: git [tag v1]: fork/exec /opt/homebrew/bin/git: permission denied`
     (helper `e.git`); 2 other tests executing + victim, 98 paused; install/swiftpminterop/swiftpmsource.
   - 35510984798 (head 5f99188): `TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState`,
     cont 12:48:35.857Z → fail 12:48:35.902Z (0.04 s), `stage_test.go:723: git [add .]: …`
     (helper `e.buildSkill → e.git`); 3 executing + victim, 89 paused; same three packages.
   - The producer's counts (1/3/3 executing, 38/98/90 paused) match mine within the ±1
     victim-counting convention; the "~100 in flight" in the brief were paused tests.
   - P2 by my grep: no `os.Setenv`/`os.Chdir`/`Umask` anywhere in `internal/` or `cmd/`
     non-test code; the install binary's 22 `t.Setenv` sites are all top-level sequential by
     Go's own rule (panics under a parallel ancestor), and top-level parallel tests stay paused
     until the sequential pass ends, so no PATH rewrite can overlap a parallel spawn. P3: the 3
     test `Chmod`s (`install_test.go:277`, `draftbuild_test.go:494`, `generation_test.go:150`)
     and the product chmods spot-checked (`privatedir_unix.go:30`, `buildcache/protection_unix.go:329-356`,
     `transaction/files.go`, `staging.go`) set 0700 or copy file modes on owned paths; none
     drops owner search permission on a shared ancestor. P4 holds (run 1 had no `cmd.Dir`).
     P5: my own web search found no darwin transient-EACCES `os/exec` tracker entry either
     (golang/go#66654, #46819, #22315, #22059 and the golang-nuts M1 "operation not permitted"
     thread are all different shapes). Also noted from the same artifacts: on darwin Go spawns
     with fork+execve (`syscall/exec_libc2.go`), not `posix_spawn`, and `exec.LookPath` skips
     non-executable candidates (`findExecutable` → `eaccess(X_OK)`), so the error naming
     `/opt/homebrew/bin/git` proves `eaccess` passed microseconds before `execve` returned
     EACCES — which weakens (c) "mode wrong" further and leaves the kernel/MAC-layer transient
     as the leading (a) variant. Still bounded, as results.md says.
5. **Call sites.** Read all four diffs: argv (`gitArgs` incl. the `-c commit.gpgsign=false -c
   tag.gpgSign=false` prefix), env (`append(os.Environ(), extraEnv...)`), dir (`cmd.Dir = dir`
   iff non-empty — identical effect to the old unconditional assignment), `CombinedOutput`, and
   the `git <args>: <err>\n<out>` failure text are unchanged on the success path;
   `draftBareFixture.run` keeps its pre-resolved absolute `git` via `gitFixtureWithBinary` and
   `t.Helper()`; `draftGitTool`'s `--exec-path`/`--version` probes (`drafttransport_test.go:391,395`)
   and the non-git spawns (`consumer-tool`, helper-process re-exec) are deliberately untouched.
   The two package copies differ only in the package clause and a 4-line header (`diff`).
   The hosted gate (run 35524579878, all 11 lanes success, Test (macos-latest) 17:02:25→17:13:15)
   ran on this exact tree; its macOS `go-test.json` has 0 fails and no `gitfixture` lines.
   Lint/vet/wired-path regressions: accepted from that green gate, not rerun locally.
6. **B3 honesty.** Stated plainly in results.md (§2 B3, §6): a run-1-shaped occurrence
   (product spawn in `gitops.writeBlobs` from a sequential test) is neither retried nor
   diagnosed by this change. The AC as written does not require product-side action and R1
   forbids it here, so this is a residual for a separate follow-up leaf (wrap the
   `writeBlobs` `Start()` error with the op and dir facts at minimum), not a finding against
   this leaf. Note for the orchestrator: 1 of the 3 historical hits had that shape, so the
   observation window can still fail from it.

## Residuals (not blocking)

- Fixture git spawns outside the two wired packages keep the old bare shape
  (`cmd/curator/{draft_diagnostics,draft_sources,draft_transport,envconfig,main,profile}_test.go`,
  `internal/{closure,envprofile,gitignore,gitops,interop/environments,rustsource,snapshot}`
  tests) and the `draftGitTool` probes in `internal/install`; all three historical hits were in
  `internal/install`, so this is a bound, not a defect.
- The helper duplication (install ↔ atomicity) is the accepted cost of the test-only ruling;
  keep the two in sync when fixing F1/F2.

## What I reran vs accepted

Reran myself: patch/tree identity (3-way), the 10 rows × 2 packages, the two wired-path probes,
7 mutants, the three artifact analyses, the 208953b source checks, the Setenv/Chdir/Umask/Chmod
greps, the x/sys constant and sysctl/rlimit probes. Accepted from attached evidence: the hosted
gate run 35524579878 on tree `20fb22ef…` (all lanes green) as the full-suite, lint and
cross-platform arbiter.

## Route

`set_status(BUG-260920-3vfwch, status=to-dev)` with this file as verdict evidence. Revision 2
needs: F1 + a driven outer-entry retry-log row, F2 + a content-pinning rlimits row (both
package copies), optionally N1–N3, results.md updated (the "logged via t.Log" and "Darwin has no
RLIMIT_NPROC" statements corrected, row count 10), a fresh green gate on the new tree.

Attached alongside: `BUG-260920-3vfwch_review-rev1-evidence.tar.gz` (probe source, probe
stream, rows log, mutant logs, artifact analysis, mutant driver).
