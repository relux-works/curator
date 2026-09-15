# Curator: local sources, directive context, MCP, and binary delivery

Date: 2026-09-10. Task: TASK-260910-nys3f2, under STORY-260908-2haegq.
Status: research and discussion proposal; no normative change is accepted.

## Scope and evidence

The request prioritizes ordinary project-local sources and multi-skill directory
installation, then rules/knowledge and public/private project instructions, then
MCP configuration acquisition/materialization. Prebuilt CLI distribution is an
open adjacent question. The intended result of this task is a discussion-ready
proposal, not feature implementation or a silently accepted protocol revision.

Specification baseline: curator-spec main and origin/main both
`d019f0e7179520b5c8dcde321c4fe51e04552f58`, verified after fetching origin.
Reference implementation source inspected at
`683364ce233df872d6cbb194e0e6b205127f5bca` in the adjacent curator checkout.
The six inspected implementation files had no working-tree delta: README.md,
install.sh, internal/manifest/manifest.go, internal/contextmaterialize/mcp.go,
internal/envprofile/managed.go, and internal/adapters/adapters.go. Other existing
changes in that checkout were not curated or modified. Implementation statements
below are source observations, not fresh end-to-end runtime verification.
Vendor behavior below is documentation evidence checked on 2026-09-10, not a
claim that all installed agent versions were tested.

Local authoritative references, relative to curator-spec:

- protocol/core.md sections 3, 4.4, 5, 6, 7, 9, 10, and 12.
- protocol/environments.md sections 1-3, 5, 6, 7, 8, 9, and 10.
- profiles/manager.md sections 4-6 and 12.
- schemas/v1/skillfile-v1.schema.json, skillfile-dev-v2.schema.json,
  agent-context-v1.schema.json, agent-mcp-v1.schema.json, context-lock-v1.schema.json.
- decisions/0012-context-packages-and-semver-locks.md, including its erratum,
  and decisions/0013-execution-ownership-and-launch-plans.md.

The normative environments text is the current authority. Historical ADR status
labels still say proposed; those labels do not undo the subsequently landed text.

## What already exists and what is missing

| Request | Current contract | Actual gap |
| --- | --- | --- |
| Project-local skills | Core section 5 has a source path below the manager's source root, exact git refs, and local Git development substitutions. | Arbitrary relative/absolute project paths, non-Git and uncommitted source trees as ordinary sources, without a dev-substitution exception. |
| Several skills from one directory | A skill may be a directory within a snapshot, but project entries are individually named and ref-bound. | One declaration selecting multiple skill roots, deterministic expansion, shared acquisition, exact membership recording. |
| Local contexts | Environments section 1 supports absolute/project-relative path roots; section 6 allows path overlays. Copies become immutable state-hash snapshots. | Project Skillfile composition and local dependency edges. Package requires remain git-only; path is not a general dependency source. |
| Rules and knowledge | Context modules have class root/system, environment selectors, ordering and weights. Skills can already carry references. | Distinct semantic types, activation semantics, independent shared knowledge, project scope and native rule adapters. |
| Managed AGENTS.md / CLAUDE.md | Global profiles already compose and materialize root context, with ownership markers, backups and drift detection. | Project instruction ownership, public/private composition, optional generated uncommitted outputs and launch-time behavior. |
| MCP wiring | Environments sections 2.2, 5.8 and 7.8 define agent-mcp.json, managed-home config files and launch channels for Claude Code, Codex and OpenCode. | Endpoint/registry/plugin acquisition, local MCP dependency packages, richer runtime/auth bindings, and project/native-home writing. |
| Prebuilt Curator executable | Adjacent curator/install.sh downloads release archives and checks checksums; README documents binary/package distribution. | Do not mistake manager bootstrap for a missing protocol feature. Prebuilt executables supplied by skills are a separate new acquisition contract. |

MCP materialization is also present in implementation source:
internal/contextmaterialize/mcp.go renders the three formats, and
internal/envprofile/managed.go calls MCPFile. This establishes implementation
presence, not proof of every real launch or authentication flow.

## Ecosystem findings

