# TASK-260916-11lwua — CLI aliases claude/codex (curator side) — results

## What changed

`claude` and `codex` are accepted as CLI aliases of `claude_code` and
`codex_cli`, normalized before validation/lookup; all outputs, markers,
fragments, config records, and locks keep the canonical id; aliases are
never persisted; unknown ids are refused as before.

- `internal/envregistry/envregistry.go`: new `EnvAliases` table +
  `NormalizeEnvID(id)`. `ByID` unchanged (raw aliases still
  `environment_unknown` at the wire layer; pinned by test).
- `cmd/curator/envalias.go` (new): alias usage sentences,
  `normalizeEnvKnob` (env positions: `forms.<env>`,
  `in_place_mode.<env>`, `scoped_current.<env-or-target>`,
  `isolation.<profile>.<env>`), `normalizeEnvKnobValue` (wholesale map
  keys, nested isolation keys, `shadow_acknowledged` env members;
  explicit canonical key wins on collision, deterministic).
- `cmd/curator/env.go`: `env resolve` normalizes its operand; usage
  lists the aliases.
- `cmd/curator/profile.go`: `profile use --env` normalized (incl.
  `--clear`); usage + flag help list the aliases.
- `cmd/curator/envconfig.go`: `env config show/set/unset` normalize
  knob paths; `set` also normalizes wholesale values; usage lists the
  aliases.
- `cmd/curator/umbrella.go`: `curator run <env-id>` rewrites a leading
  alias to canonical before dispatch; flag values (`run --profile
  claude`) and unknown spellings pass through verbatim — the launcher
  owns classification/refusal. §11 comments updated (rewrite is a pure
  function of operator argv; no profile/marker/fragment/config data
  influences dispatch).
- `cmd/curator/main.go`, `README.md`: top-level usage + environment
  profiles section list the aliases (README also gains the missing
  `env resolve|status|config` rows).
- Tests: `cmd/curator/envalias_test.go` (new, 10 tests through the real
  `run()` entry point) + 2 tests in
  `internal/envregistry/envregistry_test.go`.

## Scope readings (spec-faithful)

- `env status` has no environment operand in spec cli/curator.md
  (`env status [--check] [--json]`) and in code; no new operand was
  added (that would be a spec/CLI-surface change, out of scope). The AC
  row is satisfied as: after alias-driven operations, `env status`
  prints canonical ids only (tested, boundary-aware leak checks).
- `env unmanage` does not exist in this tree (spec names it; no
  implementation). Nothing to wire; when implemented it must call
  `NormalizeEnvID` on `--env`. No new subcommand was built for this
  alias task.
- Wire ids, markers, fragment schemas, defaults, spec: untouched.

## Evidence (all commands run in the story worktree, exit codes real)

- `go build ./...` → exit 0
- `go vet ./...` → exit 0
- `gofmt -l cmd internal` → no output, exit 0
- `golangci-lint run ./cmd/curator/... ./internal/envregistry/...` →
  0 issues, exit 0
- `go test ./internal/envregistry/ -count=1` → ok, exit 0
- `go test ./internal/config/` → ok, exit 0
- `go test ./cmd/curator/ -run
  'TestEnvResolveAcceptsAliases|TestEnvResolveAliasAdjacentUnknownsRefused|TestProfileUseEnvAliases|TestProfileUseUnknownEnvRefused|TestEnvStatusPrintsCanonicalAfterAliasUse|TestEnvConfigSetNormalizesEnvKnobs|TestEnvConfigSetNormalizesWholesaleValues|TestRunDispatchNormalizesAliasOperand|TestUsageListsAliases|TestNormalizeEnvKnobPositions'`
  → ok (10/10), exit 0
- Neighbors `go test ./cmd/curator/ -run
  'TestEnv|TestProfile|TestUmbrella|TestImplemented|TestRun'` → ok,
  exit 0
- Mutant probe: `NormalizeEnvID` replaced by identity → all 6
  entry-point alias tests + the unit test FAIL (killed); restored →
  green. No survivors.
- NOT run: full `go test ./...` / landing suite (brief: narrow only;
  remote gate runs at handoff).

## Coverage notes

- Driven rows (through `run()`): env resolve both spellings incl.
  env/shell formats + byte-identity with canonical; unknown/near-miss
  refusal (`cursor`, `Claude`, `claude-code`, `codexcli`, `claud`,
  `CLAUDE`); profile use/clear `--env` both spellings + canonical
  scoped-record filenames on disk; status canonical-only rows;
  config set/show/unset via alias knob paths + raw-file JSON key
  assertions; wholesale + shadow_acknowledged normalization;
  run-dispatch rewrite vs flag-value/unknown passthrough (scripted
  provider); help-text assertions at every operand surface.
- Bounds: `TestNormalizeEnvKnobPositions` (helper-direct knob table),
  `TestNormalizeEnvID` / `TestAliasesNeverReachTheWireLookup`
  (helper + layering pin).
