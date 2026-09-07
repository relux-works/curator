# TASK-260908-ggxfte — reviewer verdict, CR rev2

**Verdict: accepted (`accept_cr` revision 2).** Read-only review; no repo edits, no LOGBOOK, no tags, no ax.

## Empty repository delta is correct

Analysis leaf, brief forbids launcher/module/spec edits. The deliverable is `TASK-260908-ggxfte_results.md`;
zero paths in the patch is the right shape. Refs re-resolved: `v0.5.10^{commit} = origin/main = HEAD = 12f443d`
in skill-agents-management, tree clean.

## F1 (providerlimits routing) — verified fixed

- `pkg/providerlimits/groups.go:103 brokerForRuntime` walks `vendorplugin.FrozenRuntimes()` only;
  `HasClassifier` (`groups.go:362`) is true for `anthropic|openai`. A frozen `pi-anthropic` row with
  `Vendor: "anthropic"` therefore makes `unclassifiableVerdict` (`verdict.go:174`) return `ok=false`, so
  `AvailabilityFor` (`verdict.go:114`) proceeds to `IdentityFor(runtime, home)` — keyed on the **runtime id**
  (`identity.go:137,149`), i.e. (runtime, home) as the rework brief requires — then `LoadIdentityState`,
  then `table.Group(runtime, model)` (`groups.go:348`, unmapped → `pi-anthropic-unmapped:<model>`), then
  `groupVerdict` for an existing record. The §4 F1 negative (limited record → Limited; drop the row →
  Healthy via carve-out; other home → Healthy) is exactly what this code path does. Correct.
- `pi-google` resolves to broker `google`, which has no classifier → still carve-out. The report does not
  claim otherwise (it names anthropic/openai classifiers only). Acceptable, not hidden.
- Frozen-row consequences checked: `RuntimeDeclaration.Validate` requires non-empty `Broker.Checked`
  (`runtime.go:223`) — the design supplies it. No test pins `len(FrozenRuntimes())`; `broker_migration_test`
  and `TestEveryFrozenRuntime*` iterate the slice, so they extend as the report says. The
  `admissionpin_test.go:626` stub list must gain `pi` — correctly flagged.
- Rejected alternative (registry-threaded `brokerForRuntimeIn`) is rejected for a real reason: the
  package comment on `brokerForRuntime` and `HasClassifierForRuntime` states the frozen-slice-only
  invariant explicitly.

## F2 (vendor carried, provider-qualified argv) — verified fixed

- `agentic.LaunchRequest.Runtime` (`system.go:456`) is set unconditionally at `spawn.go:191` after
  `checkLaunchFidelity` (`spawn.go:359`, which does not inspect Runtime); a sibling `Vendor` field set from
  `binding.VendorID` at the same site follows the established pattern. `passthroughLaunchRequest`
  (`plugin.go:72`) does not need it. No parity test pins LaunchRequest field count (`parity/*_test.go`
  NumField loops are over `Snapshot`). Existing plugins never read it. Muse gets `""`.
- Pi 0.84.2 provider data files are literally `anthropic.json`, `openai.json`, `google.json`, so
  `<VendorID>/<id>` needs no map. Probe log re-read: `anthropic/claude-opus-5` resolves and stops at
  `/login`; bare id exits 1 ambiguous; wrong prefix falls to custom id with a warning. `gpt-5.5` is in 6
  provider files — bare id would never be admissible.
- Catalog membership re-checked in the installed data: anthropic has `claude-opus-5, claude-opus-4-8,
  claude-sonnet-5, claude-haiku-4-5, claude-haiku-4-5-20251001, claude-fable-5`; openai has `gpt-5.5,
  gpt-5.3-codex`; google has `gemini-3.1-pro-preview`; `claude-fable-5-1` appears in 0 files. The proposed
  `Systems` membership list matches, and excluding `claude-fable-5-1` (registry row at
  `vendors/anthropic/models.go:123`) is right — otherwise a launch would take Pi's custom-model fallback.
  Alias mirror rule `sameSystems` exists (`vendor.go:853`), so alias rows must gain `pi-native` too — stated.
- Negative test list (argv never bare; `Vendor=""` → refusal not bare id; `Preflightable` type assertion
  false; gate narrowing via `BuildLaunch`) is adequate and names the production call site
  (`spawn.go:222` for preflight, launcher SPEC §4.4 for `AvailabilityFor`).

## F3 — verified fixed

`pi-native` moved to implied defaults with the correct argument (rebinding `pi` breaks
`local-models.toml system="pi"` and local-qwen goldens). Operator-only items are exactly the signed
`v0.5.11` tag and provider login in the managed home. No product decision remains.

## No hidden semantics

No new request field is read by a vendor Spawn; effort stays the row's vocabulary with argv transport;
no argv rebuild in the launcher; no fallback (unknown model → `ErrUnknownModel`, non-native →
`ErrModelNotDrivenBySystem`); `systems/pi` and local-model runtime untouched. Minor items from rev1
(ref identity spelling, Preflightable assertion) are incorporated.

## Checklist

Implementation matches AC / fits architecture: yes, additive on existing extension contracts.
Tests green: no code changed; producer's narrow `systems/pi`+`vendorplugin` probes exited 0 (accepted
from rev1 evidence, not rerun — nothing in the delta could change them). Attacked, not read: the F1
carve-out and bare-id ambiguity were both reproduced from source/probe before being accepted as fixed.
LOGBOOK: operator no-LOGBOOK instruction; this outcome is the equivalent record.

Handoff: parent (solution-architect/analyst producer) owns the upstream implementation and delivery.
