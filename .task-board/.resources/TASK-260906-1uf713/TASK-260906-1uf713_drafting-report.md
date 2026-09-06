# TASK-260906-1uf713 drafting report — stage (c) increment 1: config schema 2 + env-config/compose CLI

Branch `feat/agent-environments-stage-c` in `/Users/iv/Developer/ReluxWorks/.worktrees/curator-stage-c`,
one signed commit on top of `b056e5da`. Authority curator-spec `550579d`
(`protocol/environments.md` rev 1.1, `manager-config-v2` / `system-config-v2`
schemas, `cli/curator.md` rows).

## What this increment delivers

Manager-config schema 2 and system-config schema 2, i.e. producer-brief
scope item 1 plus the two CLI rows that edit it:

- `internal/config/environments.go` (new): every §12.1 knob with its exact
  name, value grammar and default; the §12.2 lockable subset; `EffectiveOverlays`
  (locked `overlays_allowed: false` empties every overlay list);
  `CheckMachineUse` (locked `require_current_profile` refuses machine-scope
  `profile use` of any other profile); `EnvLockKey` / `SplitEnvKnob`;
  `EffectiveJSON` (vector-effective render); `ReadRaw` / `WriteRaw`.
- `internal/config/config.go`: `Parse` accepts schema 1 and 2 — a schema-1
  file stays valid, a schema-1 file carrying `environments` is rejected, an
  unknown `schema_version` is rejected explicitly; `applySystem` accepts
  system schema 1 and 2, merges `environments.<key>` per manager §1 rules 2
  and 3 with the system-file warning, rejects uncarriable knobs and
  `isolated`-direction isolation locks, and records the locked set plus the
  system path on `Config`.
- `internal/envprofile`: `Policy` gains `OverlaysAllowed` (+`ForbidsOverlays`);
  `PolicyFromConfig` carries the MCP package allowlist and `overlays_allowed`
  from the effective configuration (this resolves the stage-(a) stated bound
  that the allowlist had no machine-config surface).
- CLI rows exactly as `cli/curator.md` spells them at `550579d`:
  `profile compose add|remove|list`, `env config show|set|unset`, and the
  locked-`require_current_profile` gate inside `profile use` (machine scope
  only; scoped switches unaffected).
- Conformance: `manager-config-v2` (41 cases) and `system-config-v2`
  (24 cases) schema-case families plus the 13 `manager-config-v2` vectors,
  all driven through the production `Load`/`Parse` entry points.
- CI: `internal/config` row in `root-artifacts.tsv` (three artefacts) and
  three `platform-cases.tsv` rows with `root-unset` tolerance
  (`deferred-only`, so the candidate lane fails closed).

## AC coverage — 9 of 15 stage-(c) surfaces driven, 6 stated bounds

| # | Brief surface | Production call site | Named test | State |
|---|---|---|---|---|
| 1 | schema-2 read, every §12.1 knob/grammar/default | `config.Parse` ← `config.Load` | `TestSchema2Defaults`, `TestSchema2EveryKnobParses`, `TestSchema2KnobRejections` (40 subtests) | driven |
| 2 | §12.2 lockable subset via §1 locked machinery | `config.applySystem` ← `config.Load` | `TestSystemV2LockedKnobs`, `TestSystemV2UnlockedEnvIsDefault`, `TestSystemV2Refusals` (8 subtests) | driven |
| 3 | schema-1 stays valid; unknown version rejected | `config.Parse` | `TestSchema1StaysValid`, `TestUnknownSchemaVersionsRejected` | driven |
| 4 | `require_current_profile` refusal | `Config.CheckMachineUse` ← `cmdProfileUse` ← `run` | `TestRequireCurrentProfileGate`, `TestProfileUseLockedRequireRefuses` | driven |
| 5 | `overlays_allowed: false` empties lists | `Environments.EffectiveOverlays` ← `PolicyFromConfig` | `TestSchema2EveryKnobParses`, `TestSystemV2LockedKnobs`, `TestPolicyFromConfigCarriesEnvGates` | driven |
| 6 | isolation lockable only toward shared | `parseIsolation(systemOnly)` ← `Load` | `TestSystemV2Refusals/isolated_direction`, system family `invalid-isolation-isolated-direction` | driven |
| 7 | `env config show\|set\|unset` rows | `cmdEnvConfig` ← `run` | `TestEnvConfigShowDefaults`, `TestEnvConfigSetUnsetRoundTrip`, `TestEnvConfigSetListKnob`, `TestEnvConfigRefusals`, `TestEnvConfigLockedKnobRefuses` | driven |
| 8 | `profile compose add\|remove\|list` row | `cmdProfileCompose` ← `run` | `TestProfileComposeAddListRemove`, `TestProfileComposeRefusals`, `TestProfileComposeListsDeclaredOverlays` | driven |
| 9 | conformance families + CI ledger rows | `suite-plan.sh`, `test-gate.sh`, `ledger-consistency.sh` | `TestManagerConfigV2SchemaCases`, `TestSystemConfigV2SchemaCases`, `TestManagerConfigV2Vectors` | driven |
| 10 | composition: overlays join closure, weights, precedence emission | — | — | STATED BOUND: config surface only; resolution is the next increment |
| 11 | `path` source §1 in full | pre-existing `Install` path operand (stage a/b) | existing `profile install <path>` tests | pre-existing, untouched |
| 12 | onboarding §9.5 + `--takeover` rows | — | — | STATED BOUND: not implemented |
| 13 | import §9.6 + `profile import` row | — | — | STATED BOUND: not implemented |
| 14 | `env status` requirement row (§12) | — | — | STATED BOUND: not implemented |
| 15 | `profile install <path>` row wording | pre-existing | existing | pre-existing, untouched |

