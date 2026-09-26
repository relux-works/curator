# BUG-260920-3vfwch — review verdict, revision 3 (CR-BUG-260920-3vfwch-3)

Reviewer: claude-opus-5 (RUN-260920-c3fb31), independent exact-head review, 2026-09-21.
Shell: zsh on the host (darwin/amd64, go1.26.0); every throwaway test, stress round and
mutant ran in disposable clones (`git clone --no-checkout` control root → `checkout --detach
c3f9eeea` → `apply --index` rev3 patch → commit; `HEAD^{tree}` = `bde68e85…` in both clones,
`clone` for stress, `clone2` for mutants), never in the Story worktree. Evidence tarball
attached: `BUG-260920-3vfwch_review-rev3-evidence.tar.gz` (drivers, every stress/mutant/
regression log, gate-artifact analysis, rev2→rev3 diff, copies diff).

## Verdict: ACCEPTED (→ `accept_cr(BUG-260920-3vfwch, revision=3, …)`)

Revision 3 resolves the single rev2 finding (F3) exactly as ruled in 3vfwch-rework-2.md and
changes nothing else. Everything accepted at revision 2 (F1/F2/N1–N3, retry rule, diagnostic,
call sites, R1) is carried byte-identical and was not re-opened.

## Tree identity (verified)

- Story worktree temp-index `write-tree` (HEAD `c3f9eeea` + the 4 modified + 8 untracked
  `*_test.go`) = `bde68e853081cf1e3a2e3d3d66867ba1092a6441` = CR candidate tree.
- `git diff --binary c3f9eeea..bde68e85` sha256 `38a15234…` = attached
  `BUG-260920-3vfwch_change-request_rev3.patch`; 12 changed paths, all `*_test.go`
  (1246+/31−), no product file, no `.github/` change (R1 holds).
- rev2 → rev3 (`21bde827…` → `bde68e85…`): exactly 2 paths,
  `internal/install/gitfixture_retry_test.go` and
  `internal/install/atomicity/gitfixture_retry_test.go` (39 lines each: the
  `gitFixtureRecordingTB` type + `TestGitFixtureOuterEntryLogsRetryOnTransientEACCES`
  tail). The two hunks are identical (`diff` of the two per-file diffs: empty).
  `gitfixture_test.go`, the rlimit files and the four wired call sites are unchanged since
  rev2 — the retry policy and the diagnostic were not touched.
- Copies in sync at rev3: `gitfixture_retry_test.go`, `gitfixture_rlimit_unix_test.go`,
  `gitfixture_rlimit_other_test.go` differ only in the package clause;
  `gitfixture_test.go` only in the package clause + the standing atomicity header comment.
- Gate run 35535957629 (branch `gate/STORY-260915-3w11un/260920-203400-55228-1`): `headSha`
  `01ffa90a166f03682e657adfa095182f7436e8f1`, `01ffa90a^{tree}` = `bde68e85…` (candidate),
  parent `c3f9eeea` (base). Conclusion success; all 11 executed lanes success — Test/Race
  ubuntu+macos, Test windows (58 min, inside the 120 m budget), Gate self-test ×3, Lint,
  Interop, Naming; Candidate suite and rose-air skipped as in rev1/rev2.

## F3 — outer-entry row is event-driven (verified resolved)

Shape (`internal/install/gitfixture_retry_test.go:284-347`, atomicity same lines):
`gitFixtureRecordingTB` gains `flip func()`; its `Logf` records the line, forwards it to the
real `t.Logf`, and runs `flip` when the line contains `retrying once` (`:293-300`).
`runGitFixture` emits that line synchronously after attempt 1 failed and before
`sleep(gitFixtureBackoff)` (`gitfixture_test.go:127-128`, unchanged), all in the test
goroutine, so attempt 1 always sees the 0644 blocker (real kernel EACCES on an absolute
path — no LookPath, no X_OK pre-check) and attempt 2 always sees the real git. The row
(`:331-339`) builds the flip from `os.Symlink(realGit, fresh)` + `os.Rename(fresh, link)`
with `t.Fatal(err)` on both (no `_ =`), passes it in the recording TB (`:343`) and asserts
the two logged lines (`:344-346`). No `go func`, no `time.Sleep`, no `_ = os.` anywhere in
either file (grep). This is my rev2 reference (`zz_fixshape_test.go`) modulo names and
comments. `t.Fatal` inside `flip` is legal: the call chain is synchronous from the test
goroutine.

