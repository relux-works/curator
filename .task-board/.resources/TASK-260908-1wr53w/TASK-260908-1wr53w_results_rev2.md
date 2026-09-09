# TASK-260908-1wr53w producer evidence rev2 — R1-R3 rework

Role: developer (implementer), Muse Spark xhigh. Platform: Darwin arm64,
go1.25.5. No Linux coverage claimed. No dependency, tag, or release change
(go.mod/go.sum untouched). Existing rev1 candidate preserved and extended
in place; execution upstream untouched. No commits on the managed Story
branch; candidate left UNCOMMITTED for the handoff snapshot.

Base: 3ff66a9421ff6ddf675a49fc0c2868309f6e3de3 (execution PR11).

## R1 — launcher-owned multiline detail is framed (high)

`internal/diagnostics/diagnostics.go` `Line` now folds every CR and LF in
the launcher-owned detail to LF plus two spaces. The first rendered line
is therefore always the only code line; each continuation starts with
whitespace, so `IsDiagnosticLine` never recognizes it — even when the
embedded bytes spell `curator-run: <code>: <detail>`. The byte rule is
documented on `Line` and in the package comment. Only launcher-owned
detail passes through `Line`: Curator stderr, provider verdict evidence,
and ax Structured Error bytes are forwarded verbatim by their owners and
are never transformed here.

Production call site: `run()` in `cmd/curator-run/main.go` — both the
resolve path (`fragment.IsResolve` detail, which carries the raw profile
value through the argv rendering) and the usage path (`cli.UsageError`
detail, which `%q`-quotes raw flag tokens). No `main.go` logic change was
needed; the central framing covers both.

Tests: `TestDiagnosticDetailCannotForgeLine` (unit, 6 hostile details
through `Line`/`Emit`) and `TestRunDetailInjectionCannotForgeLine`
(production entry: real `fragment.Resolver` with only the binary absent
plus the reviewer's exact `normal\ncurator-run: usage: forged` profile,
and the hostile-flag usage path). Narrowing mutant D12 folds the framing
to a no-op while preserving the searched-for token; the unit test fails.

## R2 — closed per-owner families, typed-nil safe (medium)

`CodeOf` now accepts only each typed owner's own closed family
(`resolveFamily` / `layerFamily` / `refusalFamily` transcribed from the
owner constants; `UsageError` pinned to `Code() == "usage"`;
`axconfig.Error` has no code field and always reads
`defaults_config_invalid`). A wrapped `ResolveError{Code:"usage"}` or
`{Code:"invented_code"}`, and foreign codes on `LayerError`/`Refusal`,
yield no code.

Root cause of the panic: the owners' `Unwrap` methods dereference the
receiver, so even stdlib `errors.As` panics on a typed-nil owner. `CodeOf`
therefore walks wrap chains (`Unwrap() error`, `Unwrap() []error`) with a
nil-guarded traversal that ends a typed-nil branch with no match instead
of panicking. Owner packages are untouched.

Tests: `TestCodeOfRejectsForeignFamily` (direct/wrapped/joined foreign
codes per owner, plus own-family acceptance so the gate cannot
over-narrow) and `TestCodeOfTypedNil` (direct and wrapped typed nils for
all five owners). Narrowing mutant D10 admits exactly `"usage"` into the
resolve family; the foreign-family test fails.

Bound, kept honest: `CodeOf` has no production caller in this build;
`main` classifies through `cli.IsUsage` / `fragment.IsResolve` and chooses
mapping/execution codes at their call sites. Stated in the package
comment.

## R3 — honest mutants, SPEC-derived completeness (medium)

`TestClosedSetMatchesSpec` no longer hand-enumerates: it parses the Codes
column of the SPEC section 6 table out of `SPEC.md` (located from the
test file) and requires `Codes()` to equal it in table order.

`.scripts/diagnostics-mutants.sh` now carries 12 mutants and labels them
honestly: NARROWING (D01 exit admits one operational code as usage; D03
membership widened by one unreachable code; D06 token-preserving anchor
weakening; D09 unknown-code gate admits `""` as success; D10 resolve
family admits `"usage"`; D11 main resolve path admits one operational
failure as success; D12 token-preserving framing no-op) versus RETAINED
non-narrowing probes (D02/D04/D07 whole-clause replacements, D05
drop-one, D08 formatting). D04/D05 anchors were updated for the `CodeOf`
rewrite (the old `errors.As` anchors no longer exist). Result: 12 of 12
KILLED; candidate bytes restored after every mutant (verified via
`git status`: only the intended files differ).

## AC coverage: 15 of 18 diagnostic-code rows driven, 3 declared bounds

