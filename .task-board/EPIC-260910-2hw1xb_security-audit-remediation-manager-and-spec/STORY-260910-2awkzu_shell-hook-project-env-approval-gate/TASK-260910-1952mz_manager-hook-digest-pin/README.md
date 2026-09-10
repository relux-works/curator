# TASK-260910-1952mz: manager-hook-digest-pin

## Description
curator: record the digest of every written env file (env.sh/env.ps1) in the project trust surface; generate the hook so it verifies the digest (or manager ownership) before sourcing and warns + skips on mismatch.

## Scope
(define task scope)

## Acceptance Criteria
Hook refuses foreign .agents/env.sh bytes; tests simulate a hostile project checkout
