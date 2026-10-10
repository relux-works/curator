# Project surfaces beyond the Skillfile

Research for **TASK-261010-1992si — research-project-surfaces-coverage**, addressing [Curator issue #113][issue]. Collected 2026-10-10. **Recommendation: amend CIP-0002 with a versioned surface inventory, typed import/admission rules and adapter capability records.** Keep skill selection and reproducible package identity in the Skillfile and its lock; keep project admission, non-skill items and the derived profile/project composition outside the checkout. This follows the operator input already recorded in the draft, rather than introducing a competing project model. [cip]

Research proposal only: no implementation, harness, test or build. Findings are in §8 and task notes; the campaign prohibits editing `LOGBOOK.md`.

## 1. Decision, versions and evidence bounds

Decision: ownership and admission of project surfaces outside the Skillfile. Bounds: one document, no source archives, 60 KiB ceiling; consuming slice in §7. `grammar_frozen: not applicable`: no wire grammar or executable vectors are published.

Implementation baseline: `relux-works/curator` **`1d7eb18c24c4f156730f5c14aa4790a8c86ed891`**. CI pins curator-spec **`43bf0a2506d5c354a73bbc3ea4623d4653db10c7`**, `v1.0.0-rc.14`. Forward-looking design is read separately at PR #136 head **`2f0531b4edcc99c6118c277deb00e8736392c04d`**. CIP-0002 is still Draft: its operator input confirms the model but is not a released protocol amendment. [c-ci] [s-env] [cip] [pr]

“Pin” below means Curator's `VerifiedRelease` evidence baseline. The registry does not install or enforce that exact harness version; a mismatch is an unverified-version diagnostic. OpenCode has no recorded verified release. [c-reg] [s-env]

| Environment | Curator baseline | Latest stable publication observed on 2026-10-10 | Evidence binding |
|---|---|---|---|
| Claude Code | 2.1.261 | 2.1.296, published 2026-10-09 | Public changelog at baseline commit `d7dbd9a09f59775726ed14bbea8fc9dfdff62f7b`; latest tag commit `2301018b1f61073c501a8e7a4813ef48c239163b`. Detailed vendor docs are rolling, not an archived binary contract. [cl-pin] [cl-new] [cl-release] |
| Codex | 0.153.2 | 0.162.1, published 2026-10-09 | Exact source tags peel to `657a993cbee87acf52d14b758ce49dbd46d1b8eb` and `092d3acd6bec3e3a14bdc7e7a2810ab628ab759d`. [cx-release] [cx-doc-pin] [cx-doc-new] |
| OpenCode | **Unpinned** | 1.18.35, published 2026-10-06 | Latest tag commit `53d1eabb61e21162157817bf677da0a4ad3332e3`. Source inventory below concerns its V1 implementation; rolling V2 documentation is a separate contract. [oc-release] [oc-instructions] [oc-v2] |
| Pi | 0.84.2 | 1.1.0, published 2026-10-07 | Exact source commits `914cf1472e715297caa30db4b9535d534a9eb718` and `abe508e1b89912adde45528136c3221eb69acdd7`. The old repository URL redirects to `earendil-works/pi`; links below use that repository. [pi-release] [pi-usage] [pi-new-config] |
| Muse | 1.4.1-R4503.1 | 1.4.2 is the newest entry in the public changelog | The current build suffix and publication date are not established by that page. Rolling docs and the released Curator specification are available; exact-release implementation is not established. Do not promote a historical locally observed build into a “latest” claim. [mu-release] [s-env] |

Coverage concerns discovery mechanisms and resource families, not every scalar or undocumented feature. Unknown semantics mean refusal. **5/5 environments and all requested categories are addressed; exact baseline source: 2/5 (Codex, Pi); latest source: 3/5 (also OpenCode); runtime qualification: 0/5.** Claude/Muse rolling docs cannot prove exhaustive older-pin discovery; OpenCode lacks a baseline pin.

## 2. Per-environment inventory

`<project>` and `<managed-home>` are symbolic roots. Home-relative names describe product conventions, not actual personal locations. “B/L” means baseline/latest. A code such as **I** or **X** maps to the full resolution recipe in §3; each row's discovery locations must be excluded from native activation as described in §4. Native precedence is documented here to explain what the importer must capture; admitted precedence is deliberately defined separately.

### Claude Code

Baseline confidence is the released Curator contract plus the 2.1.261 changelog; detailed discovery is current vendor documentation unless a version boundary is named. Do not interpret an unchanged row as proof that every subfield existed in 2.1.261. [s-env] [cl-pin]

| Surface | Baseline and latest inventory | Project-layer disposition |
|---|---|---|
| Instructions and overrides | Ancestor/current `CLAUDE.md`, `.claude/CLAUDE.md`, `CLAUDE.local.md`; nested files load when relevant files are accessed. Local instructions supplement project instructions. Latest adds `AGENTS.md` support introduced in 2.1.277; it is not a baseline input automatically read by Claude. [cl-memory] [cl-new] | **I**: snapshot the selected chain, record original scope and selection, render approved text into the managed root. Imported `AGENTS.md` can be rendered for the old release without creating a checkout wrapper. |
| Conditional rules | `.claude/rules/**/*.md`, with `paths` frontmatter; unconditional rules are a separate case. Native malformed frontmatter can become unconditional text. The baseline changelog includes project-settings-source and symlink exclusion fixes. Latest 2.1.288 changes activation on Write/Edit, so startup/Read-only exclusion evidence is insufficient. [cl-memory] [cl-pin] [cl-new] | **R**: parse a named glob dialect; reject malformed or unknown applicability rather than widen it. Preserve runtime path conditions or refuse that capability. |
| Knowledge and memory | `@` imports form an eager reference graph; skill references can be read on demand. Auto-memory lives under the configuration root's `projects/<project>/memory/`; settings can redirect it. [cl-memory] [cl-skills] | **K** for approved references; **W** for mutable memory. Disable ambient memory initially; never treat generated notes as approval. |
| Settings and permissions | `.claude/settings.json`, `.claude/settings.local.json`; native order is managed → launch settings → local → project → user. Lists may merge. Permission rules, sandbox controls, environment assignments and executable settings must be distinguished from preferences. [cl-settings] | **S/P/X**: split by field; no wholesale settings copy. Preserve enforced policy and record the effective permission result. |
| MCP | Project `.mcp.json`; project-local records also exist in the user configuration's project map. Plugins, user servers and managed servers are further inputs. These are distinct from `.claude/settings.local.json`. [cl-mcp] | **M**: resolve each server and its dependencies, including helpers, as separate admitted items; render an exclusive set, including an explicit empty set. |
| Hooks and executable configuration | Hooks in project/local settings, skill or agent frontmatter, and plugin bundles. Current hooks include command, HTTP, prompt/agent and MCP handlers; exact event/type support needs a release dialect. Settings may also select helper executables. [cl-hooks] [cl-settings] | **X**, plus **M/P** dependencies. Importer reads declarations without invoking handlers or helpers. Each event/matcher/handler is separately visible in approval. |
| Plugins, extensions and bundled surfaces | Project/local plugin enablement and marketplace declarations can activate whole bundles. Components include skills, commands, agents, hooks, MCP/LSP, output styles, workflows, themes, monitors and `bin/` executables. Current mods execute JavaScript function hooks; terminal default availability is documented from 2.1.287, beyond the baseline. [cl-plugins] [cl-mods] | **X** bundle with a closed component manifest; decompose policy/MCP/context effects. Hash dependencies and executables, not only the plugin manifest. Unknown components refuse. |
| Subagents, styles and workflows | `.claude/agents/`, `.claude/output-styles/`, `.claude/workflows/`; project definitions can shadow personal ones, with nearest nested scope winning. Workflows are JavaScript; output styles alter system instructions. Latest adds `omitClaudeMd` for subagents in 2.1.271. [cl-agents] [cl-styles] [cl-workflows] [cl-new] | **A**, **I(system)** and **X** respectively. Do not flatten any of these into ordinary knowledge. Track child-agent entry points and inherited composition. |
| Commands and skills | `.claude/commands/`, `.claude/skills/<name>/SKILL.md`, plugin contributions and added-directory resources. Commands/skills can contain inline shell expansion and frontmatter capabilities. Gitignore is not an admission boundary; the changelog records discovery of commands/agents/styles ignoring it. [cl-skills] [cl-pin] | **C/T**, with **X/P/M** for active frontmatter or expansion. A textual slash command is not a shell dispatcher export. |

### Codex

The following uses exact source at both versions. Current official OpenAI documentation supplements source inspection; it is not substituted for older-release behavior. [cx-doc-pin] [cx-doc-new] [cx-official]

| Surface | Baseline and latest inventory | Project-layer disposition |
|---|---|---|
| Instructions and overrides | At each project directory: `AGENTS.override.md`, then `AGENTS.md`, then configured fallback filenames. Discovery runs from the project boundary to CWD; home instructions are separate. There is no native `@` import expander in these loaders. Latest rejects fallback entries that are not filenames, accounting for executor path syntax. [cx-doc-pin] [cx-doc-new] [cx-home] | **I**: retain per-directory selection, scope and ordered bytes; do not concatenate the overridden file as well. Suppress project discovery independently of the home output. |
| Execution-policy rules | `rules/*.rules` under effective configuration folders, including project `.codex/rules/`, are execution policy. Disabled project layers are excluded by the policy loader. This is not Markdown. Official documentation describes Starlark `prefix_rule` and most-restrictive matching decisions. [cx-policy-pin] [cx-policy-new] [cx-rules] | **P** with a versioned rule AST. Start with a literal, bounded subset; unsupported Starlark constructs refuse, rather than executing arbitrary input during inspection or turning policy into prose. |
| Native settings and policy | Trusted project `.codex/config.toml` layers, with ancestor/root discovery and worktree handling. Config can supply model/developer instructions, sandbox/approval and environment policy, paths and capabilities. Selecting `-p` is one layer, not exclusive project-config suppression. [cx-config-pin] [cx-config-new] [cx-local] | **S/P/I/X** split. Preserve requirements outside project authority; deny home/PATH/trust/provider redirection as ordinary project preferences. |
| Knowledge and mutable memory | Knowledge is instruction/skill data unless an explicit loader is selected. Memory has a separate home/state pipeline and artifacts such as `raw_memories.md`, rollout summaries and consolidated notes. A repository `.agents/memory` loader is **not established** by the inspected sources. [cx-mem-pin] [cx-mem-new] | **K/W**. Approved snapshots are separate from enabled native memory and its state; no copying personal rollouts into the public project layer. |
| MCP | `mcp_servers` in admitted config and plugin MCP declarations; server command/arguments, URL, environment and referenced authentication are consequential inputs. Curator's current MCP profile transport does not make other native layers disappear. [cx-toml-pin] [cx-plugin-pin] [c-reg] | **M**, with **X** for executable helpers and **P** for access effects. No direct project `.mcp.json` discovery is asserted independently of plugins. |
| Hooks | Native hook events in configuration and plugin hook sources; managed hook files and enablement/trust state are separate inputs. Worktree project-hook loading can refer to the root checkout configuration. [cx-hook-pin] [cx-hook-new] [cx-local] [cx-plugin-pin] | **X**. Include root-worktree source identity, event, matcher, handler and transitive files. A copied hook trust hash is not Curator approval. |
| Plugins and custom agents | Configured marketplaces/plugins contribute resources; plugins may carry skills, hooks, MCP and apps. `[agents.<role>]` can reference a separate `config_file`, with paths relative to the defining configuration. [cx-plugin-pin] [cx-plugin-new] [cx-toml-pin] [cx-toml-new] | **X/A**. Close role-file and bundle graphs; a role may request new model, instruction or permission settings, so admitting its description alone is insufficient. |
| Skills and commands | The inspected host resolver includes `.codex/skills` from project layers and `.agents/skills` from CWD to project root, plus home/admin/bundled/plugin and explicit roots. The two versions' `host_roots.rs` are identical. This function uses all project layers for roots, unlike trusted-layer execution policy. Skill metadata and executable resources remain part of the closure. [cx-skills-pin] [cx-skills-new] [cx-skill-loader] | **T/C/X**. Do not infer skill exclusion from untrusted config or a zero instruction budget. No independent project prompt-directory loader is established here; plugin/skill commands are inventoried instead. |

### OpenCode

There is no Curator-pinned release: the baseline column is **unknown**, not V1 or V2 by assumption. The table covers stable **1.18.35 source** and its in-tree V1 docs. Rolling V2 docs describe AGENTS-only dynamic instructions and currently unresolved `instructions` entries; those facts must not be mixed into the V1 adapter. [c-reg] [oc-release] [oc-instructions] [oc-v2]

| Surface | Latest stable inventory; baseline unpinned | Project-layer disposition |
|---|---|---|
| Instructions and overrides | Startup prefers project `AGENTS.md`, then compatibility `CLAUDE.md`, then deprecated `CONTEXT.md`; global instructions have a separate fallback. The read tool calls an additional nested-instruction resolver. There is no `AGENTS.override.md` entry in the inspected filename list. [oc-instructions] [oc-read] | **I**. Capture both startup and nested scope; preserve compatibility-dialect selection. Do not assume V2 semantics. |
| Rules and imported knowledge | `instructions` accepts local files/globs and URLs in V1. Ordinary Markdown `@` references are not automatically parsed. Cursor-style files can be explicitly selected, but selection alone does not interpret `.mdc` conditions. [oc-instructions] [oc-rules] | **I/R/K**. Freeze glob/URL results at acquisition; no network fetch during launch and no unqualified `.mdc` translation. |
| Reference repositories/directories | `references` in `opencode.json(c)` can identify local paths or Git repositories/ref selectors and descriptions. The docs say referenced directories are automatically admitted through the external-directory permission boundary. [oc-references] | **K + P**: this is both knowledge and a filesystem grant. Snapshot content and separately approve access; a moving Git branch is not a launch identity. |
| Settings and permission policy | Project `opencode.json(c)` and `.opencode/opencode.json(c)` merge after global/custom-file config. Config directories add resources. `{env:...}` and `{file:...}` substitution can read ambient inputs. Tool permissions are distinct from `experimental.policies`, whose documented action is `provider.use`; global policy takes priority over project policy. [oc-config] [oc-paths] [oc-vars] [oc-permissions] [oc-policies] | **S/P**. Parse before substitution; capture file references and environment names without reading secret values. Re-evaluate policy after composition. |
| MCP and other process servers | `mcp` plus LSP/formatter commands in native configuration can start programs or connect externally. Setting `OPENCODE_CONFIG` alone does not remove the higher project layer. [oc-config-doc] [oc-config] | **M/X**. Pin each command/server, dependencies and access scope; refuse untyped executable settings. |
| Plugins, tools and hooks | `.opencode/plugin(s)` local modules, configured npm plugins and dependency installation; `.opencode/tools/` JavaScript/TypeScript definitions; plugin events provide hooks. Custom tools can replace built-in tool names. Merely loading native config can start dependency acquisition. [oc-config] [oc-plugin] [oc-tools] | **X**. Inspect JSON/Markdown/JavaScript as data, never import the plugin to discover exports. Use a static admitted component declaration and approve tool replacement separately. |
| Agents and commands | `.opencode/agent(s)` and `command(s)` Markdown plus config declarations; compatibility `mode(s)` definitions become primary agents. Commands can reference files or run shell substitutions, and agent definitions can alter tools/permissions. [oc-command] [oc-agent] [oc-commands-doc] | **A/C**, with **X/P** dependencies. Resolve prompt references and classify executable substitutions before approval. |
| UI configuration and themes | Project `tui.json(c)` is separate from `opencode.json(c)`; `.opencode/themes/` supplies theme resources. Legacy theme/keybinding/TUI settings can trigger migration. [oc-config-doc] | **S** for allowlisted preferences and immutable theme data. Suppress original UI configuration and migration writes; a theme does not authorize plugin code. |
| Skills and memory | `.opencode/skill(s)`, project `.claude/skills`, `.agents/skills`, configured skill paths/URLs and personal roots are separate scans. No dedicated project memory-file discovery is established by this inspected set; instructions, references and plugin-defined memory remain possible. [oc-skills] [oc-skill-doc] | **T**, **K/W** for separately declared knowledge/memory. Suppress external skill scans as well as native config directories. |

### Pi

| Surface | 0.84.2 baseline and 1.1.0 latest | Project-layer disposition |
|---|---|---|
| Instructions and overrides | B/L: agent-home and ancestor-to-CWD context. Per directory, first readable candidate among `AGENTS.override.md`, `AGENTS.md`, `AGENTS.MD`, `CLAUDE.md`, `CLAUDE.MD`. Context loads independently of project trust; no native conditional-rules directory or `@` import expansion is established by these loaders. [pi-loader] [pi-new-loader] [pi-new-config] | **I**. Record exact selection and scope; materialize approved text separately from repository discovery. |
| System instruction files | B/L: `.pi/SYSTEM.md` and `.pi/APPEND_SYSTEM.md`, trust-gated; corresponding agent-home files also exist. The project candidate replaces the same-role global candidate, rather than concatenating both. Explicit prompt sources are another channel. [pi-loader] [pi-new-config] | **I(system-replace/system-append)**. A replacement requires its own approval and ceiling; never import it as harmless project prose. |
| Settings and startup state | B/L: `.pi/settings.json`, resource paths/packages, shell/tools/model/session preferences; project values override global values. Baseline `main.ts` constructs a project-reading settings manager for session selection before trust. Latest documents the `sessionDir` exception explicitly. [pi-settings] [pi-main] [pi-new-config] [pi-new-security] | **S/X/W**. Distinguish startup state routing from later trust-gated resources. `--no-approve` alone is insufficient for a strict no-raw-config-read claim. |
| Knowledge and memory | Context files, skill reference files and session/compaction state; no dedicated repository auto-memory directory is established by the inspected loader/configuration docs. Extensions can provide memory. [pi-loader] [pi-skills] [pi-new-config] | **K/W**; extension-backed memory additionally requires **X**. Session data never becomes an admitted declaration automatically. |
| MCP and execution policy | **Baseline explicitly has no built-in MCP or permission popups. Latest adds `.pi/mcp.json`** and agent-home `mcp.json`; project servers require trust. Same-name project entries can replace a server, or selectively override exposure/enabled state. Latest MCP env/header values can execute `!command`. [pi-usage] [pi-new-mcp] | Baseline MCP is **unsupported** unless a separately admitted **X** extension supplies it. Latest is **M/P/X**; reject command-valued secrets from the initial import subset. No universal permission-mode translation. |
| Extensions, packages and hooks | B/L: `.pi/extensions`, configured local/npm/Git packages and resource paths; packages bundle extensions/skills/prompts/themes. Trusted startup can install missing packages; extensions execute code and handle lifecycle events. [pi-packages] [pi-loader] [pi-new-packages] [pi-new-security] | **X**, with a pinned resource/dependency manifest. Do not run package managers, extension imports or lifecycle callbacks during discovery. |
| Skills | B/L: `.pi/skills`, project `.agents/skills` through the repo boundary, settings/package roots, explicit `--skill`. `--no-skills` disables discovery but retains explicitly supplied skills. Baseline also accepts standalone Markdown in Pi-specific skill roots. [pi-skills] [pi-new-skills] | **T**. Canonical Skillfile package or separately imported skill snapshot; preserve invocation policy and supporting closure. |
| Commands, prompts and themes | B/L: `.pi/prompts` templates, extension-registered commands, `/skill:name`, `.pi/themes` and package resources. Prompt files can be passed explicitly while default discovery is off. No separate native project subagent-definition directory is established; extension-provided agents remain executable integration. [pi-usage] [pi-prompts] [pi-new-config] [pi-new-cli] | **C/X/S**. Text templates remain prompt commands; extension commands are executable. Pure themes can be typed settings/resources but cannot authorize loading a package. |

### Muse

Baseline 1.4.1-R4503.1 has an XDG layout recorded by Curator, but no admitted root-instruction, system-prompt or MCP output channel; even discovery of the recorded skill path is unverified. The latest public docs accompany 1.4.2 without exact-build binding. Thus the table is a **documented candidate inventory for both-version qualification**, not a retroactive baseline support claim. [c-reg] [s-env] [mu-release]

| Surface | Documented inventory and explicit limits | Project-layer disposition |
|---|---|---|
| Instructions and prose rules | Workspace-to-nearest-`.git` chain; at each level the first existing `AGENTS.md`, `CLAUDE.md`, `.agents/AGENTS.md`, `.claude/CLAUDE.md` wins. Project beats user guidance; deeper project guidance wins. Project rules require workspace trust. Conditional rule files and override filenames beyond this list remain unknown. [mu-config] | **I/R**, only after a versioned output/discovery contract is established. Otherwise retain as candidates and refuse required activation. |
| Knowledge and memory | `.agents/memory/MEMORY.md` plus topic files; startup injects an index and up to 48 topic paths, with on-demand recall. Personal-project and personal memory are separate scopes. [mu-config] | **K/W**. Snapshot shared knowledge; keep writable memory private to the view and non-authoritative. Native index/observer reads require suppression too. |
| Native settings and permissions | User `settings.json`, managed defaults/policy and workspace trust/permission state. Docs describe workspace-scoped prefix grants; deny wins over allow. A native project `.muse/settings.json` loader is **not established** by these pages. [mu-config] [mu-permissions] | **S/P/W**. Import only confirmed dialects; never infer project settings from another harness's filenames. No project request may choose bypass or grant trust. |
| MCP | `mcp_servers` in settings, stdio or streamable HTTP, environment references and required/optional modes. A direct project `.mcp.json` loader is **not established**. [mu-extending] | **M**, gated on an admitted manager output channel. Unknown transport/target refuses; copying into an arbitrary project file is not a design. |
| Hooks, plugins and extensions | `.muse/hooks.json` project hooks; user hooks and `managed_hooks_path` are different sources. Project hooks require trust. Enabled plugins contribute skills; the reviewed pages do not establish a complete project plugin manifest/discovery grammar. [mu-extending] | **X** per event/handler; plugin closure remains unsupported until its grammar is known. Native malformed-hook warnings are not acceptable admission success. |
| Skills and slash invocation | Project `.agents/skills`, `.codex/skills`, `.claude/skills`; built-in, user, foreign personal and plugin roots also contribute. Current docs place user skills under the XDG config root, whereas Curator records an unverified XDG data-root location. [mu-extending] [s-env] | **T/C**. Preserve this discrepancy as an adapter qualification item, not an implicit registry correction. |
| Workflows and child agents | `.agents/workflows/*.js`, loaded at session start after trust; project names shadow user names. Children can use managed worktrees. Latest 1.4.2 changelog adds multiple workspace roots whose own rules use their recorded trust. [mu-workflows] [mu-extending] [mu-release] | **X/A**. Pin scripts and bind every workspace/child to the admitted composition. Extra roots cannot silently enlarge source discovery. |

### Cross-environment import-only formats

Issue #113/CIP-0002 include Cursor-style `.mdc` imports, without adding a Cursor launcher. `.cursor/rules/` needs its own description/globs/activation dialect; never strip unknown conditions into unconditional prose. Distinguish native imports, declared knowledge references and ordinary Markdown links. [issue] [cip] [oc-rules]

## 3. Resolution design for every surface

Everything below is **proposed**, elaborating CIP-0002's existing admission/precedence model. The recipes apply to every table row carrying their code; §4 supplies the environment-specific materialization and exclusion requirements. They are not claims about implemented support. [cip]

### Common discovery, normalization and admission

1. Bind discovery to the registered checkout and explicit external roots. Read native paths, ancestors, nested/worktree scopes and bundle paths as untrusted bytes. Distinguish `absent`, `unreadable`, `malformed`, `outside-boundary`, `unsupported-dialect` and `present`. Never run a harness, module, Git filter, MCP server, helper or package installer to inspect it.
2. Closed, versioned declaration: ID; kind; environment/release dialect; relative source; content/closure digests; scope/activation; load mode; dependencies; capabilities; requiredness; ordering/replacement intent. Native details require closed schemas too.
3. Freeze imports and finite glob results. Enforce file kind, no-follow/containment, byte/count/depth limits and cycle detection. External/URL references require approved acquisition of pinned bytes. Changing any dependency invalidates its approval, even if the top-level file matches.
4. Preview components, exclusions, order, permissions, command owners and profile changes. Public evidence uses neutral IDs/digests; source locations, memory and another operator's material stay private.
5. Operators approve policy-admissible items. Agents may approve only within a manager-issued ceiling bound to actor, project, kinds, environments, grants, destinations, validity and delegation. Bind approval to the reviewed digest and actual approver. Repository approval files, native trust stores and self-minted hashes confer no authority.
6. Compose over the selected profile lock; stacks resolve one shared closure. Record approval generation, profile/Skillfile digests, dialect/release, policy, order and output hashes separately. Do not mutate profiles or refresh sources at launch. Approve optional omissions; failed reads never trigger absence fallback.

These steps refine the discovery/admission cycle already proposed in CIP-0002; a future normative amendment must freeze exact field names, limits and refusal codes before implementation. [cip]

### Typed recipes: precedence, protected output and suppression

| Code / surface | Closed normalization and admission | Precedence against profile | Protected materialization and native discovery to suppress |
|---|---|---|---|
| **I — instructions** | UTF-8 bytes, exact import closure, scope, root/system role and append/replace intent. Replacement and altered system priority need explicit capability approval. | Preserve profile sequence; project-general follows, then project-environment, then deeper supported scope. Replacement is separate consent. This is text ordering, not semantic enforcement. | Render a regular managed file or an approved import tree; references target protected copies. Suppress original instruction chains, overrides, additional-root discovery, lazy reads and reloads. Do not put wrappers or symlinks in the checkout. |
| **R — conditional prose rules** | Named grammar for metadata and relative path predicates; distinguish unconditional, path-triggered and model-selected conditions. Reject malformed or unsupported selectors. | Add scoped project rules after applicable profile rules; deterministic explicit ordering. Conflicting requirements are shown, not supposedly solved by a filename. | Preserve the selector against the original project namespace. A rule whose original base moved into the managed home needs an adapter mapping, not unchanged relative globs. If the target cannot preserve activation, refuse or explicitly omit an optional item. Suppress original rule directories and imported aliases. |
| **K — knowledge** | Exact files/tree, eager/indexed/on-demand mode, description and finite reference graph. An external-directory access grant is a separate **P** dependency. | Additive, namespace by owner; same ID/digest deduplicates, differing content requires explicit replacement. | Protected files and index with rewritten bounded references. No mutable URLs, remote branch refresh or links back into the checkout. Suppress native knowledge indexes and auto-imports; ordinary code reads remain task data. |
| **W — mutable memory/state** | Explicit opt-in policy for scope, read/write channels, retention and export. Existing notes are candidates, not implicit imports. | Does not replace profile policy or project approval. Memory may inform a session only in its admitted scope. | Separate writable private state from immutable context. Disable ambient recall initially. Promotion to durable instructions takes another admission. Suppress native personal/project indexes, observers and session routing that escape the selected view. |
| **S — settings** | Closed per-release key allowlist with known types and effects; split instructions, execution and policy into their recipes. No generic native-config merger. | Project may override approved preference keys; enforced policy and manager identity/path controls remain above it. Resolve arrays/maps by key-specific rules, not universal last-wins. | Generate one effective managed config and explicit launch settings. Reject auth material, arbitrary interpolation and tool-owned state in immutable config. Suppress every project/local/ancestor setting source and runtime update path that can reintroduce it. |
| **P — execution/permission policy** | Typed native atoms and a bounded grammar; request effect, subject, scope, matcher and inherited constraints. Operator or delegated ceiling must authorize every new grant. | Keep hard denies/requirements; project may narrow, never silently weaken. Native ordered rules need semantic evaluation, not blind concatenation. | Emit protected policy/config with an effective-policy digest. Suppress original rule files, project sandbox/bypass settings and persisted native grants. Unknown semantics or an unverifiable ceiling refuse. |
| **M — MCP** | Stable server ID, transport, pinned executable closure or endpoint, argv, working-directory rule, environment/secret reference names, tool/access policy and requiredness. Helper commands are **X**. | Union; identical definitions deduplicate. Different same-ID server requires explicit whole-server replacement approval. Preserve profile/manager ceilings. | Generate a protected exclusive server set, even when empty. Resolve secrets through an admitted private channel; no secret values in the declaration. Suppress project, local-map, plugin and remote/native server sources outside the admitted set. |
| **X — executable integrations** | Static component manifest: event/entry point, module/executable digest, transitive package files, interpreter/runtime, arguments, environment references, access effects and component types. Hooks, plugins, LSP, formatters, workflow scripts and inline shell substitutions are included. | Add only explicitly enabled items. Name conflicts refuse unless replacement is approved; hooks require explicit order. Bundles cannot grant themselves priority. | Verified immutable dependency closure; native code executes only from protected copies under a qualified execution boundary. Separate writable caches/data. No runtime download, dependency discovery or package post-install surprise. Suppress plugin dirs, marketplaces, package auto-install and helper discovery. |
| **A — agent definitions** | Prompt plus model/tool/permission settings, imported role files, skills, MCP, hooks and nested-spawn policy. Normalize each active dependency separately. | Explicit same-ID replacement; child's permissions remain within the parent and manager ceiling. | Generate protected role files/arguments. Carry composition identity into descendants, worktrees and resume. Suppress repository role watchers, implicit child defaults and inherited foreign context. |
| **C — commands** | Distinguish prompt templates/slash commands, script commands and compiled shell exports. Parse placeholders; classify substitution as **X**. Skill-exported commands retain package ownership. | Unique effective command owner; project falls back to the selected profile, never machine-current. Replacement applies to the whole owning skill. | Prompt commands use protected native files; shell exports use CIP-0002's protected dispatcher. Suppress native command folders, plugin `bin/` additions and repository `.agents/bin` exposure. |
| **T — skills** | Skillfile-selected package identity and closure plus admission of active resources, invocation metadata and capabilities. Native undeclared skills need explicit import; they are not automatically enrolled. | Identical pins deduplicate. Approved replacement is whole-skill, including its context, commands and providers; re-resolve all constraints. | Install into the project view/store and expose only approved environment copies. Suppress all repository and ambient skill roots and plugin discoveries, including roots that do not follow config trust. |

“Protected” requires preventing the running agent and admitted extensions from replacing instructions, policy or approval records, including via writable parents or aliases. Read-only file bits owned by the same unrestricted user are not by themselves that boundary. Stable homes, publication transactions, leases, resume identity and revocation must retain CIP-0002's rules. A plain native home redirect supplies neither a filesystem sandbox nor proof against all prompt injection from task data. [cip] [s-env] [pi-new-security]

## 4. Adapter output and suppression contract

Current output capabilities below come from Curator's registry/materializer. Proposed extensions are conditional on version-specific evidence; **none is certified by this research**. [c-reg] [c-render] [c-managed]

| Environment | Where admitted copies would go | Required native controls and remaining limits |
|---|---|---|
| Claude | Existing managed `CLAUDE.md` and `skills/`; generated settings via a manager-owned settings source; protected agent/style/workflow/plugin trees and explicit MCP file only when their dialects are admitted. | Settings-source exclusion, instruction/rule exclusions, skill/command/plugin controls and strict MCP must cover distinct loaders. Baseline includes fixes for rules ignoring settings-source exclusion. `--safe-mode` is broad and can remove desired resources; `--bare` changes auth and resource behavior, and latest changes it again. No proven selective all-surface re-admission recipe is established here. [cl-pin] [cl-new] [cl-cli] [cl-mcp] |
| Codex | Existing managed `AGENTS.md` and skills; generated home/profile config, `rules/`, closed role files and admitted plugin components. `model_instructions_file` remains a distinct system-replacement channel. | Zero `project_doc_max_bytes` suppresses the project instruction loader, not settings or skills. Keep home overrides under manager control. Untrusted project layers address configuration/policy but the inspected skill-root resolver still enumerates project layers/`.agents` separately. Plugin, memory, custom-agent and execution-environment discovery need their own qualification. [cx-doc-pin] [cx-home] [cx-skills-pin] [cx-policy-pin] [c-reg] |
| OpenCode | Existing `<managed-xdg-config>/opencode/AGENTS.md`; V1 referenced output uses a generated `instructions` array. New output could include typed config and protected resource dirs only after choosing the release/launcher contract. | At 1.18.35 `OPENCODE_DISABLE_PROJECT_CONFIG` gates startup project files/config dirs. The inspected nested resolver has no equivalent guard and is called by the read tool; external skill scans have separate flags. This is a source-level counterexample to treating the one flag as blanket exclusion, not a runtime exploit measurement. Personal `.opencode` and compatibility roots also remain relevant. [oc-instructions] [oc-read] [oc-skills] [oc-paths] |
| Pi | Existing managed `AGENTS.md`, skills and optional system files; explicit protected `--skill`/`--prompt-template`/extension paths are possible channels. Latest needs a separately revised MCP adapter; the current registry has none. | `--no-context-files` disables home and project context, so it also disables the current managed root output. `--no-approve` skips trusted project resources, not ancestor instructions or all bootstrap config reads. Disable discovery per resource and re-admit explicitly. Baseline startup still reads project settings for session selection; a session-dir override bounds selection but does not prove no raw read. A strict launch cannot be certified from these flags alone. [pi-loader] [pi-usage] [pi-main] [pi-new-mcp] [c-reg] |
| Muse | Curator defines four managed XDG parents and a skill storage location. No root/system/MCP materialization target is admitted. Proposed resources must remain unresolved candidates until the adapter contract exists. | Whole-workspace trust can activate skills, rules, hooks and workflows together. XDG redirects do not establish suppression of foreign personal roots or repository memory. Current skill-location docs differ from the registry, and extra workspace roots add discovery. Refuse required strict capability; do not guess an `AGENTS.md` target. [s-env] [mu-extending] [mu-workflows] [mu-release] |

Capability records bind **environment × release × platform × surface × entry point**, with paths/exclusions. Cover startup, nested access, reload, added roots, resume, children and worktrees. Distinguish supported, unsupported and unknown; only supported evidence admits strict activation. Explicit legacy launches claim no per-item admission. [cip]

## 5. Skillfile boundary and repository-install migration

Today the closed Skillfile parser accepts skills, agent/locale/project metadata, and schema-2 sources; it has no arbitrary rule, memory, hook or settings block. Schema-2 resolve freezes source identities into `Skillfile.lock.json`, while machine bindings remain outside the project. Install currently chooses `<project>/.agents/skills` and `.agents/bin`, applies gitignore hygiene, and stages adapter mirrors such as `.codex/skills` and `.claude/skills`. The project-install adapter registry is a different set from the environment-home registry; Pi/Muse must not be assumed to have legacy mirrors just because they have managed-home adapters. [c-manifest] [c-cli] [c-install] [c-adapters] [c-reg]

| Remains in committed Skillfile / skill lock | Lives in manager-owned project layer / derived composition |
|---|---|
| Portable skill source declarations, selections, versions/revisions, agent/locale choices and frozen skill package/closure identities under the existing grammar. | Registered checkout binding; candidate imports of instructions, rules, knowledge, settings, MCP, hooks, agents, plugins and workflows; exact approved snapshots, dialects, capability effects and omissions. |
| Skill package exports and dependencies remain owned by that package; do not duplicate every packaged script or reference document into a second project manifest. | Admission links to the skill lock and package exports, approver/ceiling records, profile choice/stack, replacement decisions, effective policy and command owner map. |
| A lock is reproducibility input. Repository bytes, including a valid lock, do not authorize activation. | Private paths/secret references, writable memory/state policy, protected output hashes, leases and revocation. No credential values in either public declaration. |

This split is a recommendation under the operator's explicit Skillfile boundary, not a new Skillfile schema. Skill-declared MCP requirements or commands still need their normal package audit plus project activation admission; installing a skill must not implicitly approve unrelated project hooks. [cip] [c-manifest]

Proposed migration, without changing source ownership:

1. Read the existing Skillfile/lock. Preserve schema 1 through an equivalent immutable plan, without a silent schema upgrade. Installed artifacts and native-only skills are candidates, not authority.
2. Acquire or verify each locked skill snapshot and its dependencies in the external protected store. Match exact identities/digests; never bless drifted gitignored installed bytes solely because their names match the lock. Locally edited skills require an explicit source/snapshot decision.
3. Preview and approve the project skill selection and active exports. Compose the selected profile and project requirements jointly. Deduplicate identical pins; replace a whole same-identity skill only by explicit approval; reject incompatible dependencies and command collisions.
4. Materialize outside the checkout. Keep the Skillfile lock; store admission/view/launch records separately. Bind dispatcher and adapter outputs to this generation.
5. Once the adapter can suppress old roots, managed launches no longer consult repository skill mirrors or `.agents/bin`. Remove only demonstrably manager-owned generated entries through the existing ledger/marker ownership contract, after a reviewable migration preview. Leave authored files untouched. Do not remove a broad `.agents/` ignore entry blindly: that directory may also contain authored material or state. Keep legacy mode available explicitly until enrollment succeeds.
6. Launch/resume never refresh tags or copy fresh repository bytes into an admitted generation. Drift becomes another candidate; revocation blocks subsequent admissions and needs explicit handling for already running processes.

These steps reuse existing freeze/install ownership mechanics and CIP-0002's composition/lease model; they are not operations performed by this research. [c-cli] [c-install] [c-adapters] [cip]

## 6. Options and recommendation

| Option | Advantages | Costs / reason to choose or reject |
|---|---|---|
| Expand Skillfile into a universal native-configuration manifest | One portable entry point; familiar dependency lock. | Mixes committed declaration with private admission/state; needs a large grammar change and duplicates CIP-0002. Conflicts with the operator's retained skill-only responsibility. Reject for this scope. [cip] |
| Amend CIP-0002 with typed surface imports and release capability records | Keeps one owner for registration, approval, composition, protected views and launch refusal; incorporates the now-known release deltas. | Makes the draft larger; needs small separately versioned dialect contracts and explicit unsupported rows. **Recommended.** [cip] |
| New companion CIP for a general resource packaging protocol | Could serve independent package authors and multiple consumers. | Adds a second acceptance/dependency chain before this project feature; no independent protocol consumer is required by issue #113. Reserve for a later reusable interchange format, not this study's first slice. [issue] [cip] |
| Copy native config wholesale or trust the repository folder | Quick to reproduce native behavior. | Cannot provide per-item admission or a closed graph; imports, executable helpers, plugin defaults and late discovery bypass the proposed authority boundary. Reject. [cip] [oc-vars] [pi-packages] |

Amendment scope: inventory/confidence; closed item dialects; per-surface precedence; approval ceilings/identity; Skillfile migration; refusal on unresolved exclusion. CIP-0002 remains subject to normal spec adoption. [cip] [pr]

## 7. Smallest first slice

**Recommend an unconditional project-context/knowledge slice using the existing Codex monolithic serializer, ending at an admitted, protected project generation and preview.** Inputs: explicitly selected UTF-8 instruction files and finite local knowledge references, one environment selector, and a selected profile lock. No conditional rules, system replacement, live memory, native settings, MCP, executable integrations or new command dispatcher in this slice. It addresses an actual missing project-layer resolution path while reusing Curator's existing output rather than growing a second template engine. [c-render] [c-reg] [cip]

Freeze the declaration and refusal vectors, then connect discovery → parser → approval/ceiling → composition → protected store → real Codex materialization/preview. Require stable re-resolution and dependency-sensitive digests; reject bad reads, escapes, forged approvals and authority expansion. Represent project-only context explicitly: the renderer currently emits nothing when the root package declares no context. [c-render]

**Strict native launch remains refused until its separate source-control capability is proven.** That is a deliberate delivery boundary, not an “admitted launch” claim based on serialization. This study does not recommend starting with the earlier draft's Pi launch recipe unchanged: pinned Pi reads bootstrap project settings, and disabling context also removes its managed `AGENTS.md`; Codex likewise needs independent skill exclusion. Neither supplies a proven strict recipe from the inspected switches alone. [pi-main] [pi-loader] [cx-skills-pin] [cip]

Qualify the real loader/launcher on an authorized platform, without another general research prerequisite. If native controls cannot separate approved inputs from discovery, use a supported upstream control or an explicit filesystem/execution boundary with product review. Serialization alone does not establish activation capability. [cip]

Future activation checks need positive admitted and negative unapproved inputs at startup, nested access, reload, resume, child/worktree entry and added roots; empty/nonempty MCP; imported-byte drift; collisions; protected-file writes; and unsupported releases. **None ran on this host.**

## 8. Findings register and verification

| Finding / anomaly | Consequence |
|---|---|
| OpenCode is unpinned in Curator; latest V1 source and rolling V2 docs differ. [c-reg] [oc-instructions] [oc-v2] | Select an exact release/dialect before admitting output; no universal OpenCode contract. |
| Latest Pi adds native MCP absent from 0.84.2. [pi-usage] [pi-new-mcp] | A registry upgrade changes the control-surface inventory, not just the version string. |
| Pi baseline reads project settings for bootstrap session selection; latest documents it. [pi-main] [pi-new-security] | Whole-folder trust denial is not proof of no project-config read. |
| OpenCode V1 nested instruction resolution and external skill scans are distinct from the project-config gate. [oc-instructions] [oc-read] [oc-skills] | A startup-only flag claim misses production call sites. |
| Codex skill roots use all project layers and a separate `.agents` walk. [cx-skills-pin] [cx-skills-new] | Instruction-budget/config-trust evidence must not be reused as skill-discovery proof. |
| Claude adds AGENTS support, later Write/Edit rule activation and mods after the recorded baseline. [cl-new] [cl-mods] | Release capability records must cover late discovery and executable bundle changes. |
| OpenCode references imply external-directory access; latest Pi MCP can run command-valued env/header inputs. [oc-references] [pi-new-mcp] | “Knowledge” and “MCP configuration” are not inherently passive item kinds. |
| Muse docs and Curator disagree on the advertised user-skill location; output channels remain unverified. [mu-extending] [s-env] | Record drift and refuse unknown activation; do not silently change the registry. |

Fact-checking: Curator/released-spec source, exact upstream trees/files, public latest-release queries, vendor docs, changelogs and baseline/latest comparison. Earlier studies supplied navigation only; **no prior runtime evidence was rerun or accepted as qualification**.

Read failures: Claude Markdown returned 403; vendor HTML/changelogs supplied evidence. A guessed Codex Markdown URL returned 404; the official page worked. Muse workflows required its vendor link. Missing-path searches and board-query/schema errors were corrected, never treated as absence or passing validation.

Editorial verification (standalone processes): Python citation/source/hygiene audit first exited **1** for exceeding 60 KiB; the shortened document rerun exited **0**. It resolved **87/87** citations, verified **50** upstream Git blobs and **8** local files against pinned Git content, and checked UTF-8/LF, whitespace and private-locator patterns. Four other pinned snapshots and 25 web/release links have the evidence bounds above. `git diff --quiet <baseline> -- LOGBOOK.md` and `git diff --quiet <baseline> --` each exited **0**. HEAD remains the baseline; status shows only this uncommitted research file. These are editorial checks, not runtime gates.

Brief coverage: question 1 → §§1–2; question 2 → §§3–4; question 3 → §5; question 4 → §§6–7. Unknown native behavior is explicitly represented rather than silently filled in. One task-scoped outcome uses the requested basename `project-surfaces-coverage.md` under a task-ID resource path.

## References

All repository links pin path and full commit. Release links identify exact tags; latest-release queries were read on 2026-10-10. Vendor URLs are rolling documentation accessed on that date, with version attribution limited as described in §1. Symbolic paths in the study are product conventions only.

[issue]: https://github.com/relux-works/curator/issues/113
[pr]: https://github.com/relux-works/curator-spec/pull/136
[cip]: https://github.com/relux-works/curator-spec/blob/2f0531b4edcc99c6118c277deb00e8736392c04d/cips/CIP-0002-project-context-in-managed-launches.md
[s-env]: https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md
[c-ci]: https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/.github/workflows/ci.yml
[c-reg]: https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/internal/envregistry/envregistry.go
[c-manifest]: https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/internal/manifest/manifest.go
[c-cli]: https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/docs/cli.md
[c-install]: https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/internal/install/install.go
[c-adapters]: https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/internal/adapters/adapters.go
[c-render]: https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/internal/contextmaterialize/contextmaterialize.go
[c-managed]: https://github.com/relux-works/curator/blob/1d7eb18c24c4f156730f5c14aa4790a8c86ed891/internal/envprofile/managed.go
[cl-pin]: https://github.com/anthropics/claude-code/blob/d7dbd9a09f59775726ed14bbea8fc9dfdff62f7b/CHANGELOG.md
[cl-new]: https://github.com/anthropics/claude-code/blob/2301018b1f61073c501a8e7a4813ef48c239163b/CHANGELOG.md
[cl-release]: https://github.com/anthropics/claude-code/releases/tag/v2.1.296
[cl-memory]: https://code.claude.com/docs/en/memory
[cl-skills]: https://code.claude.com/docs/en/skills
[cl-settings]: https://code.claude.com/docs/en/settings
[cl-mcp]: https://code.claude.com/docs/en/mcp
[cl-hooks]: https://code.claude.com/docs/en/hooks
[cl-plugins]: https://code.claude.com/docs/en/plugins-reference
[cl-mods]: https://code.claude.com/docs/en/plugins/mods/overview
[cl-agents]: https://code.claude.com/docs/en/sub-agents
[cl-styles]: https://code.claude.com/docs/en/output-styles
[cl-workflows]: https://code.claude.com/docs/en/workflows
[cl-cli]: https://code.claude.com/docs/en/cli-reference
[cx-release]: https://github.com/openai/codex/releases/tag/rust-v0.162.1
[cx-doc-pin]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/core/src/agents_md.rs
[cx-doc-new]: https://github.com/openai/codex/blob/092d3acd6bec3e3a14bdc7e7a2810ab628ab759d/codex-rs/core/src/agents_md.rs
[cx-home]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/codex-home/src/instructions/mod.rs
[cx-official]: https://learn.chatgpt.com/docs/agent-configuration/agents-md
[cx-rules]: https://learn.chatgpt.com/docs/agent-configuration/rules
[cx-policy-pin]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/core/src/exec_policy.rs
[cx-policy-new]: https://github.com/openai/codex/blob/092d3acd6bec3e3a14bdc7e7a2810ab628ab759d/codex-rs/core/src/exec_policy.rs
[cx-config-pin]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/config/src/loader/mod.rs
[cx-config-new]: https://github.com/openai/codex/blob/092d3acd6bec3e3a14bdc7e7a2810ab628ab759d/codex-rs/config/src/loader/mod.rs
[cx-local]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/config/src/loader/local.rs
[cx-mem-pin]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/memories/README.md
[cx-mem-new]: https://github.com/openai/codex/blob/092d3acd6bec3e3a14bdc7e7a2810ab628ab759d/codex-rs/memories/README.md
[cx-toml-pin]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/config/src/config_toml.rs
[cx-toml-new]: https://github.com/openai/codex/blob/092d3acd6bec3e3a14bdc7e7a2810ab628ab759d/codex-rs/config/src/config_toml.rs
[cx-plugin-pin]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/core-plugins/src/loader.rs
[cx-plugin-new]: https://github.com/openai/codex/blob/092d3acd6bec3e3a14bdc7e7a2810ab628ab759d/codex-rs/core-plugins/src/loader.rs
[cx-hook-pin]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/config/src/hook_config.rs
[cx-hook-new]: https://github.com/openai/codex/blob/092d3acd6bec3e3a14bdc7e7a2810ab628ab759d/codex-rs/config/src/hook_config.rs
[cx-skills-pin]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/ext/skills/src/host_roots.rs
[cx-skills-new]: https://github.com/openai/codex/blob/092d3acd6bec3e3a14bdc7e7a2810ab628ab759d/codex-rs/ext/skills/src/host_roots.rs
[cx-skill-loader]: https://github.com/openai/codex/blob/657a993cbee87acf52d14b758ce49dbd46d1b8eb/codex-rs/ext/skills/src/loader/discovery.rs
[oc-release]: https://github.com/anomalyco/opencode/releases/tag/v1.18.35
[oc-v2]: https://opencode.ai/v2/docs/instructions
[oc-instructions]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/session/instruction.ts
[oc-read]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/tool/read.ts
[oc-rules]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/web/src/content/docs/rules.mdx
[oc-references]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/web/src/content/docs/references.mdx
[oc-config]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/config/config.ts
[oc-config-doc]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/web/src/content/docs/config.mdx
[oc-paths]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/config/paths.ts
[oc-vars]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/config/variable.ts
[oc-permissions]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/web/src/content/docs/permissions.mdx
[oc-policies]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/web/src/content/docs/policies.mdx
[oc-plugin]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/config/plugin.ts
[oc-tools]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/web/src/content/docs/custom-tools.mdx
[oc-command]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/config/command.ts
[oc-agent]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/config/agent.ts
[oc-commands-doc]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/web/src/content/docs/commands.mdx
[oc-skills]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/opencode/src/skill/index.ts
[oc-skill-doc]: https://github.com/anomalyco/opencode/blob/53d1eabb61e21162157817bf677da0a4ad3332e3/packages/web/src/content/docs/skills.mdx
[pi-release]: https://github.com/earendil-works/pi/releases/tag/v1.1.0
[pi-usage]: https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/docs/usage.md
[pi-loader]: https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/src/core/resource-loader.ts
[pi-main]: https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/src/main.ts
[pi-settings]: https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/docs/settings.md
[pi-skills]: https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/docs/skills.md
[pi-prompts]: https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/docs/prompt-templates.md
[pi-packages]: https://github.com/earendil-works/pi/blob/914cf1472e715297caa30db4b9535d534a9eb718/packages/coding-agent/docs/packages.md
[pi-new-config]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/configuration.md
[pi-new-loader]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/src/core/resource-loader.ts
[pi-new-security]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/security.md
[pi-new-mcp]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/mcp.md
[pi-new-skills]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/skills.md
[pi-new-packages]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/packages.md
[pi-new-cli]: https://github.com/earendil-works/pi/blob/abe508e1b89912adde45528136c3221eb69acdd7/packages/coding-agent/docs/cli.md
[mu-release]: https://dev.meta.ai/docs/muse-code/changelog
[mu-config]: https://dev.meta.ai/docs/muse-code/configuration
[mu-extending]: https://dev.meta.ai/docs/muse-code/extending
[mu-permissions]: https://dev.meta.ai/docs/muse-code/permissions
[mu-workflows]: https://dev.meta.ai/docs/muse-code/workflows
