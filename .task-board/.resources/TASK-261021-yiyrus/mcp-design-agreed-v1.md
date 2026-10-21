# MCP servers in the project Skillfile: agreed design v1 (cocoaskills + Curator orchestrators, 2026-10-18)

Status: agreed in collective topic `#curator` › `mcp-skillfile-url`. It is P1 from the cocoaskills orchestrator plus
the Curator orchestrator's amendments A1–A6. It binds the cocoaskills implementation. It is not normative Curator
spec text: Ivan finalizes the spec separately.

## 1. Skillfile schema 3

- A project Skillfile with `"schema_version": 3` is a **strict superset of schema 2**. Every schema-2 member,
  including the source-extension grammar, keeps its meaning. The only addition is the optional top-level member
  `mcp_servers`.
- Schema 3 is **project scope only**. A machine-global Skillfile stays on schema 1 and rejects 3 with the upgrade
  error, as the source extension does for 2.
- A user writes schema 3 only when the Skillfile needs `mcp_servers`. csk's own Skillfile writers never upgrade a
  Skillfile to 3 by themselves.
- Curator releases up to rc.7 reject schema 3 with the upgrade error. That refusal is clean, and the csk docs say so
  until Curator supports schema 3.

## 2. Declaration grammar

`mcp_servers` maps NAME to DECL. NAME follows the M1 grammar `[A-Za-z0-9][A-Za-z0-9_-]{0,127}`. DECL is one of two
forms.

### 2.1 URL shorthand

```json
"mcp_servers": { "docs": "https://mcp.example.com/mcp" }
```

A string is exactly `{"server": {"transport": "http", "url": "<string>"}}`. The URL rules are those of Curator
environments §2.2:
- the scheme is https;
- the host is ASCII;
- there is no userinfo, no query and no fragment.

M1's further bounds apply as well: length, no whitespace, control or format characters, no `$`, braces or
backslashes. The secret-material detector runs over `url` (and over `args` in the object form). A token in a URL is
a blocking finding.

### 2.2 Object form

```json
"mcp_servers": {
  "tools": {
    "server": {
      "transport": "stdio",
      "command": "uvx",
      "args": ["example-mcp==1.2.3"],
      "env_names": ["TOOLS_TOKEN"],
      "environments": ["claude_code", "cursor"]
    },
    "csk": { "auth": { "kind": "env", "help_url": "https://example.com/tokens" } }
  },
  "api": {
    "server": { "transport": "http", "url": "https://api.example.com/mcp" },
    "csk": {
      "headers": { "Authorization": { "env": "API_TOKEN", "prefix": "bearer" } },
      "auth": { "kind": "env" }
    }
  }
}
```

- **`server`** (required) MUST be a valid Curator `agent-mcp.json` `server` object, and nothing more (A1). Its
  member set is closed:
  - `transport`: `stdio` or `http`;
  - `url`, for http;
  - `command` and `args`, for stdio, where `command` is a bare executable name;
  - `env_names`;
  - `environments`.

  The M1 rules apply on top: reserved env names, runtime pins for `npx` and `uvx`, and the URL rules above.
  `environments` lists csk agent ids. Absent means the Skillfile's resolved `agents`; Curator reads absent as every
  adapter. Naming an agent that csk cannot write in slice 1 refuses with `mcp_capability_unsupported`.
- **`csk`** (optional) holds the csk-only members that a Curator `server` cannot express (A2). Curator can later
  ignore or refuse it deliberately. Its members, all optional:
  - `headers` (http only): `{Header-Name: {"env": NAME, "prefix"?: "bearer"}}`, with the M1 header rules;
  - `auth`: `{"kind": "none" | "env" | "oauth", "help_url"?}`. The default is `env` when the declaration has any env
    reference (`env_names` or `headers`), otherwise `none`;
  - `scope`: `project` only; `user` refuses with `mcp_capability_unsupported` in slice 1;
  - `git_policy`: `local` only; `shared` refuses with `mcp_capability_unsupported` in slice 1.
- Any other DECL member refuses with `mcp_declaration_invalid`. Later, not slice 1: `{"package": {"git": …,
  "range" | "tag" | "revision": …, "directory"?}}` resolves a Curator `agent-mcp.json` declaration package. That is
  the convergence point: both managers read one artifact under one allowlist (A6).

Internally csk normalizes both forms to the same model as M1 (`ServerSpec`, `AuthSpec`, agents, scope,
git_policy). Digests are computed over the normalized model, so the shorthand and its expanded object form have
equal digests.

## 3. What `csk install` does (project scope)

It follows `slice1-decisions.md` D1, D3–D8 and A1–A10, unchanged:
- **Consent:** required for stdio or any env reference; a plain URL needs none.
- **Writes:** native project configs only, for Claude Code `.mcp.json`, Codex `.codex/config.toml` and Cursor
  `.cursor/mcp.json`.
- **Edits:** lossless span edits.
- **Ownership:** a per-entry ledger in the csk home; a foreign same-name entry is a conflict.
- **Drift:** skip, restore or release.
- **Removal:** an undeclared owned entry is removed.
- **Transaction:** one transaction for native files, the ledger and the skills.
- **Git:** gitignore for created files.
- **Auth:** none at install, only hints.
- **Processes:** no server process is ever started.

## 4. Reporting is total (A4)

Every (server, agent) pair gets exactly one line, naming the agent and the exact project-relative path, whatever
the outcome:

```text
mcp docs -> claude_code .mcp.json (added)
mcp docs -> codex_cli .codex/config.toml (unchanged)
mcp docs -> cursor .cursor/mcp.json (skipped: drift)
mcp api -> cursor .cursor/mcp.json (conflict: foreign entry)
```

- Success actions: `added`, `updated`, `unchanged`, `restored`, `removed`, `released`, `skipped: drift`.
- A refused install still prints one line per affected pair, with the refusal (`conflict: foreign entry`,
  `refused: <code>`), then the refusal summary. Nothing is written.
- `--dry-run` prefixes every line with `would:`.
- After a successful write: `MCP configuration installed. Authentication not checked.` Then per-agent hints, printed
  and never executed.

## 5. Managed launches (A5)

Curator-managed sessions do not yet see the project configs that csk writes. Claude suppresses other MCP sources
when its managed channel exists, and Codex loads a project config only for trusted projects. The csk docs state
that csk targets native agent sessions. The managed path arrives with Curator's project-contribution work.