| SPEC section 6 family | Production call site | Named test | Status |
|---|---|---|---|
| usage (1) | `run()` via `cli.Parse` | `TestRunDiagnosticsContract` + `TestRunUsageErrorsExit2` + `TestRunDetailInjectionCannotForgeLine/usage` | driven through main, exit 2 |
| resolve (6) | `run()` via `fragment.Resolver.Resolve` | `TestRunDiagnosticsContract` + `TestRunResolveFailuresExit1` + `TestRunDetailInjectionCannotForgeLine/resolve` | driven through main, exit 1 |
| environment `env_unsupported` (1) | `run()` via `mapping.Resolve` | `TestRunDiagnosticsContract` | driven through main, exit 1 |
| defaults `defaults_config_invalid` (1) | `axconfig.Load` (real temp dirs) + `CodeOf` | `TestCodeOfProductionTypes`, absence-vs-broken pair in helper | driven at API, exit 1 |
| mcp (2) | `composition.Compose` + `Value.CheckLaunchBoundary` (real temp files) + `CodeOf` | `TestCodeOfProductionTypes`, missing/dir/readable triple in helper | driven at API, exit 1 |
| system prompt (2) | `systemprompt.Select` / `ProbeFiles` (real fragments, temp homes) + `CodeOf` | `TestCodeOfProductionTypes`, absent-vs-dir pair in helper | driven at API, exit 1 |
| exec `exec_provider_missing` (1) | `execution.Launch.Run` | existing `TestLateChecksBothModes` (accepted checkpoint evidence, not rerun) | driven, preserved |
| ax `ax_handoff_failed` (1) | `execution.Launch.Run` + verbatim ax stderr | existing `TestAxFailureNoFallback` (accepted checkpoint evidence, not rerun) | driven, preserved |
| `defaults_unresolvable`, `plan_refused`, `plan_provider_limited` (3) | no producer in this build | `TestRemainingObligationsAreStated` pins the declaration | BOUND (below) |

Negative evidence: `TestCodeOfUnknownFailsClosed`,
`TestCodeOfRejectsForeignFamily`, `TestCodeOfTypedNil`,
`TestDiagnosticLineDistinguishesTransport`,
`TestDiagnosticDetailCannotForgeLine`, absence-vs-failure pairs for
layer/config/home-file probes, terminal-failure assertions (exit 1, no
fallback, no stdout). No silent degradation: `ExitForCode` maps every
unknown code to 1, never 0 or 2 (`TestExitForCode` + narrowing D01/D09).

## Gates rerun by the producer (real exits)

- `go build ./...`: exit 0
- `gofmt -l internal/diagnostics/ cmd/curator-run/ .scripts/`: clean
- `go vet ./internal/diagnostics/ ./cmd/curator-run/`: exit 0
- `go test ./internal/diagnostics/ ./cmd/curator-run/ -count=1`: exit 0
  (11 diagnostics tests + 13 main tests incl. subtests, all pass)
- `go test -race ./internal/diagnostics/ -count=1`: exit 0
- `.scripts/diagnostics-mutants.sh .temp/diagnostics-mutants`: exit 0,
  12 of 12 KILLED (7 narrowing incl. 2 token-preserving; 5 retained)
- Reviewer rev1 probes replayed verbatim against the fixed tree
  (`TestReviewUnknownTypedCode`, `TestReviewTypedNil`,
  `TestReviewOneLineAtRealResolver`): all pass; throwaway files removed
  afterwards (`git status` shows only the intended candidate files)
- Full `make check` NOT replayed here per the rework brief; it runs at the
  runtime handoff.

## Bounds and review notes (not claimed)

- No installed working launcher, no real ax/model calls, no Linux
  runtime. Provider-limits evidence shape and ax Structured Error
  pass-through remain owned by their stages (`execution.Launch.Run`
  writes one `ax_handoff_failed` line then `axStderr.Bytes` unchanged),
  not re-implemented. Direct child exits/signals never route through
  `ExitForCode`.
- The pre-existing main `not_implemented` path is outside the closed
  registry and must disappear with finished pipeline wiring.
- README and `RemainingObligations` ownership corrected to the actual
  task descriptions: defaults/lineup origin diagnostics belong to
  TASK-260909-2vy977 (SPEC 4.3), plan admission and provider-limits
  enforcement to TASK-260908-2so46q (SPEC 4.4); TASK-260908-1o7i8y keeps
  install/real launches plus the call sites below.

## Call-site obligations for TASK-260908-1o7i8y (attachable section)

Each wiring step must choose the named family code at its boundary,
render through `diagnostics.Emit` (one framed code line), and exit
through `diagnostics.ExitForCode`, with no fallback to a weaker launch:

1. `axconfig.Load` before `cli.Parse`, so a present-but-unreadable
   `ax.json` reports `defaults_config_invalid` even when argv is also a
   usage error.
2. `composition.Value.CheckLaunchBoundary` with the binary check and
   `systemprompt.PrepareLaunch` as the execution boundary immediately
   before BOTH ax handoff and direct exec.
3. `systemprompt.PrepareLaunch` selection, file-kind probe, and warning
   emission on every launch.
4. Preserve: forwarded Curator/provider/ax bytes stay verbatim (never
   routed through `Line` framing); launched child statuses never routed
   through `ExitForCode`.