Stress (my rev2 harness `zz_stress_test.go` verbatim, copied into both packages, precompiled
+ ad-hoc signed, run from the package dir with `-test.count=10 -test.parallel 32`, mask
`TestGitFixtureOuterEntryLogsRetryOnTransientEACCES$|TestZZStressSpawn`, source restored and
tree re-verified `bde68e85…` before the rounds):

| binary | rounds × iterations | committed row | stress rows | `DATA RACE` |
|---|---|---|---|---|
| install plain | 5 × 10 | 50/50 PASS | 0 FAIL | – |
| atomicity plain | 3 × 10 | 30/30 PASS | 0 FAIL | – |
| install `-race` | 2 × 10 | 20/20 PASS | 0 FAIL | 0 |
| atomicity `-race` | 1 × 10 | 10/10 PASS | 0 FAIL | 0 |

**110/110** (rev2 timer shape under the same harness: 4 failures / 80). Every iteration
logged exactly one `retrying once after 200ms` and one `succeeded on retry` (10+10 per
round log) and took 0.26–0.62 s — two spawns plus the 200 ms backoff — i.e. the retry path
was exercised every time and never skipped by an early flip. Round timings 30–38 s per
round; the first install round shows 317 s wall only because the host's exec-stall window
(fresh binaries stranded at `_dyld_start`, SIGKILL after ~90 s, three attempts, `syspolicyd`
restarted 01:59 local) preceded it — a host artefact, not a test result.

## Mutants (both packages, byte-restore + sha check between mutants; `mutants3-driver.log`)

Baseline 11/11 PASS in both packages (install `gitfixture_test.go` sha `c9133e622dcb`,
atomicity `8b83e41f6138` — the shas the producer reports). Every mutant killed, with
identical FAIL-row sets in the two packages:

- A retry unreachable → 5 FAIL (EACCES/EAGAIN/PersistentDiagnostic/RealSpawn/OuterEntry).
- B-full retry on any error → 5 FAIL (NeverRetriesNonZeroExit/WrappedExitError/
  NonRetryableFailsFast/RetryPredicate/Unresolvable).
- **C** backoff 5 ms → 1 FAIL: `TestGitFixtureRetriesOnceOnSpawnEACCES` (N2 literal pin).
  The outer-entry row now passes under C — expected: the flip keys off the logged line, not
  the clock; the pin still kills C in both copies.
- E predicate + EPERM → 1 FAIL (`RetryPredicate`).
- **W** `logf: t.Logf` dropped from `gitFixtureWithBinary` → 1 FAIL:
  `TestGitFixtureOuterEntryLogsRetryOnTransientEACCES` fails fatal at `:343` with
  `git [--version]: fork/exec …/001/git: permission denied (first attempt: … permission
  denied)` plus the diagnostic block (`binary: lstat: mode=Lrwxr-xr-x -> …/blocker …;
  stat: mode=-rw-r--r-- size=15`, dir stat, cwd, PATH, `rlimits: NOFILE cur=122880
  max=…; NPROC cur=5568 max=8352`) — without the log line the flip never fires, so the
  retry hits the blocker; the wiring pin holds, with the honest new failure mode the
  producer describes.
- R NPROC "n/a" → 1 FAIL (`PersistentSpawnFailureCarriesDiagnostic`).
Both clones ended at `bde68e85…` with `git status` clean.

