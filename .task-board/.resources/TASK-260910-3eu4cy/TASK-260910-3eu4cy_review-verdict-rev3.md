# TASK-260910-3eu4cy — review verdict, revision 3: ACCEPT (with one landing-time finding for the orchestrator)

Reviewer run RUN-260919-759f8f (claude-opus-5), 2026-09-19. Change Request
CR-TASK-260910-3eu4cy-3, base 6f173778, candidate tree 93688bde, patch
sha256 `1907d391…` (matches the resource). Shell for every command:
`#!/bin/bash` drivers with `set -o pipefail`, one log line per command with
the binary's own exit code; nothing piped through `tail` for an exit code.
Evidence bundle: `TASK-260910-3eu4cy_review-rev3-evidence.tar.gz` (probe
tests, raw `-v` outputs, driver logs, mutant scripts and outputs, the gate
timing scripts and their outputs).

## Verdict

**ACCEPT → `accept_cr(TASK-260910-3eu4cy, revision=3, evidence=this file)`.**
Every rev1 finding is closed exactly per the orchestrator rulings (F1 option A,
F2 refusal path, F3 `Result.Attestations` row, F4 probe D + legacy-marker row),
the rev1 probes now pass on the candidate, every mutant I asked for dies at the
production entry, the parts of the delta confirmed in rev1 are byte-identical,
and the hosted gate is green on the exact candidate tree (Windows included).

One finding is NOT about the code and does not block acceptance, but it needs
an orchestrator decision **before the landing PR**: **F-W1** — the Windows
`internal/install` lane finished in 3516 s of its 3600 s package budget on the
rev3 gate (84 s margin); the identical code timed out on rev2 at the same
runner speed. Numbers, root cause and options are below. Nothing in this leaf
hangs; the lane is simply over budget on about half of the observed hosted
Windows runner speeds, and this leaf's sweep is the marginal 21–25 minutes.

## Candidate identity (exact tree, three ways)

| proof | result |
|---|---|
| Story worktree working tree (`GIT_INDEX_FILE=$(mktemp) git read-tree HEAD && git add -A . && git write-tree`) | `93688bdeb7cd808b7b6c8f2af77aa6579a39fa8c` = CR candidate ✔ |
| Disposable clones `/tmp/3eu4cy-rev3` and `/tmp/3eu4cy-rev3-mut`: `git checkout --detach 6f173778 && git apply --index <rev3 patch> && git commit` → `HEAD^{tree}` | `93688bde…` ✔ both; both end the review at `93688bde…` with 0 dirty paths (only my two `zz_review_*_test.go` probes untracked in the first) |
| Hosted gate run 35424565415: `headSha` `0d53f0f2…` ("remote gate snapshot of STORY-260910-1s75e1") → `git rev-parse 0d53f0f2^{tree}` | `93688bde…` ✔ — `conclusion: success`: Test ubuntu/macos/windows, Race ubuntu/macos, Lint, Naming, Interop, Gate self-test ×3 all `success`; rose-air and candidate suite skipped (main-push only). Windows Test job 05:40:14Z → 06:43:46Z |

Delta shape (all by `git diff` between trees, not by reading the patches):

- rev2 → rev3: exactly one added file, `internal/install/cache_rebuild_deadline_test.go`
  (the two patches are otherwise byte-identical after dropping `index` lines). **No
  production change in rework 2**, so nothing could have weakened a protected-store
  proof; the "bounded wait → typed refusal" question is moot for this revision.
- rev1 → rev3 (rework 1): 6 files, +423/−86 — production: `cmd/curator/main.go`
  (stale-lock diversion removed, comment), `cmd/curator/draft_status.go`
  (`draftLockIsStale` deleted, `CheckStale != nil` → error), `internal/install/install.go`
  (v5 moved-tag arm and `recordedPackageCommit` deleted, the bound stated at the
  gate and its call site; the legacy predicate textually unchanged). Tests:
  `draft_status_test.go` (+5 CLI/unit rows, stale rows re-pointed),
  `draftpublish_test.go` (moved-tag row → two negative rows + legacy control + F3 row).
- `internal/transaction`, `internal/adapters`, `internal/runtimestore`,
  `internal/staging`: no diff base → rev3 (trunk engine untouched).
  `internal/closure`, `internal/registry`, `internal/sourcelock`, `internal/marker`,
  `internal/install/global.go`, `internal/install/draftatomic_test.go` (the sweep): no
  diff rev1 → rev3 — everything the rev1 verdict confirmed is byte-identical.
