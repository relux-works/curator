# TASK-261004-34brhn: research-claude-login-transfer-modes

## Description
Research (read-only, no operator credential reads or writes): mechanisms to reuse a Claude Code login in a CLAUDE_CONFIG_DIR managed home on macOS (setup-token / CLAUDE_CODE_OAUTH_TOKEN, apiKeyHelper, per-home Keychain item suffix, .credentials.json under CLAUDE_CONFIG_DIR, refresh semantics), against Decision 0017 (Q1 gate, Q7 no-copy). Output: a modes proposal for the operator.

## Scope
(define task scope)

## Acceptance Criteria
Research document (CIP draft per cip-template.md where the brief says so) under .research/ with options, tradeoffs, recommendation, evidence (file:line or measured probes) and decision-ready open questions; no product code changes; no secrets read or printed; LOGBOOK.md untouched.
