# TASK-260906-1uf713 rework report 1: four blocking, three major, four minor

Head: `7997d97b` on `feat/agent-environments-stage-c` (5 signed commits on `833918d2`).
Base: curator main `b056e5da`. Authority: curator-spec `550579d`.
Worktree: `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`. No push, PR #61 untouched.

Commits (all signed, each building green):

- `1010cd79` B1: gate locked require_current_profile at the activation seam
- `522e6bd6` B2: copied surfaces never follow a link
- `ee6743a2` B3+B4+m2: managed roots skipped, divergent skills are loss, ledger failure is loss
- `dade7d8b` M2+M3+m1+m4+M1: unreadable leads, status reports require, class-wide locks, inert list, ledger wording
- `7997d97b` ledger: register new hermetic rows and fix path-overlay wording

Note: `1010cd79` carries the `switch.go` remove-first hardening as well as the seam gate (one file touched for two reasons before the split was planned); the split below names the exact lines per finding. `7997d97b` registers the new tests in `platform-cases.tsv` after both lanes had already executed the same test binaries; `ledger-consistency.sh` re-checked green against both evidences (see gate table).

## Attestation correction (read first)

- Increment 2 listed the `path`-overlay AC row as driven via `TestPathOverlayJoinsClosure`, which builds `Policy{Overlays: ...}` in Go. `PolicyFromConfig` cannot produce that state (reader requires a form, resolution refuses any form on a path source, `compose add` requires a form). The row was not reachable from any production surface. This revision carries it as an explicit stated bound naming TASK-260906-3x0w4y (see M1). The test is kept as machinery coverage, not as a driven row.
- Increment 2 claimed 15 of 15 while dropping increment 1's declared bound (`env status` not reporting the locked `require_current_profile`). This revision implements it (M3) and drives it through `run()`.

## Finding -> resolution (all eleven)

### B1 — `profile install --use` bypasses locked `require_current_profile`

Root cause agreed: `Config.CheckMachineUse` had one production caller (`cmdProfileUse`); `Install` performs the same §9.2 machine-scope switch on `--use` and first install without consulting it.

Fix at the seam: `Policy` carries `RequireCurrent *string` + `RequireCurrentLocked bool` from `PolicyFromConfig`; new `Policy.CheckMachineUse` mirrors the §12.2 rule; `useLocked` gates every machine-scope switch (`scope == "" && !clearScope`) before `readSource` (configuration error precedes not-found, so an uninstalled other profile is still refused). `cmdProfileUse` no longer calls `cfg.CheckMachineUse`; it relies on `UseWithPolicy`. Coverage of the seam: `profile use`, `profile install --use`, first-install auto-activation, `import --use` (via `installLocked` -> `useLocked`), and `resyncCurrentScopes` (machine + scoped resync). `Sync` moves no current pointer and is correctly ungated.

