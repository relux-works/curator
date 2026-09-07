# TASK-260908-1c0fwn — Plan.Env value contract (A0 E4) resolution

Read-only source analysis; no ax call, no ax PR or repo mutation, no code. Probe artifacts under the story worktree `.temp/TASK-260908-1c0fwn/` (probe module + `probe-01.log`).

## 1. Exact sources

- Launcher `SPEC.md` `0.2.1-draft` @ `484933b` (story worktree): §4.4 (L335-372, says nothing about what `LaunchRequest.Env` the launcher supplies), §4.5 environment layers L440-446, SHOULD-warn L452-456, `env_names` rule + untracked note L458-475, §4.6 `env_literals` row L546.
- `curator-spec/decisions/0013-…launch-plans.md` @ `87a0d00` (file head `83de1a5`), status *Proposed, draft*: §3.2 document grammar L140-180 (closed set: `argv|argv_suffix`, `env_names`, `env_literals`, `stdin`, `extensions`; unknown member rejected); §6.3 L487-500; §6.4 L519-524.
- `skill-agents-management` tag `v0.5.10` = `13d167e2c5cabb226eb0563f591ae6c63095855a`, go.sum `h1:mrpOOLcW+YaNvwdsQldULZ7H5WLTSoZoUVgFvLdinAQ=`:
  `pkg/agentic/plan.go:38` `Plan.Env []string`; `plan.go:280` `env, err := sys.ChildEnv(req.Env, req)`; `system.go:419` `LaunchRequest.Env` = "the parent environment the child inherits from, before the system's own filtering"; `system.go:534-541` `ChildEnv(parent, req)` contract; `runcontext.go:61-100` `SetEnvValue` (whole key, empty value removes) and `WithRunContext` (six `TASK_BOARD_*` keys always rewritten); `systems/claude/env.go` (strips exactly `CLAUDECODE`); `systems/codex/env.go` + `internal/runtimeenv` (11 keys, the tokens the two `*_AUTH_TOKEN_ENV` pointers name, `PATH` sanitizer dropping `codex-path` dirs); `systems/claude/claude.go:151` and `codex/codex.go:129` `resolveBinary(req.Env)`; `internal/launchenv:39` `ErrNoPath` — an Env without `PATH` is a refusal; `internal/paritycase/paritycase.go:262-275` `WholeEnvWipeSystem` = `ChildEnv(nil, req)` (the module already tests the nil-parent shape).
- A0: `TASK-260908-qblycn_a0-verification-findings.md` §3.5, E4.

## 2. Verified Plan.Env semantics (probe, 5/5 PASS, exit 0)

| Case (seeded parent) | claude | codex |
|---|---|---|
| inherited unchanged (`HOME`, `FAKE_SECRET`, `CLAUDE_CONFIG_DIR`) | kept, values intact | kept |
| removed | `CLAUDECODE`, `TASK_BOARD_RUN_ID`, `TASK_BOARD_DIR` | `CODEX_THREAD_ID`, pointer `TASK_BOARD_CODEX_APP_SERVER_AUTH_TOKEN_ENV`, the token it named (`MY_TOK`), `TASK_BOARD_RUN_ID`, `TASK_BOARD_DIR` |
| changed | none | `PATH` (`/tmp/parent/codex-path` dropped) |
| added (zero `Run`) | none | none |
| `ChildEnv(nil, req)` zero `Run` | `[]` | `[]` |
| `ChildEnv(nil, req)` with `Run` | exactly `TASK_BOARD_{RUN_ID,TASK_ID,BOARD_DIR,DIR,CONTEXT_ID}` | same |

Facts beyond E4's wording: (a) `Plan.Env` is a complete environment, not a delta — E4 confirmed; (b) with the zero `RunContext` a terminal launch carries, **every** plugin also removes the six run-context keys from the parent, not only its own strip list; (c) codex *changes* `PATH`; (d) no plugin writes the home variable (§3.5 confirmed); (e) naive "all `Plan.Env` → `env_literals`" serializes the inherited `FAKE_SECRET` (negative probe `TestNaiveLiteralsLeakInheritedValues`); (f) `BuildPlan` with `Env=nil` is refused by `ErrNoPath`, so the own-name set cannot come from a second `BuildPlan`.

## 3. Composition rule (both modes, one sentence)

**Layer 1 is the input of the plan request, never a composition layer. The composed environment starts at `Plan.Env`; fragment `env` and the variable-kind channel apply on top. The plan's *own names* are `System.ChildEnv(nil, req)` — the plugin's answer over an empty parent — and only those, never `Plan.Env` itself, enter `env_literals`.**

