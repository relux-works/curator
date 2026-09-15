# Skillfile source and selection examples

Date: 2026-09-10. Follow-up research resource for TASK-260910-16vtxi.
Status: design examples only. Version 2 and its field names are provisional;
the existing CLI only supports Skillfile version 1. No schema was modified.

## Existing v1 declaration

```json
{
  "schema_version": 1,
  "skills": [
    { "name": "review", "git": "https://example.org/review.git", "tag": "v1.2.0" }
  ]
}
```

The current optional source field has a different meaning from the proposed
path field: it selects a portable relative path below the manager's configured
source root. With no source field, the implementation defaults to the skill
name. Git is the clone source when that repository is absent. An existing
configured-source declaration can therefore also look like this:

```json
{
  "schema_version": 1,
  "skills": [
    { "name": "review", "source": "team/review", "tag": "v1.2.0" }
  ]
}
```

References: schemas/v1/skillfile-v1.schema.json, protocol/core.md section 5,
and ../curator/internal/manifest/manifest.go. These v1 semantics must remain
unchanged when read by a future CLI.

## Common source-relative layout

Assume the selected source root contains:

```text
<source-root>/
└── skills/
    ├── review/SKILL.md
    ├── docs/SKILL.md
    └── release/SKILL.md
```

For a path source, source-root is the declared directory. For a Git source,
it is the repository root at the resolved commit. The alias toolkit below
has no built-in semantics and need not change when the acquisition changes.

## One skill from a relative local source

```json
{
  "schema_version": 2,
  "sources": {
    "toolkit": { "path": "./agents" }
  },
  "skills": [
    { "name": "review", "from": "toolkit", "directory": "skills/review" }
  ]
}
```

The authored file is <Skillfile-directory>/agents/skills/review/SKILL.md.
It can be uncommitted and the directory need not be a Git repository.
The authored `agents/` directory is distinct from the managed `.agents/`
installation directory. See `local-source-runtime-boundaries.md` for physical
overlap guards and the full runtime/build contract for local acquisition.

## Source replacements with the same skill selector

The following snippets replace sources.toolkit, not the whole Skillfile.

Project-relative directory:

```json
{ "path": "./agents" }
```

Sibling checkout or non-Git directory, resolved relative to Skillfile.json:

```json
{ "path": "../shared-agents" }
```

Absolute host directory:

```json
{ "path": "/work/shared-agents" }
```

Git source at a tag:

```json
{ "git": "https://example.org/kit.git", "tag": "v1.2.0" }
```

The same Git source over SSH, using the existing protocol's SCP form:

```json
{ "git": "git@example.org:kit.git", "tag": "v1.2.0" }
```

Authentication is machine-owned; the declaration does not embed credentials.
The acquisition lane determines which operator Git/SSH settings and credential
providers it may use. In particular, external builds retain the closed
transport and credential-broker contract in profiles/manager.md section 11;
arbitrary inherited configuration is not allowed. HTTPS versus SSH changes
acquisition transport, not the single-skill/collection selector. The transport
grammar is grounded in protocol/core.md section 6.1; the new source object
remains a proposal. Automatic per-machine transport resolution is a separate
proposal tracked as TASK-260910-3du5nd.

Git source at an exact illustrative commit:

```json
{ "git": "https://example.org/kit.git", "revision": "0123456789abcdef0123456789abcdef01234567" }
```

Git source at a project-declared branch:

```json
{ "git": "https://example.org/kit.git", "branch": "main" }
```

The commit above is a placeholder, not a fetched or verified revision. Branch
selection is proposed for root project declarations, preserving the existing
project/dependency distinction: this does not permit floating branches in
published transitive skill dependencies. Tag/branch resolution is recorded as
an exact commit in the proposed project lock and changes only on explicit
resolution/update. Status and launch do not advance it implicitly.

A source is exactly path or git. Git has exactly one of tag/branch/revision;
path has no Git ref. A path source snapshots the selected filesystem bytes,
including local modifications, even if .git exists. It does not mean committed
HEAD. The meaning of a local path cannot change implicitly based on .git.

## Selection replacements with any of those sources

