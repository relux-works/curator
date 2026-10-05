# CIP-0002 evidence: project context in managed launches

Companion to [CIP-0002: Project context in managed launches](261004_CIP-0002-project-context-in-managed-launches.md). Research for TASK-261004-2asduq — launch-command-environment-fragment-design, under STORY-261004-1ffwh8 — design-launch-project-views. Collected 2026-10-04. Design remains pending; no product implementation or scheduling is authorized by this evidence.

## Scope, baseline and research bounds

Decision to unblock: select the project × profile view, admission and command-transport architecture while supporting all five requested environment inventories. Budget: one researcher session, approximately 60 minutes including packaging; one CIP plus this companion, no archives, target below 100 KiB combined. No serial research dependency is introduced. Exit criterion: an option comparison, recommended design, decision-ready questions and consuming implementation leaves. First proposed production slice: approved context and a protected skill command through manager resolve and direct curator-run with the pinned Pi adapter, after design acceptance. `grammar_frozen: not applicable` for this research: the draft schema is a decision input, and no implementation consumer is yet authorized. Leaf 1 freezes the accepted subset before a parser is implemented.

| Repository | Read baseline | Relationship verified |
|---|---|---|
| Curator | `ca1b776fb580ec0cee0173bf150daf063023aeaa` | Worktree HEAD and local main identical; initial worktree clean. |
| curator-spec | `43bf0a2506d5c354a73bbc3ea4623d4653db10c7` | HEAD, main and peeled tag `v1.0.0-rc.14` identical. |
| curator-agent-launcher | `d0920353556dd0a3cea2864c616c230915977bc7` | HEAD and main identical; source checkout clean. |

These are observed local source pins, not an assertion that a remote branch remained unchanged throughout the run. No fetch, checkout, commit, merge or product edit was needed. The attached author's design material was considered as input; its original document was not copied into these public outcomes. Only this researcher's synthesis and public-source evidence are included. Personal paths and actual account details are excluded.

Source inspection used bounded `rg`, `nl`/`sed` and Git reads. Measurements used researcher-created files under a fresh scratch HOME/CWD, with a minimal child environment supplied directly to `subprocess.run` and explicit timeouts. `HOME`, tool homes and XDG roots existed only in each child environment dictionary; the session's own environment was not reassigned. No operator settings/auth file or Keychain item was opened, and no login, logout, inference request or credential transfer occurred. Pi's probe sent an RPC state request, not an inference prompt. Synthetic extension execution was intentional and confined to that scratch fixture.

## Pinned source evidence

Line ranges below refer to the exact commits above. Links use public repository paths, not machine paths.

### E1 — Home identity and launch directory

[Curator managed.go](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L63), lines 63–85: `ManagedParent` joins environments root, profile and environment; OpenCode adds a child directory. `ResolveRequest.LaunchDir` is a separate field. This establishes the existing profile × environment key, not an assertion that LaunchDir has no other use.

### E2 — Fragment revisions and influence boundary

