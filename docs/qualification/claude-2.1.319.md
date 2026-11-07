# Claude Code 2.1.319 permission grammar qualification

Observed 2026-11-07 on macOS 27.2 (26B123), Darwin 27.2.0, arm64.
Task: TASK-261107-oe1ki9 — qualify-claude-permission-grammar-2-1-317-and-319.
This is a sanitized real-binary help/parser record following the
[Codex precedent](codex-0.164.0.md). It does not admit a runtime or Linux tuple.
The [gate analysis and producer handoff](claude-permission-grammar-handoff.md)
contains the baseline comparison, proposed rows, isolation controls and gaps.

## Exact binary

| Property | Observation |
| --- | --- |
| `probe --version` | `2.1.319 (Claude Code)`; stdout; exit 0; stderr empty |
| SHA-256 before and after all probes | `1f53587e92dee8e089fa349188a4cda64fa34ace6e7b0ca2ee278ddaa3cc2ebd` |
| Executable format | Mach-O 64-bit arm64 (header CPU type `0x0100000c`) |
| Provenance | Installed native executable, read-only use; original locator omitted. |

Version and hash identify the executable actually tested, not a reported
installation on another machine. Hashing used Python `hashlib.sha256` over
the executable bytes; no installed file was modified.

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
| P01 | `probe --version` | `2.1.319 (Claude Code)`; stderr empty | 0 |
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
Error: Settings file not found: $SCRATCH/319/cwd-035/not-json
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
Error: Settings file not found: $SCRATCH/319/cwd-102/{
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
handoff for proposed upstream V2 rows and the exact agents2 evidence request.
