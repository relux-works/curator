# TASK-260916-3oh0u8 results — provider lookup from trust roots (E4, manager)

Story STORY-260916-2otjbn, wave 1. Spec: curator-spec `0da4020`
(`protocol/environments.md` §11/§11.1/§12/§12.1/§12.2,
`profiles/manager.md` §1, `cli/curator.md`), vectors
`conformance/v1/vectors/umbrella-provider-resolution.json`.
Shipped profile: **revision A (warning release)**; revision B is
implemented behind the same switch and covered by vectors, but not
shipped. The flip is a later release.

## Per-file changes

- `internal/config/environments.go` — new `Environments.ProviderDirectories`
  knob: closed list of absolute paths (POSIX-absolute or Windows
  drive-absolute with either separator, 1–4096 chars, unique),
  default `[]`, rendered as `[]` (never null); added to
  `EnvKnobNames` (so `env config show/set/unset` and `SplitEnvKnob`
  accept it), `LockableEnvKeys`, and `EnvLockKey`.
- `internal/config/config.go` — `environments.provider_directories`
  added to the manager §1 `LockableKeys` set (system-file carry +
  lock allowed; locked-but-unset fails closed via the existing gate).
- `cmd/curator/umbrella.go` — trust-root resolution. One rollout
  constant (`activeProviderRevision = providerRevisionA`); both
  revisions implemented in `resolveProvider`:
  - trust roots = install dir (running executable resolved through
    symlinks) then `provider_directories` in order; first executable
    regular file directly inside a root wins, non-executables
    skipped, never descending; Windows searches `name+PATHEXT` in
    order (bare name never matches, as the platform lookup does).
  - revision A: PATH selects (own ordered search matching LookPath
    semantics incl. empty-entry and PATHEXT handling); unreadable
    root fails first (`subcommand_provider_root_unreadable`, first
    root in order, never absence/fallback); managed/published
    candidate refused (`subcommand_provider_untrusted`, path +
    roots); inside-roots resolves silently; outside resolves with
    `subcommand_provider_outside_trust_roots` (path, roots, hint);
    no PATH match is `subcommand_provider_missing` (roots named)
    even when a root holds it.
  - revision B: roots scanned in order (unreadable fails at
    encounter, earlier match wins); managed/published match
    refused; diagnostic-only PATH probe refused; else missing.
    Five outcomes disjoint.
  - refused set = user-bin shim dir (directly inside), managed
    skill bin dirs (global `bin` + every configured project's
    `.agents/bin`, directly inside), environments root (at or
    below); symlink-aware on both sides.
  - discovery (`discoverProviderNames`) + posture rows
    (`providerPosture`) for §12.
- `cmd/curator/main.go` — umbrella branch threads the loaded config
  (nil when unloadable → install dir alone) into `cmdUmbrella`.
- `cmd/curator/env.go` — `env status` attaches provider posture after
  `StatusOf`.
- `cmd/curator/envstatus.go` — `provider <name>:` rows (path +
  verdict + current/non-current) and `attachProviderPosture`
  (refused/missing/unreadable rows set `NonCurrent`, so `--check`
  fails; the warning row stays current).
- `internal/envprofile/status.go` — data-only `ProviderState` +
  verdict constants and `Status.Providers` (`providers` in JSON);
  no §11 logic there (resolution stays in `cmd/curator` per scope).
- `cmd/curator/umbrella_conformance_test.go` (new) — executes every
  vector case for both revisions through `resolveProvider`.
- `cmd/curator/umbrella_test.go` (new) — hostile plant (warned
  end-to-end under A, refused under B), selection matrix, refused
  dirs incl. symlink, unreadable-never-absence, missing/warning
  text, posture rows, name/candidate parsing, shipped-revision pin.
- `cmd/curator/envconfig_test.go`, `cmd/curator/env_test.go` —
  knob CLI round-trip + lock tests; status matrix plants stub
  providers (warned-current under A) and asserts the new rows and
  the `providers` JSON member (see contract note below).
- `internal/config/environments_test.go` — knob grammar/default/
  render/lock/write tests; §12.2 transcription extended to seven
  keys per the landed spec.
- `.github/ci/platform-cases.tsv` — ledger row for
  `TestUmbrellaProviderResolutionVectors` (all platforms,
  `root-content`, mirroring the godriver row).
- `CHANGELOG.md` — Unreleased E4 warning-release entry (revision B
  later, 0016 dependency).

## How each AC line is met (file:line)

- "Conformance subset green" — `TestUmbrellaProviderResolutionVectors`
  (`cmd/curator/umbrella_conformance_test.go:58`): all 14 cases ×
  both revisions pass against `0da4020` (exit 0). Provider schema
  cases: all four `invalid-provider-directories-*` schema cases
  (manager + system families) pass through `Load`, and the minimal
  invalid vector `schema2-provider-directories-relative` passes
  through `Parse`; the full-file *valid* provider cases reject
  only for the sibling story's `transitive_system_modules` /
  `system_module_waivers` knobs (verified: zero failures mention
  `provider_directories`).
- "a curator-run planted through a PATH entry outside the trust
  roots is refused" — refused under B:
  `TestUmbrellaHostilePathPlantRefusedUnderB`
  (`cmd/curator/umbrella_test.go`), plus the vector case
  `s6-planted-path-provider-warns-then-refuses` (B expects
  `subcommand_provider_untrusted`, null resolved); warned under
  the shipped A: `TestUmbrellaHostilePathPlantWarnsUnderAEndToEnd`
  (end-to-end through `run()`: dispatches + warns).

