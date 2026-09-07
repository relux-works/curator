# A0 installed-contract verification — findings

Task: TASK-260908-qblycn (STORY-260908-3mcpz5, EPIC-260908-2wp8wn). Role: researcher.
Date: 2026-09-08. Machine: the operator's macOS host (Darwin 25.6.0, arm64).
Subject: launcher `SPEC.md` 0.2.1-draft (worktree HEAD `484933b`), curator-spec main `87a0d00`,
curator remote main `04550e28`, agents-management tag `v0.5.10`, relux-agents-infra `dee5403`.

Confidence labels used below: **verified** = observed on this machine by running the installed
binary or reading the installed/tagged source; **docs-confidence** = taken from a document or a
prior sprint's evidence and not reproduced here; **unknown** = could not be established.

Evidence logs (task-local, `.temp/TASK-260908-qblycn/`, paths sanitized with `~`):
`01-readiness.log`, `02-profiles.log`, `03-resolve-readonly.log`, `04-{claude,codex,pi}-help.log`,
`05-module-resolve.log`, `06-resolve-repair.log`, `06b-resolve-repair-second.log`,
`07-codex-layer-probe.log`, `07b-codex-layer-probe.log`, `08-claude-argv-probe.log`,
`09-curator-main-build.log`, `10-installed-curator-update.log`, `11-resolve-installed.log`.
The full evidence bundle is attached as a second outcome resource (tar.gz).

## 0. Summary

Every A0 item was exercised. Three real fragments were produced for `claude_code`, `codex_cli`
and `pi` from the one installed profile (`default`). The agents-management API at `v0.5.10`
is verified from the tagged source and resolves through the public Go proxy. Native argv
boundaries were probed on the pinned-or-newer releases installed here with isolated homes.

The verification **falsifies or materially qualifies** the SPEC in six places (§6, errata
E1–E6). The two that block implementation as drafted:

- **E1** — the `pi` system plugin at `v0.5.10` plans `Binary = agents-infra`, argv
  `["pi", "--model", <id>]`, and the agents-infra wrapper strips an inherited
  `PI_CODING_AGENT_DIR` and sets its own cache-root home. A `curator run pi` composed per
  SPEC §4.5/§4.6 can never launch into the Curator-managed pi home.
- **E2** — no runtime with `System: "pi"` exists in the module's frozen runtime table; the pi
  system is reachable only through the `local-models` vendor and a machine-local
  `local-models.toml` that is absent here. SPEC §4.3 level 3 (lineup fallback) and §4.4 have no
  vendor to consult for `pi`, and `BuildPlan` needs a `Model.Effort` support value that only a
  vendor row carries.

Missing external inputs (not substituted): no installed profile carries a context package with
system modules or an MCP package, so no real fragment with `system_prompt` or `mcp` sections
exists on this machine (§2.4). `ax` is not installed (§1); untouched as required.

## 1. Tool readiness and installed versions (verified)

| Tool | Path | Version | Note |
|---|---|---|---|
| curator (installed at task start) | `~/.local/bin/curator` | `v0.14.0-rc.3-113-g66e34a2` (built 2026-09-02) | **no `env` or `profile` subcommands** — exit 2 "unknown command" (`01-readiness.log`) |
| curator (task-local candidate, control tree) | `.temp/…/bin/curator` | `v0.14.1-0.20260906225215-a66eec88fa4d+dirty` | built from the control checkout at `a66eec88` (dirty: board activity files only); used for §2 first pass |
| curator (isolated clone of remote main) | `.temp/…/curator-main/bin/curator` | `v0.14.1-0.20260907213730-04550e282705` | `git clone --branch main https://github.com/relux-works/curator.git`, HEAD `04550e28`, `vcs.modified=false`; `go build -o bin/curator ./cmd/curator` exit 0 |
| curator (installed after update) | `~/.local/bin/curator` | see §1.1 | |
| claude | `~/.local/bin/claude` | `2.1.263` | environments.md §7.3/§7.8 record 2.1.261 |
| codex | `~/.local/bin/codex` | `codex-cli 0.153.4` | environments.md records 0.153.2 |
| pi | `/opt/homebrew/bin/pi` → `@earendil-works/pi-coding-agent/dist/cli.js` | `0.84.2` | matches the recorded 0.84.2 |
| agents-infra | `~/.local/bin/agents-infra` | `v1.6.1-128-gab60e0d` | the pi plugin's `Binary` (§3.5) |
| ax | — | **not installed** | untracked machine; no `ax.json` family exists |
| go | `/opt/homebrew/bin/go` | `go1.25.5 darwin/arm64` | |
| curator-run | — | not installed (expected: A1 not started) | |