These snippets are entries of skills, not complete Skillfiles. The sources
table above applies unchanged. They are alternatives; do not paste overlapping
selections into one required set unless the specification defines deduplication.

One explicit skill:

```json
{ "name": "review", "from": "toolkit", "directory": "skills/review" }
```

Two selected immediate child directories:

```json
{ "from": "toolkit", "directory": "skills", "include": ["review", "docs"] }
```

All immediate child skill directories:

```json
{ "from": "toolkit", "directory": "skills", "include": ["*"] }
```

All except release (proposed explicit exclusion field):

```json
{ "from": "toolkit", "directory": "skills", "include": ["*"], "exclude": ["release"] }
```

Include/exclude select child directory names. Actual skill identity comes from
validated skill metadata and must agree with the declared single-skill name.
No implicit renaming is introduced. Every selected skill retains its own
dependencies, audit and installation. Validate the full selected set before
publishing it; missing/malformed selections and conflicting identities fail.
Membership and member hashes are recorded in the lock. Initial wildcard
semantics are immediate children, not recursive ** discovery.

## One skill at a repository root

Assume review.git contains SKILL.md directly at the root:

```json
{
  "schema_version": 2,
  "sources": {
    "review-repo": { "git": "https://example.org/review.git", "tag": "v1.2.0" }
  },
  "skills": [
    { "name": "review", "from": "review-repo", "directory": "." }
  ]
}
```

The proposed directory selector explicitly admits . as the selected source
root. This requires its own versioned validation rule, not blind reuse of the
old portablePath definition. The same selector works when a path source points
directly at an authored skill directory.

## Skills nested in a larger repository

If the Git repository root contains agents/skills/review rather than
skills/review, use that real source-relative path:

```json
{ "name": "review", "from": "toolkit", "directory": "agents/skills/review" }
```

To select all skill children in that repository:

```json
{ "from": "toolkit", "directory": "agents/skills", "include": ["*"] }
```

Directory is always relative to the selected source root. It cannot escape
that root using ../ or symlinks. An operator-declared source path may itself
point outside the project; that is an explicit source binding, not a selector
escape. Downloaded metadata does not gain permission to bind host paths.

## Mixed legacy, local and Git sources

```json
{
  "schema_version": 2,
  "sources": {
    "project": { "path": "./agents" },
    "team": { "git": "https://example.org/kit.git", "tag": "v1.2.0" }
  },
  "skills": [
    { "name": "security-check", "git": "https://example.org/security-check.git", "tag": "v3.0.0" },
    { "from": "project", "directory": "skills", "include": ["docs", "release"] },
    { "name": "review", "from": "team", "directory": "skills/review" }
  ]
}
```

The intended result is four distinct skill identities: security-check, docs,
release and review. The old single-skill form remains convenient for isolated
repositories; optional source aliases avoid repeating shared acquisition data.
Old and new forms must be disjoint and cannot mix git/tag with from/directory
inside the same entry.

## One repository at different refs

Refs belong to a source alias. If two selected skills need different refs,
declare two aliases for the same repository with different refs:

```json
{
  "schema_version": 2,
  "sources": {
    "stable": { "git": "https://example.org/kit.git", "tag": "v1.2.0" },
    "next": { "git": "https://example.org/kit.git", "tag": "v2.0.0" }
  },
  "skills": [
    { "name": "review", "from": "stable", "directory": "skills/review" },
    { "name": "docs", "from": "next", "directory": "skills/docs" }
  ]
}
```

This selects distinct skill names. It does not define simultaneous installation
of two versions of the same skill under one name. Conflicting identities must
be reported rather than resolved by the last source alias in the file.

## Compatibility and scope

- Existing v1 files remain readable without modification by the proposed new CLI.
- Existing CLIs explicitly reject version 2/new fields.
- Version 2 examples are syntactic design proposals, not CLI integration tests.
- Collection selection does not require rewriting SKILL.md or agent-skill.json.
- Private source alias overrides, HTTP/plugin imports and skill-owned typed
  dependencies need separate contracts; no syntax for them is invented here.

Validation: all JSON examples were parsed for JSON syntax. Complete v2 examples
were checked for defined source aliases and disjoint declared selection forms.
No remote URL, commit, directory contents or future-schema conformance was
claimed to have been verified.
