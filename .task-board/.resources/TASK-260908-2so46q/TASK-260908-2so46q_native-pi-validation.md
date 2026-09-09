# TASK-260908-2so46q — native-Pi validation at real v0.5.11 (rev2)

## Tag verification (actual, not pseudo)

- `git ls-remote https://github.com/relux-works/skill-agents-management.git
  'refs/tags/v0.5.11' 'refs/tags/v0.5.11^{}'` exit 0:
  - `0ea486e46765ecf12fff7c2ac526e12da02e95ed  refs/tags/v0.5.11`
  - `a2a6e9f377f62a5872d99ecdfff0d1690e385f2a  refs/tags/v0.5.11^{}`
- Both OIDs equal the operator-published values in
  `native-pi-tag-published-resume.md` (tag object, peeled PR23 head).
- `go list -m github.com/relux-works/skill-agents-management` = `v0.5.11`.
- Pin is a plain `require` in `go.mod`; no `replace`, no `go.work`
  (`ls go.work*` absent, `grep replace go.mod` empty). Evidence:
  `.temp/TASK-260908-2so46q/tag-03.log`, `pin-01.log`.

## Candidate changes (UNCOMMITTED, checkpoint 289ff42f037b9f86411fe7852000c466b3fe970d)

- `go.mod`/`go.sum`: `skill-agents-management` `v0.5.10` -> `v0.5.11`
  via `go get` (exit 0). No other dependency change.
- `internal/plan/plan.go`: registers the real native-Pi contracts with
  production blank imports —
  `pkg/agentic/systems/pinative` (system `pi-native`) and
  `pkg/vendorplugin/vendors/google` (completes the vendor set for
  `pi-google`; anthropic/openai were already registered). No logic change:
  `Build` already passes the runtime through to the real tagged
  `vendorplugin.BuildLaunch`; the v0.5.10 Pi refusal came from the module,
  not from a launcher gate. Package doc updated from the v0.5.10
  no-provider bound to v0.5.11 native support.
- `internal/plan/plan_test.go`:
  - `fixture` creates a `pi` stub binary beside `claude`/`codex` and
    resolves per-runtime model/effort from the tagged rows:
    `pi-anthropic`/`claude-opus-5`/`high`,
    `pi-openai`/`gpt-5.6-sol`/`max`,
    `pi-google`/`gemini-3.1-pro-preview`/effort-none.
    Every Pi word is one installed Pi 0.84.2 runs as requested
    (parser-valid, not catalog-null-clamped).
  - `TestTaggedInteractivePlans` gains the third environment row
    `pi` -> `pi-anthropic`/`pi-native` with argv golden
    `["--model", "anthropic/claude-opus-5", "--thinking", "high"]`
    and a `binary` column (`pi`, not the runtime id).
  - `TestPiReleaseBound` (refusal-only v0.5.10 bound) replaced by
    `TestNativePiRuntimes`: all three Pi runtimes through production
    `plan.Build` with `plan.DefaultDeps` (real tagged `BuildLaunch` +
    real store, determinate-absent read admits), asserting system, named
    Interactive mode, managed home/workdir, `pi` binary, exact argv
    goldens (`anthropic/claude-opus-5 --thinking high`,
    `openai/gpt-5.6-sol --thinking max`,
    `google/gemini-3.1-pro-preview` with no `--thinking`),
    exact limits-query triple, and no child started.
- `README.md`: plan boundary paragraph moved from the v0.5.10 Pi block
  to v0.5.11 three-runtime goldens; main-wiring bound unchanged.
- Preserved: prior `internal/plan` refusal/limits/isolation tests,
  `.scripts/plan-mutants.py` (12 mutants, unmodified), prior outcome
  `TASK-260908-2so46q_results.md` (historical evidence, not rewritten).

## Coverage: 13 of 16 AC rows driven at production `plan.Build`; 3 stated bounds

