# curator CLI: accept claude/codex as aliases of claude_code/codex_cli

## Description
Curator machine CLI (cmd/curator env resolve|status|config, run/umbrella dispatch operand, profile use --env, env unmanage --env): accept claude and codex as aliases, normalize to claude_code and codex_cli before validation/lookup, print canonical ids everywhere, never persist aliases; help text lists the aliases; tests for both spellings and for the refusal of unknown ids. Spec: curator-spec main 2d6d497 profiles/manager.md CLI alias rule. Launcher side is a separate task.

## Scope
(define task scope)

## Acceptance Criteria
curator run claude and curator run codex launch the claude_code and codex_cli environments; curator env status accepts both spellings; goldens/tests green; help text lists the aliases.
