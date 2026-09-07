# TASK-260908-2kapmh — native Pi upstream implementation (STORY-260908-3lmnfs)

Worktree `.temp/STORY-260908-3lmnfs/worktree`, branch `task-board/story/STORY-260908-3lmnfs`,
base `12f443d` (= `origin/main` = `v0.5.10`, `git rev-list --count HEAD..origin/main` = 0).
Candidate left UNCOMMITTED for the handoff snapshot: 21 files changed, 381 insertions, 49 deletions,
3 new files. Implements accepted TASK-260908-ggxfte rev2 §3 verbatim, additive only.

## Delta (files / APIs)

| Area | File | Change |
| --- | --- | --- |
| Core | `pkg/agentic/system.go` | `LaunchRequest.Vendor string` beside `Runtime`, same contract (opaque, BuildLaunch-owned) |
| Core | `pkg/vendorplugin/spawn.go` | `launch.Vendor = string(binding.VendorID)` right after `launch.Runtime`, after `Spawn`+fidelity check; system-only bindings get `""` |
| Plugin | `pkg/agentic/systems/pinative/{pinative,args,binary,env}.go` | id `pi-native`; modes `{DryRun, Interactive}`; `EffortTransportArgv`; `GrammarNone`; `HomeEnvVar=PI_CODING_AGENT_DIR`, `DefaultHome=~/.pi/agent`; binary `launchenv.LookPath(env,"pi")`; argv `["--model", Vendor+"/"+LaunchIdentity]` + `["--thinking", Effort]` when non-empty; `ErrVendorMissing` on empty vendor; NOT `Preflightable`; env = run-context passthrough; stdin detached |
| Runtimes | `pkg/vendorplugin/runtime.go` | frozen rows `pi-anthropic`/`pi-openai`/`pi-google` = (`pi-native` × vendor), `Broker.Checked=["pi 0.84.2 pi-ai providers/data/<vendor>.json"]` |
| Vendors | `vendors/{anthropic,openai,google}/models.go` + `*.go` blank import of `systems/pinative` | `pi-native` added to `Systems` of catalog-verified rows (see provenance) |
| Tests moved | `admissionpin_test.go` (stub list +`pi`), `runtime_test.go` (9 seeded rows), `tools/.../vendors_test.go` (9 rows), `openai_astra_alias_test.go` (two mutants now mirror systems per `sameSystems`), `sourceport_test.go` (compares only source-known systems + narrowing test) | |
| Regress | `internal/regress/{interactive,parity_smoke,localmodels_conditional_registration}_test.go` | `pi-native` interactive case (exec-less); smoke guard rule extracted to `smokeCoverageGap` with its own narrowing test; isolated fake registry gains `pi-native` |
| Docs | `README.md`, `docs/architecture.md`, `docs/consuming-the-module.md`, `docs/shipped-state.md` §7 | plugin, runtimes, launcher contract, limitations, operator tag handoff |

Untouched: `pkg/agentic/systems/pi`, `vendors/local-models`, `pkg/localruntime`, every existing registry admission, goldens, group table, v2 snapshot.

## Catalog provenance (bytes read on this host, no model call, no credential)

`pi --version` = `0.84.2`. Files under
`/opt/homebrew/lib/node_modules/@earendil-works/pi-coding-agent/node_modules/@earendil-works/pi-ai/dist/providers/data/`:

| File | sha256 | Registry ∩ catalog → names `pi-native` | Registry rows absent from catalog (refused `ErrModelNotDrivenBySystem`) |
| --- | --- | --- | --- |
| anthropic.json (api `anthropic-messages`, 13 ids) | `e22c277e3a1ffddc3d2701b72787c9e0bd67b835de6b4cb806677b6b6a89a2f7` | claude-fable-5, claude-opus-5, claude-opus-4-8, claude-sonnet-5, claude-opus-4-6, claude-sonnet-4-6, claude-haiku-4-5, claude-haiku-4-5-20251001 (8) | claude-fable-5-1 |
| openai.json (api `openai-responses`, 38 ids) | `47746dfe79d92a7da58d7fae2c02a5d481b2eba2f5a914a42792552940b78b78` | gpt-5.6-sol, gpt-5.6-terra, gpt-5.6-luna, gpt-5.5, gpt-5.4, gpt-5.4-mini, gpt-5.3-codex-spark, gpt-5.3-codex, gpt-5.2 (9) | gpt-6-astra, astra (alias mirror), gpt-5.2-codex, gpt-5.1-codex-max, gpt-5.1-codex-mini |
| google.json (api `google-generative-ai`, 22 ids) | `c6d9822f7eda23cfeb7d29caee0baa7d2fa7bf3561252734a17428937fbcbfd1` | gemini-3.1-pro-preview, gemini-3.5-flash, gemini-3-flash-preview, gemini-3.1-flash-lite, gemini-2.5-pro, gemma-4-31b-it, gemma-4-26b-a4b-it (7) | all antigravity-only `gemini-3.x-*-{high,medium,low}` rows |

