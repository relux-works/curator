# TASK-260908-2kapmh — reviewer verdict, CR-TASK-260908-2kapmh-1 rev1

**Verdict: changes_requested → `to-dev`.** `repeat-of: none`.
Read-only review of tree `38639918` against base `12f443d`; one temporary probe test written and removed
(tree diff against the candidate empty afterwards). No auth mutation, no model call, no ax, no tags, no LOGBOOK.

## Finding R1 (blocking): a required effort that BuildLaunch admits is silently dropped or rewritten by Pi

The reviewer brief names this exact shape: "a required effort accepted by the module but silently
ignored/downgraded or rejected by Pi is a real mismatch, not a reason to bless a fake-only golden".

Installed Pi 0.84.2 facts (source read on this host):
- `dist/cli/args.js:6` `VALID_THINKING_LEVELS = ["off","minimal","low","medium","high","xhigh","max"]`;
  `args.js:97-108` pushes a WARNING diagnostic for any other word and leaves `result.thinking` unset, so the
  session runs at Pi's default level (`agent-session.js:1329`, `settings.defaultThinkingLevel ?? DEFAULT`).
- `pi-ai/dist/models.js:548-575` `getSupportedThinkingLevels` / `clampThinkingLevel`: a parser-valid word the
  catalog model maps to `null` is silently substituted by the nearest supported level at session creation
  (`agent-session.js:1277`, `main.js:668-669`).

Reproduced (log `pi-probe-thinking-02.log`, isolated `PI_CODING_AGENT_DIR`, `env -i`, `--offline -p`, no secrets):

| argv | Pi 0.84.2 result | exit |
| --- | --- | ---: |
| `--model openai/gpt-5.6-sol --thinking ultra` | `Warning: Invalid thinking level "ultra"…` then `No API key found for openai` (session would proceed at default thinking) | 1 |
| `--model anthropic/claude-opus-5 --thinking high` / `max`, `openai/gpt-5.6-sol --thinking xhigh` / `max`, `google/gemini-3.1-pro-preview` | no warning; stops at `No API key` (/login) | 1 |
| `--model anthropic/claude-fable-5-1 --thinking high` | custom-model fallback warning (confirms the membership exclusion) | 1 |
| `--model claude-opus-5 --thinking high` | `ambiguous across providers` (confirms qualified argv is required) | 1 |

Module side (temporary probe test through the production entry point, log `go-ultra-probe-01.log`):
`BuildLaunch(pi-openai, gpt-5.6-sol, "ultra")` and `(…, gpt-5.6-terra, "ultra")` return `err=nil` with argv
`[--model openai/gpt-5.6-sol --thinking ultra]`. The candidate's refusal test only covers `ultra` on
`pi-anthropic` (`pinative_runtime_test.go:231`), where the ROW vocabulary refuses it; on the two openai rows
whose vocabulary contains `ultra` the gate admits it and the plan is a lie.

Full registry-vocabulary vs Pi-supported comparison for all 24 pi-native rows (`vocab-vs-pi-01.txt`, computed
from the same catalog bytes the producer digested):

| Row | Row vocabulary word | Pi 0.84.2 | Effect |
| --- | --- | --- | --- |
| `gpt-5.6-sol`, `gpt-5.6-terra` | `ultra` | not a parser level | warning, level dropped, default applied |
| `gpt-5.3-codex`, `gpt-5.2` | `minimal` | `thinkingLevelMap.minimal = null` | clamped to `low` silently |

Every other (row × word) pair is supported by Pi for that model (`max` on the anthropic rows, `xhigh`/`max`
on openai are all mapped). Google/haiku rows are effort-none and emit no `--thinking`; Pi then applies its
own settings default on a reasoning-capable model. That last point is a harness default, not a module
injection; it must be STATED in the docs as a bound (currently the docs only say "transported verbatim").

Why this is rework and not a stated bound: the producer lists "not clamped to what installed Pi accepts
(invariant 4)" as a bound, and the accepted design §6 left the clamp as a harness residual. The parent's
reviewer brief supersedes that for this review and asks for an explicit harness refusal where the plan
would otherwise misrepresent the effort. The candidate's golden proves the argv bytes, not that a session
runs at the requested effort; for 4 admitted pairs it provably does not.

### Required rework (precise)
1. An explicit refusal, reached through `vendorplugin.BuildLaunch` (production call site: the `Argv` step
   after effort validation, or a pi-native-specific gate before it), for a (model, effort) pair the installed
   Pi 0.84.2 catalog does not support: `ultra` on `gpt-5.6-sol`/`gpt-5.6-terra`, `minimal` on
   `gpt-5.3-codex`/`gpt-5.2`. No clamp, no default, no change to the rows' global vocabularies (Codex still
   drives `ultra`/`minimal`). The error must name model, runtime, the words native Pi accepts for that model
   and the row's recommendation (invariant 4 error shape).
