# TASK-261018-adogt7: implement-cip-0021-skill-directory

## Description
Parse, resolve, lock, audit and materialize skill directory per CIP-0021; schema 1 still refuses; snapshot discipline; migration unchanged.

## Scope
internal/contextpkg, contextresolve, contextlock, envprofile materialization and audit, tests, CHANGELOG, docs note. Root README untouched.

## Acceptance Criteria
Items 1-9 of gw-SKILLDIR-brief.md hold against CIP-0021 (curator-spec PR #162): schema 2 directory on skills, schema 1 refusal, agreement, lock 4, audit and materialization below the directory, escape and symlink refusals, migration unchanged, vector-driven tests with narrowing mutants, full gate green.
