# Claude Code 2.1.317 permission grammar measurements

Observed 2026-11-07 on macOS 27.2 (26B123), Darwin 27.2.0, arm64.
Task: TASK-261107-oe1ki9 — qualify-claude-permission-grammar-2-1-317-and-319.
This is a sanitized real-binary help/parser record following the
[Codex precedent](codex-0.164.0.md). It does not admit a runtime or Linux tuple.
The [gate analysis and producer handoff](claude-permission-grammar-handoff.md)
contains the current V3 gate requirements, historical comparison, isolation
controls and gaps. No admission row is proposed.

## Exact binary

| Property | Observation |
| --- | --- |
| `probe --version` | `2.1.317 (Claude Code)`; stdout; exit 0; stderr empty |
| SHA-256 before and after all probes | `09f153b46b5df9e8b3454716d682a91901de020cbca630f61dec1a015b32b5e9` |
| Executable format | Mach-O 64-bit arm64 (header CPU type `0x0100000c`) |
| Provenance | Direct executable from the integrity-verified published darwin-arm64 optional package; no installer or package script executed. |

Version and hash identify the executable actually tested, not a reported
installation on another machine. Hashing used Python `hashlib.sha256` over
the executable bytes; no installed file was modified.

## Published npm intake, without installation

The wrapper package declares the exact-version `darwin-arm64` optional
dependency. Its `bin/claude.exe` is an install-missing stub; its `postinstall`
would run `node install.cjs`. Neither was executed. The published platform
package contains a ready executable; direct extraction and execution worked,
so no install step was required. This qualifies that executable, not the npm
wrapper's installation or dispatch behavior.

Python `urllib.request.urlopen` fetched each version metadata URL and its
`dist.tarball` from `registry.npmjs.org`; both fetch stages completed without
error (intake process exit 0). Before `tarfile.extractall(filter="data")`, the
downloaded bytes passed both the published SHA-512 SRI and SHA-1 `shasum`.
No npm command, lifecycle script, setup script or provider request ran.

