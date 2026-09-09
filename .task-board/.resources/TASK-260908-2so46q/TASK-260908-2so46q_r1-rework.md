# TASK-260908-2so46q — R1 rework (F1) producer evidence

Producer: Muse Spark xhigh (developer/implementer). Scope: plan-review-r1-rework.md on the
preserved candidate at checkpoint 289ff42f037b9f86411fe7852000c466b3fe970d. No commits,
no tags, no replace/pseudo/go.work, no main integration, no new architecture.

## Candidate delta (UNCOMMITTED, HEAD still 289ff42)

- `internal/plan/plan_test.go`: +3 tests, no production change (propagation gate was
  already correct; the gap was test coverage).
  - `TestModelNotDrivenBySystem` — R1 rework regression. Real v0.5.11 triple
    pi-openai / gpt-6-astra / high through production `plan.Build` with
    `plan.DefaultDeps` + real tagged `vendorplugin.BuildLaunch` (capture wrapper).
    Asserts `RefusedError`, `errors.Is(err, vendorplugin.ErrModelNotDrivenBySystem)`,
    detail words pi-openai/pi-native/gpt-6-astra, zero `agentic.Plan{}`, exactly one
    launch call in `LaunchModeInteractive`, zero limits calls, preserved Cause.
  - `TestUnresolvedVendorScope` — real fixture for the unresolved-vendor scope. Frozen
    runtime `muse` is the sole `VendorUnresolved` declaration; its system `muse` has no
    plugin in this binary (plan.go imports only claude/codex/pinative), and
    `Registry.ResolveRuntime` checks system registration before vendor resolution, so
    the real tagged `BuildLaunch` refuses with `ErrRuntimeSystemUnregistered` through
    the same `plan.Build` gate. Asserts RefusedError, sentinel, zero plan, 1/0 calls.
    `ErrRuntimeVendorUnresolved` itself is unreachable here (system check fires first;
    had the system been present, `resolveLaunchBinding`'s system-only binding would
    absorb it — evidence: registry.go ResolveRuntime order, spawn.go resolveLaunchBinding).
  - `TestInteractiveDeclaredForMappedSystems` — executable unreachable bound for
    `agentic.ErrUnsupportedLaunchMode`. `plan.Build` spells `LaunchModeInteractive` by
    name and takes no other mode; claude-code/codex/pi-native all declare Interactive
    (observed: claude-code [exec dry-run interactive], codex [exec dry-run
    managed-session interactive], pi-native [dry-run interactive]; muse/antigravity/
    qwen-code unregistered in this binary). No mapped-runtime Interactive call can
    produce it; any unmapped refusal would travel the same proven `if err != nil` gate.
- `.scripts/plan-mutants.py`: +P13, the surviving R1 narrowing mutant verbatim:
  `if err != nil` → `if err != nil && !errors.Is(err,
  vendorplugin.ErrModelNotDrivenBySystem)` on the spawn-plane propagation gate,
  bound `admit model-not-driven-by-system refusal only`, killed by
  `TestModelNotDrivenBySystem`. Prior P1–P12 untouched.
- Preserved: all prior tests (incl. `TestNativePiRuntimes` 3 Pi goldens,
  `TestTaggedInteractivePlans` 3 env rows), `internal/plan/plan.go` production code
  unchanged, real v0.5.11 pin (`go.mod` require, no replace; `ls go.work*` absent).

## Coverage: 14 of 16 AC rows driven at production `plan.Build`

