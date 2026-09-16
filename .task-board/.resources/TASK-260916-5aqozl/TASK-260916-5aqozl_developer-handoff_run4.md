# TASK-260916-5aqozl — developer handoff (run 4: Windows output-root privacy fix)

## Starting state
Run 3 handed off rev1 (Decision 2 Option B). CR rev1 validation ran remote gate 35091265197: ubuntu/macos Test green with the three TestRealPinnedPNPM* PASSING each; windows-latest Test FAILED — the three tests EXECUTED but failed with `closure_input_undeclared: portable output root is not a private real directory` (conformance_test.go:854/876/1129, the DerivePrivateStore calls). Run 3 ticked item 2 as submitted-to-gate/void-if-red; it went red. Recovery queued this run (RUN-260916-3a9ba3) to fix and re-complete. An orchestrator nudge mid-run pointed at the same failure and a new precondition resource 5aqozl-windows-failure.md (read; reconciled below).

## Diagnosis (evidence chain, all from run 35091265197 artifacts)
- Only failing job: Test (windows-latest), step `go test + platform-case gate`; platform-case gate itself exit 0; every other package green.
- test-evidence-windows-latest/test/go-test.json: exactly 4 fail events — the 3 TestRealPinnedPNPM* + their package; every fail line the same `portable output root is not a private real directory`.
- ubuntu + macos artifacts: the same 3 tests Action=pass, zero fail events. No TestRealPinnedPNPM* rows in the Windows skips-observed.tsv (they executed).
- Root cause: portable_runner.go ensureEmptyDirectory privacy-validates OutputRoot before every Run. The pnpm harness PRE-CREATED executionRoot/output via os.MkdirAll(0o700) and RE-CREATED it on every Run rotation. 0o700 passes privatedir.Validate on unix but is meaningless on Windows, where privacy = owner-only protected DACL; an MkdirAll dir inherits the parent DACL, so Validate refuses it. The fake-runner tests never hit this (fake Run never touches the portable runner), and the concrete path never executed on Windows before (pnpm absent → skip). Production is correct — fail-closed on a pre-existing non-private dir is a deliberate security property (privatedir_windows.go documents why Mkdir-then-secure is not equivalent). The harness is at fault; no product-code change.

## Proof against the path-identity hypothesis (orchestrator resource 5aqozl-windows-failure.md)
The resource hypothesizes a path-identity cause (junction/short-8.3/case) and directs creating the output root under a resolved real path (EvalSymlinks). Two independent proofs show path form cannot be the cause of the observed error, so this run fixes the creation call instead and does NOT add EvalSymlinks:
1. Code proof: windows validatePrivate performs ZERO path-string comparisons — it opens the path and inspects handle attributes, owner SID, SE_DACL_PROTECTED, and ACEs. Resolving the name cannot change any of those verdicts for the same directory object; an inherited DACL fails before and after resolution.
2. Controlled in-repo proof (the npm sibling): npmsource newConcreteNPMRunner uses the IDENTICAL unresolved TempDir base (filepath.Join(t.TempDir(), execution), no EvalSymlinks), never pre-creates output/, and has no Run rotation — so production ensureEmptyDirectory creates its output/ via privatedir.MakeAll. Its real-npm tests (TestN01RealNPMCIUsesOnlyDerivedPrivateCache, TestVerifiedProviderObservesRealNodeLaunchedNPMBoundary) PASS on the same red Windows run (npmsource 65 pass / 0 skip / 0 fail). Same base, same form, opposite outcome; the only differing variable is how output/ comes into existence. The full npm pipeline (derive + materialize + invoke) passing with unresolved roots also rules out a downstream path-identity failure mode.
The fix below therefore makes the pnpm fixture do exactly what production already does for npm output dirs. Directive hard constraints honored: fixture-only change, no skips, rest of rev1 untouched. Reviewer/orchestrator: if you still want EvalSymlinks hygiene on top, it is a compatible follow-up, but the evidence says it is unnecessary for this error.

