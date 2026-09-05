# Drafting report: `LaunchModeInteractive` in skill-agents-management

Task `TASK-260905-2zxm3s`, per the producer brief and curator-spec Decision 0013 §5
(read at curator-spec `83de1a5`, story worktree copy).

## Where

- Repo worktree `/Users/iv/Developer/ReluxWorks/.worktrees/agents-management-interactive`,
  branch `feat/launch-mode-interactive`, base `91bf945` (= main). Go 1.25.5.
- Not pushed, not tagged, no PR opened. The orchestrator lands.

## Commits (signed, verified with `git log --show-signature`)

| SHA | Subject | Signature |
| --- | --- | --- |
| `cfd43f2d20a99b0b84787a89120d795b72fefa64` | Add LaunchModeInteractive for claude, codex and pi per Decision 0013 §5 | Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM |
| `3edbde888e55ef598ce983103146a7c0f235eb57` | Document LaunchModeInteractive and its invariant | Good "git" signature for oparin@me.com with ECDSA key SHA256:V6JiKG7J29mjsvikcLoSVp0bLa77VTsFy12gnLO81cM |

Author on both: Ivan Oparin <oparin@me.com>. 18 files changed, 1305 insertions, 29 deletions.

## What changed

### Core (`pkg/agentic`)

- `system.go`: `LaunchModeInteractive` appended after `LaunchModeManagedSession` (value 3; existing
  iota values untouched). `Valid()` and `String()` (`"interactive"`) extended. Doc comment states the
  decision's closed grammar constraints.
