# TASK-260908-ggxfte — native Pi upstream contract, rev2 (F1–F3 rework)

Read-only research. No LOGBOOK (operator instruction; this outcome is the equivalent evidence), no
repo/spec edits, no tags, no ax. Refs: `v0.5.10^{commit} = origin/main = 12f443d` (skill-agents-management);
relux-agents-infra `dee5403`; Pi 0.84.2 (`/opt/homebrew/lib/node_modules/@earendil-works/pi-coding-agent`).
Sections 1–2 of rev1 (gap, wrapper refusal, Pi parser/catalog/thinking, native-home non-secret defaults)
stand as verified by the reviewer and are not repeated. Confidence: high (source + bounded probes).

## 3. Minimal upstream design (additive; one field, three frozen rows, one plugin)

1. **New plugin `pkg/agentic/systems/pinative`, id `pi-native`.** Binary `launchenv.LookPath(env,"pi")`
   (muse `binary.go` shape). Capabilities: modes `{DryRun, Interactive}`, `EffortTransportArgv`,
   `GrammarNone`, no goal/budget/tier, `HomeEnvVar "PI_CODING_AGENT_DIR"`, `DefaultHome "~/.pi/agent"`.
   Does **not** implement `agentic.Preflightable` (`vendorplugin/spawn.go:222` gates on the interface).
   Argv: `["--model", <Vendor>+"/"+LaunchIdentity]` + `["--thinking", Effort]` when non-empty. Never a
   bare id (F2). `ChildEnv` = run-context passthrough. `systems/pi` untouched.
2. **`agentic.LaunchRequest.Vendor string`** (`pkg/agentic/system.go`, beside `Runtime` at :456), same
   contract as `Runtime`: opaque string, set unconditionally by `BuildLaunch` (`spawn.go:191`,
   `launch.Vendor = string(binding.VendorID)`), never by a vendor `Spawn`. System-only bindings (muse)
   get `""`; `checkLaunchFidelity` needs no change because the field is overwritten after `Spawn`.
   `passthroughLaunchRequest` (`plugin.go:72`) unchanged. Existing plugins never read it (Profile pattern).
   Pi provider ids `anthropic|openai|google` equal `VendorID` spellings, so no mapping table.
3. **Vendor rows add `"pi-native"` to `Systems`** (`vendors/anthropic/models.go:128…198`, openai, google)
   for ids present in the 0.84.2 built-in catalog: anthropic opus-5, opus-4-8, sonnet-5, haiku-4-5,
   fable-5 (+ alias rows, e.g. `claude-haiku-4-5-20251001`, mirror per `vendor.go:837 sameSystems`);
   openai gpt-5.5, gpt-5.3-codex; google gemini-3.1-pro-preview. **Not** `claude-fable-5-1` (absent
   from the catalog; `anthropic/claude-fable-5-1` would take the custom-model fallback with a warning —
   an unverified launch, refused as `ErrModelNotDrivenBySystem`). Effort vocabulary stays the row's.
4. **Three frozen rows** in `vendorplugin/runtime.go frozenRuntimes`: `pi-anthropic`/`pi-openai`/
   `pi-google` = (`pi-native` × vendor), `Broker.Checked: ["pi 0.84.2 pi-ai providers/data/<vendor>.json"]`.
   Frozen, not consumer-declared, because **F1**: `providerlimits.brokerForRuntime` (`groups.go:103`)
   walks `FrozenRuntimes()` only, and identity.go/ladder.go/broker_migration_test share that single
   source. A frozen row makes `HasClassifierForRuntime("pi-anthropic")=true`, so `AvailabilityFor`
   (`verdict.go:114`) skips the `unclassifiableVerdict` carve-out, resolves
   `IdentityFor("pi-anthropic", managedHome)` — identity is **(runtime id, home)** (`identity.go:76`), not
   (vendor, home) — reads state, and returns Limited for a limited group record. Model groups resolve to
   `pi-anthropic-unmapped:<model>` singletons (`DefaultGroups` rows key on `claude`/`codex`); no group
   rows added — anthropic/openai classifiers are broker-keyed (D4) and apply, but Pi's failure
   output has not been captured, so a `pi-*` group row would assert an unexercised mapping.
   Rejected alternative: additive `brokerForRuntimeIn(registry)` — needs a registry threaded into
   `Store`/`VerdictQuery` and a second lookup path; larger and touches the frozen-only invariant the
   package doc states.
5. Docs: `architecture.md` system table (:71) and §Runtimes (:151); `consuming-the-module.md`
   interactive paragraph; `shipped-state.md`.

