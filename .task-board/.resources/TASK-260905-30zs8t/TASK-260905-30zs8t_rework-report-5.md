# TASK-260905-30zs8t — rework report 5 (F16, final cycle)

Worktree `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-a-core`,
branch `feat/agent-environments-stage-a`, head **`834b40f6`**
("Stage (a) rework 5: machine pass skips adapters with a scope record (F16)"),
one signed commit on the cycle-5 head `b6f00e1a`, 3 files, +377/−10.
Twelve commits on curator main `a2406dfe`, all `G Ivan Oparin <oparin@me.com>`.
No push, no tag, no PR. No rewrite at any point.

Authority: curator-spec main `f39f4a9`, environments revision 1.1 §9.2/§9.3.

## Finding → disposition

**F16 (major)** — a machine-scope switch overwrote a scoped adapter's
surfaces while the scope record and `profile list` kept claiming the scoped
profile. Took **option (a)** per the brief: the machine pass **skips adapters
that carry a scope record**.

- `internal/envprofile/switch.go` — `materializeScope` consults
  `ScopedCurrents(home)` for the unnarrowed case (`environment == ""`) and
  skips every adapter with an `env:<id>` scope record. The narrowed case
  (`--env`) is untouched. This one join fixes all three machine-scope paths
  at once: `useLocked` (hence `profile use` and `Install` activation, which
  routes through `useLocked`), `SyncWithPolicy`'s machine pass (the scoped
  pass then writes each skipped home exactly once from its record — the
  double-write backup churn the reviewer observed is gone by construction),
  and `resyncCurrentScopes`' machine branch. Doc comments on the package,
  `Use`, `Sync`/`SyncWithPolicy`, and `materializeScope` state the skip.
- No other unnarrowed materialization path exists: the only
  `materializeScope` callers are `useLocked` and `SyncWithPolicy`
  (machine + scoped passes).

## New tests (all drive production entry points)

Library (`internal/envprofile/envprofile_f16_test.go`), all with machine
current alpha, `env:codex_cli` scoped to beta:

- `TestMachineUseSkipsScopedAdapter` — machine `Use(home,"gamma")`: results
  hold every adapter but `codex_cli`, all OK; codex bytes+marker stay beta;
  the other three homes move to gamma; current is gamma; `List` reports gamma
  current and beta `ScopedFor=[env:codex_cli]`.
- `TestInstallUseSkipsScopedAdapter` — `Install(..., Use:true)`: activation
  skips codex, reports activated, current moves, codex stays beta.
- `TestScopedUseStillSwitchesOnlyThatHome` — narrowed
  `Use(home,"delta","codex_cli")` after a machine switch: exactly one result
  (codex), machine current unmoved, other homes untouched.
- `TestSyncWritesScopedHomeOnce` — `Sync`: every adapter appears exactly
  once in the results and every home gains exactly one backup generation;
  codex bytes+marker are beta.

CLI (`cmd/curator/profile_test.go`, through `run()`):

- `TestProfileMachineUseSkipsScopedAdapter` — scope codex to two, machine
  `profile use three` (stdout carries no `codex_cli` line), codex bytes+marker
  stay two, claude moves to three, `profile list` keeps `env:codex_cli=two`;
  then `profile install four --use` with the same assertions.

## Mutant table (narrowing, not deletion)

| Mutant | Weakening | Named test that fails |
|---|---|---|
| M-F16: unnarrowed branch restored to `adapters = Adapters` (gate present, scope record ignored) | machine pass re-admits the unconditional overwrite for every adapter | `TestMachineUseSkipsScopedAdapter` — FAIL (`machine pass must skip the scoped adapter`, codex in results) |
| same mutant | same | `TestInstallUseSkipsScopedAdapter` — FAIL (`activation must skip the scoped adapter`) |
| same mutant | same | `TestSyncWritesScopedHomeOnce` — FAIL (`adapter codex_cli appears 2 times`, want exactly once) |
| same mutant | same | `TestProfileMachineUseSkipsScopedAdapter` (CLI) — FAIL, exit 1 |

