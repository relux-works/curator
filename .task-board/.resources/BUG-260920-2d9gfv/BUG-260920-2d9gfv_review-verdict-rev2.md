# BUG-260920-2d9gfv — review verdict, revision 2 (reviewer, claude-opus-5, RUN-260921-6f01cc)

**Verdict: ACCEPT** — `accept_cr(BUG-260920-2d9gfv, revision=2, evidence=BUG-260920-2d9gfv_review-verdict-rev2.md)`.

Change Request `CR-BUG-260920-2d9gfv-2`, base `e0f52c95`, candidate tree
`48ad45d6769d3172028155fea80f5f2f529c10a0`. Exact-tree proofs: Story worktree temp-index
`git read-tree HEAD && git add -A . && git write-tree` = `48ad45d6…` (worktree read-only, `git status`
unchanged); disposable clone base + `BUG-260920-2d9gfv_change-request_rev2.patch` (sha256
`649351d5…` = CR record) committed → `HEAD^{tree}` = `48ad45d6…`; hosted gate run 35555033097
(`gate/STORY-260919-37szes/260921-024159-52487-1`, head `2e82f98c`, parent `e0f52c95`) →
`2e82f98c^{tree}` = `48ad45d6…`, conclusion success, all 11 served jobs green (Candidate suite and
rose-air skipped as on every leaf of this Story).

Revision 2 does exactly what verdict rev1 §7 / `2d9gfv-rework-1.md` asked: the future-timestamp
tolerance is now the checker's own latency since **function entry**, so the bound at both production
call sites is the post-fetch wall clock + configured skew regardless of how many trusted registries are
configured or how slow their siblings were; the rev1 residual (F1) is closed and is pinned by the
inverted sibling unit row and the committed two-registry production-entry row (both fail on the exact
rev1 product file — mutant M3 below). R1–R4 are met; the AC (`attestation-evidence-unreadable` /
`-wrong-key` deterministic under `-race`, classification fixed with a production-entry test and
narrowing mutants) is met — and the determinism is now structural, not statistical: with
`bound = now + skew + time.Since(start)` and `start` the first statement after the call site's
`time.Now()`, a serve-time mint satisfies `created_at = trunc(t_serve) ≤ t_serve < t_check ≈ bound`
for every registry in the list.

## 1. rev1 → rev2 delta (review-note: "nothing else")

`git diff f5c03b9c…(rev1 tree) 48ad45d6…(rev2 tree) --stat`: `CHANGELOG.md` (+6/−6),
`internal/install/registry_e2e_test.go` (+68/−9), `internal/registry/registry_test.go` (+62/−31),
`internal/registry/snapshot.go` (+22/−14 — comments, `start := time.Now()` as the first statement of
`checkSnapshotsWithPolicy`, the in-loop `checkStart` removed, the check
`parsed.CreatedAt.After(now.Add(clockSkew).Add(time.Since(start)))`). The crossconformance harness
file is unchanged between revisions (its R3 comment came with rev1; `crossSecondBoundary` stays 50 ms
with the corrected comment, F2). Nothing else. Non-test diff base→rev2 = `snapshot.go` + CHANGELOG
+ one harness comment (5 files, +288/−17 overall).

## 2. R1 shape and the rev1 residual (note item: two-registry probe)

- Product: `start` is the first statement; the only uncovered interval is call-site
  `time.Now()` → function entry (`install.go:1347/1353` evaluate `time.Now()` as an argument and
  call immediately — nanoseconds). Both comments rewritten as requested (checker's own latency since it
  sampled its clock; post-fetch clock + skew for production callers; ~zero tolerance for injected-`now`
  callers with instant fetches; a genuinely future timestamp still refused). Stale check unchanged
  (`snapshot.go:167`, pre-fetch `now` — can only under-report staleness).
- My rev1 probe `TestZZReviewSlowSiblingFlipsInstantRegistryUnderZeroSkew` (untracked copy in the
  candidate clone, removed afterwards) against rev2: `go test ./internal/install/ -race -count=3 -run
  TestZZReview -v` → **3/3 PASS**, `status=ok errors=[]`, no future warning, installs 6.4–7.8 s
  (rev1: 3/3 FAIL `registry reg-b snapshot timestamp is too far in the future`).
