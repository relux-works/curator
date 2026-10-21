# Codex 0.162.1 project-source exclusion qualification

TASK-261021-boww1n. **A complete strict-context adapter is not qualified.**
The exact build can suppress startup project instructions while preserving
managed-home context, and project distrust gates the measured config-driven
surfaces. Repository skills survive both controls. Disabling known skill
paths fails for a new path. Resume, lazy entry and child-agent exclusion lack
complete discriminating evidence. No Curator adapter or product code changes.

## Identity and method

| Item | Pin or boundary |
| --- | --- |
| Vendor | `codex-cli 0.162.1`, standalone executable, macOS arm64 |
| Executable SHA-256 | `74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0` |
| Opt-in | `CODEX_EXCL_BINARY`; missing gate skips before fixture creation |
| Fixtures | Fresh absolute `CODEX_EXCL_RUN_ROOT` under task-local `.temp/` |
| Runtime | Go 1.25.5; Apple Seatbelt; stdio app-server and synthetic stdio MCP |
| Credentials and inference | No credentials, login/auth commands, model turns or remote inference |
| Process environment | `env -i`; disposable HOME, CODEX_HOME, four XDG roots and TMPDIR |
| Isolation | All network denied, including loopback; real native directories, system Codex policy and Keychains-shaped files denied |

The [harness README](../../.research/project-exclusion-codex/README.md)
contains commands, tool entry points, outputs and the full replay boundary.
The digest is checked before any fixture exists; vendor version checks and
all vendor entries run in Seatbelt. Git metadata is separate from checkouts.
The approved copies are synthetic files inside the managed home; they do not
assert production admission or publication provenance.

