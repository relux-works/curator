# BUG-260920-3vfwch — review verdict, revision 2 (CR-BUG-260920-3vfwch-2)

Reviewer: claude-opus-5 (RUN-260920-f8d90f), independent exact-head review, 2026-09-20/21.
Shell: zsh on the host (darwin/amd64, go1.26.0), `set -o pipefail` where piped; every throwaway
test, probe and mutant ran in a disposable clone (`git clone --no-checkout` control root →
`checkout --detach c3f9eeea` → `apply --index` rev2 patch → commit; `HEAD^{tree}` =
`21bde827…`), never in the Story worktree. Evidence tarball attached:
`BUG-260920-3vfwch_review-rev2-evidence.tar.gz`.

## Verdict: CHANGES REQUESTED (→ `to-dev`)

Both rev1 findings are fixed exactly as ruled and every rev1 mutant plus the two new pins die
(details below). One new finding blocks acceptance:

- **F3 (must fix) — the new outer-entry row is itself a spawn-pressure flake.**
  `TestGitFixtureOuterEntryLogsRetryOnTransientEACCES`
  (`internal/install/gitfixture_retry_test.go:296-333`, atomicity copy same lines) flips the
  symlink from a goroutine on a **50 ms timer** (`:321-325`) and requires the first spawn to
  hit the blocker before that timer fires. Under fork contention inside the test process —
  exactly the condition of the hosted `internal/install` lane (~100 `t.Parallel` tests
  spawning git; the very condition this leaf is about) — the first `fork/exec` routinely
  takes 100–200 ms, lands after the flip, succeeds on the real git without a retry, and the
  row fails with `logs = [], want the retry line and the success line from the outer entry`
  (`:332`). Reproduction (attached `zz_stress_test.go`: 24 parallel subtests each spawning
  `git --version` 25×, run alongside the row with `-test.count=10 -test.parallel 32` from a
  precompiled binary of the exact candidate): **4 failures in 80 plain iterations** (rounds:
  0,1,0 / 0,1,0,2,0 of 10), 0 in 20 `-race` iterations; every failure is `logs = []` with
  elapsed 0.15–0.23 s (one contended spawn, no backoff), while passing iterations take
  0.30–0.44 s (two spawns + 200 ms). The one hosted sample we have already sits on the
  boundary: in gate run 35531307725 the install copy took 0.31 s vs 0.22 s for the
  atomicity copy, i.e. ~50 ms per spawn during the install package's parallel phase on the
  3-vCPU macOS runner. The row ships in 2 packages × 4 unix lanes (Test + Race on
  ubuntu/macos) = 8 instances per gate; a false red here costs the same republish + ~60 min
  gate as the flake being fixed, and would be indistinguishable at lane level from a
  recurrence during the 10-gate observation window. The mirrored window (flip must land
  before the 200 ms retry) is not observed but is the same design.
  **Fix shape (verified):** make the flip event-driven instead of timer-driven — have the
  recording TB perform the symlink swap when it sees the `retrying once` line, which
  `runGitFixture` emits synchronously after attempt 1 failed and before
  `sleep(gitFixtureBackoff)` (`gitfixture_test.go:127`); drop the goroutine and the 50 ms
  timer, and let the swap's `os.Symlink`/`os.Rename` errors fail the test instead of `_ =`.
  Attempt 1 then always sees the 0644 blocker (real kernel EACCES) and attempt 2 always sees
  the real git — no timing anywhere. Reference implementation attached as
  `zz_fixshape_test.go` (`TestZZReviewFixOuterEntryEventDrivenFlip`): run in the same stress
  harness next to the committed row, **70/70 pass** (50 plain + 20 `-race`) while the
  committed row failed 4×. Mutant C (backoff 5 ms) is still killed by
  `TestGitFixtureRetriesOnceOnSpawnEACCES`'s literal pin (N2), so nothing is lost; mutant W
  (wiring dropped) is still killed because the event-driven row records through the same
  `t.Logf` seam. Apply to both copies and keep them in sync.