## Mutants — every new refusal narrowed, each kills a named test

| Mutant | What it narrows the gate to | Named test that fails | Survival bound |
|---|---|---|---|
| M1: `systemOnly && mode != "shared"` weakened to admit `isolated` (gate stays, one value admitted) | isolation-direction lock | `TestSystemV2Refusals/isolated_direction` (`err = <nil>`) | none — killed, reverted |
| M2: `*RequireCurrent == name` weakened to `len(name) > 0` (gate stays, admits the refused class) | require-current refusal | `TestRequireCurrentProfileGate` (`other profile admitted`) | none — killed, reverted |
| M3: `EnvLockKey("precedence…")` returns `""` (lock map stays, one knob unguarded) | env-config locked refusal | `TestEnvConfigLockedKnobRefuses` (`set on a locked knob = 0`) | none — killed, reverted |
| M4: `EffectiveOverlays` empties only multi-entry lists (flag check stays, single-entry admitted) | overlays_allowed emptying | `TestSchema2EveryKnobParses` (one declaration survives) | none — killed, reverted |
| M5: range class widened with `()` (grammar stays, paren spelling admitted; token-preserving) | overlay range grammar | `TestSchema2KnobRejections/overlay_range_grammar` (`err = <nil>`) | none — killed, reverted |

No mutant survived. Delete-only mutants were not used as evidence.

## Contract migrations (existing tests edited, not weakened)

1. `internal/config/config_test.go` `TestParseRejections/schema`: `schema_version: 2`
   rejected → `schema_version: 3` rejected. Schema 2 is now valid per manager
   §12; the unknown-version gate stays and is still proven by this row plus
   `TestUnknownSchemaVersionsRejected` (0, 3, 99).
2. `TestMalformedSystemConfigFailsClosed`: `{"schema_version": 2, "locked": []}`
   accepted-as-invalid → `{"schema_version": 3, ...}`. A v2 system file with an
   empty lock set is valid under system-config-v2; the malformed-system gate
   stays.
3. `cmd/curator/profile_test.go` `TestProfileComposeIsRefused` (stub refusal)
   → `TestProfileComposeListsDeclaredOverlays` (implemented dispatch). The
   stub existed only because schema 2 was unimplemented; the brief implements
   the row.

## Stated bounds carried forward

- Knob-level explicit nulls (`forms: null`, `overlays: null`, …) read as
  absent (codebase null convention); only top-level `environments: null` is
  rejected. No published schema-case distinguishes either way.
- `env status` does not yet report a locked `require_current_profile` (§12
  sentence); the refusal itself is enforced.
- Unlocked user-set `require_current_profile` carries no refusal: the refusal
  is a locked-key effect (§12.2 tied to the manager §1 rules).

## Windows-fixture sweep (lesson b)

