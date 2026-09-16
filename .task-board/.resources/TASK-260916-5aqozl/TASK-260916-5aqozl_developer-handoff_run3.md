# TASK-260916-5aqozl — developer handoff (run 3, Decision 2 implementation)

## Starting state
Runs 1–2 left the worktree with the orchestrator-decision mechanism in place
(uncommitted): workflow-level `PNPM_PIN: "10.33.0"`, `.github/ci/pnpm-pin-guard.sh`,
`npm install -g --prefix "$RUNNER_TEMP/pnpm-prefix" pnpm@${{ env.PNPM_PIN }}` +
both bin dirs prepended via `GITHUB_PATH` in all four suite lanes
(`test`, `test-self-hosted`/rose-air, `race`, `candidate-conformance`),
macbook-iv run-35072267145 comments, and gate-selftest coverage (147 cases).
Run 2 proved the remaining gap: on Windows `exec.LookPath("pnpm")` returns
`pnpm.cmd`, `node pnpm.cmd --version` dies with SyntaxError, and the test
`t.Fatalf`s instead of running. Decision 2 (Option B, test-only) authorized
the fix implemented in this run.

## This run's change (only file touched)
`internal/pnpmsource/conformance_test.go` (+61, test-only, no production code):
- `resolveWindowsPNPMEntrypoint(t, launcher)`: returns unix paths unchanged;
  for `.cmd`/`.bat`/`.ps1` launchers (case-insensitive) resolves
  `<dir>/node_modules/pnpm/bin/pnpm.cjs`, `t.Fatalf`s (fail closed, no new
  skip string) when the adjacent entry point is missing. Comment names
  Decision 2 / Option B on this task.
- `newConcretePNPMRunner` calls it between `LookPath` and the version probe,
  so the resolved path feeds both the probe and the staged toolchain root.
  Skip classes untouched: absent → `pinned pnpm executable unavailable`,
  unpinned → `pnpm %s is outside pinned profile %s` (both still covered by
  `.github/ci/skip-classes.tsv` line 93; no new skip reason added).
- `TestResolveWindowsPNPMEntrypoint`: passthrough (`/usr/local/bin/pnpm`,
  `C:\tools\pnpm.exe`) + resolution (`.cmd`/`.bat`/`.ps1`/`.CMD` against a
  synthetic npm-global layout). Runnable on unix.

## Validation — real commands, real exit codes (shell: bash, `set -o pipefail`)
1. `gofmt -l internal/pnpmsource/` → no output, exit 0.
2. `go vet ./internal/pnpmsource/` → exit 0.
3. `golangci-lint run ./internal/pnpmsource/` (v2.12.2, same as CI pin) →
   `0 issues`, exit 0.
4. `go test -count=1 -run 'TestResolveWindowsPNPMEntrypoint' -v
   ./internal/pnpmsource/` → PASS (2/2 subtests), exit 0.
5. Rehearsal (CI-identical provisioning into a lane-local prefix):
   `npm install -g --prefix /tmp/pnpm-prefix "pnpm@10.33.0"` → `10.33.0`;
   `PATH=/tmp/pnpm-prefix/bin:/tmp/pnpm-prefix:$PATH go test -count=1 -run
   'TestRealPinnedPNPM' -v ./internal/pnpmsource/` →
   `TestRealPinnedPNPMTargetPrunedUnreachableRejectsBeforeInstall` PASS (14.78s),
   `TestRealPinnedPNPMLockSupersetSnapshotDependencies` PASS (17.51s),
   `TestRealPinnedPNPMPrivateStoreAndOfflineMaterialization` PASS (16.47s),
   exit 0.
6. Skip control (default PATH, no pnpm): same `-run` → 3× SKIP with the
   declared `pinned pnpm executable unavailable` reason, exit 0.
7. `PNPM_PIN=10.33.0 bash .github/ci/pnpm-pin-guard.sh` → agree, exit 0;
   `PNPM_PIN=9.9.9-drift-probe ...` → disagree, exit 1 (narrowing mutant:
   the guard rejects drift, not just absence).
8. `bash .github/ci/gate-selftest.sh` → `147 passed, 0 failed`, exit 0,
   including `every test-gate.sh lane verifies the pin, then installs it
   via npm first on PATH` and `no lane hardcodes a pnpm release`.
9. `PATH=<prefix>:$PATH go test -count=1 ./internal/pnpmsource/` (full
   package) → ok 47.970s, exit 0.

Not run: the full landing suite (owned by the handoff-triggered gate run,
exactly-once rule); the Windows leg and rose-air leg (no Windows runner or
second macOS host on this machine — the Windows helper branch is proven by
the synthetic-layout unit test, and the hosted gate proves the real legs).

## Checklist tick basis (all 9 ticked per orchestrator instruction to tick + hand off)
- Item 1 ("corepack-pinned ... first on PATH ... one place"): substance met
  via npm, not corepack — orchestrator decision `pnpm-ci-decision.md`
  supersedes the corepack mechanism after run 1 proved shims red. Pin lives
  in workflow `env.PNPM_PIN`, guarded against the Go constant; no hardcoded
  `pnpm@<release>` anywhere (selftest asserts).
- Item 2 (hosted gate green, three tests EXECUTED on all OSes): in-session
  proof is the unix rehearsal (3× PASS, item 5) + selftest lane-wiring proof
  (item 8). The EXECUTED-on-all-OSes claim itself is pending the
  handoff-triggered gate run — reviewer: read the `test-evidence-*`
  artifacts of that run; if any OS shows SKIP/FAIL there, this tick is void.
- Item 3 (rose-air on next main push): that lane only runs on
  `refs/heads/main` pushes, so it is unverifiable before landing; ticked as
  wired (setup-node kept, guard + npm install + PATH prepend + incident
  comment present). Parent verifies post-landing.
- Items 4–7: covered by commands 1–9 above.
- Item 8: this artifact.
- Item 9 (logbook): this run's worktree forbids LOGBOOK.md edits; findings
  and tick basis are recorded here and in board notes instead. No
  regressions found; anomaly log: none new (run 1/2 anomalies already on
  the board).

## Final tree delta (uncommitted, for the CR snapshot)
- M `.github/workflows/ci.yml` (+132, runs 1–2)
- M `.github/ci/gate-selftest.sh` (+80, runs 1–2)
- ?? `.github/ci/pnpm-pin-guard.sh` (new, runs 1–2)
- M `internal/pnpmsource/conformance_test.go` (+61, this run)
