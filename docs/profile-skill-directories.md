# Profile skill directories

CIP-0021 (curator-spec PR #162, commit
`e41f96387287f5e76604627f9e63662f7d6d86d4`) permits Git skill requirements
in context manifest schema 2 to select a package within a repository:

```json
{
  "schema_version": 2,
  "name": "sample-profile",
  "version": "1.0.0",
  "requires": {
    "skills": {
      "formatter": {
        "git": "https://github.com/example/skill-packages.git",
        "tag": "v1.0.0",
        "directory": "skills/formatter"
      }
    }
  }
}
```

Omit `directory` for repository-root selection. All requirers must agree
on the canonical source and directory, including root omission. Disagreement
reports `context_range_conflict` with each requirer's source and directory.
A selector must be a portable relative path: no absolute paths, dot/parent
components, empty components or symlink traversal. The directory must exist.
Snapshot violations report `profile_source_invalid`; missing required skill
content reports `manifest_invalid`. A manifestless context skill remains valid.
With an explicit directory, `SKILL.md` frontmatter uses the source-package
reader, including its name, description and triggers; its name must match the
required skill name before any profile or store entry is published.

Schema 1 still rejects skill `directory` as `context_manifest_invalid`.
Schema 2 retains closed fields and admits only absent or empty
`codex_settings`; it does not admit machine policy. Successor locks use schema
4/hash version 2, bind the directory and effective precedence, and retain
required empty settings and contributions. Lock schemas 1–2 remain readable
and reject skill directories. Draft lock schema 3 remains unsupported.

Audit content identity, manifest capabilities and findings use the selected
package tree. For skills, Markdown guidance participates in network-declaration detection only within
agent-facing context (core section 3), using the shared context roots and
exclusions, validated runtime/build roots, and command-less scripts rule.
Non-Markdown scripts and the skill manifest retain their prior scan surface.
Markdown enumeration uses the same filesystem walk as context copying, including
canonical root aliases on case-insensitive filesystems. Descendant exclusions
retain case-sensitive protocol comparison.
Context members instead audit every declared module through the same reader
used for materialization, including root and system classes, all environment
selectors, and non-Markdown extensions. Skill exclusions do not apply to those
modules. Context and MCP source audits retain implicit-none capabilities. MCP
publishes server configuration rather than Markdown and retains the historical
all-Markdown scan. Unknown kinds and invalid context manifests conservatively
use that all-Markdown scan as well. Verdict equality also binds the subject kind.
Detector ruleset 8 uses a new verdict cache namespace: ruleset 7 could cache
allowances for context modules, ruleset 6 could cache
allowances for root case aliases, ruleset 5 could cache
false refusals from repository-only Markdown, while rulesets 1–4 (including
ruleset 4, which did not inspect Markdown) could cache stale allowances. Those
records are cache misses and the package is scanned again. Old
records remain untouched; fresh verdicts bind the current ruleset. Root and
sibling files
cannot lend capabilities or contribute findings to the selected skill.
Successor installation audits the whole closure before publishing store
entries. Managed-home and native skill links now use a shared core section 3.1
projection of the selected package, never its raw repository tree. The same
`whitelist.WalkContext` selects audit Markdown and projection files;
`whitelist.CopyContext` publishes that projection. Link and copy-fallback trees
have identical paths and bytes. Repository-only files, fixed exclusions at any
depth, runtime/build roots and exported scripts are absent from installed
context; command-less scripts remain eligible.

Projections live under `contexts/skill-context/core-3.1-v1/v<hash-version>/<digest>`.
Planning and status only read files; materialization publishes a complete
staged projection atomically and reuses its content-addressed store directory.
Skills surface identities include `core-3.1-v1`, so old raw-root links and old
markers become stale even when their file bytes happen to match. Native and
managed repair replace Curator-recorded links without takeover. Old recorded
managed tree copies are replaced as complete recoverable transaction entries,
removing excluded files while retaining their copy-fallback record. Missing
projection stores and regular projections with missing, extra or changed files
are stale. Explicit repair stages one rebuild from the pinned store entry,
verifies its identity and replaces the complete directory through the existing
recoverable transaction. A staged identity mismatch returns
`ErrSkillProjectionIdentity` without replacing the old directory; linked store
routes remain refused. Shared projections are rebuilt once for all consumers.
On macOS and Linux, repair of an existing projection atomically exchanges the
published directory and a freshly staged sibling. Readers resolving the shared
link see the old or new tree without a missing-directory interval. The journal
retains the old tree until commit cleanup and restores it by exchange on rollback.
Published bytes, private permissions and link targets do not change. Windows
and other platforms retain the existing recoverable backup/install sequence,
including its brief missing-directory interval; repair adds no new refusal there.
Identity migration also publishes projections before activating new links or fallback copies.

The command publishers in `install/targets.go` and `runtimestore` continue to
read snapshots and write independent runtime stores/build caches. No command
dispatch path reads the profile skill link. Build inputs stay out of installed
context and commit-keyed command runtime trees; generated artifacts stay in
build caches. Legacy global/project context already uses `CopyContext` through
`install.stageNode`; adapter discovery continues to publish that installed
canonical tree. Profile-global operations share the profile publishers above.

Projection planning and tree hashing retain the legacy refusal of NUL bytes
before any v1 hash computation; v2 keeps its byte-preserving framing.

Audit ruleset 8 remains unchanged: this publication revision changes neither
the audited bytes nor verdict cache identity. Publication identity is kept in
projection paths and surface hashes, outside the audit cache key.
Legacy global migration remains repository-root addressed and never infers a
directory from layout.

## Verification

The exact spec corpus is vendored at
`internal/contextpkg/testdata/profile-skill-directories-v1/`. The structural
suite drives 56 published cases through `contextpkg.ParseManifest` and
`contextlock.Parse`; the identity suite checks two independent canonical locks.
The lifecycle suite substitutes real fixture Git commits for symbolic pins
and drives install, update, sync, migration, managed resolve and in-place use
with temporary homes. The vendored corpus remains unchanged; its historical
undeclared-contact witness is relocated from package-root `notes.md` to
`references/notes.md` when materializing the lifecycle case, so it exercises
agent-facing context under core section 3. A declared-host positive control
uses the same visible path. Rejections assert diagnostics and preservation of
profile files and surfaces. The narrowing harness uses Go overlays, including
an inside-snapshot symlink control whose target contains valid skill content.
Additional metadata regression cases pair valid frontmatter with malformed,
missing, mismatched-name, empty-description and invalid-trigger refusals.

```sh
go test ./internal/contextpkg -run TestSkillDirectory -count=1
go test ./internal/envprofile -run TestSkillDirectory -count=1
```

The broader CI entry points and artifact locations remain documented in
[ci-gates.md](ci-gates.md). Tests above print their results to stdout; redirect
local evidence under `.temp/` when needed. No operator home is used.
