# Profile MCP first-install writer

The separate `profile mcp` entrypoint consumes MCP packages already admitted
by an explicitly named profile lock. It introduces no global or profile
`mcp_servers` grammar and never delivers project declarations into a shared
profile home. Project delivery belongs to the project-view adapter contract.

```sh
curator profile mcp status example
curator profile mcp install example --env claude_code,codex_cli --dry-run
curator profile mcp install example --env claude_code,codex_cli
curator profile mcp repair example --env claude_code
```

Output is JSON in the `profile-mcp-native-writer-v1` namespace, bound to the
profile id and exact lock digest. Each determinable NAME/environment pair
has one sorted row. Dry-run states begin with `would:`. Unsupported adapters
have a null path and no code; a foreign conflict has its diagnostic in the
command summary, rather than a row code. A failed batch converts tentative
`added` rows to refusals and publishes none of that batch.

Only absent entries can be first-installed. Current source/MCP allowlists,
signer policy, revocations, strict audit, package snapshot identity and
passable environment-name ceilings apply. The writer does not fetch, run MCP
commands, inspect native credentials, contact endpoints or attempt login.
A successful native publication reports exactly:
`MCP configuration installed. Authentication not checked.`
An unchanged or status-only operation does not claim new publication.

## Native files and protected ownership

Paths are relative to the profile's managed home, never the operator's
native home:

| Environment | File | Initial writable qualification |
| --- | --- | --- |
| `claude_code` | `.agent-context/mcp/claude_code.json` | stdio, Claude Code 2.1.305, macOS arm64, executable SHA-256 `ea3ea3c08d2a9634a4651e687fcf900d410c489bce2b00f22bd7352b91ba704f` |
| `codex_cli` | `curator-mcp.config.toml` | stdio, Codex CLI 0.162.1, macOS arm64, executable SHA-256 `74c6a6d263f40d65f99c27a815275f5b4ed0bcc2baecc396b4cf3255498bdfc0` |
| `muse` | No certified writable file | Refuses required writes |
| `opencode`, `pi` | No file in this capability | Unsupported |

Executable qualification reads bytes without executing the binary. HTTP,
other builds and other platforms refuse `mcp_contribution_policy_refused`
with `native_writer_unqualified`. Read-only inspection does not need a vendor
binary. These pins come from the [adaptation sheet](https://github.com/relux-works/curator-spec/blob/f02b5aeff07f81ee7d39ca4d2bc9aba1f8e3f816/protocol/mcp-native-writer.md)
and qualify the native surface only. This first-install subset does not
qualify managed launch, native-source exclusion, reload, resume, children,
authentication or ownership actions.

The private manager namespace is `mcp-native-v1/ownership/<K>.json`. `K` is
the lowercase SHA-256 hex of CCJ-1 `[profile_id, environment, relative_path]`.
The closed canonical record binds the namespace/version, profile, environment,
path and sorted per-NAME records with `owner: curator`, normalized server and
package-content hashes, null absent preimage and installed native-entry hash.
The marker is a separate protected input, never a key inside native config.
Malformed, linked, unreadable or insufficiently private ownership refuses.
A reserved native filename alone confers no ownership.

The editor preserves foreign record bytes, order and unrelated settings.
Claude entry digests cover the exact JSON value bytes; TOML entry digests
cover the selected table statements and subtables, excluding trailing
container whitespace. Status uses the same spans. Duplicate keys and
unreadable native syntax refuse. Inline or dotted TOML records whose exact
table span cannot be certified are inspectable for conflicts, but prevent
first install of additional tables. The writer refuses a lossy conversion.
An existing legacy store symlink is not followed or silently enrolled; it
needs a separately qualified migration. The legacy whole-file materializer
refuses enrolled MCP surfaces, including removed declarations.

For Codex, every operation inspects the live managed `config.toml`. A
same-NAME base entry conflicts even if its bytes equal the desired server,
and regardless of whether seed A, B or a later external edit introduced it.
The writer never modifies or recopies that base. Disjoint base servers remain
ungoverned; this writer does not claim a profile-only effective server set.

## Status, unavailable actions and recovery

`status` reports `unchanged`, `drift`, `missing`, `replacement-required`,
`removed-declaration`, conflict, unreadable and unsupported observations.
Markerless same-NAME entries always conflict, including equal bytes.
Install and repair preserve drift and refuse operations requiring authority.
Explicit `restore`, `replace`, `remove`, `release` and `--takeover` refuse
`mcp_native_writer_authority_unavailable`. No option, repository file,
project approval, content pin or passive hash can mint that authority.
The separate platform authority leaf must establish operator custody and an
exact signed act before ownership actions can be implemented.

Native edits, protected ownership, applicable skill context files and the
unchanged exact profile lock are targets of one existing managed-home journal.
Existing skill bytes are reused only when the complete projection and modes
already match; changed or additional skill files require the materializer.
Preflight failures preserve native/ownership payloads. Target full preimages
and no-follow parent/leaf boundaries are rechecked at publication. Read-only
source, policy, live Codex base and executable inputs are durably pinned by
bytes as well as the existing boundary proof, so recovery cannot silently
adopt an in-place policy edit after process restart.

Status and preview never recover or acquire the mutation lock. They report
`recovery_required` and refuse pair inspection if durable transaction state
is pending. Mutating install/repair first recovers already admitted durable
intent under the manager-home lock; `recovered` reports completed prior
recovery separately from `published`. A changed frozen input refuses and
rolls the interrupted payload back under the existing transaction rules.
No launch fragment is exposed by this entrypoint.

## Validation

On Windows, Git-backed profile and project MCP caches obey the
[measured metadata-root bound](windows-git-paths.md). An over-budget acquisition
returns `git_metadata_path_too_long` through `profile_source_invalid` before
creating cache parents; it does not leave a partial clone or change existing
home bytes. A short private metadata layout remains a separate follow-up.

The tests exercise production planning, actual journal publication and fresh
engine recovery in disposable homes. They replace only vendor identity
observation with a private test seam; they make no claim of new vendor runtime
qualification. Each test states its specific claim beside the test.

```sh
env -u GIT_DIR -u GIT_WORK_TREE go test ./internal/mcpnative ./internal/transaction ./internal/config
env -u GIT_DIR -u GIT_WORK_TREE go test ./internal/envprofile -run TestProfileMCP
env -u GIT_DIR -u GIT_WORK_TREE go test ./cmd/curator -run 'Test(ProfileMCPCLI|RegisteredSubcommandHelpNoWrites|CommandHelp)'
```

Use a private HOME, short canonical TMPDIR, private Go caches and a PATH that
excludes ambient native tools. Store validation logs and bounded narrowing
plants under task-scoped `.temp/`; these are disposable artifacts. The test
suite does not read real native authentication state.