## Validation transcripts (real exit codes, `sh`, unpiped gates)

Conformance root exported unless noted:
`CURATOR_CONFORMANCE_ROOT=…/curator-spec/conformance/v1` (`0da4020`).

- `go build ./...` → exit 0.
- `go vet ./...` → exit 0.
- `gofmt -l .` → clean (no output).
- `golangci-lint run ./cmd/curator/... ./internal/config/...
  ./internal/envprofile/...` → exit 0, 0 issues (one
  `ineffassign` in a new test found and fixed first).
- `go test ./cmd/curator/ -run '<22-test mask: all umbrella,
  vector, env-status, envconfig-knob, TestRunUnknownCommand>'`
  → exit 0, 22/22 PASS, 14/14 vector subtests PASS.
- `go test ./cmd/curator/ -run 'TestEnv|TestUmbrella|TestRun|
  TestConfig|TestGlobal|TestProfile'` → exit 0, `ok` (454 s;
  96 tests matched, zero FAIL/panic lines in the log).
- `go test ./internal/config/ -run '<knob mask>'` → exit 0.
- `go test ./internal/config/` (full) → exit 1: 44 subtests fail,
  ALL solely for out-of-scope knobs (`transitive_system_modules`,
  `system_module_waivers`, and the S4 `passable_env_names`
  default change). Proven same-root causes, none mine:
  `unsupported field` counts are 11× system_module_waivers, 8×
  transitive_system_modules, 1× lock refusal for transitive; and
  `schema2-passable-env-absent-empty` fails identically on the
  pristine baseline (verified via stash + rerun + pop, exit 1
  both ways). On the CI pin (`87a0d00`) these cases do not exist
  and the package is green.
- `go test ./internal/envprofile/ -run 'TestStatus|TestProvider|
  StatusOf'` → exit 0. Full package: NOT green in this session —
  `TestSCPOverlayResolvesAsGit` (git-subprocess overlay test in
  an untouched file) hung past the 10 m `go test` timeout under
  concurrent-agent load on this host; unrelated to this change
  (my envprofile diff adds only a struct field/type/constants).
- Skip paths: unset root → exit 0, `SKIP
  (CURATOR_CONFORMANCE_ROOT is not set)`; root without the vector
  → exit 0, `SKIP (… publishes no umbrella-provider-resolution
  vector)` — the `publishes no ` root-content pattern.
- Narrowing mutants (each reverted after; revert verified by
  grep): B-probe-refusal→missing → exit 1 (4 vector subtests +
  hostile test fail; trust-root refusal still passes, proving
  narrowness); A-warn→silent → exit 1 (4 vector subtests + 3
  warn tests fail); grammar accepts relative → exit 1
  (`schema2-provider-directories-relative` vector + 3 unit
  subtests fail).
- Full `go test ./cmd/curator/` (incl. compiled-build lanes) was
  NOT run here: the package TestMain serializes on the host
  GOROOT lock, which concurrent agents on this host hold for
  minutes at a time (one attempt failed to acquire it within
  TestMain's 5 min budget without running a single test). The
  runtime's `scripts/remote-gate.sh` runs the full suite at
  handoff. Relevant-blast-radius subsets above are green.

## Profile shipped

Revision A (warning release): `activeProviderRevision =
providerRevisionA` (`cmd/curator/umbrella.go`), pinned by
`TestActiveRevisionIsWarningRelease`. PATH still selects; the
trust verdict warns. Revision B ships in a later release.

## Deliberately out of scope

- Launcher side (`TASK-260916-16ys92`), proposal 0016 itself,
  SPEC_PIN, spec edits.
- Sibling knobs in the same root families
  (`transitive_system_modules`, `system_module_waivers`, S4
  `passable_env_names` default): not implemented; their red
  cases are reported, not patched around.
- `curator status` (project-skills surface) carries no provider
  rows: the spec defines provider posture under `env status`
  (§12) only, and the brief names `envstatus.go`. Adding rows to
  `curator status` would invent spec surface.

## Spec gaps / decisions (reported, not patched)

- §11 does not order the revision-A unreadable check against the
  managed-dir refusal. Implemented: readability first (the
  general "MUST NOT fall through … to PATH" bullet forbids
  consulting PATH selection while a root is unreadable).
- §11 does not say whether a revision-B earlier match outranks a
  later unreadable root. Implemented encounter order (earlier
  match wins; "no LATER root … fires"), with a committed test.
- "Ownership/writability check" (task description): no MUST-level
  writability enforcement exists in §11 (operator-administered
  roots are SHOULD-level and unvectorized); implemented as the
  manager-owned-directory refusal, extended to skill bin dirs.
- The root's invalid `provider-directories` schema cases are
  full every-knob files: they reject for the sibling knobs first
  and cannot isolate this grammar until that story lands. The
  minimal invalid vector + unit tests isolate it (proven by M3).
- Existing `TestEnvStatusMatrix` asserted post-repair `--check`
  exit 0 under the pre-E4 contract; the landed spec makes
  missing provider rows non-current, so the test now plants stub
  providers (warned-current under A) with that standing noted.