| Row | Required behavior | Production call site / named test | Result |
|---|---|---|---|
| 1 | §4.2 Claude/Codex/Pi runtime matches mapped system | `plan.Build` → real `vendorplugin.BuildLaunch`; `TestTaggedInteractivePlans` (3 env rows), `TestNativePiRuntimes` (3 Pi runtimes) | driven |
| 2 | Explicit resolved model/effort unchanged | `plan.Build` → `SpawnRequest`; `TestSpawnRequestShape` | driven |
| 3 | Managed Home passed; blank Home cannot use native default | `plan.Build` → `SpawnRequest`+`Availability`; `TestTaggedInteractivePlans`, `TestRequiredInputs` | driven |
| 4 | Supplied WorkDir unchanged, required | `plan.Build` → `SpawnRequest`; `TestSpawnRequestShape`, `TestRequiredInputs` | driven |
| 5 | Supplied Env unchanged; child Env preserves inherited marker | `plan.Build` → `BuildLaunch`; `TestSpawnRequestShape`, `TestTaggedInteractivePlans`, `TestNativePiRuntimes` | driven |
| 6 | Empty Composition, zero Run, no goal/budget/tier/assignment/profile/engine | `plan.Build` → `SpawnRequest`; `TestSpawnRequestShape` | driven |
| 7 | Named Interactive; exact argv without bypass; unattached stdin; no child | `plan.Build` → `BuildLaunch`; `TestTaggedInteractivePlans`, `TestNativePiRuntimes` (5 argv goldens) | driven |
| 8 | Module refusal retained, missing-effort guidance, invalid/empty model/runtime, invalid effort, cancellation/nil registry terminal | `plan.Build` → `BuildLaunch`, `RefusedError`; `TestTaggedAdmissionRefusals` + `TestModelNotDrivenBySystem` (mismatch class) | driven |
| 9 | Separate single Store read keyed by exact runtime/model/Home | `plan.Build` → `Availability`; `TestTaggedInteractivePlans`, `TestNativePiRuntimes`, `TestRealStoreIsolation` | driven |
| 10 | Only Serviceable; Unknown/Limited/Unreachable/invalid refused with full evidence | `plan.Build` → `Serviceable`, `LimitedError`; `TestProviderVerdicts` | driven |
| 11 | Reader error terminal, no retry/default/downgrade; single calls | `plan.Build` → `Availability`; `TestProviderReadError`, `TestTaggedAdmissionRefusals` | driven |
| 12 | Determinate absence admits; corruption refuses; profile/model-group isolation | `plan.Build` → `DefaultDeps` → real `Store.AvailabilityFor`; `TestRealStoreIsolation` | driven |
| 13 | All three adapters at real native-Pi tag | v0.5.11 pinned, no replace; 3 env rows + 3 Pi-runtime goldens at `plan.Build` | driven |
| 14 | Both main tracked/untracked paths use current cwd / `os.Environ` | main still `not_implemented`; owned by TASK-260908-1o7i8y | stated bound |
| 15 | Other module admission classes (model not driven, unsupported mode, unresolved vendor) | `plan.Build` → `BuildLaunch` propagation gate; `TestModelNotDrivenBySystem` (real mismatch fixture + P13 mutant), `TestUnresolvedVendorScope` (real muse fixture), `TestInteractiveDeclaredForMappedSystems` (executable Interactive bound) | driven |
| 16 | Complete publication, independent Astra-medium acceptance, signed delivery | next CR published via this handoff; acceptance pending | pending |

## Negative evidence: 13/13 narrowing mutants killed

`python3 .scripts/plan-mutants.py .temp/TASK-260908-2so46q/mutants-04` exit 0; every
mutant exit 1 with its named behavioral test failing (`summary.tsv` + per-mutant logs in
`mutants-04/`, gitignored scratch). P13 log shows `--- FAIL: TestModelNotDrivenBySystem`
+ `EXIT: 1`. P1–P12 anchors still unique, still killed. No production gate scans source
text; token-preserving source-checker attack remains inapplicable per rev1 verdict.

## Fresh validation (this producer, exact commands/exits observed)

| Command | Exit | Evidence |
|---|---|---|
| `go test ./internal/plan -count=1 -v` | 0 | all suites PASS incl. `TestModelNotDrivenBySystem`, `TestUnresolvedVendorScope`, `TestInteractiveDeclaredForMappedSystems`, `TestNativePiRuntimes/{pi-anthropic,pi-openai,pi-google}` |
| `python3 .scripts/plan-mutants.py .temp/TASK-260908-2so46q/mutants-04` | 0 | 13/13 killed |
| `go test ./internal/plan -count=1` | 0 | ok |
| `go test ./internal/plan -count=1 -race` | 0 | ok (narrow package race, not full suite) |
| `go vet ./internal/plan`, `go build ./internal/plan`, `gofmt -l internal/plan/` (empty), `git diff --check` | 0 | clean |
| `go list -m github.com/relux-works/skill-agents-management` | v0.5.11, no replace | pre-existing pin, untouched |

NOT run by this producer: `make check` (full suite). Per plan-review-r1-rework.md the
configured runtime finalization runs the full suite once; per rev1 verdict there was a
prior manual full suite plus the runtime full suite (duplicate). This run does not add a
third: narrow package gates only. No hosted CI, installs, daemon restarts, real `ax`,
runtime-home edits, LOGBOOK/control-root writes, or branch operations. Candidate left
UNCOMMITTED; `git status` shows only the pre-existing `M README.md/go.mod/go.sum` plus
untracked `internal/plan/` and `.scripts/plan-mutants.py`; no commit, tag, or release.

## Real-tag provenance (accepted, not rerun)

Real v0.5.11 tag/signature/API/registration evidence is accepted from rev1 review
(tag object 0ea486e46765ecf12fff7c2ac526e12da02e95ed, peeled a2a6e9f377f62a5872d99ecdfff0d1690e385f2a,
PR23 MERGED head) and the prior native-pi-validation outcome — not independently
re-fetched here. This run's new module-behavior probes (mismatch sentinel, mapped-system
Interactive modes, muse system-unregistered vs vendor-unresolved order) were observed
fresh against the pinned v0.5.11 module in this worktree (exit 0 probes, outputs in run
log). No pseudo-version, replace, workspace override, or agent-created tag.