- The committed row `TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry`
  (`registry_e2e_test.go:270`) is that probe's shape (v1 lane, strict policy, zero skew, `reg-a` slow
  + pending with `crossSecondBoundary`, `reg-b` instant + audited + serve-time mint); order is
  config order (`config.TrustedRegistries` preserves `AuditRegistries` order), so the slow sibling is
  always checked first. Under rev2 it cannot flip (see header); under the rev1 file it fails
  deterministically except in the ≈3 ms/s window where the pre-fetch `now` sits within reg-b's own
  check latency of the boundary — a rare non-kill of the mutant, never a false failure of the candidate.

## 3. Semantics (note item 2 of rev1, re-checked on rev2)

- `now` stays the reference time: exact-bound pins `TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew`
  (12 subtests × 3 skews incl. `exactly on the bound` accepted / `one second past` refused) and
  `TestSnapshotZeroClockSkewIsLiteral` unchanged and green 5/5 under `-race`.
- Inverted sibling rows (`registry_test.go:785` `TestSnapshotFutureBoundToleratesCheckerLatency`):
  after a 1.5 s sibling an instant registry with `created_at = now + 1 s` (≥ 0.5 s behind the wall
  clock) is accepted; with `created_at = now + skew + 3 s` (≥ 0.5 s ahead of the post-fetch wall
  clock, skews 0 and 30 s) it is still refused with `fast snapshot timestamp is too far in the
  future`; own-fetch accept row and `now + skew + 2 s` refuse row kept byte-identical. Accept rows are
  deterministic (`time.Sleep` never returns early, so elapsed ≥ 1.5 s > 1 s); refuse rows need the
  in-process latency to stay < 1.5 s / < 2 s — observed 1.7–2.0 s total for the sibling row under
  `-race` at load 13 (i.e. ≈0.2–0.5 s of overhead on top of the sleep).
- Read-only status path: shared body, fix applies by construction; still no committed row at zero
  skew with a serve-time mint — bound, documented in results.md as allowed in rev1.

## 4. Legacy v1 lane (R2)

CHANGELOG `## Unreleased` → `### Fixed` (lines 139–151) with the function-entry wording ("its own
latency since it sampled its clock … genuinely ahead of the post-fetch clock plus skew is still
refused"). The three e2e rows run the frozen v1 lane (schema-1 Skillfile, `install.Project` →
`resolveRegistries` → `CheckSnapshotsWithPolicy`, zero skew, serve-time-signed snapshot; negative row
refuses with the same class + text). Legacy goldens `TestDraftTransportLegacyGolden` (23.5 s) and
`TestProjectResolveLegacyUntouched` PASS (my rerun, `ok 24.2 s`); also PASS on all hosted lanes where
they run (the golden skips on Windows by its own pre-existing POSIX-wrapper guard,
`draft_transport_test.go:482`).

## 5. Determinism (R3, note item 4)

- Producer (results.md): `go test ./internal/crossconformance/ -race -count=2 -run
  'TestDraftSourcesSemanticCases/attestation-evidence-' -v` × 10 consecutive chunks on the rev2 tree,
  200/200 evidence subtests, 2262 s, both reproduction rows 20/20 — command and per-chunk timings
  recorded ✓.
- Mine (candidate clone, `go test -c -race` binary compiled from that clone, run from
  `internal/crossconformance`): `/tmp/2d9gfv-rev2/xconf-race.test -test.count=5 -test.run
  'TestDraftSourcesSemanticCases/attestation-evidence-' -test.v -test.timeout 60m` → **50/50 evidence
  subtests PASS** (5 × 10 rows; `unreadable` 5/5 at 10.2–22.5 s, `wrong-key` 5/5 at 13.0–23.8 s),
  0 FAIL, 0 SKIP, 0 host-stall signatures, 772 s wall at load 8–14 while the registry/install/mutant
  drivers ran concurrently; exit 1 only from the documented `executed(10) != total(94, want 94)` filter
  guard (5 parent FAIL lines).
- Hosted gate on the exact tree: all 10 evidence rows PASS on Race ubuntu (1.9–2.2 s) and Race macos
  (4.2–8.3 s) and on Test ubuntu/macos; Windows Test SKIPs them at Layer 3 (POSIX transport wrapper) as
  before. Ratio lines: macOS lanes `semantic cases: 93 driven, 0 known-gap, 1 bound, 0 skipped, 94
  total`; ubuntu `92 driven … 1 skipped` (pre-existing platform skip, not an evidence row); Windows
  `42 driven … 51 skipped`.

## 6. Harness (R3, note item 5)

