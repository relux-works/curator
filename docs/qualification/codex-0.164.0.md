# Codex 0.164.0 permission grammar qualification

Observed 2026-10-27 on macOS arm64. This is a sanitized selection of the
real-binary observations supporting the skill-agents-management 0.164.0 row.
Commands below use `codex` for the installed executable, `$CODEX_HOME` for
an existing disposable home, and `$SCRATCH` for disposable output. Native
user config, authentication and sessions were not copied or edited. No model
turn, shared daemon restart or consumer launch was performed. Every native
invocation explicitly set `CODEX_HOME`; other home selectors and `XDG_*` were cleared.

## Binary identity

```sh
codex --version
shasum -a 256 "$(command -v codex)"
```

Relevant output (the checksum output path is normalized to `codex`):

```text
codex-cli 0.164.0
d8990e3ee7acfbc26f91265b12a1a83f0fee40d0e0f8f1f40a37d4ec21f499cf  codex
```

## CLI grammar

`codex --help` exited 0 and declared these forms:

```text
exec              Run Codex non-interactively [aliases: e]
-s, --sandbox <SANDBOX_MODE>
[possible values: read-only, workspace-write, danger-full-access]
-a, --ask-for-approval <APPROVAL_POLICY>
- on-request: The model decides when to ask the user for approval
- never:      Never ask for user approval
--dangerously-bypass-approvals-and-sandbox
Skip all confirmation prompts and execute commands without sandboxing.
--approve-for-me
Route approval requests through automatic review using the workspace-write sandbox
```

Descriptions above are relevant line extracts, with indentation normalized.
The following parsing observations use help only, not execution:

| Command | Relevant output or result |
| --- | --- |
| `codex exec --help` | `Usage: codex exec [OPTIONS] [PROMPT]`; exit 0 |
| `codex e --help` | Same exec usage; exit 0 |
| `codex resume --help` | `Usage: codex resume [OPTIONS] [SESSION_ID] [PROMPT]`; exit 0 |
| `codex -s read-only --help` | Root help; exit 0 |
| `codex --sandbox=workspace-write --help` | Root help; exit 0 |
| `codex -s danger-full-access --help` | Root help; exit 0 |
| `codex --sandbox=unrestricted --help` | `error: invalid value 'unrestricted' for '--sandbox <SANDBOX_MODE>'`; exit 2 |
| `codex -a on-request --help` | Root help; exit 0 |
| `codex --ask-for-approval=never --help` | Root help; exit 0 |
| `codex -a untrusted --help` | `error: invalid value 'untrusted' for '--ask-for-approval <APPROVAL_POLICY>'`; exit 2 |
| `codex -a on-failure --help` | `error: invalid value 'on-failure' for '--ask-for-approval <APPROVAL_POLICY>'`; exit 2 |
| `codex --ask-for-approval=sometimes --help` | `error: invalid value 'sometimes' for '--ask-for-approval <APPROVAL_POLICY>'`; exit 2 |
| `codex --dangerously-bypass-approvals-and-sandbox --help` | Root help; exit 0 |
| `codex --yolo --help` | Root help; exit 0 |
| `codex --approve-for-me --help` | Root help; exit 0 |
| `codex --not-so-yolo --help` | Root help; exit 0 |
| `codex -anever -h` | Root usage; exit 0 |
| `codex -sdanger-full-access -h` | Root usage; exit 0 |
| `codex --yolo -h` | Root usage; exit 0 |
| `codex --not-so-yolo -h` | Root usage; exit 0 |
| `codex exec --sandbox=read-only -h` | Exec usage; exit 0 |
| `codex exec -a never -h` | `error: unexpected argument '-a' found`; exit 2 |

The last rejection is retained. Its native suggestion to put the flag after
`--` does not qualify a permission override there; that position is data.
Parsing hidden aliases or compact short forms does not establish their
interaction with the canonical bypass.

## Config overrides and startup validation

`codex --help` describes `-c, --config <key=value>`: values are parsed as TOML,
with a raw-string fallback; dotted keys address nested values. It describes
base config under `$CODEX_HOME` and named profile config layered above it.

| Command | Relevant output or result |
| --- | --- |
| `codex -c 'approval_policy="never"' --help` | Root help; exit 0 |
| `codex '--config=sandbox_mode="workspace-write"' --help` | Root help; exit 0 |
| `codex '-csandbox_permissions=["disk-full-read-access"]' --help` | Root help; exit 0 |

These three commands establish syntax only. Actual config validation used
`codex -c '<override>' app-server --print-startup-config` for each override:

| Override in that command | Relevant output or result |
| --- | --- |
| `approval_policy="on-request"` | `"appServerVersion":"0.164.0"`; exit 0 |
| `approval_policy="never"` | `"appServerVersion":"0.164.0"`; exit 0 |
| `approval_policy="on-failure"` | `"appServerVersion":"0.164.0"`; exit 0 |
| `approval_policy="untrusted"` | `approval_policy = "untrusted" is no longer supported; remove this setting`; exit 1 |
| `approval_policy="sometimes"` | `unknown variant` for `sometimes`; exit 1 |
| `sandbox_mode="read-only"` | `"appServerVersion":"0.164.0"`; exit 0 |
| `sandbox_mode="workspace-write"` | `"appServerVersion":"0.164.0"`; exit 0 |
| `sandbox_mode="danger-full-access"` | `"appServerVersion":"0.164.0"`; exit 0 |
| `sandbox_mode="unrestricted"` | `unknown variant` for `unrestricted`; exit 1 |
| `approval_policy={granular={sandbox_approval=true,rules=true,mcp_elicitations=true}}` | `"appServerVersion":"0.164.0"`; exit 0 |
| `approval_policy={granular={}}` | `missing field` for `sandbox_approval`; exit 1 |

Successful startup output also reported `"requirements":null`. This does not
establish every policy combination or the rights of an executing session.

## Binary-generated schema

```sh
codex app-server generate-json-schema --help
codex app-server generate-json-schema --out "$SCRATCH/schema"
```

Both exited 0. The generated `ConfigReadResponse.json` under the schema's
`v2` directory had SHA-256
`69a32e3d84e16045802dff8abf01770d759ebc758965c3258a30ea4a93159fce`.
Selected definitions were inspected with this normalized extraction command:

```sh
python3 - "$SCRATCH/schema/v2/ConfigReadResponse.json" <<'PY'
import json, sys
s = json.load(open(sys.argv[1]))
print(json.dumps({k: s['definitions'][k] for k in
    ['AskForApproval', 'SandboxMode', 'ConfigLayerSource', 'SandboxWorkspaceWrite']}))
PY
```

Relevant fields: `AskForApproval` lists `untrusted`, `on-request`, `never`
and a granular object requiring `mcp_elicitations`, `rules`, `sandbox_approval`.
`SandboxMode` lists `read-only`, `workspace-write`, `danger-full-access`.
`SandboxWorkspaceWrite.network_access` defaults false. Config layers include
packaged defaults, system, user, project, session flags and managed sources.
This is the config/read protocol schema, not a complete config-file schema;
`Config` permits additional properties. Real startup rejects `untrusted`
despite that enum. Schema presence never supersedes runtime validation.

## Standalone config/read

The following command ran as its own temporary process, terminated after replies:

```sh
codex -c 'approval_policy="never"' -c 'sandbox_mode="danger-full-access"' app-server --listen stdio://
```

Stdin used `initialize` (client name `permission-qualification`, version `1`,
`experimentalApi:true`), `initialized`, then this normalized request:

```json
{"id":2,"method":"config/read","params":{"includeLayers":true,"cwd":"$SCRATCH"}}
```

Relevant reply fields, extracted with paths normalized:

```json
{"codexHome":"$CODEX_HOME","platformOs":"macos"}
{"approval_policy":"never","sandbox_mode":"danger-full-access"}
{"type":"sessionFlags"}
```

Both permission origins were `sessionFlags`; the user config layer was empty.
The process stopped with exit 0. No thread or model request was sent.

## Classification drift and admitted contract

| Command | Relevant output or result |
| --- | --- |
| `codex --remote-auth-token exec --help` | `error: unexpected argument '--remote-auth-token' found`; exit 2 |
| `codex --remote-auth-token-env EXAMPLE_AUTH_ENV --help` | Root help; exit 0 |
| `codex --no-daemon --help` | Root help; exit 0 |
| `codex --worktree --help` | Root help; exit 0 |

The exact 0.164.0 row uses `permission-grammar-v3`, `YoloSupported=true`, and
maps explicit yolo to `--dangerously-bypass-approvals-and-sandbox` once.
It retains V2's closed config/conflict subset. Hidden `--yolo`/`--not-so-yolo`
and compact `-a...`/`-s...` caller forms refuse as unqualified. Known explicit
sandbox/approval selectors conflict with mapped yolo; unknown values refuse.
No rejected argument is moved after `--`. The 0.153.2 row remains V2;
unknown neighbouring releases have no qualified mapping.

Classification recognizes `exec`/`e` only at established command positions.
The removed token option and unmodeled new options remain indeterminate.
Stored-policy inspection reports explicit `never`, workspace-write and full
access as relaxations; restrictive, missing and malformed controls remain
distinct. It reads bounded sources, not native precedence or managed policy;
legacy values recognized by the inspector do not prove successful startup.

Operator-selected unrestricted/no-approval access means explicit yolo.
Native or absent selection contributes no permission flag and makes no access
guarantee. Neither failures, roles nor stored settings imply yolo. These
observations do not qualify consumer activation, network access or model execution.
