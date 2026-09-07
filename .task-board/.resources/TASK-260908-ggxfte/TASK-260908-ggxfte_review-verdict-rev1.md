# TASK-260908-ggxfte — reviewer verdict, CR rev1

**Verdict: changes requested → `analysis`.** Read-only review; no repo edits, no LOGBOOK, no tags, no ax.

## Empty repository delta is correct

This leaf is analysis: its deliverable is the outcome resource, and the brief forbids
launcher/module/spec edits. Zero paths in the patch is the right shape. Not a reason to reject.

## Verified (exact sources, own probes)

- `v0.5.10^{commit}` = `12f443d` = `origin/main` (local `main` is stale at `abe11ec`; the
  producer's "= main HEAD" holds for the remote). Nothing unreleased on main.
- `pkg/agentic/systems/pi`: binary `agents-infra`, interactive argv `["pi","--model",M]`,
  `EffortTransportNone`, `HomeEnvVar ""`; `Preflightable`, and `vendorplugin/spawn.go:221-229`
  runs Preflight for every non-dry-run mode → `agents-infra runtime status` (`localruntime/client.go:116`).
  A0 E1/E2 hold.
- relux-agents-infra `dee5403` `pi_launch_posix.go:135-139` refuses inbound `PI_CODING_AGENT_DIR`
  (`pi_execution_environment_invalid`); `:332` sets its own. Wrapper is unusable for a Curator home.
- Pi 0.84.2 (`/opt/homebrew/lib/node_modules/@earendil-works/pi-coding-agent`): `--model/--provider/--thinking`
  in `cli/args.js`; `getAgentDir` honours `PI_CODING_AGENT_DIR`; anthropic catalog has no `claude-fable-5-1`;
  `clampThinkingLevel` silently nearest-fits; anthropic vocabulary `{low,medium,high,xhigh,max}` ⊂ pi levels.
- Native home non-secret: `defaultProvider=llama-cpp`, `defaultModel=qwen3.8-27b`, thinking `medium`,
  `models.json` providers `[llama-cpp]`, `auth.json` 2 bytes (size only, never read).
- Design skeleton (new `pi-native` plugin, muse-shaped binary lookup, vendor-row `Systems` membership,
  `DeclareRuntime` seeding, `checkAliases` mirror rule, `ErrUnknownModel` via `selectModel`) matches
  the module's extension contracts. Producer probe logs re-read: `systems/pi` and `vendorplugin` subsets exit 0.
- Own negative probes, fresh throwaway home, `--offline`: `--model claude-fable-5-1` → exit 1;
  `--thinking ultra` → exit 1; `claude-opus-5:ultra` → exit 1. No model call reached.

## Findings requiring rework

### F1 — provider-limits claim is false for a declared runtime (blocks SPEC §4.4)
`providerlimits.brokerForRuntime` (`groups.go:103`) walks **`vendorplugin.FrozenRuntimes()` only**.
A runtime seeded via `DeclareRuntime` (`pi-anthropic`) resolves to broker `""`, so
`AvailabilityFor` (`verdict.go:114-126`) takes the `unclassifiableVerdict` carve-out: always
Healthy, `Checked=[frozen table]`, no identity, no state read. Also `IdentityKey` keys on the
**runtime id** string, not the vendor (`identity.go:76`, `verdict.go:135`). So report §3.3 ("brokerForRuntime
then keys state (vendor, managed home); frozen ids untouched") and the §4 test
("`AvailabilityFor{Runtime:"pi-anthropic",Home}` → identity (anthropic, home)") are wrong as written:
the test would observe the carve-out, and SPEC §4.4's admission ("a verdict that is not observed
healthy is not serviceable ... keyed by (provider, home)") becomes vacuous for every native Pi launch.
Rework: the design must name the providerlimits change — either frozen rows for `pi-*` (accepting the
digest/`crossbinary` consequences it already flags) or an additive `brokerForRuntime` that also consults
declared runtimes' `Vendor` — and restate the test as a negative: a limited record under the managed
home's identity must refuse `pi-anthropic`, proven by narrowing, with the `AvailabilityFor` call site named.

### F2 — bare `--model <id>` is ambiguous for every proposed id (probe-backed)
Fresh home, no auth: `pi --model claude-opus-5` → exit 1, "ambiguous across providers: anthropic,
cloudflare-ai-gateway, github-copilot, opencode. No matching provider is authenticated." `gpt-5.5`
is in 6 provider files. `resolveCliModel:340-370` admits a bare id only when exactly one matching
provider is authenticated. Consequences for the recommended argv: a freshly seeded managed home (the
launcher's own §7.4 case, no cloud auth) cannot even reach Pi's `/login`; a home with Copilot or
OpenCode auth added later exits 1 on every launch. Report §6 defers this as "`--provider` would need
a vendor id on LaunchRequest" without the evidence that all candidate ids are multi-provider.
`agentic.LaunchRequest`/`agentic.Model` carry no vendor today (`system.go`), so this is a required
additive upstream change, not deferrable: carry the runtime's vendor to the plugin (e.g. a `Vendor`
field populated by BuildLaunch from the declaration) and spell `<vendor>/<id>` — pi's provider ids
`anthropic`, `openai`, `google` match the vendor ids exactly. Golden and a negative test
(argv never bare id) belong in §4.

### F3 — `pi-native` id is not a human product decision
The goal fixes both constraints: native Pi from the managed home **and** the wrapper-backed
local-model runtime preserved. Option (b) violates the second by construction, so (a) is the only
backward-compatible implementation and needs no operator input. Move it from "Unavoidable" to the
implied-defaults list; the real operator-only items are the `v0.5.11` tag and provider login.

## Minor
- State ref identities as `v0.5.10^{commit} = origin/main = 12f443d`.
- §4 "no preflight" test should also assert `pi-native` does not satisfy `agentic.Preflightable`
  (type assertion), since `spawn.go:222` gates on the interface, not on capabilities.

## Routing
`analysis` (producer archetype analyst). F1 and F2 change the proposed upstream delta and its test
list; F3 changes the decision section. Everything else may stand as is.