## This run change (only file touched)
internal/pnpmsource/conformance_test.go (+19/-2 over run 3):
- Execution-root creation loop: os.MkdirAll → privatedir.MakeAll (unix-identical: MkdirAll 0o700; Windows: owner-only protected DACL). Only output/ is validated; bin/work ride along.
- concretePNPMRunner.Run rotation: os.MkdirAll → privatedir.MakeAll (the rotation recreates the validated dir before EVERY operation; without this the 2nd op would fail on Windows even with creation fixed).
- import privatedir (already a pnpmsource production dependency — no new dep, no cycle). Comments cite gate run 35091265197 + the npm control. Test-only, no production code, no new skip, skip classes untouched.
- Authorization note: Decision 2 authorized the entry-point resolution only. This extends the same test-only class to output-root creation in the same fixture — no options tradeoff (unlike run 2 A/B/C), no product/API decision, the repo own helper used exactly as production uses it. Implemented as ordinary rework; the hosted gate (not local judgment) proves Windows.

## Validation — real commands, real exit codes (shell: bash; serialized under host contention)
1. `gofmt -l cmd internal` → no output, exit 0.
2. `go vet ./internal/pnpmsource/` → exit 0.
3. `GOOS=windows go vet ./internal/pnpmsource/` → exit 0 (test file compiles for Windows).
4. `golangci-lint run ./internal/pnpmsource/` (v2.12.2 = CI pin) → 0 issues, exit 0. (First attempt timed out at 300s under host contention; serialized retry green.)
5. `go test -run TestResolveWindowsPNPMEntrypoint` → PASS, exit 0.
6. Rehearsal (PATH=/tmp/pnpm-prefix/bin:<spare-node>/bin:$PATH; `node <pnpm> --version` → 10.33.0 exit 0): `go test -run TestRealPinnedPNPM -v` → 3x PASS (9.86/15.65/18.86s), exit 0. Prefix is run 3 /tmp/pnpm-prefix leftover (same immutable pnpm@10.33.0 tarball, verified by probe; fresh-install attempt timed out under contention; layout identical to CI lane-local prefix). Node is spare v24.21.0 — the primary ~/.local/bin node binary HANGS on this host (`node --version` killed at 30s, no output; x86_64 arch matches; cause unknown, host-state, unrelated to this task). Only the pnpm release is pinned.
7. Skip control (default PATH, pnpm absent): 3x SKIP `pinned pnpm executable unavailable`, exit 0.
8. Full package with pnpm: ok 56.7s, exit 0.
9. `bash .github/ci/gate-selftest.sh` → 147 passed, 0 failed, exit 0 (pin-agree, drift-reject, 3x fail-closed, lane-wiring, no-hardcode all ok).
Not run: full landing suite (handoff-triggered, exactly-once); Windows/rose-air legs (no local runners — hosted gate proves them).
Durable-test-collateral standing: the maintained coverage for this fix IS the three TestRealPinnedPNPM* integration tests (failed-before on Windows gate, pass-after pending rev2 gate; unix-identical proven by rehearsal). No new synthetic test: a MakeAll-passes-Validate unit test would test privatedir itself (already covered by privatedir_test.go), not these call sites.

## Checklist basis (all 9 remain ticked)
- Item 2 (hosted gate green, EXECUTED on all OSes): rev1 proved ubuntu+macos EXECUTED-and-PASS and Windows EXECUTED-and-FAIL with a single-mechanism fixture cause, fixed this run. Tick = submitted-to-rev2-gate (run3 convention; handoff requires all-checked); void if rev2 red. Reviewer: read rev2 test-evidence-* artifacts.
- Others: unchanged from run3 + commands 1–9 above; item 8 = this artifact + rehearsal log.

## Final tree delta (uncommitted, for the CR snapshot)
- M .github/workflows/ci.yml (+132, runs 1–2)
- M .github/ci/gate-selftest.sh (+80, runs 1–2)
- ?? .github/ci/pnpm-pin-guard.sh (new, runs 1–2)
- M internal/pnpmsource/conformance_test.go (+80/-2 total: +61 run3, +19/-2 this run)

## Residual risks / host anomalies
- Windows proof is gate-only (no local Windows runner); reasoned from privatedir_windows.go + exact rev1 error + npm control. If rev2 shows a deeper Windows failure, it is a new diagnosis, not this one.
- rose-air verifiable only post-landing (main-push-only lane).
- Host anomalies (not task blockers): primary node binary hangs; heavy parallel-agent contention caused three timeouts (lint, go test, npm install) — all resolved by serialization/retry; another agent concurrently analyzed the same gate run (left running, untouched).
