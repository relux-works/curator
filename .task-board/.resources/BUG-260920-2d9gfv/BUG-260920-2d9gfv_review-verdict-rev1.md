# BUG-260920-2d9gfv — review verdict, revision 1 (reviewer, claude-opus-5)

**Verdict: CHANGES_REQUESTED → `to-dev`.** Change Request `CR-BUG-260920-2d9gfv-1`,
base `e0f52c95`, candidate tree `f5c03b9c…` (Story worktree temp-index `write-tree` =
`f5c03b9c690a1d158d2932b9e5bf6408d0bd0a89`; disposable clone base+rev1 patch, sha256
`c0bbb7fc…`, `HEAD^{tree}` = the same OID; hosted gate 35545031567 head `d9318e6f` → tree
`f5c03b9c…`, all 11 served jobs green).

The root cause is confirmed, the named reproduction rows are green and effectively
deterministic, and every ruling but one is met. The one open item is R1 itself: the
candidate's bound is not "a clock reading taken after the snapshot fetch" — it is the
pre-fetch `now` plus only the current registry's own latency, so the refusal class still
depends on fetch duration (the siblings' and the pre-loop state work) whenever more than one
trusted registry is configured. I drove that residual deterministically at the production
entry (`install.Project → resolveRegistries → CheckSnapshotsWithPolicy`, zero skew, two
registries): the instant second registry is refused as `too far in the future` and a strict
install fails `is not audited by any trusted registry` — 3/3 under the candidate, 3/3 pass
under the brief's preferred function-entry shape. The "slow sibling" unit row that pins the
per-registry choice demands refusal of a timestamp that is already **1.5 s behind the wall
clock** when it is checked, so it pins the residual, not a security property; and the
harness margin widening (`crossSecondBoundary` 2 ms → 50 ms) exists only to absorb the part
of the latency the product leaves uncovered. The rework is small (§7).

Against the AC literally: `attestation-evidence-unreadable` (and `-wrong-key`) are
deterministic under `-race` — producer 20/20, mine 5/5, both hosted race lanes — because with
the harness's single registry the uncovered gap is only the pre-loop state work (0.2–0.6 ms
measured) and a flip additionally needs that gap to exceed the post-serve processing
(2–3 ms measured), so it is effectively closed; the second clause ("the classification is
fixed with a production-entry test and a narrowing mutant") is met for one registry and not
for two or more, where the class re-opens at (sibling latency)/1 s per install — the
original rate, on real networks.

## 1. Root cause (review-note item 1) — CONFIRMED

- Pre-fix proof rerun: reverting the one line (`M1`, `After(now.Add(clockSkew))`) in a
  disposable clone makes `TestRegistrySnapshotMintedDuringFetchIsNotFuture` fail with exactly
  the gate's class and text:
  `Errors:[every trusted audit registry served a tampered snapshot]`,
  `Messages:[… test: registry: registry test-reg snapshot timestamp is too far in the future]`
  (`logs/mut-M1-install.log:25`). The negative row
  `TestRegistrySnapshotTwoSecondsPastSkewStillRefuses` passes under M1 (check intact) and
  under the candidate (same class + text asserted: `every trusted audit registry served a
  tampered snapshot` and `registry test-reg snapshot timestamp is too far in the future`).
- The driven row forces the crossing, not sleep-and-hope: `crossSecondBoundary` runs inside
  the snapshot handler, after `now` was sampled, and sleeps until the next whole second +
  50 ms before `time.Now().UTC().Truncate(time.Second)` is stamped, so `created_at` is
  always > `now` (by 0.25–0.96 s in my instrumented runs) and always ≤ the post-serve clock.
- Mechanism verified with an instrumented copy (env-gated stderr, mutant clone only): the
  gap `now → checkStart` (MkdirAll + Chmod + catalog read + decode) measured 207–635 µs under
  `-race` on this host; `checkStart → check` for an instant fetch 2.2–3.0 ms; for the driven
  row 0.31–1.03 s with `created_at − now` = elapsed − 52…75 ms (= the 50 ms hook margin plus
  post-serve processing minus the pre-loop gap).

## 2. Semantics (item 2) — the open finding

**F1 (R1, product): the bound is `now + skew + (t_check − t_checkStart_k)`, and
`t_checkStart_k − now` is not covered.** For registry k that gap is the pre-loop state work
plus every earlier registry's complete check (its HTTP fetch up to the 30 s
`RegistryGetTotalDeadline`, parse, verify, and its `writeSnapshotState` + catalog fsyncs).
A snapshot minted at serve time is therefore still refused as "too far in the future" under
a literal zero skew whenever a whole-second boundary falls inside that gap — the identical
mechanism as the bug, moved from "own fetch" to "siblings' fetches". Driven proof (throwaway
`TestZZReviewSlowSiblingFlipsInstantRegistryUnderZeroSkew`, `internal/install`, production
entry `install.Project`, v1 lane, `registry_policy=strict`, `SnapshotClockSkewSeconds=0`,
two `fakeRegistry` servers: `reg-a` first, status `pending`, `beforeSnapshotResponse(crossSecondBoundary)`;
`reg-b` second, status `audited`, `snapshotMintedAtServeTime()`):