[Curator envfragment.go](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envfragment/envfragment.go#L23), lines 23–70, 106–138 and 347–414: v2 default, Muse v3, closed emitted fields and environment/path checks. [rc.14 environments](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L3155), §10.2 lines 3155–3251: three fragment versions; a reserved single `path_prepend`; Muse home/channel constraints. §10.3 governs profile influence. No command-environment or project-root member is emitted by the inspected construction site.

### E3 — Machine-current command limitation

[rc.14 environments §9.4](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L2583), lines 2583–2606: singleton command shims represent machine-current; managed-home command availability is explicitly limited; `curator-*` exports are reserved; hybrid scope remains separate. [Manager §3](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/profiles/manager.md#L647), lines 657–710: portable project shims, PATH augmentation, runtime/artifact binding and a separate enforced-script lane. Inference: a new appended dispatcher alone neither displaces old earlier shims nor fixes an old command wrapper's own PATH augmentation.

### E4 — Conditional strict MCP

[Curator buildFragment](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L2290), lines 2290–2312: the MCP member depends on a recorded MCP surface and supported channel. [Registry](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envregistry/envregistry.go#L243), line 251: Claude's descriptor carries strict MCP. [Launcher composition](https://github.com/relux-works/curator-agent-launcher/blob/d0920353556dd0a3cea2864c616c230915977bc7/internal/composition/composition.go#L93), lines 93–122: flags are added only when `frag.MCP != nil`. [rc.14 environments §10.2](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L3226), lines 3226–3233, specifies presence for a nonempty set.

Bound: this is source-path evidence, not a new Claude empty-set launch test. It disproves the unconditional interpretation of “the launcher passes strict MCP.”

### E5 — Real CWD and launcher support

[Launcher main.go](https://github.com/relux-works/curator-agent-launcher/blob/d0920353556dd0a3cea2864c616c230915977bc7/cmd/curator-run/main.go#L249), lines 249–304: plan build receives `WorkDir: wd`; the real composition and execution preparation are called there. [Mapping](https://github.com/relux-works/curator-agent-launcher/blob/d0920353556dd0a3cea2864c616c230915977bc7/internal/mapping/mapping.go#L18), lines 18–32: only Claude, Codex and Pi map to providers. [Fragment parser](https://github.com/relux-works/curator-agent-launcher/blob/d0920353556dd0a3cea2864c616c230915977bc7/internal/fragment/fragment.go#L324), lines 324–333: outer v1/v2 admission. Muse's manager v3 therefore does not imply launcher support.

### E6 — Owned literals, destination PATH and identity

[Launcher composition](https://github.com/relux-works/curator-agent-launcher/blob/d0920353556dd0a3cea2864c616c230915977bc7/internal/composition/composition.go#L33), lines 33–46 and 74–139: full direct environment is separate from serialized owned literals/name lookups; no PATH-transform field exists in this structure. [Decision 0013](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0013-execution-ownership-and-launch-plans.md#L87), §§1, 3.2 and 6.4: one composer, tracked request and fragment digest. [Curator canonical fragment](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envfragment/envfragment.go#L141), lines 141–154: canonical bytes excluding trailing LF define the fragment digest. Proposed destination transform/lease identity is new design, not existing capability.

### E7 — Login cost and unresolved credential mechanism

[Decision 0017](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/decisions/0017-environment-credential-modes.md#L113), lines 113–136 and question 1 at line 170: credential sharing is distinct from sandboxing; Claude/macOS shared mode remains unsupported pending a properly authorized experiment. The decision records a conflicting file-vs-Keychain observation; no account-specific details are repeated here. [Curator first-resolve notice](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envprofile/managed.go#L2338), lines 2338–2352, tells Claude users to log in within a new home. [rc.14 environment version table](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L1868), §7.9, retains earlier version-specific facts and an erratum procedure.

TASK-261004-34brhn — research-claude-login-transfer-modes was queried twice for status/outcome names during this run. Both reads showed `analysis` and only a system spawn log, no research outcome. No worker transcript was read and no credential experiment was rerun. Login reuse is therefore an open design cost, not a proven migration path.

### E8 — Native residuals and evidence drift

[rc.14 environments](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L1260), §§7.1–7.3, 7.8–7.9: OpenCode global skill/state residuals; Muse's unverified context/MCP/skill discovery; tool pins. [Curator registry](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/internal/envregistry/envregistry.go#L228), lines 228–339, mirrors those adapter channels. Curator `internal/envprofile/managed.go:2348–2350` warns about ambient Muse context but says Pi has no trust wall. P5 contradicts that Pi notice on its own recorded 0.84.2 release; the draft recommends an erratum, not a code change in this task.

### E9 — Old home-depth assumption

[Launcher fragment.go](https://github.com/relux-works/curator-agent-launcher/blob/d0920353556dd0a3cea2864c616c230915977bc7/internal/fragment/fragment.go#L493), lines 493–506: reserved prepend validation derives environments root from the old two-level home layout. A deeper project-view hierarchy needs a version-specific boundary rule. The proposed external project-root field cannot simply be added to the old path map.

### E10 — Provider-discovery refusal boundary

[Curator umbrella.go](https://github.com/relux-works/curator/blob/ca1b776fb580ec0cee0173bf150daf063023aeaa/cmd/curator/umbrella.go#L75), lines 75–125 and 278–303: provider trust roots are distinct from published/managed refused directories. [rc.14 environments §11](https://github.com/relux-works/curator-spec/blob/43bf0a2506d5c354a73bbc3ea4623d4653db10c7/protocol/environments.md#L3377) defines umbrella discovery. Adding dispatcher locations and aliases to that refusal set is a proposed defense against a skill supplying `curator-run`; no new vector was implemented here.

## Public native documentation

Accessed 2026-10-04. These sources establish documented behavior only; current documentation must not be retroactively treated as a measurement of Curator's older pins. Summaries are deliberately bounded; full vendor pages and another operator's material are not attached.

| ID | Primary source | Claim used and bound |
|---|---|---|
| D1 | [Claude memory](https://code.claude.com/docs/en/memory) | Instruction locations/imports, scoped rules, auto-memory and AGENTS fallback since 2.1.277. Excluding startup instructions does not imply exclusion of lazy nested reads. |
| D2 | [Claude settings](https://code.claude.com/docs/en/settings), [environment variables](https://code.claude.com/docs/en/env-vars) | User/project/local settings, permissions and environment assignment. Treat `env.PATH` as replacement, not an append channel; this is documentation/task-input confidence, not a PATH-setting runtime probe. |
| D3 | [Claude CLI reference](https://code.claude.com/docs/en/cli-reference) | Strict MCP scope and settings source selection. Installed help additionally confirms `--bare` authentication restrictions and broad `--safe-mode`; it does not prove a selective re-admission recipe. |
| D4 | [Codex config basics](https://learn.chatgpt.com/docs/config-file/config-basic), [AGENTS.md](https://learn.chatgpt.com/docs/agent-configuration/agents-md) | Trusted project layers and project instruction ordering are separate from home configuration. P2/P3 test the distinction. |
| D5 | [Codex advanced config](https://learn.chatgpt.com/docs/config-file/config-advanced), [reference](https://learn.chatgpt.com/docs/config-file/config-reference) | Project config/trust, hooks and literal `shell_environment_policy.set`, with exclusions/set/include ordering. No append operator is documented there. |
| D6 | [Codex rules](https://learn.chatgpt.com/docs/agent-configuration/rules) | Execution policy rules are distinct from prose guidance; their discovery and semantics need an adapter-specific contract. |
| D7 | [Codex skills](https://developers.openai.com/codex/skills/), [plugin packaging](https://developers.openai.com/plugins/build/plugins) | Repository `.agents/skills` discovery and project plugin configuration broaden the control-file inventory beyond `.codex/config.toml` alone. |
| D8 | [OpenCode config](https://opencode.ai/docs/config/) | Global/custom/project/resource merge ordering; project config follows `OPENCODE_CONFIG`. No installed OpenCode runtime was available. |
| D9 | [OpenCode rules](https://opencode.ai/docs/rules/) | Legacy AGENTS/CLAUDE behavior and instruction paths/globs/URLs, including Cursor-style inputs. |
| D10 | [OpenCode V2 config](https://opencode.ai/v2/docs/config) | V2 has its own project-config hierarchy; do not combine it indiscriminately with legacy behavior. |
| D11 | [OpenCode V2 instructions](https://opencode.ai/v2/docs/instructions) | AGENTS-only behavior, dynamic discovery, project-config disable control and currently inactive explicit `instructions` loading. Every V2 fact here remains docs-confidence. |
| D12 | Pi installed `@earendil-works/pi-coding-agent` 0.84.2 docs; [upstream usage](https://github.com/earendil-works/pi/blob/main/packages/coding-agent/docs/usage.md) | Installed `docs/usage.md:100–131,222–248,304`, `docs/settings.md:7–22,334`, `docs/skills.md:27–41`: context/trust split, native settings/resources, flags and no built-in MCP. Local hashes below pin the actual docs used; upstream main is a navigation link, not the source pin. |
| D13 | [Muse configuration/context](https://dev.meta.ai/docs/muse-code/configuration) | Instruction precedence, trust and project memory. Does not establish project settings.json/.mcp.json discovery. |
| D14 | [Muse extending](https://dev.meta.ai/docs/muse-code/extending) | Project/foreign skills, project hooks and settings-based MCP. Its current skill-root description differs from the rc.14 recorded data path; do not silently rewrite that registry fact. |
| D15 | [Muse permissions](https://dev.meta.ai/docs/muse-code/permissions) | Whole-workspace trust enables project resources; permissions and sandbox controls are distinct. Sandbox effectiveness was not reproduced in this research. |

One Claude settings-reference fetch exceeded the web reader's size limit; its Markdown endpoint returned HTTP 403. A guessed Codex skills URL returned 404; the official `/codex/skills/` page was then retrieved successfully. A guessed shell-policy Markdown endpoint failed; the official advanced-configuration page and Markdown body were retrieved. Those failed fetches establish nothing and are not validation passes. No inaccessible page was treated as reviewed.

Installed Pi documentation digests (SHA-256):

```text
docs/usage.md     a6a76e733c50ea8a08701456858d3937e1d60caa8709a7f15125e6f5ba6cabca
docs/settings.md  f36d3a918d87d18e13d22ca98f4e429428b5f2bc06316ff7d2e7adace59a973b
docs/skills.md    de38956dcdb3f060b62a891c23b5c0facc568d26039cd0addd6d5adfbed2762f
```

## Measured probes

All commands below ran as separate native processes with captured stdout/stderr; none was piped through `tee`. Every listed native process exited normally with the real exit code shown. Harness exit 0 is not substituted for a child status. No long-running check was abandoned at turn end.

### P1 — Version and help inventory

Each tool received `--version` and `--help` in the scratch workspace. Timeout: 10 seconds per process. Minimal environment: fixed executable search PATH, scratch HOME, tool config home, four scratch XDG roots, `CI=1`, `NO_COLOR=1`, `TERM=dumb`. No inherited credential variables. The executable was resolved before launching the scratch child.

| Tool | Version output | `--version` exit | `--help` exit | Evidence scope |
|---|---|---:|---:|---|
| Claude | `2.1.287 (Claude Code)` | 0 | 0 | Flag descriptions only; newer than registry 2.1.261. |
| Codex | `codex-cli 0.159.0` | 0 | 0 | Flag descriptions only; newer than registry 0.153.2. |
| Pi | `0.84.2` | 0 | 0 | Same release as registry; project trust and resource flags present. |
| Muse | `Muse Code 1.4.2 (1.4.2-R4684.1)` | 0 | 0 | Newer than registry 1.4.1-R4503.1. |
| OpenCode | Not on this run's executable search path | Not run | Not run | No installed/runtime claim. |

Claude help lines 41–58 describe `--bare` skipping automatic context and Keychain/OAuth reads while requiring a supported alternate auth route; lines 212–226 describe `--safe-mode` disabling customizations and the settings-source selector. These flag descriptions do not prove they satisfy this CIP. Pi help lines 42–56 expose resource-discovery switches and trust overrides. Muse help lines 79–82 says workspace trust loads skills/rules without persisting the decision.

Two additional Codex processes, `codex -c 'cli_auth_credentials_store="file"' debug --help` and the corresponding `debug prompt-input --help`, both exited **0**. They established an offline prompt-input inspection command before P2/P3 used it.

### P2 — Codex home context versus project AGENTS

Fixtures: scratch home `AGENTS.md` contains `CIP0002_PROFILE_SENTINEL`; scratch project `AGENTS.md` contains `CIP0002_PROJECT_SENTINEL`. The scratch project contains a `.git` marker directory. Scratch config selects file credential storage and disables startup update checking. No auth file exists. The commands only render prompt input; they do not send a model request.

```text
codex debug prompt-input CIP0002_USER_SENTINEL
codex -c project_doc_max_bytes=0 debug prompt-input CIP0002_USER_SENTINEL
```

| Command | Exit | Home sentinel | Project sentinel | stdout characters | stderr |
|---|---:|---|---|---:|---|
| Default | 0 | present | present | 12896 | empty |
| Zero project document cap | 0 | present | absent | 12736 | empty |

Measured result: the cap suppresses this startup project-document chain without suppressing home instructions. Bound: it does not prove anything about skills, hooks, MCP, dynamic native reads or source-file prompt injection.

### P3 — Codex trusted project configuration is a separate channel

A second scratch project was initialized with `git init --quiet --template=` using no user/system Git config; exit **0**. Its `.codex/config.toml` contained a researcher-authored `developer_instructions` sentinel. The scratch home config retained `project_doc_max_bytes=0` and used a project-specific `trust_level`, varied between `trusted` and `untrusted`. The same `codex debug prompt-input CIP0002_USER_SENTINEL` command ran once per trust state, with a 15-second timeout.

| Trust state | Exit | `CIP0002_PROJECT_CONFIG_SENTINEL` in model-visible input | stderr |
|---|---:|---|---|
| trusted | 0 | present | empty |
| untrusted | 0 | absent | empty |

This is a native-entry counterexample to treating `project_doc_max_bytes=0` as suppression of all project control. It measures one allowed project config key; it does not establish that every key is accepted or that an MCP server started.

### P4 — MCP list diagnostic: deliberately limited evidence

Before P3's Git initialization, the second fixture used a `.git` marker directory, a generated `curator-mcp.config.toml` layer containing `manager_probe`, and a project config with `project_probe` plus a differing `manager_probe`. No real endpoint or credentials were configured. `codex -p curator-mcp mcp list --json` ran with trusted and untrusted scratch-home records; both exited **0**, printed only `manager_probe` with command `/usr/bin/false`, and had empty stderr.

That observation **does not** demonstrate project MCP merging or suppression during an actual agent launch. It also does not establish why this diagnostic did not include the project config. The reason was not investigated beyond the separate P3 production prompt-input check. The CIP's broader MCP inventory is documentation confidence; no MCP connection or project-MCP runtime proof is claimed.

### P5 — Pi real RPC startup: project extension trust

Fixture: `.pi/extensions/cip-marker.ts` under a new scratch project contains exactly:

```typescript
export default function() { process.stderr.write("CIP0002_EXTENSION_LOADED\n"); }
```

Scratch Pi settings disable install telemetry and update checking. No credential files or project packages are supplied. The native command template was:

```text
pi <trust-flag> --no-context-files --no-skills --no-prompt-templates --no-themes --no-session --mode rpc
```

The child receives one stdin line, `{"id":"probe","type":"get_state"}`, followed by EOF. `--no-extensions` is intentionally absent: the tested variable is project trust. Each process had a 15-second bound and exited without a model request.

| Trust flag | Exit | RPC get_state response | Extension sentinel on stderr |
|---|---:|---|---|
| `--no-approve` | 0 | present | absent; stderr empty |
| `--approve` | 0 | present | present |

This proves the trust-sensitive project extension path is reachable and suppressed by the negative case on installed Pi 0.84.2. It does not prove all project discovery is suppressed, permissions are sandboxed, SYSTEM/APPEND_SYSTEM files are disabled or every explicit-resource combination works. Installed docs explicitly distinguish context loading from trust. The current manager's “No trust wall” notice therefore needs review.

### P6 — PATH append preserves earlier command selection

Two scratch directories contain an executable `cip-probe`, each a researcher-authored `/bin/sh` script printing a constant. No repository script was sourced. Each probe invoked `/bin/sh -c cip-probe` with a child environment containing only scratch HOME and the specified PATH; timeout 5 seconds.

| PATH | Exit | stdout |
|---|---:|---|
| `<base>:<dispatcher>` | 0 | `BASE` |
| `<dispatcher>` | 0 | `DISPATCH` |

This demonstrates ordinary POSIX search order, not a Curator dispatcher implementation or an integrity guarantee. It is why a stale machine-current shim earlier on PATH must be migrated or refused, rather than assuming append selects the new view. Windows behavior was not tested.

## Coverage and unknowns

- Inventory: **5/5** requested environments addressed. Native help/version processes: **4/5** environments available, **8/8** executed processes exited 0.
- Behavioral native probes: **2/5** environments (Codex and Pi). P2/P3 cover six explicitly reported sentinel-presence outcomes, **6/6** matching the reported observations. P5 covers the two extension-trust states, **2/2**. These are evidence bounds, not a release-conformance score.
- No full product suite, platform qualification, live account login, MCP connection, tracked-session launch, dispatcher implementation, lease/GC race or hostile-process confinement was tested. No such checklist gate is represented as green.
- OpenCode is documentation-only and version-split. Muse context/MCP target admission remains unknown in the pinned spec. Claude's complete selective suppression/re-admission recipe is unproven. Codex's project skill/plugin exclusion and live MCP behavior remain unproven by these probes.
- Some exploratory path lookups returned 1/2 for nonexistent guessed filenames/directories; searches were corrected to actual `umbrella.go`, launcher composition/mapping and installed Pi docs. An initial board query without semicolon separators returned 1 and was corrected. Scoped schema lookup for `update_checklist_item` returned 1 because that mutation does not exist. These were exploratory failures, not passing gates.

## Artifact verification and lifecycle

The document validation process exited **0**, with **13/13** checks: template sections, options/recommendation, draft/pending status, five environments, evidence IDs, artifact budget, relative links, personal-path scan, whitespace, JSON example parsing, tracked-file scope, unchanged HEAD and unchanged LOGBOOK.md. Its three Git subprocess checks (`git diff --quiet HEAD --`, `git rev-parse HEAD`, and `git diff --quiet HEAD -- LOGBOOK.md`) each exited **0**. The first pass measured 69,580 bytes across the two files; a subsequent record-only update adds this validation result.

The two standalone commands below each returned **exit 1**, with no output:

```text
git diff --no-index --check /dev/null .research/261004_CIP-0002-project-context-in-managed-launches.md
git diff --no-index --check /dev/null .research/261004_CIP-0002-project-context-in-managed-launches_evidence.md
```

These are **not green gates**: no-index reports differences for a nonempty new file against `/dev/null`. No whitespace diagnostic was emitted; the explicit whitespace validation above supplies the passing whitespace check. The source/evidence inspections and native probes do not imply a passing product test suite.

Research changes are restricted to the two CIP Markdown files; no product source/config/test or LOGBOOK.md edit is part of the deliverable. The generic logbook checklist conflicts with the task's explicit prohibition, so its evidence belongs in this companion and task notes; the board checklist is reconciled to that task-specific requirement before handoff.

Packaging encountered a command-runner stall: a second source-locator audit, board calls and even shell liveness probes remained pending across bounded waits. They were not counted as passing while pending. Execution subsequently recovered: the second audit exited **0**, with **14/14** checks and **23/23** pinned source locators verified; every source-reading Git subprocess exited **0**. The checklist replacement also returned exit **0**. The initial 13/13 audit and native probe results above had already returned their recorded exit codes. Outcome attachment and role handoff must be established by actual board command results, not inferred from these files being present.

Both files must be attached as new outcome resources named with TASK-261004-2asduq — launch-command-environment-fragment-design before the researcher handoff. The last board command is the requested `task-board handoff TASK-261004-2asduq --role researcher`.

Republished for review
