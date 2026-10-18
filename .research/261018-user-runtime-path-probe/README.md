# User-level discovery under runtime configuration-home overrides

Research for BUG-261018-k2tn36, tracked by STORY-261018-aws5ss (`user-level-runtime-path-probe`) and TASK-261018-nd18ex (`probe-user-runtime-discovery-paths`). This is a discovery probe, not an isolation implementation or release qualification.

## Result

**Observed:** on macOS arm64, both Codex 0.162.0 and 0.163.0 include skills from `$HOME/.agents/skills` while `CODEX_HOME` points elsewhere. Both also discover a local plugin marketplace at `$HOME/.agents/plugins/marketplace.json`. Changing only `CODEX_HOME` does not exclude either HOME-based source. Changing HOME moves both lookups to the new HOME in the paired control.

**Observed:** Claude Code 2.1.302 follows `CLAUDE_CONFIG_DIR` for the tested user skills, instructions, settings/hooks and installed-plugin registries. With that variable set and HOME fixed, the old `$HOME/.claude` sentinels are not read by the selected entry points, while the configured-directory positive controls are read and the selected SessionStart hook executes.

Let `H` mean the fixed, synthetic HOME and `C` a separate runtime configuration directory. All paths below describe fixture paths, not an operator's actual home.

| Runtime | Surface | With only configuration directory overridden | Evidence |
| --- | --- | --- | --- |
| Codex 0.162.0 and 0.163.0 | User skills | Reads `C/skills/*/SKILL.md` **and** `H/.agents/skills/*/SKILL.md`; old `H/.codex/skills` excluded | Observed file reads and rendered prompt sentinel names; known valid skill controls |
| Both Codex versions | Global instructions | Reads `C/AGENTS.md`; `C/AGENTS.override.md` takes precedence when present | Observed file reads and rendered instruction content; separate precedence control |
| Both Codex versions | User hooks | Reads `C/config.toml` and `C/hooks.json`; old `H/.codex/hooks.json` excluded | Observed reads with default-directory and overridden-directory controls; execution/trust verdict not tested |
| Both Codex versions | Plugin marketplace discovery | Reads `H/.agents/plugins/marketplace.json` and its referenced local plugin manifest | Observed `plugin list --available` result and file reads, including shifted-HOME positive control; availability is not installation or activation |
| Claude Code 2.1.302 | User skills | Reads `C/skills/*/SKILL.md`; old `H/.claude/skills` and `H/.agents/skills` excluded | Observed reads during init-only plus the read-only plugin listing; valid selected-directory skill control |
| Claude Code 2.1.302 | Global instructions | Reads `C/CLAUDE.md`; old `H/.claude/CLAUDE.md` excluded | Observed file reads; selected global and project instruction controls |
| Claude Code 2.1.302 | User hooks | Reads `C/settings.json`, executes its SessionStart hook; old user settings excluded | Observed reads and exact selected hook marker |
| Claude Code 2.1.302 | Installed plugins | Reads `C/plugins/installed_plugins.json`, `C/plugins/known_marketplaces.json` and the registered local marketplace/plugin files | Observed enabled plugin listing without errors and plugin skill reads; fixture uses a registered directory marketplace |

Codex did not read the planted skill roots at `C/.agents/skills`, `H/.claude/skills`, or the displaced `H/.codex/skills` in the fixed-HOME override cases. Claude did not read the generic `H/.agents/skills` fixture. These are bounded negative observations under the exact commands and fixtures below, not universal filesystem isolation claims.

## Method and controls

The [probe](probe.py) creates a new physical scratch tree under `/private/tmp`. HOME, the runtime configuration directory, XDG directories and temporary directories all point into that tree. The real user directories are neither planted nor modified. The parent session's authentication, session, proxy and integration environment variables are not inherited.

Every native probe runs under macOS `sandbox-exec`: writes are permitted only inside its own scratch lane and network access is denied. Controls verify an allowed write, a rejected write outside the lane with no resulting file, and a rejected loopback socket connection with the expected permission error. No inference prompt, turn submission, login, install, global setup or native home migration is performed. Scratch hooks only append a known marker inside the scratch HOME. Processes are recorded by PID and reaped; timeouts stop only the exact process started by the probe.

The no-model entry points are:

```text
codex --version
codex debug prompt-input
codex plugin list --marketplace <planted-marketplace> --available --json
claude --version
claude --init-only --debug-file <scratch-file>
claude plugin list --json
```

`debug prompt-input` builds the local prompt input without requesting a model response. The installed Claude binary describes `--init-only` as running Setup and SessionStart hooks, then exiting. The network guard also prevents outbound model requests. The probes use no credentials.

File-read observations use access times: after writing each candidate fixture, the harness sets its atime to a known old value, runs the native entry points, then compares atimes. Every lane contains a direct-read instrument control and an unrelated unread negative control. Selected configuration, instruction and skill controls must also be read. Thus an empty listing, skipped positive, setup failure or unsupported atime instrument does not establish exclusion. Reads are aggregated across a lane's selected commands. Atime is evidence of file access, not proof that all of the file's semantics were applied; Codex rendered content and Claude's hook/plugin results provide the additional evidence named in the table.

The harness uses Python's standard library for subprocess deadlines, fixture construction, access-time observation and JSON evidence. It adds no product dependency. Native product behavior is exercised through the release binaries, not reimplemented in Python.

## Executed matrix

Host: macOS 26.5.1, arm64; Python 3.14.7. See [provenance.json](provenance.json) for binary digests, npm package URLs/integrity values and exact source commits. npm platform tarball bytes matched the registry SHA-512 integrity values before extraction. Source tags were peeled through the official GitHub API. The binaries were not rebuilt from those source tags; source-to-binary reproducibility is not claimed. Claude provenance is the pre-existing native binary's version readback and SHA-256, not an independent source build.

| Runtime | Default config | Config override, HOME fixed | Config override, HOME shifted | Instruction override precedence |
| --- | --- | --- | --- | --- |
| Codex 0.162.0 | Passed | Passed; HOME skill and marketplace present | Passed; original HOME sources absent, shifted sources present | Passed |
| Codex 0.163.0 | Passed | Passed; HOME skill and marketplace present | Passed; original HOME sources absent, shifted sources present | Passed |
| Claude Code 2.1.302 | Passed | Passed; configured-directory controls present | Passed; configured-directory controls still present | Not selected; Codex-specific precedence case |

The final normal matrix ran all 11 selected lanes and exited 0. The [machine-readable evidence](evidence.json) retains every candidate's read verdict, command exit codes, prompt sentinels, plugin results and control verdicts. Full built-in prompt bodies and unrelated debug output are omitted from the public evidence. Scratch prefixes are replaced with `$SCRATCH`.

Two executable instrument controls protect the harness's verdicts:

- With the first lane's valid selected skill removed, that lane fails the specific selected-skill read and rendered-prompt assertions. All ten later lanes still execute and pass. The overall process exits 1. Actual native command exit codes remain unchanged in the evidence; an expected-red checker result is not relabeled as a successful runtime command.
- With the first runtime binary deliberately absent, its four lanes report `cannot_start`, then all four lanes of the second, available Codex runtime execute and pass. The overall process exits 1. This is distinct from an executed lane failing a discovery assertion.

Both controls are recorded in [controls.json](controls.json). These checks establish bounded instrument behavior, not exhaustive mutation coverage of the native runtimes.

## Source-only paths and limits

The following conclusions are **source-only**, beyond the observed rows above:

- Codex's HOME marketplace finder tries these layouts, in order, taking the first supported manifest: `.agents/plugins/marketplace.json`, `.agents/plugins/api_marketplace.json`, `.claude-plugin/marketplace.json`, `.cursor-plugin/marketplace.json`. Only the first layout was dynamically qualified here.
- Codex's installed plugin cache/data roots are derived from `codex_home`: `plugins/cache` and `plugins/data`. A HOME marketplace being available does not prove an installed plugin is activated in a thread. This probe does not install plugins or qualify all cache/activation paths.
- Codex hook discovery visits configuration layers and obtains each layer's hook configuration folder before loading `hooks.json` and TOML hooks. Only the selected user files' reads were observed here; hook trust, effective policy and execution remain unqualified.

The examined skill-root, global-instruction, marketplace, hook-discovery and Codex-home resolution source files are byte-identical between the two release tags; [source-comparison.json](source-comparison.json) records their digests. No general equivalence of the entire releases is asserted.

