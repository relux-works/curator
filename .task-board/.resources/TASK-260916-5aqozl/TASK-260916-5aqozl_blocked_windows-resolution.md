# TASK-260916-5aqozl — BLOCKED: Windows leg cannot execute the real-pnpm tests

Date: 2026-09-16. Developer: implementation of the npm mechanism is complete
and verified on unix; the Windows Test leg of the AC is infeasible under the
current test code. No handoff was triggered (handoff runs the hosted gate,
which would go red on windows-latest). Work is uncommitted in the story
worktree, purely additive vs HEAD.

## 1. Implemented (per orchestrator decision pnpm-ci-decision.md)

Switched all four test-gate lanes in `.github/workflows/ci.yml` from the
prior corepack draft to npm-installed pinned pnpm; no Go changes:

- `test`, `test-self-hosted` (rose-air), `race`, `candidate-conformance`:
  after `actions/setup-node@v4` (added to the three lanes that lacked it),
  `Verify pnpm pin matches Go constant`
  (`bash .github/ci/pnpm-pin-guard.sh`, unchanged) then
  `Install pinned pnpm via npm`:
  `npm install -g --prefix "$RUNNER_TEMP/pnpm-prefix" "pnpm@${{ env.PNPM_PIN }}"`
  plus both `$prefix/bin` (unix) and `$prefix` (Windows) prepended via
  `GITHUB_PATH`.
- Workflow-level `env: PNPM_PIN: "10.33.0"` kept as the single copy of
  `internal/pnpmsource.SupportedPNPMVersion`; env comment keeps the
  macbook-iv run 35072267145 incident note and adds the why-not-corepack
  rationale (dispatcher shim provides no package root; rehearsal 3x FAIL
  MODULE_NOT_FOUND vs 3x PASS).
- `.github/ci/gate-selftest.sh`: the lane-wiring section now asserts
  guard -> npm-install -> PATH-prepend order in every test-gate lane
  (was: corepack enable/prepare); the no-hardcoded-`pnpm@<release>` check
  is unchanged. Skip classes untouched.

Files: `.github/workflows/ci.yml` (+132), `.github/ci/gate-selftest.sh`
(+80), `.github/ci/pnpm-pin-guard.sh` (new, pre-existing draft, untouched).

## 2. Verified green (each command run directly; real exit codes)

- `bash .github/ci/gate-selftest.sh` -> exit 0, `147 passed, 0 failed`
  (includes: workflow pin agrees with Go constant; drift/unset/unreadable/
  no-declaration fail closed; all 4 lanes wired guard->install->PATH;
  no hardcoded release). Log: `TASK-260916-5aqozl_gate-selftest.log`.
- Ruby `YAML.load_file` on ci.yml -> parses; 8 jobs; `PNPM_PIN: 10.33.0`.
- `bash -n` on both gate scripts -> clean.
- Unix rehearsal, `PATH=/tmp/pnpmprobe/bin:$PATH`
  (`npm install -g --prefix /tmp/pnpmprobe pnpm@10.33.0`, exit 0):
  `go test ./internal/pnpmsource/ -run TestRealPinnedPNPM -v -count=1`
  -> exit 0, 3x PASS (8.2s / 11.0s / 12.5s). Log: `..._pnpm-rehearsal.log`.
- Skip control, pnpm absent from PATH -> exit 0, 3x SKIP with the declared
  `host-capability` reason `pinned pnpm executable unavailable`
  (skip-classes.tsv:93). Log: `..._pnpm-skip-control.log`.
- Not run: hosted gate (would run at handoff; deliberately not triggered
  while blocked), `go vet`/gofmt (no Go changes; the rehearsed package
  compiled and ran under `go test`).

## 3. The blocker: Windows resolution chain ends in t.Fatalf, not a skip

`newConcretePNPMRunner` (internal/pnpmsource/conformance_test.go:1226-1238)
does `exec.LookPath("pnpm")`, then `node <that> --version`, and calls
`t.Fatalf("read pnpm version...")` on any error -- a FAIL, to which no
skip class applies. On Windows, with the lane's npm-installed pnpm first
on PATH:

1. npm's Windows global layout puts `pnpm.cmd` (batch) + extensionless sh
   `pnpm` + `pnpm.ps1` in the prefix root. Verified byte-exact: generated
   with npm's own vendored cmd-shim (`.../npm/node_modules/cmd-shim`)
   against the installed pnpm@10.33.0 `bin/pnpm.cjs`.
2. Go's Windows `LookPath("pnpm")` returns the `.cmd`: for an
   extensionless name only `name`+PATHEXT is probed, in order
   (`findExecutable`, `lp_windows.go`; the extensionless file itself is
   never returned). First hit in our first-on-PATH dir is `pnpm.cmd`.
3. `node pnpm.cmd --version` -> `SyntaxError: Invalid or unexpected token`
   at `@ECHO off`, exit 1. Observed with the exact generated shim bytes
   (node module loading of a non-JS extension is platform-independent, so
   this is a faithful proxy of the Windows lane step). Transcript:
   `TASK-260916-5aqozl_windows-shim-probe.log`.
4. Non-nil error -> `t.Fatalf` -> all three tests FAIL ->
   `internal/pnpmsource` package fails -> windows-latest Test lane red ->
   hosted gate red. The test has no Windows branch (zero GOOS/.cmd/.exe/
   PATHEXT hits in internal/pnpmsource).

The corepack variant fails identically on Windows (corepack also writes
batch `pnpm.cmd` shims), so the mechanism choice is not the cause: the
test's `node <LookPath> --version` model assumes a node-runnable LookPath
result, which no real pnpm distribution provides on Windows.

## 4. Failed assumption

"PATH-provisioning the pinned pnpm suffices on every OS." Disproven for
Windows by the probe above. Installing per the decision turns the Windows
leg's 3 tolerated skips into 3 failures -- a predictable gate regression,
so the tree was deliberately NOT handed off.

## 5. Options

- A. Narrow the AC: exempt the Windows Test leg from pnpm provisioning
  (documented carve-out, incl. gate-selftest); the 3 tests keep skipping
  there under the already-declared `host-capability` class. Gate green;
  Windows real-pnpm coverage stays absent; contradicts "every lane".
- B. Authorize a minimal TEST-ONLY Go change: resolve the real entry point
  on Windows in `newConcretePNPMRunner` (e.g. derive the sibling
  `pnpm.cjs` when LookPath yields a `.cmd`/`.bat` launcher). No production
  behavior change; Windows then executes for real. Needs operator sign-off
  against "No Go code changes besides the optional guard".
- C. Lane-side `pnpm.js` bridge on Windows only (synthetic node-runnable
  file + PATHEXT dependence). Works on paper; fragile and exactly the
  compensating-hack shape the standing orders forbid without approval.
  Not recommended.

Recommendation: B if real Windows coverage is wanted (it is the honest
fix, confined to the test harness); otherwise A with the narrowed AC
recorded on the task.

## 6. Exact decision needed

Which option -- A (documented Windows skip exemption), B (authorize the
test-only Windows resolution fix), or C (authorize the lane-side bridge)?
Until then: DO NOT hand off or land this tree; the Windows Test leg will
go red as analyzed above.
