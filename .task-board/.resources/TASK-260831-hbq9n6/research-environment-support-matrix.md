# Research: agent environment support matrix for global context management

Task: TASK-260831-hbq9n6 · Story: STORY-260831-6bbhow · Date: 2026-08-31

Verification legend: **local** = confirmed against the binary installed on this
machine (help output, string table, or an actual run that wrote files);
**docs** = vendor documentation only.

Versions checked locally: claude 2.1.251, codex-cli 0.151.0, pi 0.84.2,
gemini-cli 0.54.4. opencode, cursor-agent, copilot, goose, crush, amp, aider,
qwen-code, droid: docs only.

## 1. Home-directory isolation mechanisms

| Environment | Mechanism | Semantics | Confidence |
| --- | --- | --- | --- |
| claude_code | `CLAUDE_CONFIG_DIR` | names the config home itself (default `~/.claude`); `.claude.json` state file lives at its root | local |
| codex_cli | `CODEX_HOME` | names the home itself (default `~/.codex`): `config.toml`, `AGENTS.md`, `skills/`, `prompts/`, profiles, `auth.json` | local |
| opencode | `XDG_CONFIG_HOME` | names the parent; tool reads `<parent>/opencode/`. No dedicated variable. `OPENCODE_CONFIG` / `OPENCODE_CONFIG_CONTENT` override the config file only | docs |
| pi | `PI_CODING_AGENT_DIR` | names the agent dir itself (default `~/.pi/agent`) | local (run wrote `auth.json`, `models-store.json` into target) |
| gemini | `GEMINI_CLI_HOME` | names the parent; tool creates `<parent>/.gemini/` | local (run created `.gemini/{projects.json,history,tmp}`) |
| cursor (cursor-agent) | `CURSOR_CONFIG_DIR` (Linux/BSD also `$XDG_CONFIG_HOME/cursor/cli-config.json`) | names the config dir | docs |
| copilot CLI | `COPILOT_HOME` | names the home (default `~/.copilot`); XDG support is dot-prefixed and non-conforming — prefer `COPILOT_HOME` | docs |
| goose | `GOOSE_PATH_ROOT` | names the parent; tool creates `config/`, `data/`, `state/` beneath it | docs |
| crush | `CRUSH_GLOBAL_DATA`, `CRUSH_SKILLS_DIR` | partial (data dir + skills dir); no single home variable | docs |
| amp | `AMP_SETTINGS_FILE`, XDG | config-file override only | docs |
| aider | `--config FILE`, `AIDER_*` env prefix | config-file override only | docs |
| qwen-code | `QWEN_*` per-setting | no confirmed dir swap | docs |
| droid (Factory) | none found | `~/.factory` appears hardcoded | docs (negative) |
| Xcode 26.3 CodingAssistant | fixed path | `~/Library/Developer/Xcode/CodingAssistant/ClaudeAgentConfig/` (`.claude.json` inside), Codex analog beside it; restricted env, does not inherit shell config | docs |

Fallback everywhere: `HOME=<profile-home>` relocates every tool, but drags
`~/.gitconfig`, `~/.ssh`, `~/.npmrc`, shell config, and can break
Keychain-backed auth on macOS. Usable only with a seeded home (symlinks to the
real `.gitconfig`/`.ssh`); last resort for tier-3 tools.

Key gotcha: the variables are inconsistent about what they point at — some name
the config dir itself (`CODEX_HOME`, `CLAUDE_CONFIG_DIR`, `PI_CODING_AGENT_DIR`,
`COPILOT_HOME`, `CURSOR_CONFIG_DIR`), some name the parent and the tool appends
its own subdirectory (`GEMINI_CLI_HOME` → `.gemini/`, `GOOSE_PATH_ROOT` →
`config|data|state/`, `XDG_CONFIG_HOME` → `opencode/`). An adapter registry must
encode per-tool layout; a single shape cannot be assumed.

## 2. Global root instruction files

| Environment | Global root context file | Notes | Confidence |
| --- | --- | --- | --- |
| claude_code | `<home>/CLAUDE.md` | supports `@path` imports; project CLAUDE.md discovered separately | local |
| codex_cli | `<home>/AGENTS.md` | no per-file disable flag; `project_doc_max_bytes=0` kills only project docs; `model_instructions_file` replaces base instructions | local |
| opencode | `~/.config/opencode/AGENTS.md` | also reads `~/.claude/CLAUDE.md` unless `OPENCODE_DISABLE_CLAUDE_CODE_PROMPT=1`; `instructions` config key appends more files | docs |
| pi | `<agentDir>/AGENTS.override.md` → `AGENTS.md` → `AGENTS.MD` → `CLAUDE.md` → `CLAUDE.MD` (first match), plus `<agentDir>/APPEND_SYSTEM.md` appended to system prompt | `--no-context-files` disables discovery | local (dist source) |
| gemini | `<home>/GEMINI.md` | filename configurable via `contextFileName` setting | docs |
| cursor | global rules dir `~/.cursor/rules` | AGENTS.md/CLAUDE.md read at project root only | docs |
| copilot | `<home>/` instructions (global copilot-instructions) | docs |
| goose | `.goosehints` global variant in config dir | docs |

## 3. Configurable surfaces inside each global home

Surfaces: R = root instructions, S = skills, M = MCP servers, C = settings /
permissions / model config, A = subagents, P = prompts / commands, H = hooks,
T = themes, K = memory / knowledge. "✓" = exists in the tool's global home;
"—" = no such global surface found.