- `plan.go`: new sentinels `ErrCompositionNotInteractive` (the decision's) and
  `ErrParameterNotInteractive` (see Deviations). `BuildPlan` in interactive mode, after the mode
  check and before composition-grammar validation, refuses a non-zero `Composition` (prefix, servers,
  or both) and any goal, budget, service tier, `PromptPath` or `Prompt`. After `Stdin` it refuses an
  attached stdin (even an attached empty one) unless the system's `EffortTransport` is stdin, as
  `ErrPluginContract`. `Home`/`WorkDir` carry as before; `ErrEffortMissing` and
  `ErrUnsupportedLaunchMode` unchanged. Undeclared mode → `ErrUnsupportedLaunchMode` before dispatch.

### Systems

| System | Mode declared | Interactive argv (exact) | Effort | Stdin | Binary |
| --- | --- | --- | --- | --- | --- |
| `claude-code` | yes | `--model <id>` `[--effort <e>]` | argv | detached | `claude` on launch PATH (unchanged resolver) |
| `codex` | yes | `-m <id>` `[-c model_reasoning_effort="<e>"]` | argv | detached | codex three-path resolver (unchanged) |
| `pi` | yes | `pi --model <id>` | none (no flag) | detached | `agents-infra` wrapper on launch PATH (unchanged; never raw `pi`) |
| `agy`, `gemini-cli`, `muse`, `qwen-code` | no | — | — | — | refuse with `ErrUnsupportedLaunchMode` |

Each argv is built inside the plugin's existing single construction site:
claude `args.go` `Args` → `interactiveArgs` + shared `appendEffort` (both added to the argvguard
allowlist with reasons; the guard's narrowing and mutant tests still pass); codex `args.go` `Args`
interactive case reusing `appendReasoningAndTier` (tier refused first, so it contributes the effort
override alone); pi `args.go` `interactiveArgs`, with `prepareLaunchRequest` passing an interactive
request through (no profile assertion, no prompt read).

Plugin-level repeats of core refusals for a caller holding the plugin directly: claude refuses
goal/budget in `interactiveArgs`; codex refuses service tier and **profile** (profile is codex's own
fact, no core rule names it — see Deviations); pi refuses an empty model id.

### Verified binary spellings (installed binaries, this machine, 2026-09-05)

| Binary | Version | Spelling verified | Evidence |
| --- | --- | --- | --- |
| `claude` | 2.1.261 (Claude Code) | `--model <model>` "Model for the current session"; `--effort <level>` "Effort level for the current session"; `-p, --print` is the non-interactive switch and is absent | `claude --help` |
| `codex` | codex-cli 0.153.2 | top-level `-m, --model <MODEL>` and `-c, --config <key=value>`; `exec` is the non-interactive subcommand and is absent; `-p, --profile` absent | `codex --help` |
| `pi` | 0.84.2 | `--model <pattern>`; `--print, -p` is non-interactive and absent | `pi --help` |
| `agents-infra` | v1.6.1-128-gab60e0d | `agents-infra pi <args>` with first arg not `spawn`/`turn`/`lifecycle` starts the interactive primary session and passes the remaining args through (`tools/agents-infra/main.go` `runPi`); `agents-infra pi --help` prints raw pi's help | source + `agents-infra pi --help` |

Not verified at runtime: whether the local-models model id (e.g. `qwen-3.8-27b-mlx-8bit`) resolves
under raw pi's `--model <pattern>` matcher through the wrapper. That is the vendor/wrapper's fact;
the plugin passes `Model.ID` through verbatim as the decision requires. No process was started.

### Goldens: why interactive is not a `parity/testdata/goldens` entry

`pkg/agentic/parity/doc.go`: every golden there was captured by the SOURCE repository's harness and
"this package captures nothing"; a golden captured next to the port "proves only that the new code
agrees with itself". Interactive mode has no source capture (the source's interactive path filtered a
human's argv rather than constructing one). The parity package declares no way to add non-source
goldens. So the interactive goldens are plugin-local tests, in the style of
`systems/claude/args_test.go` and codex's frozen managed-session proof:

- `pkg/agentic/systems/claude/interactive_test.go`, `codex/interactive_test.go`,
  `pi/interactive_test.go`: exact argv, exact env (run context written, `CLAUDECODE` stripped, no
  codex tier variable, pi passes `AGENTS_INFRA_CALLER_CWD` through), detached stdin, `Home`/`WorkDir`,
  binary — all through the real registry and `agentic.BuildPlan` (production call site). Negative:
  exec-mode marker sweep (`-p`, `--print`, `--output-format`, `--dangerously-skip-permissions`,
  `--max-budget-usd`, `--append-system-prompt-file`, `/goal ` prefix, `--mcp-config`, `exec`,
  `--dangerously-bypass-approvals-and-sandbox`, `--skip-git-repo-check`, sandbox/approval policy,
  `-p/--profile`, `-C`, `--add-dir`, `-`, `spawn`, `--prompt`, `--deadline`, `--result-schema`,
  `service_tier=`) must be empty on the interactive argv AND must fire on the same system's exec
  (and codex managed-session) argv — the narrowing that makes silence a measurement.
- Refusals through `BuildPlan`: composition → `ErrCompositionNotInteractive`; goal, budget, tier,
  prompt path, prompt bytes → `ErrParameterNotInteractive`; missing required effort →
  `ErrEffortMissing`; pi effort word → `ErrEffortNotTransportable`; codex profile → plugin refusal
  naming the profile.
- `pkg/agentic/interactive_test.go` (core, test double): mode appended/valid/named; positive baseline;
  undeclared-mode refusal with `Argv` never dispatched; composition refusal before
  `ValidateComposition` is called, with the same composition admitted in exec mode; each forbidden
  parameter refused alone with the error naming it; attached stdin refused under argv/none
  transports (including attached-empty), admitted under stdin transport, admitted in exec mode.

### `internal/regress`

`internal/regress/interactive_test.go` (new class "INTERACTIVE"): for `claude-code` and `codex`,
build the interactive plan through `paritycase.BuildPlan`, sweep one union marker list, require
detached stdin, and require the sweep to fire on the same system's exec plan; the composition
sentinel once per system; and a registry walk asserting every non-declaring registered system
refuses the mode and every declaring one has a case here.

**pi is not in the regress case**, by that package's own rule: importing the pi plugin registers it
into the default registry, and the existing `TestEveryLayerOneSystemHasASmokeCase` then demands a
golden pi cannot have (the package's own comment says it never imports pi for this reason; I hit
the failure and backed out). pi's cross-cutting negative lives in its own package with the same
marker discipline. `make regress` runs both.

### Gate-narrowing evidence (mutants planted, then restored; not committed)

| Mutant | Result |
| --- | --- |
| composition refusal disabled in `BuildPlan` | FAIL: `pkg/agentic` TestBuildPlanRefusesAnInteractiveLaunchCarryingAComposition; claude+codex TestAnInteractiveLaunchRefusesWhatItsGrammarCannotCarry; regress TestAnInteractivePlanRefusesACompositionForEveryMappedSystem |
| interactive stdin gate disabled | FAIL: TestBuildPlanRefusesAnAttachedInteractiveStdinUnderANonStdinTransport |
| `-p --dangerously-skip-permissions` planted in claude interactive; `--dangerously-bypass-approvals-and-sandbox` planted in codex interactive | FAIL: both plugins' TestTheInteractiveArgvIsModelAndEffortOnly and TestTheInteractiveArgvCarriesNoExecMarker; regress TestAnInteractivePlanCarriesNoExecMarkerForAnyMappedSystem |

## Gates (standalone commands, real exit codes; logs in curator-spec `.temp/TASK-260905-2zxm3s/`)

| Command | Exit | Log |
| --- | ---: | --- |
| `gofmt -l .` | clean (no output) | — |
| `make build` | 0 | `make-build-01.log` (silent) |
| `make vet` | 0 | `make-vet-01.log` (silent) |
| `make test` | 0 | `make-test-01.log` |
| `make regress` | 0 | `make-regress-01.log` |

`make test` tail:

```
ok  	github.com/relux-works/skill-agents-management/internal/regress	2.215s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic	2.922s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic/parity	4.535s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic/systems/agy	4.989s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic/systems/claude	3.785s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic/systems/codex	3.564s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic/systems/gemini	5.659s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic/systems/muse	4.175s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic/systems/pi	5.142s
ok  	github.com/relux-works/skill-agents-management/pkg/agentic/systems/qwen	3.307s
ok  	github.com/relux-works/skill-agents-management/pkg/inferenceengine	5.836s
ok  	github.com/relux-works/skill-agents-management/pkg/inferenceengine/engines/mlx	5.917s
ok  	github.com/relux-works/skill-agents-management/pkg/localruntime	5.930s
ok  	github.com/relux-works/skill-agents-management/pkg/plugin	5.707s
ok  	github.com/relux-works/skill-agents-management/pkg/providerlimits	14.041s
ok  	github.com/relux-works/skill-agents-management/pkg/vendorplugin	5.586s
ok  	github.com/relux-works/skill-agents-management/pkg/vendorplugin/vendors/local-models	7.629s
ok  	github.com/relux-works/skill-agents-management/tools/agents-management/cmd	9.750s
```

`make regress` tail:

```
ok  	github.com/relux-works/skill-agents-management/internal/regress	1.055s
```

## Docs touched

- `docs/architecture.md`: launch-mode list under the agentic-system plugin declaration; new
  invariant 6 (interactive grammar bound, refusals, stdin rule, M2 rationale).
- `README.md`: `pkg/agentic` bullet listing the four modes and the interactive refusals; codex
  "one argv construction site" bullet now names the interactive grammar.
- `docs/consuming-the-module.md`: "Launch and availability entry points" gains an
  "Interactive sessions" paragraph (who calls it, what the plan must not contain).
- `docs/shipped-state.md`: untouched — it enumerates no launch modes.

## Deviations from the brief / decision, stated

1. **`ErrParameterNotInteractive` added** beyond the decision's one named sentinel. The decision
   forbids goal/budget/tier/prompt on the interactive argv and stdin; admitting them and building an
   argv without them would be a silent drop, which this module refuses everywhere else. Core-level
   enforcement keeps it a single source instead of three plugin copies. If the reviewer prefers
   plugin-only refusals, the core check is one function (`refuseNonInteractiveParameters`).
2. **codex refuses `Profile` in interactive mode** (plugin-level, plain error). The decision's
   argv list is model + effort only; a profile that reached no `-p` flag would be a configured
   profile dropped. The launcher today sets no profile for codex. pi is different: local-models'
   `Spawn` always contributes a profile, and the interactive wrapper resolves its own from project
   config, so pi accepts the profile on the request and spells nothing — documented in
   `pi/args.go`.
3. **pi absent from `internal/regress`** for the reason above; covered in its own package.
4. **agy, gemini, muse, qwen skipped**: no one-line trivially correct argv for them was verified,
   so they are undeclared and refuse with `ErrUnsupportedLaunchMode` (asserted in regress).

## Not done here, on purpose

- No push, no tag, no PR (orchestrator lands). Release: a compatible minor after the latest tag
  `v0.5.7` is the operator's call; the change is additive (new enum value, new sentinels, new mode
  declarations).
- No LOGBOOK.md written anywhere (brief forbids writing into the control root); findings are in
  this report and the board notes.
- No process was launched to validate the interactive argv end to end.
