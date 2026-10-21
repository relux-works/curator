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
entries. Managed-home and in-place skill links both use the selected package
root. Existing in-place destinations are unchanged by this amendment.
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
