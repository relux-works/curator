# TASK-260916-rkrphg: launcher-accept-claude-codex-aliases

## Description
curator-agent-launcher: curator-run <env-id> accepts claude and codex as aliases of claude_code and codex_cli, normalized before the environment lookup and before env resolve is invoked; origin/provenance lines print the canonical id; README SPEC flag table and help updated; goldens/tests for both spellings. Spec: curator-spec main 2d6d497.

## Scope
(define task scope)

## Acceptance Criteria
curator-run claude and curator-run codex behave exactly as curator-run claude_code / codex_cli; goldens green; help lists aliases.
