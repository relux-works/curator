# TASK-260908-1wr53w producer evidence — SPEC §6 diagnostics delivery

Role: developer (implementer), Muse Spark xhigh. Platform: Darwin arm64,
go1.25.5. No Linux coverage claimed. No dependency, tag, or release change
(go.mod/go.sum untouched). No commits on the managed Story branch; candidate
left uncommitted for the handoff snapshot.

## What was built

`internal/diagnostics` (new): the stable SPEC §6 contract with no
behavior of its own — the closed 18-code table, `ExitForCode` (usage 2,
every operational failure 1, unknown codes 1 — never 0), deterministic
`Line`/`Emit` rendering (`curator-run: <code>: <detail>`), `CodeOf`
classification over the owners' concrete error types (never invents a
code), `IsDiagnosticLine` transport separation (warnings/child stderr are
never code lines), and `RemainingObligations` (5 stated call sites).

`cmd/curator-run/main.go` (rewired, byte-identical rendering): the usage,
resolve, and mapping failure paths now render through `Emit` and exit
through `ExitForCode`. All pre-existing entry tests pass unchanged.

`.scripts/diagnostics-mutants.sh` (new): 8 narrowing mutants, all KILLED
(see below). README gained the §6 status paragraph, a "Diagnostics
contract and remaining call sites" section, and the harness table row.

## AC coverage: 15 of 18 codes driven, 3 declared bounds

| SPEC §6 family | Production call site | Test | Status |
|---|---|---|---|
| usage | `run()` via `cli.Parse` | `TestRunDiagnosticsContract` (4 cases) + existing `TestRunUsageErrorsExit2` | driven, exit 2 |
| resolve ×6 | `run()` via `fragment.Resolver.Resolve` | `TestRunDiagnosticsContract` (6 cases) + existing `TestRunResolveFailuresExit1` | driven, exit 1 |
| environment `env_unsupported` | `run()` via `mapping.Resolve` | `TestRunDiagnosticsContract` | driven, exit 1 |
| defaults `defaults_config_invalid` | `axconfig.Load` (real temp dirs) + `CodeOf` | `TestCodeOfProductionTypes`, absence-vs-broken pair in helper | driven at API, exit 1 |
| mcp ×2 | `composition.Compose` + `Value.CheckLaunchBoundary` (real temp files) + `CodeOf` | `TestCodeOfProductionTypes`, missing/dir/readable triple in helper | driven at API, exit 1 |
| sysprompt ×2 | `systemprompt.Select` / `ProbeFiles` (real parsed fragments, temp homes) + `CodeOf` | `TestCodeOfProductionTypes`, absent-vs-dir pair in helper | driven at API, exit 1 |
| exec `exec_provider_missing` | `execution.Launch.Run` | existing `TestLateChecksBothModes` (accepted checkpoint evidence, not rerun) | driven, preserved |
| ax `ax_handoff_failed` | `execution.Launch.Run` + verbatim ax stderr | existing `TestAxFailureNoFallback` (accepted checkpoint evidence, not rerun) | driven, preserved |
| defaults `defaults_unresolvable`, plan ×2 | no producer in this build | `TestRemainingObligationsAreStated` pins the declaration | BOUND (see below) |

Entry-point contract (`TestRunDiagnosticsContract`, 11 cases): every
failure this build can produce exits with `ExitForCode`, prints exactly
one `IsDiagnosticLine` line carrying its code, forwards Curator stderr
verbatim ahead of it, and writes nothing to stdout.

Negative evidence: `TestCodeOfUnknownFailsClosed` (plain/wrapped/bare
parse/empty-typed/nil/mapping-shaped errors yield no code),
`TestDiagnosticLineDistinguishesTransport` (13 non-lines incl. embedded
tokens, stale and invented codes, Curator lines), absence-vs-failure
pairs for layer/config/home-file probes, terminal-failure assertions
(exit 1, no fallback, no stdout).

## Narrowing mutants: 8 of 8 KILLED

`.scripts/diagnostics-mutants.sh .temp/diagnostics-mutants`, per-mutant
behavioral suites (`go test ./internal/diagnostics/
./cmd/curator-run/`), logs + `summary.tsv` under `.temp/`
(gitignored evidence, reproduced on demand):

- D01 exit gate admits one operational code as usage → `TestExitForCode`
- D02 unknown codes exit 0 → `TestExitForCode`
- D03 `Valid` admits unreachable `environment_home_stale` → transport test
- D04 unknown class hidden in resolve family → no-invention test
- D05 `CodeOf` drops exactly `mcp_layer_missing` → production-types test
- D06 anchored parse weakened to contains (token preserved), behavioral suite → transport test
- D07 entry resolve path returns 0 → `TestRunDiagnosticsContract`
- D08 `Emit` drops the code separator → `TestEmitDeterministic`

## Gates rerun by the producer

`go build ./...`, `make fmt-check`, `go vet ./...`, unit tests for
cmd/curator-run, diagnostics, cli, mapping, axconfig, fragment,
composition, systemprompt (all ok), `go test -race` for diagnostics +
cmd (ok), diagnostics-mutants harness 8/8 KILLED. The execution package
suite was not replayed (untouched by this change; covered by build+vet);
its two diagnostics rows rest on accepted checkpoint evidence.

## Declared bounds and remaining call sites (not claimed)

- `defaults_unresolvable`, `plan_refused`, `plan_provider_limited` have
  no producer; wired by TASK-260908-1o7i8y (defaults/lineup,
  axconfig-before-parse) and TASK-260909-2vy977 / TASK-260908-2so46q
  (BuildLaunch admission, provider-limits verdict with verbatim
  evidence). Indeterminate limit reads stay terminal refusals.
- MCP layer stat, sysprompt probe, and binary check are classified at
  their APIs here; binding them immediately before both handoff and
  exec is final-pipeline wiring (see `RemainingObligations` + README).
- No real installed launches, no real ax/model calls, no Linux runtime
  claimed. Provider-limits evidence shape and ax Structured Error
  pass-through are preserved as owned by their stages, not re-implemented.