2. The fact must be catalog-verified like membership: extend `TestPiNativeMembershipMatchesTheInstalledCatalog`
   (or a sibling) to compare the module's per-model unsupported set against `thinkingLevelMap` and the
   parser list in the installed bytes, failing in both directions.
3. Negative tests through `BuildLaunch`: `pi-openai × gpt-5.6-sol × ultra` and `pi-openai × gpt-5.3-codex ×
   minimal` refused; positive goldens that `pi-openai × gpt-5.6-sol × max` and `pi-anthropic × claude-opus-5 ×
   max` still build. Narrowing mutant: the gate stays present and admits exactly `ultra` → a named test fails.
4. Docs (`README.md`, `docs/architecture.md` "transported verbatim", `docs/shipped-state.md` "No thinking-word
   clamp") rewritten to the new contract, plus the stated bound for effort-none rows above.
5. Attach the no-secret probe log for the refused pairs as the producer's own evidence.

## Everything else: verified, would be accepted with R1 fixed

- **Catalog membership**: re-read `anthropic.json`/`openai.json`/`google.json`; sha256 match the producer's
  digests exactly (`e22c277e…`, `47746dfe…`, `c6d9822f…`). All 24 `pi-native` rows are catalog ids; every
  registry row WITHOUT `pi-native` (`claude-fable-5-1`, `gpt-6-astra`, `astra`, `gpt-5.2-codex`,
  `gpt-5.1-codex-max/mini`, all `gemini-3.x-*-{high,medium,low}`) is absent from the catalog. The expansion
  from the design's example list to registry∩catalog is the settled rule applied, not unsupported admission.
- **Vendor propagation**: `spawn.go` sets `launch.Vendor = string(binding.VendorID)` after Spawn and fidelity;
  `TestBuildLaunchOverwritesAVendorSetBySpawn` + the unguarded-path narrowing cover the spoof; muse gets `""`.
- **Argv/binary/env/stdin/home**: `launchenv.LookPath(env,"pi")`, `--model <vendor>/<id>`, `ErrVendorMissing`
  on empty vendor, stdin detached, run-context passthrough, `PI_CODING_AGENT_DIR` / `~/.pi/agent`.
  Not `Preflightable`; plan builds with only a stub `pi` on PATH.
- **Refusals** through BuildLaunch: exec, composition, empty model, other-vendor model, non-catalog row,
  vocabulary and missing effort — all present and pass.
- **Provider limits**: frozen `pi-anthropic/openai/google` rows; `TestALimitedRecordOnPiAnthropicReadsAsLimitedFromItsOwnHome`,
  the unfrozen-id carve-out narrowing, other-home Healthy, `Observe` write gate, `pi-google` carve-out pinned.
- **Existing guards**: `sourceport_test` filter is narrowed to post-port systems with its own under-report
  test; producer's M7 states the over-filter bound honestly (full-set pin catches it). Regress smoke exemption is
  narrowed to exec-less systems (M8). Astra alias mutants mirror systems per `sameSystems`; original failure
  classes preserved.
- **Legacy**: `systems/pi`, `vendors/local-models`, `pkg/localruntime` untouched; local-models suite ok.

## Commands run by me (exit codes)

| Command | Exit |
| --- | ---: |
| `go vet ./...` | 0 |
| `go test ./pkg/agentic/systems/pinative/` | 0 |
| `go test ./pkg/vendorplugin/ -run 'Pi|Native|Alias|Astra|SourcePort|Frozen|Runtime|Vendor'` | 0 |
| `go test ./pkg/providerlimits/ -run 'Pi|Frozen|Broker'` | 0 |
| `env -u TASK_BOARD_DIR go test ./internal/regress/` | 0 |
| `go test ./pkg/vendorplugin/vendors/... ./tools/...` | 0 |
| temporary `TestReviewProbeUltraOnPiOpenAI` (removed) | 0 (admits ultra — the finding) |
| 8 isolated Pi probes (`pi-probe-thinking-02.log`) | all 1 at `No API key` / ambiguity, as expected |

Full `make vet/test/regress` exit 0 accepted from the producer's run bound to this tree; the narrow reruns
above are the discriminating subset. Logs under `.temp/review-2kapmh/` in the Story worktree.

## Checklist
Implementation matches AC: NO (R1). Fits architecture: yes, additive. Tests green: yes. Attacked, not read:
yes — the effort boundary was driven against the installed Pi parser and catalog and the gate was shown to
admit four pairs it must refuse. LOGBOOK: operator no-LOGBOOK instruction; this artifact is the record.