An initial app-server JSON-RPC probe could initialize but configuration reads failed with `Failed to synchronize managed preferences` under the scratch sandbox. Its empty skills/hooks responses are **inconclusive**, not exclusion evidence. The final matrix uses the supported local prompt-rendering and plugin-list CLI entry points, with working positive controls. No managed-preferences guard was disabled or worked around.

Claude's directory documentation says its user directory follows `CLAUDE_CONFIG_DIR`; the runtime evidence above independently covers the planted primary paths in version 2.1.302. Alternate instruction fallback names, user command aliases, additional rule directories, synced content, arbitrary plugin source types, admin/system settings, credentials, shell startup and interactive sessions are not qualified by this probe. It is not a Windows/Linux qualification and does not establish a security boundary against symlinks, explicit extra paths or ambient system sources.

## Reproduce

Obtain the exact native binaries identified in `provenance.json` without installing them globally. Pass their executable paths explicitly. Keep output/logs in a task-scoped `.temp/` directory:

```sh
mkdir -p .temp/BUG-261018-k2tn36
python3 .research/261018-user-runtime-path-probe/probe.py \
  --codex 0.162.0=/absolute/path/to/codex-0.162.0 \
  --codex 0.163.0=/absolute/path/to/codex-0.163.0 \
  --claude /absolute/path/to/claude-2.1.302 \
  --output .temp/BUG-261018-k2tn36/results.json \
  > .temp/BUG-261018-k2tn36/run.log 2>&1
```

Add `--control-remove-first-skill` to repeat the expected-red first-lane control. For the unavailable-binary control, supply a nonexistent first Codex path and a real second Codex path. The printed scratch location contains fixture configurations and owned PID records. Preserve it for inspection; no automatic cleanup removes evidence. Do not run the probe on a filesystem that does not update atimes and treat its failed instrument as passing evidence.

## References and decision

Exact primary source references for Codex 0.163.0, commit `83086188ae0614bd882818ff1aa23f3194962012`:

- [User skill roots](https://github.com/openai/codex/blob/83086188ae0614bd882818ff1aa23f3194962012/codex-rs/ext/skills/src/host_roots.rs#L84)
- [Global instructions provider](https://github.com/openai/codex/blob/83086188ae0614bd882818ff1aa23f3194962012/codex-rs/codex-home/src/instructions/mod.rs#L30)
- [Marketplace HOME resolution and layouts](https://github.com/openai/codex/blob/83086188ae0614bd882818ff1aa23f3194962012/codex-rs/core-plugins/src/marketplace.rs#L20)
- [Marketplace HOME scan](https://github.com/openai/codex/blob/83086188ae0614bd882818ff1aa23f3194962012/codex-rs/core-plugins/src/marketplace.rs#L472)
- [Installed plugin store](https://github.com/openai/codex/blob/83086188ae0614bd882818ff1aa23f3194962012/codex-rs/core-plugins/src/store.rs#L86)
- [Hook discovery](https://github.com/openai/codex/blob/83086188ae0614bd882818ff1aa23f3194962012/codex-rs/hooks/src/engine/discovery.rs#L128)
- [Local prompt renderer](https://github.com/openai/codex/blob/83086188ae0614bd882818ff1aa23f3194962012/codex-rs/cli/src/main.rs#L1995)
- [Managed-preferences synchronization](https://github.com/openai/codex/blob/83086188ae0614bd882818ff1aa23f3194962012/codex-rs/config/src/loader/macos.rs#L133)

Other primary documentation: [OpenAI's prompt assembly description](https://openai.com/index/unrolling-the-codex-agent-loop/), [Claude's user directory contract](https://code.claude.com/docs/en/claude-directory), [Claude configuration inspection](https://code.claude.com/docs/en/debug-your-config). These are current documentation, not substitutes for pinned runtime evidence.

The implementation decision is now specific: an environment that promises to exclude ambient Codex user skills and marketplaces cannot rely solely on `CODEX_HOME` for these two versions. The observed shifted-HOME control removes the original HOME sources, but this report does not select or implement Curator's isolation design. That design, its broader environment/credential consequences and landing remain with the owning lane. No product code or real user configuration changes are included.
