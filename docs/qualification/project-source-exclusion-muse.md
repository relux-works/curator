# Muse 1.4.11 project-source exclusion

TASK-261021-4bo0a7, 2026-10-21. Research only: **complete strict project-context
exclusion is not qualified**. Fresh untrusted headless launches exclude the
measured startup sources while retaining selected home copies. Lazy tool reads,
native child execution and permission-policy enforcement remain unmeasured;
interactive entry and resume have additional trust behavior.

## Identity and method

- Muse Code `1.4.11-R6667.1`, macOS/arm64, SHA-256
  `6c04ce1719844685153ef6f04b98113cb61b942d97f0a8b0bfb3932c343984d9`.
- Curator source baseline `cdcba3bd7a8de6028d3d1dbf621ce4cb19b130a7`.
- Design read: curator-spec `e41f96387287f5e76604627f9e63662f7d6d86d4`,
  [CIP-0002](https://github.com/relux-works/curator-spec/blob/e41f96387287f5e76604627f9e63662f7d6d86d4/cips/CIP-0002-project-context-in-managed-launches.md).
  Its admission boundary includes startup, nested discovery, reload, resume and
  children. Startup success alone does not satisfy that boundary.
- [Harness and replay instructions](../../.research/project-exclusion-muse/README.md),
  [native startup excerpts](../../.research/project-exclusion-muse/evidence/observations.txt),
  [test transcript](../../.research/project-exclusion-muse/evidence/gated.log),
  [control transcript](../../.research/project-exclusion-muse/evidence/controls.log).

Every session uses a digest-checked executable, `env -i`, disposable HOME/XDG,
`MUSE_NO_AUTO_UPDATE=1`, `--provider echo`, no approval judge or reminder agents,
and an outer macOS sandbox denying network and real native/credential stores.
No real credentials or remote inference are involved. Cases use separate real
Git checkouts with external metadata. Trust never escapes the synthetic home.
The sandbox's network and file denials have positive and negative controls.

The startup control is an **empty, newly created trust store and no
`--trust-workspace` or `--yolo`**. Merely omitting a flag in a reused trusted home
is not this recipe. The neighboring baseline adds `--trust-workspace`.
Selected synthetic home copies use `$XDG_CONFIG_HOME/muse/AGENTS.md`,
`skills/<name>/SKILL.md` and `settings.json`. This demonstrates native delivery,
not manager approval or immutable protection. The fixture owns both home and
checkout; same-UID hostile-code containment is not qualified.

## P1-P8 observations

| Row | Source and native evidence | Exclusion verdict and exact scope | Selected home positive and remaining gap |
| --- | --- | --- | --- |
| P1 | Individually planted `AGENTS.md`, `CLAUDE.md`, `.agents/AGENTS.md`, `.claude/CLAUDE.md`, at checkout root or an ancestor of a nested launch. Trusted startup reports `project_sources=1`. A source above the Git root is not discovered in either trust mode. | **Excludable at exec startup** with the fresh-untrusted recipe: `project_sources=0`, home `user_sources=1`. | Home `AGENTS.md` loads in every case. Each filename is tested separately; precedence, imports and every possible rule dialect are not certified. |
| P2 | Starting in `sub/` with root `AGENTS.md` and nested `CLAUDE.md` reports two trusted project sources. | **Excludable for nested startup**: untrusted reports zero, home one. **Lazy entry unmeasured**: echo cannot issue the real read/entry tool. | Home root remains loaded. A separate launch in a subdirectory is not a lazy discovery test. |
| P3 | Project `.agents/skills`, `.claude/skills`, `.codex/skills`: trusted catalogs contain one project skill beside one distinct home skill. | **Excludable at exec startup**: untrusted `project_skills=0`, `user_skills=1`. | Native home skill catalog still loads. Skill invocation, later reload and plugin-provided skills need separate coverage. |
| P4 hooks | Project `.muse/hooks.json`, `UserPromptSubmit` command: trusted catalog has two runnable handlers including the selected user hook, and both marker files are written. | **Excludable at exec startup**: untrusted project handler is inactive; only the user marker is written. `handlers=2`, `runnable=1`, `inactive=1`. | The selected `settings.json` `hooks` member executes its marker hook while project execution is absent. Other hook events are not certified. |
| P4 plugins | A raw valid repository plugin bundle is inert. Explicit `plugins install <path> --scope project --json` refuses in an untrusted checkout; a trusted project install admits its command capability. | **Measured admission gate**, not passive project-plugin auto-discovery. No complete ambient plugin exclusion claim. | The same bundle copied outside the checkout and explicitly installed `--scope user` gives `installed_plugins=1`, commands increment from one built-in to two. Cached installs, other capability families and project config dialects remain outside this observation. |
| P5 | Project `.mcp.json`, both `{mcpServers:{}}` and a non-empty stdio definition, appears as `trusted_project_source_count=1` only when trusted. | **Excludable configuration resolution at exec startup**: untrusted project count zero while selected home source remains one. | Home `settings.json` `mcpServers` remains admitted. Empty home MCP does **not** impose an exclusive empty set in a trusted checkout. `/usr/bin/true` is deliberately only a config marker; successful server handshakes, tools and authorization are unmeasured. |
| P6 | Poison parser-control bytes in project `.muse/settings.json`, `.claude/settings.json` and `.claude/settings.local.json` do not stop a trusted echo launch. The identical user `settings.json` bytes refuse with `malformed settings file`, exit 1. | **Non-loading effect for these candidate files**, not proof that no project settings dialect exists. **Permission-policy exclusion unmeasured.** | A valid selected home model setting gives `model_source="settings"`. No tool-policy decision is exercised, so neither permission expansion nor preservation is certified. |
| P7 | Headless untrusted session, then real `resume --last`: the TUI offers trust or quit. Root rules retain their saved listing rather than a fresh rules-loader event. Live catalogs are tested separately after granting trust. | **No complete resume recipe qualified.** Granting trust revives the project skill (`project_skills=1`), hook (`runnable=2` with the selected user hook), and MCP source (`trusted_project_source_count=1`); the project marker hook executes. Root snapshot behavior cannot prove exclusion of those live surfaces or children. | Selected home sources remain in the resumed root/catalog observations. Untrusted exec explicitly reports delegation unavailable; a native child tool invocation is **unmeasured**, because echo emits no tool calls. |
| P8 | A project `AGENTS.md` created between two trusted TUI echo turns is absent from the repeated `/rules` listing; a fresh trusted process loads the same file. | **No late-file exclusion recipe qualified.** This observes the session's rule listing without a lazy read tool. A fresh untrusted interactive entry offers only trust or quit. | Home root appears in both listings and fresh startup. Late skills/hooks/plugins/MCP and reload-triggering tools remain unmeasured. |

P7 retains `user: $CONFIG_DIR/AGENTS.md`, `project: none` in `/rules` while
live catalogs admit the project sources after trust. These are separate observed
behaviors, not complete isolation. Echo is used only to flush logs and establish liveness; no assertion
relies on its textual opinion about loaded instructions.

## Controls, validation and disposition

The narrowing variant restores workspace trust for P1's excluded root case.
The row must fail with `project_sources: [1]; want 0`. Separate controls remove
the network and synthetic Keychains denies and require their specific failures.
Wrong-binary and invalid-run-root controls require refusal before vendor startup
or fixture writes. The observation checker itself admits clean zero and one
records and rejects missing, failed, duplicate and contradictory observations.

The [task outcome](../../.research/project-exclusion-muse/evidence/validation.md)
records exact commands and exit codes, including failed exploratory assumptions,
corrected replays, build/vet, default skip checks and the exact-head default suite.
A passing research test can witness an unsupported recipe; it is not a runtime
admission certificate.

Do not enable a CIP-0002 strict Muse adapter from this evidence. Required follow-up
is a native untrusted interactive/resume contract, actual lazy read and native
child probes, permission-policy/tool decisions, and broader lifecycle coverage
for hooks/plugins/MCP. No product API, adapter registry, board status or release
qualification is changed by this research.