- No scratch files in the candidate (`git ls-files | grep zz_` finds only board
  resources); `grep -rn "recordedPackageCommit|draftLockIsStale"` over cmd/ internal/ is empty.

## Rulings verified at the production entries

| rev1 finding | ruling | rev3 evidence (candidate clone, prebuilt binaries, `-count=1 -v`) |
|---|---|---|
| F1 v5 moved-tag false positive | option A: drop the v5 arm | `install.go:1245-1247` `if recorded.Package != nil { continue }` with the bound stated at :1229-1234 and the call site :583-585; my rev1 probe A (`zz_review_tagbump_test.go`, unchanged) now PASSES: declared bump v1→v2 under `StrictTags` → `status=ok`, no "moved tag", marker binds v2; legacy control PASS. Producer rows `TestDraftDeclaredTagBumpIsNotAMovedTag`, `TestDraftSameTagMoveInstallsWithoutWarning` (refresh binds the moved commit; install warns nothing), `TestLegacyDeclaredTagBumpIsNotAMovedTag` PASS; trunk `TestMovedTagWarningAndStrict`, `TestGlobalStrictTagsDetectMovedTag` PASS (legacy predicate intact). Mutant **M-v5** (arm restored) KILLED exit 1 — both draft rows fail with the rev1 false positive, both legacy rows still pass. |
| F2 stale lock fail-open | refusal path | `main.go:766-771`: `!BuildsComplete` → `printResult` + `exitFail` + `continue` for every project (the block is byte-identical to trunk except comments); `draft_status.go:83-89` returns the `CheckStale` error. My rev1 probes B/C re-armed for the ruling PASS: zero-member lock + declared skill → `status` exit 1, stderr `error: manifest_sha256: source_lock_stale: Skillfile changed since lock; explicit refresh required`, no rows; `--check` exit 1; `--json` exit 1 with `[]` (the historical failed-dry-run shape); swapped selection (lock {review}, include ["other"]) → same refusal, stdout never mentions review or other; **recovery**: `project refresh` → `app: other not-installed`, `--check` 1; `install` → `app: other up-to-date`, `--check` 0, the dropped `review` directory removed by the managed removal. Producer rows `TestDraftStatusStale{ManifestRefuses,EmptyLockRefuses,SwappedSelectionRefuses}` + unit `TestDraftStatusDriftFailsClosedOnStaleLock` PASS. Mutant **M-div** (rev1 diversion + rows restored, both halves) KILLED exit 1 — all three CLI rows fail with "status = 0, want nonzero". |
| F3 `Result.Attestations` unverified | dry-run row at `install.Project` | `TestDraftDryRunExposesResolvedAttestations` PASS (0.8 s): injected `ResolveAttest` map equals `Result.Attestations`; failed resolution → `failed` + nil. Mutant **MR7** (`result.Attestations = nil`, `install.go:532`) KILLED exit 1 ("Attestations = map[], want the injected map"). The optional signed-registry `status` row was not added (producer's stated bound; the comparison is covered at `classifyDraftMember` 12 rows + this exposure row) — acceptable. |
| F4 classifier gates killed only at the helper | probe D + legacy-marker CLI row | `TestDraftStatusLockOnlyChangeNeedsInstall` (probe D verbatim in shape) and `TestDraftStatusLegacyMarkerNeedsInstall` PASS; my own probe D PASS (`app: other not-installed`, `app: review needs-install`, `--check` 1). Mutant **MR1** (drop `lock_sha256` comparison, `draft_status.go:134`) KILLED at the CLI (exit 1, "app: review up-to-date") and at the classifier; **MR8** (legacy marker reported current, :121-125) KILLED at the CLI exit 1. |

Additional mutant: **M-stale** (`draftStatusDrift` ignores `CheckStale` and classifies
rows from a stale lock) SURVIVES the four CLI rows (exit 0) and is KILLED by the unit
row (exit 1, "drift with a stale lock returned rows: map[review:not-installed]"). This
is an equivalent mutant at the CLI, not a gap: the read-only dry run's frozen-consumption
gate (`source_lock_stale`, trunk 3vxe3y) refuses before classification is reached, so the
drift-level check is only reachable in the race window between the dry run and the
classification — exactly what its comment says. Recorded as a bound.

## Rework 2 (Windows) — what I verified

- The producer's forensics are correct and I reproduced them from the gate artifacts
  (`gh run download`, `test-evidence-windows-latest`, `go-test-served.json`) of four runs:
  the 1xs0pj rev3 gate 35402478828 (trunk + sibling = the pre-leaf baseline), rev1
  35413340404, rev2 35418260010, rev3 35424565415. Reconstructed per-test timelines
  (`timeline.py`) show the serial phase (non-`t.Parallel` tests, run one after another)
  followed by a ~6–9 minute parallel phase; the rev2 alarm fired during the parallel phase
  with `TestAuthoritativeCacheRejectionsAreRebuiltNeverAdopted` 5 s into its 9th subtest,
  as the producer says — no hang anywhere.
- rev2 → rev3 has no production change (above), and rev3 passed Windows with the same
  code, so there was nothing to bound in production. The new bounded regression
  `TestCacheWrongTargetRebuildCompletesWithinDeadline` runs at `install.Project` without
  the conformance root (PASS locally 7.9 s, both `nonReusableStatuses`; 26.5 s on the
  Windows gate). Watchdog proofs reproduced: hang injected via `builder.observe = func(StageRequest){ select{} }`
  with the watchdog at 5 s → exit 1 in 16 s ("install did not return within 5s; refusing
  to wait out the package timeout"); the same hang with the `timer.C` arm removed (plain
  `<-done`) → `panic: test timed out after 1m0s`, exit 2 under `-test.timeout 60s`. The
  bound is test-side only; production keeps `runWithRestarts`/`MaxRestarts` (trunk).

### F-W1 — Windows `internal/install` lane is at the edge of its 60-minute package budget (landing risk, orchestrator decision)

Measured from the four Windows evidence streams (package = `internal/install`, budget
`GO_TEST_TIMEOUT` 60m at `.github/workflows/ci.yml:213`, also `:702` for the candidate
suite; the Test job itself has no `timeout-minutes`):

| run | package result | serial phase | parallel phase | this leaf's sweep | runner-speed proxy¹ |
|---|---|---|---|---|---|
| baseline 35402478828 (trunk + 1xs0pj, no 3eu4cy) | pass 2928 s (81 % of budget) | 2485 s | 443 s | — | 1.00 |
| rev1 35413340404 | pass 2894 s | 2526 s | 368 s | 1257 s | 0.55 |
| rev2 35418260010 | **timeout 3600 s** | 3132 s | ≥ 468 s (cut) | 1529 s | 0.74 |
| rev3 35424565415 | pass **3516 s (97.7 %)** | 2980 s | 536 s | 1491 s | 0.74 |

¹ sum of the 187 top-level install tests present in all four runs, relative to the
baseline run — i.e. how fast that day's runner was; the baseline's runner was the slowest
observed (e.g. `TestDraftAuditRenewalMarkersNeverRenewable` 755 s there vs 191–251 s in
the other three runs), rev1's the fastest.

Reading: the pre-existing wave-3 suite already consumes 48 min on a slow runner; this
leaf adds ~1500 s of serial time on Windows, of which the sweep
`TestDraftFailureAtEveryTargetClassRestoresPriorState` is 1257–1529 s (its per-class
cost grows with the number of committed targets to roll back: 10-context 94 s …
80-removal 344 s, 90-consumer 331 s on rev3; 6–12 s per class on this Mac). rev2 and
rev3 ran on equally fast runners (0.74) and landed on opposite sides of the budget by
~150 s of noise; on a baseline-speed runner the rev3 suite extrapolates to
3516 / 0.74 ≈ 4750 s ≈ 79 min. So after landing, every push to main has a substantial
(order of one in two, on the four samples) chance of a red Windows lane in
`internal/install`, and the same applies to every later Story's gate.

This is not a defect in the delta: the sweep is the AC's proof ("inject a failure at
every publication step"), it must not be weakened, and the per-install cost on Windows
is the trunk transaction engine's (the rev2 timeout stack sits in
`saveJournal → validateIndependentTargetNamespaces → canonicalNamespacePath →
filepath.EvalSymlinks` — one symlink walk per target per journal save, which is what
makes a late-class rollback cost 5–6 minutes on Windows). It IS a landing-time
capacity problem that this leaf tips over the edge. Options, with what I measured:

1. **Raise the Windows package budget** (`ci.yml:213` and `:702`, `'60m'` → `'90m'`; the
   ci.yml comment already frames it as "a hosted-runner budget, not a behavioural bar …
   a hang is still bounded and still fatal"). One line, outside this leaf's scope line,
   so it is the orchestrator's to carry (landing PR or a tiny CI leaf). 90m covers the
   baseline-speed extrapolation (79 min) with ~11 min to spare; 120m also covers a
   runner 1.5× slower than baseline. **Recommended as the stopgap.**
2. **In-leaf test-cost trim**: `t.Parallel()` on the sweep moves its 1491 s out of the
   serial phase; the parallel phase then becomes the sweep's own length, so the package
   drops by at most the current parallel phase (≈ 536 s, ~9 min) → ~83 % of the 60m
   budget on a 0.74 runner and still > 60m on a baseline-speed one. Optional, not
   sufficient alone; costs another gate + review cycle.
3. **Follow-up performance leaf on the engine** (Windows journal validation /
   EvalSymlinks per target per save; fsync policy): cuts both the sweep and users'
   real rollback time. Right fix, not this Story.

Decision needed from the orchestrator before the landing PR: option 1 (and at which
value), with or without 2. I have not changed anything.

## Independent reruns (candidate clone at 93688bde, prebuilt binaries run from the package dirs, `-test.count=1 -test.v`)

Build/vet/fmt: `go build ./...` exit 0 (12 s); `go vet` over cmd/curator, internal/install,
internal/closure, internal/registry, internal/sourcelock, internal/marker exit 0;
`gofmt -l cmd internal` empty (exit 0); `go test -c` ×6 exit 0; probe binaries ×2 exit 0.

| run | mask | result |
|---|---|---|
| `internal/sourcelock` | whole package | exit 0 — 31 pass |
| `internal/registry` | `TestAttestRoot\|TestV5AttestIdentity` | exit 0 — 5 pass |
| `internal/closure` | `TestRefreshDraft\|TestOpenDraftFrozen` | exit 0 — 9 pass, 0 skip |
| `internal/marker` | whole package | exit 0 |
| `cmd/curator` probes | `TestReview` (B, C+recovery, D) | exit 0 — 3 pass, 11 s |
| `cmd/curator` draft | `TestDraftStatus\|TestClassifyDraftMember` | exit 0 — 16 pass (rev1's 11 + 5 new), 30 s |
| `cmd/curator` legacy | `TestStatus\|TestProjectResolve` | exit 0 — 34 pass, 177 s |
| `internal/install` probe A + legacy control | `TestReview` | exit 0 — 3 pass, 50 s |
| `internal/install` F1 triple + F3 | the four names | exit 0 — 4 pass, 37 s |
| `internal/install` deadline regression | `TestCacheWrongTargetRebuildCompletesWithinDeadline` | exit 0 — 1 pass (2 subtests), 8 s |
| `internal/install` legacy | `TestLegacyInstallUntouchedWhenDraftOff\|TestMovedTagWarningAndStrict\|TestGlobalStrictTagsDetectMovedTag\|TestAuthoritativeCacheRejections` | exit 0 — 3 pass, 1 skip (authoritative suite needs `CURATOR_CONFORMANCE_ROOT`; gate-only), 19 s |
| `internal/install` sweep | `TestDraftFailureAtEveryTargetClassRestoresPriorState` | exit 0 — 7/7 classes, 83 s |
| `internal/install` publish rows | the 8 `draftpublish_test.go` rows (refresh-then-install, retarget, admitted replacement, repair ×3, stale bindings, unmanaged adapter) | exit 0 — 8 pass, 79 s |
| `internal/install` full prescribed mask | `TestAuthoritativeCacheRejections\|TestDraft\|TestLegacyInstallUntouchedWhenDraftOff` | exit 0 — 70 pass, 0 fail, 1 skip (authoritative suite, gate-only), 479 s |

Windows/Ubuntu rows and the race lanes: hosted gate on the exact candidate tree only
(this host is macOS). No new skip class; the platform-case gate passed on all three lanes.

## Narrowing mutants (mutant clone `/tmp/3eu4cy-rev3-mut`, candidate committed, `git checkout --` restores; final tree 93688bde, 0 dirty)

| id | mutant | killer | result |
|---|---|---|---|
| MR7 | `result.Attestations = nil` (`install.go:532`) | `TestDraftDryRunExposesResolvedAttestations` | KILLED exit 1 |
| M-v5 | rev1 v5 moved-tag arm restored (`install.go:1245`) | `TestDraftDeclaredTagBumpIsNotAMovedTag`, `TestDraftSameTagMoveInstallsWithoutWarning` | KILLED exit 1 (2 fail; `TestLegacyDeclaredTagBump…`, `TestMovedTagWarningAndStrict` still pass) |
| M-wd-hang | hang in `builder.observe`, watchdog 5 s | deadline regression | KILLED exit 1 in 16 s (bound trips) |
| M-wd-unbounded | same hang, `timer.C` arm removed | deadline regression under `-test.timeout 60s` | exit 2, `panic: test timed out after 1m0s` (without the bound the hang consumes the timeout) |
| MR1 | drop `recorded.LockSHA256 != lockSHA256` (`draft_status.go:134`) | `TestDraftStatusLockOnlyChangeNeedsInstall` (CLI) + classifier `lock-mismatch` | KILLED exit 1 at both |
| MR8 | legacy marker → `up-to-date` (`draft_status.go:121`) | `TestDraftStatusLegacyMarkerNeedsInstall` (CLI) | KILLED exit 1 |
| M-div | rev1 stale-lock diversion + rows restored (`main.go`, `draft_status.go`) | the three `…Stale…Refuses` CLI rows | KILLED exit 1 (3 fail) |
| M-stale | `draftStatusDrift` ignores `CheckStale` | 4 CLI rows: SURVIVED exit 0 (equivalent — the dry run refuses first); `TestDraftStatusDriftFailsClosedOnStaleLock`: KILLED exit 1 | bound, see above |
| MR3 (rev1) | engine rollback skips `20-runtime` targets | sweep | accepted from rev1: the mutated engine (`internal/transaction`) and the killer (`draftatomic_test.go`) are byte-identical between the rev1 and rev3 trees |

## Acceptance clauses — status

| clause | evidence | status |
|---|---|---|
| one recoverable transaction; faults roll back consistently | sweep 7/7 at `install.Project` (whole-state digest, reverse-order rollback, no journal); refresh first/second-write/crash-window rows (closure 9 pass); files byte-identical to rev1 where MR3 died | holds |
| recheck boundaries at write time | retarget + admitted-replacement rows fail `source_output_overlap` from a `PointPrepared` mutation with full rollback (8 publish rows pass; identical to rev1) | holds |
| status enforces exact currentness and frozen inputs | up-to-date/needs-install/content-drift/not-installed/invalid-marker/live-mutation rows; **stale lock now refuses** (3 CLI rows + probes B/C, M-div dead); lock-only change (probe D, MR1 dead at the CLI); legacy marker (MR8 dead at the CLI); attestation exposure (MR7 dead) | holds |
| repair revalidates the locked source and evidence | 3 repair rows pass (identical to rev1) | holds |
| refresh replaces lock+marker only after success | refresh-then-failed-install / stale-bindings rows pass (identical to rev1) | holds |
| unmanaged files protected | `TestDraftUnmanagedAdapterDestinationRefuses` pass | holds |
| retarget / copy mutation failure paths | proven (above) | holds |
| identity consumers deferred by 1xs0pj | `AttestRoot` v5 (registry 5 pass, unchanged); `scopeStatusDrift` replaced by the draft lane for schema-2 + switch, 34 legacy rows byte-identical; `detectMovedTagsIn`: **bounded per ruling A** — schema-5 markers never match; `--strict-tags` is therefore inert on the draft lane by design (a move is observable only at explicit refresh) | holds (bound stated at the gate and its call site) |
| legacy goldens with the switch off | `TestLegacyInstallUntouchedWhenDraftOff`, legacy moved-tag rows, 34 legacy status/resolve rows | holds |
| hosted gate green on the exact tree | run 35424565415, tree 93688bde | holds — see F-W1 for the margin |

## Bounds (not findings)

- `status --json` on a stale lock prints `[]` with exit 1 and the refusal on stderr —
  the historical failed-dry-run shape shared with the legacy lane; a JSON consumer must
  read the exit code. Not new in this leaf.
- `draftStatusApplies` returns false on a manifest read failure (falls back to the legacy
  surface), which only matters in the race between the dry run and classification; the
  legacy surface then reports its own read failure. Unchanged from rev1; not re-opened.
- The authoritative cache-rejection suite skips locally (no `CURATOR_CONFORMANCE_ROOT`);
  the self-contained deadline regression is its always-on counterpart; the gate ran both.
- `internal/transaction` rollback cost on Windows (5–6 min for a late-class failed
  install in this fixture) is trunk behaviour surfaced by the sweep — a performance
  follow-up candidate, outside this leaf.

## Required for landing (orchestrator)

1. Decide F-W1 before opening the landing PR: raise the Windows `GO_TEST_TIMEOUT`
   (`ci.yml:213`, `:702`) to ≥ 90m (120m recommended for runner variance), optionally
   pair it with the `t.Parallel()` trim, and/or queue the engine performance leaf.
   Without it, expect the Windows `internal/install` lane to time out on roughly every
   other push at current test cost.

Verdict recorded via `accept_cr(TASK-260910-3eu4cy, revision=3, evidence=TASK-260910-3eu4cy_review-verdict-rev3.md)`.
