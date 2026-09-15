# TASK-260909-2vy977 producer evidence

State: ready for review. Candidate is UNCOMMITTED on the managed Story
branch `task-board/story/STORY-260908-1wxjbs`; no commit, rebase, or
reset performed. Parent owns checkpoint and signed delivery.

## Scope and decisions

Completed SPEC §4.3 against the real tagged `agents-management` module
(`v0.5.10`, pin unchanged) and wired it into the production entry point:

- New `internal/defaults/lineup.go`: `Files.Complete` applies levels 1-2
  through the untouched `Files.Resolve`, then level 3 for each member left
  unset. Runtime discovery scans `RuntimeDeclarations` by mapped system —
  no launcher-side runtime table, so no invented runtime id. The top of
  `vendorplugin.Lineup` over the binding's `DrivenBy` rows supplies the
  model; that row's `Effort.Recommended` supplies the effort (nothing when
  `EffortSupport` is `None`). A configured model takes its own row's
  effort; an unknown configured model leaves effort unset for admission to
  refuse. Zero or several declarations, resolution failure, and empty
  driven sets all fail with typed `defaults_unresolvable`; a `Resolve`
  failure propagates without consulting the lineup; unmapped environments
  keep the mapping refusal. `Resolved.Describe`/`Line`/`EmitGroup` render
  the one-line origin group with `diagnostics.Line`-style folding.
- `NewRegistry` registers exactly the `claude-code`/`codex` systems, the
  frozen declarations, and the `anthropic`/`openai` vendors. The legacy
  `pi` system plugin is deliberately not registered (wrapper replaces the
  managed home; backs no runtime). `pi-native` needs no launcher change
  when declared upstream: the scan is system-keyed.
- `cmd/curator-run/main.go`: after mapping, `Load` → `Complete` →
  `EmitGroup` → pending plan stage. `usage` (locked flags) exits 2;
  `defaults_config_invalid` / `defaults_unresolvable` exit 1; nothing
  launches on any failure. `main` builds deps from process
  `XDG_CONFIG_HOME`/home/registry; tests inject temp paths + real
  registries (never ambient `/etc`/home).
- `internal/diagnostics`: `defaults_unresolvable` graduated from
  `RemainingObligations` to produced (owner pair pinned); the plan
  families stay stated for TASK-260908-2so46q, which owns `BuildLaunch`
  admission, `ErrEffortMissing` completion, and the provider-limits
  verdict. Composition/exec wiring stays a later story.
- `README.md` milestone note and `.scripts/defaults-mutants.py` (12 new
  narrowing mutants + per-mutant package + `DEFAULTS_MUTANT_IDS` filter)
  updated. `internal/defaults/defaults.go` untouched: prior
  file-resolution evidence stands.

Pinned v0.5.10 observations (re-verify on pin change): lineup tops are
`claude-fable-5-1`/`high` and `gpt-6-astra`/`max` — score leaders, not the
display-Recommended rows (`claude-opus-5`, `sol`).

## Behavioral AC coverage

11 of 11 implementation rows driven; 7 stated bounds follow the table.
`run` below is `cmd/curator-run.run`; `Complete` is
`defaults.Files.Complete`.

| AC row | Production call site | Named test |
|---|---|---|
| Lineup supplies only missing members per member | `Complete` via `run` | TestCompletePerMemberFill, TestCompleteLineupTops, TestRunLineupEnvsPrintGroupBeforeRefusal |
| Typed failure, no invented model | `Complete` via `run` | TestCompleteUnresolvable, TestCompleteAmbiguousRuntime, TestCompleteNilRegistry, TestCompleteRuntimeResolutionFailure, TestCompleteWrongVendorModels, TestRunPiUnresolvable |
| No completion past refusal/failure | `Complete` via `run` | TestCompleteNoFirePastFailure, TestRunDiagnosticsContract/usage-locked-flag |
| Row recommendation (not display pick) | `Complete` | TestCompleteLineupTops, TestCompletePerMemberFill/model-takes-own-row-effort |
| No-effort semantics (None/unknown unset) | `Complete` | TestCompleteNoEffortSemantics |
| Origin line-group at every launch | `EmitGroup` in `run` | TestResolvedLine, TestRunLineupEnvsPrintGroupBeforeRefusal, TestRunMapping |
| Group ordered before plan-request point | `run` | TestRunLineupEnvsPrintGroupBeforeRefusal, TestRunMapping |
| Locks/failures refuse with no launch | `run` | TestRunDiagnosticsContract (locked/invalid/unresolvable), TestRunPiUnresolvable, TestCompleteMappingRefusal |
| All three envs through the real module | `NewRegistry`+`Complete` in `run` | TestNewRegistryResolvesLaunchableSystems, per-env tests above |
| Pin is a real verified tag, no override | `go.mod`/`go.sum` unchanged | `git diff --exit-code go.mod go.sum` clean; no `replace`; no `go.work` |
| Prior file-resolution evidence valid | `defaults.go` untouched | 20 prior mutants re-killed in this run (below) |

