# spec: CLI aliases claude/codex normalized to claude_code/codex_cli

## Description
curator-spec amendment (small): profiles/manager.md CLI section and the launcher SPEC reference state that the command-line environment operand accepts claude and codex as aliases of claude_code and codex_cli, normalized before validation; diagnostics and outputs print the canonical id; aliases are never written to config, markers, fragments or locks; no schema/vector change. CHANGELOG entry.

## Scope
(define task scope)

## Acceptance Criteria
One alias rule in manager.md (+ launcher SPEC pointer); make validate green; frozen schemas untouched.
