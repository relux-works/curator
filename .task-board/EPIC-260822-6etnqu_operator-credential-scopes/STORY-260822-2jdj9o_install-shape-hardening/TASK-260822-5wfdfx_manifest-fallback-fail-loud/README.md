# TASK-260822-5wfdfx: manifest-fallback-fail-loud

## Description
Prove that a present canonical manifest (agent-skill.json) that is malformed or declares a newer schema_version fails loud and never falls back to agents/runtime.json (internal/skillspec/parse.go). Add the regression test: canonical manifest with schema_version 99 plus a runtime.json alongside must error with the upgrade hint, not parse the fallback.

## Scope
(define task scope)

## Acceptance Criteria
Regression test in internal/skillspec; fallback reachable only when no manifest file exists at all; go test green.