Consequences checked: admitted-pair digests (invariant 3) are per runtime over `RuntimeModels`
(`v2snapshot.go:301`, `DrivenBy(runtime.SystemID)`) — `claude`/`codex` sets unchanged; `pi-*` are not in
the v2 snapshot (`V2SnapshotTiers` false → `ExpandCeiling` refuses; interactive launch does not use it).
`crossbinary_test` pins existing ids only. `TestEveryFrozenRuntimeBuildsThroughTheOneLaunchEntryPoint`
stub list (`admissionpin_test.go:626`) must add `"pi"`; `TestEveryFrozenRuntimeWithAVendorResolvesThroughTheDefaultRegistry`
requires item 3. Launcher (parent): SPEC §4.2 `pi → pi-native`; runtime = frozen `pi-*` row whose
vendor `Models()` carry the id and `pi-native`; `BuildLaunch(…, Interactive)`; SPEC §4.4 verdict via
`AvailabilityFor{Runtime:"pi-<vendor>", Home:<managed>, Model}`; no argv rebuild.

## 4. Required behavioural tests (upstream)

- Golden (`pi-anthropic`, opus-5, high): `Binary`=`pi` on launch PATH; argv exactly
  `["--model","anthropic/claude-opus-5","--thinking","high"]`; effort-none row → `["--model","google/…"]`;
  alias row launches the target identity with the vendor prefix; stdin detached; env passthrough; Home.
- **Negative, argv never bare:** sweep every `pi-*` row × model: no argv element equals `model.ID` or
  lacks `<vendor>/`. Narrowing proof: build with `Vendor=""` → plugin refuses (`ErrVendorMissing`),
  not a bare id. `checkLaunchFidelity`-style test: a vendor `Spawn` setting `Vendor` is overwritten.
- Marker sweep (`spawn --profile --prompt --deadline --result-schema -p --print --mode exec
  --output-format --dangerously-* --max-budget-usd`) = 0; shown firing on `systems/pi` exec argv.
- Gate narrowing via `BuildLaunch(pi-anthropic)`: `ultra`→`ErrEffortNotInVocabulary`; empty effort on a
  required row→`ErrEffortMissing`; empty model→`ErrModelMissing`; exec→`ErrUnsupportedLaunchMode`;
  composition→`ErrCompositionNotInteractive`; `gpt-5.5`→`ErrUnknownModel`; `claude-fable-5-1`→
  `ErrModelNotDrivenBySystem`.
- No preflight: `_, ok := pinative.System.(agentic.Preflightable)` is false; plan builds with no
  `agents-infra` on PATH.
- **F1 negative (providerlimits):** fixture store, `IdentityFor("pi-anthropic", home)`, `Observe`/limit
  record for group `pi-anthropic-unmapped:claude-opus-5` → `AvailabilityFor{Runtime:"pi-anthropic",
  Home:home, Model:"claude-opus-5"}` = Limited, `Checked` names the state source, not
  `SourceFrozenTable`. Narrowing: drop the frozen row → same query returns Healthy via the carve-out
  (the reviewer's observed failure), proving the row is load-bearing. Different home → Healthy.
  Production call site: launcher SPEC §4.4 `AvailabilityFor` before `BuildLaunch`.
- `broker_migration_test`/`TestBrokerFactsAgreeWithTheFrozenTable` extend automatically; alias-mirror
  refusal; regression `systems/pi` + `local-models` suites unchanged.

## 5. Operator-only prerequisites

- Signed tag `v0.5.11` after the change lands on main; launcher `go.mod` requires it; SPEC §7 names it.
- Provider login in the managed Pi home. Probe (`pi-probe-provider-qualified-01.log`, fresh home, no
  auth, `--offline -p`): `anthropic/claude-opus-5` resolves and stops at "No API key found for
  anthropic … /login" (exit 1 in print mode; interactive reaches `/login`); bare `claude-opus-5` exits 1
  "ambiguous across providers"; `openai/claude-opus-5` warns and falls to custom-model id. No model call.

## 6. Decisions

None unavoidable. `pi-native` as the new id with `pi` left as the wrapper is implied by the goal (native
Pi **and** local-model runtime preserved; rebinding `pi` breaks `local-models.toml system="pi"` and
frozen local-qwen goldens). Implied defaults: three vendors, catalog-verified ids only,
provider-qualified argv, thinking clamp left as harness residual (muse `max` stance).

## 7. Checklist evidence (analysis-only leaf)

Items "Implementation matches AC" / "fits architecture": this outcome maps every AC clause (released/main
support §1 rev1, source proposal §3, runtime/model/effort semantics §2–3, tests §4, tag §5, decision §6);
design is additive on existing extension contracts. "Tests green": no repository code changed; rev1
narrow probes (`systems/pi`, `vendorplugin` `-run` subsets) exited 0, Pi probes above exited as recorded.
"Attacked, not read": bare-id ambiguity, wrong-prefix fallback and no-auth stop were exercised on Pi 0.84.2;
the F1 carve-out failure was reproduced from source and each proposed gate ships a narrowing negative in §4.
Review rev1 verdict is attached (`_review-verdict-rev1.md`); F1–F3 reworked here and routed back to review.