- **Untracked:** `LaunchRequest.Env = os.Environ()`; exec `Plan.Env ⊕ fragment env ⊕ channel`. Removals (strip lists, run-context keys) stay removed; codex's sanitized `PATH` is used; inherited values pass through untouched. Fixes E4 (`CLAUDECODE` re-admission).
- **Tracked:** `BuildPlan` still takes `os.Environ()` (binary resolution and admission need it; `Binary`/`Argv`/`Stdin` do not depend on `Env` beyond `PATH`). `env_literals = keys/values of ChildEnv(nil, req) ⊕ fragment env ⊕ channel`. At `v0.5.10` with a zero `Run` that is `fragment env ⊕ channel` only — no inherited name, no `PATH`, no secret can enter the document by construction. No diffing against the launcher environment is needed or allowed.
- **SHOULD-warn (§4.5):** "a name layer 2 actually set" is ambiguous once layer 2 is a full environment (it would fire on any inherited `CLAUDE_CONFIG_DIR`). Redefine: warn when fragment/channel overrides one of the plan's own names. At `v0.5.10` this never fires.
- **`env_names` collision:** unchanged in intent, but it must subtract the *own-name* literals, not `Plan.Env`. Under the current wording every inherited name (e.g. an operator's `FIGMA_API_KEY`) would collide and be dropped from `env_names`, silently disabling the destination lookup. The E4 fix is load-bearing for `env_names`.
- **Destination-local inherited env:** untouched by the launcher; it is ax's layer, as Decision 0013 §6.4 already says.

## 4. Unexpressible constraint (recorded residual, not a product decision)

The launch-plan document (0013 §3.2) has no member meaning "unset name X on the destination" or "rewrite `PATH`". The plugin's strip list — `CLAUDECODE`, codex's runtime family and token pointers, `PATH` sanitizing, run-context removal — therefore cannot reach a tracked child. Observable consequence: if ax's terminal backend hands the child `CLAUDECODE=1` (ax started from inside a Claude session), the claude child refuses as nested. Whether ax's own provider plugins strip these names was **not read** and is unknown. This needs no decision now: record it in SPEC §9 and 0013 residuals; a future ax revision may add an `env_unset` member (ax change, out of scope here). No choice between correct alternatives remains — the composition above is the only one that satisfies own-names-only, no inherited leak, and plugin ownership without a new API.

## 5. Erratum (smallest, evidence-backed)

1. SPEC §4.4: add "The launcher passes its process environment as `LaunchRequest.Env` in both modes; the plan's `Env` is the plugin's complete child environment over it."
2. SPEC §4.5 environment: replace layers 1-2 by one layer "the plan's `Env`" (three layers total); define *own names* = `ChildEnv` over an empty parent; reword the SHOULD-warn to own names.
3. SPEC §4.5 `env_names`: "minus every name in the composed `env_literals` of §4.6" (unchanged text, now correct because of 4).
4. SPEC §4.6 `env_literals` row: "the plan's **own names** (its `ChildEnv` over an empty parent) ⊕ fragment `env` ⊕ channel — never `Plan.Env`, which is a full environment".
5. SPEC §9: residual "plugin strip list does not reach a tracked destination".
6. Decision 0013 §6.3 L487-491 and §6.4 L519-522: same one-line substitution ("plan `Env`" → "the plan's own names"; drop "inherited ⊕"). 0013 is *Proposed*, so this is an amendment, not an erratum on a closed document.

## 6. Deterministic tests for the implementation story (fake ax only)

- T1 untracked: parent `CLAUDECODE=1` → composed env lacks it (fails under the overlay reading).
- T2 tracked, negative bound: parent `FAKE_SECRET`, `HOME` → `env_literals` lacks both; **narrowing**: substituting `Plan.Env` for `ChildEnv(nil)` at the composer call site must fail this test.
- T3 tracked: claude with fragment env → `env_literals == {CLAUDE_CONFIG_DIR: <home>}` exactly; codex → `PATH` never in literals.
- T4 collision: fragment `env_names=[FIGMA_API_KEY]` + inherited `FIGMA_API_KEY` → name stays; `env_names=[CLAUDE_CONFIG_DIR]` → dropped with warning.
- T5 own-name consistency: `keys(ChildEnv(nil,req)) ⊆ keys(Plan.Env)` with equal values (guards plugin drift between the two calls).
- T6 `BuildPlan(Env=nil)` returns `ErrNoPath` (pins why the composer must not use a second `BuildPlan`).

## 7. Board notes

Solution-architect decomposition checklist items about board sizing/dependencies apply to the story's follow-up (one spec-erratum task, one composer implementation task with T1-T6); no new elements created here — the story owner decides after review. No logbook write (operator rule); this resource is the equivalent record.