**Deviation, stated:** the design's §3.3 example list named a subset (anthropic 5+aliases, openai 2,
google 1). The brief instructs "check current installed catalog bytes before changing the listed
memberships"; the rule the design settled is "ids present in the 0.84.2 built-in catalog", and the
installed bytes contain more registry ids than the example list. Membership was set to the full
registry ∩ catalog intersection (24 rows) under that rule. `TestPiNativeMembershipMatchesTheInstalledCatalog`
re-reads the bytes on every run where Pi is installed (skips loudly otherwise, `PI_AI_PROVIDER_DATA_DIR`
override) and fails in both directions (native-but-absent, present-but-not-native). If the reviewer wants
the narrower example list, it is a per-row edit with no other consequence.

## Gates (standalone processes, real exit codes; logs in `.temp/STORY-260908-3lmnfs/logs/`)

| Command | Exit | Log |
| --- | ---: | --- |
| `make vet` | 0 | make-vet-02.log |
| `make regress` (`env -u TASK_BOARD_DIR`) | 0 | make-regress-02.log |
| `make test` (25 packages ok, 0 FAIL) | 0 | make-test-02.log |
| `gofmt -l pkg internal tools` | 0, empty | — |

Run 01 of each gate (before the regress guard refactor and docs) also exited 0. The configured suite
is expected to run once more at handoff; nothing above is accepted from prior evidence.

## AC coverage: 24 of 24 rows driven through production entry points

Production call sites: `vendorplugin.BuildLaunch` (`spawn.go`, Vendor set at the `launch.Runtime` site;
preflight gate at the `Preflightable` assertion), `agentic.BuildPlan` (`plan.go`),
`providerlimits.Store.AvailabilityFor` (`verdict.go`) / `Store.Observe` (`state.go`), launcher SPEC §4.4.