| tree | result (3 runs each) |
|---|---|
| candidate `f5c03b9c` | 3/3 FAIL: `status=failed errors=[skill-a is not audited by any trusted registry (registry_policy is strict)] messages=[… registry reg-b snapshot timestamp is too far in the future]` |
| candidate + function-entry start (`start := time.Now()` at `checkSnapshotsWithPolicy` entry, `time.Since(start)`) | 3/3 PASS (`status=ok`, no future warning) |

Every timestamp in that probe is at or behind the client's wall clock when it is checked, so
the brief's criterion ("a timestamp genuinely ahead of the client's clock after the fetch
must still be refused" — and, by the same token, one that is not ahead must not be) is
violated by the candidate and met by the preferred shape.

**Why the producer's rationale does not hold.** `now` is `time.Now()` at both production
call sites (`install.go:1349`, `:1355`), so with a function-entry start the bound is
`≈ wall clock at check + skew` — exactly the specified bound — and it can never exceed
`wall clock + skew` (`now ≤ t_entry`). A slow sibling does not "widen" anything beyond the
specification; the wall clock simply advanced. The per-registry bound is *stricter than
specified* by the uncovered gap and buys no security: any snapshot it refuses that the
function-entry bound accepts has `created_at ≤ wall clock + skew`. The candidate's own pin
shows this: in `TestSnapshotFutureBoundToleratesOnlyItsOwnFetchLatency/a slow sibling …`
the fast registry's `created_at = now + 1 s` is fetched ≥ 1.5 s after `now`; instrumented
(`injected now=…:15Z, fast created_at=…:16Z, wall clock when fast was fetched=…:17.527Z ⇒
created_at is 1.527 s BEHIND the wall clock; verdict: tampered=true, "registry fast snapshot
timestamp is too far in the future"`). The row asserts that a past timestamp is refused as
future. Hoisting the start (`M3`) fails exactly that row and nothing else — verified on a third
clone (`fix2` = candidate + the 2-line hoist + my probe): install rows + probe `-race
-count=3` 33/33 PASS; registry rows `-race -count=3` all PASS except the sibling row 3/3
FAIL; crossconformance evidence rows `-race` (binary compiled from `fix2`) 10/10 PASS —
i.e. the row discriminates per-registry from function-entry, as the note asked, and what it
discriminates is the residual defect.

Stale check: unchanged (`snapshot.go:164`), correct per R1. Read-only status path: the fix
is in the shared body (`CheckSnapshotsWithPolicyReadOnly → checkSnapshotsWithPolicy`), so it
applies by construction; no committed row drives it at zero skew with a serve-time mint
(Layer 3 of the harness uses the loader-default 300 s skew; `registry_test.go:551/576`
exercise read-only *state* semantics) — a bound, acceptable as such, note it in results.md.