Package [`@anthropic-ai/claude-code@2.1.317`](https://registry.npmjs.org/@anthropic-ai/claude-code/2.1.317):

```text
dist.integrity: sha512-k467+KT/LcqJXr3Ye/C4lXnfRNmxOa/TZcIiWKoiyc/XbZGzAR0fxiI/IEOkpic+BFhYRP6BiXa7cvTx4fsOSg==
dist.shasum: e0940e4de0a4ea860e41af8af01248a11e85a1db
tarball SHA-256: bf8543506740f8ba36bbec24a5e02937dd8b514237fc951db69d9de394965c74
```

Package [`@anthropic-ai/claude-code-darwin-arm64@2.1.317`](https://registry.npmjs.org/@anthropic-ai/claude-code-darwin-arm64/2.1.317):

```text
dist.integrity: sha512-Xnf9K4aGBuoYaODmEZ0o/lRvuYUCv2ZIDP+AXv15WVITryGgrsSA7CFL7KH+43kC6WhK8i6oysF6EbGdgBW4Aw==
dist.shasum: bea435ebb36755ddb825b047c0d46d45da2f7bb8
tarball SHA-256: d572e006fecf09f14c544f2995906d332272647470472f2c8b7158686d90b4f7
```

The integrity assertion was equivalent to:

```python
algorithm, expected = metadata["dist"]["integrity"].split("-", 1)
assert algorithm == "sha512"
assert base64.b64encode(hashlib.sha512(payload).digest()).decode() == expected
assert hashlib.sha1(payload).hexdigest() == metadata["dist"]["shasum"]
```

The metadata and tarball share the registry trust boundary; matching integrity
is package-byte verification, not an independent publisher signature.

## Isolated command form

`$CLAUDE` is the exact executable above; `$DISPOSABLE_HOME` and the working
directory are separate newly created absolute disposable directories for
every invocation. `$SANDBOX` is the offline Apple sandbox profile described
in the handoff. `$DISPOSABLE_HOME/tmp` exists. No real configuration or
credentials were copied. All probes used closed stdin and a 25-second bound;
none timed out. The equivalent sanitized command prefix is:

```sh
probe() {
  env -i HOME="$DISPOSABLE_HOME" PATH=/usr/bin:/bin:/usr/sbin:/sbin \
    TMPDIR="$DISPOSABLE_HOME/tmp" LANG=C LC_ALL=C \
    DISABLE_AUTOUPDATER=1 DISABLE_TELEMETRY=1 \
    CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1 \
    /usr/bin/sandbox-exec -f "$SANDBOX" "$CLAUDE" "$@" </dev/null
}
```

`probe --version` and `probe --help` both exited 0. Every other probe included
`--help` before any native `--`. The missing-mode case consumed it as a value
and refused before startup; all other runs printed help or refused. Print
mode was inspected using help only; no print response was requested.

## Root and print help

Root help: `Usage: claude [options] [command] [prompt]`.
Print help (`probe -p --help` and `probe --print --help`):
`Usage: claude [options] [prompt]`. All exited 0 with empty stderr.
Root help SHA-256:
`5ca3a9dac2026e56c221fdd34380c7d7d45ebbb51c1d01b7c17c75cb076a0be8`.

Relevant root-help extracts, with whitespace normalized:

```text
--dangerously-skip-permissions  Bypass all permission checks.
--allow-dangerously-skip-permissions  Enable bypassing all permission checks
as an option, without it being enabled by default.
--permission-mode <mode>  Permission mode to use for the session
(choices: "acceptEdits", "auto", "bypassPermissions", "manual", "dontAsk", "plan")
--allowedTools, --allowed-tools <tools...>
--disallowedTools, --disallowed-tools <tools...>
--settings <file-or-json>  Path to a settings JSON file or a JSON string
to load additional settings from
--mcp-config <configs...>  Load MCP servers from JSON files or strings
(space-separated)
--strict-mcp-config  Only use MCP servers from --mcp-config,
ignoring all other MCP configurations
--setting-sources <sources>  Comma-separated list of setting sources to load
(user, project, local).
--permission-prompts <target>  (choices: "host", "none", default: "host")
-p, --print  Print response and exit (useful for pipes).
```

The full exact sanitized help and both output streams of every probe are
stored in the adjacent observations JSON. `stdout_ref` / `stderr_ref` name
exact strings in its `outputs` map; this avoids duplicating help per command.
The table below is a readable projection, with stderr kept under output IDs.
Exit 0 with help proves only the observed help path. In particular it can
mask unknown flags and invalid settings/MCP content.

## Command matrix

Every command starts with the isolated `probe` above. `Root help` and `Print
help` refer to the exact corresponding output; error IDs are expanded below.

| Probe | Sanitized command | Observed output | Exit |
| --- | --- | --- | ---: |
| P01 | `probe --version` | `2.1.317 (Claude Code)`; stderr empty | 0 |
| P02 | `probe --help` | Root help; stderr empty | 0 |
| P03 | `probe -p --help` | Print help; stderr empty | 0 |
| P04 | `probe --print --help` | Print help; stderr empty | 0 |
| P05 | `probe --permission-mode acceptEdits --help` | Root help; stderr empty | 0 |
| P06 | `probe --permission-mode=acceptEdits --help` | Root help; stderr empty | 0 |
| P07 | `probe --permission-mode auto --help` | Root help; stderr empty | 0 |
| P08 | `probe --permission-mode=auto --help` | Root help; stderr empty | 0 |
| P09 | `probe --permission-mode bypassPermissions --help` | Root help; stderr empty | 0 |
| P10 | `probe --permission-mode=bypassPermissions --help` | Root help; stderr empty | 0 |
| P11 | `probe --permission-mode manual --help` | Root help; stderr empty | 0 |
| P12 | `probe --permission-mode=manual --help` | Root help; stderr empty | 0 |
| P13 | `probe --permission-mode dontAsk --help` | Root help; stderr empty | 0 |
| P14 | `probe --permission-mode=dontAsk --help` | Root help; stderr empty | 0 |
| P15 | `probe --permission-mode plan --help` | Root help; stderr empty | 0 |
| P16 | `probe --permission-mode=plan --help` | Root help; stderr empty | 0 |
| P17 | `probe --permission-mode invalid-mode --help` | stderr `O05`; stdout empty | 1 |
| P18 | `probe --permission-mode=invalid-mode --help` | stderr `O05`; stdout empty | 1 |
| P19 | `probe --permission-mode default --help` | Root help; stderr empty | 0 |
| P20 | `probe --permission-mode --help` | stderr `O06`; stdout empty | 1 |
| P21 | `probe --help --permission-mode=invalid-mode` | stderr `O05`; stdout empty | 1 |
| P22 | `probe --dangerously-skip-permissions --help` | Root help; stderr empty | 0 |
| P23 | `probe --dangerously-skip-permissions=true --help` | Root help; stderr empty | 0 |
| P24 | `probe --allow-dangerously-skip-permissions --help` | Root help; stderr empty | 0 |
| P25 | `probe --restricted --help` | Root help; stderr empty | 0 |
| P26 | `probe --sandbox danger-full-access --help` | Root help; stderr empty | 0 |
| P27 | `probe --sandbox-mode=unrestricted --help` | Root help; stderr empty | 0 |
| P28 | `probe --allowedTools Read Edit --help` | Root help; stderr empty | 0 |
| P29 | `probe --allowedTools=Read,Edit --help` | Root help; stderr empty | 0 |
| P30 | `probe --allowed-tools=Read --help` | Root help; stderr empty | 0 |
| P31 | `probe --disallowedTools=AskUserQuestion --help` | Root help; stderr empty | 0 |
| P32 | `probe --disallowed-tools=AskUserQuestion --help` | Root help; stderr empty | 0 |
| P33 | `probe --tools '' --help` | Root help; stderr empty | 0 |
| P34 | `probe --settings '{}' --help` | Root help; stderr empty | 0 |
| P35 | `probe '--settings={"permissions":{"defaultMode":"dontAsk","allow":[],"deny":["mcp__*"]}}' --help` | Root help; stderr empty | 0 |
| P36 | `probe --settings not-json --help` | stderr `O07`; stdout empty | 1 |
| P37 | `probe --mcp-config '{"mcpServers":{}}' --help` | Root help; stderr empty | 0 |
| P38 | `probe '--mcp-config={"mcpServers":{}}' --help` | Root help; stderr empty | 0 |
| P39 | `probe --strict-mcp-config --help` | Root help; stderr empty | 0 |
| P40 | `probe --setting-sources user --help` | Root help; stderr empty | 0 |
| P41 | `probe --setting-sources '' --help` | Root help; stderr empty | 0 |
| P42 | `probe --setting-sources invalid-source --help` | stderr `O08`; stdout empty | 1 |
| P43 | `probe --permission-prompts none --help` | Root help; stderr empty | 0 |
| P44 | `probe --permission-prompts=invalid-target --help` | stderr `O09`; stdout empty | 1 |
| P45 | `probe --output-format=json -p --help` | Print help; stderr empty | 0 |
| P46 | `probe --output-format=invalid-format -p --help` | stderr `O10`; stdout empty | 1 |
| P47 | `probe --print=true --help` | Root help; stderr empty | 0 |
| P48 | `probe --model claude-fable-5-1 --effort low --dangerously-skip-permissions --disallowedTools=AskUserQuestion --help` | Root help; stderr empty | 0 |
| P49 | `probe --setting-sources user '--mcp-config={"mcpServers":{}}' --strict-mcp-config --permission-mode dontAsk --tools '' --help` | Root help; stderr empty | 0 |
| P50 | `probe QUALIFICATION_TEXT --permission-mode=invalid-mode --help` | stderr `O05`; stdout empty | 1 |
| P51 | `probe --help -- --permission-mode=invalid-mode` | Root help; stderr empty | 0 |
| P52 | `probe --help -- --dangerously-skip-permissions` | Root help; stderr empty | 0 |
| P53 | `probe --settings '{"permissions":{"defaultMode":"invalid-mode"}}' --help` | Root help; stderr empty | 0 |
| P54 | `probe --settings '{"sandbox":{"enabled":true}}' --help` | Root help; stderr empty | 0 |
| P55 | `probe --settings '{' --help` | stderr `O11`; stdout empty | 1 |
| P56 | `probe --mcp-config '{"mcpServers":{"invalid":{}}}' --help` | Root help; stderr empty | 0 |
| P57 | `probe --permission-prompts host --help` | Root help; stderr empty | 0 |

Exact stderr for refusal rows (disposable locators normalized):

`O05`:

```text
error: option '--permission-mode <mode>' argument 'invalid-mode' is invalid. Allowed choices are acceptEdits, auto, bypassPermissions, manual, dontAsk, plan.
```

`O06`:

```text
error: option '--permission-mode <mode>' argument '--help' is invalid. Allowed choices are acceptEdits, auto, bypassPermissions, manual, dontAsk, plan.
```

`O07`:

```text
Error: Settings file not found: $SCRATCH/317/cwd-035/not-json
```

`O08`:

```text
Error processing --setting-sources: Invalid setting source: invalid-source. Valid options are: user, project, local
```

`O09`:

```text
error: option '--permission-prompts <target>' argument 'invalid-target' is invalid. Allowed choices are host, none.
```

`O10`:

```text
error: option '--output-format <format>' argument 'invalid-format' is invalid. Allowed choices are text, json, stream-json.
```

`O11`:

```text
Error: Settings file not found: $SCRATCH/317/cwd-102/{
```

## Interpretation and limits

All six declared permission modes reached help in separate and equals forms.
`invalid-mode` refused in both forms, after positional text, and even when
`--help` appeared first. The missing-value probe consumed `--help` as the mode
and refused it as an invalid choice. In contrast, `--help --` followed by
flag-looking text printed help; the terminator protects data from option
parsing. The launcher's outer `--` forwards a native suffix and is different
from a native `--` inside that suffix.

`default` reached help despite not being listed. `--print=true` and
`--dangerously-skip-permissions=true` also reached root help. None establishes
runtime semantics or expands the gate's closed selector subset. No original
2.1.261 binary was tested, so these are observations, not proven new features.

Claude's generated yolo mapping uses no independent sandbox selector. Neither
`--sandbox` nor `--sandbox-mode` appears in root help. Both unknown examples
nevertheless reached help, so their exit 0 is deliberately retained as a
counterexample to interpreting help success as flag support.

Settings with an invalid permission default and an invalid MCP server object
also reached help. Missing settings paths and invalid setting sources did
refuse; malformed `{` was interpreted as a missing filename. These paths do
not qualify settings application, JSON schema validation, precedence, sandbox
enforcement, effective tool lists, or MCP connection behavior. `--allowedTools`
was inspected because the scanner checks re-enable attempts; the generated
launcher denial is `--disallowedTools=AskUserQuestion`, not an allow list.

The two combined-prefix probes cover launcher model/effort/bypass/tool-denial
spellings and Curator's managed settings-source/MCP/dontAsk/tools spellings
on the help path. They do not qualify model availability or a real launch.
No model turn, authentication, credential access, provider traffic, session,
MCP process or shared daemon was requested. Linux, other architectures,
consumer activation and independent acceptance remain unqualified. See the
handoff for current V3 requirements and the Linux evidence boundary.


## Review supplement: current V3 concerns

The following 39 observations supplement, rather than replace, P01–P57.
The executable digest above was checked again before and after this supplement
and remained identical. Every invocation used the same isolation prefix,
a fresh HOME and cwd, closed stdin, OS network denial and a 25-second bound;
all completed within the bound. No inherited selectors or credentials were
provided. H01 and H02 repeat identity and root help as positive controls.

`$FIXTURE_ROOT` is an absolute directory in this invocation's disposable cwd.
`$SETTINGS_FILE` contains `{"env":{},"permissions":{"allow":[],"deny":[]}}`;
`$DISABLE_FILE` contains
`{"permissions":{"disableBypassPermissionsMode":"disable"}}`.
Both are newly written fixture files beneath that directory, with a trailing
newline. Nothing was copied from real configuration.

| Probe | Sanitized command | Observed output | Exit |
| --- | --- | --- | ---: |
| H01 | `probe --version` | `2.1.317 (Claude Code)`; stderr empty | 0 |
| H02 | `probe --help` | Root help; stderr empty | 0 |
| H03 | `probe --inherit-permission-mode plan --help` | Root help; stderr empty | 0 |
| H04 | `probe --inherit-permission-mode=bypassPermissions --help` | Root help; stderr empty | 0 |
| H05 | `probe --inherit-permission-mode invalid-mode --help` | stderr `O12`; stdout empty | 1 |
| H06 | `probe --inherit-permission-mode=invalid-mode --help` | stderr `O12`; stdout empty | 1 |
| H07 | `probe --plan-mode-required --help` | Root help; stderr empty | 0 |
| H08 | `probe --enable-auto-mode --help` | Root help; stderr empty | 0 |
| H09 | `probe --managed-settings '{}' --help` | Root help; stderr empty | 0 |
| H10 | `probe '--managed-settings={}' --help` | Root help; stderr empty | 0 |
| H11 | `probe --project-config-root '$FIXTURE_ROOT' --help` | Root help; stderr empty | 0 |
| H12 | `probe --fast-mode on --help` | Root help; stderr empty | 0 |
| H13 | `probe --fast-mode=off --help` | Root help; stderr empty | 0 |
| H14 | `probe --fast-mode invalid-fast-mode --help` | Root help; stderr empty | 0 |
| H15 | `probe --runner-declares grammar-control --help` | Root help; stderr empty | 0 |
| H16 | `probe --runner-declares=grammar-control --help` | Root help; stderr empty | 0 |
| H17 | `probe --inherit-permission-mode=invalid-mode --help` | stderr `O12`; stdout empty | 1 |
| H18 | `probe --runner-declares --inherit-permission-mode=invalid-mode --help` | Root help; stderr empty | 0 |
| H19 | `probe --runner-declares grammar-control --inherit-permission-mode=invalid-mode --help` | stderr `O12`; stdout empty | 1 |
| H20 | `probe --settings '{"env":{}}' --help` | Root help; stderr empty | 0 |
| H21 | `probe --settings '{"env":{"CLAUDE_CODE_SUBPROCESS_ENV_SCRUB":"1"}}' --help` | Root help; stderr empty | 0 |
| H22 | `probe '--settings={"env":{"CLAUDE_CODE_SUBPROCESS_ENV_SCRUB":"1"}}' --help` | Root help; stderr empty | 0 |
| H23 | `probe --settings '{"env":"invalid-shape"}' --help` | Root help; stderr empty | 0 |
| H24 | `probe --settings '{"hooks":{}}' --help` | Root help; stderr empty | 0 |
| H25 | `probe --settings '{"permissions":{"disableBypassPermissionsMode":"disable"}}' --help` | Root help; stderr empty | 0 |
| H26 | `probe --settings '{"permissions":{"defaultMode":"plan"}}' --help` | Root help; stderr empty | 0 |
| H27 | `probe --settings '{"permissions":{"allow":["Read"],"deny":[],"ask":[],"additionalDirectories":[]}}' --help` | Root help; stderr empty | 0 |
| H28 | `probe --settings '{"permissions":{"deny":null}}' --help` | Root help; stderr empty | 0 |
| H29 | `probe --settings '{"model":1}' --help` | Root help; stderr empty | 0 |
| H30 | `probe --settings '{"futureSetting":1}' --help` | Root help; stderr empty | 0 |
| H31 | `probe --settings '$SETTINGS_FILE' --help` | Root help; stderr empty | 0 |
| H32 | `probe --settings '$DISABLE_FILE' --help` | Root help; stderr empty | 0 |
| H33 | `probe --settings '{"permissions":{"disableBypassPermissionsMode":"disable"}}' --settings '{}' --help` | Root help; stderr empty | 0 |
| H34 | `probe --settings '{}' --settings '{"permissions":{"disableBypassPermissionsMode":"disable"}}' --help` | Root help; stderr empty | 0 |
| H35 | `probe --runner-declares --settings '{}' --help` | Root help; stderr empty | 0 |
| H36 | `probe --fast-mode --settings '{}' --help` | Root help; stderr empty | 0 |
| H37 | `probe --help -- --inherit-permission-mode=invalid-mode` | Root help; stderr empty | 0 |
| H38 | `probe --fast-mode --inherit-permission-mode=invalid-mode --help` | Root help; stderr empty | 0 |
| H39 | `probe --fast-mode on --inherit-permission-mode=invalid-mode --help` | stderr `O12`; stdout empty | 1 |

`O12` (exact stderr):

```text
error: option '--inherit-permission-mode <mode>' argument 'invalid-mode' is invalid. Allowed choices are acceptEdits, auto, bypassPermissions, manual, dontAsk, plan.
```

H05/H06/H17 refuse an invalid inherited mode. H03/H04 reach help with known
modes; this establishes parsing of `--inherit-permission-mode`, including on
2.1.319. It does not establish how inheritance interacts with the bypass.
H18 versus H17/H19, and H38 versus H17/H39, show that `--runner-declares` and
`--fast-mode` consume the next flag-looking token in these concrete cases.
Their values are supplied without consuming the final help flag. H35/H36
alone cannot establish effective settings selection.

H20–H34 reach help even for settings that current V3 would refuse under yolo:
nonempty `env`, `hooks`, permission defaults, bypass disabling, unknown keys
and malformed member shapes. These are native help-path observations, not
SAM refusal measurements. They do not demonstrate settings application,
last-source precedence, or the eager scanner's agreement with Commander.
The known permission-prompts choice refusal is already in P43/P44/P57.

No unknown hidden-option enumeration or missing-value hidden scalar probes
were run: a required scalar could consume the only help flag and start a
session. Runtime bypass interaction, hook execution, managed configuration
loading, effective permissions and session-based settings checks are
**NOT_RUN** under the no-session constraint; work stops at the recorded help
or parser boundary for those rows. No session was attempted to fill them.
The pinned parser/list/eager comparisons used for 2.1.312 remain **NOT_RUN**.
Help can mask unknown options, so success alone does not prove that every
hidden selector shown here is recognized or has the assumed arity.
