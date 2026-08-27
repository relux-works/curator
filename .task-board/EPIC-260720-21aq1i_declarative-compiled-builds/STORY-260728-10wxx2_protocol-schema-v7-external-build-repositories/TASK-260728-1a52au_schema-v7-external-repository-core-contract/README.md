# TASK-260728-1a52au: schema-v7-external-repository-core-contract

## Description
Translate the accepted external-build-repository architecture into curator-spec normative decision and core security text. Define schema-7 versioning, first-class build_repositories, declared/effective source identity, immutable object-format plus full-commit locks, optional exact-tag assertion, repository descriptor trust boundary, audit equivalence, failure taxonomy, signing ownership, and closed future-driver admission.

## Scope
curator-spec decision records and normative protocol/security prose only. Use architecture-v6 SHA-256 2abae77d80eba6789f9911db7e9722595b4f21ba47391ca9eafd0064af03d67e and its accepted review as binding input. Do not implement schemas, generated vectors, or manager code in this task.

## Acceptance Criteria
Normative text uses MUST/MUST NOT language for every accepted boundary; schemas 1-6, go-v1, receipt v1, marker v1/v2, and rc.4 are explicitly frozen; exact source access, protected offline snapshot, audit-before-cache/compiler, manager-derived command/output, credential and signing ownership, monorepo target selection, typed failures, and future closed-driver rules are complete and internally consistent; protocol validation and documentation checks pass; an outcome resource maps every architecture-v6 section to the resulting spec text.
