# TASK-260924-10d3l1: drive-exact-repository-and-commit-mismatches-through-install

## Description
Add source_identity-only and commit-only bad-record rows to TestDraftEvidenceExactMatch at install.Project (gap matrix TASK-260924-3re9jo, BUG-260920-2wyzde B1); no ResolveExact behaviour change.

## Scope
(define task scope)

## Acceptance Criteria
exact positive evidence installs; wrong repository only and wrong commit only each fail closed at install.Project, preserve prior state, expose no registry endpoint/key; narrowing mutants of either comparison killed