**F2 (R3/R4, harness): `crossSecondBoundary` 2 ms → 50 ms is compensation for F1.** With
the per-registry start the driven row's acceptance margin is
`hook margin + post-serve processing − (t_checkStart − now)`; the comment says so ("keeps the
serve-time mint comfortably ahead of the checker's pre-fetch work"). With a function-entry
start the pre-fetch work is *inside* the tolerated interval and the margin is
`hook margin + post-serve processing + (t_entry − now)` — positive by construction even at
0 ms. Under the rework the 50 ms is harmless but its justification is gone; restore 2 ms or
keep 50 ms with an honest comment (it then only guards the hook's own scheduling, not the
product).

**F3 (comments now contradicted by the candidate, minor):** `registryFixture` doc
("A published snapshot is minted once and then served; minting created_at inside the handler
… is the defect this file carried"), `fakeRegistry` ("Mint the snapshot timestamp once, here,
and never again: the served snapshot must not move relative to the clock reading the checker
takes before it fetches") and `TestSnapshotZeroClockSkewIsLiteral` ("The install fixtures
depend on this: they stamp their snapshot before the run rather than during the fetch,
because under this policy any drift forward is tampering") describe the pre-2d9gfv contract;
`snapshotMintedAtServeTime` and the product tolerance are the opposite. Refresh them.
`TestRegistrySnapshotSurvivesASecondBoundaryDuringFetch` (1bdotx) is not contradictory —
it is now the weaker mint-once cousin of the new row — keep it.

## 3. Legacy v1 lane (item 3) — MET
CHANGELOG `## Unreleased → ### Fixed` (lines 139–151) ✓. Legacy-lane rows: both new e2e rows
run the frozen v1 lane (schema-1 Skillfile, `install.Project` without `DraftSourcesV1`,
zero skew, serve-time-signed snapshot; negative row refuses) ✓. Legacy goldens
`TestDraftTransportLegacyGolden`, `TestProjectResolveLegacyUntouched` PASS (my rerun,
22.6 s) ✓. Non-test diff = `internal/registry/snapshot.go` only (+CHANGELOG, +one comment in
the harness) ✓. Under the rework the CHANGELOG wording "per-registry latency" must change to
"its own latency since it sampled its clock".

## 4. Determinism (item 4)
Producer: 5 × `go test ./internal/crossconformance/ -race -count=4 -run
'TestDraftSourcesSemanticCases/attestation-evidence-' -v`, 200/200 subtests, ~35 min,
command and per-chunk wall times recorded ✓; full unfiltered `-race` matrix 95/95, ratio
`93 driven, 0 known-gap, 1 bound, 0 skipped, 94 total` ✓.
Mine (candidate clone, precompiled `go test -c -race`, `-test.count=5`, health-gated
rerun after an exec-stall window 01:04–01:20Z that killed Layer-3 CLI execs
(`bootstrap = -1`, `attested status = -1`, `golangci-lint` rc=137 — host artefacts, logged
in `xconf-evidence-race-x5-attempt1-stallwindow.log`)): **50/50 evidence subtests PASS** (5 × 10 rows, both reproduction rows 5/5), 642 s wall at load 7–11, exit 1 only from the documented `executed(10) != total(94)` filter guard — command: `cd internal/crossconformance && /tmp/2d9gfv-rev/xconf-race.test -test.count=5 -test.run 'TestDraftSourcesSemanticCases/attestation-evidence-' -test.v -test.timeout 60m`.
Hosted gate 35545031567 (exact candidate tree): all 10 evidence rows PASS on Race ubuntu
(2.9–7.5 s) and Race macos (2.6–4.1 s), Test ubuntu/macos PASS, Windows SKIP at Layer 3 as
before; ratio line `93 driven … 94 total` on the macOS lanes, `92 driven … 1 skipped` on
ubuntu (a pre-existing platform skip, not an evidence row).

## 5. Harness (item 5) — MET
`draftInstallConfig` stays a bare literal (zero skew) with the R3 comment ✓; no skew widening
✓; the crossconformance stub still mints at serve time ✓. 1bdotx: see F3 (comments stale,
row not redundant).

## 6. Reruns, mutants, ratio, Windows (item 6)

Candidate clone reruns (`bash` driver, rc per step, clone tree before/after = `f5c03b9c…`):

| step | command | result |
|---|---|---|
| registry `-race -count=5` | `go test ./internal/registry/ -race -count=5 -run 'TestSnapshotFutureBound\|TestSnapshotZeroClockSkew\|TestSnapshotVerification\|TestSnapshotRequiresCompleteShape\|TestSnapshotRollback' -v` | ok 34.8 s, 105 PASS, 0 FAIL |
| install `-race -count=5` | `go test ./internal/install/ -race -count=5 -run 'TestRegistry\|TestStrictRegistry\|TestDraftEvidenceExactMatch' -v` | ok 164.6 s, 60 PASS (12 tests × 5), 0 FAIL |
| registry full | `go test ./internal/registry/ -count=1` | ok 7.8 s |
| gofmt / vet / build | `gofmt -l internal/ cmd/` = 0 files; `go vet` 4 packages; `go build ./...` | clean |
| legacy goldens | `go test ./cmd/curator/ -count=1 -run 'TestDraftTransportLegacyGolden$\|TestProjectResolveLegacyUntouched$' -v` | ok 22.6 s, 2/2 PASS |
| golangci-lint | `golangci-lint run ./internal/registry/... ./internal/install/... ./internal/crossconformance/...` | 0 issues (7 s; first attempt rc=137 inside the stall window, rerun clean) |

Mutants (`internal/registry/snapshot.go`, second clone, `git checkout --` between mutants,
tracked tree back at `f5c03b9c…`; rows = registry `TestSnapshotFutureBound*|ZeroClockSkew|Verification`
+ install `TestRegistrySnapshot*|TestRegistryFuture*|AttestationLandsInMarker` + my probe):

| mutant | shape | killed by | survivors of note |
|---|---|---|---|
| M1 | drop the elapsed term (= pre-fix) | e2e `MintedDuringFetchIsNotFuture` (gate class+text), unit accept rows 2/2, probe | negatives pass (check intact) — killed |
| M2 | `if false &&` (drop the future check) | unit `now+skew+2s`, exact-bound past-bound 6/6, literal-zero, e2e `TwoSecondsPast`, `FutureSnapshot(+1h)` | killed |
| M3 | hoist start to function entry (R1 preferred shape) | only unit `a slow sibling …` | everything else passes, probe passes — **this is the requested change, and the row that dies is F1's pin** |
| M4 | `.Add(time.Second)` on top of the elapsed term | exact-bound `one second past the bound` ×3 skews, literal-zero, slow-sibling | killed |
| M5 | replace elapsed by a 2 s constant | exact-bound past-bound ×3, literal-zero, unit `now+skew+2s`, slow-sibling | killed |
| M6 | start taken after the fetch returned | e2e `MintedDuringFetchIsNotFuture`, unit accept rows 2/2, probe | negatives pass — killed |
| M7 | start taken just before the fetch (state read excluded) | **only my two-registry probe** | survives every committed row (registry rc=0, all committed install rows PASS): the committed rows cannot tell whether the pre-fetch state read is inside the tolerance — the two-registry row in §7.3 is what resolves this class |

Ratio line unchanged (`registerDraftSemantic` untouched; 93 driven / 1 bound / 94 on the
macOS lanes) ✓. Windows: the two new production-entry rows PASS on the Windows Test lane
(`TestRegistrySnapshotMintedDuringFetchIsNotFuture` 22.3 s,
`TestRegistrySnapshotTwoSecondsPastSkewStillRefuses` 1.1 s); crossconformance evidence rows
SKIP at Layer 3 on Windows as before; no `-race` lane on Windows (workflow matrix) ✓.

## 7. Requested changes (revision 2)

1. `internal/registry/snapshot.go`: take the monotonic start once at function entry
   (`start := time.Now()` as the first statement of `checkSnapshotsWithPolicy`) and compare
   against `now.Add(clockSkew).Add(time.Since(start))` — the brief's preferred shape. Rewrite
   the two comments: the tolerance is the checker's own latency since it sampled its clock,
   which makes the bound ≈ the post-fetch wall clock + skew for the production callers and a
   ~zero tolerance for injected-`now` callers with instant fetches; a genuinely future
   timestamp (ahead of the post-fetch clock + skew) is still refused.
2. `internal/registry/registry_test.go`: replace `a slow sibling does not widen another
   registry's bound` by its inverse — after a 1.5 s sibling, an instant registry whose
   `created_at = now + 1 s` (behind the wall clock) is **accepted**, and one whose
   `created_at = now + skew + 3 s` (ahead of the post-fetch clock) is still refused with
   `too far in the future`. Keep the other two rows. Optional hardening: pre-create the
   catalog in the exact-bound pins so the first-use catalog fsyncs do not sit inside their
   1 s / 500 ms margins (under function-entry they are inside the elapsed term).
3. `internal/install/registry_e2e_test.go`: add the two-registry production-entry row (shape
   of my probe above: slow first registry crossing the boundary, instant serve-time-minted
   second registry, strict policy, zero skew → `ok`, no future warning); it fails on the
   candidate and passes with (1). Restore `crossSecondBoundary` to 2 ms or keep 50 ms with a
   corrected comment (F2). Refresh the three stale comments (F3).
4. CHANGELOG entry: "per-registry latency" → the function-entry wording; results.md: add the
   read-only-path bound (§2) and the mutant that now kills the sibling row.
5. Rerun the R3 determinism command on the reworked tree and publish only on a green gate.

## 8. Evidence files (reviewer host, `bash` drivers, `set -u`, rc logged per step)
`/tmp/2d9gfv-rev/logs/` — `driver-cand.out`, `registry-race-x5.log`, `install-race-x5.log`,
`registry-full.log`, `legacy-goldens.log`, `lint.log`, `xconf-evidence-race-x5.log`
(+ `…-attempt1-stallwindow.log`), `mut-M{1..7}-{registry,install}.log`, `driver-mut*.out`;
probe `/tmp/2d9gfv-rev/zz_review_probe_test.go`, also attached as
`BUG-260920-2d9gfv_review-rev1-two-registry-probe_test.go.txt` (never written to the Story
worktree; the worktree stayed read-only — `git status` unchanged, temp-index tree
`f5c03b9c…`). Leak hygiene: the three `curator-conformance-bin*` dirs from my runs were
deleted, nothing else.