`TestScopedUseStillSwitchesOnlyThatHome` passes under M-F16 by design (the
narrowed path is unaffected); it pins the other half of the shape.
`TestProfileMachineUseSkipsScopedAdapter` (CLI, through `run()`) was run
separately under the same mutant: FAIL, exit 1
(`--- FAIL: TestProfileMachineUseSkipsScopedAdapter (0.96s)`).

Mutant procedure: single-file backup of `switch.go`, mutant applied,
`go test -run F16... ./internal/envprofile/` → 3 FAIL as above (exit
non-zero), fixed file restored byte-identical (`cmp` clean), tree green
again.

## Gate outputs (each run directly, real exit codes)

| Gate | Result |
|---|---|
| `go build ./...` | exit 0 |
| `go vet ./...` | exit 0 |
| `gofmt -l cmd internal` | clean (exit 0). One self-found lint issue fixed before commit: `revive` unused-parameter on the first draft of `scopeFixture`; `golangci-lint` re-run clean |
| `golangci-lint run ./internal/envprofile/... ./cmd/curator/...` | 0 issues, exit 0 |
| `go test -count=1 -race ./internal/envprofile/` | ok 31.615s, exit 0 |
| `go test -count=1 -race -run 'TestProfile' ./cmd/curator/` | ok 10.020s, exit 0 |
| interop vectors, `CURATOR_CONFORMANCE_ROOT=<spec>/conformance/v1` | 95 `--- PASS`, 0 FAIL, exit 0; exactly 7 sub-skips, all `stage-deferred`/`tolerated-by-ledger` (`referenced-{claude-code-composed,opencode,opencode-zero-modules}`, `mcp-{claude-code,codex-cli,opencode,pi-none}`) |
| `bash .github/ci/gate-selftest.sh` | 81 passed, 0 failed, exit 0 |
| `bash .github/ci/no-broad-suppression.sh` | ok, exit 0 |
| `bash .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger` (with `CURATOR_CONFORMANCE_ROOT` set, as ci.yml) | 103 rows, ok, exit 0 |
| platform-case gate over a real `go test -json ./internal/interop/` stream, ledger scoped to interop rows, `CI_GATE_GOOS=linux\|darwin\|windows` | ok on all three, exit 0, 7 skips recorded each |
| 15 stage-A/dependency packages (`envprofile interop contextlock contextmaterialize contextresolve contextpkg contextaudit contextstore envmarker pkgversion identity audit closure managerlock transaction`) | all ok, exit 0 |
| remaining 54 packages | all ok (3 with no test files), exit 0 |
| `go test -count=1 -timeout 30m ./cmd/curator/` | ok 264.022s, exit 0 |

## Stated bounds and follow-ups (named, not silent)

- **FU-1** (`profile_source_path_missing`/`profile_source_path_unreadable`
  diagnostic names do not exist; both shapes refuse fail-closed as
  `profile_source_invalid`): larger than a one-line correction in the touched
  code — carried as a named follow-up for the next touch of the path branch.
- **FU-2** (a partial install whose activation fails before
  `materializeScope` prints no "installed profile" line,
  `cmd/curator/profile.go:78`): outside the touched code — carried as a named
  follow-up.
- All cycle-5 "verified, held" bounds (`loadMachinePolicy` absence rule,
  stage-(b) surfaces, non-journaled per-entry payloads with re-run recovery,
  MCP allowlist awaiting manager-config v2, `fetchRaw` first-spelling-wins,
  F7 waivers, F2 migration warning) are unchanged by this commit; the
  `switch.go` package doc already states the transaction bound.

## AC coverage note

The cycle-5 gap is closed: the machine-switch-over-a-scope row is now
**3 of 3** — machine `Use`, `Install --use`, and CLI `profile use` /
`install --use`, each by a named committed test through the production entry
point — plus the narrowed-scope pin and the sync-once pin. Brief items 1–8
stand as verified in cycles 1–5 with this fix; no other behavior changed
(`git diff b6f00e1a..834b40f6` touches only `switch.go` docs+branch, the new
test file, and the new CLI test).