New fixtures build every path with `filepath.Join` over `t.TempDir()` (a
platform-absolute base); no test interpolates a raw temp path into git config
(the `gitFileURL` helper rule is untouched); no POSIX-rooted literals feed
`filepath` gates — the only absolute literals are `/tmp/skills`-style
`skills_root` strings, which `Parse` treats as opaque non-empty strings and
never probes. `gofmt`, `go vet`, and `ledger-consistency.sh` (which proves
every ledger row compiles on linux, darwin, and windows) are green. Hosted
Windows lanes were not run here; CI was not consulted (no `gh` calls).

## Gate table

Root-sensitive gates appear twice, once per root. CI was not consulted: no
hosted lane output is quoted anywhere in this report. Every command below ran
as a standalone process; exit codes are the process's own.

- `go build ./...` → exit 0
- `go vet ./...` → exit 0
- `gofmt -l cmd internal` → exit 0, no files listed
- `golangci-lint run ./...` → exit 0, `0 issues.`
- `bash .github/ci/gate-selftest.sh` → exit 0, `81 passed, 0 failed`
- `bash .github/ci/no-broad-suppression.sh` → exit 0, `ok`
- `bash .github/ci/ledger-consistency.sh .temp/ci-evidence/ledger-local` → exit 0
- `go test -count=1 ./internal/config/` → exit 0
- `go test -count=1 -race ./internal/config/` → exit 0
- `go test -count=1 ./internal/config/ ./internal/envprofile/` → exit 0
- `go test -count=1 -race ./internal/envprofile/` → exit 0 (184s)
- `go test -count=1 -run 'TestSchema' ./internal/config/` → exit 0 (null-strictness fix)
- `CURATOR_CONFORMANCE_ROOT=<spec-main>/conformance/v1 go test -count=1 -run 'TestManagerConfigV2|TestSystemConfigV2' -v ./internal/config/` → exit 0 (41+24+13 subtests pass)
- `bash .github/ci/suite-plan.sh <pin-root> …` → exit 0, `served=69 deferred=3`, internal/config deferred naming the three new artefacts
- `bash .github/ci/suite-plan.sh <spec-main-root> …` → exit 0, `served=72 deferred=0`
- Negative control `CI_REQUIRE_FULL_ROOT=1 bash .github/ci/suite-plan.sh <root-minus-new-families> …` → exit 1, `FAIL internal/config was deferred… missing: schema-cases/manager-config-v2 schema-cases/system-config-v2 vectors/manager-config-v2.json` (fail-closed registration proven)
- `CURATOR_CONFORMANCE_ROOT=/tmp/curator-roots/rc9/conformance/v1 bash .github/ci/test-gate.sh /tmp/curator-roots/ev-rc9` → exit 1. Suite-plan correct (`served=69 deferred=3`, internal/config deferred naming the three new artefacts). The single red case, `cmd/curator :: TestCompiledProjectRepairsCorruptCompiledState` (`repairing install = 1`, empty output, ~688s/subtest), ran while this machine concurrently executed the candidate lane and a `-race` package suite (my own parallelization, not the lane's design). See forensics below.
- `CI_REQUIRE_FULL_ROOT=1 CURATOR_CONFORMANCE_ROOT=<spec-main>/conformance/v1 bash .github/ci/test-gate.sh /tmp/curator-roots/ev-main` → exit 1. Suite-plan correct (`served=72 deferred=0`). `cmd/curator` died at 300.8s with `acquire package host GOROOT test lock: context deadline exceeded` — the machine-wide GOROOT test lock both lanes were fighting over — so its ledger cases are recorded `never ran`; `internal/install :: TestStrictRegistryPolicyFailsUnknown` failed beside it (`snapshot timestamp is too far in the future`).
- Forensics (all solo, machine idle): `go test -count=1 -run 'TestStrictRegistryPolicyFailsUnknown' ./internal/install/` → exit 0 (6s). The lane failures reproduce only under the 3× parallel load I created; neither touches the schema-2 code path (schema-1 project install/upgrade and audit-registry timestamps). `go test -count=1 -timeout 30m ./cmd/curator` (the brief's standalone gate, uncontended) was launched to close the remaining item — result below.
- `go test -count=1 -timeout 30m ./cmd/curator` → exit 0 (`ok … 1318.281s`, uncontended). This run includes `TestCompiledProjectRepairsCorruptCompiledState`, so the lane-A red does not reproduce solo: both test-gate lanes failed only from the 3× parallel load on one machine (mine), never from the tree. Lesson for the next increment: run the two lanes SEQUENTIALLY on one host.