| # | AC row | Named test (package) | Entry point |
| --- | --- | --- | --- |
| 1 | golden pi-anthropic × claude-opus-5 × high: binary `pi` on launch PATH, argv `["--model","anthropic/claude-opus-5","--thinking","high"]`, stdin detached, env passthrough incl. `PI_CODING_AGENT_DIR`, Home | `TestPiAnthropicGoldenPlan` (vendorplugin) | BuildLaunch |
| 2 | effort-none row → `["--model","google/gemini-3.1-pro-preview"]` | `TestPiGoogleEffortNoneRowCarriesNoThinkingFlag` | BuildLaunch |
| 3 | alias row launches target with vendor prefix | `TestAnAliasRowLaunchesItsTargetQualified` (mutant admits astra pair), `TestAnAliasLaunchesItsTargetWithTheVendorPrefix` (pinative) | BuildLaunch / BuildPlan |
| 4 | sweep every pi-* × model: no bare id, every `--model` value prefixed | `TestEveryPiRuntimeModelLaunchesProviderQualified` — drove 24 pairs | BuildLaunch |
| 5 | `Vendor=""` → `ErrVendorMissing`, never a bare id (both modes, and plugin held directly) | `TestAMissingVendorIsRefusedNeverDowngradedToABareID` | BuildPlan / `pinative.Args` |
| 6 | vendor `Spawn` setting Vendor/Runtime is overwritten; narrowing shows the spoof reaches argv on the unguarded path | `TestBuildLaunchOverwritesAVendorSetBySpawn`, `TestTheSpoofWouldHaveReachedArgvWithoutTheOverwrite` | BuildLaunch |
| 7 | exec-marker sweep = 0 on pi-native; fires on `systems/pi` exec argv | `TestNoArgvElementIsEverABareModelID`, `TestTheSweepFiresOnTheWrapperExecArgv` (pinative); regress `TestAnInteractivePlanCarriesNoExecMarkerForAnyMappedSystem/pi-native` | BuildPlan |
| 8 | `ultra` → `ErrEffortNotInVocabulary` | `TestBuildLaunchRefusalsOnPiRuntimes/effort_outside_the_vocabulary` | BuildLaunch |
| 9 | empty effort on required row → `ErrEffortMissing` | `…/empty_effort_on_a_required_row` | BuildLaunch |
| 10 | empty model → refused | `…/empty_model` (BuildLaunch answers `ErrUnknownModel`, the pre-existing 93abeae behaviour); `TestAnEmptyModelIsRefusedBeforeArgv` (pinative, `ErrModelMissing`); regress `TestAPlanRefusesAnEmptyModelForEveryMappedSystemInEveryMode/pi-native` | BuildLaunch / BuildPlan |
| 11 | exec → `ErrUnsupportedLaunchMode` | `…/exec_mode`, `TestExecIsNotDeclaredAndIsRefused` | BuildLaunch / BuildPlan |
| 12 | composition → `ErrCompositionNotInteractive` | `…/composition`, `TestACompositionIsRefusedOnTheInteractiveLaunch`; regress `…RefusesAComposition…/pi-native` | BuildLaunch / BuildPlan |
| 13 | `gpt-5.5` on pi-anthropic → `ErrUnknownModel` | `…/another_vendor's_model` | BuildLaunch |
| 14 | `claude-fable-5-1` → `ErrModelNotDrivenBySystem` (+ `gpt-6-astra`, `astra`, antigravity-only google row) | `…/a_row_absent_from_the_Pi_catalog` and siblings | BuildLaunch |
| 15 | no preflight: `Preflightable` assertion false (wrapper `pi` true, so the assertion discriminates); plan builds with no `agents-infra` on PATH | `TestNoPreflightIsWiredAndAPlanBuildsWithoutTheWrapper` | BuildPlan (BuildLaunch gate at `spawn.go` Preflightable assertion) |
| 16 | F1: limited record on `pi-anthropic-unmapped:claude-opus-5` at (pi-anthropic, home) → Limited, `Checked` ≠ frozen table | `TestALimitedRecordOnPiAnthropicReadsAsLimitedFromItsOwnHome` (providerlimits) | AvailabilityFor |
| 17 | F1 narrowing: same store/record/home under an UNFROZEN runtime id → Healthy via carve-out (reviewer's observed failure reproduced) | `TestAConsumerDeclaredPiRuntimeWouldTakeTheCarveOut` | AvailabilityFor |
| 18 | different home → Healthy; `claude` runtime on the same home → Healthy | `TestTheSameRecordIsInvisibleFromAnotherPiHome` | AvailabilityFor |
| 19 | write side: `ClassifyManaged("pi-anthropic")` mints a class, `Observe` records under the runtime id | `TestObserveAdmitsPiAnthropicThroughTheProductionWriteGate` | Observe |
| 20 | managed home resolves from the plugin declaration (`PI_CODING_AGENT_DIR`, else `~/.pi/agent`) | `TestTheManagedHomeResolvesFromThePluginDeclaration` | DefaultProviderHomeIn |
| 21 | alias-mirror refusal for one-sided `pi-native` | `TestAliasMirrorRefusesAOneSidedPiNativeMembership` | Registry.Register |
| 22 | frozen rows and provenance; `broker_migration_test`/`TestBrokerFactsAgreeWithTheFrozenTable`/`TestEveryFrozenRuntime*` extend | `TestThePiRuntimesAreFrozenWithTheirCatalogProvenance`; existing tests over 9 rows in `make test` | FrozenRuntimes |
| 23 | installed-Pi no-secret probe | `TestPiNativeMembershipMatchesTheInstalledCatalog` (ran, not skipped, on this host) | vendor `Models()` |
| 24 | legacy `systems/pi` + local-models parity | `pkg/agentic/systems/pi`, `vendors/local-models`, regress class 6 all ok in `make test`/`make regress`; no file in those packages changed | — |

Stated bounds (not driven, by design): `pi-google` can never be suppressed (google ships no
classifier) — pinned positively by `TestPiGoogleStaysUnclassifiableAndIsNeverWritten`; the thinking
word is transported verbatim and not clamped to what installed Pi accepts (invariant 4); no Pi process
is launched by any test (brief: no model calls); the empty-model refusal through BuildLaunch is
`ErrUnknownModel`, not `agentic.ErrModelMissing` (vendor lookup runs first).

## Mutants (logs `mutants-01.log`, `mutants-02.log`; each restored after the run, tree diff unchanged)

| Mutant | Narrows the gate to | Named failing test | Exit |
| --- | --- | --- | ---: |
| M1 `args.go`: empty Vendor emits a bare id instead of refusing | admits exactly the empty-vendor request | `TestAMissingVendorIsRefusedNeverDowngradedToABareID` | 1 |
| M2 `spawn.go`: Vendor set only when Spawn left it empty | admits exactly a spoofing vendor | `TestBuildLaunchOverwritesAVendorSetBySpawn` | 1 |
| M3 `runtime.go`: `pi-anthropic` row kept, vendor rebound to `google` | row present, wrong broker | `TestPiAnthropicGoldenPlan`, `TestThePiRuntimesAreFrozenWithTheirCatalogProvenance`; providerlimits `TestALimitedRecord…`, `TestAConsumerDeclared…` | 1 / 1 |
| M3b `runtime.go`: `pi-anthropic` row deleted (delete-only; existence proof, not accepted as the bound) | — | `TestALimitedRecordOnPiAnthropicReadsAsLimitedFromItsOwnHome` | 1 |
| M4 `anthropic/models.go`: `claude-fable-5-1` names `pi-native` | admits exactly one non-catalog member | `TestBuildLaunchRefusalsOnPiRuntimes/a_row_absent_from_the_Pi_catalog`, `TestPiNativeMembershipMatchesTheInstalledCatalog` | 1 |
| M5 `pinative.go`: `LaunchModeExec` added, interactive kept | admits exactly exec | `TestExecIsNotDeclaredAndIsRefused`; `…/exec_mode` | 1 / 1 |
| M6 `args.go`: `EffortTransportArgv` and the `--thinking` token kept in source, flag never emitted (token-preserving) | declaration without construction | `TestTheArgvTransportIsActuallyCarried` | 1 |
| M7 `sourceport_test.go`: filter drops every system, not only post-port ones | over-filter | `TestEverySourceModelRowIsPorted` (42 rows reported). **Bound:** `TestTheSourcePortPinStillSeesASourceSystemDropped` alone does NOT discriminate this mutant (exit 0 against it); it guards under-reporting, the full-set pin guards over-filtering | 1 (vs full-set pin) |
| M8 `parity_smoke_test.go`: smoke exemption widened to any system with an interactive case | admits exec-capable systems | `TestTheSmokeExemptionIsOnlyForExecLessSystems` | 1 |

No surviving mutant. No gate added by this task inspects source text (no argvguard signature was
registered for `--thinking`); M6 is the token-preserving attack on the behavioural surface.

## Deliberate limitations

- Interactive/dry-run only; no exec, managed-session, goal, budget, tier or composition for `pi-native`.
- No `pi-*` group rows; no Pi failure-output classifier; `pi-google` stays carve-out.
- `claude-fable-5-1`, `gpt-6-astra`/`astra`, `gpt-5.2-codex`, `gpt-5.1-codex-max`, `gpt-5.1-codex-mini`
  do not drive `pi-native` (absent from the 0.84.2 catalog).
- Regress smoke guard now excuses ONLY exec-less systems that have an interactive case
  (`smokeCoverageGap`), because a golden is a capture of the extraction source's headless launch and
  native Pi has none; narrowed by M8.
- No LOGBOOK entry (operator instruction: no LOGBOOK/ordinary files in the control root); this outcome
  and the board notes are the record.

## Operator handoff

1. Parent: signed branch → PR → actual comment review → local `make vet/test/regress` → exact-head
   fast-forward to `main` (no GitHub merge/squash/rebase).
2. **Operator only**, after the head is on `main` and verified: `git tag -s v0.5.11 <landed sha>` (signed,
   annotated), `git verify-tag v0.5.11`, push the tag. Nothing here creates it; no pseudo-version or
   committed `replace` was added. The consumer launcher's `go.mod` then requires `v0.5.11` (pre-tag
   development may use an ignored `go.work`).
3. Provider login in the managed `PI_CODING_AGENT_DIR` home via Pi's `/login` (no credential touched here).
