# BUG-260920-2d9gfv — results (developer, revision 2)

Nondeterministic refusal class on attestation-evidence rows under `-race`:
`attestation-evidence-unreadable` (gate 35486280770, Race ubuntu) and
`attestation-evidence-wrong-key` (gate 35529945079, Race macos) returned
`every trusted audit registry served a tampered snapshot` instead of
`is not audited by any trusted registry`.

Revision 2 reworks revision 1 per `2d9gfv-rework-1.md` (review verdict
CHANGES_REQUESTED, RUN-260921-976df9): the rev1 per-registry `checkStart`
bound still let the refusal class depend on fetch duration with more than
one trusted registry (reviewer's driven two-registry proof, 3/3 fail on
rev1 / 3/3 pass on function-entry). This revision adopts the brief's
preferred function-entry shape. Root cause (confirmed by reviewer) and the
single-registry rows are unchanged; §"Revision 2" below lists exactly what
moved.

## Root cause (product defect, confirmed at the production entry)

`resolveRegistries` (`internal/install/install.go:1346-1356`) samples
`time.Now()` **before** the snapshot fetch and passes it as `now` to
`registry.CheckSnapshotsWithPolicy`. The registry mints `created_at` at
serve time, truncated to whole seconds
(`internal/crossconformance/draftsources_semantic_evidence_test.go:137`;
same shape in `internal/install/draftevidence_test.go:75`). Whenever a
whole-second boundary falls between the two clock reads, `created_at > now`
by up to 1 s, and with the zero clock skew of the bare Go-API harness
config (`draftInstallConfig`, a literal `0`, not the 300 s loader default)
`parsed.CreatedAt.After(now.Add(clockSkew))`
(`internal/registry/snapshot.go:158` pre-fix) classifies the registry as
tampered ("snapshot timestamp is too far in the future"). The flip window
is the fetch latency — milliseconds on Test lanes, tens–hundreds of ms
under `-race` — matching the lane pattern and spanning every evidence row
that runs Layer 1 with a serve-time minter.

Note: BUG-260906-1bdotx fixed this same shape in the install-package
fixture only (mint-once `createdAt`), leaving the product defect live for
serve-time minters. This fix is at the product root.

Driven pre-fix proof (rev1, accepted by review §1): new
`TestRegistrySnapshotMintedDuringFetchIsNotFuture` (stub signs strictly
past the next second boundary, zero-skew config) **failed** on the unfixed
tree with exactly the gate class
(`Errors:[every trusted audit registry served a tampered snapshot]`,
`Messages:[... registry test-reg snapshot timestamp is too far in the
future]`) and passes with the fix. The reviewer re-ran this proof on a
disposable clone (mutant M1) and confirmed the gate's exact class+text.
The negative companion
(`TestRegistrySnapshotTwoSecondsPastSkewStillRefuses`) passed pre- and
post-fix.

## Fix (revision 2: function-entry shape)

`internal/registry/snapshot.go` (`checkSnapshotsWithPolicy`, shared by the
draft lane, the legacy v1 lane, and the read-only status path): capture
`start := time.Now()` once as the first statement of the function and
evaluate the future check as

```go
parsed.CreatedAt.After(now.Add(clockSkew).Add(time.Since(start)))
```

- Exported signatures and `now` semantics unchanged: `now` stays the
  check's reference time. Both production call sites pass `time.Now()`
  (`install.go:1347`, `:1353`), so for production callers the bound is
  the post-fetch wall clock plus the configured skew; fixed-`now`
  callers with an instant fetch see a ~zero tolerance, so the
  exact-bound pins still hold.
- Function-entry start, not per registry: the tolerance is the checker's
  own latency since it sampled its clock. A slow sibling does not widen
  anything beyond the specification — the wall clock simply advanced —
  and any snapshot this bound accepts has
  `created_at <= wall clock + skew`. Pinned by the inverted sibling unit
  row and the two-registry production-entry row (both fail on the rev1
  shape; see mutants).
- Stale check unchanged: a pre-fetch reference can only make a snapshot
  look *less* stale, never falsely stale, so fetch duration cannot flip
  that class.

## Rulings (from 2d9gfv-brief.md + 2d9gfv-rework-1.md)

- R1 — product fix at the root; future check independent of fetch
  duration; genuinely future timestamps still refused with the same
  class and text; stale semantics untouched. Done as above (brief's
  preferred shape per rework-1).
