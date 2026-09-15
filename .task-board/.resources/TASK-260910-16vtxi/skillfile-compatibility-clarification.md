# Skillfile compatibility clarification

Date: 2026-09-10. Follow-up to TASK-260910-16vtxi.
Status: proposed design clarification; no normative schema change accepted.

The existing project format is skillfile-v1.schema.json: schema_version 1,
skills array, name per declaration and exactly one tag/branch/revision.
The current source field is a relative path under the manager's configured
source root; it is not the proposed arbitrary project-local path field.
References: protocol/core.md section 5; schemas/v1/skillfile-v1.schema.json;
../curator/internal/manifest/manifest.go.

The proposed next Skillfile version keeps the individual skill as the unit of
dependency resolution, audit, command activation and installation. Sources are
acquisition declarations, not a replacement package type. A collection expands
to individually identified skills before the normal required closure is built.

Refinement for continuity: make the sources table optional. Keep the existing
single-skill declaration form readable in the new project format, with its
existing semantics. Add disjoint forms for a named skill selected from a
source alias and a collection selected from a source alias. Reject ambiguous
mixes of legacy source/ref fields and the new from/directory selector fields.

Illustrative next-version example (version number provisional):

```json
{
  "schema_version": 2,
  "sources": {
    "local": { "path": "./agents" }
  },
  "skills": [
    { "name": "review", "git": "https://example.org/review.git", "tag": "v1.2.0" },
    { "from": "local", "directory": "skills", "include": ["*"] }
  ]
}
```

Selecting one skill from that same source instead of the collection would use
`{"name":"docs","from":"local","directory":"skills/docs"}`.
For Git aliases the ref moves to the source declaration and applies to each
selected member. Collection expansion must detect duplicate names, validate
each selected SKILL.md and freeze the member set and content identities.

Compatibility recommendation:

- New managers continue reading v1 unchanged; no automatic on-disk migration.
- Next-version declarations normalize to the shared resolved-skill model,
  while preserving the legacy source-root/ref semantics for legacy entries.
- Existing managers reject the unsupported schema/new fields rather than
  silently ignoring a required collection or context dependency.
- v1 retains its original expressiveness; collections, project-relative or
  absolute filesystem snapshots and standalone context declarations need the
  new version.
- Skillfile versioning is independent of agent-skill.json versioning. Adding
  collection selection alone does not require rewriting each skill manifest.
  Extending a skill's own dependencies to rules/knowledge is separate schema
  work. SKILL.md remains the authored agent-facing content.

This clarification is a refinement of the earlier source-alias examples: those
examples demonstrated collections and therefore omitted a per-entry name.
They did not intend to remove named individual skills or require all existing
single-skill repositories to be repackaged.