Stated bounds (not driven, declared): B1 `ErrEffortMissing`→`plan_refused`
and `BuildLaunch` admission belong to TASK-260908-2so46q. B2
provider-limits verdict gate belongs to 2so46q. B3 composition/exec
wiring is a later story (`not_implemented` remains). B4 `pi` live
resolution needs operator-only `v0.5.11`, absent upstream (evidence
below); `pi` honestly refuses at this pin. B5 `NewRegistry` error
branches (61.5% func coverage; `Complete` and all helpers 100%,
`defaults` package 96.0%): unreachable with the real module, success
path pinned. B6 `/etc` fixed-path read in the binary unicode test is
SPEC-fixed; `HOME`/`XDG_CONFIG_HOME` are hermetic there. B7 LOGBOOK not
written: prohibited by the task brief; findings live here instead.

## Upstream dependency evidence (2026-09-09, personally verified)

- `git ls-remote https://github.com/relux-works/skill-agents-management.git 'refs/tags/v0.5.10' 'refs/tags/v0.5.11'` exit 0:
  only `13d167e2... refs/tags/v0.5.10` (+ `12f443d... v0.5.10^{}`, the
  commit SPEC records). `v0.5.11` absent: no tag, no peeled ref.
- Full `v0.5.*` listing tops out at `v0.5.10`; `v0.5.8` also exists.
  Nothing was created, pinned beyond `v0.5.10`, replaced, or vendored.
- Final publication gating on the operator-only `v0.5.11` tag belongs to
  the parent/orchestrator; this candidate needs zero launcher changes
  when it lands (system-keyed scan).

## Commands personally executed

| Command | Real exit | Evidence |
|---|---|---|
| `go test ./internal/defaults/ -count=1 -cover` | 0 | 96.0% statements; `Complete` 100% |
| `go test ./cmd/curator-run/ -count=1` | 0 | all entry tests incl. 3 new defaults rows |
| `python3 .scripts/defaults-mutants.py .temp/TASK-260909-2vy977/mutants` | 0 | 34 of 34 killed (20 prior + 14 new); per-mutant logs + `summary.tsv` |
| `make check` | 0 | build, fmt-check, vet, all tests, all race tests green |
| `git diff --exit-code go.mod go.sum` | 0 | pin clean |
| `git status --short` | 0 | 7 modified + 2 new files, all uncommitted |

`make check` ran once at the end on the finished tree. Narrow package
tests ran during iteration. No hosted CI, installs, daemon restarts, ax,
runtime-home edits, or main-branch writes.

## Narrowing mutant evidence

34 of 34 killed; every child gate failed with exit 1 on its named test.
One first-round survivor (`effort-overwrite-explicit`, exit 0) exposed a
genuinely redundant inner guard subsumed by the full-pair early return;
the guard was removed so the gate lives in one place, the mutant was
rewritten to narrow the early return itself, and it is killed. A later
`runtime-drop` anchor collision (after the same restructure) was fixed by
anchoring with context.