`draftInstallConfig` stays a bare literal (zero skew) with the R3 comment ✓; no skew widening ✓; the
crossconformance stub still mints at serve time ✓. `crossSecondBoundary` 50 ms kept with the honest
comment (guards only the hook's own scheduling; the product bound makes the acceptance margin positive by
construction) ✓ — F2 closed. F3: `registryFixture`, `fakeRegistry` and `TestSnapshotZeroClockSkewIsLiteral`
comments now describe the serve-time mint as a legitimate publication covered by the product bound ✓.
1bdotx: `TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch` kept as the mint-once cousin (its
comment is history, still accurate); nothing from 1bdotx is contradictory.

## 7. Reruns and mutants (note item 6)

Candidate clone reruns (`bash` driver, `set -u -o pipefail`, rc per step, clone tree before/after =
`48ad45d6…`, status clean):

| step | command | result |
|---|---|---|
| gofmt / vet / build | `gofmt -l internal/ cmd/` = []; `go vet` on registry, install, crossconformance, cmd/curator; `go build ./...` | clean |
| registry `-race -count=5` | `go test ./internal/registry/ -race -count=5 -run 'TestSnapshotFutureBound\|TestSnapshotZeroClockSkew\|TestSnapshotVerification\|TestSnapshotRequiresCompleteShape\|TestSnapshotRollback\|TestReadOnlySnapshotCheck' -v` | ok 51.2 s, 115 PASS, 0 FAIL |
| registry full | `go test ./internal/registry/ -count=1` | ok 13 s |
| install `-race -count=3` | `go test ./internal/install/ -race -count=3 -run 'TestRegistry\|TestStrictRegistry\|TestDraftEvidenceExactMatch' -v` | ok 98.8 s, 30 PASS (10 tests × 3, incl. the two-registry row 3/3), 0 FAIL |
| legacy goldens | `go test ./cmd/curator/ -count=1 -run 'TestDraftTransportLegacyGolden$\|TestProjectResolveLegacyUntouched$' -v` | ok 24.2 s, 2/2 |
| golangci-lint 2.12.2 | `golangci-lint run ./internal/registry/... ./internal/install/... ./internal/crossconformance/...` | 0 issues |
| rev1 probe on rev2 | see §2 | 3/3 PASS |

Mutants (`internal/registry/snapshot.go`, second clone, python line-replacer with sha check,
`git checkout --` between mutants, tree back at `48ad45d6…` / file sha `ef92a04e…`; rows = registry
`TestSnapshotFutureBound*|TestSnapshotZeroClockSkew|TestSnapshotVerification` + install
`TestRegistrySnapshot*|TestRegistryFuture*|TestRegistryAttestationLandsInMarker`, `-count=1`):