The goal file names curator main `919e2e9c`; remote main was `04550e28` at verification time
(ten commits later), local control-tree main `a66eec88`. All three carry `env resolve`.

### 1.1 Installed-binary update (parent directive, run RUN-260907-adbf23)

The parent's nudge authorized updating the installed curator from a verified current-main
build as a prerequisite. Done as follows (`09-…`, `10-…` logs):

1. Isolated checkout: `git clone --branch main https://github.com/relux-works/curator.git`
   into `.temp/TASK-260908-qblycn/curator-main`; HEAD `04550e28`, `git status --short` empty,
   `git describe` `v0.14.0-50-g04550e2`.
2. Build: `go build -o bin/curator ./cmd/curator` exit 0; `curator --version` →
   `v0.14.1-0.20260907213730-04550e282705`; `go version -m` → `vcs.revision=04550e28…`,
   `vcs.modified=false`.
3. Verification (real exit codes): the combined
   `go test ./cmd/curator ./internal/envfragment ./internal/envprofile ./internal/envregistry`
   under a 500 s bound **timed out, exit 124** (the `cmd/curator` suite is large). Split runs:
   `go test -count=1 ./internal/envfragment ./internal/envregistry` → ok, exit 0;
   `go test -count=1 ./internal/envprofile` → ok (79 s), exit 0;
   `go test -count=1 -run 'Env|Profile|Umbrella|Resolve' ./cmd/curator` → ok (108 s), exit 0.
   The rest of the `cmd/curator` suite was **not** run here; CI on `04550e28` is the
   authority for it.
4. Backup: `~/.local/bin/curator` (sha256 `725068f1…5920`) copied to
   `.temp/TASK-260908-qblycn/backup/curator.v0.14.0-rc.3-113-g66e34a2`.
5. Install: the Makefile has no `install` target and the README's flows (brew, installer,
   `go install …@latest`) resolve to the rc tag without `env`; so
   `GOBIN=~/.local/bin go install ./cmd/curator` from the clean clone, exit 0.
   `~/.local/bin/curator --version` → `v0.14.1-0.20260907213730-04550e282705`,
   `vcs.modified=false`.

No runtime directory was hand-edited; `~/.claude`, `~/.codex`, `~/.agents` untouched.

## 2. Fragments: `curator env resolve <env> --profile default --repair --format json`

### 2.1 Profile inventory before any mutation (verified, `02-profiles.log`)

`curator profile list` → exactly one profile: `default`, source `local`, version `0.0.0`,
lock `sha256:726310f80f44…f30e19`. No profile is current (`current_profile: null` in
`env config show`). `env status`: all four adapters `non-current, unprovisioned, mode
managed-home`. Tool rows: claude recorded 2.1.261 / detected 2.1.263; codex recorded 0.153.2 /
detected 0.153.4; pi 0.84.2 both. No takeover was performed; `~/.claude`, `~/.codex`,
`~/.agents` symlinks verified unchanged after the runs (`06b-…log` tail).

### 2.2 Read-only resolve (fail-closed, verified, `03-resolve-readonly.log`)

| Invocation | stderr | exit |
|---|---|---|
| `env resolve <env> --format json` (no `--profile`) | `curator: profile_unknown: no profile is current` | 1 |
| `env resolve <env> --profile default --format json` | `curator: environment_home_stale: home unprovisioned` | 1 |

Both match environments.md §10.4 and SPEC §4.1's read-only/fail-closed statement. Note the
diagnostic is printed as `curator: <code>: <detail>` on stderr with no JSON; the launcher's
§4.1 code mapping must parse the code token from that line.

### 2.3 Repair resolve, first run (provisioning) and second run (current home) — verified

`06-resolve-repair.log` (first run, home provisioned) and `06b-…log` (second run, no mutation).
stdout is exactly one line of CCJ-1 JSON plus one LF (`internal/envfragment.JSON` marshals
through `protocoljson.MarshalCanonical`); notices and warnings go to **stderr**. Sanitized
stdout (identical on both runs):

```json
{"env":{"CLAUDE_CONFIG_DIR":"~/.curator/environments/default/claude_code"},"environment":"claude_code","fragment":"launch-env-fragment-v1","precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19","name":"default"}}
{"env":{"CODEX_HOME":"~/.curator/environments/default/codex_cli"},"environment":"codex_cli","fragment":"launch-env-fragment-v1","precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19","name":"default"}}
{"env":{"PI_CODING_AGENT_DIR":"~/.curator/environments/default/pi"},"environment":"pi","fragment":"launch-env-fragment-v1","precedence":{"placement":"winner-last","winner":"higher-weight"},"profile":{"lock_sha256":"726310f80f442428a9a640d2d49ad9b635f22857fb4ad22167832c7c6ff30e19","name":"default"}}
```

