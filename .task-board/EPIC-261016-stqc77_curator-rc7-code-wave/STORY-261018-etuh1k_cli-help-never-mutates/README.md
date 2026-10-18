# STORY-261018-etuh1k: cli-help-never-mutates

## Description
Field report (curator #cli-help-mutation-261017): `curator init -h` dispatches initialization and creates Skillfile.json and .gitignore instead of printing help. A help request must never mutate a project or home.

## Scope
(define story scope)

## Acceptance Criteria
Help flags on every subcommand print help and write nothing; regression tests cover every mutating subcommand.