| mutant | shape | killed by | note |
|---|---|---|---|
| M1 | drop the elapsed term (= pre-fix `After(now.Add(clockSkew))`) | unit own-fetch accept + sibling accept; e2e `MintedDuringFetchIsNotFuture` with the gate's exact class `Errors:[every trusted audit registry served a tampered snapshot]` + `registry test-reg snapshot timestamp is too far in the future`; e2e `SlowSiblingDoesNotFlipInstantRegistry` | negatives (`TwoSecondsPast`, `FutureSnapshot(+1h)`, exact-bound refuse rows) still pass — check intact; root cause re-proven |
| M2 | `if false &&` (drop the future check) | exact-bound `one second past` + `an hour past` ×3 skews, literal-zero, unit `now+skew+2s`, sibling `now+skew+3s` ×2 skews; e2e `TwoSecondsPastSkewStillRefuses`, `FutureSnapshotDeniesInstall` | killed |
| M3 | the exact rev1 product file (per-registry `checkStart`) | **only** unit `…/behind the post-fetch clock is accepted` (`map[https://fast:true] [registry fast snapshot timestamp is too far in the future]`) and e2e `SlowSiblingDoesNotFlipInstantRegistry` (`registry reg-b snapshot timestamp is too far in the future`) | the requested discriminator: hoist vs per-registry is now pinned by acceptance rows; every other committed row passes → narrowing |
| M4 | `.Add(time.Second)` on top of the elapsed term | exact-bound `one second past the bound` ×3 skews, literal-zero | install rows pass by design (edge pinned in registry) |
| M5 | elapsed term → constant 2 s | exact-bound past-bound ×3, literal-zero, unit `now+skew+2s`; e2e `TwoSecondsPastSkewStillRefuses` | killed |
| M6 | `start` taken just before the registry loop (after MkdirAll/Chmod/catalog read) | — | **survives** every committed row; bound, see below |
| M7 | `start` in-loop just before the fetch (rev1's M7 survivor) | unit sibling accept; e2e `SlowSiblingDoesNotFlipInstantRegistry` | now killed, as results.md claims |
| M8 | `start` in-loop after the fetch returned | unit own-fetch accept + sibling accept; e2e `MintedDuringFetch`, `SlowSibling` | killed |

M6 bound: the interval M6 leaves uncovered is the pre-loop state work (measured 207–635 µs under
`-race` in the rev1 review; that code is byte-identical in rev2). A serve-time mint can only be refused
by M6 and accepted by the candidate if `created_at = trunc(t_serve) > t_check − δ_preloop`, which needs
the post-serve processing (HTTP read + parse + ed25519 verify, measured 2.2–3.0 ms in-process, more over
a network) to be *smaller* than δ_preloop — impossible without a multi-millisecond stall inside a
directory create / small-file read. No honest row can kill M6 without injecting such a stall, and the
same argument shows the shape has no observable product effect; function-entry is nevertheless the
right placement (a slow catalog read on a stalled disk would otherwise reopen the class). Recorded as a
stated bound, not a defect.

## 8. Ratio line, Windows

Ratio `93 driven, 0 known-gap, 1 bound, 0 skipped, 94 total` (macOS lanes; `registerDraftSemantic`
untouched — no corpus/driver rows added or removed; the 1 bound is the pre-existing capture-mutation
bound) ✓. Windows: no `-race` lane exists (workflow matrix); the Windows Test lane PASSes all three
production-entry rows (`MintedDuringFetchIsNotFuture` 13.1 s, `SlowSiblingDoesNotFlipInstantRegistry`
13.2 s, `TwoSecondsPastSkewStillRefuses` 1.4 s) and the registry pins
(`TestSnapshotFutureBoundToleratesCheckerLatency` 6.8 s, exact-bound 0.76 s, literal-zero 0.10 s);
crossconformance evidence rows SKIP at Layer 3 on Windows as before. Windows race behaviour stays
unverified by construction of the matrix — stated, not claimed.

## 9. Residuals (none blocking)

- R-A `M6` bound (§7): start-before-loop is indistinguishable from function-entry by any serve-time-mint
  row; reasoned + measured.
- R-B The product fix turns the previously clock-free injected-`now` pins into rows with an in-process
  latency dependence: `TestSnapshotZeroClockSkewIsLiteral` needs the pre-check work < 500 ms (whole test
  incl. two checks observed 0.18–0.28 s under `-race` at load 13; 0.10 s on hosted Windows), the
  exact-bound `one second past` pins < 1 s (12 subtests in 1.35–2.08 s), the sibling refuse rows < 1.5 s
  of overhead. The producer declined the optional catalog pre-creation with a measured justification;
  acceptable, but if these ever flake on a hosted lane the pre-created catalog (or a larger offset) is the
  fix — not a skew bump.
- R-C Read-only status path at zero skew with a serve-time mint: bound (shared body), as in rev1.
- R-D No Windows `-race` lane (matrix); Windows Test lane covers the rows functionally.

## 10. Evidence files (reviewer host, `bash` drivers, rc logged per step)

`/tmp/2d9gfv-rev2/logs/`: `driver-cand.out`, `gofmt.log`, `vet.log`, `build.log`,
`registry-race-x5.log`, `registry-full.log`, `install-race-x3.log`, `legacy-goldens.log`, `lint.log`,
`probe-race-x3.log`, `driver-xconf.out`, `xconf-compile.log`, `xconf-evidence-race-x5.attempt1.log`,
`driver-mut.out`, `mut-M{1..8}-{registry,install}.log`; mutator `/tmp/2d9gfv-rev2/mutate.py`; gate
artifacts `/tmp/2d9gfv-rev2/gate/{race,test}-evidence-*/…/go-test.json`. Clones:
`/tmp/2d9gfv-rev2/cand` (tree `48ad45d6…`), `/tmp/2d9gfv-rev2/rev1` (tree `f5c03b9c…`),
`/tmp/2d9gfv-rev2/mut` (restored to `48ad45d6…`). The Story worktree stayed read-only (temp-index tree
`48ad45d6…`, `git status` = the five modified candidate paths, nothing else). Leak hygiene: the one
`curator-conformance-bin*` dir from my run (07:32 local) deleted, nothing else.