## Gate artifacts (run 35535957629, all five evidence artifacts downloaded)

- `test-evidence-macos-latest` / `-ubuntu-latest` and `race-evidence-macos-latest` /
  `-ubuntu-latest`: 22 `TestGitFixture*` rows each, all pass; outer-entry row elapsed
  0.21–0.29 s (two spawns + 200 ms — the retry path fired on the hosted runners too).
  `test-evidence-windows-latest`: 18 pass + 4 declared skips (the two exec-bit rows per
  package). 0 `fail` events in any lane.
- The only `gitfixture` output lines in all five lanes come from the outer-entry rows (both
  packages), each with the `retrying once after 200ms` + `succeeded on retry` pair — no
  recurrence of the flake during this gate, and the retry log is visible in `go-test.json`
  for passing tests, as the observation window needs. The single other `permission denied`
  string per unix lane is the expected refusal diagnostic (`mkdir … permission denied`) of the
  passing `crossconformance :: TestDraftSourcesCLIInstallRestoresPriorState` row — a planted
  read-only directory, not a git spawn.

## Unchanged surface re-checked (cheap, not re-opened)

- Wired call sites (bytes unchanged since rev1/rev2): from the clean baseline binaries of the
  exact tree, `TestEndToEndInstall` (5.13 s), `TestGlobalDryRunPlansBuildsWithoutSessionOr
  PersistentState` (0.43 s), `TestDraftGitRuntimeMaterializesUnderSourceV1Key` (8.17 s),
  `TestDraftTransportAuth` (2.87 s), `TestAcquireDraftNetworkSelection` (5.44 s), atomicity
  `TestStableHybridActivationCommitsWithoutRestarting` (6.87 s) all PASS with 0 `gitfixture`
  lines (no retry needed locally); `TestDryRunEffectBindingsSeeWhatARealOperationWrites`
  skips locally (conformance root not set) and passed on the hosted lanes.
- `gofmt -l` on the 12 changed files: clean. `go vet` on both packages for darwin, linux and
  windows: clean. golangci-lint accepted from the green Lint job.

## results.md revision 3 — checked against my observations

Two-path rev2→rev3 delta, copies identical modulo package clause, stress 50/50 (mine
110/110), mutant C's changed kill set, mutant W's changed failure mode, restored shas,
rev2 §F1 correction note, observation-window text: all consistent with what I measured.

## Non-blocking notes

- `gitFixtureRecordingTB`'s comment says the flip "runs exactly once"; the code runs it on
  every line containing `retrying once`. In this row that is exactly one line, so the
  comment is true for the row but not enforced — cosmetic, no action needed.
- Observation window (AC disjunct 1) remains the orchestrator's: the next 10
  `Test (macos-latest)` runs after landing. The rev3 row cannot false-red under spawn
  pressure, so a red in the window is signal; a helper-shaped recurrence now shows as the
  `retrying once` line + diagnostic in `test-evidence-macos-latest/test/go-test.json`.
- Residual B3 (product-shaped run-1 occurrence in `gitops.writeBlobs`, not retried or
  diagnosed) stands as a follow-up leaf, not a finding against this leaf (R1).

## What I reran vs accepted

Reran myself: tree/patch/gate-commit identity, rev2→rev3 diff and copies diff, 110 stress
iterations (plain + `-race`, both packages), 6 mutants × 2 packages with baselines, 7
wired-path regression rows, gofmt/vet ×3 GOOS, the five gate artifacts (rows, skips, output
lines). Accepted from attached evidence: the hosted full-suite/lint/race lanes of run
35535957629 on tree `bde68e85…` as the cross-platform arbiter.

## Route

`accept_cr(BUG-260920-3vfwch, revision=3, evidence=BUG-260920-3vfwch_review-verdict-rev3.md)`
— accepted, routed to `integrating`; landing is the orchestrator's integration step.
