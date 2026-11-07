# Claude Code 2.1.317 / 2.1.319 permission grammar: producer handoff

TASK-261107-oe1ki9 — qualify-claude-permission-grammar-2-1-317-and-319,
under STORY-261105-fevga9 — profile-launches-on-linux-agent-hosts.
Observed 2026-11-07. PR base: `main` at
`ed5688b7121c56f23ca18e80e8e51d65d82fbbd2`.

**The launcher-used bypass spelling and six documented permission modes show
no observed drift from the recorded 2.1.261 contract.** These are macOS arm64
help/parser observations, not permission enforcement, Linux qualification or
runtime acceptance. The gate remains closed for both releases. No gate,
launcher, dependency pin or installed executable was changed.

- [2.1.317 record](claude-2.1.317.md) and [exact sanitized observations](claude-2.1.317-observations.json).
- [2.1.319 record](claude-2.1.319.md) and [exact sanitized observations](claude-2.1.319-observations.json).

## Gate ownership and consumption

Source inspection used `skill-agents-management` commit
`46d17fb7d8eb0de79dafd91a971de6344d34ebfb` and
`curator-agent-launcher` commit `2045339f703ee8190705c116a05197b12b684b71`.
These are source checkpoints, not evidence of the installed launcher's build.

| Surface | Actual responsibility |
| --- | --- |
| agents-management [`internal/releasegrammar/release.go`](https://github.com/relux-works/skill-agents-management/blob/46d17fb7d8eb0de79dafd91a971de6344d34ebfb/internal/releasegrammar/release.go) | `IsReleaseTriple` accepts three numeric components. It contains no permission-capability rows and no version range admission. |
| [`pkg/agentic/systems/claude/probe.go`](https://github.com/relux-works/skill-agents-management/blob/46d17fb7d8eb0de79dafd91a971de6344d34ebfb/pkg/agentic/systems/claude/probe.go) | Resolves the launch binary, runs `--version`, parses the first field of its first line as an exact release triple. |
| [`pkg/agentic/systems/claude/policy.go`](https://github.com/relux-works/skill-agents-management/blob/46d17fb7d8eb0de79dafd91a971de6344d34ebfb/pkg/agentic/systems/claude/policy.go) | `verifiedReleases` contains only `2.1.261`, grammar V2, yolo supported. `permissionMapping` and `scanNativePolicy` own the bypass mapping and closed conflict set. |
| [`pkg/agentic/system.go`](https://github.com/relux-works/skill-agents-management/blob/46d17fb7d8eb0de79dafd91a971de6344d34ebfb/pkg/agentic/system.go) | `LookupReleaseCapability` matches an exact release, refuses unknown releases/grammars, and supplies the admitted mapping. A neighboring version is not admitted. |
| Launcher [`cmd/curator-run/main.go`](https://github.com/relux-works/curator-agent-launcher/blob/2045339f703ee8190705c116a05197b12b684b71/cmd/curator-run/main.go) | `resolvePermission` calls `ProbeToolRelease` and the registry's `PermissionMapping`. An unverified yolo mapping becomes `plan_refused` before execution. Native tolerates an unverified mapping and adds no permission override. |
| Launcher [`internal/plan/plan.go`](https://github.com/relux-works/curator-agent-launcher/blob/2045339f703ee8190705c116a05197b12b684b71/internal/plan/plan.go) | Carries release/mode/native suffix through `SpawnRequest` into `vendorplugin.BuildLaunch` in interactive mode; wraps spawn-plane refusals as `plan_refused`. |

The launcher source pins agents-management `v0.5.74`. Curator's root module
does not own or link this permission table. Its `internal/envregistry` Claude
`VerifiedRelease` is a separate environment/credential fact, not this gate.
Adding these documents does not change any launch decision.

The [Codex 0.164.0 precedent](codex-0.164.0.md) records a V3 exact-release
mapping supported by help, invalid values, startup config/schema and a
standalone config/read exchange, with no model turn. It explicitly retains
refusals and distinguishes schema from runtime validation. Claude has no
equivalent config/read observation here: help success must not be promoted
to effective settings. The adjacent Muse records likewise separate passing
observations from failed/pending admission and consumer ownership.

## Evidence against the recorded 2.1.261 row

The baseline is the stored row, source comments and existing policy tests,
not a newly tested 2.1.261 executable. Some comments cite help at 2.1.273/274,
and the option-ownership tables are pinned to 2.1.288. No claim of a complete
historical 2.1.261 help diff follows.

| Contract | Recorded baseline | Both tested releases | Assessment |
| --- | --- | --- | --- |
| Yolo mapping | `--dangerously-skip-permissions` | Documented; help invocation exits 0 | Spelling preserved; enforcement untested |
| Permission modes | `acceptEdits`, `auto`, `bypassPermissions`, `manual`, `dontAsk`, `plan` | Same help choices; all six tested in separate and equals forms, exit 0; `invalid-mode` refused, exit 1 | No drift observed in this closed subset |
| Extra selectors | `--allow-dangerously-skip-permissions`, `--restricted` conflict with mapped yolo | Both remain documented; help invocations exit 0 | Preserve upstream conflict refusals; no combined execution tested |
| Undocumented mode | Gate refuses values outside six | `--permission-mode default --help` exits 0 despite absence from choices | New observation, not a proven change since 2.1.261; do not expand the gate's set from this |
| Flag placement | Scan excludes native `--` data; generated bypass precedes native suffix | Invalid mode before native `--`, including after positional text and after `--help`, refuses; after native `--`, help exits 0 | Observed terminator boundary preserved |
| Print form | `-p` / `--print` | Both produce print-specific help, exit 0 | No response/session was requested |
| Settings/MCP/tool flags | Launcher uses `--settings`, variadic `--mcp-config`, `--disallowedTools=AskUserQuestion`; policy scanner also recognizes allowed-tool forms | Relevant forms reach help; combined generated-prefix probes reach help | Syntax only; no MCP process or tool was started |
| Sandbox | Mapping uses bypass, no separate Claude sandbox flag | Root help lists no `--sandbox` or `--sandbox-mode`; unknown examples still reach help | Help masks unknown options; those examples do not qualify sandbox flags |

Between 2.1.317 and 2.1.319, complete root help bytes are identical (SHA-256
`5ca3a9dac2026e56c221fdd34380c7d7d45ebbb51c1d01b7c17c75cb076a0be8`).
All 57 corresponding observations have the same argv, exit and output after
normalizing the disposable per-version output locator and the version banner.
This comparison excludes timing and binary bytes.

`--settings` containing an invalid `permissions.defaultMode` and an invalid
MCP server object both reach help successfully. In contrast, an absent
settings filename and invalid `--setting-sources` refuse before help. These
counterexamples prevent treating all help successes as settings validation.
`--print=true` and a boolean bypass with `=true` also reach root help; that
does not prove their runtime meaning or equivalence to the canonical flags.

## Proposed upstream rows, not applied

For the gate owner, exactly the existing `policy.go` row format:

```go
{Release: "2.1.317", Grammar: agentic.PermissionGrammarV2, YoloSupported: true},
{Release: "2.1.319", Grammar: agentic.PermissionGrammarV2, YoloSupported: true},
```

These proposals are backed only by the measured macOS arm64 parser subset.
`ReleaseCapability` contains no OS, architecture, binary digest or evidence
field: adding a row would enable an exact **version** lookup more broadly than
these observations. The gate owner must resolve that admission boundary and
obtain Linux evidence before claiming Linux support. This producer does not
authorize insertion, release, pin bump, installation or activation.

No change to Curator or launcher check logic is indicated by the measured
flags. The separate lane would need to review/add rows and relevant tests in
agents-management, release it, and update the launcher's linked module pin
through its own delivery flow. If the owner requires platform/hash-bound
admission, the present table cannot express it: that needs a separate design
decision in agents-management and its consumer. No such code is written here.

## Linux evidence required from agents2

Supply an independently produced sanitized record for every exact version and
Linux OS/architecture/libc tuple being admitted. Record the executed binary's
`--version`, SHA-256 before/after, package provenance and published integrity
when npm is used, kernel/distribution/architecture/libc, exact sanitized argv,
stdout/stderr and exit codes. Bind all results to that binary, not a report
from another machine or a matching version string alone.

Replay this parser matrix under a fresh absolute HOME and scrubbed selectors,
with no copied configuration/credentials, no provider network, no login/model
turn/session, no shared daemon and no install scripts. Supply a Linux-native
network-denial control and sanitizer positive/negative controls; macOS
`sandbox-exec` is not portable evidence. Preserve actual invalid-value refusals,
settings/MCP validation limits, print help and native `--` placement. If the
binary needs an install step, stop that version and record the limitation.
No Linux binary was executed here; neither Linux architecture is qualified.

## Isolation and evidence controls

Each native invocation ran in its own empty absolute HOME and working
directory, with stdin closed and a 25-second timeout. The child environment
was built from an empty dictionary, equivalent to `env -i`: only HOME, a
system-only PATH, disposable TMPDIR, `LANG=C`, `LC_ALL=C`,
`DISABLE_AUTOUPDATER=1`, `DISABLE_TELEMETRY=1` and
`CLAUDE_CODE_DISABLE_NONESSENTIAL_TRAFFIC=1` were present. Thus XDG, Claude,
Codex, Curator, provider, credential, proxy, socket and inherited Git selectors
were absent apart from the explicitly named traffic-disable control.

Every native invocation used an Apple sandbox profile with `(allow default)`,
`(deny network*)`, file-read/write denial for the real native configuration
directories and Keychains directory, and security-service mach-lookup denial.
No configuration or credentials were copied. Only version/help invocations
were admitted; help remained before native `--`. No background session, daemon,
MCP server, login, permission exercise or model turn was launched. The installed
2.1.319 file was read/executed in place; its SHA-256 remained unchanged.

| Control | Positive | Negative / observed refusal | Claim and limit |
| --- | --- | --- | --- |
| OS network guard | Disposable Python process binds ephemeral loopback: `LOOPBACK_BIND_OK`, exit 0 | Same operation under probe sandbox: `PermissionError: [Errno 1] Operation not permitted`, exit 1 | Network operation denied; no external endpoint or provider contacted |
| OS file guard | `cat` of synthetic protected fixture: `CONTROL_ONLY`, exit 0 | Same file with an added literal deny: `Operation not permitted`, exit 1; bytes unchanged | Sandbox file-denial mechanism works; no real configuration file was opened for a control |
| Sanitizer | Clean symbolic binary record accepted | In-memory record planted with the actual home prefix refused; replacement with `$REAL_HOME` accepted | The actual home string was used without reading it and never printed or persisted in public output |

Sanitization replaced exact private executable and scratch locators with
`$CLAUDE319` / `$SCRATCH`, then the remaining actual home prefix with
`$REAL_HOME`. The public-output checker rejected that actual prefix, Unix
user-home prefixes and forbidden scratch-directory spellings. The control
persisted only these results:

```json
{"real_home_plant_detected":true,"sanitized_plant":"binary record: $REAL_HOME/.claude/CONTROL_ONLY","sanitized_plant_passed":true,"clean_positive_passed":true}
```

The core of that executed control was:

```python
actual_home = str(pathlib.Path.home())  # string only; no file read
def contains_actual_home(text):
    return actual_home in text
plant = "binary record: " + actual_home + "/.claude/CONTROL_ONLY"
assert contains_actual_home(plant)
cleaned = sanitize(plant)  # same sanitizer used for the probe output
assert not contains_actual_home(cleaned)
assert cleaned == "binary record: $REAL_HOME/.claude/CONTROL_ONLY"
assert not contains_actual_home("binary record: $CLAUDE319")
```

Private intake, PID and command logs remain in ignored task-scoped scratch.
Only sanitized records are published. No full suite was run for this
documentation/evidence change. Producer handoff is ready for review; hosted
checks, independent acceptance, landing, tags and runtime qualification are
outside this delivery.
