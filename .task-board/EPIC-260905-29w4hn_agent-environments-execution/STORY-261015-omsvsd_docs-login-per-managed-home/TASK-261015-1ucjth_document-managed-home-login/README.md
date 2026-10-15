# TASK-261015-1ucjth: document-managed-home-login

## Description
Add an operator-facing section to docs/environment-config.md on logging in to managed homes, with a per-adapter table (shared/isolated defaults per curator-spec environments section 7.4), the one-time login command for a claude_code home on macOS, and the failure a launch into a home without a login sees. Link it from README.

## Scope
(define task scope)

## Acceptance Criteria
1. Table matches curator-spec section 7.4 for all five adapters. 2. States that a claude_code home on macOS needs one login per profile x environment and that curator never copies tokens. 3. Gives the exact login command and how to find the home path. 4. README links the section. 5. No personal paths or names.
