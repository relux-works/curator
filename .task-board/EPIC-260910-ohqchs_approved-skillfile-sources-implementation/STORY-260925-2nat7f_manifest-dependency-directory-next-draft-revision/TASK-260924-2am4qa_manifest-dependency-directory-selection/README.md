# TASK-260924-2am4qa: manifest-dependency-directory-selection

## Description
curator-spec amendment to core section 4.4: a skill manifest dependencies.skills entry MAY carry directory (a portable relative subfolder of the dependency repository at the pinned ref, same grammar and containment rules as the Skillfile schema 2 individual selector directory in protocol/skillfile-sources.md), selecting the skill in that subfolder; absent directory keeps today's meaning (repository root). Update the skill-manifest schema(s), closure/identity rules (the selected package identity includes the directory), lock and audit implications, conformance vectors (valid subfolder dependency, invalid escapes/absolute/empty/backslash, missing folder, folder without SKILL.md, diamond of two dependencies selecting different folders of one repo), CHANGELOG.

## Scope
(define task scope)

## Acceptance Criteria
normative text in core 4.4 (+ cross-references in skillfile-sources and manager); schema + vectors; validate recipe lines + regenerate-check green; schema-1 / existing dependency meaning unchanged
