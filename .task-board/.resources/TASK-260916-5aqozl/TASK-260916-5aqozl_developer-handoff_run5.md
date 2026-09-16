# TASK-260916-5aqozl — developer handoff (run 5: Windows store-registry deferral, rev3)

## Starting state

Run 4 handed off rev2 (privatedir output-root fix). CR rev2 validation ran remote gate 35098955988: every job green
except Test (windows-latest), where exactly two cases fail — both in internal/pnpmsource, both with
`closure_input_undeclared: pnpm writable store registry contains an undeclared member`:

- TestRealPinnedPNPMLockSupersetSnapshotDependencies (conformance_test.go:883)
- TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization (conformance_test.go:1136)

The third case (TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall) now PASSES on Windows — run 4's fix
held. Ubuntu/macOS Test and both Race lanes green with all three cases PASSING.

Orchestrator ruling (nudge on RUN-260916-1cdab7 + precondition resource 5aqozl-windows-failure-2.md): the remaining
two failures are a real Windows product gap in the writable-store closure model, OUT of scope for the pin task, filed
as BUG-260916-2f3xbf. This run declares exactly those two cases Windows-deferred in the platform-case ledger naming
the bug, keeps everything else, and hands off rev3.

## This run change (only files touched)

1. `internal/pnpmsource/conformance_test.go` (+26): new test-only helper `skipOnWindowsForStoreRegistryGap(t)` —
   `t.Skip("deferred on windows pending BUG-260916-2f3xbf: pnpm writable store registry contains an undeclared
   member")` on GOOS=windows only, with a comment citing gate run 35098955988, the bug, and the removal condition.
   Called in exactly the two failing tests, immediately after `newConcretePNPMRunner(t)` so a missing/unpinned pnpm
   still skips with the existing host-capability reason (review-brief item 3 preserved on Windows too). The third
   test is untouched and keeps running on all three OSes. No production code, no closure-check weakening.
2. `.github/ci/skip-classes.tsv` (+3): re-added the `deferred delivery stages` section (precedent: commit 4458637,
   which introduced stage-deferred with exactly this three-part shape and removed its row when that stage landed)
   with one row: `stage-deferred / deferred on windows pending BUG-260916-2f3xbf / allow`. The regex is the literal
   bug id — it admits exactly this deferral; any future bug needs its own reviewed row.
3. `.github/ci/platform-cases.tsv` (+22): new commented section + two rows —
   `internal/pnpmsource <case> linux,darwin windows stage-deferred` — requiring both cases on unix (where the pin
   makes them pass) and tolerating the bug-naming skip on windows only. A skip on unix, or for any other reason on
   Windows, stays fatal.
4. `.github/ci/gate-selftest.sh` (+101): 10 new assertions reading the wiring back from the shipped files — reason
   extracted from the Go source (never restated); call-site narrowness (exactly the two Tests); first-match-wins
   classification as stage-deferred; ledger shape per case; behavioural gate drives on otherwise-passing shipped
   streams (windows tolerated + verdicts recorded, same-skip-on-linux fatal, wrong-class-on-windows fatal).

Why stage-deferred and not host-capability/platform-control: the failure is a product gap, not a host inability or a
protocol-defined carve-out; stage-deferred's documented meaning ("belongs to a later delivery stage and is asserted
by that stage's suite") matches, and BUG-260916-2f3xbf's AC ("green on windows-latest with no ledger deferral") is
the removal condition, mirroring how the stage-(b) rows were removed when that stage landed.

## Validation — real commands, real exit codes (shell: bash)

1. `gofmt -l cmd internal` → no output, exit 0.
2. `go vet ./internal/pnpmsource/` → exit 0.
3. `GOOS=windows go vet ./internal/pnpmsource/` → exit 0 (test file compiles for Windows).
4. `go test -count=1 -run TestResolveWindowsPNPMEntrypoint ./internal/pnpmsource/` → ok, exit 0.
5. Rehearsal (PATH=/tmp/pnpm-prefix/bin:$PATH; `node <pnpm> --version` → 10.33.0, exit 0; prefix is the
   run-3 npm-installed layout, release re-verified by probe): `go test -count=1 -run TestRealPinnedPNPM -v` →
   3x PASS (7.17/9.72/11.31s), exit 0. The Windows-only deferral does not affect unix.
6. Skip control (default PATH, pnpm absent): 3x SKIP `pinned pnpm executable unavailable`, exit 0 — the
   pnpm-availability ordering is preserved.
7. `PNPM_PIN=10.33.0 bash .github/ci/pnpm-pin-guard.sh` → agree, exit 0; drift probe `9.9.9-drift-probe` → exit 1.
8. `bash .github/ci/gate-selftest.sh` → 157 passed, 0 failed, exit 0 (147 before + 10 new deferral assertions).
9. `bash .github/ci/ledger-consistency.sh` → 237 rows checked, ok, exit 0 (both new rows listed).
10. Full package with pnpm: `go test -count=1 ./internal/pnpmsource/` → ok 40.1s, exit 0.
11. `golangci-lint run ./internal/pnpmsource/` (2.12.2 = CI pin) → 0 issues, exit 0.
12. Reason classification proof (same first-match logic as the gate): the printed reason first-matches skip-classes
    row 52, class=stage-deferred, policy=allow — no earlier row shadows it.

Not run: full landing suite (handoff-triggered, exactly-once); Windows/rose-air legs (no local runners — hosted
gate proves them). No Go unit test for the helper itself: on non-Windows it is a no-op return, so a local test would
be vacuous; its behaviour is proven by the gate (Windows skip recorded) and the 10 selftest assertions above.
Primary-node hang from run 4 did not reproduce (node --version → v24.21.0, exit 0).

## Checklist basis (all 9 remain ticked)

- Item 2 (hosted gate green, EXECUTED on all OSes): rev2 proved ubuntu+macOS EXECUTED-and-PASS (all three) and
  Windows EXECUTED with 1 pass + 2 product-gap fails, now narrowly deferred per orchestrator ruling with the bug
  filed. Tick = submitted-to-rev3-gate (run3/run4 convention; handoff requires all-checked); void if rev3 red.
  Reviewer: read rev3 test-evidence-* artifacts — expect the two stage-deferred skips in the Windows
  skips-observed.tsv and passes everywhere else.
- Item 3 (rose-air): unchanged — verifiable only post-landing (main-push-only lane); the lane carries the same
  provisioning + guard + (now) deferral wiring.
- Others: unchanged from run 4 + commands 1–12 above; item 8 = this artifact + attached logs.

## Final tree delta (uncommitted, for the CR snapshot)

- M .github/workflows/ci.yml (+132, runs 1–2, untouched this run)
- M .github/ci/gate-selftest.sh (+181 total: +80 runs 1–2, +101 this run)
- M .github/ci/platform-cases.tsv (+22, this run)
- M .github/ci/skip-classes.tsv (+3, this run)
- ?? .github/ci/pnpm-pin-guard.sh (new, runs 1–2, untouched this run)
- M internal/pnpmsource/conformance_test.go (+106/-2 total: +61 run 3, +19/-2 run 4, +26 this run)

## Residual risks

- Windows proof is gate-only (no local Windows runner); the deferral path is reasoned from the exact rev2 error,
  the passing third test (same runner setup), and the behavioural selftest drives.
- If BUG-260916-2f3xbf's fix changes the store layout, the two deferred cases plus the ledger rows must come back
  together — the removal condition is stated in all three places (Go comment, ledger comment, class note).
