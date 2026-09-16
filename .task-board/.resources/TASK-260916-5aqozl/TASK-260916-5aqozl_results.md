# TASK-260916-5aqozl — STOP: the specified corepack wiring fails the real-pnpm harness

Status: BLOCKED — implemented faithfully per brief, then proven red by controlled local rehearsal. Do not land the worktree as-is. One operator decision unblocks.

## Result in one table (darwin/amd64, go 1.26.0, node 24, corepack 0.34.0)

| pnpm first on PATH | T1 TargetPruned | T2 LockSuperset | T3 PrivateStore | go test exit |
|---|---|---|---|---|
| absent (hosted today) | SKIP | SKIP | SKIP | 0 (baseline, /tmp/nopnpm-baseline.log) |
| corepack shim, `node <shim> --version` = 10.33.0 (brief as specified) | FAIL | FAIL | FAIL | 1 (attached: ..._corepack-shim-3xFAIL.log) |
| self-contained npm-installed pnpm 10.33.0 (control) | PASS | PASS | PASS | 0 (attached: ..._real-pnpm-3xPASS.log) |

All three shim failures share one signature: `closure_derivation_drift: portable process exited 1 ... Error: Cannot find module ".home/.cache/node/corepack/v1/pnpm/12.4.2/bin/pnpm.cjs"`.

## Why shims cannot work (mechanism, with code refs)

1. Admission PASSES: `node <shim> --version` prints 10.33.0 under the real HOME where `corepack prepare --activate` wrote the payload, so `newConcretePNPMRunner` (internal/pnpmsource/conformance_test.go:1226-1238) proceeds instead of skipping.
2. Staging copies the WRONG tree: the harness resolves symlinks and stages `Dir(Dir(pnpm))` as the toolchain. For a corepack shim that is corepack`s own package dir; the staged entrypoint becomes the dispatcher (`dist/pnpm.js`), not a self-contained pnpm (the harness models `toolchain/pnpm/bin/pnpm.cjs`; conformance_test.go:1540).
3. The staged dispatcher runs with the EXACT permit env — no host merge (internal/closureexec/portable_runner.go:157-165): `HOME=.home` (staged, empty), no COREPACK_HOME, offline (`Network: none`, COREPACK_ENABLE_DOWNLOAD_PROMPT=0; internal/pnpmsource/materialize.go:987-988). The 10.33.0 payload activated under the real HOME is invisible.
4. Payload-less dispatcher falls back to its bundled default (observed: 12.4.2; no `packageManager` field exists in fixtures — verified by grep), finds nothing cached under staged HOME, cannot download → MODULE_NOT_FOUND → exit 1 → test FAIL.

The staging/env code in this path has no GOOS branches, so linux behaves identically; only Windows differs (see risk below). This is not an npx-vs-bundled artifact: bundled corepack shims are the same symlinks-to-a-dispatcher.

## Failed assumption

"A `corepack enable` shim is a pnpm the harness can stage and run." False: it is a dispatcher that needs a corepack home the staging deliberately does not carry. Any shim-first-on-PATH variant fails the same way; working around it inside CI (scraping corepack`s internal cache layout, per-OS home guessing) would be compensating hacks around a broken assumption — hence this stop, per standing orders.