Everything else in revision 2 is accepted as verified below; revision 3 needs only F3 in
the two copies, a results.md addendum (row shape + why, stress evidence if rerun), and a
fresh green gate on the new tree.

## Rev1 findings — verified resolved

- **F1 (wired retry silent) — fixed.** `gitFixtureWithBinary` builds
  `gitFixtureConfig{binary: binary, logf: t.Logf}` (`internal/install/gitfixture_test.go:84`,
  atomicity `:88`); `gitFixture`/`gitFixtureWithBinary` take `testing.TB`, callers pass
  `*testing.T`. My rev1 probe rerun unchanged against rev2 through the wired `env.git`
  (PATH-resolved `git` script, 0644 Go-binary interpreter, chmod +x at 50 ms):
  `TestReviewProbeWiredEnvGitRetrySucceeds` PASSES in 0.88 s and the `-v` stream now
  carries `gitfixture_test.go:127: gitfixture: git [-c commit.gpgsign=false -c
  tag.gpgSign=false --version] spawn failed (fork/exec …/bin/git: permission denied);
  retrying once after 200ms` + the diagnostic block + `gitfixture_test.go:132: … succeeded
  on retry` (`retrying once`=2, `succeeded on retry`=1, `gitfixture diagnostic`=3 across the
  two probes; attached `probe-stream.log`). Rev1 count was 0. The hosted gate's
  `go-test.json` (macOS and ubuntu) records the same two lines for the passing outer-entry
  row in both packages, so the retry is visible in `test-evidence-<os>` once it fires on a
  wired path — the observation window can now tell "gone" from "masked".
- **F2 (darwin rlimits garbage) — fixed.** `gitfixture_rlimit_darwin_test.go` deleted in both
  packages; `gitfixture_rlimit_unix_test.go` (`//go:build linux || darwin`) renders
  `unix.Getrlimit` NOFILE and NPROC, no sysctl; `other` file unchanged. Rendered here
  (persistent-EACCES probe, darwin/amd64): `rlimits: NOFILE cur=122880
  max=9223372036854775807; NPROC cur=5568 max=8352`; on the hosted macOS runner (gate
  artifact): `rlimits: NOFILE cur=10240 max=9223372036854775807; NPROC cur=1333 max=2000`;
  ubuntu: `NOFILE cur=65536 max=65536; NPROC cur=63838 max=63838`; windows rows report
  unavailable and skip the two exec-bit rows. Content pinned by
  `gitFixtureRlimitsPattern` (`NOFILE cur=\d+ max=\d+; NPROC cur=\d+ max=\d+`) in
  `TestGitFixturePersistentSpawnFailureCarriesDiagnostic` on linux/darwin; my mutant R
  (`nprocText := "n/a"`) fails that row in both packages.
- **N1** EPERM row present in the predicate table (`:158`); mutant E now dies (RetryPredicate
  FAIL, both packages). **N2** backoff pinned against the literal `200*time.Millisecond`
  (`:65-67`); mutant C dies (RetriesOnceOnSpawnEACCES + the outer-entry row). **N3** row
  count corrected: 11 `TestGitFixture*` rows per package, matches the tree.

## What I verified (rev2 note items)

1. **Tree identity.** Story-worktree temp-index `write-tree` = `21bde827…` = CR candidate;
   attached patch sha256 `ea81a023…` = `git diff --binary base..candidate`; 12 changed paths,
   0 non-`_test.go` (R1 holds; `.github/`, `cmd/`, product `internal/**` untouched). Gate run
   35531307725: `headSha` `416d62a0` (parent `c3f9eeea`), `416d62a0^{tree}` = `21bde827…`;
   all 11 executed lanes success (Test/Race/Gate self-test/Lint/Naming/Interop; rose-air and
   candidate-suite skipped as in rev1). The disposable clone ended every driver at
   `21bde827…` with `git status` clean.