| Row | Required behavior | Production call site / named test | Result |
|---|---|---|---|
| 1 | §4.2 Claude/Codex/Pi runtime matches mapped system | `plan.Build` → `vendorplugin.BuildLaunch`; `TestTaggedInteractivePlans` (3 env rows), `TestNativePiRuntimes` | driven |
| 2 | Explicit resolved model/effort unchanged | `plan.Build` → `SpawnRequest`; `TestSpawnRequestShape` | driven |
| 3 | Managed Home passed; blank Home cannot use native default | `plan.Build` → `SpawnRequest`+`Availability`; `TestTaggedInteractivePlans`, `TestRequiredInputs` | driven |
| 4 | Supplied WorkDir unchanged, required | `plan.Build` → `SpawnRequest`; `TestSpawnRequestShape`, `TestRequiredInputs` | driven; current-cwd choice is row 14 |
| 5 | Supplied Env unchanged; child Env preserves inherited marker | `plan.Build` → `BuildLaunch`; `TestSpawnRequestShape`, `TestTaggedInteractivePlans`, `TestNativePiRuntimes` | driven; `os.Environ` wiring is row 14 |
| 6 | Empty Composition, zero Run, no goal/budget/tier/assignment/profile/engine | `plan.Build` → `SpawnRequest`; `TestSpawnRequestShape` | driven |
| 7 | Named Interactive; exact argv without bypass; unattached stdin; no child | `plan.Build` → `BuildLaunch`; `TestTaggedInteractivePlans`, `TestNativePiRuntimes` (5 argv goldens total) | driven, all envs |
| 8 | Module refusal retained, missing-effort guidance, invalid/empty model/runtime, invalid effort, cancellation/nil registry terminal | `plan.Build` → `BuildLaunch`, `RefusedError`; `TestTaggedAdmissionRefusals` | driven |
| 9 | Separate single Store read keyed by exact runtime/model/Home | `plan.Build` → `Availability`; `TestTaggedInteractivePlans`, `TestNativePiRuntimes`, `TestRealStoreIsolation` | driven |
| 10 | Only Serviceable; Unknown/Limited/Unreachable/invalid refused with full evidence | `plan.Build` → `Serviceable`, `LimitedError`; `TestProviderVerdicts` | driven |
| 11 | Reader error terminal, no retry/default/downgrade; single calls | `plan.Build` → `Availability`; `TestProviderReadError`, `TestTaggedAdmissionRefusals` | driven |
| 12 | Determinate absence admits; corruption refuses; profile/model-group isolation | `plan.Build` → `DefaultDeps` → real `Store.AvailabilityFor`; `TestRealStoreIsolation` | driven |
| 13 | All three adapters at real native-Pi tag | v0.5.11 pinned; 3 env rows + 3 Pi-runtime goldens at `plan.Build` | driven (was BLOCKED) |
| 14 | Both main tracked/untracked paths use current cwd / `os.Environ` | main still `not_implemented`; owned by TASK-260908-1o7i8y | stated bound |
| 15 | Other module admission classes (model not driven, unsupported mode, unresolved vendor) | delegated to `BuildLaunch`; not independently exercised here | stated bound |
| 16 | Complete publication, independent Astra-medium acceptance, signed delivery | no CR from this run; orchestrator owns integration | pending |

## Negative evidence: 12/12 narrowing mutants killed

`python3 .scripts/plan-mutants.py .temp/TASK-260908-2so46q/mutants-03`
exit 0; every mutant exit 1 with its named behavioral test failing
(`summary.tsv` + per-mutant logs in `mutants-03/`). Unmodified harness from
prior run; anchors still unique at v0.5.11. No production gate scans source
text, so a token-preserving source-checker attack remains inapplicable;
acceptance is the named behavioral suite failure in each log.

## Fresh validation (this producer, exit codes observed)

| Command | Exit | Evidence |
|---|---|---|
| `go test ./internal/plan -count=1 -v` | 0 | `.temp/TASK-260908-2so46q/test-04-native-pi.log`; all suites incl. `TestNativePiRuntimes/{pi-anthropic,pi-openai,pi-google}` and `TestTaggedInteractivePlans/pi` PASS |
| `python3 .scripts/plan-mutants.py .temp/TASK-260908-2so46q/mutants-03` | 0 | 12/12 killed |
| `go test ./internal/plan -count=1 -race -cover` | 0 | `race-03.log`; 90.4% statements |
| `go vet ./internal/plan`, `go build ./internal/plan`, `gofmt -l`, `git diff --check` | 0 | clean |
| `make check` (build, fmt-check, vet, test, race — full suite, run exactly once) | 0 | all 10 packages PASS in both `test` and `race` |

No hosted CI, installs, daemon restarts, real `ax`, runtime-home edits,
or control-root/LOGBOOK writes. Candidate left UNCOMMITTED; no commit, tag,
release, or branch operation performed. Prior fallback provenance
(RUN-260909-b70807 transport failure, successors c6584a/752162/83a1ac) is
accepted from the attached precondition logs, not rerun.