The normative motivation is [CIP-0002](https://github.com/relux-works/curator-spec/blob/main/cips/CIP-0002-project-context-in-managed-launches.md),
read for this run: strict entry must exclude unadmitted automatic discovery
at startup, reload, nested entry, resume and child entry. A native home change
alone is insufficient. This record qualifies observations at a pin, not that
draft's implementation or acceptance.

## Surface verdicts

A verdict names the measured control and entry point. It never extends an
absence observation beyond its matching positive baseline.

| Row | Runtime observation and approved-copy control | Verdict and exact control |
| --- | --- | --- |
| P1 root and ancestor instructions | `debug prompt-input` loads managed `AGENTS.md` and repository `AGENTS.md`; a nested cwd loads the repository ancestor too. An instruction outside the Git root and root `CLAUDE.md` are absent by default. `AGENTS.override.md` replaces root `AGENTS.md`. | **Excludable for measured startup injection:** `-c project_doc_max_bytes=0`. The approved managed-home marker remains while project and override markers disappear. This is an injection budget, not a file-access sandbox. |
| P2 nested and lazy instructions | Startup from `nested/` includes both root and nested markers; the zero cap removes both and keeps the approved home marker. | **Nested startup excludable with the same cap; lazy entry unmeasured.** A new cwd startup is not evidence of a tool entering a directory in an existing turn. |
| P3 project skills | A repository `.agents/skills` marker loads with `project_doc_max_bytes=0` and explicit project distrust. The approved home skill also loads. Exact-path `skills.config` entries suppress the known project paths, keeping the approved skill; a newly created path loads on the next prompt probe. | **No complete exclusion with the tested controls.** `-c 'skills.config=[{path="<project>/.agents/skills/repo-agents/SKILL.md",enabled=false},{path="<project>/.codex/skills/repo-codex/SKILL.md",enabled=false}]'` covers those exact paths only. The `.codex/skills` path is planted but is not claimed as a supported loader root without a positive marker. |
| P4 hooks and plugins | `hooks/list` includes both home and project SessionStart hooks for trusted projects, only the approved home hook under distrust. Effective config merging replaces the same-event array, whereas hook enumeration retains both sources. Installed local plugin skill bytes load through `skills/list`; a project enable override affects the installed project plugin only when trusted. Approved home plugin skill bytes remain available in either case. | **Measured hook discovery and installed-plugin enablement excludable with explicit project distrust:** protected home `[projects."<absolute-project>"] trust_level="untrusted"`. The home explicitly disables `project@fixture`, and the trusted project attempts to re-enable it. Hook execution, all plugin marketplaces and other plugin entry points remain unmeasured. Hook metadata still requires its separate native hash-trust decision. |
| P5 MCP, empty and non-empty | Four real-runtime cases cross project trust with empty/non-empty MCP config. The approved home stdio tool initializes and is called in all four. The project process starts and its tool is called only in trusted/non-empty. In other cases its receipt is absent and the server is absent from runtime status. Empty project config retains the approved home tool. | **Excludable for measured config sources with explicit project distrust.** No model, network MCP, OAuth, remote server or secret transport is involved. |
| P6 permission and settings files | `config/read` shows trusted project overrides of approval, sandbox and reasoning settings. `thread/start` applies `on-request`/`workspaceWrite` when trusted and approved home `never`/`readOnly` under distrust. The home configuration still loads. | **Measured `.codex/config.toml` values excludable with explicit project distrust.** Other permission inputs, execution-policy `.rules`, malformed-config handling and every native setting are unmeasured. This is not a blanket policy-bypass proof. |
| P7 resume and child agents | Real `thread/start`, history injection, unsubscribe and `thread/resume` retain the approved managed-home instruction source with cap 32768 and 0. `instructionSources` omits the project source even in the uncapped baseline. | **Project exclusion unmeasured.** That source list cannot distinguish the required project-positive case. Synthetic history is not previously tainted model history. Child agents are unmeasured under this run's no-other-agents constraint; no spawn or model turn is used. |
| P8 late-created files | Within one app-server, `skills/list(forceReload=true)` sees a newly created project skill despite distrust, while the approved home skill remains. `config/read(cwd=...)` reloads changed project config: distrust excludes the late MCP entry and preserves the approved home entry. Trusted config reload also observes removal of the project MCP entry. | **No complete late-source exclusion with the tested recipe.** Explicit skill reload is a failure witness. Automatic in-turn reload, late instructions and all lifecycle entry points remain unmeasured. |

Exact control construction and assertions live in the named tests, not a
wrapper that deletes or rewrites repository files. The project, config and
skill bytes remain in place during exclusion probes.

## Evidence integrity and narrowing controls

[Observations](../../.research/project-exclusion-codex/observations.json)
retain bounded extracts with command exits, markers, runtime policies, MCP
startup/call results, hook sources and the resume instrument limitation.
Full generated prompt dumps, fixture databases and raw synthetic histories
remain ignored in task-local `.temp/`. The external task outcome records the
handed-over head, exact validation commands and exits.

| Control plant | Intended failing assertion |
| --- | --- |
| Drop only the project document cap | Excluded `PROJECT_ROOT` returns in a valid JSON prompt dump |
| Drop only the known-path skill disables | Excluded `PROJECT_AGENTS_SKILL` returns while the approved skill remains |
| Replace explicit project distrust with trust | Excluded `PROJECT_HOOK` returns in effective configuration |
| Allow only loopback through the sandbox | A reachable synthetic endpoint connects with curl exit 0 instead of the required connect-denial exit 7 |

The mutant runner requires a nonzero child exit and the matching diagnostic;
a skip, setup failure, unrelated error or timeout is not accepted. Intake
refusals also require the specific digest/run-root verdict before fixture
creation. The offline control pairs readable ordinary synthetic bytes with
an Operation-not-permitted Keychains-shaped read and unchanged source bytes.

## Adapter consequence and remaining evidence

Do not promote this record into a strict-context capability. Combining
CODEX_HOME, the document cap and project distrust still admits repository
skills. An inventory-based skill disable list does not close late creation.
The measured hook/plugin/MCP/settings controls remain useful individual facts,
subject to the lifecycle and source limits above.

A follow-up qualification needs a selective repository-skill exclusion control
that retains approved skills, discriminating lazy/resume/child probes,
previously tainted history cases, late instructions, execution-policy rules,
plugin discovery beyond installed local skills and TUI parity. Unsupported
or unmeasured surfaces must remain refused by any proposed strict adapter.

Supporting official OpenAI documentation, consulted before vendor probing:
[configuration reference](https://learn.chatgpt.com/docs/config-file/config-reference)
and [AGENTS.md discovery](https://learn.chatgpt.com/docs/agent-configuration/agents-md).
Current docs provide candidate controls; the exact binary observations above
are the qualification evidence. Tagged source was used to locate API schemas
and construct fixtures, including [config loading](https://github.com/openai/codex/blob/rust-v0.162.1/codex-rs/config/src/loader/mod.rs),
[CLI prompt input](https://github.com/openai/codex/blob/rust-v0.162.1/codex-rs/cli/src/main.rs)
and [app-server protocol](https://github.com/openai/codex/tree/rust-v0.162.1/codex-rs/app-server-protocol).