2. **Rows.** Precompiled `go test -c` + ad-hoc codesign (host watchdog), run from the package
   dir: install 11/11 PASS, atomicity 11/11 PASS (`rows-*.log`). Gate artifacts: 22 rows per
   unix OS all pass, windows 18 pass + 4 declared skips (`gate-artifact-analysis.txt`); 0 fail
   events in any lane; no `gitfixture` output outside the outer-entry rows (no recurrence
   during the gate).
3. **Mutants (both packages, byte-restore + sha check between mutants, `mutants-driver.log`):**
   A retry unreachable → 5 FAIL (EACCES/EAGAIN/PersistentDiagnostic/RealSpawn/OuterEntry);
   B-full retry-on-any-error → 5 FAIL (NeverRetriesNonZeroExit/WrappedExitError/
   NonRetryableFailsFast/RetryPredicate/Unresolvable); C backoff 5 ms → 2 FAIL; E +EPERM →
   1 FAIL; **W** `logf: t.Logf` dropped from `gitFixtureWithBinary` → OuterEntry FAIL (the F1
   pin); **R** NPROC "n/a" → PersistentDiagnostic FAIL (the F2 pin). All killed, identical in
   both copies; producer's kill lists match.
4. **Copies in sync.** `diff` of the four `gitfixture*` files: only the package clause and the
   atomicity header comment differ.
5. **Call sites unchanged since rev1** (`git diff 20fb22ef..21bde827` touches only the 8
   gitfixture files): install `env.git`, `testGit`, `draftBareFixture.run`, atomicity `env.git`
   keep argv/env/dir/output byte-identical on the success path (rev1 item 5 stands).
   `TestEndToEndInstall` + `TestGlobalDryRunPlansBuildsWithoutSessionOrPersistentState` PASS
   locally through the wired helper (3.95 s / 0.33 s, 0 gitfixture lines);
   `TestDryRunEffectBindingsSeeWhatARealOperationWrites` skips locally (conformance root).
6. **Lint/vet.** `gofmt -l` on the 12 files: clean; `GOOS=windows`/`GOOS=linux go vet` on both
   packages: clean. golangci-lint accepted from the green Lint job. `-race` on the new rows
   accepted from the green Race (ubuntu/macos) lanes (they run the full test gate under
   `-race`), plus my 20 race iterations under stress.

## Observations for the orchestrator (not findings against this leaf)

- The hosted macOS runner's per-uid process ceiling is `RLIMIT_NPROC cur=1333 max=2000`
  (gate artifact) — low for a lane running ~100 parallel install tests plus two other package
  binaries; exhaustion would surface as `fork … EAGAIN`, not the observed EACCES, but the new
  diagnostic now records it at failure time, which is what the bounded cause (a) needs.
- results.md rev2 says "a script interpreter fails on darwin with `exec format error` on retry
  (kernel rejects nested shebang)": true only for a *script* interpreter; a binary interpreter
  (my probe) works. The deviation to the symlink shape is fine; only the generalisation is
  loose. Correct the sentence when touching results.md.
- B3 (product-shaped run-1 occurrence not mitigated) unchanged: residual for a follow-up leaf,
  not required by the AC or R1.

## What I reran vs accepted

Reran myself: tree/patch/gate-commit identity, 11+11 rows, the two wired-path probes, 7
mutants × 2 packages, 100 stress iterations of the committed row (80 plain + 20 race) and 70 of
the fix shape, gofmt/vet, the three gate artifacts (rows, skips, output lines). Accepted from
attached evidence: the hosted full-suite/lint/race lanes of run 35531307725 on tree
`21bde827…` as the cross-platform arbiter.

## Route

`set_status(BUG-260920-3vfwch, status=to-dev)` with this file as verdict evidence. Revision 3:
F3 in both copies (event-driven flip, errors not ignored), results.md addendum, both copies
diffed, fresh green gate; everything else carried unchanged.