| Environment | Authored directives | Knowledge and portability consequence |
| --- | --- | --- |
| Claude Code | CLAUDE.md, CLAUDE.local.md, .claude/rules/*.md; rules may be scoped by paths. | Authored context and agent-written auto memory are separate. A private local file is additive. [Memory documentation](https://code.claude.com/docs/en/memory) |
| Codex | AGENTS.md and AGENTS.override.md, with directory hierarchy and configurable fallback names. | Override replaces the base file at that directory; it is not an automatic private append mechanism. [Official OpenAI documentation](https://learn.chatgpt.com/docs/agent-configuration/agents-md) |
| Cursor | .cursor/rules/*.mdc with always, file, relevance and manual activation; AGENTS.md also supported. | Rules and skills are distinct host surfaces. [Rules](https://cursor.com/docs/rules) |
| OpenCode | AGENTS.md plus explicitly configured instruction files/globs/URLs. | A document reference alone is not an automatic import. [Rules](https://opencode.ai/docs/rules/) |
| Windsurf/Cascade documentation | The current page redirects to Devin Desktop: .devin/rules is preferred, with .windsurf/rules fallback; AGENTS.md and local memories are documented. | Persisted agent memories are not equivalent to authored portable knowledge packages. This illustrates adapter-version drift. [Memories and rules](https://docs.devin.ai/desktop/cascade/memories) |

Rules are clearly a recurring ecosystem concept. This sample does not establish
a shared KNOWLEDGE.md filename or interoperable knowledge-package schema.
Agent Skills instead specifies optional references and progressive disclosure.
Curator can define an authored knowledge component without claiming it is a
universally discovered native file. Agent-created memory should remain separate
from immutable installed inputs. [Agent Skills specification](https://agentskills.io/specification)

One concrete adapter discrepancy deserves its own bounded review: manager
section 5 and internal/adapters/adapters.go map Cursor skill installations to
.cursor/rules, while current vendor docs discover skills through
.cursor/skills or .agents/skills and require .mdc for native project rules.
This is a documentation/source mismatch; no installed-Cursor repro was run.
[Cursor skill directories](https://cursor.com/docs/skills)

## Recommended model

Keep source acquisition, component meaning and installation scope independent:

- Source: a declared git repository or local directory; later, a typed metadata
  endpoint/import format. Source location does not determine loading semantics.
- Components: skills, authored rules, reference knowledge and MCP declarations.
  Several components can share one source and snapshot. A collection is a
  selection convenience, not a new execution runtime.
- Scope/output: project or global profile, optionally modified by operator-local
  input; native adapter surfaces are generated from the resolved desired state.

Reuse the existing resolver, snapshots, locks, audit, deterministic composition,
markers, rollback and launcher boundary. Add typed context components within the
context-package machinery; avoid a separate package manager for rules and another
for knowledge. AGENTS.md/CLAUDE.md are adapter outputs, not additional dependency
package kinds. Imported plugins are a packaging input.

## Priority 1: local sources and collections

Conceptual declaration fragment, deliberately without a schema version. This is
not valid current Skillfile syntax and reserves no future wire identifier:

```json
{
  "sources": {
    "project": { "path": "./agents" },
    "shared": {
      "git": "https://github.com/example/agent-toolkit.git",
      "tag": "v1.2.0"
    }
  },
  "skills": [
    { "from": "project", "directory": "skills", "include": ["*"] },
    { "from": "shared", "directory": "skills", "include": ["review", "docs"] }
  ]
}
```

The local tree can simply be agents/skills/review/SKILL.md and
agents/skills/docs/SKILL.md. It needs neither a separate repository nor an
umbrella package manifest. The standard SKILL.md validation still applies;
agent-skill.json is needed for Curator-specific runtime/dependency metadata, not
merely because a directory contains multiple ordinary skills.

Recommended initial semantics:

1. Resolve root-declared relative paths against the directory of Skillfile.json,
   never the shell's transient cwd. Permit absolute paths. Keep machine-specific
   source overrides private when portability matters; absolute-path declarations
   remain usable rather than being categorically forbidden.
2. Keep path and git distinct. Path consumes the selected live filesystem bytes,
   including uncommitted/new files, into an immutable snapshot. Git consumes an
   exact commit. Never silently choose HEAD merely because a path contains .git.
3. A declaration with include is a collection of immediate child skill roots;
   one without it can address an individual package root. Start with literal
   directory names and the wildcard *, plus explicit exclusions. Recursive
   discovery needs a separately specified depth/selection contract.
4. Validate each selected skill; take its declared identity from SKILL.md and
   reject name/path inconsistencies, conflicting duplicate names and malformed
   selected packages. A named member that is missing or a collection matching
   no skills should fail with a specific diagnostic. Do not silently install
   only a successful subset.
5. Expand and sort the selection before audit/publication. Record source,
   package-relative directory, member names and hashes. Install/update/explicit
   local refresh changes that record; read-only resolve/status never does.
6. Reuse content hashes for path pins, not pretend Git revisions. Reconcile this
   with currently commit-keyed skill stores/markers explicitly. A hash proves
   exact bytes, not that another machine can retrieve an absolute path.
7. Keep authoring under agents/ and generated installation under .agents/.
   Snapshot selected roots; exclude administration/generated outputs. Reject
   output/input overlap, symlink escapes and other existing snapshot violations.
8. A repository-declared dependency can address another package within the same
   declared source root. It cannot interpret an absolute path from downloaded
   package metadata as permission to read the operator's filesystem. Host paths
   are bound by the root Skillfile or private operator configuration.

The same selector should work for local and Git sources. Multiple skill names
at one commit remain separate components; do not require one repository per name.

A project lock is the right companion, already identified as Decision 0012 open
question 6. Agree its location, refresh and private-override behavior before
adding ranges. P1 need not wait for a general semver resolver: exact Git pins,
path hashes and frozen collection membership are sufficient initially. Ordinary
sync can be the explicit local refresh operation; a launch must not rescan and
silently add a newly created skill. Live symlink mode can be a later explicit
development mode with distinct guarantees.

## Priority 2: directive and knowledge semantics

Extend context packages and project declarations together. Example project
fragment using the same source aliases, again proposed syntax only:

```json
{
  "rules": [
    {
      "from": "project",
      "files": ["rules/*.md"],
      "activation": { "mode": "always" }
    }
  ],
  "knowledge": [
    {
      "from": "project",
      "files": ["knowledge/**/*.md"],
      "activation": { "mode": "on-demand" }
    }
  ]
}
```

Plain files can be declared directly for project convenience. Published reusable
sets can export named typed modules from agent-context.json. Their source roots
can be local or Git using the same containment rules. Stable module identity
should include its owning member and relative path, not only the basename.

The project composition belongs in Skillfile.json. If a skill itself needs a
shared rule or knowledge unit, put that machine-readable dependency in the next
version of agent-skill.json; SKILL.md remains agent-facing prose and references.
Downloaded skill metadata must not create machine-wide overlays or promote its
own context to a higher instruction scope. Dependency activation must be explicit
and supported by the target adapter.

Core section 3.1 does not admit arbitrary rules/ or knowledge/ directories into
existing skill context. Existing references/ can serve skill-local reference
documents now; standalone shared components need the new explicit contract.
Specify installed reference locations so dependency links still work after
copying from the authoring tree.

Distinguish three questions:

| Question | Proposed concept |
| --- | --- |
| What does the component mean? | Skill workflow, rule instruction, knowledge reference. |
| When should it load? | Always, on matching files, or on demand. Agent-selected activation is adapter-specific until a portable contract exists. |
| Through which channel? | Preserve the existing root/system channel distinction; do not overload class with rule/knowledge meaning. |

Existing root/system modules may contain mixed prose; do not silently relabel
all old modules as rules. Add explicit new fields/versioned variants, preserving
old behavior. Unknown fields are rejected today, so adding them to schema 1 is
not an interoperable extension.

Rules can render to a supported native rule surface; unconditional instructions
can use existing root-context composition. Knowledge should materialize as
addressable documents with a compact description/path index. Reading it remains
agent-driven unless the host offers a stronger loading interface. Do not
concatenate all knowledge into every prompt or turn a path-scoped rule into an
always-on rule without a declared, visible fallback. If the adapter cannot
preserve the requested activation, report unsupported semantics.

Keep module bodies opaque. Adapters may generate a metadata wrapper/index, but
that transformation needs a specified form and hash inputs; do not violate the
current prohibition on implicit templating or imports inside module bodies.
Content ordering and weights can reuse existing composition rules. They order
text, not enforce policy or solve semantic contradictions in prose. Knowledge
does not acquire policy priority merely because it has a later output position.

## Managed project instructions: outline only

Recommended future opt-in shape:

- Public authored inputs live in agents/context/ or a context package, committed
  with Skillfile.json and the public project lock.
- An explicitly declared agents.private.md, or a private source override pointing
  outside the checkout, contributes operator-local project input.
- Curator composes project-owned shared packages, project inputs and private
  refinements in a defined order. The global profile stays in its existing
  native scope and is included only once in the effective context; do not copy
  it into project output and load the same global context again.
- AGENTS.md and CLAUDE.md can be generated uncommitted outputs. The composition
  input and output must not be the same file.

Existing tracked or unmanaged AGENTS.md/CLAUDE.md need a reviewed ownership
transition with backup and drift handling. Do not append generated text into an
arbitrary tracked file, silently untrack it or overwrite another agent's edits.
The option to keep public instructions committed should remain available.

Do not implement the private layer just by renaming it AGENTS.override.md: Codex
would replace the base at that directory. Claude's local file is additive, so the
same filename trick is not portable. The composed output must include all
intended inputs exactly once.

Separate a shareable public lock from the local effective state that binds private
inputs. Keep private contents, absolute machine paths and generated composites
out of committed locks, remote audits and public diagnostics. Private here means
uncommitted; loaded instructions are still visible to the chosen agent/model and
should not contain credentials.

Defer automatic per-launch generation until there is a supported project-scope
delivery channel and a concurrency rule. Two profiles launched in the same
checkout must not race by overwriting one AGENTS.md. An ignored file left in the
project works for direct launches after sync; a per-launch immutable artifact
requires an adapter channel that preserves project scope. Adding content to a
global home or a replacement system prompt is not automatically equivalent.

## Priority 3: MCP acquisition, wiring and plugins

Distinguish a metadata endpoint from the running MCP endpoint. A URL accepting
MCP requests is not thereby an installer-configuration URL. MCP standardizes
stdio and Streamable HTTP communication. Server discovery metadata can come from
the MCP Registry's server.json/API or an explicitly declared configuration source.
[MCP transports](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports)
[MCP Registry](https://modelcontextprotocol.io/registry/about)

The Registry points at remote services or external packages; it does not host all
server binaries or make metadata a security approval. Curator's audit registry
and the MCP discovery registry have different roles. Private catalogues can use
an operator-selected endpoint; a public registry need not be a mandatory hop.

Proposed acquisition pipeline:

1. Select declared format: Curator MCP package from local/Git, pinned Registry
   metadata, or a recognized plugin format. For HTTP documents declare the schema
   or format explicitly rather than guessing JSON structure.
2. Fetch and validate inert metadata. Resolve the selected server/version and
   transport, normalize into a Curator-owned descriptor, retain original source
   and digest. Metadata and imported package bytes enter the ordinary audit.
3. Lock the exact declaration and artifact identities. Fetch updates explicitly;
   status/launch uses the locked state, including offline. An ordinary MCP server
   URL remains a runtime service whose behavior is not pinned by a config hash.
4. Bind machine-specific environment/credential references locally. Emit the
   adapter's configuration in its declared owned scope. Show the desired diff and
   use the existing marker/rollback discipline for project/native config writes.
   Preserve unrelated entries and detect concurrent edits to shared files; do
   not replace a whole mutable native configuration from a downloaded template.
5. The agent host launches stdio servers, performs OAuth and connects remotely.
   Curator installation does not run imported scripts, hooks, npx or a server as
   an acquisition side effect. Server-runtime installation is a separate feature.

Current env_names only lists process variable names. It does not express mapping
a variable to an HTTP Authorization header, nor guarantee a client forwards it
to each stdio subprocess. Add explicit name-only per-server runtime bindings and
adapter support where needed; keep actual values in the operator environment or
client credential store. Codex documents bearer_token_env_var and env_vars.
[Official OpenAI MCP documentation](https://learn.chatgpt.com/docs/extend/mcp?surface=cli)

Use the existing managed-home launch channel as the first wiring path. A project
Skillfile MCP declaration must also have a project path that works outside a
managed launch if that is the intended UX. That is a new surface, not a silent
reinterpretation of manager section 6's read-only requirement check. Native home
editing should remain a separate opt-in ownership mode. Pi has no MCP channel in
the currently admitted Curator revision; do not infer one from plugin packaging.

### A real common plugin standard now exists

Agent Plugins 1.0.0 is published. Its portable core contains skills/ and mcp.json
under a root plugin.json; rules/knowledge remain outside that core. It supplies
an interoperability floor while installation/distribution remain client-owned.
[Agent Plugins](https://agent-plugins.org/)

The format has immediate-child skill discovery, plugin-relative MCP commands and
working directories, optional transport support and client-owned authorization.
Its error boundaries permit some components to be skipped; Curator's declared
required closure must not silently inherit that behavior. It also defines no
portable credential-reference field. These differences require an explicit import
contract, not a file rename. [Specification](https://agent-plugins.org/specification)

Cursor supports the common format alongside its own richer plugin format.
OpenAI documents portable packaging alongside the .codex-plugin compatibility
layout. Claude Code documents its own .claude-plugin layout. Do not infer that
all clients accept every vendor-specific extension.
[Cursor plugins](https://cursor.com/docs/plugins)
[Official OpenAI plugin packaging](https://developers.openai.com/plugins/build/plugins)
[Claude Code plugin reference](https://code.claude.com/docs/en/plugins-reference)

Recommendation: make Agent Plugins a recognized import/export packaging format
for compatible skills and MCP components, while Curator remains responsible for
source selection, exact locks, audit and environment composition. Initially scope
imports to representable data. Report unsupported required components, relative
executables, runtime-data needs or auth requirements rather than silently losing
them. A Curator importer is not automatically a conforming plugin runtime.
Do not forward unknown extension directories to hosts where they might activate
hooks outside the reviewed Curator plan.

MCPB is a separate archive format for local MCP server distribution, with a
manifest and bundled server files. Consider it later for local server artifacts;
it does not replace project context composition.
[MCPB](https://github.com/modelcontextprotocol/mcpb)

## Prebuilt CLI distribution

Two meanings need separate decisions:

1. The Curator manager itself: prebuilt download installation already exists in
   the checked-in installer. No claim was made that every currently published
   release asset was downloaded and tested in this research task.
2. CLI executables exported by skills: propose a separately versioned prebuilt
   artifact acquisition mechanism, alongside existing source-build mechanisms.

For the second, bind immutable version/source provenance, target OS/architecture
and ABI requirements, download URL, size/digest, executable path and publisher
verification policy. Reuse protected storage, receipt binding, command shims,
atomic publication and rollback. Do not issue a source-build receipt for a
download, or call a matching checksum proof that the binary matches audited
source. A source audit does not attest the compiled artifact.

Default product direction: prefer an admitted, verified prebuilt for the exact
platform; permit source fallback only when declared and enabled by operator
policy. Missing/unsupported artifacts may trigger that allowed fallback. A
signature or integrity failure should fail rather than silently switch routes.
Local authored skills can declare the same mechanism. Installation should never
execute the downloaded program merely to discover its version.

## Proposed specification work and order

| Slice | Contract changes | Focused acceptance evidence |
| --- | --- | --- |
| P1 | A new project manifest version/capability; shared source selector; directory collections; path skill snapshots; project lock/marker identity. | Relative/absolute/non-Git/dirty sources; multiple and filtered skills; duplicate names; no matches; containment; membership/hash change; reproducible reinstall; no implicit launch refresh. |
| P2a | Versioned typed context modules and activation; matching Skillfile and skill-dependency declarations; project composition; adapter capability matrix. | Always versus lazy knowledge; file-scope preservation or explicit failure; local and Git dependencies; stable ordering; no duplicated context; current Cursor rule/skill discovery. |
| P2b, outline first | Public/private project inputs and owned generated instruction outputs. | Public/private separation, correct append semantics, output drift/takeover, direct versus managed launch, no shared-output race. |
| P3a | Local/Git MCP packages in project scope; typed metadata importer; name-only auth/runtime bindings; existing managed launch and explicit project output. | Exact imported data, representability, idempotent writes, ownership conflict, missing runtime/auth state, zero installer execution. |
| P3b | Agent Plugins import/export profile and optional later native vendor/MCPB adapters. | Multi-skill discovery, unsupported component diagnostics, namespace/version/transport mapping, no implicit hooks. |
| Separate open slice | Prebuilt artifact acquisition and verification profile. | Target selection, trusted digest/signature binding, artifact receipt, rollback and explicit fallback behavior. |

The next design artifact should define the project source/collection contract and
its lock before broadening the directive or MCP surfaces. Coordinate the local
dependency representation across those drafts now, but do not make P1 wait for
the whole roadmap.

After discussion, an accepted change needs normative prose, new strict schemas,
positive/negative conformance vectors, marker/lock migration, manager/CLI text,
compatibility documentation and implementation-consumption evidence together.
Do not widen frozen core 1.0 objects in place or merely sprinkle path/rules fields
into current schema-1 documents. A new versioned capability can preserve old
readers' explicit unsupported-version behavior. Preserve Decision 0013's roles:
Curator resolves context; the launcher composes execution; the agent hosts MCP.

## Validation and research tooling

- git --version, rg --version, python3 --version, task-board --help, gh --version
  established tool readiness; logs are in .temp/curator-context-review/.
- git fetch origin main and git rev-parse HEAD origin/main fixed the spec baseline.
- Bounded rg/sed/source reads established current contract and implementation
  presence. Official web pages supplied external behavior evidence.
- No application builds, test suites, package installs, native config edits,
  signed release operations or implementation changes were required or performed.
- JSON examples are syntax-checked only; current schemas intentionally reject them.
- This report and UNRESOLVED_QUESTIONS.md are persisted as task-board resources.
  The discussion remains open; the research deliverable does not accept a design.