| Mutant | Gate narrowed to / admitted case | Named failing test | Exit |
|---|---|---|---:|
| schema-version | admit schema v2 only | TestLoadRejectsInvalid/wrong-schema | 1 |
| locked-null | admit null lock only | TestLoadRejectsInvalid/locked-null | 1 |
| defaults-null | admit null defaults only | TestLoadRejectsInvalid/defaults-null | 1 |
| unknown-env | admit future env only | TestLoadRejectsInvalid/unknown-env | 1 |
| empty-entry | admit empty pi entry only | TestLoadRejectsInvalid/empty-entry | 1 |
| entry-null | admit null env entry only | TestLoadRejectsInvalid/null-entry | 1 |
| member-null | admit null string members only | TestLoadRejectsInvalid/model-null | 1 |
| extra-member | admit extra entry member only | TestLoadRejectsInvalid/unknown-member | 1 |
| sig | admit top-level sig only; token retained | TestLoadRejectsInvalid/unknown-sig | 1 |
| required-schema | allow absent schema only | TestLoadRejectsInvalid/missing-schema | 1 |
| required-defaults | allow absent defaults only | TestLoadRejectsInvalid/missing-defaults | 1 |
| duplicate | admit duplicate model only in strict JSON reader | TestLoadRejectsInvalid/duplicate-model | 1 |
| broken-link | misclassify ENOENT symlink targets only as absence | TestLoadFilesystem/broken-link | 1 |
| permission | treat permission-denied reads only as absence | TestLoadFilesystem/read-permission | 1 |
| directory | treat directories only as absence | TestLoadFilesystem/directory | 1 |
| resolve-env | admit future resolve environment only | TestLoadKnownEnvironmentsAndPresence | 1 |
| model-lock | enforce locks only for effort | TestResolveLocks/model | 1 |
| effort-lock | enforce locks only for model | TestResolveLocks/effort | 1 |
| operator-ignore | ignore operator only when machine sets effort | TestResolveLocks/model | 1 |
| presence | ignore explicitly empty flags only | TestResolveExplicitEmptyOverrides | 1 |
| lineup-bottom | supply bottom row only | TestCompleteLineupTops | 1 |
| effort-none-fill | fill effortless rows and drop required ones | TestCompleteNoEffortSemantics | 1 |
| effort-overwrite-explicit | resolve full pairs through the row effort | TestCompletePerMemberFill | 1 |
| top-effort-overwrite | overwrite explicit effort with the lineup top effort | TestCompletePerMemberFill | 1 |
| unresolvable-admit-empty | admit empty resolution | TestCompleteUnresolvable | 1 |
| ambiguity-first-wins | admit first of several runtimes | TestCompleteAmbiguousRuntime | 1 |
| fire-past-resolve-error | complete past a lock refusal | TestCompleteNoFirePastFailure | 1 |
| mapping-skip | admit unmapped environments | TestCompleteMappingRefusal | 1 |
| findrow-fuzzy | take another row's effort | TestCompletePerMemberFill | 1 |
| origin-mislabel | label lineup values as flags | TestCompleteLineupTops | 1 |
| runtime-drop | drop the recorded binding | TestCompletePerMemberFill | 1 |
| exit-usage-downgrade | downgrade usage to exit 1 | TestRunDiagnosticsContract/usage_locked_flag | 1 |
| resolution-failure-admit | admit failed resolution as success | TestCompleteRuntimeResolutionFailure | 1 |
| empty-lineup-admit | admit empty lineup as success | TestCompleteWrongVendorModels | 1 |

No source-text-inspection gate was added (typed `SystemID`
comparison only), so the token-preserving-mutant clause is not
applicable — same position as the prior leaf.

## Files changed (uncommitted)

- `internal/defaults/lineup.go` (new): registry, `Complete`, origins, group
- `internal/defaults/lineup_test.go` (new): 13 tests incl. real-module pins
- `cmd/curator-run/main.go`: `launchDeps`, §4.3 wiring, group, refusal codes
- `cmd/curator-run/main_test.go`: entry coverage per env, ordering, 3 new contract rows, hermetic binary-test HOME/XDG
- `cmd/curator-run/gate_framing_test.go`: call-site signature only
- `internal/diagnostics/diagnostics.go`: unresolvable producer named, obligation graduated
- `internal/diagnostics/diagnostics_test.go`: owner pair, obligations tokens
- `README.md`: milestone note now describes the wiring + pin
- `.scripts/defaults-mutants.py`: 14 new mutants, per-mutant package, `DEFAULTS_MUTANT_IDS` filter

Baseline HEAD `289ff42f037b9f86411fe7852000c466b3fe970d` plus accepted
leaf `ae8676c` (checkpoint `e6827e35`, preserved, never restarted).
Platform: macOS arm64, go1.25.5.