| Environment | R | S | M | C | A | P | H | T | K |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| claude_code | ✓ `CLAUDE.md` | ✓ `skills/` | ✓ `~/.claude.json` `mcpServers` (user scope) | ✓ `settings.json` | ✓ `agents/` | ✓ `commands/` | ✓ in `settings.json` | ✓ output-styles, keybindings | ✓ auto-memory under `projects/<dir>/memory` |
| codex_cli | ✓ `AGENTS.md` | ✓ `skills/` (+ `skills.config` in config.toml) | ✓ `config.toml` `mcp_servers` | ✓ `config.toml` (+ `<name>.config.toml` profiles) | multi-agent config (`agents/` roles TOML) | ✓ `prompts/` | ✓ hooks in config | ✓ themes in config | ✓ `memories/` (experimental) |
| opencode | ✓ `AGENTS.md` | ✓ `skills/` | ✓ `opencode.json` `mcp` | ✓ `opencode.json` | ✓ `agents/` | ✓ `commands/` | ✓ `plugins/` | ✓ `themes/` | — |
| pi | ✓ `AGENTS.md` + `APPEND_SYSTEM.md` | ✓ `skills/` | — (no MCP; extensions instead) | ✓ `settings.json` | — | ✓ `prompt-templates/` | via extensions | ✓ `themes/` | — |
| gemini | ✓ `GEMINI.md` | ✓ `skills/` (0.54+) | ✓ `settings.json` `mcpServers` | ✓ `settings.json` + policies | — | ✓ `commands/` | ✓ hooks (0.54) | — | — |
| cursor | ✓ `rules/` | ✓ (rules-as-skills) | ✓ `mcp.json` | ✓ `cli-config.json` | — | — | — | — | — |
| windsurf | — | native `~/.agents/skills` | ✓ `~/.codeium/windsurf/mcp_config.json` | — | — | — | — | — | — |

## 4. Credential and mutable state placement (managed-home hazard map)

| Environment | Credentials | Mutable session state in home |
| --- | --- | --- |
| claude_code | macOS Keychain (survives home swap); `.credentials.json` in home on Linux; `.claude.json` holds userID/onboarding state | `.claude.json` (per-project trust, MCP approvals, metrics), `projects/` (sessions, memory), `todos/`, `statsig/` |
| codex_cli | `auth.json` in `CODEX_HOME` | `sessions/`/thread history, `history.jsonl`, sqlite stores (`goals_1.sqlite`, `memories_1.sqlite`, `logs_2.sqlite`), caches |
| opencode | auth in data dir (`~/.local/share/opencode`), NOT config dir — XDG_DATA_HOME untouched by config swap → shared automatically | session storage in data dir |
| pi | `auth.json` in agent dir | `sessions/`, `models-store.json` |
| gemini | `oauth_creds.json` in `.gemini` | `history/`, `projects.json`, `tmp/` |

Consequences for the design: (a) auth must be shared or seeded across profile
homes per adapter — it is never profile content; (b) sessions land inside the
profile home for codex/pi/gemini/claude → session resume is naturally
profile-affine; the launcher and ax must reproduce the same home to resume.

## 5. Existing Curator assets that generalize

- Adapter model + `.csk-managed.json` ledger (protocol core §11, manager §5):
  symlink/copy materialization with never-overwrite-unmanaged discipline.
- Global scope (manager §4.2): machine-local Skillfile + store mirrored into
  each agent home — becomes "one profile" under the new model.
- Canonical git source identity + exact refs (§6.1–6.3): reuse verbatim for
  profile repositories.
- Content hash (§8): reuse for context IR output determinism + drift detection.
- MCP surface table (manager §6): read-only verification today; the same file
  map is the write-path target for a later revision.

## 6. ax (agent-session-manager) integration surface

ASM spec v0.5.0: provider plugin `launch`/`resume`/`fork` return a `SpawnPlan`
containing `argv`, `cwd`, `env_names` (pass-through allowlist, 0..64) and
`env_literals` (map name→value, 0..64). Directory invariants require spawn via
argv + explicit cwd + environment allowlist, no shell. So an environment
fragment of literal env vars is exactly the shape ax can inject; no new ax
mechanism is required — only a contract for who computes the fragment and where
it is recorded (Session Record extensions) for resume fidelity.

## 7. Sources

- Local binaries/dist: claude 2.1.251 (string table: `CLAUDE_CONFIG_DIR`
  resolution, disable flags), codex 0.151.0 (config keys, skills config), pi
  0.84.2 (`loadContextFileFromDir`, `PI_CODING_AGENT_DIR`), gemini 0.54.4
  (`GEMINI_CLI_HOME` fallback chain; empirical home creation).
- https://opencode.ai/docs/config/ · https://opencode.ai/docs/rules/
- https://pi.dev/docs/latest/usage · https://github.com/badlogic/pi-mono
- https://cursor.com/docs/cli/reference/configuration
- https://docs.github.com/en/copilot/reference/copilot-cli-reference/cli-config-dir-reference
- https://goose-docs.ai/docs/guides/config-files/ (+ `GOOSE_PATH_ROOT` env-vars doc)
- https://deepwiki.com/charmbracelet/crush/2.2-configuration
- https://aider.chat/docs/config/aider_conf.html
- https://qwenlm.github.io/qwen-code-docs/en/users/configuration/settings/
- https://docs.factory.ai/reference/cli-reference
- https://zenn.dev/usagimaru/articles/42cddb457a204e (Xcode ClaudeAgentConfig)
- agent-session-manager-spec SPEC.md v0.5.0 (SpawnPlan, DIR-INV spawn rules)