(`~` stands for the absolute home path; real output is absolute, as §10.2 requires.)

stderr, first run: `warning: environment_tool_version_unverified: claude_code detected 2.1.263,
recorded 2.1.261` (claude), the same for codex 0.153.4/0.153.2, none for pi; then a three-line
provisioning notice per env ("Managed home provisioned at …", state-root sentence, first-run
steps). stderr, second run: only the version warning for claude/codex, nothing for pi. Exit 0
on all six invocations.

Managed home contents after repair: claude `.agent-environment.json`, `.claude.json` (written
seed); codex `.agent-environment.json`, `auth.json` (passthrough), `config.toml` (seed copy);
pi `.agent-environment.json`, `auth.json` (passthrough). No `curator-mcp.config.toml`, no
`.agent-context/`, no `SYSTEM.md`/`APPEND_SYSTEM.md` — the `default` profile is empty.

Fragment digests (sha256 over the stdout bytes without the trailing LF, i.e. over the CCJ-1
bytes): claude `c7d17c39…545e`, codex `04459237…2e32`, pi `c0512f55…01bf` (full values in
the shell transcript, `06b` section of the bundle). These are what §4.1 must reproduce from the
parsed object.

### 2.4 What the fragments do not show — missing external input

- No `system_prompt` section and no `mcp` section in any fragment: the only installed profile
  has no context package with `class: system` modules and no MCP declaration package. **Real
  fragments with those members cannot be produced on this machine until a profile carrying
  them is installed** (Workstream B1/B3/B5). Not substituted with fixtures. The shapes are
  taken from environments.md §10.2 and curator `internal/envregistry` (docs-confidence for the
  bytes, verified for the descriptor table — §3.7).
- The `pi` fragment correctly has **no `mcp` section by protocol** (environments.md §7.8: pi
  0.84.2 has no MCP channel; `envregistry` `Pi.MCP: nil`). This is the expected absence, not a
  gap.
- `path_prepend` absent (revision 1 never emits it) — verified absent.

## 3. agents-management API at the imported tag (verified from source at `v0.5.10`)

### 3.1 Tag identity and fetchability

- `v0.5.10` → `12f443d1`, tag signature `Good "git" signature` (ECDSA, oparin@me.com),
  tagged 2026-09-07. `LaunchModeInteractive` is present from `v0.5.8` onward (`v0.5.7` lacks it).
- `go list -m -json github.com/relux-works/skill-agents-management@v0.5.10` through the default
  proxy: resolves, `Origin.Hash = 12f443d1…` (`05-module-resolve.log`). SPEC §7 "public module,
  no replace" holds.
- `docs/consuming-the-module.md` at the tag still says "require `v0.4.3`" in its header; the
  interactive-session paragraph names the curator launcher `0.2.0-draft` §4.1 as the consumer.

### 3.2 Entry points and request shapes

`agentic.BuildPlan(r *agentic.Registry, req agentic.LaunchRequest, mode agentic.LaunchMode) (agentic.Plan, error)`
— `pkg/agentic/plan.go`. `LaunchRequest` members relevant to the launcher: `System SystemID`,
`Model agentic.Model{ID string, Effort agentic.EffortSupport, AliasOf string}`, `Effort string`,
`WorkDir`, `Home`, `Env []string`, `Composition`, plus `Profile`, `Run`, `Goal`, `Budget`,
`ServiceTier`, `PromptPath/Prompt`, `Runtime`. `Plan` = `System, Mode, Binary, Argv, Env,
Stdin StdinPayload, WorkDir, Home, ModelIdentity{Requested, Launched}, Provenance, Nodes`.

Two facts the SPEC does not state and the implementation must handle:

- **`Model.Effort` is an `EffortSupport` enum, not a word.** `BuildPlan` refuses
  `ErrEffortMissing` only when `req.Model.Effort == EffortSupportRequired && Effort == ""`. The
  support value comes from the vendor catalog row (`vendorplugin.Model.Effort.Support`); a
  launcher calling `BuildPlan` directly with a bare model string would have to look the row up
  itself. The module's own path for that is `vendorplugin.BuildLaunch(ctx, r *vendorplugin.Registry,
  req SpawnRequest{Runtime, Model, Effort, WorkDir, Home, Env, Profile, …}, mode)`, which selects the
  row, validates the effort word against the row's vocabulary (`resolveEffort`), then calls
  `agentic.BuildPlan`. SPEC §4.4 names `BuildPlan` as "the declared entry point"; `BuildLaunch` is
  the entry that realizes §4.3's "admission is the spawn plane's" for the effort word. (Erratum E3.)
- **`BuildPlan`/`BuildLaunch` never consult provider limits.** `BuildLaunch`'s doc: "What it does
  NOT do is ask whether the launch is allowed RIGHT NOW". The verdict the SPEC's
  `plan_provider_limited` needs comes from a separate read:
  `providerlimits.Store.AvailabilityFor(VerdictQuery{Runtime, Model, Home, Table})` →
  `vendorplugin.Availability{State, Until, Checked, Observed, Failures}`; `Serviceable()` is true
  only for `AvailabilityHealthy`. Store layout: `$XDG_STATE_HOME` or `os.UserConfigDir()` +
  `task-board/provider-limits/<identity>.state.json`; identity = `IdentityKey(provider,
  NormalizeProviderHome(home))`. `VerdictQuery.Runtime` is the runtime id (`"claude"`,
  `"codex"`), and `Home` empty resolves the native default home — so the launcher must pass the
  managed home explicitly, exactly as SPEC §4.4 intends. (Erratum E3, same item.)

### 3.3 Mode and the interactive refusals

`LaunchMode` = `Exec, DryRun, ManagedSession, Interactive` (iota 0–3; `Valid()`/`String()`
cover all four). In interactive mode `BuildPlan` calls `refuseNonInteractiveParameters` before
the composition-grammar checks: a non-zero `Composition` → `ErrCompositionNotInteractive`
("interactive launch carries no composition; the MCP prefix is the composer's"); a `Goal`,
`Budget`, `ServiceTier` or prompt → `ErrParameterNotInteractive`. Also present:
`ErrUnknownSystem`, `ErrUnsupportedLaunchMode`, `ErrModelMissing` (empty `Model.ID`, added at
`93abeae`), `ErrEffortNotTransportable`, `ErrEffortMissing`, `ErrPluginContract`.

### 3.4 Interactive argv per system (verified from the tagged goldens)

| System | Interactive `Argv` golden | Effort transport | Binary |
|---|---|---|---|
| `claude-code` | `["--model", M, "--effort", E]`; `["--model", M]` when effort unset | argv | `claude` on the request `Env`'s PATH |
| `codex` | `["-m", M, "-c", "model_reasoning_effort=\"E\""]`; `["-m", M]` without effort | argv (`-c`, TOML-quoted) | `codex` (or `codex.js` shim) on PATH |
| `pi` | `["pi", "--model", M]` | **none** (`EffortTransportNone`; an effort word is refused `ErrEffortNotTransportable`) | **`agents-infra`** on PATH (`pkg/agentic/systems/pi/binary.go`: `executableName = "agents-infra"`, "never raw pi") |

`Stdin` is unattached for all three (`TestTheInteractiveArgvIsModelAndEffortOnly` in each
plugin). Negative goldens pin the absence of `-p`, `--output-format`,
`--dangerously-skip-permissions`, `exec`, `--dangerously-bypass-approvals-and-sandbox`,
`--max-budget-usd`. SPEC §4.4 "stdin unattached for every launchable environment" holds.

### 3.5 Plan `Env` and the home variable

`Plan.Env` is the plugin's `ChildEnv(req.Env, req)`: claude clears `CLAUDECODE` (nesting
marker) and nothing else; codex filters the `CODEX_*`/`TASK_BOARD_*` runtime family and
sanitizes PATH; pi passes through with `AGENTS_INFRA_CALLER_CWD`. **No plugin writes
`CLAUDE_CONFIG_DIR`/`CODEX_HOME`/`PI_CODING_AGENT_DIR` into `Plan.Env`** — `Plan.Home` is a
separate field (`Capabilities.HomeEnvVar` is declared for claude/codex and empty for pi, and is
only read by `providerlimits`). Consequences for SPEC §4.5:

- the fragment's `env` map is the only source of the home variable in the composed
  environment — consistent with the SPEC;
- the §4.5 "SHOULD warn when layer 3 overrides a name layer 2 set" can only fire if the
  inherited environment already carried the variable (layer 2 is a filtered copy of layer 1);
- **layering "inherited ⊕ plan Env per name" re-admits keys the plugin deliberately removed**
  (`CLAUDECODE`, codex's `CODEX_*` family): the plan's `Env` is a complete replacement of the
  inherited environment, not a delta. (Erratum E4.)

### 3.6 Runtime table, lineup and the pi gap

`vendorplugin.FrozenRuntimes()` at the tag declares exactly: `claude`→`claude-code`/`anthropic`,
`codex`→`codex`/`openai`, `qwen`→`qwen-code`/`alibaba`, `gemini`→`gemini-cli`/`google`,
`agy`→`antigravity`/`google`, `muse`→`muse`/unresolved. **No runtime declares `System: "pi"`.**
The only documented pi runtime is `local-qwen` (`local-models` vendor, registered conditionally
from `~/.agents/.configs/local-models.toml`, which does **not exist** on this machine). The
runtime "compatibility registry" SPEC §4.3 names is `RuntimeDeclaration.System` plus
`Model.Systems` on each vendor row (`model.DrivenBy(system)` in `BuildLaunch`).

`vendorplugin.Lineup(models []Model) []RankedModel{Model, Position, Tied}` and
`LineupOf(vendor)` exist as the SPEC says; `Model.Effort` is
`EffortDeclaration{Support, Vocabulary, Recommended}` — the SPEC's "`Effort.Recommended`" is the
right member. Anthropic rows at the tag: `claude-fable-5-1` (required effort, recommended
`high`), `claude-fable-5`, `claude-opus-5` (`Recommended: true`), `claude-opus-4-8` (recommended
`xhigh`), `claude-sonnet-5`, …, `claude-haiku-4-5` (`effortNone`). Score values are in
`models.go`; the lineup's top row for `claude-code` is a runtime question for A1's golden.

### 3.7 Curator-side channel descriptors (verified in `internal/envregistry`, remote main)

claude: sysprompt `flag/append --append-system-prompt-file (path)`, `flag/replace
--system-prompt-file (path)`; MCP `flag --mcp-config (path) with ["--strict-mcp-config"]`.
codex: sysprompt `config-key/replace model_instructions_file`; MCP `flag -p (name) name
"curator-mcp"`. pi: sysprompt `flag/append --append-system-prompt (path)`, `file/append
APPEND_SYSTEM.md`, `file/replace SYSTEM.md`; MCP `nil`. opencode: MCP `variable
OPENCODE_CONFIG`. All match environments.md §7.3/§7.8 and SPEC §4.5/§5.

## 4. Native interactive argv boundaries (verified on the installed releases, isolated homes)

### 4.1 claude 2.1.263 (`08-claude-argv-probe.log`, `CLAUDE_CONFIG_DIR` = empty temp dir)

- Usage: `claude [options] [command] [prompt]`; interactive by default; `[prompt]` is the one
  positional → the first non-option operand is the user turn. Commander parses options
  anywhere on the line, so a channel flag after a native operand is still a flag; the SPEC's
  ordering is safe here but not load-bearing.
- `--model`, `--effort <low|medium|high|xhigh|max>`, `--append-system-prompt-file <path>`,
  `--system-prompt-file <path>`, `--mcp-config <configs...>`, `--strict-mcp-config` are all
  **recognized**: each probe failed only on the trailing sentinel `--bogus-launcher-flag`
  ("error: unknown option", exit 1).
- **`--effort` with an unknown word is not refused**: `claude --effort bogus --version` prints
  `Warning: Unknown --effort value 'bogus' — ignoring it and using the default effort` and exits
  0. Effort vocabulary admission therefore rests entirely on the spawn plane's row vocabulary
  (§3.2); the tool will not catch a wrong word.
- `--mcp-config <configs...>` is **variadic** ("JSON files or strings, space-separated"): a
  native operand immediately following `--mcp-config <path> --strict-mcp-config` is safe because
  `--strict-mcp-config` terminates the list, which is exactly the SPEC §4.5 order (`with`
  companions after the path). A composition that put the native args directly after the path
  would have them swallowed as MCP config sources — the order is load-bearing for this flag.
- `resume` is **not** a claude subcommand (2.1.263 commands: agents, attach, auth, auto-mode,
  doctor, gateway, import, install, logs, mcp, plugin, project, respawn, rm, setup-token, stop,
  ultrareview, update); resume is `-r/--resume`. A `-- resume --last` tail on claude becomes
  the prompt "resume" plus an unknown option error — tool-owned, as the SPEC intends.
- Unknown flags before the operand: hard error, exit 1 — no silent forwarding at the tool.
- Nothing was written into the probe home by the failed invocations.

### 4.2 codex 0.153.4 (`07-…`, `07b-…`, `CODEX_HOME` = temp dir)

| Probe | Result | SPEC claim | Status |
|---|---|---|---|
| P1 `codex -p curator-mcp mcp list`, layer file absent | "No MCP servers configured yet", **exit 0** | missing layer silently ignored | **verified** |
| P3 layer = `[mcp_servers.probe-srv] command="true"` | server listed, exit 0 | mcp_servers-only layer composes | **verified** |
| P6 `-p curator-mcp -p other` | clap: "cannot be used multiple times", exit 2 | `-p` exactly once | **verified** |
| P7 `codex mcp list -p curator-mcp` | "unexpected argument '-p'", exit 2 | — | `-p` must precede the subcommand for `mcp`; `resume` accepts `-p` in both positions (P9/P10 printed help) |
| P8 layer path is a directory | "failed to load configuration … is not a file", exit 1 | `mcp_layer_unreadable` class | tool itself errors for a directory; the stat is still the only guard for *absence* |
| P2/P4/P11/P13 `--strict-config` | "`--strict-config` is not supported for `codex mcp` / `codex debug`" | missing layer ignored under `--strict-config` too | **not reproducible non-interactively** on 0.153.4; only runtime commands (`codex`, `exec`, `resume`, …) accept it. Stays **docs-confidence** (environments.md §7.8, 0.153.2 sprint evidence) |
| P12 `codex -p x debug models` | "`--profile` only applies to runtime commands and `codex mcp`" | — | informative |

`[PROMPT]` is codex's single positional; `-m`, `-c`, `-p` are global options parsed before or
after the prompt by clap, so the SPEC order is safe. For subcommand tails (`resume`, `fork`,
`exec`) the composed `-m … -c … -p …` prefix precedes the subcommand, which clap accepts (P9).

### 4.3 pi 0.84.2 (`04-pi-help.log`, installed `dist/cli/args.js` and `dist/core/resource-loader.js`)

Argv parser (`cli/args.js`, verified): `--model <pattern>` (next arg), `--system-prompt <text>`,
`--append-system-prompt <text>` (repeatable, pushed to a list), `@file` operands, and **every
other non-dash argument is pushed to `messages`** (the initial user turn). An unknown
`--flag` consumes the next non-dash argument as its value (`unknownFlags`), and a single-dash
unknown is an error. There is **no `--` terminator**: a literal `--` is treated as an unknown
flag named `""` that swallows the following operand — the launcher must never forward `--`
to pi (SPEC §3 already strips it). Consequence for SPEC §4.5: any pi native operand is a
message; channel flags after a native operand would still parse, but a native `--x` would
swallow the flag's value. The SPEC order is correct and load-bearing.

`resolvePromptInput(input)` (`resource-loader.js:16`): `existsSync(input)` → read the file;
a read error → **warns and uses the literal string**; non-existent → literal text. This is the
polymorphism environments.md §7.3 records; SPEC §5's "verify the path is a readable regular
file immediately before exec" is necessary and verified against source.

`SYSTEM.md`/`APPEND_SYSTEM.md` detection (`resource-loader.js:808–829`, verified):

- `discoverSystemPromptFile()`: `<cwd>/.pi/SYSTEM.md` **if the project is trusted**, else
  `<agentDir>/SYSTEM.md`; `discoverAppendSystemPromptFile()` likewise for `APPEND_SYSTEM.md`.
  `agentDir` = `$PI_CODING_AGENT_DIR` (`config.js`: `ENV_AGENT_DIR = "<APP>_CODING_AGENT_DIR"`).
- **A `--system-prompt` value suppresses SYSTEM.md discovery; a `--append-system-prompt` value
  suppresses APPEND_SYSTEM.md discovery** (`systemPromptSource ?? discover…()`,
  `if (!appendSources) { discover… }`). The file and the flag of the same semantics are
  **alternatives, not additive**. (Erratum E5.)
- The project-local `.pi/SYSTEM.md`/`.pi/APPEND_SYSTEM.md` in the launch cwd wins over the
  managed home's file when the project is trusted — a channel outside the managed home that
  the SPEC §5.1 probe does not see. (Erratum E5.)

The `agents-infra pi` wrapper (`~/.local/bin/agents-infra`, source `dee5403`
`tools/agents-infra/internal/infra/pi_launch_posix.go:137,332`): strips inherited
`PI_CODING_AGENT_DIR`, `PI_CODING_AGENT_SESSION_DIR`, `PI_SKIP_VERSION_CHECK`, `PI_TELEMETRY`
and re-sets them to its own cache-root home. Its `--help` is pi's own help (it forwards).

## 5. Contract matrix for A1 (what the implementation must code against)

| SPEC item | Verified fact | Source |
|---|---|---|
| §4.1 subprocess | `curator env resolve <env> [--profile p] --repair --format json`; stdout = one CCJ-1 line + LF; diagnostics on stderr as `curator: <code>: <detail>`; exit 1 on refusal, 2 on usage | §2.2, §2.3 |
| §4.1 digest | sha256 over the stdout bytes minus LF equals sha256 over CCJ-1 of the parsed object (curator prints canonical bytes) — the launcher must still recompute from the parsed object | §2.3 |
| §4.1 fragment members | `fragment, environment, profile{name,lock_sha256}, precedence{winner,placement}, env{<VAR>:path}`, optional `system_prompt`, `mcp`, `path_prepend` | §2.3, §10.2 |
| §4.2 mapping | agentic system ids `claude-code`, `codex`, `pi`; runtime ids `claude`, `codex`, none for pi | §3.6 |
| §4.3 level 3 | `vendorplugin.LineupOf(vendor)`; rows filtered by `Model.Systems` containing the system; `Model.Effort.Recommended`; `Support == EffortSupportNone` → no effort | §3.6 |
| §4.4 request | `SpawnRequest{Runtime, Model: ModelID, Effort, Home, WorkDir, Env: os.Environ()}` → `vendorplugin.BuildLaunch(ctx, reg, req, agentic.LaunchModeInteractive)`; wiring by blank-importing `systems/{claude,codex,pi}` and `vendors/{anthropic,openai}`; `providerlimits.Store.AvailabilityFor(VerdictQuery{Runtime, Model, Home})` for admission | §3.2 |
| §4.4 refusals | `ErrUnknownSystem`, `ErrUnsupportedLaunchMode`, `ErrModelMissing`, `ErrEffortMissing`, `ErrEffortNotTransportable`, `ErrCompositionNotInteractive`, `ErrParameterNotInteractive`; vendor-layer `ErrUnknownModel`, `ErrModelNotDrivenBySystem`, effort-vocabulary refusal | §3.3 |
| §4.5 argv | claude `--model M --effort E` + `--append-system-prompt-file P`/`--system-prompt-file P` + `--mcp-config P --strict-mcp-config` + tail; codex `-m M -c model_reasoning_effort="E"` + (config-key spelling: **docs-confidence**, `-c model_instructions_file=P` unverified) + `-p curator-mcp` + tail; pi `pi --model M` + `--append-system-prompt P` + tail — but see E1 | §3.4, §4 |
| §4.5 env | plan `Env` is a full filtered environment; use it as the base, not as an overlay on the inherited one (E4); fragment `env` names win | §3.5 |
| §4.5 codex stat | absence is silent at the tool (P1); directory errors at the tool (P8); stat remains required | §4.2 |
| §5.1 probe | `<home>/SYSTEM.md`, `<home>/APPEND_SYSTEM.md`; the tool also reads `<cwd>/.pi/<same>` when trusted (E5) | §4.3 |
| §5 pi flag | `--append-system-prompt <path>` verified readable before exec; it **replaces** file discovery for append (E5) | §4.3 |
| §4.6 untracked | exec `Plan.Binary` with composed argv; for pi this is `agents-infra` (E1) | §3.4 |
| §7 dependency | `github.com/relux-works/skill-agents-management v0.5.10` fetchable, signed tag | §3.1 |

## 6. Errata against SPEC 0.2.1-draft (evidence-backed; each needs its own story before code)

**E1 — pi plans through the agents-infra wrapper, which discards the managed home.** SPEC
§4.5/§4.6 treat "the interactive plan's `Binary`" as the tool. At `v0.5.10` the pi plugin
resolves `agents-infra` and argv `["pi","--model",M]` (§3.4), and the wrapper strips
`PI_CODING_AGENT_DIR` and sets its own cache-root home (§4.3). The fragment's
`PI_CODING_AGENT_DIR` (layer 3) is therefore overwritten inside the child before pi starts:
`curator run pi` can never run in the Curator-managed pi home, and Workstream B6 (retire
`agents-infra` launchers) would remove the launcher's own pi binary. Minimal correction:
either (a) agents-management gains a raw-`pi` interactive binary for a Curator-driven runtime
(a module change, tagged), or (b) SPEC §4.2 marks `pi` as `env_unsupported` in this revision
with the reason recorded, or (c) SPEC §4.5 admits a wrapper `Binary` and requires the wrapper
to honor an inherited home variable (an agents-infra change). Recommendation: (a) plus a
declared `pi` runtime (E2); until then (b).

**E2 — no vendor/runtime for `pi`; §4.3 level 3 and §4.4 cannot be satisfied.** No frozen
`RuntimeDeclaration` has `System: "pi"`; the pi system needs the `local-models` vendor and a
`local-models.toml` that is absent here (§3.6). `BuildLaunch` needs a `Runtime` id and a vendor
row; `BuildPlan` needs `Model.Effort` support from that row. SPEC §4.2's row `pi` → system `pi`
implies a runtime that does not exist at the tag. Correction: name the runtime id column in
§4.2 (`claude`, `codex`, and a pi runtime to be declared upstream), and state that `pi` is
`defaults_unresolvable`/`plan_refused` until a runtime exists. Ties to E1.

**E3 — the entry point is `vendorplugin.BuildLaunch` + `providerlimits`, not `BuildPlan`
alone.** `agentic.BuildPlan` takes a `Model{ID, Effort EffortSupport}` the launcher cannot fill
without the vendor row, does not validate the effort word, and never reads provider limits
(§3.2). SPEC §4.4 ("declared entry point `BuildPlan`", "admission is the spawn plane's",
"`plan_provider_limited`") describes `BuildLaunch` followed by
`providerlimits.Store.AvailabilityFor(VerdictQuery{Runtime, Model, Home})` — a read the
launcher must make itself, keyed by the managed home (M7). Correction: §4.4 names both calls
and the `SpawnRequest` members (`Runtime`, `Model`, `Effort`, `Home`, `WorkDir`, `Env`); §4.2
gains the runtime-id column (E2); §6 `plan` row maps the vendor-layer errors.

**E4 — §4.5 environment layer 2 must replace layer 1, not overlay it.** `Plan.Env` is the
plugin's complete filtered child environment built from `LaunchRequest.Env`; claude removes
`CLAUDECODE`, codex removes its runtime family (§3.5). "Inherited ⊕ plan Env, later overriding
earlier per name" re-inserts the removed names, so a `curator run claude_code` started from
inside a Claude session would be refused by the child as nested. Correction: layer 1 is the
input to the plan request (`LaunchRequest.Env`), and the composed environment starts from
`Plan.Env`; layers 3–4 apply on top. Decision 0013 D6.3 says the same "inherited ⊕ plan
`Env`" and needs the same one-line erratum.

**E5 — pi file-kind channels are not additive with the flag, and the tool reads a second
location.** SPEC §5.1 says that with `--system-prompt append` and `APPEND_SYSTEM.md` present
"both channels engage" and the launcher "MUST NOT deduplicate". In 0.84.2 a
`--append-system-prompt` value suppresses `APPEND_SYSTEM.md` discovery (and `--system-prompt`
suppresses `SYSTEM.md`), and a trusted project's `<cwd>/.pi/SYSTEM.md`/`APPEND_SYSTEM.md`
takes precedence over the agent-dir file (§4.3). Correction: §5.1 states that the flag
replaces file discovery of the same semantics on pi (the warning should say which one wins);
the probe set, or §9's residual list, records the project-local `.pi/` files as a channel
the launcher does not probe. environments.md §7.3's "applied unconditionally when the file
exists" needs the same qualification (curator-spec erratum).

**E6 — §4.1 diagnostic transport and the version warning.** Curator prints `curator:
<code>: <detail>` on stderr with exit 1 for `profile_unknown` and `environment_home_stale`
(§2.2), and emits `warning: environment_tool_version_unverified: …` on stderr with exit 0
when the detected tool version differs from the recorded one (§2.3) — the case on this
machine for claude and codex. SPEC §4.1 maps codes but does not say where they are read from,
and says nothing about passing Curator's warnings through. Correction (minor): §4.1 states
the stderr `curator: <code>` line as the code source and that Curator stderr is forwarded to
the operator verbatim; §6 `resolve` row unchanged.

Minor, not errata: §4.6's `ax` untracked behaviour cannot be exercised (`ax` absent, as
designed); codex "missing layer ignored under `--strict-config`" could not be re-verified on
0.153.4 non-interactively and stays docs-confidence (§4.2); the codex `model_instructions_file`
override spelling stays docs-confidence (already listed in §9).

## 7. Re-run on the installed curator (`11-resolve-installed.log`, verified)

With `~/.local/bin/curator` = `v0.14.1-0.20260907213730-04550e282705` (§1.1), for each of
`claude_code`, `codex_cli`, `pi`:

- `curator env resolve <env> --profile default --format json` (read-only, homes now current)
  → exit 0, stdout byte-identical to §2.3; stderr only the tool-version warning for
  claude/codex, empty for pi. This is the §4.1 "current home emits its fragment without any
  lock" path, observed.
- `curator env resolve <env> --profile default --repair --format json` → exit 0, same stdout,
  same stderr (no provisioning notice: nothing repaired).
- `curator env status` → the three homes `current, provisioned`; codex and pi show
  `passthrough: auth.json (file-link)`; claude `seeds: .claude.json`; `opencode` stays
  unprovisioned (never resolved).

The task-local candidate builds (§1) and the installed binary agree on every byte of the
three fragments. Evidence from the installed binary is the record of §2; the earlier
candidate runs are retained as the provisioning transcript (`06-…`).

## 8. Definition-of-done checklist mapping

- Versions and sanitized real fragment output for all three environments: §1, §2.3.
- Tagged agents-management API and interactive argv boundaries, codex layer, pi prompt files:
  §3, §4.
- Findings and errata attached as outcome resources; handed off to review: this file +
  evidence bundle; board `to-review`.
- Fact-checking: every "verified" row names its log or source path; docs-confidence rows are
  labelled.
- Questions from the task description: all A0 items answered; missing inputs named in §2.4.