- R2 — legacy v1 lane shares `CheckSnapshotsWithPolicy*`; declared bug
  fix with CHANGELOG `Unreleased → Fixed` entry ("its own latency
  since it sampled its clock" wording). Legacy-lane rows: the three new
  e2e tests run through the frozen v1 lane (schema-1 Skillfile,
  `install.Project → resolveRegistries → CheckSnapshotsWithPolicy`,
  zero skew, on-demand-signed snapshot; real future timestamp still
  refuses). `TestDraftTransportLegacyGolden` and
  `TestProjectResolveLegacyUntouched` green.
- R3 — harness keeps the zero-skew Go-API config as the strict case,
  now documented in a `draftInstallConfig` comment; no skew bump, no
  fixed `created_at` in the stub. Determinism: 20 consecutive `-race`
  iterations of all evidence rows (see below).
- R4 — deterministic production-entry rows (single- and two-registry;
  pass only with the fix); negative row (serve clock + 2 s, ≥ 1 s past
  now + skew at skew 0); narrowing mutants; unit rows in
  `internal/registry` with injected `now`. Done; see below.

One deliberate test-shape note (carried from rev1): the e2e negative
stamps serve-relative (`snapshotFutureBy(2s)`) rather than fixture-time
`now + 2s`, because slow setup under parallel `-race` slid a
fixture-anchored stamp into the past (a test-harness flaw, not a product
flaw). The exact R4 shape `created_at = now + skew + 2s` is pinned at
the unit level with an injected `now`, where setup cannot intervene.
The `crossSecondBoundary` hook keeps its 50 ms margin with a corrected
comment: it now only guards the hook's own scheduling — the product
bound covers the checker's latency, so the acceptance margin is
positive by construction (review F2).

## Revision 2 (what moved vs revision 1)

Per rework-1 §7, continued from the rev1 tree (no checkout/clean/stash):

1. `internal/registry/snapshot.go`: `start := time.Now()` as the first
   statement of `checkSnapshotsWithPolicy`; comparison
   `parsed.CreatedAt.After(now.Add(clockSkew).Add(time.Since(start)))`;
   both comments rewritten (tolerance = checker's own latency since it
   sampled its clock; genuinely future = ahead of the post-fetch clock
   + skew, still refused).
2. `internal/registry/registry_test.go`: `a slow sibling does not widen
   another registry's bound` replaced by its inverse `a slow sibling's
   latency is inside the bound` — after a 1.5 s sibling, an instant
   registry with `created_at = now + 1 s` (behind the wall clock) is
   ACCEPTED, and ones with `created_at = now + skew + 3 s` (ahead of
   the post-fetch clock, skews 0 and 30 s) are still refused with `too
   far in the future`. The test function is renamed to
   `TestSnapshotFutureBoundToleratesCheckerLatency` (old name pinned
   the rev1 semantics). The other two rows are kept byte-identical.
   F3 comment on `TestSnapshotZeroClockSkewIsLiteral` refreshed (zero
   = zero *configured* tolerance; the product's own-latency tolerance
   additionally applies). The optional catalog pre-creation was NOT
   taken: first-use fsyncs measure milliseconds against 500 ms / 1 s
   margins (exact-bound pins green 3/3 under `-race` on this tree).
3. `internal/install/registry_e2e_test.go`: new two-registry
   production-entry row
   `TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry` (slow
   first registry crossing the boundary with `pending` status, instant
   serve-time-minted second registry with `audited` status, strict
   policy, zero skew → `ok`, no future warning). F2: 50 ms kept with a
   corrected comment. F3: `registryFixture` and `fakeRegistry`
   comments refreshed (serve-time mint is a legitimate
   publication-during-fetch covered by the product bound).
4. CHANGELOG: "per-registry latency" → "its own latency since it
   sampled its clock"; first line "its own fetch" → "a fetch".
5. R3 determinism re-run on the reworked tree (below).

## Evidence (all on the candidate tree, macOS host, shell `bash`, go1.26.0 darwin/amd64)

Determinism — evidence rows, `-race`, 20 consecutive iterations in ten
bounded `-count=2` chunks (chunking is consecutive: no code changed
between chunks; each chunk fits the ~10 min single-shell bound where a
single `-count=20` does not):

`go test ./internal/crossconformance/ -race -count=2 -run
'TestDraftSourcesSemanticCases/attestation-evidence-' -v`

| chunk | wall time | evidence subtests PASS | subtest FAIL/SKIP |
|---|---|---|---|
| 1 (iters 1–2) | 243 s | 20/20 | 0 |
| 2 (iters 3–4) | 267 s | 20/20 | 0 |
| 3 (iters 5–6) | 232 s | 20/20 | 0 |
| 4 (iters 7–8) | 270 s | 20/20 | 0 |
| 5 (iters 9–10) | 205 s | 20/20 | 0 |
| 6 (iters 11–12) | 195 s | 20/20 | 0 |
| 7 (iters 13–14) | 218 s | 20/20 | 0 |
| 8 (iters 15–16) | 229 s | 20/20 | 0 |
| 9 (iters 17–18) | 198 s | 20/20 | 0 |
| 10 (iters 19–20) | 205 s | 20/20 | 0 |
| total | 2262 s (~38 min) | **200/200** | 0 |

Each chunk exits 1 solely from the harness's documented filter tally
(`executed(10) != total(94, want 94); run the full matrix without -run
subtest filters`, 2/2 per chunk; the only `--- FAIL` lines are the two
parent `TestDraftSourcesSemanticCases` tallies); zero evidence subtests
failed or skipped in any iteration. Both reproduction rows
(`attestation-evidence-unreadable`, `attestation-evidence-wrong-key`)
passed 2/2 in every chunk — 20/20 each. Full `-v` logs retained on the
producer host (`xconf-rev2-chunk1..10.log`); counts above were
re-verified with anchored greps
(`--- PASS: TestDraftSourcesSemanticCases/attestation-evidence-` =
20/chunk, non-parent `--- FAIL` = 0/chunk, `--- SKIP` = 0/chunk).

Production-entry rows under `-race` on this tree:

- `go test ./internal/install/ -race -count=3 -run
  'TestRegistrySnapshotMintedDuringFetchIsNotFuture|TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry|TestRegistrySnapshotTwoSecondsPastSkewStillRefuses' -v`
  → ok 26.6 s, 9/9 PASS (3 tests × 3).
- `go test ./internal/registry/ -race -count=3 -run
  'TestSnapshotFutureBoundToleratesCheckerLatency|TestSnapshotFutureBoundIsExactAtEveryConfiguredSkew|TestSnapshotZeroClockSkewIsLiteral'`
  → ok 25.8 s.

Two-registry row output (fail-on-rev1 proof, same tree, product file
temporarily at the rev1 per-registry shape, restored after — sha256
`ef92a04e…` before and after):

- rev1 shape:
  `TestRegistrySnapshotSlowSiblingDoesNotFlipInstantRegistry` FAILS —
  `a snapshot at or behind the post-fetch clock was refused as future:
  ... registry reg-b snapshot timestamp is too far in the future`
  (1.5 s). The unit sibling-accept
  (`.../a slow sibling's latency is inside the bound/behind the
  post-fetch clock is accepted`) FAILS identically
  (`map[https://fast:true] [registry fast snapshot timestamp is too
  far in the future]`); all other unit rows pass — the mutant is
  narrowing.
- rev2 shape: both PASS (install 6/6 incl. the pre-existing boundary,
  future-refusal, and within-bound rows, 14.0 s non-race run).

Registry package full: `go test ./internal/registry/ -count=1` → ok
10.5 s. Legacy goldens `TestDraftTransportLegacyGolden`,
`TestProjectResolveLegacyUntouched` (`./cmd/curator/`) → both PASS,
25.5 s. Static gates: `go vet` (4 touched packages) clean; `gofmt -l`
clean; `golangci-lint run` on the three touched `internal/` packages
→ 0 issues; `go build ./...` → ok.

Full-matrix ratio re-run on rev2: NOT re-run here — a single unfiltered
`-race` matrix (~6.5 min, rev1) exceeds the bounded-shell budget and a
non-race attempt was terminated after backgrounding per headless
constraints. The ratio is accepted from rev1's attached evidence
(unfiltered `-race` 95/95, `93 driven, 0 known-gap, 1 bound, 0 skipped,
94 total`) plus the verified fact that `registerDraftSemantic` is
untouched by this diff (0 matches), so no corpus row was added or
removed. The hosted gate re-proves the ratio on the exact candidate
tree.

## Mutant table (`internal/registry/snapshot.go`, rev2 tree)

| mutant | shape | expected kill | observed |
|---|---|---|---|
| M1 | drop the elapsed term (`After(now.Add(clockSkew))`) | deterministic rows fail | e2e `MintedDuringFetch` FAIL, e2e `SlowSibling` FAIL, unit accept rows FAIL (slow-fetch + sibling-accept); e2e `TwoSecondsPast` and unit refuse rows still pass (check intact) — killed |
| M2 | drop the future check (`if false && …`) | negative rows fail | e2e `TwoSecondsPast` FAIL, e2e `FutureSnapshot(+1h)` FAIL, unit `now+skew+2s` FAIL (+ exact-bound past-bound and literal-zero pins, same as rev1) — killed |
| M-rev1 | scope the start per registry (the rev1 shape: `checkStart` in-loop, `time.Since(checkStart)`) | sibling rows fail, nothing else | e2e `SlowSibling` FAIL (reg-b `too far in the future`), unit sibling-accept FAIL; every other committed row passes — killed. This is the inverse of rev1's M3 (which hoisted to function-entry and killed only the old sibling-refusal row): the row that dies is now the acceptance pin |
| (bound) | threshold-widening (`+1 s` on the bound, 2 s constant) | exact-bound pins fail | accepted from rev1 review (M4/M5 killed by exact-bound `one second past the bound` ×3 skews, literal-zero, unit `now+skew+2s`); the pins are byte-identical on this tree and green 3/3 under `-race` |

No survivors. The M7 class from rev1 (start taken just before the fetch,
state read excluded — survived every rev1 committed row) is now killed
by the committed two-registry row: under M7 the second registry's
tolerance excludes the sibling's fetch, so the boundary-crossing sibling
flips it exactly as M-rev1 does.

Read-only-path bound (review §2): the fix is in the shared body
(`CheckSnapshotsWithPolicyReadOnly → checkSnapshotsWithPolicy`), so it
applies by construction; no committed row drives the read-only path at
zero skew with a serve-time mint (Layer 3 of the harness uses the
loader-default 300 s skew; `registry_test.go` read-only rows exercise
state semantics). Accepted as a bound, as the reviewer allowed.

## Files changed

- `internal/registry/snapshot.go` — the fix (function-entry elapsed
  term + comments). Only non-test product file touched.
- `internal/registry/registry_test.go` — unit rows: slow-fetch accept
  (kept), slow-sibling inverse accept + `now+skew+3s` refuse (new),
  `now+skew+2s` refuse (kept); F3 literal-zero comment refresh.
- `internal/install/registry_e2e_test.go` — v1-lane production-entry
  deterministic rows (single + two-registry) + tight negative;
  `snapshotMintedAtServeTime` / `snapshotFutureBy` fixture options;
  F2/F3 comment refresh.
- `internal/crossconformance/draftsources_fixtures_test.go` — strict
  zero-skew config comment (R3, carried from rev1).
- `CHANGELOG.md` — `Unreleased → Fixed` entry (R2, function-entry
  wording).

Non-test diff = `internal/registry/snapshot.go` + CHANGELOG + one
harness comment — same footprint as rev1.

## Ratio line

`semantic cases: 93 driven, 0 known-gap, 1 bound, 0 skipped, 94 total`
(`registerDraftSemantic` untouched — no corpus or driver rows
added/removed; the 1 bound is the pre-existing `capture-mutation`
bound, proven in-package). Accepted from rev1's attached full-matrix
evidence on the same driver; hosted gate re-proves on this tree.

## Windows proof status

Unverified locally (host is macOS). No platform-specific code touched
(`time`/`time.Since` only). CI runs no `-race` lane on Windows (Go race
detector needs a C toolchain; race matrix is ubuntu+macos per
`.github/workflows/ci.yml`); the Windows Test lane covers the
production-entry rows functionally (rev1: both e2e rows PASSed there).
Deferred to the hosted gate on the exact candidate tree.

## Reran-vs-accepted (headless compliance)

- Reran on the exact rev2 tree: R3 20× `-race` evidence chunks (10/10,
  200/200); install new rows `-race -count=3` (9/9); registry bound
  rows `-race -count=3`; registry full; legacy goldens; vet/gofmt/
  build/lint; mutants M1/M2/M-rev1 with fail/pass pairs; rev1-shape
  fail proof for the two-registry row.
- Accepted from already-attached evidence: pre-fix single-registry
  proof (rev1 producer + reviewer M1 rerun, verdict §1); M4/M5
  threshold-widening kills (byte-identical pins, green here);
  full-matrix ratio (same driver, untouched); Windows lane behavior
  (rev1 gate 35545031567).
- Terminated, not accepted: one `-count=4` evidence attempt (exceeded
  the shell yield; replaced by the ten `-count=2` chunks above) and
  one unfiltered non-race matrix attempt (backgrounded; replaced by
  the driver-untouched argument + gate).