Secondary finding (kept in tree): `corepack enable --install-directory` requires the directory to EXIST (observed ENOENT/lstat failure); the brief`s step omits `mkdir -p`. All four lane steps now create it first (harmless if some corepack ever creates it).

## What is implemented in the worktree (faithful to brief; ON HOLD)

- `.github/workflows/ci.yml` (+103): workflow-level `PNPM_PIN: 10.33.0` with the macbook-iv/run-35072267145 incident comment; setup-node 22 + pin-guard + corepack-install steps in test, test-self-hosted (rose-air, setup-node kept), race, candidate-conformance. interop untouched (runs only ./internal/interop, which does not import pnpmsource — verified). skip-classes.tsv untouched. No Go changes.
- `.github/ci/pnpm-pin-guard.sh` (new, +x): fails the lane unless workflow PNPM_PIN equals `SupportedPNPMVersion` grepped from internal/pnpmsource/errors.go; unset pin / unreadable source / missing declaration all fail closed.
- `.github/ci/gate-selftest.sh` (+79): 5 guard cases (match, drift, unset, unreadable, no-declaration) + lane-wiring check (every test-gate.sh lane has guard→enable→prepare in order) + no-hardcoded-`pnpm@<ver>` check. Full selftest: 147 passed, 0 failed, exit 0. Narrowing mutant (one `corepack prepare` line deleted): 146/1, exit 1 — the wiring check bites.
- Verified: guard direct (match exit 0; drift/unset/unreadable exit 1); workflow parses as YAML (ruby), 8 jobs, PNPM_PIN=10.33.0.

The pin/guard/PATH-precedence/selftest design transfers unchanged to option A below; only the 4 install-step bodies (and 2 selftest search patterns) would change.

## Options and tradeoffs

A. Provision SELF-CONTAINED pnpm: `npm install -g "pnpm@$PNPM_PIN" --prefix <lane-temp>`, prepend its bin via GITHUB_PATH. Keeps every goal: single-source pin + guard, first-on-PATH outranks broken shims, incident comment, skip-classes untouched. PROVEN 3xPASS on darwin. Deviation: installer choice vs the 2026-09-16 operator decision. RECOMMENDED.
B. Corepack-download + cached-payload-dir first on PATH. Nominally "through corepack" but depends on undocumented cache layout (`v1/pnpm/<ver>/bin/pnpm.cjs`) x 3 OS-specific cache homes; deviates from the AC step anyway (no enable/shim-dir). Fragile. NOT recommended.
C. Change the Go harness to support dispatchers (stage corepack home / pass COREPACK_HOME). Forbidden by the brief, and architecturally wrong: a dispatcher breaks the pinned-toolchain fingerprint the closure model attests. NOT recommended.
D. Land the specified wiring. Proven red on unix. REJECTED.

Windows risk (applies to A; unverifiable from this darwin machine): the admission probe runs `node <pnpm> --version`, written for a JS entry point; Windows PATH wrappers are .cmd/.ps1/shell, which node cannot execute (likely `read pnpm version` Fatal with ANY provisioner). The hosted gate at handoff is the only Windows proof available. If Windows fails at admission, the follow-up decision is: minimal Go probe fix (currently out of scope) vs declaring Windows execution out of scope (Windows keeps its declared host-capability skip).

## Exact decision needed (operator)

May the provisioner be `npm install -g "pnpm@$PNPM_PIN" --prefix <lane-temp-dir>` with that bin dir prepended via GITHUB_PATH, instead of corepack shims — keeping the single-source PNPM_PIN, `pnpm-pin-guard.sh`, PATH precedence, and the incident comment? A yes unblocks a small rework of the 4 install steps; a no needs an alternative that survives the staged-offline mechanism above.

## Validation commands and real exit codes (this session, shell bash)

- `PNPM_PIN=10.33.0 bash .github/ci/pnpm-pin-guard.sh` → 0; drift/unset/unreadable variants → 1/1/1.
- `bash .github/ci/gate-selftest.sh` → 0 (147 passed, 0 failed); with one `corepack prepare` line deleted → 1 (146/1).
- `ruby -ryaml YAML.load_file(ci.yml)` → parses; jobs=test,test-self-hosted,race,lint,gate-selftest,interop,naming-gate,candidate-conformance; PNPM_PIN=10.33.0.
- `go test ./internal/pnpmsource/ -run TestRealPinnedPNPM -v -count=1`: no-pnpm → 0 (3xSKIP); corepack-shim-first → 1 (3xFAIL); real-pnpm-first → 0 (3xPASS).
- NOT run: hosted gate (would be red by construction — not triggered); rose-air lane; Windows/macOS-hosted lanes (no access from here); full landing suite (blocked, no handoff).
