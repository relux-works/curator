# TASK-260910-1952mz: manager-hook-digest-pin

## Description
curator: record the digest of every written env file (env.sh/env.ps1) in the project trust surface; generate the hook so it verifies the digest (or manager ownership) before sourcing and warns + skips on mismatch.

## Scope
curator internal/shell (hook generation, both rollout profiles, default A-warning), new internal/hookapproval state package, internal/envfiles manager-recorded digests, vector-execution test for conformance/v1/vectors/shell-hook-trust.json with root-content skip and ledger row, CHANGELOG

## Acceptance Criteria
Hook refuses foreign .agents/env.sh bytes; tests simulate a hostile project checkout