Narrowing mutant (gate stays, admits exactly the install path's profile):

```go
// internal/envprofile/envprofile.go — Policy.CheckMachineUse
- if p.RequireCurrent == nil || *p.RequireCurrent == name {
+ if p.RequireCurrent == nil || *p.RequireCurrent == name || name == "other" {
```

Killed by `TestProfileInstallUseLockedRequireRefuses` (`cmd/curator`, via `run()`): `install --use of another profile = 0` (expected `exitFail` with `environments.require_current_profile`). The use-path test with `personal` still passes under the mutant, proving exactly one member admitted.

New tests (all via `run()`): `TestProfileInstallUseLockedRequireRefuses`, `TestProfileInstallFirstActivationLockedRequireRefuses`; extended `TestPolicyFromConfigCarriesEnvGates` to assert the policy carries the knob + lock and `CheckMachineUse` refuses/admits.

### B2 — `--takeover` of a foreign-manager symlink writes THROUGH the link

Agreed, most serious. `materializeOne` wrote the `claude_code` copy with `os.WriteFile` without removing the symlink first; `os.WriteFile` follows links. Stage (c) first routes foreign symlinks into that write.

Fix: remove first before every copied write, exactly as `replaceLink` does. `switch.go`: `_ = os.Remove(target)` before the `claude_code` `WriteFile` and defensively before the fallback copy (already removed by `replaceLink`, kept explicit); `_ = os.Remove` before the marker write. `managed.go` `applyPlan` copies (managed homes): remove first as well. Sweep for other writes landing outside the manager's tree: `openBackup` reads through the link (correct, backs up foreign bytes) then the write replaces the link; `purgeHomes` only removes; `writeStoreDocument` writes inside the manager home; `copyLinkFallback` target was already removed by `replaceLink`. No other native-home write follows a link after this change.

Narrowing mutant (gate stays, admits exactly the copied surface):

```go
// internal/envprofile/switch.go — materializeOne, claude_code branch
- _ = os.Remove(target)
  if err := os.WriteFile(target, document, 0o644); ...
```

Killed by strengthened `TestForeignSymlinkStopsSwitch` (production `UseWithPolicy`): `takeover left .../CLAUDE.md a symlink; a copied surface must never follow a link`. The test now also asserts the foreign bytes are unchanged (`foreign\n`) and a backup generation exists. Before, it asserted only `result.OK`.

### B3 — `profile import` swallows curator's own managed files

Agreed: `readRootSurface` had no marker check.

Fix: `readRootSurface(envID, native, path)` skips when `envmarker.Read(native)` returns a valid marker (managed home reaches managed state via §9.2, never via import; §9.5 inventories unmanaged files). Skipped surfaces are neither detected nor losses.

Narrowing mutant (gate stays, admits exactly one member):

```go
// internal/envprofile/import.go — readRootSurface
- if marker, err := envmarker.Read(native); err == nil && marker != nil {
+ if marker, err := envmarker.Read(native); err == nil && marker != nil && envID != "claude_code" {
```

Killed by new `TestImportSkipsManagedRootContext` (production `Import` + `detectNative`): `managed roots detected: [{envID:claude_code ...}]` (the full §5.1 header bytes). The test also asserts the reassembled manifest contains no `claude_code.md`/`codex_cli.md` modules.

### B4 — same-named skills collapse silently

Decision from §9.6: one mapping entry plus a recorded loss. Deciding sentence: reassembly's "One `requires.skills` entry per mapping skills entry". A JSON object cannot hold duplicate keys, so the excess mapping entry is an unmappable detected surface; classification's "The loss list names each loss" then requires it in the loss list. Two adapters carrying the same name at the same commit/source are one declaration (collapse, no loss); divergent commits/sources are lossy.

Fix: `detectedSkill` carries `path`; `deduplicateSkills` (skills are pre-sorted ascending env) keeps the first declaration per name, drops identical repeats silently, and appends an `ImportLoss` per divergent drop (`duplicate skill "foo" diverges from claude_code at <rev>; only the first declaration carries over`). `importLocked` appends these losses before the consent gate, so divergent imports stop with `environment_import_lossy` and proceed only under `--allow-lossy` (losses re-reported as warnings). `reassembleImport` is unchanged (now receives deduped input).

Narrowing mutant (gate stays, admits exactly one member without loss):

```go
// internal/envprofile/import.go — deduplicateSkills
  if winner.git == skill.git && winner.revision == skill.revision {
      continue
  }
+ if skill.envID == "codex_cli" {
+     continue
+ }
```

Killed by new `TestImportDivergentSkillsAreLoss` (production `Import`): `divergent skills err = <nil>, want environment_import_lossy`. Under `--allow-lossy` the test asserts the survivor is the ascending winner (`claude_code`'s commit) and both the loss warning and the foreign warning name `foo`.

### M1 — path overlays unreachable (spec contradiction)

Agreed, including the three-gate composition and the ledger row stating the opposite of the test. No workaround invented; TASK-260906-3x0w4y (form required for git sources only, still in `reviewing`, rev 1 ready) owns the spec fix.

This revision: ledger row fixed to `a path overlay carrying a revision parses per the published family; the section 1 refusal fires at resolution (see TASK-260906-3x0w4y bound)`; `TestPathOverlayDeclarationParses` and `TestPathOverlayJoinsClosure` kept as machinery coverage but reported as a stated bound (below), not as driven rows. `cmdComposeList`'s `default: form = "path"` branch is dead until the spec fix lands and is named as such.

### M2 — unreadable-path diagnostic shadowed (refutes inc2 M3 survivor rationale for the unreadable branch)

Agreed. Two changes: `stateForPath` now returns `preservePathDiag(err)` (a store error already leading with missing/unreadable/invalid is returned as-is; anything else is wrapped invalid), removing the doubled `profile_source_invalid: profile_source_path_unreadable` prefix (m3) and letting the specific diagnostic lead; `pathManifestDiag` now maps an existing-but-unreadable operand (permission error via `errors.Is(err, fs.ErrPermission)`) to `profile_source_path_unreadable` instead of invalid, so the canonical chmod-000 root reports unreadable leading (previously `profile_source_invalid: read agent-context.json: ... permission denied`).

Narrowing mutant:

```go
// internal/envprofile/envprofile.go — stateForPath
- return nil, preservePathDiag(err)
+ return nil, fmt.Errorf("%s: %v", DiagSourceInvalid, err)
```

Killed by strengthened `TestUnreadablePathIsUnreadable` (now asserts `strings.HasPrefix(err, "profile_source_path_unreadable:")` and no `profile_source_invalid: profile_source_path_unreadable` doubling) and new `TestUnreadablePathRootLeadsUnreadable` (chmod-000 root, production `Install`, asserts leading unreadable).

Re-run of increment 2's M3 mutant (collapse absence→unreadable in `EnsureState`): still survives (`go test -run 'TestMissingPathOperandIsPathMissing|TestMissingOverlayPathIsPathMissing|TestUnreadablePath' ./internal/envprofile/` → ok). The survivor is not an artefact of the wrapper for the missing branch: `Install` decides missing via `pathManifestDiag` (Stat IsNotExist) before `EnsureState` runs, so the store's missing branch fires only on manifest/snapshot TOCTOU race. The review's refutation concerned the unreadable branch (ordinary addressing mode below the root reaches the store), which this revision fixes and kills with the M2 mutant above. The missing-branch survivor stands with that race bound; the production missing gate itself (`pathManifestDiag`) is killed by the inc2 M3r mutant (missing→invalid), still green.

### M3 — `env status` does not report locked `require_current_profile`

Implemented. `Status` carries `require_current_profile` (+ locked flag) from `req.Policy`; `PolicyFromConfig` already threads the knob (B1). JSON key `require_current_profile` present when set; text prints `require_current_profile: acme (locked)`. Driven by new `TestEnvStatusReportsLockedRequire` via `run()` (JSON decodes to `acme`; text names it).

### m1 — reviewer's surviving narrowing mutant on the lockable-subset gate

Fixed class-wide. New `TestSystemV2LockableSubsetIsClassWide` carries every §12.2 lockable key (`overlays_allowed`, `precedence`, `mcp_package_allowlist`, `passable_env_names`, `require_current_profile`, `isolation` toward shared) as accepted system defaults and a representative non-lockable set (`forms`, `current_profile`, `overlays`, `overlay_default_weight`, `backup_retention`) as refused with `not lockable`. The reviewer's mutant (`&& key != "forms"`) now fails on `non-lockable/forms` (`admitted: err = <nil>`). Verified with and without `CURATOR_CONFORMANCE_ROOT` in the original review; re-verified here without the root (config package hermetic).

### m2 — ledger read failure treated as nothing ledgered

Fixed. `readSkillsLedger` returns `(map, error)`: absence → empty + nil; read/decode/version failure → error. `readSkillsSurface` records the ledger file itself as the loss and contributes no skills from that surface. Driven by new `TestImportLedgerFailureIsLoss` (corrupt `.csk-managed.json`, production `Import` stops with `environment_import_lossy` naming the ledger).

### m3 — doubled diagnostic prefixes

Gone via M2 (`preservePathDiag` + permission-aware `pathManifestDiag`). The strengthened unreadable test asserts no `profile_source_invalid: profile_source_path_unreadable` doubling.

### m4 — `compose list` prints inert declarations

Fixed as a judgement call on the informative side: the list still prints the file verbatim (it edits the file), but when `!cfg.Env.OverlaysAllowed` and the profile declares any, it emits `warning: overlays_allowed is false: the listed declarations are inert; resolution joins the root alone` on stderr. Resolution behavior unchanged (root alone + manager §1 warning at load).

## AC coverage (honest ratio)

Stage (c) rows driven through the production entry point (`run()` in `cmd/curator` or `Install`/`UseWithPolicy`/`Import` in `internal/envprofile`): 14 of 15. The one stated bound is the `path`-overlay declaration (M1/TASK-260906-3x0w4y): resolution machinery for a form-free path overlay is covered by `TestPathOverlayJoinsClosure` (manufactured `Policy`, kept), but no operator can declare it from config or CLI until the spec fix lands, so it is not claimed as driven. `cmdComposeList`'s `default: form = "path"` branch is dead for the same reason.

## Bounds (honest)

- Path overlays: see M1. Bound names TASK-260906-3x0w4y (overlay form required for git sources only). Consume the spec fix when it lands; until then the file-level grammar (form required) + resolution refusal (form on path invalid) compose into an unusable §6 sentence.
- Store missing branch TOCTOU: see M2/M3 re-run. The production missing gate is `pathManifestDiag` (killed by M3r); the store's own missing branch is race-only.
- `requires` never names a path source structurally (carried from inc2; requirement entries carry git identity).
- Secondary fixed-home targets join inventory/backup through existing seams; divergent-file comparison has no probe path in revision 1 (carried).
- Import audit is the shared always-strict path; no import-specific secret-material vector added (carried).
- Hard-link rejection inherited from store discipline (carried).
- Opencode global skills below operator home; ledgered/store-target entries belong to §9.4 migration, never import (carried).
- Machine configuration cannot pre-record import consent; `Policy.Takeover` operation-scoped, never from config (carried).
- Fold-collision skips on case-folding filesystems (host-capability tolerance, ledgered).

## Gate table (exact standalone commands, observed exit codes)

All commands run as standalone processes in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`. Exit codes are the process's own (no `tee`, no unguarded pipe; where output was tailed for display the exit was captured separately via `$?` to a file).

- `go build ./...` → exit 0 (final tree, `/tmp/build.log` empty).
- `go vet ./...` → exit 0 (final tree, `/tmp/vet.log` empty).
- `gofmt -l cmd internal` → exit 0, empty output.
- `golangci-lint run ./...` → exit 0, `0 issues.` (`/tmp/lint.log`).
- `bash .github/ci/gate-selftest.sh` → exit 0, `94 passed, 0 failed` (`/tmp/selftest.log`).
- `bash .github/ci/no-broad-suppression.sh` → exit 0 (`no-broad-suppression: ok`).
- `bash .github/ci/ledger-consistency.sh /tmp/lane-default-evidence` → exit 0, `191 rows checked`, `ok` (`/tmp/ledger-default.log`).
- `bash .github/ci/ledger-consistency.sh /tmp/lane-candidate-evidence` → exit 0, `191 rows checked`, `ok` (`/tmp/ledger-candidate.log`).
- Default lane `CURATOR_CONFORMANCE_ROOT=/tmp/spec-pin-root/conformance/v1 bash .github/ci/test-gate.sh /tmp/lane-default-evidence` (root materialized via `git -C <spec> archive 0ed5c691e9208eea52f21db2fc05e226ce3516fd conformance/v1 | tar -x -C /tmp/spec-pin-root`, exit 0): shell exit masked by `| tail` (no `pipefail`), so no exit claimed; evidence states the result instead — go-test JSON 5260 test-level passes / 0 fails, `platform-case gate: ok` (32 skips), `suite-plan: served=69 deferred=3 excluded=0` (`defer internal/config`, `internal/envfragment`, `internal/envmarker` — the pin predates those families; new stage (c) rows are hermetic and ran). All eight wanted new tests observed passing in `go-test.json`.
- Candidate lane `CURATOR_CONFORMANCE_ROOT=<spec-worktree>/conformance/v1 CI_REQUIRE_FULL_ROOT=1 bash .github/ci/test-gate.sh /tmp/lane-candidate-evidence` → exit 0 (captured via `$?` to `/tmp/candidate.exit`, no pipe on the gate itself). `test-gate: go test exit=0, platform-case gate exit=0`; `suite-plan: served=72 deferred=0 excluded=0` with `CI_REQUIRE_FULL_ROOT=1 -- every package must be served`; `platform-case gate: ok` (20 skips). New CLI rows (`TestProfileInstallUseLockedRequireRefuses`, `TestProfileInstallFirstActivationLockedRequireRefuses`, `TestEnvStatusReportsLockedRequire`) and new envprofile rows all `ok` in `/tmp/candidate.log`.
- The two lanes ran sequentially, never concurrently and never alongside a `-race` suite.
- `go test -count=1 -timeout 30m ./cmd/curator` → exit 0, `ok ... 290.945s`, zero `--- FAIL` (`/tmp/cmd-suite.log`).
- Targeted suites on the final tree: `go test -count=1 ./internal/config/ ./internal/contextstore/` → exit 0; `go test -count=1 ./internal/envprofile/` → exit 0 (54s).
- Hosted CI was not consulted: nothing pushed, PR #61 untouched per the brief, so no `gh pr checks` output exists for `7997d97b`. The two local lane reproductions above stand in for it.

Windows-fixture sweep: every new fixture builds paths with `filepath.Join` over `t.TempDir()` bases; no POSIX literal, no `filepath.IsAbs` on fixtures, no backslash in values; `pinOperatorHome`-style env pinning reused (`CLAUDE_CONFIG_DIR`/`CODEX_HOME`/`XDG_CONFIG_HOME`/`PI_CODING_AGENT_DIR`, plus `HOME`/`USERPROFILE` in probes). Permission-based tests (`TestUnreadablePathIsUnreadable`, `TestUnreadablePathRootLeadsUnreadable`) probe readability and skip under the classified host-capability reason on Windows/superuser, matching the existing ledger `host-capability` tolerance. Symlink test skips where the platform refuses symlinks. Found nothing new: no volume-in-host, no escape, no fold assumption beyond the already-tolerated fold row.
